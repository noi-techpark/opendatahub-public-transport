// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
)

func main() {
	zipPath := flag.String("zip", "", "Path to GTFS zip file (required)")
	loadShapes := flag.Bool("shapes", false, "Also load shapes.txt (large)")
	loadTranslations := flag.Bool("translations", false, "Also load translations.txt")
	queryStop := flag.String("stop", "", "Look up a stop by ID")
	queryRoute := flag.String("route", "", "Look up routes by short name")
	queryTrip := flag.String("trip", "", "Look up a trip by ID")
	queryDate := flag.String("date", "", "Show active services/trips for date (YYYYMMDD)")
	flag.Parse()

	if *zipPath == "" {
		fmt.Fprintln(os.Stderr, "error: -zip is required")
		flag.Usage()
		os.Exit(1)
	}

	// Load
	start := time.Now()
	store := gtfs.NewMemStore()
	feed, err := gtfs.Load(*zipPath, gtfs.LoadOptions{
		LoadShapes:       *loadShapes,
		LoadTranslations: *loadTranslations,
	}, store)
	if err != nil {
		log.Fatalf("load: %v", err)
	}
	elapsed := time.Since(start)

	// Memory stats
	var mem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&mem)

	s := feed.Store()
	fmt.Printf("Loaded in %v (~%dMB heap)\n", elapsed.Round(time.Millisecond), mem.Alloc/1024/1024)
	fmt.Printf("  agencies:        %d\n", len(s.AllAgencies()))
	fmt.Printf("  stops:           %d\n", len(s.AllStops()))
	fmt.Printf("  routes:          %d\n", len(s.AllRoutes()))
	fmt.Printf("  trips:           %d\n", len(s.AllTrips()))
	fmt.Printf("  stop_times:      %d\n", len(s.AllStopTimes()))
	fmt.Printf("  calendar:        %d\n", len(s.AllCalendar()))
	fmt.Printf("  calendar_dates:  %d\n", len(s.AllCalendarDates()))

	// Queries
	if *queryStop != "" {
		queryStopByID(feed, *queryStop)
	}
	if *queryRoute != "" {
		queryRouteByName(feed, *queryRoute)
	}
	if *queryTrip != "" {
		queryTripByID(feed, *queryTrip)
	}
	if *queryDate != "" {
		queryActiveOnDate(feed, *queryDate)
	}
}

func queryStopByID(feed *gtfs.Feed, stopID string) {
	fmt.Printf("\n--- Stop: %s ---\n", stopID)
	s := feed.Stop(stopID)
	if s == nil {
		fmt.Println("  not found")
		return
	}
	fmt.Printf("  name:           %s\n", s.StopName)
	fmt.Printf("  lat/lon:        %.6f, %.6f\n", s.StopLat, s.StopLon)
	fmt.Printf("  location_type:  %d\n", s.LocationType)
	fmt.Printf("  parent_station: %s\n", s.ParentStation)

	children := feed.ChildStops(stopID)
	if len(children) > 0 {
		fmt.Printf("  child stops:    %d\n", len(children))
		for _, c := range children {
			if len(children) <= 10 {
				fmt.Printf("    - %s (%s)\n", c.StopID, c.StopName)
			}
		}
	}

	stopTimes := feed.StopTimesAtStop(stopID)
	fmt.Printf("  stop_times:     %d trips serve this stop\n", len(stopTimes))
	if len(stopTimes) > 0 && len(stopTimes) <= 5 {
		for _, st := range stopTimes {
			fmt.Printf("    trip=%s arr=%s dep=%s seq=%d\n",
				st.TripID, st.ArrivalTime, st.DepartureTime, st.StopSequence)
		}
	}
}

func queryRouteByName(feed *gtfs.Feed, shortName string) {
	fmt.Printf("\n--- Route: %s ---\n", shortName)
	routes := feed.RoutesByShortName(shortName)
	if len(routes) == 0 {
		fmt.Println("  not found")
		return
	}
	fmt.Printf("  %d route(s) found:\n", len(routes))
	for _, r := range routes {
		trips := feed.TripsForRoute(r.RouteID)
		fmt.Printf("    route_id=%s  long_name=%s  type=%d  trips=%d\n",
			r.RouteID, r.RouteLongName, r.RouteType, len(trips))
	}
}

func queryTripByID(feed *gtfs.Feed, tripID string) {
	fmt.Printf("\n--- Trip: %s ---\n", tripID)
	t := feed.Trip(tripID)
	if t == nil {
		fmt.Println("  not found")
		return
	}
	fmt.Printf("  route_id:    %s\n", t.RouteID)
	fmt.Printf("  service_id:  %s\n", t.ServiceID)
	fmt.Printf("  headsign:    %s\n", t.TripHeadsign)
	fmt.Printf("  direction:   %d\n", t.DirectionID)

	r := feed.Route(t.RouteID)
	if r != nil {
		fmt.Printf("  route_name:  %s (%s)\n", r.RouteShortName, r.RouteLongName)
	}

	stopTimes := feed.StopTimesForTrip(tripID)
	fmt.Printf("  stop_times:  %d stops\n", len(stopTimes))
	for i, st := range stopTimes {
		s := feed.Stop(st.StopID)
		name := ""
		if s != nil {
			name = s.StopName
		}
		if i < 3 || i >= len(stopTimes)-2 {
			fmt.Printf("    %2d. %s dep=%s  %s (%s)\n",
				st.StopSequence, st.ArrivalTime, st.DepartureTime, st.StopID, name)
		} else if i == 3 {
			fmt.Printf("    ... (%d more stops) ...\n", len(stopTimes)-5)
		}
	}
}

func queryActiveOnDate(feed *gtfs.Feed, dateStr string) {
	date := gtfs.ParseDate(dateStr)
	if !date.IsSet() {
		fmt.Printf("invalid date: %s\n", dateStr)
		return
	}
	fmt.Printf("\n--- Active on %s (%s) ---\n", date, date.Weekday())

	activeServices := 0
	for _, c := range feed.Store().AllCalendar() {
		if feed.ServiceRunsOn(c.ServiceID, date) {
			activeServices++
		}
	}

	trips := feed.ActiveTripsOnDate(date)
	fmt.Printf("  active services: %d\n", activeServices)
	fmt.Printf("  active trips:    %d\n", len(trips))

	// Count trips per route_short_name
	routeCounts := map[string]int{}
	for _, t := range trips {
		r := feed.Route(t.RouteID)
		if r != nil {
			routeCounts[r.RouteShortName]++
		}
	}
	fmt.Printf("  active lines:    %d\n", len(routeCounts))
}
