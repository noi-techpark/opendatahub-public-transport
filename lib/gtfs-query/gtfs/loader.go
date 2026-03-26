// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfs

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// LoadOptions configures what to load from a GTFS zip.
type LoadOptions struct {
	LoadShapes       bool
	LoadTranslations bool
	ExcludeTables    []string // file names to skip (e.g. "shapes.txt")
}

func (o LoadOptions) isExcluded(name string) bool {
	for _, e := range o.ExcludeTables {
		if e == name {
			return true
		}
	}
	return false
}

// Load opens a GTFS zip and loads it into the given Store, returning a Feed.
func Load(zipPath string, opts LoadOptions, store Store) (*Feed, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	for _, file := range zr.File {
		if opts.isExcluded(file.Name) {
			continue
		}

		var loader func(io.Reader) error

		switch file.Name {
		case "agency.txt":
			loader = func(r io.Reader) error { return loadAgency(r, store) }
		case "stops.txt":
			loader = func(r io.Reader) error { return loadStops(r, store) }
		case "routes.txt":
			loader = func(r io.Reader) error { return loadRoutes(r, store) }
		case "trips.txt":
			loader = func(r io.Reader) error { return loadTrips(r, store) }
		case "stop_times.txt":
			loader = func(r io.Reader) error { return loadStopTimes(r, store) }
		case "calendar.txt":
			loader = func(r io.Reader) error { return loadCalendar(r, store) }
		case "calendar_dates.txt":
			loader = func(r io.Reader) error { return loadCalendarDates(r, store) }
		case "shapes.txt":
			if opts.LoadShapes {
				loader = func(r io.Reader) error { return loadShapes(r, store) }
			}
		case "translations.txt":
			if opts.LoadTranslations {
				loader = func(r io.Reader) error { return loadTranslations(r, store) }
			}
		}

		if loader != nil {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("open %s: %w", file.Name, err)
			}
			if err := loader(stripBOM(rc)); err != nil {
				rc.Close()
				return nil, fmt.Errorf("load %s: %w", file.Name, err)
			}
			rc.Close()
		}
	}

	store.BuildIndexes()
	return NewFeed(store), nil
}

// stripBOM wraps a reader to skip a UTF-8 BOM if present.
func stripBOM(r io.Reader) io.Reader {
	br := bufio.NewReader(r)
	b, err := br.Peek(3)
	if err == nil && len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		br.Discard(3)
	}
	return br
}

func headerIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, name := range header {
		idx[strings.TrimSpace(name)] = i
	}
	return idx
}

func col(record []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[i])
}

// --- Per-file loaders ---

func loadAgency(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddAgency(Agency{
			AgencyID:       col(rec, idx, "agency_id"),
			AgencyName:     col(rec, idx, "agency_name"),
			AgencyURL:      col(rec, idx, "agency_url"),
			AgencyTimezone: col(rec, idx, "agency_timezone"),
			AgencyLang:     col(rec, idx, "agency_lang"),
			AgencyPhone:    col(rec, idx, "agency_phone"),
			AgencyFareURL:  col(rec, idx, "agency_fare_url"),
			AgencyEmail:    col(rec, idx, "agency_email"),
		})
	}
	return nil
}

func loadStops(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddStop(Stop{
			StopID:             col(rec, idx, "stop_id"),
			StopCode:           col(rec, idx, "stop_code"),
			StopName:           col(rec, idx, "stop_name"),
			StopDesc:           col(rec, idx, "stop_desc"),
			StopLat:            parseFloat(col(rec, idx, "stop_lat")),
			StopLon:            parseFloat(col(rec, idx, "stop_lon")),
			ZoneID:             col(rec, idx, "zone_id"),
			StopURL:            col(rec, idx, "stop_url"),
			LocationType:       parseInt(col(rec, idx, "location_type")),
			ParentStation:      col(rec, idx, "parent_station"),
			StopTimezone:       col(rec, idx, "stop_timezone"),
			WheelchairBoarding: parseInt(col(rec, idx, "wheelchair_boarding")),
			LevelID:            col(rec, idx, "level_id"),
			PlatformCode:       col(rec, idx, "platform_code"),
		})
	}
	return nil
}

