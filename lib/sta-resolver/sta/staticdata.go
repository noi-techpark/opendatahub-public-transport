// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package sta resolves STA SIRI Lite references to GTFS IDs, backed by STA's
// NeTEx and GTFS static data downloaded over FTP.
package sta

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"

	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
	netexq "github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
	_ "github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex/profile" // register profiles
)

// StaticData holds NeTEx + GTFS static data with thread-safe atomic swap.
type StaticData struct {
	netexURL, gtfsURL string

	mu       sync.RWMutex
	resolver *Resolver
}

// LoadStaticData downloads and parses NeTEx + GTFS, creating a Resolver.
// netexURL may contain a "{}" date placeholder, see resolveLatestNeTExURL.
func LoadStaticData(netexURL, gtfsURL string) (*StaticData, error) {
	sd := &StaticData{netexURL: netexURL, gtfsURL: gtfsURL}
	if err := sd.refresh(); err != nil {
		return nil, err
	}
	return sd, nil
}

// GetResolver returns the current resolver under a read lock.
func (sd *StaticData) GetResolver() *Resolver {
	sd.mu.RLock()
	r := sd.resolver
	sd.mu.RUnlock()
	return r
}

// StartRefreshLoop periodically re-downloads and re-parses static data until
// ctx is done. A failed refresh keeps serving the previous data.
func (sd *StaticData) StartRefreshLoop(ctx context.Context, hours int) {
	if hours <= 0 {
		hours = 24
	}
	ticker := time.NewTicker(time.Duration(hours) * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.Info("Refreshing static data...")
			if err := sd.refresh(); err != nil {
				slog.Error("Failed to refresh static data", "err", err)
			} else {
				slog.Info("Static data refreshed successfully")
			}
		}
	}
}

func (sd *StaticData) refresh() error {
	start := time.Now()
	netexURL, gtfsURL := sd.netexURL, sd.gtfsURL

	slog.Info("Downloading GTFS...", "url", gtfsURL)
	gtfsPath, err := downloadFTP(gtfsURL)
	if err != nil {
		return fmt.Errorf("download GTFS: %w", err)
	}
	defer os.Remove(gtfsPath)

	// Resolve the latest available NeTEx file from the dated pattern, then
	// download + parse it. Listing the directory means we always pick whatever
	// date is currently on the server, surviving the provider's daily rotation
	// and retention — a fixed dated URL would eventually 404 on restart.
	resolvedURL, err := resolveLatestNeTExURL(netexURL)
	if err != nil {
		return fmt.Errorf("resolve NeTEx URL: %w", err)
	}
	slog.Info("Downloading NeTEx...", "pattern", netexURL, "url", resolvedURL)
	netexPath, err := downloadFTP(resolvedURL)
	if err != nil {
		return fmt.Errorf("download NeTEx: %w", err)
	}
	defer os.Remove(netexPath)

	resolver, err := loadStaticFromFiles(gtfsPath, netexPath)
	if err != nil {
		return err
	}

	// Atomic swap
	sd.mu.Lock()
	sd.resolver = resolver
	sd.mu.Unlock()

	slog.Info("Static data loaded", "duration", time.Since(start).Round(time.Millisecond))
	return nil
}

// loadStaticFromFiles parses a local GTFS zip and NeTEx zip into a Resolver.
func loadStaticFromFiles(gtfsPath, netexZipPath string) (*Resolver, error) {
	gtfsFeed, err := gtfs.Load(gtfsPath, gtfs.LoadOptions{
		ExcludeTables: []string{"shapes.txt", "translations.txt"},
	}, gtfs.NewMemStore())
	if err != nil {
		return nil, fmt.Errorf("load GTFS: %w", err)
	}
	s := gtfsFeed.Store()
	slog.Info("GTFS loaded",
		"routes", len(s.AllRoutes()),
		"trips", len(s.AllTrips()),
		"stops", len(s.AllStops()),
	)

	netexFeed, err := parseNeTExZip(netexZipPath)
	if err != nil {
		return nil, fmt.Errorf("parse NeTEx: %w", err)
	}
	sjCount, jpCount, routeCount, lineCount := netexFeed.Stats()
	slog.Info("NeTEx loaded",
		"service_journeys", sjCount,
		"journey_patterns", jpCount,
		"routes", routeCount,
		"lines", lineCount,
	)

	return &Resolver{GTFS: gtfsFeed, NeTEx: netexFeed}, nil
}

