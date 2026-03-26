// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package profile

import (
	"github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
)

// EPIPProfile implements the European Passenger Information Profile (Level 1).
type EPIPProfile struct{}

func init() {
	netex.RegisterProfile(&EPIPProfile{})
}

func (p *EPIPProfile) Name() string { return "epip" }

func (p *EPIPProfile) FrameParsers() map[string]netex.FrameParserFunc {
	return map[string]netex.FrameParserFunc{
		"ResourceFrame":        parseResourceFrame,
		"SiteFrame":            parseSiteFrame,
		"ServiceFrame":         parseServiceFrame,
		"ServiceCalendarFrame": parseServiceCalendarFrame,
		"TimetableFrame":       parseTimetableFrame,
	}
}

func (p *EPIPProfile) EntityParsers() map[string]netex.EntityParserFunc {
	return map[string]netex.EntityParserFunc{
		"Authority":                parseAuthority,
		"Operator":                 parseOperator,
		"VehicleType":              parseVehicleType,
		"StopPlace":                parseStopPlace,
		"ScheduledStopPoint":       parseScheduledStopPoint,
		"Direction":                parseDirection,
		"Route":                    parseRoute,
		"Line":                     parseLine,
		"FlexibleLine":             parseLine, // same structure
		"DestinationDisplay":       parseDestinationDisplay,
		"ServiceJourneyPattern":    parseServiceJourneyPattern,
		"ServiceLink":              parseServiceLink,
		"SiteConnection":           parseSiteConnection,
		"DefaultConnection":        parseSiteConnection,
		"Notice":                   parseNotice,
		"PassengerStopAssignment":  parsePassengerStopAssignment,
		"DayType":                  parseDayType,
		"UicOperatingPeriod":       parseUicOperatingPeriod,
		"DayTypeAssignment":        parseDayTypeAssignment,
		"ServiceJourney":           parseServiceJourney,
		"TrainNumber":              parseTrainNumber,
	}
}

