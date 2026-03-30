// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
)

// testGTFSURL and testNeTExURL defined in staticdata_test.go

func TestETTransformPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skip in short mode")
	}

	sd, err := LoadStaticData(testNeTExURL, testGTFSURL)
	if err != nil {
		t.Fatalf("load static data: %v", err)
	}

	resp, err := http.Get("https://efa.sta.bz.it/siri-lite/estimated-timetable")
	if err != nil {
		t.Fatalf("fetch ET feed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	et, err := siri.DeserializeET(body, siri.FormatJSON)
	if err != nil {
		t.Fatalf("deserialize ET: %v", err)
	}

	journeys := et.AllEstimatedVehicleJourneys()
	t.Logf("SIRI ET: %d journeys", len(journeys))

	resolver := sd.GetResolver()
	rt := ConvertET(et, resolver)

	t.Logf("GTFS-RT: %d trip update entities", len(rt.Entity))
	t.Logf("Resolver: trips_a=%d, trips_b=%d, unresolved=%d",
		resolver.TripsResolvedA, resolver.TripsResolvedB, resolver.TripsUnresolved)

	if len(rt.Entity) == 0 && len(journeys) > 0 {
		t.Error("no GTFS-RT entities from non-empty ET feed")
	}

	// Correctness audit
	var dirMismatch, tripNotActive, stopNotOnTrip, timeImplausible int

	for _, e := range rt.Entity {
		if e.TripUpdate == nil {
			t.Errorf("entity %s has nil TripUpdate", e.ID)
			continue
		}
		tu := e.TripUpdate
		if tu.Trip == nil || tu.Trip.TripID == "" {
			t.Errorf("entity %s has no trip descriptor", e.ID)
			continue
		}

		trip := resolver.GTFS.Trip(tu.Trip.TripID)
		if trip == nil {
			t.Errorf("entity %s: trip %s not in GTFS", e.ID, tu.Trip.TripID)
			continue
		}

		// Direction check
		if tu.Trip.DirectionID != nil && *tu.Trip.DirectionID != trip.DirectionID {
			dirMismatch++
		}

		// Service active check
		date := gtfs.ParseDate(tu.Trip.StartDate)
		if date.IsSet() && !resolver.GTFS.ServiceRunsOn(trip.ServiceID, date) {
			tripNotActive++
		}

		// Stop consistency
		tripStops := resolver.GTFS.StopTimesForTrip(tu.Trip.TripID)
		tripStopSet := make(map[string]bool, len(tripStops))
		for _, st := range tripStops {
			tripStopSet[st.StopID] = true
		}
		for _, stu := range tu.StopTimeUpdate {
			if stu.StopID != "" && !tripStopSet[stu.StopID] {
				stopNotOnTrip++
			}
		}

		// Delay plausibility
		for _, stu := range tu.StopTimeUpdate {
			if stu.Arrival != nil && math.Abs(float64(stu.Arrival.Delay)) > 3600 {
				timeImplausible++
			}
			if stu.Departure != nil && math.Abs(float64(stu.Departure.Delay)) > 3600 {
				timeImplausible++
			}
		}
	}

	// E036/E037: duplicate stop_sequence or stop_id
	dupSeq, dupStop, unsorted := 0, 0, 0
	for _, e := range rt.Entity {
		tu := e.TripUpdate
		if tu == nil {
			continue
		}
		seqSeen := map[int]bool{}
		stopSeen := map[string]bool{}
		prevSeq := -1
		for _, stu := range tu.StopTimeUpdate {
			if seqSeen[stu.StopSequence] {
				dupSeq++
				break
			}
			seqSeen[stu.StopSequence] = true
			if stopSeen[stu.StopID] {
				dupStop++
				break
			}
			stopSeen[stu.StopID] = true
			if stu.StopSequence <= prevSeq {
				unsorted++
				break
			}
			prevSeq = stu.StopSequence
		}
	}

	t.Logf("Correctness audit (%d entities):", len(rt.Entity))
	t.Logf("  Direction mismatches:  %d", dirMismatch)
	t.Logf("  Trip not active:       %d", tripNotActive)
	t.Logf("  Stops not on trip:     %d", stopNotOnTrip)
	t.Logf("  Delays > 1 hour:       %d", timeImplausible)
	t.Logf("  Dup stop_sequence:     %d", dupSeq)
	t.Logf("  Dup stop_id:           %d", dupStop)
	t.Logf("  Unsorted sequences:    %d", unsorted)

	if dirMismatch > len(rt.Entity)/2 {
		t.Errorf("too many direction mismatches: %d/%d", dirMismatch, len(rt.Entity))
	}
	if dupSeq > 0 {
		t.Errorf("E036: %d entities with duplicate stop_sequence", dupSeq)
	}
	if dupStop > 0 {
		// Duplicate stop_id is valid for loop routes when stop_sequence disambiguates.
		t.Logf("  (E037 note: %d entities with duplicate stop_id — valid for loop routes with stop_sequence set)", dupStop)
	}
	if unsorted > 0 {
		t.Errorf("E002: %d entities with unsorted stop_sequences", unsorted)
	}

	// Serialization
	pbData, err := rt.Serialize(gtfsrt.FormatProtobuf)
	if err != nil {
		t.Fatalf("serialize protobuf: %v", err)
	}
	jsonData, err := rt.Serialize(gtfsrt.FormatJSON)
	if err != nil {
		t.Fatalf("serialize JSON: %v", err)
	}
	t.Logf("Output: %d bytes protobuf, %d bytes JSON", len(pbData), len(jsonData))

	_ = fmt.Sprintf("") // keep fmt import
}
