// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GTFSTime represents a GTFS time as seconds since midnight.
// Supports values >= 86400 for service days extending past midnight (e.g., 25:30:00).
type GTFSTime int32

const GTFSTimeNotSet GTFSTime = -1

// ParseGTFSTime parses "HH:MM:SS" or "H:MM:SS" into GTFSTime.
func ParseGTFSTime(s string) GTFSTime {
	s = strings.TrimSpace(s)
	if s == "" {
		return GTFSTimeNotSet
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return GTFSTimeNotSet
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	sec, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return GTFSTimeNotSet
	}
	return GTFSTime(h*3600 + m*60 + sec)
}

// String returns "HH:MM:SS" format.
func (t GTFSTime) String() string {
	if t == GTFSTimeNotSet {
		return ""
	}
	s := int(t)
	h := s / 3600
	m := (s % 3600) / 60
	sec := s % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
}

// Seconds returns total seconds since midnight.
func (t GTFSTime) Seconds() int { return int(t) }

// IsSet returns true if the time was parsed successfully.
func (t GTFSTime) IsSet() bool { return t != GTFSTimeNotSet }

// Date represents a GTFS date (YYYYMMDD) as a compact integer.
type Date int32

const DateNotSet Date = 0

// ParseDate parses "YYYYMMDD" into Date.
func ParseDate(s string) Date {
	s = strings.TrimSpace(s)
	if len(s) != 8 {
		return DateNotSet
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return DateNotSet
	}
	return Date(v)
}

// MustParseDate parses "YYYYMMDD" or panics.
func MustParseDate(s string) Date {
	d := ParseDate(s)
	if d == DateNotSet {
		panic("invalid date: " + s)
	}
	return d
}

// String returns "YYYYMMDD" format.
func (d Date) String() string {
	if d == DateNotSet {
		return ""
	}
	return fmt.Sprintf("%08d", int(d))
}

// ToTime converts to time.Time (midnight UTC on that date).
func (d Date) ToTime() time.Time {
	v := int(d)
	y := v / 10000
	m := (v / 100) % 100
	day := v % 100
	return time.Date(y, time.Month(m), day, 0, 0, 0, 0, time.UTC)
}

// Weekday returns the day of the week.
func (d Date) Weekday() time.Weekday {
	return d.ToTime().Weekday()
}

// Year, Month, Day extract date components.
func (d Date) Year() int  { return int(d) / 10000 }
func (d Date) Month() int { return (int(d) / 100) % 100 }
func (d Date) Day() int   { return int(d) % 100 }

// IsSet returns true if the date was parsed.
func (d Date) IsSet() bool { return d != DateNotSet }

// --- GTFS Entity Structs ---

// Agency represents a GTFS agency.txt row.
type Agency struct {
	AgencyID       string
	AgencyName     string
	AgencyURL      string
	AgencyTimezone string
	AgencyLang     string
	AgencyPhone    string
	AgencyFareURL  string
	AgencyEmail    string
}

// Stop represents a GTFS stops.txt row.
type Stop struct {
	StopID             string
	StopCode           string
	StopName           string
	StopDesc           string
	StopLat            float64
	StopLon            float64
	ZoneID             string
	StopURL            string
	LocationType       int // 0=stop/platform, 1=station, 2=entrance, 3=generic node, 4=boarding area
	ParentStation      string
	StopTimezone       string
	WheelchairBoarding int
	LevelID            string
	PlatformCode       string
}

// Route represents a GTFS routes.txt row.
type Route struct {
	RouteID           string
	AgencyID          string
	RouteShortName    string
	RouteLongName     string
	RouteDesc         string
	RouteType         int // 0=Tram, 1=Subway, 2=Rail, 3=Bus, ...
	RouteURL          string
	RouteColor        string
	RouteTextColor    string
	RouteSortOrder    int
	ContinuousPickup  int
	ContinuousDropOff int
	NetworkID         string
}

// Trip represents a GTFS trips.txt row.
type Trip struct {
	RouteID              string
	ServiceID            string
	TripID               string
	TripHeadsign         string
	TripShortName        string
	DirectionID          int
	BlockID              string
	ShapeID              string
	WheelchairAccessible int
	BikesAllowed         int
}

// StopTime represents a GTFS stop_times.txt row.
type StopTime struct {
	TripID            string
	ArrivalTime       GTFSTime
	DepartureTime     GTFSTime
	StopID            string
	StopSequence      int
	StopHeadsign      string
	PickupType        int
	DropOffType       int
	ContinuousPickup  int
	ContinuousDropOff int
	ShapeDistTraveled float64
	Timepoint         int
}

// CalendarEntry represents a GTFS calendar.txt row.
type CalendarEntry struct {
	ServiceID                                                  string
	Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday bool
	StartDate, EndDate                                         Date
}

// CalendarDate represents a GTFS calendar_dates.txt row.
type CalendarDate struct {
	ServiceID     string
	Date          Date
	ExceptionType int // 1=added, 2=removed
}

// Shape represents a single point in a GTFS shape.
type Shape struct {
	ShapeID           string
	ShapePtLat        float64
	ShapePtLon        float64
	ShapePtSequence   int
	ShapeDistTraveled float64
}

// Translation represents a GTFS translations.txt row.
type Translation struct {
	TableName    string
	FieldName    string
	Language     string
	TranslationText string
	RecordID     string
	RecordSubID  string
}

// --- Helpers ---

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}

func parseBool(s string) bool {
	return strings.TrimSpace(s) == "1"
}
