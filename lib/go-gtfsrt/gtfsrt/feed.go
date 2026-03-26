package gtfsrt

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// NewFeedMessage creates a new GTFS-RT FeedMessage with header.
func NewFeedMessage() *FeedMessage {
	return &FeedMessage{
		Header: FeedHeader{
			GtfsRealtimeVersion: "2.0",
			Incrementality:      "FULL_DATASET",
			Timestamp:           time.Now().Unix(),
		},
	}
}

// AddEntity appends an entity to the feed.
func (fm *FeedMessage) AddEntity(e FeedEntity) {
	fm.Entity = append(fm.Entity, e)
}

// IntPtr returns a pointer to an int (for optional direction_id).
func IntPtr(v int) *int {
	return &v
}

// --- Serialize ---

// Serialize encodes the feed to bytes in the given format.
func (fm *FeedMessage) Serialize(format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		type Alias FeedMessage
		return json.MarshalIndent((*Alias)(fm), "", "  ")
	case FormatProtobuf:
		return serializeProtobuf(fm)
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}
}

// Dump serializes and writes to a file in the given format.
func (fm *FeedMessage) Dump(path string, format Format) error {
	data, err := fm.Serialize(format)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// --- Deserialize ---

// Deserialize decodes bytes into a FeedMessage in the given format.
func Deserialize(data []byte, format Format) (*FeedMessage, error) {
	switch format {
	case FormatJSON:
		var fm FeedMessage
		if err := json.Unmarshal(data, &fm); err != nil {
			return nil, fmt.Errorf("parse GTFS-RT JSON: %w", err)
		}
		return &fm, nil
	case FormatProtobuf:
		return deserializeProtobuf(data)
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}
}

// Load reads a file and deserializes it in the given format.
func Load(path string, format Format) (*FeedMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return Deserialize(data, format)
}
