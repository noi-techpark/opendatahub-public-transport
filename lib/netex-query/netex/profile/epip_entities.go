// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package profile

import (
	"encoding/xml"

	"github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
)

// --- ResourceFrame entities ---

func parseAuthority(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	return parseOrganisation(decoder, start, "Authority", emit)
}

func parseOperator(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	return parseOrganisation(decoder, start, "Operator", emit)
}

func parseOrganisation(decoder *xml.Decoder, start xml.StartElement, entityType string, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name_lang"] = getAttr(t, "lang")
				fields["name"] = readText(decoder)
			case "ShortName":
				fields["short_name"] = readText(decoder)
			case "LegalName":
				fields["legal_name"] = readText(decoder)
			case "PublicCode":
				fields["public_code"] = readText(decoder)
			case "OrganisationType":
				fields["organisation_type"] = readText(decoder)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			case "keyList":
				parseKeyList(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				emit(netex.Entity{Type: entityType, Fields: fields})
				return nil
			}
		}
	}
}

func parseVehicleType(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "LowFloor":
				fields["low_floor"] = readText(decoder)
			case "HasLiftOrRamp":
				fields["has_lift_or_ramp"] = readText(decoder)
			case "HasHoist":
				fields["has_hoist"] = readText(decoder)
			case "Length":
				fields["length"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "VehicleType" {
				emit(netex.Entity{Type: "VehicleType", Fields: fields})
				return nil
			}
		}
	}
}

// --- SiteFrame entities ---

func parseStopPlace(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	parentID := fields["id"]

	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name_lang"] = getAttr(t, "lang")
				fields["name"] = readText(decoder)
			case "ShortName":
				fields["short_name"] = readText(decoder)
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
			case "PublicCode":
				fields["public_code"] = readText(decoder)
			case "Centroid":
				parseCentroid(decoder, fields)
			case "TopographicPlaceRef":
				fields["topographic_place_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "TransportMode":
				fields["transport_mode"] = readText(decoder)
			case "StopPlaceType":
				fields["stop_place_type"] = readText(decoder)
			case "quays":
				parseQuays(decoder, parentID, emit)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			case "keyList":
				parseKeyList(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "StopPlace" {
				emit(netex.Entity{Type: "StopPlace", Fields: fields})
				return nil
			}
		}
	}
}

func parseCentroid(decoder *xml.Decoder, fields map[string]string) {
	// Structure: Centroid > Location > {Longitude, Latitude}
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Location":
				parseLocation(decoder, fields)
			case "Longitude":
				fields["longitude"] = readText(decoder)
			case "Latitude":
				fields["latitude"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Centroid" {
				return
			}
		}
	}
}

func parseQuays(decoder *xml.Decoder, stopPlaceID string, emit netex.EntityHandler) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "Quay" {
				parseQuay(decoder, t, stopPlaceID, emit)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "quays" {
				return
			}
		}
	}
}

func parseQuay(decoder *xml.Decoder, start xml.StartElement, stopPlaceID string, emit netex.EntityHandler) {
	fields := parseGenericAttrs(start)
	fields["stop_place_id"] = stopPlaceID

	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name_lang"] = getAttr(t, "lang")
				fields["name"] = readText(decoder)
			case "ShortName":
				fields["short_name"] = readText(decoder)
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
			case "PublicCode":
				fields["public_code"] = readText(decoder)
			case "Centroid":
				parseCentroid(decoder, fields)
			case "QuayType":
				fields["quay_type"] = readText(decoder)
			case "keyList":
				parseKeyList(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Quay" {
				emit(netex.Entity{Type: "Quay", Fields: fields})
				return
			}
		}
	}
}

// --- ServiceFrame entities ---

func parseScheduledStopPoint(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name_lang"] = getAttr(t, "lang")
				fields["name"] = readText(decoder)
			case "ShortName":
				fields["short_name"] = readText(decoder)
			case "Location":
				parseLocation(decoder, fields)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ScheduledStopPoint" {
				emit(netex.Entity{Type: "ScheduledStopPoint", Fields: fields})
				return nil
			}
		}
	}
}

func parseLocation(decoder *xml.Decoder, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Longitude":
				fields["longitude"] = readText(decoder)
			case "Latitude":
				fields["latitude"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Location" {
				return
			}
		}
	}
}

func parseDirection(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "DirectionType":
				fields["direction_type"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Direction" {
				emit(netex.Entity{Type: "Direction", Fields: fields})
				return nil
			}
		}
	}
}

