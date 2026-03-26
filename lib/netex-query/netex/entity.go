// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

// Entity is a generic parsed NeTEx entity.
// Fields captured as flat map — both CEN standard and profile additions.
type Entity struct {
	Type   string            // NeTEx entity type (e.g., "Line", "ServiceJourney")
	Fields map[string]string // field_name → value
}

// EntityHandler receives parsed entities.
type EntityHandler func(entity Entity)
