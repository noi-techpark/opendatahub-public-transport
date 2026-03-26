// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfsrt

// GTFS-RT output model — JSON-serializable structs matching the protobuf schema.

type FeedMessage struct {
	Header FeedHeader   `json:"header"`
	Entity []FeedEntity `json:"entity"`
}

type FeedHeader struct {
	GtfsRealtimeVersion string `json:"gtfs_realtime_version"`
	Incrementality      string `json:"incrementality,omitempty"`
	Timestamp           int64  `json:"timestamp"`
}

type FeedEntity struct {
	ID         string           `json:"id"`
	TripUpdate *TripUpdate      `json:"trip_update,omitempty"`
	Vehicle    *VehiclePosition `json:"vehicle,omitempty"`
	Alert      *Alert           `json:"alert,omitempty"`
}

type TripUpdate struct {
	Trip           *TripDescriptor  `json:"trip"`
	StopTimeUpdate []StopTimeUpdate `json:"stop_time_update,omitempty"`
	Timestamp      int64            `json:"timestamp,omitempty"`
}

type StopTimeUpdate struct {
	StopSequence int            `json:"stop_sequence,omitempty"`
	StopID       string         `json:"stop_id,omitempty"`
	Arrival      *StopTimeEvent `json:"arrival,omitempty"`
	Departure    *StopTimeEvent `json:"departure,omitempty"`
}

type StopTimeEvent struct {
	Delay int32 `json:"delay,omitempty"` // seconds
	Time  int64 `json:"time,omitempty"`  // POSIX timestamp
}

type VehiclePosition struct {
	Trip            *TripDescriptor    `json:"trip,omitempty"`
	Vehicle         *VehicleDescriptor `json:"vehicle,omitempty"`
	Position        *Position          `json:"position,omitempty"`
	StopID          string             `json:"stop_id,omitempty"`
	CurrentStatus   string             `json:"current_status,omitempty"`
	Timestamp       int64              `json:"timestamp,omitempty"`
	CongestionLevel string             `json:"congestion_level,omitempty"`
}

type TripDescriptor struct {
	TripID               string `json:"trip_id,omitempty"`
	RouteID              string `json:"route_id,omitempty"`
	DirectionID          *int   `json:"direction_id,omitempty"`
	StartTime            string `json:"start_time,omitempty"`
	StartDate            string `json:"start_date,omitempty"`
	ScheduleRelationship string `json:"schedule_relationship,omitempty"`
}

type VehicleDescriptor struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

type Position struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type Alert struct {
	ActivePeriod    []TimeRange       `json:"active_period,omitempty"`
	InformedEntity  []EntitySelector  `json:"informed_entity,omitempty"`
	Cause           string            `json:"cause,omitempty"`
	Effect          string            `json:"effect,omitempty"`
	HeaderText      *TranslatedString `json:"header_text,omitempty"`
	DescriptionText *TranslatedString `json:"description_text,omitempty"`
	SeverityLevel   string            `json:"severity_level,omitempty"`
}

type TimeRange struct {
	Start int64 `json:"start,omitempty"`
	End   int64 `json:"end,omitempty"`
}

type EntitySelector struct {
	RouteID     string `json:"route_id,omitempty"`
	StopID      string `json:"stop_id,omitempty"`
	DirectionID *int   `json:"direction_id,omitempty"`
}

type TranslatedString struct {
	Translation []Translation `json:"translation"`
}

type Translation struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
}
