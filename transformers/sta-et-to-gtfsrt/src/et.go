// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

// SIRI-Lite Estimated Timetable → GTFS-RT TripUpdates conversion.
//
// Mapping strategy:
//   - Route:     SIRI LineRef → GTFS route_short_name → route_id.
//   - Trip:      Two-phase resolution via Resolver (same as VM):
//                  Phase A: NeTEx ServiceJourney ID from DatedVehicleJourneyRef.
//                  Phase B: OriginAimedDepartureTime + line matching (±2min).
//                After resolution, route_id and direction_id are overridden from
//                the matched GTFS trip for consistency.
//   - Stops:     SIRI uses "IT:ITH10:ScheduledStopPoint:X:Y:Z" format.
//                Resolved to GTFS "it:22021:X:Y:Z" via case/prefix normalization.
//                SIRI and GTFS often use different quay IDs for the same physical
//                stop (platform mismatch) — these are different stop_ids in GTFS.
//   - StopTimeUpdates: matched positionally — SIRI calls and GTFS stop_times are
//                both in journey order. A cursor walks GTFS stop_times, matching
//                each SIRI call to the next unmatched entry with the same stop_id.
//                This correctly handles loop routes (same stop visited twice).
//   - Delays:    Only delay is emitted (not absolute time), computed as
//                ExpectedTime - AimedTime from the SIRI source. Emitting absolute
//                times would cause E022 because consecutive stops often share the
//                same aimed minute. Delay-only lets consumers reconstruct times
//                from GTFS static schedule + delay.
//   - Real-time only: StopTimeUpdates are only emitted for stops where SIRI
//                provides ExpectedTime (actual real-time monitoring data). ~44%
//                of calls have only AimedTime (future/unmonitored stops) — these
//                are omitted, letting consumers fall back to the static GTFS
//                schedule via GTFS-RT propagation rules.
//   - Direction: SIRI 1 → GTFS 1 (R), SIRI 2 → GTFS 0 (H).
//
// Drop decisions (correctness over completeness):
//   - Entities without a resolved trip_id are dropped.
//   - Stops not found in GTFS are dropped.
//   - Stops that exist in GTFS but are not on the matched trip's stop_times
//     are dropped (platform/quay mismatch — different stop_id at same station).
//   - Stops without ExpectedArrivalTime AND ExpectedDepartureTime are dropped
//     (no real-time data — consumer uses static schedule).
//   - Entities where stop times decrease after sorting by stop_sequence are
//     dropped (wrong-direction trip match detected).
//   - After all filtering, entities with zero StopTimeUpdates are dropped.
//
// Known gaps:
//   - ~57% of ET journeys can't be resolved (urban lines, missing NeTEx refs).
//   - SIRI ET does not provide VehicleRef — W002 is unavoidable.
//   - Loop routes produce duplicate stop_ids in output (valid per spec when
//     disambiguated by stop_sequence).

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
)

