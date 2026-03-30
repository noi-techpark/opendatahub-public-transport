// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfs

// Store abstracts the storage and retrieval of GTFS entities.
// Implementations decide how data is persisted and indexed.
type Store interface {
	// --- Write ---
	AddAgency(Agency)
	AddStop(Stop)
	AddRoute(Route)
	AddTrip(Trip)
	AddStopTime(StopTime)
	AddCalendarEntry(CalendarEntry)
	AddCalendarDate(CalendarDate)
	AddShape(Shape)
	AddTranslation(Translation)

	// --- Primary lookups ---
	StopByID(id string) *Stop
	RouteByID(id string) *Route
	TripByID(id string) *Trip

	// --- Secondary lookups ---
	RoutesByShortName(name string) []*Route
	StopsByName(name string) []*Stop

	// --- Relational traversals ---
	StopTimesForTrip(tripID string) []*StopTime
	StopTimesAtStop(stopID string) []*StopTime
	TripsForRoute(routeID string) []*Trip
	TripsForService(serviceID string) []*Trip
	ChildStops(parentID string) []*Stop
	CalendarForService(serviceID string) *CalendarEntry
	CalendarDatesForService(serviceID string) []*CalendarDate

	// --- Bulk access ---
	AllAgencies() []Agency
	AllStops() []Stop
	AllRoutes() []Route
	AllTrips() []Trip
	AllStopTimes() []StopTime
	AllCalendar() []CalendarEntry
	AllCalendarDates() []CalendarDate

	// BuildIndexes must be called after all Add* calls to build lookup maps.
	BuildIndexes()
}
