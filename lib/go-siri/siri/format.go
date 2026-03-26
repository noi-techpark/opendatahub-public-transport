// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"fmt"
	"os"
	"strings"
)

// Format represents a SIRI feed serialization format.
type Format int

const (
	FormatJSON Format = iota
	FormatXML
)

// DetectFormat guesses the format from the file extension or content.
func DetectFormat(path string) Format {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".xml") {
		return FormatXML
	}
	if strings.HasSuffix(lower, ".json") {
		return FormatJSON
	}
	// Peek first non-whitespace byte
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return FormatJSON
	}
	for _, b := range data {
		switch b {
		case '{', '[':
			return FormatJSON
		case '<':
			return FormatXML
		case ' ', '\t', '\n', '\r', 0xEF, 0xBB, 0xBF: // whitespace + BOM
			continue
		}
		break
	}
	return FormatJSON
}

// FormatString returns the format name.
func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "JSON"
	case FormatXML:
		return "XML"
	default:
		return fmt.Sprintf("Format(%d)", int(f))
	}
}