// ConvertET converts a SIRI ET feed to a GTFS-RT TripUpdates feed.
func ConvertET(feed *siri.ETFeed, resolver *Resolver) *gtfsrt.FeedMessage {
	fm := gtfsrt.NewFeedMessage()

	journeys := feed.AllEstimatedVehicleJourneys()

	for _, evj := range journeys {
		lineRef := evj.LineRef
		routeID := resolver.ResolveRouteID(lineRef)
		if routeID == "" && evj.PublishedLineName != "" && evj.PublishedLineName != lineRef {
			lineRef = evj.PublishedLineName
			routeID = resolver.ResolveRouteID(lineRef)
		}
		directionID := resolver.ResolveDirectionID(evj.DirectionRef)
		startDate := reformatDate(evj.FramedVehicleJourneyRef.DataFrameRef)

		// Try to resolve trip via NeTEx SJ or origin departure time
		tripID := resolveETTrip(evj, resolver, startDate)

		// Skip entities without a resolved trip
		if tripID == "" {
			continue
		}

		if t := resolver.GTFS.Trip(tripID); t != nil {
			routeID = t.RouteID
			directionID = t.DirectionID
		}

		trip := &gtfsrt.TripDescriptor{
			RouteID:              routeID,
			TripID:               tripID,
			DirectionID:          gtfsrt.IntPtr(directionID),
			StartDate:            startDate,
			ScheduleRelationship: "SCHEDULED",
		}

		// Match SIRI EstimatedCalls to GTFS stop_times positionally.
		// Both are in journey order. For loop routes, the same stop appears
		// multiple times — we use a cursor to match each SIRI call to the
		// next unmatched GTFS stop_time with the same stop_id.
		gtfsStopTimes := resolver.GTFS.StopTimesForTrip(tripID)
		matched := make([]bool, len(gtfsStopTimes))

		var stopTimeUpdates []gtfsrt.StopTimeUpdate
		for _, call := range evj.EstimatedCalls.EstimatedCall {
			stopID := resolveETStopRef(call.StopPointRef, resolver)
			if stopID == "" {
				continue
			}

			// Find next unmatched GTFS stop_time for this stop_id
			seq := -1
			for i, st := range gtfsStopTimes {
				if !matched[i] && st.StopID == stopID {
					seq = st.StopSequence
					matched[i] = true
					break
				}
			}
			if seq < 0 {
				continue // stop not on this GTFS trip
			}

			stu := gtfsrt.StopTimeUpdate{
				StopID:               stopID,
				StopSequence:         seq,
				ScheduleRelationship: "SCHEDULED",
			}

			// Only emit arrival/departure when SIRI provides ExpectedTime
			// (actual real-time data). Stops with only AimedTime have no
			// real-time info — omitting them lets the consumer fall back to
			// the static GTFS schedule per the GTFS-RT propagation rules.
			//
			// We emit delay only (not absolute time) to avoid E022: the SIRI
			// source often has consecutive stops with identical aimed minutes,
			// which would produce equal absolute timestamps.
			if call.AimedArrivalTime != "" && call.ExpectedArrivalTime != "" {
				aimed, err1 := parseISO8601Time(call.AimedArrivalTime)
				expected, err2 := parseISO8601Time(call.ExpectedArrivalTime)
				if err1 == nil && err2 == nil {
					stu.Arrival = &gtfsrt.StopTimeEvent{
						Delay: int32(expected.Unix() - aimed.Unix()),
					}
				}
			}

			if call.AimedDepartureTime != "" && call.ExpectedDepartureTime != "" {
				aimed, err1 := parseISO8601Time(call.AimedDepartureTime)
				expected, err2 := parseISO8601Time(call.ExpectedDepartureTime)
				if err1 == nil && err2 == nil {
					stu.Departure = &gtfsrt.StopTimeEvent{
						Delay: int32(expected.Unix() - aimed.Unix()),
					}
				}
			}

			// E043: every StopTimeUpdate must have at least arrival or departure.
			// Skip stops where we have no real-time data at all.
			if stu.Arrival == nil && stu.Departure == nil {
				continue
			}

			stopTimeUpdates = append(stopTimeUpdates, stu)
		}

		// Skip if no stop_time_updates survived filtering
		if len(stopTimeUpdates) == 0 {
			continue
		}

		// Sort by stop_sequence (E002: must be strictly sorted)
		slices.SortFunc(stopTimeUpdates, func(a, b gtfsrt.StopTimeUpdate) int {
			return a.StopSequence - b.StopSequence
		})

		// Validate times are non-decreasing (E022).
		// If they decrease, we matched a wrong-direction trip — drop the entity.
		if !timesNonDecreasing(stopTimeUpdates) {
			continue
		}

		// Parse timestamp (W001: always set it)
		var timestamp int64
		if t, err := parseISO8601Time(evj.RecordedAtTime); err == nil {
			timestamp = t.Unix()
		}
		if timestamp == 0 {
			timestamp = fm.Header.Timestamp
		}

		tu := &gtfsrt.TripUpdate{
			Trip:           trip,
			StopTimeUpdate: stopTimeUpdates,
			Timestamp:      timestamp,
		}

		// Entity ID: use DatedVehicleJourneyRef or fallback
		entityID := evj.FramedVehicleJourneyRef.DatedVehicleJourneyRef
		if entityID == "" {
			entityID = fmt.Sprintf("et-%s-%s", evj.LineRef, evj.OriginAimedDepartureTime)
		}

		fm.AddEntity(gtfsrt.FeedEntity{
			ID:         entityID,
			TripUpdate: tu,
		})
	}

	return fm
}

