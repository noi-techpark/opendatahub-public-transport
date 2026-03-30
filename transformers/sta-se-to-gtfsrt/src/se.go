package main

import (
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
)

// ConvertSE converts a SIRI SE feed to a GTFS-RT ServiceAlerts feed.
func ConvertSE(feed *siri.SEFeed, resolver *Resolver) *gtfsrt.FeedMessage {
	fm := gtfsrt.NewFeedMessage()

	situations := feed.ServiceDelivery.SituationExchangeDelivery.Situations.PtSituationElement

	for _, sit := range situations {
		alert := &gtfsrt.Alert{}

		// Active periods
		periods := siri.ParseValidityPeriods(sit.ValidityPeriod)
		for _, p := range periods {
			tr := gtfsrt.TimeRange{}
			if t, err := parseISO8601Time(p.StartTime); err == nil {
				tr.Start = t.Unix()
			}
			if t, err := parseISO8601Time(p.EndTime); err == nil {
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
