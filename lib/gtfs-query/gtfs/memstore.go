// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfs

import (
	"slices"
	"strings"
)

// MemStore is an in-memory Store implementation using Go maps and slices.
type MemStore struct {
	agencies      []Agency
	stops         []Stop
	routes        []Route
	trips         []Trip
	stopTimes     []StopTime
	calendar      []CalendarEntry
	calendarDates []CalendarDate
	shapes        []Shape
	translations  []Translation

	// indexes
	stopByID          map[string]*Stop
	routeByID         map[string]*Route
	tripByID          map[string]*Trip
	routesByShortName map[string][]*Route
	tripsByRouteID    map[string][]*Trip
	tripsByServiceID  map[string][]*Trip
	stopTimesByTripID map[string][]StopTime
	stopTimesByStopID map[string][]StopTime
	childStops        map[string][]*Stop
	calendarByService map[string]*CalendarEntry
	calDatesByService map[string][]CalendarDate
	stopsByName       map[string][]*Stop
}

// NewMemStore creates a new in-memory store with pre-allocated slices.
func NewMemStore() *MemStore {
	return &MemStore{
		stops:     make([]Stop, 0, 8000),
		routes:    make([]Route, 0, 1000),
		trips:     make([]Trip, 0, 40000),
		stopTimes: make([]StopTime, 0, 700_000),
		calendar:  make([]CalendarEntry, 0, 2000),
	}
}

// --- Add methods ---

func (m *MemStore) AddAgency(a Agency)        { m.agencies = append(m.agencies, a) }
func (m *MemStore) AddStop(s Stop)             { m.stops = append(m.stops, s) }
func (m *MemStore) AddRoute(r Route)           { m.routes = append(m.routes, r) }
func (m *MemStore) AddTrip(t Trip)             { m.trips = append(m.trips, t) }
func (m *MemStore) AddStopTime(st StopTime)    { m.stopTimes = append(m.stopTimes, st) }
func (m *MemStore) AddCalendarEntry(c CalendarEntry) { m.calendar = append(m.calendar, c) }
func (m *MemStore) AddCalendarDate(cd CalendarDate)  { m.calendarDates = append(m.calendarDates, cd) }
func (m *MemStore) AddShape(s Shape)           { m.shapes = append(m.shapes, s) }
func (m *MemStore) AddTranslation(t Translation) { m.translations = append(m.translations, t) }

// --- Primary lookups ---

func (m *MemStore) StopByID(id string) *Stop    { return m.stopByID[id] }
func (m *MemStore) RouteByID(id string) *Route   { return m.routeByID[id] }
func (m *MemStore) TripByID(id string) *Trip     { return m.tripByID[id] }

// --- Secondary lookups ---

func (m *MemStore) RoutesByShortName(name string) []*Route { return m.routesByShortName[name] }
func (m *MemStore) StopsByName(name string) []*Stop        { return m.stopsByName[strings.ToLower(name)] }

// --- Relational traversals ---

func (m *MemStore) StopTimesForTrip(tripID string) []StopTime       { return m.stopTimesByTripID[tripID] }
func (m *MemStore) StopTimesAtStop(stopID string) []StopTime         { return m.stopTimesByStopID[stopID] }
func (m *MemStore) TripsForRoute(routeID string) []*Trip             { return m.tripsByRouteID[routeID] }
func (m *MemStore) TripsForService(serviceID string) []*Trip         { return m.tripsByServiceID[serviceID] }
func (m *MemStore) ChildStops(parentID string) []*Stop               { return m.childStops[parentID] }
func (m *MemStore) CalendarForService(serviceID string) *CalendarEntry { return m.calendarByService[serviceID] }
func (m *MemStore) CalendarDatesForService(serviceID string) []CalendarDate { return m.calDatesByService[serviceID] }

// --- Bulk access ---

func (m *MemStore) AllAgencies() []Agency          { return m.agencies }
func (m *MemStore) AllStops() []Stop               { return m.stops }
func (m *MemStore) AllRoutes() []Route             { return m.routes }
func (m *MemStore) AllTrips() []Trip               { return m.trips }
func (m *MemStore) AllStopTimes() []StopTime       { return m.stopTimes }
func (m *MemStore) AllCalendar() []CalendarEntry    { return m.calendar }
func (m *MemStore) AllCalendarDates() []CalendarDate { return m.calendarDates }

// BuildIndexes constructs all lookup maps from the loaded data.
func (m *MemStore) BuildIndexes() {
	// stops
	m.stopByID = make(map[string]*Stop, len(m.stops))
	m.childStops = make(map[string][]*Stop)
	m.stopsByName = make(map[string][]*Stop)
	for i := range m.stops {
		s := &m.stops[i]
		m.stopByID[s.StopID] = s
		if s.ParentStation != "" {
			m.childStops[s.ParentStation] = append(m.childStops[s.ParentStation], s)
		}
		key := strings.ToLower(s.StopName)
		m.stopsByName[key] = append(m.stopsByName[key], s)
	}

	// routes
	m.routeByID = make(map[string]*Route, len(m.routes))
	m.routesByShortName = make(map[string][]*Route)
	for i := range m.routes {
		r := &m.routes[i]
		m.routeByID[r.RouteID] = r
		if r.RouteShortName != "" {
			m.routesByShortName[r.RouteShortName] = append(m.routesByShortName[r.RouteShortName], r)
		}
	}

	// trips
	m.tripByID = make(map[string]*Trip, len(m.trips))
	m.tripsByRouteID = make(map[string][]*Trip)
	m.tripsByServiceID = make(map[string][]*Trip)
	for i := range m.trips {
		t := &m.trips[i]
		m.tripByID[t.TripID] = t
		m.tripsByRouteID[t.RouteID] = append(m.tripsByRouteID[t.RouteID], t)
		m.tripsByServiceID[t.ServiceID] = append(m.tripsByServiceID[t.ServiceID], t)
	}

	// stop_times
	m.stopTimesByTripID = make(map[string][]StopTime, len(m.trips))
	m.stopTimesByStopID = make(map[string][]StopTime)
	for _, st := range m.stopTimes {
		m.stopTimesByTripID[st.TripID] = append(m.stopTimesByTripID[st.TripID], st)
		m.stopTimesByStopID[st.StopID] = append(m.stopTimesByStopID[st.StopID], st)
	}
	for tripID, sts := range m.stopTimesByTripID {
		slices.SortFunc(sts, func(a, b StopTime) int {
			return a.StopSequence - b.StopSequence
		})
		m.stopTimesByTripID[tripID] = sts
	}

	// calendar
	m.calendarByService = make(map[string]*CalendarEntry, len(m.calendar))
	for i := range m.calendar {
		c := &m.calendar[i]
		m.calendarByService[c.ServiceID] = c
	}

	// calendar_dates
	m.calDatesByService = make(map[string][]CalendarDate)
	for _, cd := range m.calendarDates {
		m.calDatesByService[cd.ServiceID] = append(m.calDatesByService[cd.ServiceID], cd)
	}
}
