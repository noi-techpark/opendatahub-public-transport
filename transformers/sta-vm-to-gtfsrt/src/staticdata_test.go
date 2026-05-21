// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
)

const (
	testGTFSURL      = "ftp://anonymous:guest@ftp.sta.bz.it/gtfs/google_transit_shp.zip"
	testNeTExPattern = "ftp://anonymous:guest@ftp.sta.bz.it/netex/2026/plan/EU_profil/NX-PI_01_it_apb_LINE_apb__{}.xml.zip"
)

func TestDownloadFTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping FTP download in short mode")
	}

	path, err := downloadFTP(testGTFSURL)
	if err != nil {
		t.Fatalf("download GTFS: %v", err)
	}
	defer os.Remove(path)

	if path == "" {
		t.Fatal("empty path")
	}
	t.Logf("GTFS downloaded to: %s", path)
}

func TestDownloadAndParseGTFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GTFS load in short mode")
	}

	path, err := downloadFTP(testGTFSURL)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer os.Remove(path)

	feed, err := gtfs.Load(path, gtfs.LoadOptions{
		ExcludeTables: []string{"shapes.txt", "translations.txt"},
	}, gtfs.NewMemStore())
	if err != nil {
		t.Fatalf("load GTFS: %v", err)
	}

	s := feed.Store()
	routes := s.AllRoutes()
	trips := s.AllTrips()
	stops := s.AllStops()

	t.Logf("GTFS: %d routes, %d trips, %d stops", len(routes), len(trips), len(stops))

	if len(routes) == 0 {
		t.Error("no routes loaded")
	}
	if len(trips) == 0 {
		t.Error("no trips loaded")
	}
	if len(stops) == 0 {
		t.Error("no stops loaded")
	}
}

func TestDownloadAndParseNeTEx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping NeTEx parse in short mode")
	}

	netexURL, err := resolveLatestNeTExURL(testNeTExPattern)
	if err != nil {
		t.Fatalf("resolve NeTEx URL: %v", err)
	}

	feed, err := downloadAndParseNeTEx(netexURL)
	if err != nil {
		t.Fatalf("parse NeTEx: %v", err)
	}

	sj, jp, routes, lines := feed.Stats()
	t.Logf("NeTEx: %d SJ, %d JP, %d routes, %d lines", sj, jp, routes, lines)

	if sj == 0 {
		t.Error("no service journeys")
	}
	if jp == 0 {
		t.Error("no journey patterns")
	}
	if lines == 0 {
		t.Error("no lines")
	}

	store := feed.Store()
	types := store.Types()
	t.Logf("Entity types in store: %v", types)
	if len(types) < 4 {
		t.Errorf("expected at least 4 entity types, got %d", len(types))
	}
}

func TestResolveLatestNeTExURL(t *testing.T) {
	// A pattern without the placeholder is returned verbatim (no network).
	literal := "ftp://host/netex/file.xml.zip"
	if got, err := resolveLatestNeTExURL(literal); err != nil || got != literal {
		t.Fatalf("literal URL: got (%q, %v), want (%q, nil)", got, err, literal)
	}

	if testing.Short() {
		t.Skip("skipping live FTP listing in short mode")
	}

	resolved, err := resolveLatestNeTExURL(testNeTExPattern)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	t.Logf("resolved: %s", resolved)
	if strings.Contains(resolved, "{}") {
		t.Errorf("placeholder not substituted: %s", resolved)
	}
}

func TestLoadStaticData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full static data load in short mode")
	}

	sd, err := LoadStaticData(testNeTExPattern, testGTFSURL)
	if err != nil {
		t.Fatalf("load static data: %v", err)
	}

	resolver := sd.GetResolver()
	if resolver == nil {
		t.Fatal("resolver is nil")
	}
	if resolver.GTFS == nil {
		t.Error("GTFS feed is nil")
	}
	if resolver.NeTEx == nil {
		t.Error("NeTEx feed is nil")
	}
}
