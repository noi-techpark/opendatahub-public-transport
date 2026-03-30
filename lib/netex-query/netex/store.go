// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import (
)

// Store is a generic entity store for parsed NeTEx data.
// It receives Entity objects from the parser and provides retrieval by type and ID.
// All entity types from any profile (EPIP, IT-L2, etc.) are stored.
type Store interface {
	// Put stores an entity. The entity's Type determines the "table".
	// Implementations may silently drop entities based on type filters.
	Put(entity Entity)

	// Accepts returns true if the store will accept entities of the given type.
	// Loaders should check this before parsing a file to skip unnecessary I/O.
	Accepts(entityType string) bool

	// All returns all entities of a given type.
	All(entityType string) []Entity

	// Get returns an entity by type and ID (value of the "id" field), or nil.
	Get(entityType string, id string) *Entity

	// Types returns all entity types that have been stored.
	Types() []string

	// Count returns total entity count for a type.
	Count(entityType string) int
}