func parseRoute(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	routeID := fields["id"]

	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name"] = readText(decoder)
			case "Distance":
				fields["distance"] = readText(decoder)
			case "LineRef":
				fields["line_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "DirectionType":
				fields["direction_type"] = readText(decoder)
			case "DirectionRef":
				fields["direction_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "InverseRouteRef":
				fields["inverse_route_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "pointsInSequence":
				parsePointsOnRoute(decoder, routeID, emit)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Route" {
				emit(netex.Entity{Type: "Route", Fields: fields})
				return nil
			}
		}
	}
}

func parsePointsOnRoute(decoder *xml.Decoder, routeID string, emit netex.EntityHandler) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "PointOnRoute" {
				parsePointOnRoute(decoder, t, routeID, emit)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "pointsInSequence" {
				return
			}
		}
	}
}

func parsePointOnRoute(decoder *xml.Decoder, start xml.StartElement, routeID string, emit netex.EntityHandler) {
	fields := parseGenericAttrs(start)
	fields["route_id"] = routeID

	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "RoutePointRef" {
				fields["scheduled_stop_point_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "PointOnRoute" {
				emit(netex.Entity{Type: "PointOnRoute", Fields: fields})
				return
			}
		}
	}
}

func parseLine(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name_lang"] = getAttr(t, "lang")
				fields["name"] = readText(decoder)
			case "ShortName":
				fields["short_name"] = readText(decoder)
			case "Description":
				fields["description"] = readText(decoder)
			case "TransportMode":
				fields["transport_mode"] = readText(decoder)
			case "TransportSubmode":
				fields["transport_submode"] = readSubmode(decoder)
			case "Url":
				fields["url"] = readText(decoder)
			case "PublicCode":
				fields["public_code"] = readText(decoder)
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
			case "OperatorRef":
				fields["operator_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "AuthorityRef":
				fields["authority_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "Monitored":
				fields["monitored"] = readText(decoder)
			case "TypeOfLineRef":
				fields["type_of_line_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			case "keyList":
				parseKeyList(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Line" || t.Name.Local == "FlexibleLine" {
				emit(netex.Entity{Type: "Line", Fields: fields})
				return nil
			}
		}
	}
}

func parseDestinationDisplay(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "SideText":
				fields["side_text"] = readText(decoder)
			case "FrontText":
				fields["front_text"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "DestinationDisplay" {
				emit(netex.Entity{Type: "DestinationDisplay", Fields: fields})
				return nil
			}
		}
	}
}

func parseServiceJourneyPattern(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	patternID := fields["id"]

	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name"] = readText(decoder)
			case "RouteRef":
				fields["route_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "pointsInSequence":
				parseStopPointsInJP(decoder, patternID, emit)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ServiceJourneyPattern" {
				emit(netex.Entity{Type: "ServiceJourneyPattern", Fields: fields})
				return nil
			}
		}
	}
}

func parseStopPointsInJP(decoder *xml.Decoder, patternID string, emit netex.EntityHandler) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "StopPointInJourneyPattern" {
				parseStopPointInJP(decoder, t, patternID, emit)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "pointsInSequence" {
				return
			}
		}
	}
}

func parseStopPointInJP(decoder *xml.Decoder, start xml.StartElement, patternID string, emit netex.EntityHandler) {
	fields := parseGenericAttrs(start)
	fields["journey_pattern_id"] = patternID

	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "ScheduledStopPointRef":
				fields["scheduled_stop_point_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "ForAlighting":
				fields["for_alighting"] = readText(decoder)
			case "ForBoarding":
				fields["for_boarding"] = readText(decoder)
			case "DestinationDisplayRef":
				fields["destination_display_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "RequestStop":
				fields["request_stop"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "StopPointInJourneyPattern" {
				emit(netex.Entity{Type: "StopPointInJourneyPattern", Fields: fields})
				return
			}
		}
	}
}

func parseServiceLink(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Distance":
				fields["distance"] = readText(decoder)
			case "FromPointRef":
				fields["from_point_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "ToPointRef":
				fields["to_point_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			default:
				skipElement(decoder) // skip gml:LineString etc.
			}
		case xml.EndElement:
			if t.Name.Local == "ServiceLink" {
				emit(netex.Entity{Type: "ServiceLink", Fields: fields})
				return nil
			}
		}
	}
}

func parseSiteConnection(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "WalkTransferDuration":
				parseWalkTransferDuration(decoder, fields)
			case "DefaultDuration":
				fields["default_duration"] = readText(decoder)
			case "BothWays":
				fields["both_ways"] = readText(decoder)
			case "From":
				parseConnectionEnd(decoder, "from_", fields)
			case "To":
				parseConnectionEnd(decoder, "to_", fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				emit(netex.Entity{Type: "SiteConnection", Fields: fields})
				return nil
			}
		}
	}
}

func parseWalkTransferDuration(decoder *xml.Decoder, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "DefaultDuration" {
				fields["default_duration"] = readText(decoder)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "WalkTransferDuration" {
				return
			}
		}
	}
}

