package gtfsrt

import (
	"fmt"
	"strings"
)

// Format represents a GTFS-RT feed serialization format.
type Format int

const (
	FormatJSON     Format = iota
	FormatProtobuf        // stub — not yet implemented
)

// DetectFormat guesses the format from the file extension or content.
func DetectFormat(path string) Format {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".pb") || strings.HasSuffix(lower, ".proto") || strings.HasSuffix(lower, ".bin") {
		return FormatProtobuf
	}
	return FormatJSON
}

// String returns the format name.
func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "JSON"
	case FormatProtobuf:
		return "Protobuf"
	default:
		return fmt.Sprintf("Format(%d)", int(f))
	}
}
