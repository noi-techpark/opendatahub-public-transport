package main

import (
	"strconv"
	"strings"
	"time"
)

// parseISO8601Time parses an ISO 8601 timestamp to time.Time.
func parseISO8601Time(s string) (time.Time, error) {
	// Try common formats
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

// reformatDate converts "YYYY-MM-DD" to "YYYYMMDD".
func reformatDate(s string) string {
	return strings.ReplaceAll(s, "-", "")
}

// parseFloat32 parses a string to float32.
func parseFloat32(s string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return float32(v)
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

// mapDirectionRef maps SIRI DirectionRef (1/2) to GTFS direction_id (0/1).
func mapDirectionRef(dirRef string) int {
	switch dirRef {
	case "1":
		return 0
	case "2":
		return 1
	default:
		v, _ := strconv.Atoi(dirRef)
		if v > 0 {
			return v - 1
		}
		return 0
	}
}

// mapCause maps SIRI AlertCause to GTFS-RT Cause enum string.
func mapCause(siriCause string) string {
	switch siriCause {
	case "constructionWork":
		return "CONSTRUCTION"
	case "maintenanceWork":
		return "MAINTENANCE"
	case "specialEvent":
		return "OTHER_CAUSE"
	case "accident":
		return "ACCIDENT"
	case "strike":
		return "STRIKE"
	case "weather":
		return "WEATHER"
	case "technicalProblem":
		return "TECHNICAL_PROBLEM"
	default:
		return "UNKNOWN_CAUSE"
	}
}

// mapEffect maps SIRI Consequence.Condition to GTFS-RT Effect enum string.
func mapEffect(condition string) string {
	switch condition {
	case "stopCancelled":
		return "NO_SERVICE"
	case "lineCancellation":
		return "NO_SERVICE"
	case "delayed":
		return "SIGNIFICANT_DELAYS"
	case "diverted":
		return "DETOUR"
	case "reducedService":
		return "REDUCED_SERVICE"
	case "additionalService":
		return "ADDITIONAL_SERVICE"
	case "modifiedService":
		return "MODIFIED_SERVICE"
	default:
		return "UNKNOWN_EFFECT"
	}
}

// mapSeverity maps SIRI Severity to GTFS-RT SeverityLevel enum string.
func mapSeverity(severity string) string {
	switch severity {
	case "severe":
		return "SEVERE"
	case "warning":
		return "WARNING"
	case "normal":
		return "INFO"
	default:
		return "UNKNOWN_SEVERITY"
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

	// Remove "PT" prefix
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

// mapLangCode maps SIRI language codes to ISO 639-1.
func mapLangCode(lang string) string {
	switch strings.ToUpper(lang) {
	case "DE":
		return "de"
	case "IT":
		return "it"
	case "EN":
		return "en"
	case "LAD":
		return "lld" // Ladin (ISO 639-3, but commonly used)
	default:
		return strings.ToLower(lang)
	}
}
