// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package profile

import (
	"encoding/xml"
	"fmt"

	"github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
)

// parseResourceFrame handles the ResourceFrame: organisations and vehicleTypes.
func parseResourceFrame(
	decoder *xml.Decoder, start xml.StartElement,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler,
) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("ResourceFrame: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "organisations":
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "vehicleTypes":
				if err := ParseCollection(decoder, "VehicleType", parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ResourceFrame" {
				return nil
			}
		}
	}
}

// parseSiteFrame handles the SiteFrame: stopPlaces (+ nested quays).
func parseSiteFrame(
	decoder *xml.Decoder, start xml.StartElement,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler,
) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("SiteFrame: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "stopPlaces":
				if err := ParseCollection(decoder, "StopPlace", parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder) // skip topographicPlaces, etc.
			}
		case xml.EndElement:
			if t.Name.Local == "SiteFrame" {
				return nil
			}
		}
	}
}

// parseServiceFrame handles the ServiceFrame — the largest frame with many collection types.
func parseServiceFrame(
	decoder *xml.Decoder, start xml.StartElement,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler,
) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("ServiceFrame: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "directions":
				if err := ParseCollection(decoder, "Direction", parsers, emit); err != nil {
					return err
				}
			case "routes":
				if err := ParseCollection(decoder, "Route", parsers, emit); err != nil {
					return err
				}
			case "lines":
				// Lines collection may contain Line or FlexibleLine
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "destinationDisplays":
				if err := ParseCollection(decoder, "DestinationDisplay", parsers, emit); err != nil {
					return err
				}
			case "scheduledStopPoints":
				if err := ParseCollection(decoder, "ScheduledStopPoint", parsers, emit); err != nil {
					return err
				}
			case "serviceLinks":
				if err := ParseCollection(decoder, "ServiceLink", parsers, emit); err != nil {
					return err
				}
			case "connections":
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "stopAssignments":
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "journeyPatterns":
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "notices":
				if err := ParseCollection(decoder, "Notice", parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ServiceFrame" {
				return nil
			}
		}
	}
}

// parseServiceCalendarFrame handles ServiceCalendarFrame.
// Structure: ServiceCalendar > {dayTypes, operatingPeriods, dayTypeAssignments}
func parseServiceCalendarFrame(
	decoder *xml.Decoder, start xml.StartElement,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler,
) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("ServiceCalendarFrame: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "ServiceCalendar":
				if err := parseServiceCalendarContent(decoder, parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ServiceCalendarFrame" {
				return nil
			}
		}
	}
}

func parseServiceCalendarContent(
	decoder *xml.Decoder,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler,
) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "dayTypes":
				if err := ParseCollection(decoder, "DayType", parsers, emit); err != nil {
					return err
				}
			case "operatingPeriods":
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "dayTypeAssignments":
				if err := ParseCollection(decoder, "DayTypeAssignment", parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder) // Name, FromDate, ToDate of ServiceCalendar itself
			}
		case xml.EndElement:
			if t.Name.Local == "ServiceCalendar" {
				return nil
			}
		}
	}
}

// parseTimetableFrame handles TimetableFrame: vehicleJourneys and trainNumbers.
func parseTimetableFrame(
	decoder *xml.Decoder, start xml.StartElement,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler,
) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("TimetableFrame: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "vehicleJourneys":
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			case "trainNumbers":
				if err := ParseCollection(decoder, "TrainNumber", parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "TimetableFrame" {
				return nil
			}
		}
	}
}
