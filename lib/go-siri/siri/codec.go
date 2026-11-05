// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"bytes"
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

// wrapObjects rewrites SIRI Lite JSON so that every value of a key in listKeys
// is an array, wrapping single objects in [...]. Decoding the result into
// plain slices is one pass of encoding/json; a custom UnmarshalJSON per level
// instead re-scans each subtree once per nesting level, which made ET
// decoding ~6x slower. Input without single objects is returned unchanged.
func wrapObjects(data []byte, listKeys map[string]bool) []byte {
	var out []byte    // nil until the first rewrite
	copied := 0       // data[:copied] is already in out
	var closeAt []int // object depths whose closing brace must be followed by ']'
	depth := 0

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '{':
			depth++
		case '}':
			if n := len(closeAt); n > 0 && closeAt[n-1] == depth {
				closeAt = closeAt[:n-1]
				out = append(out, data[copied:i+1]...)
				out = append(out, ']')
				copied = i + 1
			}
			depth--
		case '"':
			end := stringEnd(data, i)
			if end < 0 {
				return data
			}
			key := data[i+1 : end]
			i = end
			if !listKeys[string(key)] {
				continue
			}
			j := skipSpace(data, end+1)
			if j >= len(data) || data[j] != ':' {
				continue
			}
			j = skipSpace(data, j+1)
			if j < len(data) && data[j] == '{' {
				out = append(out, data[copied:j]...)
				out = append(out, '[')
				copied = j
				closeAt = append(closeAt, depth+1)
				i = j - 1
			}
		}
	}
	if out == nil {
		return data
	}
	return append(out, data[copied:]...)
}

// stringEnd returns the index of the quote closing the string opened at start, or -1.
func stringEnd(data []byte, start int) int {
	for i := start + 1; i < len(data); {
		k := bytes.IndexAny(data[i:], `"\`)
		if k < 0 {
			return -1
		}
		i += k
		if data[i] == '"' {
			return i
		}
		i += 2
	}
	return -1
}

func skipSpace(data []byte, i int) int {
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	return i
}