func (p *EPIPProfile) Tables() []netex.TableDef {
	return []netex.TableDef{
		{EntityType: "Authority", FileName: "authorities.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"name_lang", "name_lang"},
			{"short_name", "short_name"}, {"public_code", "public_code"}, {"organisation_type", "organisation_type"},
		}},
		{EntityType: "Operator", FileName: "operators.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"name_lang", "name_lang"},
			{"short_name", "short_name"}, {"legal_name", "legal_name"}, {"organisation_type", "organisation_type"},
		}},
		{EntityType: "VehicleType", FileName: "vehicle_types.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"low_floor", "low_floor"},
			{"has_lift_or_ramp", "has_lift_or_ramp"}, {"has_hoist", "has_hoist"}, {"length", "length"},
		}},
		{EntityType: "StopPlace", FileName: "stop_places.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"name_lang", "name_lang"},
			{"short_name", "short_name"}, {"private_code", "private_code"},
			{"longitude", "longitude"}, {"latitude", "latitude"},
			{"transport_mode", "transport_mode"}, {"stop_place_type", "stop_place_type"},
			{"topographic_place_ref", "topographic_place_ref"},
		}},
		{EntityType: "Quay", FileName: "quays.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"stop_place_id", "stop_place_id"},
			{"name", "name"}, {"name_lang", "name_lang"}, {"short_name", "short_name"},
			{"private_code", "private_code"}, {"public_code", "public_code"},
			{"longitude", "longitude"}, {"latitude", "latitude"}, {"quay_type", "quay_type"},
		}},
		{EntityType: "ScheduledStopPoint", FileName: "scheduled_stop_points.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"name_lang", "name_lang"},
			{"short_name", "short_name"}, {"longitude", "longitude"}, {"latitude", "latitude"},
		}},
		{EntityType: "Direction", FileName: "directions.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"direction_type", "direction_type"},
		}},
		{EntityType: "Route", FileName: "routes.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"distance", "distance"},
			{"line_ref", "line_ref"}, {"direction_type", "direction_type"},
			{"direction_ref", "direction_ref"}, {"inverse_route_ref", "inverse_route_ref"},
		}},
		{EntityType: "PointOnRoute", FileName: "route_points.csv", Columns: []netex.Column{
			{"id", "id"}, {"route_id", "route_id"}, {"order", "order"},
			{"scheduled_stop_point_ref", "scheduled_stop_point_ref"},
		}},
		{EntityType: "Line", FileName: "lines.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"responsibility_set_ref", "responsibility_set_ref"},
			{"name", "name"}, {"name_lang", "name_lang"}, {"short_name", "short_name"},
			{"description", "description"}, {"transport_mode", "transport_mode"},
			{"transport_submode", "transport_submode"}, {"url", "url"},
			{"public_code", "public_code"}, {"private_code", "private_code"},
			{"operator_ref", "operator_ref"}, {"authority_ref", "authority_ref"},
			{"monitored", "monitored"}, {"type_of_line_ref", "type_of_line_ref"},
		}},
		{EntityType: "DestinationDisplay", FileName: "destination_displays.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"},
			{"side_text", "side_text"}, {"front_text", "front_text"},
		}},
		{EntityType: "ServiceJourneyPattern", FileName: "journey_patterns.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"route_ref", "route_ref"},
		}},
		{EntityType: "StopPointInJourneyPattern", FileName: "stop_points_in_jp.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"journey_pattern_id", "journey_pattern_id"},
			{"order", "order"}, {"scheduled_stop_point_ref", "scheduled_stop_point_ref"},
			{"for_alighting", "for_alighting"}, {"for_boarding", "for_boarding"},
			{"destination_display_ref", "destination_display_ref"}, {"request_stop", "request_stop"},
		}},
		{EntityType: "ServiceLink", FileName: "service_links.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"distance", "distance"},
			{"from_point_ref", "from_point_ref"}, {"to_point_ref", "to_point_ref"},
		}},
		{EntityType: "SiteConnection", FileName: "connections.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"default_duration", "default_duration"},
			{"both_ways", "both_ways"}, {"from_stop_place_ref", "from_stop_place_ref"},
			{"from_quay_ref", "from_quay_ref"}, {"to_stop_place_ref", "to_stop_place_ref"},
			{"to_quay_ref", "to_quay_ref"},
		}},
		{EntityType: "Notice", FileName: "notices.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"text", "text"}, {"text_lang", "text_lang"},
		}},
		{EntityType: "PassengerStopAssignment", FileName: "stop_assignments.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"order", "order"},
			{"scheduled_stop_point_ref", "scheduled_stop_point_ref"},
			{"stop_place_ref", "stop_place_ref"}, {"quay_ref", "quay_ref"},
		}},
		{EntityType: "DayType", FileName: "day_types.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"},
		}},
		{EntityType: "UicOperatingPeriod", FileName: "operating_periods.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"},
			{"from_date", "from_date"}, {"to_date", "to_date"}, {"valid_day_bits", "valid_day_bits"},
		}},
		{EntityType: "DayTypeAssignment", FileName: "day_type_assignments.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"order", "order"},
			{"operating_period_ref", "operating_period_ref"}, {"day_type_ref", "day_type_ref"},
		}},
		{EntityType: "ServiceJourney", FileName: "service_journeys.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"name", "name"}, {"distance", "distance"},
			{"transport_mode", "transport_mode"}, {"transport_submode", "transport_submode"},
			{"departure_time", "departure_time"}, {"day_type_ref", "day_type_ref"},
			{"journey_pattern_ref", "journey_pattern_ref"}, {"vehicle_type_ref", "vehicle_type_ref"},
			{"operator_ref", "operator_ref"}, {"line_ref", "line_ref"},
			{"train_number_ref", "train_number_ref"},
		}},
		{EntityType: "TimetabledPassingTime", FileName: "passing_times.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"service_journey_id", "service_journey_id"},
			{"stop_point_in_jp_ref", "stop_point_in_jp_ref"}, {"order", "order"},
			{"arrival_time", "arrival_time"}, {"departure_time", "departure_time"},
			{"arrival_day_offset", "arrival_day_offset"}, {"departure_day_offset", "departure_day_offset"},
		}},
		{EntityType: "TrainNumber", FileName: "train_numbers.csv", Columns: []netex.Column{
			{"id", "id"}, {"version", "version"}, {"for_advertisement", "for_advertisement"},
		}},
	}
}
