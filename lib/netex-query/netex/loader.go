// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

)

// Load reads NeTEx parsed CSVs from a directory into the given Store,
// then builds and returns an indexed NeTExFeed.
// Files whose entity type is not accepted by the store are skipped entirely.
func Load(dir string, store Store) (*NeTExFeed, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".csv") {
			continue
		}

		entityType := csvFileToEntityType(entry.Name())
		if !store.Accepts(entityType) {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if err := loadCSVIntoStore(path, entityType, store); err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
	}

	feed := NewNeTExFeed(store)
	feed.BuildIndexes()
	return feed, nil
}

// csvFileToEntityType maps a CSV filename to a NeTEx entity type name.
// e.g. "service_journeys.csv" → "ServiceJourney"
func csvFileToEntityType(filename string) string {
	name := strings.TrimSuffix(filename, ".csv")
	mapping := map[string]string{
		"service_journeys":     "ServiceJourney",
		"journey_patterns":     "ServiceJourneyPattern",
		"routes":               "Route",
		"lines":                "Line",
		"stop_places":          "StopPlace",
		"quays":                "Quay",
		"scheduled_stop_points": "ScheduledStopPoint",
		"authorities":          "Authority",
		"operators":            "Operator",
		"vehicle_types":        "VehicleType",
		"directions":           "Direction",
		"route_points":         "RoutePoint",
		"destination_displays": "DestinationDisplay",
		"stop_points_in_jp":    "StopPointInJourneyPattern",
		"service_links":        "ServiceLink",
		"connections":          "Connection",
		"notices":              "Notice",
		"stop_assignments":     "StopAssignment",
		"day_types":            "DayType",
		"operating_periods":    "OperatingPeriod",
		"day_type_assignments": "DayTypeAssignment",
		"passing_times":        "TimetabledPassingTime",
		"train_numbers":        "TrainNumber",
		"journey_accountings":  "JourneyAccounting",
		"groups_of_operators":  "GroupOfOperators",
	}
	if t, ok := mapping[name]; ok {
		return t
	}
	return name // fallback: use filename as-is
}

// loadCSVIntoStore reads a CSV file and puts each row as a generic Entity into the store.
func loadCSVIntoStore(path, entityType string, store Store) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	// Clean header names
	for i, name := range header {
		header[i] = strings.TrimSpace(name)
	}

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fields := make(map[string]string, len(header))
		for i, name := range header {
			if i < len(rec) {
				fields[name] = strings.TrimSpace(rec[i])
			}
		}

		store.Put(Entity{
			Type:   entityType,
			Fields: fields,
		})
	}
	return nil
}
