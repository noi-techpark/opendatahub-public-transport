// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/noi-techpark/opendatahub-go-sdk/tel"

	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
	netexq "github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
	_ "github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex/profile" // register profiles
)

// StaticData holds NeTEx + GTFS static data with thread-safe atomic swap.
type StaticData struct {
	mu       sync.RWMutex
	resolver *Resolver
}

// LoadStaticData downloads and parses NeTEx + GTFS, creating a Resolver.
func LoadStaticData(netexURL, gtfsURL string) (*StaticData, error) {
	sd := &StaticData{}
	if err := sd.refresh(netexURL, gtfsURL); err != nil {
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

// StartRefreshLoop periodically re-downloads and re-parses static data.
// Critical panics propagate via tel.FlushOnPanic() to kill the pod.
func (sd *StaticData) StartRefreshLoop(ctx context.Context, hours int) {
	defer tel.FlushOnPanic()

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
			if err := sd.refresh(env.NETEX_FTP_URL, env.GTFS_FTP_URL); err != nil {
				slog.Error("Failed to refresh static data", "err", err)
			} else {
				slog.Info("Static data refreshed successfully")
			}
		}
	}
}

func (sd *StaticData) refresh(netexURL, gtfsURL string) error {
	start := time.Now()

	// Download and parse GTFS
	slog.Info("Downloading GTFS...", "url", gtfsURL)
	gtfsPath, err := downloadFTP(gtfsURL)
	if err != nil {
		return fmt.Errorf("download GTFS: %w", err)
	}
	defer os.Remove(gtfsPath)

	gtfsFeed, err := gtfs.Load(gtfsPath, gtfs.LoadOptions{
		ExcludeTables: []string{"shapes.txt", "translations.txt"},
	}, gtfs.NewMemStore())
	if err != nil {
		return fmt.Errorf("load GTFS: %w", err)
	}
	s := gtfsFeed.Store()
	slog.Info("GTFS loaded",
		"routes", len(s.AllRoutes()),
		"trips", len(s.AllTrips()),
		"stops", len(s.AllStops()),
	)

	// Download and parse NeTEx
	slog.Info("Downloading NeTEx...", "url", netexURL)
	netexFeed, err := downloadAndParseNeTEx(netexURL)
	if err != nil {
		return fmt.Errorf("parse NeTEx: %w", err)
	}
	sjCount, jpCount, routeCount, lineCount := netexFeed.Stats()
	slog.Info("NeTEx loaded",
		"service_journeys", sjCount,
		"journey_patterns", jpCount,
		"routes", routeCount,
		"lines", lineCount,
	)

	// Build resolver
	resolver := &Resolver{
		GTFS:  gtfsFeed,
		NeTEx: netexFeed,
	}

	// Atomic swap
	sd.mu.Lock()
	sd.resolver = resolver
	sd.mu.Unlock()

	slog.Info("Static data loaded", "duration", time.Since(start).Round(time.Millisecond))
	return nil
}

// downloadAndParseNeTEx downloads a NeTEx XML zip, parses it, and builds a NeTExFeed.
func downloadAndParseNeTEx(ftpURL string) (*netexq.NeTExFeed, error) {
	zipPath, err := downloadFTP(ftpURL)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer os.Remove(zipPath)

	// Extract XML from zip
	xmlPath, err := extractFirstXML(zipPath)
	if err != nil {
		return nil, fmt.Errorf("extract XML: %w", err)
	}
	defer os.Remove(xmlPath)

	// Parse NeTEx XML into store — only load entity types needed for querying.
	// This saves ~800MB by skipping TimetabledPassingTime (695k),
	// StopPointInJourneyPattern (160k), and PointOnRoute (149k).
	store := netexq.NewMemStore().OnlyTypes(
		"ServiceJourney",
		"ServiceJourneyPattern",
		"Route",
		"Line",
	)

	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("open XML: %w", err)
	}
	defer xmlFile.Close()

	p, err := netexq.GetProfile("epip")
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if err := netexq.Parse(xmlFile, p, store.Handler()); err != nil {
		return nil, fmt.Errorf("parse NeTEx: %w", err)
	}

	// Build query feed
	feed := netexq.NewNeTExFeed(store)
	feed.BuildIndexes()
	return feed, nil
}

// extractFirstXML extracts the first .xml file from a zip archive to a temp file.
func extractFirstXML(zipPath string) (string, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".xml") {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			tmp, err := os.CreateTemp("", "netex-*.xml")
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(tmp, rc); err != nil {
				tmp.Close()
				os.Remove(tmp.Name())
				return "", err
			}
			tmp.Close()
			return tmp.Name(), nil
		}
	}
	return "", fmt.Errorf("no XML file found in zip")
}

// downloadFTP downloads a file from an FTP URL to a temp file.
// Supports URLs like ftp://user:pass@host:port/path/to/file.ext
func downloadFTP(ftpURL string) (string, error) {
	u, err := url.Parse(ftpURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":21"
	}

	conn, err := ftp.Dial(host, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return "", fmt.Errorf("dial %s: %w", host, err)
	}
	defer conn.Quit()

	user := "anonymous"
	pass := "guest"
	if u.User != nil {
		user = u.User.Username()
		if p, ok := u.User.Password(); ok {
			pass = p
		}
	}

	if err := conn.Login(user, pass); err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

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
