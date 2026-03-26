// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
)

// TestVMTransformPipeline does a full end-to-end test:
// download real static data, fetch a live SIRI VM feed, convert to GTFS-RT.
func TestVMTransformPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping pipeline test in short mode")
	}

	// Load static data
	sd, err := LoadStaticData(testNeTExURL, testGTFSURL)
	if err != nil {
		t.Fatalf("load static data: %v", err)
	}

	// Fetch live SIRI VM feed
	vmURL := "https://efa.sta.bz.it/siri-lite/vehicle-monitoring"
	resp, err := http.Get(vmURL)
	if err != nil {
		t.Fatalf("fetch VM feed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("VM feed returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	vm, err := siri.DeserializeVM(body, siri.FormatJSON)
	if err != nil {
		t.Fatalf("deserialize VM: %v", err)
	}

	activities := vm.ServiceDelivery.VehicleMonitoringDelivery.VehicleActivity
	t.Logf("SIRI VM: %d vehicle activities", len(activities))

	// Convert
	resolver := sd.GetResolver()
	rt := ConvertVM(vm, resolver)

	t.Logf("GTFS-RT: %d entities", len(rt.Entity))
	t.Logf("Resolver stats: trips_a=%d, trips_b=%d, unresolved=%d",
		resolver.TripsResolvedA, resolver.TripsResolvedB, resolver.TripsUnresolved)

	if len(rt.Entity) == 0 && len(activities) > 0 {
		t.Error("no GTFS-RT entities produced from non-empty SIRI feed")
	}

	// Verify serialization
	pbData, err := rt.Serialize(gtfsrt.FormatProtobuf)
	if err != nil {
		t.Fatalf("serialize protobuf: %v", err)
	}
	if len(pbData) == 0 {
		t.Error("empty protobuf output")
	}

	jsonData, err := rt.Serialize(gtfsrt.FormatJSON)
	if err != nil {
		t.Fatalf("serialize JSON: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("empty JSON output")
	}

	t.Logf("Output: %d bytes protobuf, %d bytes JSON", len(pbData), len(jsonData))
}
