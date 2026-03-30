// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"encoding/json"
	"fmt"
	"os"
)

// serialize encodes any SIRI feed struct to bytes in the given format.
func serialize(feed any, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(feed, "", "  ")
	case FormatXML:
		return nil, fmt.Errorf("XML serialization not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}
}

// deserialize decodes bytes into a SIRI feed struct in the given format.
func deserialize(data []byte, format Format, target any) error {
	switch format {
	case FormatJSON:
		return json.Unmarshal(data, target)
	case FormatXML:
		return fmt.Errorf("XML deserialization not yet implemented")
	default:
		return fmt.Errorf("unsupported format: %v", format)
	}
}

// loadFromFile reads a file and deserializes into target in the given format.
func loadFromFile(path string, format Format, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return deserialize(data, format, target)
}

// dumpToFile serializes and writes to a file in the given format.
func dumpToFile(path string, format Format, feed any) error {
	data, err := serialize(feed, format)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// unmarshalArrayOrObject handles SIRI Lite JSON polymorphism where a field
// can be either an array [...] or a single object {...}.
// Returns a slice of T in both cases.
func unmarshalArrayOrObject[T any](raw json.RawMessage) ([]T, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var arr []T
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var single T
	if err := json.Unmarshal(raw, &single); err == nil {
		return []T{single}, nil
	}
	return nil, fmt.Errorf("expected JSON array or object, got: %.40s", raw)
}
