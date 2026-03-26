// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package profile

import (
	"encoding/xml"

	"github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
)

// ITL2Profile implements the Italian NeTEx Level 2 profile.
// It composes EPIP (Level 1) and adds/overrides:
// - GeneralFrame handler (for JourneyAccounting)
// - Extended DayType: +ShortName, +Description, +PrivateCode
// - Extended Operator: +TradingName, +CompanyNumber, +PrivateCode, +Description, +PrimaryMode
// - Extended VehicleType: +Name, +Description, +PrivateCode, +FuelType, +EuroClass
// - New entity: JourneyAccounting
// - New entity: GroupOfOperators
type ITL2Profile struct {
	base *EPIPProfile
}

func init() {
	netex.RegisterProfile(&ITL2Profile{base: &EPIPProfile{}})
}

func (p *ITL2Profile) Name() string { return "it-l2" }

func (p *ITL2Profile) FrameParsers() map[string]netex.FrameParserFunc {
	fps := p.base.FrameParsers()
	// Add GeneralFrame handler for L2-specific entities (JourneyAccounting)
	fps["GeneralFrame"] = parseGeneralFrameL2
	return fps
}

func (p *ITL2Profile) EntityParsers() map[string]netex.EntityParserFunc {
	eps := p.base.EntityParsers()
	// Override DayType to capture L2 additions
	eps["DayType"] = parseDayTypeL2
	// Override Operator to capture L2 additions
	eps["Operator"] = parseOperatorL2
	// Override VehicleType to capture L2 additions
	eps["VehicleType"] = parseVehicleTypeL2
	// Add new entity types
	eps["JourneyAccounting"] = parseJourneyAccounting
	eps["GroupOfOperators"] = parseGroupOfOperators
	return eps
}

func (p *ITL2Profile) Tables() []netex.TableDef {
	tables := p.base.Tables()

	// Extend DayType table with L2 columns
	for i, t := range tables {
		switch t.EntityType {
		case "DayType":
			tables[i].Columns = append(tables[i].Columns,
				netex.Column{Header: "short_name", Field: "short_name"},
				netex.Column{Header: "description", Field: "description"},
				netex.Column{Header: "private_code", Field: "private_code"},
			)
		case "Operator":
			tables[i].Columns = append(tables[i].Columns,
				netex.Column{Header: "trading_name", Field: "trading_name"},
				netex.Column{Header: "company_number", Field: "company_number"},
				netex.Column{Header: "private_code", Field: "private_code"},
				netex.Column{Header: "description", Field: "description"},
				netex.Column{Header: "primary_mode", Field: "primary_mode"},
			)
		case "VehicleType":
			tables[i].Columns = append(tables[i].Columns,
				netex.Column{Header: "name", Field: "name"},
				netex.Column{Header: "description", Field: "description"},
				netex.Column{Header: "private_code", Field: "private_code"},
				netex.Column{Header: "fuel_type", Field: "fuel_type"},
				netex.Column{Header: "euro_class", Field: "euro_class"},
			)
		}
	}

	// Add JourneyAccounting table
	tables = append(tables, netex.TableDef{
		EntityType: "JourneyAccounting",
		FileName:   "journey_accounting.csv",
		Columns: []netex.Column{
			{Header: "id", Field: "id"},
			{Header: "version", Field: "version"},
			{Header: "name", Field: "name"},
			{Header: "description", Field: "description"},
			{Header: "accounted_object_ref", Field: "accounted_object_ref"},
			{Header: "organisation_ref", Field: "organisation_ref"},
			{Header: "supply_contract_ref", Field: "supply_contract_ref"},
			{Header: "accounting_code", Field: "accounting_code"},
			{Header: "accounting_type", Field: "accounting_type"},
			{Header: "partial", Field: "partial"},
			{Header: "distance", Field: "distance"},
			{Header: "duration", Field: "duration"},
		},
	})

	// Add GroupOfOperators table
	tables = append(tables, netex.TableDef{
		EntityType: "GroupOfOperators",
		FileName:   "groups_of_operators.csv",
		Columns: []netex.Column{
			{Header: "id", Field: "id"},
			{Header: "version", Field: "version"},
			{Header: "name", Field: "name"},
			{Header: "short_name", Field: "short_name"},
			{Header: "description", Field: "description"},
			{Header: "private_code", Field: "private_code"},
		},
	})

	return tables
}

