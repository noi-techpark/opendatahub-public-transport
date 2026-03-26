package netex

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Parse reads a NeTEx XML from r, using the given profile to drive parsing and emitting entities.
func Parse(r io.Reader, prof Profile, emit EntityHandler) error {
	decoder := xml.NewDecoder(r)

	frameParsers := prof.FrameParsers()
	entityParsers := prof.EntityParsers()

	// Navigate into: PublicationDelivery > dataObjects > CompositeFrame > frames
	if err := navigateToFrames(decoder); err != nil {
		return fmt.Errorf("navigate to frames: %w", err)
	}

	// Dispatch each frame to the profile's frame parser
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading frame: %w", err)
		}

		start, ok := tok.(xml.StartElement)
		if !ok {
			// EndElement for </frames> or </CompositeFrame> — we're done
			if _, isEnd := tok.(xml.EndElement); isEnd {
				return nil
			}
			continue
		}

		frameName := start.Name.Local
		if fp, ok := frameParsers[frameName]; ok {
			if err := fp(decoder, start, entityParsers, emit); err != nil {
				return fmt.Errorf("frame %s: %w", frameName, err)
			}
		} else {
			// Unknown frame — skip entirely
			skipUntilEnd(decoder, frameName)
		}
	}
}

// navigateToFrames positions the decoder at the <frames> element inside
// PublicationDelivery > dataObjects > CompositeFrame.
func navigateToFrames(decoder *xml.Decoder) error {
	path := []string{"PublicationDelivery", "dataObjects", "CompositeFrame", "frames"}
	idx := 0

	for idx < len(path) {
		tok, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("looking for %s: %w", path[idx], err)
		}
		if start, ok := tok.(xml.StartElement); ok {
			if start.Name.Local == path[idx] {
				idx++
			} else {
				// Skip elements we're not interested in at this level
				skipUntilEnd(decoder, start.Name.Local)
			}
		}
	}
	return nil
}

// skipUntilEnd skips all tokens until the EndElement matching the given local name.
func skipUntilEnd(decoder *xml.Decoder, name string) {
	depth := 1
	for depth > 0 {
		tok, err := decoder.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			if t.Name.Local == name && depth == 1 {
				return
			}
			depth--
		}
	}
}
