// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package sta

import (
	"strconv"
	"strings"
	"time"
)

// ParseISO8601Time parses an ISO 8601 timestamp to time.Time.
func ParseISO8601Time(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05+07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{Value: s, Message: "unrecognized ISO 8601 format"}
}

// ReformatDate converts "YYYY-MM-DD" to "YYYYMMDD".
func ReformatDate(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

// stripNeTExStopPrefix removes the NeTEx ScheduledStopPoint prefix from a StopPointRef.
// "it:apb:ScheduledStopPoint:it:22021:240:0:337" → "it:22021:240:0:337"
// "it:22021:240:0:337" → "it:22021:240:0:337" (unchanged)
func stripNeTExStopPrefix(ref string) string {
	const prefix = "it:apb:ScheduledStopPoint:"
	if strings.HasPrefix(ref, prefix) {
		return ref[len(prefix):]
	}
	return ref
}

// MapDirectionRef maps SIRI DirectionRef (1/2) to GTFS direction_id (0/1).
// STA convention: SIRI 1 = Rückfahrt (return) = GTFS 1, SIRI 2 = Hinfahrt
// (outbound) = GTFS 0.
func MapDirectionRef(dirRef string) int {
	switch dirRef {
	case "1":
		return 1
	case "2":
		return 0
	default:
		v, _ := strconv.Atoi(dirRef)
		if v > 0 {
			return v - 1
		}
		return 0
	}
}

// parseISO8601Duration parses an ISO 8601 duration like "PT1M15S", "-PT2M36S" to time.Duration.
func parseISO8601Duration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	s = strings.TrimPrefix(s, "PT")
	s = strings.TrimPrefix(s, "P")
	s = strings.TrimPrefix(s, "T")

	var totalSeconds int
	num := ""
	for _, c := range s {
		switch c {
		case 'H':
			v, _ := strconv.Atoi(num)
			totalSeconds += v * 3600
			num = ""
		case 'M':
			v, _ := strconv.Atoi(num)
			totalSeconds += v * 60
			num = ""
		case 'S':
			v, _ := strconv.Atoi(num)
			totalSeconds += v
			num = ""
		default:
			num += string(c)
		}
	}

	d := time.Duration(totalSeconds) * time.Second
	if negative {
		d = -d
	}
	return d
}
