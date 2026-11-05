// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import (
	"encoding/xml"
	"fmt"
	"io"
)

// ParseOption configures Parse.
type ParseOption func(*Decoder)

// WithTypes restricts Parse to the entity types accepted by fn. Other entities
// are never emitted, and subtrees holding only such entities are skipped
// without being parsed.
func WithTypes(accepts func(entityType string) bool) ParseOption {
	return func(d *Decoder) { d.accepts = accepts }
}

// Parse reads a NeTEx XML from r, using the given profile to drive parsing and emitting entities.
func Parse(r io.Reader, prof Profile, emit EntityHandler, opts ...ParseOption) error {
	decoder := NewDecoder(r)
	for _, opt := range opts {
		opt(decoder)
	}
	if decoder.accepts != nil {
		all := emit
		emit = func(e Entity) {
			if decoder.accepts(e.Type) {
				all(e)
			}
		}
	}

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
			if err := decoder.Skip(); err != nil {
				return fmt.Errorf("skip frame %s: %w", frameName, err)
			}
		}
	}
}

// navigateToFrames positions the decoder at the <frames> element inside
// PublicationDelivery > dataObjects > CompositeFrame.
func navigateToFrames(decoder *Decoder) error {
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
				if err := decoder.Skip(); err != nil {
					return fmt.Errorf("skip %s: %w", start.Name.Local, err)
				}
			}
		}
	}
	return nil
}
