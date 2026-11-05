// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

// SIRI-Lite Situation Exchange → GTFS-RT ServiceAlerts conversion.
//
// Mapping strategy:
//   - Alert ID:       SIRI SituationNumber → GTFS-RT entity ID.
//   - Active periods: SIRI ValidityPeriod (polymorphic array/object) → TimeRange.
//   - Cause:          SIRI AlertCause → GTFS-RT Cause enum
//                     (constructionWork→CONSTRUCTION, strike→STRIKE, etc.).
//   - Effect:         SIRI Consequence.Condition → GTFS-RT Effect enum
//                     (lineCancellation→NO_SERVICE, delayed→SIGNIFICANT_DELAYS, etc.).
//   - Severity:       SIRI Consequence.Severity → GTFS-RT SeverityLevel.
//   - Affected stops: SIRI AffectedStopPoint.StopPointRef → GTFS stop_id
//                     (strip NeTEx prefix). Only emitted if stop exists in GTFS.
//   - Affected lines: SIRI AffectedLine.LineRef → all matching GTFS route_ids
//                     via route_short_name lookup. Fallback to PublishedLineName.
//   - Header text:    SIRI ReasonName (multilingual, polymorphic) → TranslatedString.
//
// Drop decisions:
//   - Affected stops not found in GTFS are silently omitted from InformedEntity.
//   - Affected lines that can't be resolved to any GTFS route are omitted.
//   - Alerts with no InformedEntity are still emitted (the alert itself is valid;
//     the affected entities may just not be in our GTFS dataset).
//
// No trip resolution needed — SE alerts don't reference specific trips.

import (
	"strings"

	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
	"github.com/noi-techpark/opendatahub-public-transport/lib/sta-resolver/sta"
)

// ConvertSE converts a SIRI SE feed to a GTFS-RT ServiceAlerts feed.
func ConvertSE(feed *siri.SEFeed, resolver *sta.Resolver) *gtfsrt.FeedMessage {
	fm := gtfsrt.NewFeedMessage()

	situations := feed.ServiceDelivery.SituationExchangeDelivery.Situations.PtSituationElement

	for _, sit := range situations {
		alert := &gtfsrt.Alert{}

		// Active periods
		periods := siri.ParseValidityPeriods(sit.ValidityPeriod)
		for _, p := range periods {
			tr := gtfsrt.TimeRange{}
			if t, err := sta.ParseISO8601Time(p.StartTime); err == nil {
				tr.Start = t.Unix()
			}
			if t, err := sta.ParseISO8601Time(p.EndTime); err == nil {
				tr.End = t.Unix()
			}
			alert.ActivePeriod = append(alert.ActivePeriod, tr)
		}

		// Cause
		alert.Cause = mapCause(sit.AlertCause)

		// Effect and severity from consequences
		consequences := siri.ParseConsequences(sit.Consequences.Consequence)
		if len(consequences) > 0 {
			alert.Effect = mapEffect(consequences[0].Condition)
			alert.SeverityLevel = mapSeverity(consequences[0].Severity)
		}

		// Informed entities — stops (only if they exist in GTFS)
		if sit.Affects.StopPoints != nil {
			for _, asp := range sit.Affects.StopPoints.AffectedStopPoint {
				stopID := resolver.ResolveStopID(asp.StopPointRef)
				if resolver.GTFS.Stop(stopID) != nil {
					alert.InformedEntity = append(alert.InformedEntity, gtfsrt.EntitySelector{
						StopID: stopID,
					})
				}
			}
		}

		// Informed entities — lines
		if sit.Affects.Networks != nil {
			networks := siri.ParseAffectedNetworks(sit.Affects.Networks.AffectedNetwork)
			for _, net := range networks {
				lines := siri.ParseAffectedLines(net.AffectedLine)
				for _, line := range lines {
					routeIDs := resolver.ResolveAllRouteIDs(line.LineRef)
					if len(routeIDs) == 0 {
						// Fallback: try PublishedLineName as route_short_name
						routeIDs = resolver.ResolveAllRouteIDs(line.PublishedLineName)
					}
					for _, rid := range routeIDs {
						alert.InformedEntity = append(alert.InformedEntity, gtfsrt.EntitySelector{
							RouteID: rid,
						})
					}
				}
			}
		}

		// Header text from ReasonName (multilingual)
		reasonTexts := siri.ParseReasonNames(sit.ReasonName)
		if len(reasonTexts) > 0 {
			ts := &gtfsrt.TranslatedString{}
			for _, rt := range reasonTexts {
				ts.Translation = append(ts.Translation, gtfsrt.Translation{
					Text:     rt.Text,
					Language: mapLangCode(rt.Lang),
				})
			}
			alert.HeaderText = ts
		}

		fm.AddEntity(gtfsrt.FeedEntity{
			ID:    sit.SituationNumber,
			Alert: alert,
		})
	}

	return fm
}

// mapCause maps SIRI AlertCause to GTFS-RT Cause enum string.
func mapCause(siriCause string) string {
	switch siriCause {
	case "constructionWork":
		return "CONSTRUCTION"
	case "maintenanceWork":
		return "MAINTENANCE"
	case "specialEvent":
		return "OTHER_CAUSE"
	case "accident":
		return "ACCIDENT"
	case "strike":
		return "STRIKE"
	case "weather":
		return "WEATHER"
	case "technicalProblem":
		return "TECHNICAL_PROBLEM"
	default:
		return "UNKNOWN_CAUSE"
	}
}

// mapEffect maps SIRI Consequence.Condition to GTFS-RT Effect enum string.
func mapEffect(condition string) string {
	switch condition {
	case "stopCancelled":
		return "NO_SERVICE"
	case "lineCancellation":
		return "NO_SERVICE"
	case "delayed":
		return "SIGNIFICANT_DELAYS"
	case "diverted":
		return "DETOUR"
	case "reducedService":
		return "REDUCED_SERVICE"
	case "additionalService":
		return "ADDITIONAL_SERVICE"
	case "modifiedService":
		return "MODIFIED_SERVICE"
	default:
		return "UNKNOWN_EFFECT"
	}
}

// mapSeverity maps SIRI Severity to GTFS-RT SeverityLevel enum string.
func mapSeverity(severity string) string {
	switch severity {
	case "severe":
		return "SEVERE"
	case "warning":
		return "WARNING"
	case "normal":
		return "INFO"
	default:
		return "UNKNOWN_SEVERITY"
	}
}

// mapLangCode maps SIRI language codes to ISO 639-1.
func mapLangCode(lang string) string {
	switch strings.ToUpper(lang) {
	case "DE":
		return "de"
	case "IT":
		return "it"
	case "EN":
		return "en"
	case "LAD":
		return "lld" // Ladin (ISO 639-3, but commonly used)
	default:
		return strings.ToLower(lang)
	}
}
