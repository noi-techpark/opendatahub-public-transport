// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import "strings"

// NeTExFeed is the query overlay over a generic NeTEx Store.
// It provides typed access to the entity types needed for SIRI→GTFS-RT mapping,
// while the underlying store holds ALL entity types from any profile.
type NeTExFeed struct {
	store Store

	// typed indexes built from the store for fast querying
	sjByID             map[string]*ServiceJourney
	sjBaseIndex        map[string]*ServiceJourney
	jpByID             map[string]*JourneyPattern
	routeByID          map[string]*Route
	lineByID           map[string]*Line
	lineIDToPublicCode map[string]string
}

// NewNeTExFeed creates a query feed from a store. Call BuildIndexes() after the store is populated.
func NewNeTExFeed(store Store) *NeTExFeed {
	return &NeTExFeed{store: store}
}

// Store returns the underlying generic store.
func (f *NeTExFeed) Store() Store { return f.store }

// BuildIndexes reads typed entities from the store and builds query indexes.
func (f *NeTExFeed) BuildIndexes() {
	// ServiceJourneys
	sjEntities := f.store.All("ServiceJourney")
	f.sjByID = make(map[string]*ServiceJourney, len(sjEntities))
	f.sjBaseIndex = make(map[string]*ServiceJourney, len(sjEntities))
	for _, e := range sjEntities {
		sj := &ServiceJourney{
			ID:                e.Fields["id"],
			DepartureTime:     e.Fields["departure_time"],
			DayTypeRef:        e.Fields["day_type_ref"],
			JourneyPatternRef: e.Fields["journey_pattern_ref"],
			OperatorRef:       e.Fields["operator_ref"],
			TransportMode:     e.Fields["transport_mode"],
		}
		f.sjByID[sj.ID] = sj
		f.sjByID[strings.TrimSuffix(sj.ID, ":")] = sj
		if base := extractSJBase(sj.ID); base != "" {
			f.sjBaseIndex[base] = sj
		}
	}

	// JourneyPatterns
	jpEntities := f.store.All("ServiceJourneyPattern")
	f.jpByID = make(map[string]*JourneyPattern, len(jpEntities))
	for _, e := range jpEntities {
		jp := &JourneyPattern{
			ID:       e.Fields["id"],
			RouteRef: e.Fields["route_ref"],
		}
		f.jpByID[jp.ID] = jp
		f.jpByID[strings.TrimSuffix(jp.ID, ":")] = jp
	}

	// Routes
	routeEntities := f.store.All("Route")
	f.routeByID = make(map[string]*Route, len(routeEntities))
	for _, e := range routeEntities {
		r := &Route{
			ID:           e.Fields["id"],
			LineRef:      e.Fields["line_ref"],
			DirectionRef: e.Fields["direction_ref"],
		}
		f.routeByID[r.ID] = r
		f.routeByID[strings.TrimSuffix(r.ID, ":")] = r
	}

	// Lines
	lineEntities := f.store.All("Line")
	f.lineByID = make(map[string]*Line, len(lineEntities))
	f.lineIDToPublicCode = make(map[string]string, len(lineEntities))
	for _, e := range lineEntities {
		l := &Line{
			ID:         e.Fields["id"],
			PublicCode: e.Fields["public_code"],
		}
		f.lineByID[l.ID] = l
		f.lineByID[strings.TrimSuffix(l.ID, ":")] = l
		if l.PublicCode != "" {
			f.lineIDToPublicCode[l.ID] = l.PublicCode
			f.lineIDToPublicCode[strings.TrimSuffix(l.ID, ":")] = l.PublicCode
		}
	}
}

// --- Query methods (unchanged API) ---

func (f *NeTExFeed) ServiceJourneyByID(id string) *ServiceJourney {
	return f.sjByID[id]
}

func (f *NeTExFeed) LinePublicCode(lineID string) string {
	return f.lineIDToPublicCode[lineID]
}

func (f *NeTExFeed) ResolveSJToLineInfo(sjID string) (LineInfo, bool) {
	sj := f.sjByID[sjID]
	if sj == nil {
		sj = f.sjByID[strings.TrimSuffix(sjID, ":")]
		if sj == nil {
			return LineInfo{}, false
		}
	}

	jp := f.jpByID[sj.JourneyPatternRef]
	if jp == nil {
		return LineInfo{}, false
	}

	route := f.routeByID[jp.RouteRef]
	if route == nil {
		return LineInfo{}, false
	}

	publicCode := f.lineIDToPublicCode[route.LineRef]
	if publicCode == "" {
		return LineInfo{}, false
	}

	direction := extractDirection(route.DirectionRef)

	return LineInfo{
		PublicCode:    publicCode,
		Direction:     direction,
		DepartureTime: sj.DepartureTime,
		DayTypeRef:    sj.DayTypeRef,
	}, true
}

func (f *NeTExFeed) FindServiceJourneyByBase(base string) string {
	if sj, ok := f.sjBaseIndex[base]; ok {
		return sj.ID
	}
	return ""
}

func (f *NeTExFeed) Stats() (sj, jp, routes, lines int) {
	return f.store.Count("ServiceJourney"),
		f.store.Count("ServiceJourneyPattern"),
		f.store.Count("Route"),
		f.store.Count("Line")
}

// --- helpers ---

func extractSJBase(sjID string) string {
	const prefix = "it:apb:ServiceJourney:"
	if !strings.HasPrefix(sjID, prefix) {
		return ""
	}
	inner := sjID[len(prefix):]
	inner = strings.TrimSuffix(inner, ":")
	if lastColon := strings.LastIndex(inner, ":"); lastColon > 0 {
		return inner[:lastColon]
	}
	return inner
}

func extractDirection(ref string) string {
	ref = strings.TrimSuffix(ref, ":")
	if idx := strings.LastIndex(ref, ":"); idx >= 0 {
		return ref[idx+1:]
	}
	return ref
}
