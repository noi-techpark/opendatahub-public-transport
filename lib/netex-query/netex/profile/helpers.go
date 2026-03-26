package profile

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/noi-techpark/opendatahub-public-transport/lib/netex-query/netex"
)

// getAttr extracts an attribute value by local name from a StartElement.
func getAttr(start xml.StartElement, name string) string {
	for _, attr := range start.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// readText reads CharData until the matching EndElement.
// Assumes the decoder is positioned right after the StartElement.
func readText(decoder *xml.Decoder) string {
	var b strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			return b.String()
		}
		switch t := tok.(type) {
		case xml.CharData:
			b.Write(t)
		case xml.EndElement:
			return strings.TrimSpace(b.String())
		case xml.StartElement:
			// Nested element inside a text element — skip it and continue
			skipElement(decoder)
		}
	}
}

// skipElement skips to the matching EndElement.
// Call after reading the StartElement.
func skipElement(decoder *xml.Decoder) {
	depth := 1
	for depth > 0 {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
}

// parseGenericAttrs extracts DataManagedObject base attributes from a StartElement.
func parseGenericAttrs(start xml.StartElement) map[string]string {
	fields := make(map[string]string, 8)
	for _, attr := range start.Attr {
		switch attr.Name.Local {
		case "id":
			fields["id"] = attr.Value
		case "version":
			fields["version"] = attr.Value
		case "responsibilitySetRef":
			fields["responsibility_set_ref"] = attr.Value
		case "created":
			fields["created"] = attr.Value
		case "changed":
			fields["changed"] = attr.Value
		case "modification":
			fields["modification"] = attr.Value
		case "status":
			fields["status"] = attr.Value
		case "dataSourceRef":
			fields["data_source_ref"] = attr.Value
		case "order":
			fields["order"] = attr.Value
		}
	}
	return fields
}

// parseValidBetween reads a ValidBetween element into fields.
func parseValidBetween(decoder *xml.Decoder, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "FromDate":
				fields["valid_between_from"] = readText(decoder)
			case "ToDate":
				fields["valid_between_to"] = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "ValidBetween" {
				return
			}
		}
	}
}

// parseKeyList reads a keyList element into fields with "key:" prefix.
func parseKeyList(decoder *xml.Decoder, fields map[string]string) {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "KeyValue" {
				parseKeyValue(decoder, fields)
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "keyList" {
				return
			}
		}
	}
}

func parseKeyValue(decoder *xml.Decoder, fields map[string]string) {
	var key, value string
	for {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Key":
				key = readText(decoder)
			case "Value":
				value = readText(decoder)
			default:
				skipElement(decoder)
			}
		case xml.EndElement:
			if t.Name.Local == "KeyValue" {
				if key != "" {
					fields["key:"+key] = value
				}
				return
			}
		}
	}
}

// readSubmode reads a TransportSubmode element (which contains a child like BusSubmode, RailSubmode, etc.)
func readSubmode(decoder *xml.Decoder) string {
	for {
		tok, err := decoder.Token()
		if err != nil {
			return ""
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// The child element name is the submode type (BusSubmode, RailSubmode, etc.)
			// and its text content is the value (localBus, regional, etc.)
			val := readText(decoder)
			// Skip remaining content until TransportSubmode closes
			for {
				tok2, err2 := decoder.Token()
				if err2 != nil {
					return val
				}
				if end, ok := tok2.(xml.EndElement); ok && end.Name.Local == "TransportSubmode" {
					return val
				}
				if _, ok := tok2.(xml.StartElement); ok {
					skipElement(decoder)
				}
			}
		case xml.EndElement:
			if t.Name.Local == "TransportSubmode" {
				return ""
			}
		default:
			_ = t
		}
	}
}

// parseCollection iterates children of a collection element,
// dispatching each child to the named entity parser.
func ParseCollection(decoder *xml.Decoder, entityType string,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler) error {
	parser, ok := parsers[entityType]
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("parseCollection(%s): %w", entityType, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if ok {
				if err := parser(decoder, t, emit); err != nil {
					return err
				}
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			return nil // end of collection
		}
	}
}

// parseCollectionAny dispatches each child based on its actual element name.
// Used for heterogeneous collections (e.g., connections may contain SiteConnection, DefaultConnection).
func ParseCollectionAny(decoder *xml.Decoder,
	parsers map[string]netex.EntityParserFunc, emit netex.EntityHandler) error {
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("parseCollectionAny: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if parser, ok := parsers[t.Name.Local]; ok {
				if err := parser(decoder, t, emit); err != nil {
					return err
				}
			} else {
				skipElement(decoder)
			}
		case xml.EndElement:
			return nil
		}
	}
}
