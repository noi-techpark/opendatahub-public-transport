package main

import (
	"log"
	"math"
	"strings"

	"github.com/noi-techpark/opendatahub-public-transport/lib/gtfs-query/gtfs"
	netexq "github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
)

// Resolver maps SIRI entity references to GTFS entities
// using the GTFS feed and NeTEx query feed for ID resolution.
type Resolver struct {
	GTFS  *gtfs.Feed
	NeTEx *netexq.NeTExFeed

	// Stats
	TripsResolvedA   int // Phase A: NeTEx SJ ID
	TripsResolvedB   int // Phase B: stop+time
	TripsUnresolved  int
	RoutesResolved   int
	RoutesUnresolved int
}

// ResolveRouteID maps a SIRI LineRef to the first matching GTFS route_id.
// VM LineRef = public_code (e.g., "240") → GTFS RoutesByShortName
// SE LineRef = NeTEx ID (e.g., "it:apb:Line:222401.26a") → NeTEx public_code → GTFS
func (r *Resolver) ResolveRouteID(lineRef string) string {
	publicCode := r.resolveToPublicCode(lineRef)
	if publicCode == "" {
		r.RoutesUnresolved++
		return ""
	}
	routes := r.GTFS.RoutesByShortName(publicCode)
	if len(routes) == 0 {
		r.RoutesUnresolved++
		return ""
	}
	r.RoutesResolved++
	return routes[0].RouteID
}

// ResolveAllRouteIDs returns all GTFS route_ids matching a SIRI LineRef.
func (r *Resolver) ResolveAllRouteIDs(lineRef string) []string {
	publicCode := r.resolveToPublicCode(lineRef)
	if publicCode == "" {
		return nil
	}
	routes := r.GTFS.RoutesByShortName(publicCode)
	ids := make([]string, len(routes))
	for i, rt := range routes {
		ids[i] = rt.RouteID
	}
	return ids
}

func (r *Resolver) resolveToPublicCode(lineRef string) string {
	// If it contains ":" it's likely a NeTEx ID → look up public_code
	if strings.Contains(lineRef, ":") {
		pc := r.NeTEx.LinePublicCode(lineRef)
		if pc != "" {
			return pc
		}
		// Try stripping trailing colon
		pc = r.NeTEx.LinePublicCode(strings.TrimSuffix(lineRef, ":"))
		if pc != "" {
			return pc
		}
		return ""
	}
	// Already a public_code
	return lineRef
}

// ResolveStopID maps a SIRI StopPointRef to a GTFS stop_id.
func (r *Resolver) ResolveStopID(stopPointRef string) string {
	return stripNeTExStopPrefix(stopPointRef)
}

// ResolveTripID attempts to match a SIRI vehicle journey to a GTFS trip_id.
// Uses two strategies:
//   Phase A: Extract NeTEx ServiceJourney ID from ref → traverse NeTEx → match GTFS
//   Phase B: Use current stop + time to match directly in GTFS
func (r *Resolver) ResolveTripID(
	lineRef, dirRef, dataFrameRef, datedVehicleJourneyRef string,
	stopPointRef, recordedAtTime, delay string,
) string {
	dateStr := reformatDate(dataFrameRef)
	date := gtfs.ParseDate(dateStr)
	if !date.IsSet() {
		r.TripsUnresolved++
		return ""
	}

	// Phase A: NeTEx ServiceJourney ID extraction
	if sjID := extractNeTExSJID(datedVehicleJourneyRef); sjID != "" {
		if tripID := r.matchViaNeTExSJ(sjID, date, dirRef); tripID != "" {
			r.TripsResolvedA++
			return tripID
		}
	}

	// Phase B: Stop + time matching
	if stopPointRef != "" && recordedAtTime != "" {
		if tripID := r.matchViaStopTime(lineRef, dirRef, stopPointRef, recordedAtTime, delay, date); tripID != "" {
			r.TripsResolvedB++
			return tripID
		}
	}

	r.TripsUnresolved++
	return ""
}

// ResolveDirectionID maps SIRI DirectionRef to GTFS direction_id.
func (r *Resolver) ResolveDirectionID(dirRef string) int {
	return mapDirectionRef(dirRef)
}

// --- Phase A: NeTEx ServiceJourney-based matching ---

// extractNeTExSJID extracts a NeTEx ServiceJourney ID from a DatedVehicleJourneyRef.
// Returns "" if no SJ ID is embedded.
// SIRI embeds: "it:apb:ServiceJourney:88464-KronplM-71-3-33600:30" (with version suffix)
// CSV has:     "it:apb:ServiceJourney:88464-KronplM-71-3-33600:T1-464A:" (with day_type suffix)
// We extract the base part before the last colon-delimited suffix.
func extractNeTExSJID(ref string) string {
	const marker = "it:apb:ServiceJourney:"
	idx := strings.Index(ref, marker)
	if idx < 0 {
		return ""
	}
	sjPart := strings.TrimSuffix(ref[idx:], "_")

	// The base ID is everything up to the last ":" segment after the journey details
	// e.g., "it:apb:ServiceJourney:88464-KronplM-71-3-33600:30" → base = "88464-KronplM-71-3-33600"
	inner := sjPart[len(marker):] // "88464-KronplM-71-3-33600:30"
	// Strip the version/variant suffix (last ":" segment)
	if lastColon := strings.LastIndex(inner, ":"); lastColon > 0 {
		inner = inner[:lastColon] // "88464-KronplM-71-3-33600"
	}
	return inner // just the base, without the full prefix (resolver will search)
}

