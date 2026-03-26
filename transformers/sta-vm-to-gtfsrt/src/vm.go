// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
)

// ConvertVM converts a SIRI VM feed to a GTFS-RT VehiclePositions feed.
func ConvertVM(feed *siri.VMFeed, resolver *Resolver) *gtfsrt.FeedMessage {
	fm := gtfsrt.NewFeedMessage()

	activities := feed.ServiceDelivery.VehicleMonitoringDelivery.VehicleActivity

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
