// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import (
	"sort"

)

// MemStore is an in-memory generic NeTEx entity store.
type MemStore struct {
	entities map[string][]Entity          // type → entities
	byID     map[string]map[string]*Entity // type → id → entity
}

// NewMemStore creates a new in-memory store.
func NewMemStore() *MemStore {
	return &MemStore{
		entities: make(map[string][]Entity),
		byID:     make(map[string]map[string]*Entity),
	}
}

func (m *MemStore) Put(entity Entity) {
	m.entities[entity.Type] = append(m.entities[entity.Type], entity)

	if id, ok := entity.Fields["id"]; ok && id != "" {
		if m.byID[entity.Type] == nil {
			m.byID[entity.Type] = make(map[string]*Entity)
		}
		e := m.entities[entity.Type][len(m.entities[entity.Type])-1]
		m.byID[entity.Type][id] = &e
	}
}

func (m *MemStore) All(entityType string) []Entity {
	return m.entities[entityType]
}

func (m *MemStore) Get(entityType string, id string) *Entity {
	if idx, ok := m.byID[entityType]; ok {
		return idx[id]
	}
	return nil
}

func (m *MemStore) Types() []string {
	types := make([]string, 0, len(m.entities))
	for t := range m.entities {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func (m *MemStore) Count(entityType string) int {
	return len(m.entities[entityType])
}

// Handler returns an EntityHandler that stores entities into this MemStore.
// This can be passed directly to the parser: netex.Parse(reader, profile, store.Handler())
func (m *MemStore) Handler() EntityHandler {
	return func(entity Entity) {
		m.Put(entity)
	}
}
