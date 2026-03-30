// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

// SIRI-Lite Vehicle Monitoring → GTFS-RT VehiclePositions conversion.
//
// Mapping strategy:
//   - Route:     SIRI LineRef (public code like "240") → GTFS route_short_name → route_id.
//   - Trip:      Two-phase resolution via Resolver:
//                  Phase A: extract NeTEx ServiceJourney ID from DatedVehicleJourneyRef,
//                           traverse NeTEx SJ → JP → Route → Line → GTFS trips,
//                           score by first-stop departure time proximity (±2min).
//                  Phase B: fallback stop+time matching — compute scheduled time at
//                           current stop (recordedTime - delay), match against GTFS
//                           stop_times at that stop on the service date (±10min).
//   - Stop:      strip NeTEx prefix "it:apb:ScheduledStopPoint:" → GTFS stop_id.
//   - Direction: SIRI 1 → GTFS 1 (R/return), SIRI 2 → GTFS 0 (H/outbound).
//                After trip resolution, direction is overridden from the matched trip.
//
// Drop decisions (correctness over completeness):
//   - Entities without a resolved trip_id are dropped — can't pair with GTFS.
//   - Stops not found in GTFS are omitted from the entity.
//   - Duplicate VehicleRefs (trip transitions): keep only the most recent
//     RecordedAtTime per VehicleRef.
//
// Known gaps:
//   - Urban city lines (1-12) use different naming in SIRI vs GTFS → unresolved.
//   - ~17% of vehicles can't be matched to a GTFS trip.

import (
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
)

// deduplicateByVehicle keeps only the most recent VehicleActivity per VehicleRef.
// When a bus transitions between trips, the feed may contain both the old and new entry.
func deduplicateByVehicle(activities []siri.VehicleActivity) []siri.VehicleActivity {
	best := make(map[string]siri.VehicleActivity, len(activities))
	for _, va := range activities {
		ref := va.MonitoredVehicleJourney.VehicleRef
		if existing, ok := best[ref]; ok {
			// Keep the one with the later RecordedAtTime
			if va.RecordedAtTime > existing.RecordedAtTime {
				best[ref] = va
			}
		} else {
			best[ref] = va
		}
	}
	result := make([]siri.VehicleActivity, 0, len(best))
	for _, va := range best {
		result = append(result, va)
	}
	return result
}

// ConvertVM converts a SIRI VM feed to a GTFS-RT VehiclePositions feed.
func ConvertVM(feed *siri.VMFeed, resolver *Resolver) *gtfsrt.FeedMessage {
	fm := gtfsrt.NewFeedMessage()

	// Deduplicate: keep only the most recent entry per VehicleRef.
	// The feed can contain stale entries from a previous trip alongside the current one
	// (e.g., bus transitioning between trips).
	activities := deduplicateByVehicle(feed.ServiceDelivery.VehicleMonitoringDelivery.VehicleActivity)

	for _, va := range activities {
		mvj := va.MonitoredVehicleJourney

		// Resolve IDs — try LineRef first, fall back to PublishedLineName
		lineRef := mvj.LineRef
		routeID := resolver.ResolveRouteID(lineRef)
		if routeID == "" && mvj.PublishedLineName != "" && mvj.PublishedLineName != lineRef {
			lineRef = mvj.PublishedLineName
			routeID = resolver.ResolveRouteID(lineRef)
		}
		stopID := resolver.ResolveStopID(mvj.MonitoredCall.StopPointRef)
		tripID := resolver.ResolveTripID(
			lineRef,
			mvj.DirectionRef,
			mvj.FramedVehicleJourneyRef.DataFrameRef,
			mvj.FramedVehicleJourneyRef.DatedVehicleJourneyRef,
			mvj.MonitoredCall.StopPointRef,
			va.RecordedAtTime,
			mvj.Delay,
		)
		directionID := resolver.ResolveDirectionID(mvj.DirectionRef)

		// Parse timestamp
		var timestamp int64
		if t, err := parseISO8601Time(va.RecordedAtTime); err == nil {
			timestamp = t.Unix()
		}

		// Parse position
		lat := parseFloat32(mvj.VehicleLocation.Latitude)
		lon := parseFloat32(mvj.VehicleLocation.Longitude)

		// Current status
		currentStatus := "IN_TRANSIT_TO"
		if mvj.MonitoredCall.VehicleAtStop == "true" {
			currentStatus = "STOPPED_AT"
		}

		// Congestion
		congestion := ""
		if mvj.InCongestion == "true" {
			congestion = "CONGESTION"
		}

		// Start date
		startDate := reformatDate(mvj.FramedVehicleJourneyRef.DataFrameRef)

		// If trip was resolved, use the trip's actual route_id and direction
		// to ensure consistency (avoids version mismatches like 86-351-26a-2 vs 26a-4)
		if tripID != "" {
			if t := resolver.GTFS.Trip(tripID); t != nil {
				routeID = t.RouteID
				directionID = t.DirectionID
			}
		}

		// Only emit stop_id if it exists in GTFS
		if resolver.GTFS.Stop(stopID) == nil {
			stopID = ""
		}

		// Skip entities without a resolved trip — they can't be paired with static GTFS
		if tripID == "" {
			continue
		}

		// Build trip descriptor
		trip := &gtfsrt.TripDescriptor{
			RouteID:              routeID,
			TripID:               tripID,
			DirectionID:          gtfsrt.IntPtr(directionID),
			StartDate:            startDate,
			ScheduleRelationship: "SCHEDULED",
		}

		vp := &gtfsrt.VehiclePosition{
			Trip: trip,
			Vehicle: &gtfsrt.VehicleDescriptor{
				ID:    mvj.VehicleRef,
				Label: mvj.PublishedLineName,
			},
			Position: &gtfsrt.Position{
				Latitude:  lat,
				Longitude: lon,
			},
			StopID:          stopID,
			CurrentStatus:   currentStatus,
			Timestamp:       timestamp,
			CongestionLevel: congestion,
		}

		fm.AddEntity(gtfsrt.FeedEntity{
			ID:      mvj.VehicleRef,
			Vehicle: vp,
		})
	}

	return fm
}