// --- L2 Frame Parsers ---

func parseGeneralFrameL2(
	decoder *xml.Decoder, start xml.StartElement,
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
			case "members":
				// GeneralFrame members can contain JourneyAccounting and other L2 entities
				if err := ParseCollectionAny(decoder, parsers, emit); err != nil {
					return err
				}
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "GeneralFrame" {
				return nil
			}
		}
	}
}

// --- L2 Entity Parsers (overrides) ---

// parseDayTypeL2 extends EPIP DayType with ShortName, Description, PrivateCode.
func parseDayTypeL2(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
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
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
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

// parseOperatorL2 extends EPIP Operator with TradingName, CompanyNumber, PrivateCode, Description, PrimaryMode.
func parseOperatorL2(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
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
			case "TradingName":
				fields["trading_name"] = readText(decoder)
			case "PublicCode":
				fields["public_code"] = readText(decoder)
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
			case "CompanyNumber":
				fields["company_number"] = readText(decoder)
			case "Description":
				fields["description"] = readText(decoder)
			case "OrganisationType":
				fields["organisation_type"] = readText(decoder)
			case "PrimaryMode":
				fields["primary_mode"] = readText(decoder)
			case "ValidBetween":
				parseValidBetween(decoder, fields)
			case "keyList":
				parseKeyList(decoder, fields)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "Operator" {
				emit(netex.Entity{Type: "Operator", Fields: fields})
				return nil
			}
		}
	}
}

// parseVehicleTypeL2 extends EPIP VehicleType with Name, Description, PrivateCode, FuelType, EuroClass.
func parseVehicleTypeL2(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
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
				fields["name"] = readText(decoder)
			case "Description":
				fields["description"] = readText(decoder)
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
			case "LowFloor":
				fields["low_floor"] = readText(decoder)
			case "HasLiftOrRamp":
				fields["has_lift_or_ramp"] = readText(decoder)
			case "HasHoist":
				fields["has_hoist"] = readText(decoder)
			case "Length":
				fields["length"] = readText(decoder)
			case "FuelType":
				fields["fuel_type"] = readText(decoder)
			case "EuroClass":
				fields["euro_class"] = readText(decoder)
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

// --- L2 New Entity Parsers ---

// parseJourneyAccounting parses a JourneyAccounting element (L2-specific entity).
func parseJourneyAccounting(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
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
				fields["name"] = readText(decoder)
			case "Description":
				fields["description"] = readText(decoder)
			case "AccountedObjectRef":
				fields["accounted_object_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "OrganisationRef":
				fields["organisation_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "SupplyContractRef":
				fields["supply_contract_ref"] = getAttr(t, "ref")
				skipElement(decoder)
			case "AccountingCode":
				fields["accounting_code"] = readText(decoder)
			case "AccountingType":
				fields["accounting_type"] = readText(decoder)
			case "Partial":
				fields["partial"] = readText(decoder)
			case "Distance":
				fields["distance"] = readText(decoder)
			case "Duration":
				fields["duration"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "JourneyAccounting" {
				emit(netex.Entity{Type: "JourneyAccounting", Fields: fields})
				return nil
			}
		}
	}
}

// parseGroupOfOperators parses a GroupOfOperators element (L2-specific entity).
func parseGroupOfOperators(decoder *xml.Decoder, start xml.StartElement, emit netex.EntityHandler) error {
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
				fields["name"] = readText(decoder)
			case "ShortName":
				fields["short_name"] = readText(decoder)
			case "Description":
				fields["description"] = readText(decoder)
			case "PrivateCode":
				fields["private_code"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "GroupOfOperators" {
				emit(netex.Entity{Type: "GroupOfOperators", Fields: fields})
				return nil
			}
		}
	}
}