// parseNeTExZip streams the first XML file of a NeTEx zip into a NeTExFeed.
func parseNeTExZip(zipPath string) (*netexq.NeTExFeed, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	var xmlFile *zip.File
	for _, f := range zr.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".xml") {
			xmlFile = f
			break
		}
	}
	if xmlFile == nil {
		return nil, fmt.Errorf("no XML file found in zip")
	}
	rc, err := xmlFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", xmlFile.Name, err)
	}
	defer rc.Close()

	// Only the types the resolver queries. The parser skips everything else
	// unparsed, notably ~700k TimetabledPassingTimes nested in ServiceJourneys.
	store := netexq.NewMemStore().OnlyTypes(
		"ServiceJourney",
		"ServiceJourneyPattern",
		"Route",
		"Line",
	)

	p, err := netexq.GetProfile("epip")
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if err := netexq.Parse(rc, p, store.Handler(), netexq.WithTypes(store.Accepts)); err != nil {
		return nil, err
	}

	feed := netexq.NewNeTExFeed(store)
	feed.BuildIndexes()
	return feed, nil
}

// ftpConnect dials and authenticates against the FTP server in u, defaulting to
// anonymous/guest when the URL carries no credentials.
func ftpConnect(u *url.URL) (*ftp.ServerConn, error) {
	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":21"
	}

	conn, err := ftp.Dial(host, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", host, err)
	}

	user := "anonymous"
	pass := "guest"
	if u.User != nil {
		user = u.User.Username()
		if p, ok := u.User.Password(); ok {
			pass = p
		}
	}
	if err := conn.Login(user, pass); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("login: %w", err)
	}
	return conn, nil
}

// resolveLatestNeTExURL turns a dated NeTEx URL pattern into a concrete URL
// pointing at the most recent file currently on the server.
//
// The pattern must contain the placeholder "{}" where an 8-digit yyyymmdd date
// appears, e.g. ".../EU_profil/NX-PI_01_it_apb_LINE_apb__{}.xml.zip". The file's
// directory is listed, entries matching the prefix/suffix around the placeholder
// are parsed for their date, and the newest is chosen. A pattern without "{}" is
// returned unchanged (treated as a literal URL).
func resolveLatestNeTExURL(pattern string) (string, error) {
	const placeholder = "{}"
	if !strings.Contains(pattern, placeholder) {
		return pattern, nil
	}

	before, after, _ := strings.Cut(pattern, placeholder)

	u, err := url.Parse(before)
	if err != nil {
		return "", fmt.Errorf("parse pattern: %w", err)
	}
	dir := path.Dir(u.Path)
	filePrefix := path.Base(u.Path) // e.g. "NX-PI_01_it_apb_LINE_apb__"
	fileSuffix := after             // e.g. ".xml.zip"

	conn, err := ftpConnect(u)
	if err != nil {
		return "", err
	}
	defer conn.Quit()

	entries, err := conn.List(dir)
	if err != nil {
		return "", fmt.Errorf("list %s: %w", dir, err)
	}

	var (
		bestDate string
		bestTime time.Time
	)
	for _, e := range entries {
		name := path.Base(e.Name)
		if !strings.HasPrefix(name, filePrefix) || !strings.HasSuffix(name, fileSuffix) {
			continue
		}
		date := name[len(filePrefix) : len(name)-len(fileSuffix)]
		t, err := time.Parse("20060102", date)
		if err != nil {
			continue // not a yyyymmdd-dated file
		}
		if bestDate == "" || t.After(bestTime) {
			bestDate, bestTime = date, t
		}
	}

	if bestDate == "" {
		return "", fmt.Errorf("no NeTEx file matching %q found in %s", filePrefix+placeholder+fileSuffix, dir)
	}

	return before + bestDate + after, nil
}

// downloadFTP downloads a file from an FTP URL to a temp file.
// Supports URLs like ftp://user:pass@host:port/path/to/file.ext
func downloadFTP(ftpURL string) (string, error) {
	u, err := url.Parse(ftpURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}

	conn, err := ftpConnect(u)
	if err != nil {
		return "", err
	}
	defer conn.Quit()

	resp, err := conn.Retr(u.Path)
	if err != nil {
		return "", fmt.Errorf("retrieve %s: %w", u.Path, err)
	}
	defer resp.Close()

	ext := filepath.Ext(u.Path)
	tmp, err := os.CreateTemp("", "ftp-*"+ext)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmp, resp); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("download: %w", err)
	}
	tmp.Close()

	slog.Info("Downloaded", "url", ftpURL, "path", tmp.Name())
	return tmp.Name(), nil
}
