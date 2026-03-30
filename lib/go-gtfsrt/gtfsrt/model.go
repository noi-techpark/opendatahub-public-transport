// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package gtfsrt provides a complete GTFS-RT data model matching the protobuf schema.
// Reference: https://gtfs.org/documentation/realtime/reference/
package gtfsrt

// --- Core feed structure ---

type FeedMessage struct {
	Header FeedHeader   `json:"header"`
	Entity []FeedEntity `json:"entity"`
}

type FeedHeader struct {
	GtfsRealtimeVersion string `json:"gtfs_realtime_version"`
	Incrementality      string `json:"incrementality"` // FULL_DATASET | DIFFERENTIAL — REQUIRED per spec
	Timestamp           int64  `json:"timestamp"`
}

type FeedEntity struct {
	ID         string           `json:"id"`
	IsDeleted  bool             `json:"is_deleted,omitempty"`
	TripUpdate *TripUpdate      `json:"trip_update,omitempty"`
	Vehicle    *VehiclePosition `json:"vehicle,omitempty"`
	Alert      *Alert           `json:"alert,omitempty"`
}

// --- TripUpdate ---

type TripUpdate struct {
	Trip           *TripDescriptor    `json:"trip"`
	Vehicle        *VehicleDescriptor `json:"vehicle,omitempty"`
	StopTimeUpdate []StopTimeUpdate   `json:"stop_time_update,omitempty"`
	Timestamp      int64              `json:"timestamp,omitempty"`
	Delay          *int32             `json:"delay,omitempty"`
	TripProperties *TripProperties    `json:"trip_properties,omitempty"`
}

type StopTimeUpdate struct {
	StopSequence             int                  `json:"stop_sequence,omitempty"`
	StopID                   string               `json:"stop_id,omitempty"`
	Arrival                  *StopTimeEvent       `json:"arrival,omitempty"`
	Departure                *StopTimeEvent       `json:"departure,omitempty"`
	DepartureOccupancyStatus string               `json:"departure_occupancy_status,omitempty"`
	ScheduleRelationship     string               `json:"schedule_relationship,omitempty"` // SCHEDULED | SKIPPED | NO_DATA | UNSCHEDULED
	StopTimeProperties       *StopTimeProperties  `json:"stop_time_properties,omitempty"`
}

type StopTimeEvent struct {
	Delay       *int32 `json:"delay,omitempty"`         // seconds; 0 = on time. nil = not provided.
	Time        *int64 `json:"time,omitempty"`           // POSIX timestamp. nil = not provided.
	Uncertainty *int32 `json:"uncertainty,omitempty"`
}

type StopTimeProperties struct {
	AssignedStopID string `json:"assigned_stop_id,omitempty"`
}

type TripProperties struct {
	TripID    string `json:"trip_id,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	ShapeID   string `json:"shape_id,omitempty"`
}

// --- VehiclePosition ---

type VehiclePosition struct {
	Trip                *TripDescriptor    `json:"trip,omitempty"`
	Vehicle             *VehicleDescriptor `json:"vehicle,omitempty"`
	Position            *Position          `json:"position,omitempty"`
	CurrentStopSequence *int               `json:"current_stop_sequence,omitempty"`
	StopID              string             `json:"stop_id,omitempty"`
	CurrentStatus       string             `json:"current_status,omitempty"`  // INCOMING_AT | STOPPED_AT | IN_TRANSIT_TO
	Timestamp           int64              `json:"timestamp,omitempty"`
	CongestionLevel     string             `json:"congestion_level,omitempty"`
	OccupancyStatus     string             `json:"occupancy_status,omitempty"`
	OccupancyPercentage *int               `json:"occupancy_percentage,omitempty"`
	MultiCarriageDetails []CarriageDetails `json:"multi_carriage_details,omitempty"`
}

type CarriageDetails struct {
	ID                   string `json:"id,omitempty"`
	Label                string `json:"label,omitempty"`
	OccupancyStatus      string `json:"occupancy_status,omitempty"`
	OccupancyPercentage  *int   `json:"occupancy_percentage,omitempty"`
	CarriageSequence     *int   `json:"carriage_sequence,omitempty"`
}

// --- Alert ---

type Alert struct {
	ActivePeriod         []TimeRange       `json:"active_period,omitempty"`
	InformedEntity       []EntitySelector  `json:"informed_entity,omitempty"`
	Cause                string            `json:"cause,omitempty"`
	CauseDetail          *TranslatedString `json:"cause_detail,omitempty"`
	Effect               string            `json:"effect,omitempty"`
	EffectDetail         *TranslatedString `json:"effect_detail,omitempty"`
	URL                  *TranslatedString `json:"url,omitempty"`
	HeaderText           *TranslatedString `json:"header_text,omitempty"`
	DescriptionText      *TranslatedString `json:"description_text,omitempty"`
	TtsHeaderText        *TranslatedString `json:"tts_header_text,omitempty"`
	TtsDescriptionText   *TranslatedString `json:"tts_description_text,omitempty"`
	SeverityLevel        string            `json:"severity_level,omitempty"`
	Image                *TranslatedImage  `json:"image,omitempty"`
	ImageAlternativeText *TranslatedString `json:"image_alternative_text,omitempty"`
}

// --- Shared types ---

type TripDescriptor struct {
	TripID               string `json:"trip_id,omitempty"`
	RouteID              string `json:"route_id,omitempty"`
	DirectionID          *int   `json:"direction_id,omitempty"`
	StartTime            string `json:"start_time,omitempty"`
	StartDate            string `json:"start_date,omitempty"`
	ScheduleRelationship string `json:"schedule_relationship,omitempty"` // SCHEDULED | ADDED | UNSCHEDULED | CANCELED | REPLACEMENT | DUPLICATED | DELETED
}

type VehicleDescriptor struct {
	ID                   string `json:"id,omitempty"`
	Label                string `json:"label,omitempty"`
	LicensePlate         string `json:"license_plate,omitempty"`
	WheelchairAccessible string `json:"wheelchair_accessible,omitempty"` // NO_VALUE | UNKNOWN | WHEELCHAIR_ACCESSIBLE | WHEELCHAIR_INACCESSIBLE
}

type Position struct {
	Latitude  float32  `json:"latitude"`
	Longitude float32  `json:"longitude"`
	Bearing   *float32 `json:"bearing,omitempty"`
	Odometer  *float64 `json:"odometer,omitempty"`
	Speed     *float32 `json:"speed,omitempty"`
}

type TimeRange struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

type EntitySelector struct {
	AgencyID    string          `json:"agency_id,omitempty"`
	RouteID     string          `json:"route_id,omitempty"`
	RouteType   *int            `json:"route_type,omitempty"`
	DirectionID *int            `json:"direction_id,omitempty"`
	Trip        *TripDescriptor `json:"trip,omitempty"`
	StopID      string          `json:"stop_id,omitempty"`
}

type TranslatedString struct {
	Translation []Translation `json:"translation"`
}

type Translation struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
}

type TranslatedImage struct {
	LocalizedImage []LocalizedImage `json:"localized_image"`
}

type LocalizedImage struct {
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	Language  string `json:"language,omitempty"`
}
