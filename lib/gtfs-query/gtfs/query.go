// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfs

// Functor query API — generic predicate and scoring-based queries.

// FindTrips returns all trips matching a predicate.
func (f *Feed) FindTrips(match func(*Trip) bool) []*Trip {
	trips := f.store.AllTrips()
	var result []*Trip
	for i := range trips {
		if match(&trips[i]) {
			result = append(result, &trips[i])
		}
	}
	return result
}

// FindStops returns all stops matching a predicate.
func (f *Feed) FindStops(match func(*Stop) bool) []*Stop {
	stops := f.store.AllStops()
	var result []*Stop
	for i := range stops {
		if match(&stops[i]) {
			result = append(result, &stops[i])
		}
	}
	return result
}

// FindRoutes returns all routes matching a predicate.
func (f *Feed) FindRoutes(match func(*Route) bool) []*Route {
	routes := f.store.AllRoutes()
	var result []*Route
	for i := range routes {
		if match(&routes[i]) {
			result = append(result, &routes[i])
		}
	}
	return result
}

// FindStopTimes returns all stop_times matching a predicate.
func (f *Feed) FindStopTimes(match func(*StopTime) bool) []*StopTime {
	stopTimes := f.store.AllStopTimes()
	var result []*StopTime
	for i := range stopTimes {
		if match(&stopTimes[i]) {
			result = append(result, &stopTimes[i])
		}
	}
	return result
}

// FindTripsWhere returns trips matching compound criteria.
func (f *Feed) FindTripsWhere(match func(trip *Trip, stopTimes []*StopTime) bool) []*Trip {
	trips := f.store.AllTrips()
	var result []*Trip
	for i := range trips {
		t := &trips[i]
		sts := f.store.StopTimesForTrip(t.TripID)
		if match(t, sts) {
			result = append(result, t)
		}
	}
	return result
}

// MatchTrip finds the best-matching trip using a scoring function.
func (f *Feed) MatchTrip(score func(*Trip, []*StopTime) float64) *Trip {
	trips := f.store.AllTrips()
	var best *Trip
	var bestScore float64
	for i := range trips {
		t := &trips[i]
		sts := f.store.StopTimesForTrip(t.TripID)
		s := score(t, sts)
		if s > bestScore {
			bestScore = s
			best = t
		}
	}
	return best
}

// MatchStop finds the best-matching stop using a scoring function.
func (f *Feed) MatchStop(score func(*Stop) float64) *Stop {
	stops := f.store.AllStops()
	var best *Stop
	var bestScore float64
	for i := range stops {
		s := &stops[i]
		sc := score(s)
		if sc > bestScore {
			bestScore = sc
			best = s
		}
	}
	return best
}

// MatchTripIn finds the best-matching trip from a pre-filtered candidate set.
func (f *Feed) MatchTripIn(candidates []*Trip, score func(*Trip, []*StopTime) float64) *Trip {
	var best *Trip
	var bestScore float64
	for _, t := range candidates {
		sts := f.store.StopTimesForTrip(t.TripID)
		s := score(t, sts)
		if s > bestScore {
			bestScore = s
			best = t
		}
	}
	return best
}