// matchViaNeTExSJ resolves via NeTEx: SJ → JP → Route → Line → GTFS trip.
// sjBase is the base part of the ServiceJourney ID (without day_type suffix).
func (r *Resolver) matchViaNeTExSJ(sjBase string, date gtfs.Date, siriDirRef string) string {
	// Search for a ServiceJourney whose ID contains this base
	fullID := r.NeTEx.FindServiceJourneyByBase(sjBase)
	if fullID == "" {
		return ""
	}
	info, ok := r.NeTEx.ResolveSJToLineInfo(fullID)
	if !ok {
		return ""
	}

	// Find GTFS trips for this line on this date
	candidates := r.GTFS.FindTripsForLine(info.PublicCode, date)
	if len(candidates) == 0 {
		return ""
	}

	// Filter by direction
	gtfsDir := netexDirToGTFS(info.Direction)
	var dirFiltered []*gtfs.Trip
	for _, t := range candidates {
		if t.DirectionID == gtfsDir {
			dirFiltered = append(dirFiltered, t)
		}
	}
	if len(dirFiltered) == 0 {
		dirFiltered = candidates // fallback: don't filter if no match
	}

	// Parse NeTEx departure time
	depTime := gtfs.ParseGTFSTime(info.DepartureTime)
	if !depTime.IsSet() {
		return ""
	}

	// Score by first departure time proximity
	best := r.GTFS.MatchTripIn(dirFiltered, func(trip *gtfs.Trip, sts []gtfs.StopTime) float64 {
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
		diff := math.Abs(float64(firstDep.Seconds() - depTime.Seconds()))
		if diff > 120 { // >2 min tolerance for first stop
			return 0
		}
		return 1.0 / (1.0 + diff)
	})

	if best != nil {
		return best.TripID
	}
	return ""
}

// --- Phase B: Stop + time matching ---

// matchViaStopTime finds GTFS trip by matching the scheduled time at the current stop.
func (r *Resolver) matchViaStopTime(
	lineRef, dirRef, stopPointRef, recordedAtTime, delayStr string,
	date gtfs.Date,
) string {
	stopID := stripNeTExStopPrefix(stopPointRef)

	// Compute scheduled time at this stop = recorded time - delay
	recordedTime, err := parseISO8601Time(recordedAtTime)
	if err != nil {
		return ""
	}
	delaySecs := parseISO8601Duration(delayStr)
	scheduledTime := recordedTime.Add(-delaySecs)

	// Convert to GTFS time (seconds since midnight)
	scheduledGTFS := gtfs.GTFSTime(scheduledTime.Hour()*3600 + scheduledTime.Minute()*60 + scheduledTime.Second())

	// Get all stop_times at this stop
	stopTimes := r.GTFS.StopTimesAtStop(stopID)
	if len(stopTimes) == 0 {
		return ""
	}

	// Resolve line to public code for filtering
	publicCode := r.resolveToPublicCode(lineRef)
	gtfsDir := mapDirectionRef(dirRef)

	var bestTrip string
	var bestDiff float64 = 999999

	for _, st := range stopTimes {
		stTime := st.ArrivalTime
		if !stTime.IsSet() {
			stTime = st.DepartureTime
		}
		if !stTime.IsSet() {
			continue
		}

		diff := math.Abs(float64(stTime.Seconds() - scheduledGTFS.Seconds()))
		if diff > 600 { // 10 min tolerance (vehicles can be significantly delayed)
			continue
		}

		trip := r.GTFS.Trip(st.TripID)
		if trip == nil {
			continue
		}

		// Service active on date
		if !r.GTFS.ServiceRunsOn(trip.ServiceID, date) {
			continue
		}

		// Line check (route_short_name)
		if publicCode != "" {
			route := r.GTFS.Route(trip.RouteID)
			if route == nil || route.RouteShortName != publicCode {
				continue
			}
		}

		// Direction match is a bonus (tiebreaker), not a filter
		score := 1.0 / (1.0 + diff)
		if trip.DirectionID == gtfsDir {
			score *= 1.5 // prefer matching direction
		}

		if score > 1.0/(1.0+bestDiff) || (diff < bestDiff) {
			bestDiff = diff
			bestTrip = trip.TripID
		}
	}

	return bestTrip
}

// netexDirToGTFS maps NeTEx direction "H"/"R" to GTFS direction_id 0/1.
func netexDirToGTFS(dir string) int {
	switch dir {
	case "H":
		return 0
	case "R":
		return 1
	default:
		return 0
	}
}

// PrintStats logs resolver statistics.
func (r *Resolver) PrintStats() {
	totalTrips := r.TripsResolvedA + r.TripsResolvedB + r.TripsUnresolved
	if totalTrips > 0 {
		log.Printf("Trip resolution: %d/%d (%.1f%%) — Phase A (NeTEx SJ): %d, Phase B (stop+time): %d, unresolved: %d",
			r.TripsResolvedA+r.TripsResolvedB, totalTrips,
			float64(r.TripsResolvedA+r.TripsResolvedB)/float64(totalTrips)*100,
			r.TripsResolvedA, r.TripsResolvedB, r.TripsUnresolved)
	}
	totalRoutes := r.RoutesResolved + r.RoutesUnresolved
	if totalRoutes > 0 {
		log.Printf("Route resolution: %d/%d (%.1f%%)",
			r.RoutesResolved, totalRoutes, float64(r.RoutesResolved)/float64(totalRoutes)*100)
	}
}
