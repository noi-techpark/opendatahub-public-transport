package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
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

		// Build set of stops that are actually on this trip's stop sequence
		tripStopSet := make(map[string]bool)
		for _, st := range resolver.GTFS.StopTimesForTrip(tripID) {
			tripStopSet[st.StopID] = true
		}

		// Build StopTimeUpdates from EstimatedCalls
		var stopTimeUpdates []gtfsrt.StopTimeUpdate
		for _, call := range evj.EstimatedCalls.EstimatedCall {
			stopID := resolveETStopRef(call.StopPointRef, resolver)
			if stopID == "" || !tripStopSet[stopID] {
				continue // skip stops not in GTFS or not on this trip
			}

			stu := gtfsrt.StopTimeUpdate{
				StopID: stopID,
			}

			// Arrival
			if call.AimedArrivalTime != "" {
				stu.Arrival = &gtfsrt.StopTimeEvent{}
				if t, err := parseISO8601Time(call.AimedArrivalTime); err == nil {
					stu.Arrival.Time = t.Unix()
				}
				if call.ExpectedArrivalTime != "" {
					if t, err := parseISO8601Time(call.ExpectedArrivalTime); err == nil {
						delay := t.Unix() - stu.Arrival.Time
						stu.Arrival.Delay = int32(delay)
					}
				}
			}

			// Departure
			if call.AimedDepartureTime != "" {
				stu.Departure = &gtfsrt.StopTimeEvent{}
				if t, err := parseISO8601Time(call.AimedDepartureTime); err == nil {
					stu.Departure.Time = t.Unix()
				}
				if call.ExpectedDepartureTime != "" {
					if t, err := parseISO8601Time(call.ExpectedDepartureTime); err == nil {
						delay := t.Unix() - stu.Departure.Time
						stu.Departure.Delay = int32(delay)
					}
				}
			}

			stopTimeUpdates = append(stopTimeUpdates, stu)
		}

		tu := &gtfsrt.TripUpdate{
			Trip:           trip,
			StopTimeUpdate: stopTimeUpdates,
		}

		// Parse timestamp
		var timestamp int64
		if t, err := parseISO8601Time(evj.RecordedAtTime); err == nil {
			timestamp = t.Unix()
		}
		tu.Timestamp = timestamp

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

			best := resolver.GTFS.MatchTripIn(dirFiltered, func(trip *gtfs.Trip, sts []gtfs.StopTime) float64 {
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
//   "IT:ITH10:ScheduledStopPoint:1140:0:8843" → "it:22021:1140:0:8843"
//   "IT:22021:3511:0:4" → "it:22021:3511:0:4"
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