// resolveETTrip tries to match an ET journey to a GTFS trip.
func resolveETTrip(evj siri.EstimatedVehicleJourney, resolver *Resolver, dateStr string) string {
	date := gtfs.ParseDate(dateStr)
	if !date.IsSet() {
		return ""
	}

	// Try NeTEx SJ ID extraction from DatedVehicleJourneyRef
	ref := evj.FramedVehicleJourneyRef.DatedVehicleJourneyRef
	if sjBase := extractNeTExSJID(ref); sjBase != "" {
		if tripID := resolver.matchViaNeTExSJ(sjBase, date, evj.DirectionRef); tripID != "" {
			resolver.TripsResolvedA++
			return tripID
		}
	}

	// Try matching via OriginAimedDepartureTime + line
	publicCode := resolver.resolveToPublicCode(evj.LineRef)
	if publicCode == "" {
		resolver.TripsUnresolved++
		return ""
	}

	if evj.OriginAimedDepartureTime != "" {
		originTime, err := parseISO8601Time(evj.OriginAimedDepartureTime)
		if err == nil {
			depSeconds := originTime.Hour()*3600 + originTime.Minute()*60 + originTime.Second()
			candidates := resolver.GTFS.FindTripsForLine(publicCode, date)
			gtfsDir := mapDirectionRef(evj.DirectionRef)

			var dirFiltered []*gtfs.Trip
			for _, t := range candidates {
				if t.DirectionID == gtfsDir {
					dirFiltered = append(dirFiltered, t)
				}
			}
			if len(dirFiltered) == 0 {
				dirFiltered = candidates
			}

			best := resolver.GTFS.MatchTripIn(dirFiltered, func(trip *gtfs.Trip, sts []*gtfs.StopTime) float64 {
				if len(sts) == 0 {
					return 0
				}
				firstDep := sts[0].DepartureTime
				if !firstDep.IsSet() {
					firstDep = sts[0].ArrivalTime
				}
				if !firstDep.IsSet() {
					return 0
				}
				diff := math.Abs(float64(firstDep.Seconds() - depSeconds))
				if diff > 120 {
					return 0
				}
				return 1.0 / (1.0 + diff)
			})

			if best != nil {
				resolver.TripsResolvedB++
				return best.TripID
			}
		}
	}

	resolver.TripsUnresolved++
	return ""
}

// resolveETStopRef converts an ET StopPointRef to a GTFS stop_id.
// ET uses formats like:
//
//	"IT:ITH10:ScheduledStopPoint:1140:0:8843" → "it:22021:1140:0:8843"
//	"IT:22021:3511:0:4" → "it:22021:3511:0:4"
//
// GTFS uses lowercase "it:22021:..."
func resolveETStopRef(ref string, resolver *Resolver) string {
	// Try direct match
	if s := resolver.GTFS.Stop(ref); s != nil {
		return ref
	}

	// Try lowercase (IT:22021:... → it:22021:...)
	lower := strings.ToLower(ref)
	if s := resolver.GTFS.Stop(lower); s != nil {
		return lower
	}

	// Strip "IT:ITH10:ScheduledStopPoint:" prefix → rebuild as "it:22021:..."
	for _, prefix := range []string{
		"IT:ITH10:ScheduledStopPoint:",
		"it:ITH10:ScheduledStopPoint:",
	} {
		if strings.HasPrefix(ref, prefix) || strings.HasPrefix(lower, strings.ToLower(prefix)) {
			suffix := ref[len(prefix):]
			candidate := "it:22021:" + suffix
			if s := resolver.GTFS.Stop(candidate); s != nil {
				return candidate
			}
		}
	}

	// Return empty if not found in GTFS (don't emit invalid IDs)
	return ""
}

// timesNonDecreasing checks that arrival/departure times never go backwards.
func timesNonDecreasing(stus []gtfsrt.StopTimeUpdate) bool {
	var prevTime int64
	for _, stu := range stus {
		if stu.Arrival != nil && stu.Arrival.Time > 0 {
			if prevTime > 0 && stu.Arrival.Time < prevTime {
				return false
			}
			prevTime = stu.Arrival.Time
		}
		if stu.Departure != nil && stu.Departure.Time > 0 {
			if prevTime > 0 && stu.Departure.Time < prevTime {
				return false
			}
			prevTime = stu.Departure.Time
		}
	}
	return true
}
