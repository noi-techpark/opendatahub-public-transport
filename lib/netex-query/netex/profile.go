package netex

import (
	"encoding/xml"
	"fmt"
	"sync"
)

// FrameParserFunc parses a frame element, dispatching to entity parsers.
type FrameParserFunc func(
	decoder *xml.Decoder,
	start xml.StartElement,
	parsers map[string]EntityParserFunc,
	emit EntityHandler,
) error

// EntityParserFunc parses a single entity element.
type EntityParserFunc func(
	decoder *xml.Decoder,
	start xml.StartElement,
	emit EntityHandler,
) error

// Profile is the full plugin interface.
// A profile controls BOTH parsing (what to extract) and output (what to write).
type Profile interface {
	Name() string

	// FrameParsers returns frame-level parsers keyed by frame element name.
	FrameParsers() map[string]FrameParserFunc

	// EntityParsers returns entity-level parsers keyed by entity element name.
	EntityParsers() map[string]EntityParserFunc

	// Tables returns CSV table definitions for output.
	Tables() []TableDef
}

// Registry for profile registration and lookup.
var (
	registryMu sync.RWMutex
	registry   = map[string]Profile{}
)

// RegisterProfile adds a profile to the registry.
func RegisterProfile(p Profile) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.Name()] = p
}

// GetProfile returns a registered profile by name.
func GetProfile(name string) (Profile, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown profile: %q (available: %v)", name, ListProfiles())
	}
	return p, nil
}

// ListProfiles returns all registered profile names.
func ListProfiles() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
