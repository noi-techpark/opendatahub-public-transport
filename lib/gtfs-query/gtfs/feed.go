// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfs

import "time"

// Feed is the GTFS query overlay wrapping a Store.
type Feed struct {
	store Store
}

// NewFeed creates a Feed backed by the given Store.
func NewFeed(store Store) *Feed {
	return &Feed{store: store}
}

// Store returns the underlying store.
func (f *Feed) Store() Store { return f.store }

// --- Primary lookups ---

func (f *Feed) Stop(stopID string) *Stop    { return f.store.StopByID(stopID) }
func (f *Feed) Route(routeID string) *Route  { return f.store.RouteByID(routeID) }
func (f *Feed) Trip(tripID string) *Trip     { return f.store.TripByID(tripID) }

// --- Secondary lookups ---

func (f *Feed) RoutesByShortName(shortName string) []*Route { return f.store.RoutesByShortName(shortName) }
func (f *Feed) StopsByName(name string) []*Stop             { return f.store.StopsByName(name) }

// --- Relational traversals ---

func (f *Feed) StopTimesForTrip(tripID string) []*StopTime       { return f.store.StopTimesForTrip(tripID) }
func (f *Feed) StopTimesAtStop(stopID string) []*StopTime         { return f.store.StopTimesAtStop(stopID) }
func (f *Feed) TripsForRoute(routeID string) []*Trip             { return f.store.TripsForRoute(routeID) }
func (f *Feed) TripsForService(serviceID string) []*Trip         { return f.store.TripsForService(serviceID) }
func (f *Feed) ChildStops(parentStopID string) []*Stop           { return f.store.ChildStops(parentStopID) }
func (f *Feed) CalendarForService(serviceID string) *CalendarEntry { return f.store.CalendarForService(serviceID) }
func (f *Feed) CalendarDatesForService(serviceID string) []*CalendarDate { return f.store.CalendarDatesForService(serviceID) }

// --- Calendar queries ---

// ServiceRunsOn checks whether a service_id is active on the given date.
func (f *Feed) ServiceRunsOn(serviceID string, date Date) bool {
	if exceptions := f.store.CalendarDatesForService(serviceID); len(exceptions) > 0 {
		for _, ex := range exceptions {
			if ex.Date == date {
				return ex.ExceptionType == 1
			}
		}
	}

	cal := f.store.CalendarForService(serviceID)
	if cal == nil {
		return false
	}
	if date < cal.StartDate || date > cal.EndDate {
		return false
	}

	switch date.Weekday() {
	case time.Monday:
		return cal.Monday
	case time.Tuesday:
		return cal.Tuesday
	case time.Wednesday:
		return cal.Wednesday
	case time.Thursday:
		return cal.Thursday
	case time.Friday:
		return cal.Friday
	case time.Saturday:
		return cal.Saturday
	case time.Sunday:
		return cal.Sunday
	}
	return false
}

// ActiveTripsOnDate returns all trips whose service is active on the given date.
func (f *Feed) ActiveTripsOnDate(date Date) []*Trip {
	trips := f.store.AllTrips()
	var result []*Trip
	for i := range trips {
		if f.ServiceRunsOn(trips[i].ServiceID, date) {
			result = append(result, &trips[i])
		}
	}
	return result
}

// FindTripsForLine returns all trips for routes with the given short name,
// filtered to those active on the given date.
func (f *Feed) FindTripsForLine(shortName string, date Date) []*Trip {
	routes := f.store.RoutesByShortName(shortName)
	var result []*Trip
	for _, r := range routes {
		for _, t := range f.store.TripsForRoute(r.RouteID) {
			if f.ServiceRunsOn(t.ServiceID, date) {
				result = append(result, t)
			}
		}
	}
	return result
}
