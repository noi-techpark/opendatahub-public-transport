// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"os"
	"testing"
)

func TestET_ArrayFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/et_array.json")
	if err != nil {
		t.Fatal(err)
	}
	et, err := DeserializeET(data, FormatJSON)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	frames := et.ServiceDelivery.EstimatedTimetableDelivery.EstimatedJourneyVersionFrame
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}
	journeys := frames[0].EstimatedVehicleJourney
	if len(journeys) != 1 {
		t.Fatalf("expected 1 journey, got %d", len(journeys))
	}
	if journeys[0].LineRef != "240" {
		t.Errorf("LineRef = %q, want 240", journeys[0].LineRef)
	}
	calls := journeys[0].EstimatedCalls.EstimatedCall
	if len(calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calls))
	}
	if calls[0].StopPointRef != "stop_1" {
		t.Errorf("first call StopPointRef = %q", calls[0].StopPointRef)
	}
}

func TestET_SingleObjectFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/et_single.json")
	if err != nil {
		t.Fatal(err)
	}
	et, err := DeserializeET(data, FormatJSON)
	if err != nil {
		t.Fatalf("deserialize single: %v", err)
	}

	// Single frame (not array)
	frames := et.ServiceDelivery.EstimatedTimetableDelivery.EstimatedJourneyVersionFrame
	if len(frames) != 1 {
		t.Fatalf("expected 1 frame from single object, got %d", len(frames))
	}

	// Single journey (not array)
	journeys := frames[0].EstimatedVehicleJourney
	if len(journeys) != 1 {
		t.Fatalf("expected 1 journey from single object, got %d", len(journeys))
	}

	// Single call (not array)
	calls := journeys[0].EstimatedCalls.EstimatedCall
	if len(calls) != 1 {
		t.Fatalf("expected 1 call from single object, got %d", len(calls))
	}
	if calls[0].StopPointRef != "stop_1" {
		t.Errorf("StopPointRef = %q, want stop_1", calls[0].StopPointRef)
	}
}

func TestET_AllEstimatedVehicleJourneys(t *testing.T) {
	data, err := os.ReadFile("testdata/et_array.json")
	if err != nil {
		t.Fatal(err)
	}
	et, err := DeserializeET(data, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	all := et.AllEstimatedVehicleJourneys()
	if len(all) != 1 {
		t.Errorf("expected 1 total journey, got %d", len(all))
	}
}

func TestET_RoundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/et_array.json")
	if err != nil {
		t.Fatal(err)
	}
	et, err := DeserializeET(data, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	out, err := et.Serialize(FormatJSON)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	et2, err := DeserializeET(out, FormatJSON)
	if err != nil {
		t.Fatalf("re-deserialize: %v", err)
	}
	if len(et2.AllEstimatedVehicleJourneys()) != 1 {
		t.Error("round-trip lost journeys")
	}
}