func parseConnectionEnd(decoder *xml.Decoder, prefix string, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "StopPlaceRef":
				fields[prefix+"stop_place_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "QuayRef":
				fields[prefix+"quay_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "From" || t.Name.Local == "To" {
				return
			}
		}
	}
}

func parseNotice(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Text":
				fields["text_lang"] = getAttr(t, "lang")
				fields["text"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Notice" {
				emit(netex.Entity{Type: "Notice", Fields: fields})
				return nil
			}
		}
	}
}

func parsePassengerStopAssignment(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "ScheduledStopPointRef":
				fields["scheduled_stop_point_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "StopPlaceRef":
				fields["stop_place_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "QuayRef":
				fields["quay_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "PassengerStopAssignment" {
				emit(netex.Entity{Type: "PassengerStopAssignment", Fields: fields})
				return nil
			}
		}
	}
}

// --- ServiceCalendarFrame entities ---

func parseDayType(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name_lang"] = getAttr(t, "lang")
				fields["name"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "DayType" {
				emit(netex.Entity{Type: "DayType", Fields: fields})
				return nil
			}
		}
	}
}

func parseUicOperatingPeriod(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "FromDate":
				fields["from_date"] = readText(decoder)
			case "ToDate":
				fields["to_date"] = readText(decoder)
			case "ValidDayBits":
				fields["valid_day_bits"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "UicOperatingPeriod" {
				emit(netex.Entity{Type: "UicOperatingPeriod", Fields: fields})
				return nil
			}
		}
	}
}

func parseDayTypeAssignment(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "OperatingPeriodRef":
				fields["operating_period_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "DayTypeRef":
				fields["day_type_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "DayTypeAssignment" {
				emit(netex.Entity{Type: "DayTypeAssignment", Fields: fields})
				return nil
			}
		}
	}
}

// --- TimetableFrame entities ---

func parseServiceJourney(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	journeyID := fields["id"]

	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Name":
				fields["name"] = readText(decoder)
			case "Distance":
				fields["distance"] = readText(decoder)
			case "TransportMode":
				fields["transport_mode"] = readText(decoder)
			case "TransportSubmode":
				fields["transport_submode"] = readSubmode(decoder)
			case "DepartureTime":
				fields["departure_time"] = readText(decoder)
			case "dayTypes":
				parseDayTypeRefs(decoder, fields)
			case "ServiceJourneyPatternRef":
				fields["journey_pattern_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "JourneyPatternRef":
				fields["journey_pattern_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "VehicleTypeRef":
				fields["vehicle_type_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "OperatorRef":
				fields["operator_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "LineRef":
				fields["line_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "trainNumbers":
				parseTrainNumberRefs(decoder, fields)
			case "passingTimes":
				parsePassingTimes(decoder, journeyID, emit)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ServiceJourney" {
				emit(netex.Entity{Type: "ServiceJourney", Fields: fields})
				return nil
			}
		}
	}
}

func parseDayTypeRefs(decoder *xml.Decoder, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "DayTypeRef" {
				fields["day_type_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "dayTypes" {
				return
			}
		}
	}
}

func parseTrainNumberRefs(decoder *xml.Decoder, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "TrainNumberRef" {
				fields["train_number_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "trainNumbers" {
				return
			}
		}
	}
}

func parsePassingTimes(decoder *xml.Decoder, journeyID string, emit netex.EntityHandler) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "TimetabledPassingTime" {
				parseTimetabledPassingTime(decoder, t, journeyID, emit)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "passingTimes" {
				return
			}
		}
	}
}

func parseTimetabledPassingTime(decoder *xml.Decoder, start xml.StartElement, journeyID string, emit netex.EntityHandler) {
	fields := parseGenericAttrs(start)
	fields["service_journey_id"] = journeyID

	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "StopPointInJourneyPatternRef":
				fields["stop_point_in_jp_ref"] = getAttr(t, "ref")
				fields["order"] = getAttr(t, "order")
				skipElement(decoder)
			case "ArrivalTime":
				fields["arrival_time"] = readText(decoder)
			case "DepartureTime":
				fields["departure_time"] = readText(decoder)
			case "ArrivalDayOffset":
				fields["arrival_day_offset"] = readText(decoder)
			case "DepartureDayOffset":
				fields["departure_day_offset"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "TimetabledPassingTime" {
				emit(netex.Entity{Type: "TimetabledPassingTime", Fields: fields})
				return
			}
		}
	}
}

func parseTrainNumber(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
	fields := parseGenericAttrs(start)
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "ForAdvertisement":
				fields["for_advertisement"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "TrainNumber" {
				emit(netex.Entity{Type: "TrainNumber", Fields: fields})
				return nil
			}
		}
	}
}