func loadRoutes(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddRoute(Route{
			RouteID:           col(rec, idx, "route_id"),
			AgencyID:          col(rec, idx, "agency_id"),
			RouteShortName:    col(rec, idx, "route_short_name"),
			RouteLongName:     col(rec, idx, "route_long_name"),
			RouteDesc:         col(rec, idx, "route_desc"),
			RouteType:         parseInt(col(rec, idx, "route_type")),
			RouteURL:          col(rec, idx, "route_url"),
			RouteColor:        col(rec, idx, "route_color"),
			RouteTextColor:    col(rec, idx, "route_text_color"),
			RouteSortOrder:    parseInt(col(rec, idx, "route_sort_order")),
			ContinuousPickup:  parseInt(col(rec, idx, "continuous_pickup")),
			ContinuousDropOff: parseInt(col(rec, idx, "continuous_drop_off")),
			NetworkID:         col(rec, idx, "network_id"),
		})
	}
	return nil
}

func loadTrips(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddTrip(Trip{
			RouteID:              col(rec, idx, "route_id"),
			ServiceID:            col(rec, idx, "service_id"),
			TripID:               col(rec, idx, "trip_id"),
			TripHeadsign:         col(rec, idx, "trip_headsign"),
			TripShortName:        col(rec, idx, "trip_short_name"),
			DirectionID:          parseInt(col(rec, idx, "direction_id")),
			BlockID:              col(rec, idx, "block_id"),
			ShapeID:              col(rec, idx, "shape_id"),
			WheelchairAccessible: parseInt(col(rec, idx, "wheelchair_accessible")),
			BikesAllowed:         parseInt(col(rec, idx, "bikes_allowed")),
		})
	}
	return nil
}

func loadStopTimes(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddStopTime(StopTime{
			TripID:            col(rec, idx, "trip_id"),
			ArrivalTime:       ParseGTFSTime(col(rec, idx, "arrival_time")),
			DepartureTime:     ParseGTFSTime(col(rec, idx, "departure_time")),
			StopID:            col(rec, idx, "stop_id"),
			StopSequence:      parseInt(col(rec, idx, "stop_sequence")),
			StopHeadsign:      col(rec, idx, "stop_headsign"),
			PickupType:        parseInt(col(rec, idx, "pickup_type")),
			DropOffType:       parseInt(col(rec, idx, "drop_off_type")),
			ContinuousPickup:  parseInt(col(rec, idx, "continuous_pickup")),
			ContinuousDropOff: parseInt(col(rec, idx, "continuous_drop_off")),
			ShapeDistTraveled: parseFloat(col(rec, idx, "shape_dist_traveled")),
			Timepoint:         parseInt(col(rec, idx, "timepoint")),
		})
	}
	return nil
}

func loadCalendar(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddCalendarEntry(CalendarEntry{
			ServiceID: col(rec, idx, "service_id"),
			Monday:    parseBool(col(rec, idx, "monday")),
			Tuesday:   parseBool(col(rec, idx, "tuesday")),
			Wednesday: parseBool(col(rec, idx, "wednesday")),
			Thursday:  parseBool(col(rec, idx, "thursday")),
			Friday:    parseBool(col(rec, idx, "friday")),
			Saturday:  parseBool(col(rec, idx, "saturday")),
			Sunday:    parseBool(col(rec, idx, "sunday")),
			StartDate: ParseDate(col(rec, idx, "start_date")),
			EndDate:   ParseDate(col(rec, idx, "end_date")),
		})
	}
	return nil
}

func loadCalendarDates(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddCalendarDate(CalendarDate{
			ServiceID:     col(rec, idx, "service_id"),
			Date:          ParseDate(col(rec, idx, "date")),
			ExceptionType: parseInt(col(rec, idx, "exception_type")),
		})
	}
	return nil
}

func loadShapes(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddShape(Shape{
			ShapeID:           col(rec, idx, "shape_id"),
			ShapePtLat:        parseFloat(col(rec, idx, "shape_pt_lat")),
			ShapePtLon:        parseFloat(col(rec, idx, "shape_pt_lon")),
			ShapePtSequence:   parseInt(col(rec, idx, "shape_pt_sequence")),
			ShapeDistTraveled: parseFloat(col(rec, idx, "shape_dist_traveled")),
		})
	}
	return nil
}

func loadTranslations(r io.Reader, store Store) error {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	header, err := reader.Read()
	if err != nil {
		return err
	}
	idx := headerIndex(header)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		store.AddTranslation(Translation{
			TableName:       col(rec, idx, "table_name"),
			FieldName:       col(rec, idx, "field_name"),
			Language:        col(rec, idx, "language"),
			TranslationText: col(rec, idx, "translation"),
			RecordID:        col(rec, idx, "record_id"),
			RecordSubID:     col(rec, idx, "record_sub_id"),
		})
	}
	return nil
}
