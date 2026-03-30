// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package gtfsrt

import (
	"fmt"

	pb "github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/pb"
	"google.golang.org/protobuf/proto"
)

// --- Serialize to protobuf ---

func serializeProtobuf(fm *FeedMessage) ([]byte, error) {
	return proto.Marshal(toProto(fm))
}

func toProto(fm *FeedMessage) *pb.FeedMessage {
	header := &pb.FeedHeader{
		GtfsRealtimeVersion: proto.String(fm.Header.GtfsRealtimeVersion),
		Timestamp:           proto.Uint64(uint64(fm.Header.Timestamp)),
	}
	if fm.Header.Incrementality == "DIFFERENTIAL" {
		inc := pb.FeedHeader_DIFFERENTIAL
		header.Incrementality = &inc
	} else {
		inc := pb.FeedHeader_FULL_DATASET
		header.Incrementality = &inc
	}

	msg := &pb.FeedMessage{Header: header}
	for _, e := range fm.Entity {
		msg.Entity = append(msg.Entity, entityToProto(&e))
	}
	return msg
}

func entityToProto(e *FeedEntity) *pb.FeedEntity {
	fe := &pb.FeedEntity{
		Id: proto.String(e.ID),
	}
	if e.IsDeleted {
		fe.IsDeleted = proto.Bool(true)
	}
	if e.Vehicle != nil {
		fe.Vehicle = vehiclePositionToProto(e.Vehicle)
	}
	if e.Alert != nil {
		fe.Alert = alertToProto(e.Alert)
	}
	if e.TripUpdate != nil {
		fe.TripUpdate = tripUpdateToProto(e.TripUpdate)
	}
	return fe
}

func tripDescriptorToProto(td *TripDescriptor) *pb.TripDescriptor {
	if td == nil {
		return nil
	}
	p := &pb.TripDescriptor{}
	if td.TripID != "" {
		p.TripId = proto.String(td.TripID)
	}
	if td.RouteID != "" {
		p.RouteId = proto.String(td.RouteID)
	}
	if td.DirectionID != nil {
		p.DirectionId = proto.Uint32(uint32(*td.DirectionID))
	}
	if td.StartTime != "" {
		p.StartTime = proto.String(td.StartTime)
	}
	if td.StartDate != "" {
		p.StartDate = proto.String(td.StartDate)
	}
	if td.ScheduleRelationship != "" {
		sr := scheduleRelationshipToProto(td.ScheduleRelationship)
		p.ScheduleRelationship = &sr
	}
	return p
}

func vehicleDescriptorToProto(v *VehicleDescriptor) *pb.VehicleDescriptor {
	if v == nil {
		return nil
	}
	p := &pb.VehicleDescriptor{}
	if v.ID != "" {
		p.Id = proto.String(v.ID)
	}
	if v.Label != "" {
		p.Label = proto.String(v.Label)
	}
	if v.LicensePlate != "" {
		p.LicensePlate = proto.String(v.LicensePlate)
	}
	return p
}

func vehiclePositionToProto(vp *VehiclePosition) *pb.VehiclePosition {
	p := &pb.VehiclePosition{}
	p.Trip = tripDescriptorToProto(vp.Trip)
	p.Vehicle = vehicleDescriptorToProto(vp.Vehicle)
	if vp.Position != nil {
		p.Position = &pb.Position{
			Latitude:  proto.Float32(vp.Position.Latitude),
			Longitude: proto.Float32(vp.Position.Longitude),
		}
		if vp.Position.Bearing != nil {
			p.Position.Bearing = proto.Float32(*vp.Position.Bearing)
		}
		if vp.Position.Odometer != nil {
			p.Position.Odometer = proto.Float64(*vp.Position.Odometer)
		}
		if vp.Position.Speed != nil {
			p.Position.Speed = proto.Float32(*vp.Position.Speed)
		}
	}
	if vp.CurrentStopSequence != nil {
		p.CurrentStopSequence = proto.Uint32(uint32(*vp.CurrentStopSequence))
	}
	if vp.StopID != "" {
		p.StopId = proto.String(vp.StopID)
	}
	if vp.CurrentStatus != "" {
		cs := currentStatusToProto(vp.CurrentStatus)
		p.CurrentStatus = &cs
	}
	if vp.Timestamp > 0 {
		p.Timestamp = proto.Uint64(uint64(vp.Timestamp))
	}
	if vp.CongestionLevel != "" {
		cl := congestionLevelToProto(vp.CongestionLevel)
		p.CongestionLevel = &cl
	}
	if vp.OccupancyStatus != "" {
		os := occupancyStatusToProto(vp.OccupancyStatus)
		p.OccupancyStatus = &os
	}
	if vp.OccupancyPercentage != nil {
		p.OccupancyPercentage = proto.Uint32(uint32(*vp.OccupancyPercentage))
	}
	return p
}

func alertToProto(a *Alert) *pb.Alert {
	p := &pb.Alert{}
	for _, ap := range a.ActivePeriod {
		tr := &pb.TimeRange{}
		if ap.Start > 0 {
			tr.Start = proto.Uint64(uint64(ap.Start))
		}
		if ap.End > 0 {
			tr.End = proto.Uint64(uint64(ap.End))
		}
		p.ActivePeriod = append(p.ActivePeriod, tr)
	}
	for _, ie := range a.InformedEntity {
		es := &pb.EntitySelector{}
		if ie.AgencyID != "" {
			es.AgencyId = proto.String(ie.AgencyID)
		}
		if ie.RouteID != "" {
			es.RouteId = proto.String(ie.RouteID)
		}
		if ie.RouteType != nil {
			es.RouteType = proto.Int32(int32(*ie.RouteType))
		}
		if ie.DirectionID != nil {
			es.DirectionId = proto.Uint32(uint32(*ie.DirectionID))
		}
		if ie.Trip != nil {
			es.Trip = tripDescriptorToProto(ie.Trip)
		}
		if ie.StopID != "" {
			es.StopId = proto.String(ie.StopID)
		}
		p.InformedEntity = append(p.InformedEntity, es)
	}
	if a.Cause != "" {
		c := causeToProto(a.Cause)
		p.Cause = &c
	}
	if a.Effect != "" {
		e := effectToProto(a.Effect)
		p.Effect = &e
	}
	if a.URL != nil {
		p.Url = translatedStringToProto(a.URL)
	}
	if a.HeaderText != nil {
		p.HeaderText = translatedStringToProto(a.HeaderText)
	}
	if a.DescriptionText != nil {
		p.DescriptionText = translatedStringToProto(a.DescriptionText)
	}
	if a.SeverityLevel != "" {
		sl := severityToProto(a.SeverityLevel)
		p.SeverityLevel = &sl
	}
	return p
}

func tripUpdateToProto(tu *TripUpdate) *pb.TripUpdate {
	p := &pb.TripUpdate{
		Trip: tripDescriptorToProto(tu.Trip),
	}
	p.Vehicle = vehicleDescriptorToProto(tu.Vehicle)
	if tu.Timestamp > 0 {
		p.Timestamp = proto.Uint64(uint64(tu.Timestamp))
	}
	if tu.Delay != nil {
		p.Delay = proto.Int32(*tu.Delay)
	}
	for _, stu := range tu.StopTimeUpdate {
		pstu := &pb.TripUpdate_StopTimeUpdate{}
		if stu.StopID != "" {
			pstu.StopId = proto.String(stu.StopID)
		}
		if stu.StopSequence > 0 {
			pstu.StopSequence = proto.Uint32(uint32(stu.StopSequence))
		}
		if stu.Arrival != nil {
			pstu.Arrival = stopTimeEventToProto(stu.Arrival)
		}
		if stu.Departure != nil {
			pstu.Departure = stopTimeEventToProto(stu.Departure)
		}
		if stu.ScheduleRelationship != "" {
			sr := stuScheduleRelationshipToProto(stu.ScheduleRelationship)
			pstu.ScheduleRelationship = &sr
		}
		p.StopTimeUpdate = append(p.StopTimeUpdate, pstu)
	}
	return p
}

func stopTimeEventToProto(e *StopTimeEvent) *pb.TripUpdate_StopTimeEvent {
	p := &pb.TripUpdate_StopTimeEvent{}
	if e.Delay != 0 {
		p.Delay = proto.Int32(e.Delay)
	}
	if e.Time > 0 {
		p.Time = proto.Int64(e.Time)
	}
	if e.Uncertainty != nil {
		p.Uncertainty = proto.Int32(*e.Uncertainty)
	}
	return p
}

func translatedStringToProto(ts *TranslatedString) *pb.TranslatedString {
	if ts == nil {
		return nil
	}
	p := &pb.TranslatedString{}
	for _, t := range ts.Translation {
		pt := &pb.TranslatedString_Translation{
			Text: proto.String(t.Text),
		}
		if t.Language != "" {
			pt.Language = proto.String(t.Language)
		}
		p.Translation = append(p.Translation, pt)
	}
	return p
}

// --- Deserialize from protobuf ---

func deserializeProtobuf(data []byte) (*FeedMessage, error) {
	var msg pb.FeedMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("parse GTFS-RT protobuf: %w", err)
	}
	return fromProto(&msg), nil
}

func fromProto(msg *pb.FeedMessage) *FeedMessage {
	fm := &FeedMessage{
		Header: FeedHeader{
			GtfsRealtimeVersion: msg.Header.GetGtfsRealtimeVersion(),
			Timestamp:           int64(msg.Header.GetTimestamp()),
		},
	}
	if msg.Header.Incrementality != nil {
		if *msg.Header.Incrementality == pb.FeedHeader_DIFFERENTIAL {
			fm.Header.Incrementality = "DIFFERENTIAL"
		} else {
			fm.Header.Incrementality = "FULL_DATASET"
		}
	}
	for _, e := range msg.Entity {
		fm.Entity = append(fm.Entity, entityFromProto(e))
	}
	return fm
}

func entityFromProto(e *pb.FeedEntity) FeedEntity {
	fe := FeedEntity{
		ID:        e.GetId(),
		IsDeleted: e.GetIsDeleted(),
	}
	if e.Vehicle != nil {
		fe.Vehicle = vehiclePositionFromProto(e.Vehicle)
	}
	if e.Alert != nil {
		fe.Alert = alertFromProto(e.Alert)
	}
	if e.TripUpdate != nil {
		fe.TripUpdate = tripUpdateFromProto(e.TripUpdate)
	}
	return fe
}

func tripDescriptorFromProto(td *pb.TripDescriptor) *TripDescriptor {
	if td == nil {
		return nil
	}
	p := &TripDescriptor{
		TripID:    td.GetTripId(),
		RouteID:   td.GetRouteId(),
		StartTime: td.GetStartTime(),
		StartDate: td.GetStartDate(),
	}
	if td.DirectionId != nil {
		v := int(td.GetDirectionId())
		p.DirectionID = &v
	}
	if td.ScheduleRelationship != nil {
		p.ScheduleRelationship = td.ScheduleRelationship.String()
	}
	return p
}

func vehicleDescriptorFromProto(v *pb.VehicleDescriptor) *VehicleDescriptor {
	if v == nil {
		return nil
	}
	return &VehicleDescriptor{
		ID:           v.GetId(),
		Label:        v.GetLabel(),
		LicensePlate: v.GetLicensePlate(),
	}
}

func vehiclePositionFromProto(vp *pb.VehiclePosition) *VehiclePosition {
	p := &VehiclePosition{
		StopID:    vp.GetStopId(),
		Timestamp: int64(vp.GetTimestamp()),
	}
	p.Trip = tripDescriptorFromProto(vp.Trip)
	p.Vehicle = vehicleDescriptorFromProto(vp.Vehicle)
	if vp.Position != nil {
		p.Position = &Position{
			Latitude:  vp.Position.GetLatitude(),
			Longitude: vp.Position.GetLongitude(),
		}
		if vp.Position.Bearing != nil {
			v := vp.Position.GetBearing()
			p.Position.Bearing = &v
		}
		if vp.Position.Odometer != nil {
			v := vp.Position.GetOdometer()
			p.Position.Odometer = &v
		}
		if vp.Position.Speed != nil {
			v := vp.Position.GetSpeed()
			p.Position.Speed = &v
		}
	}
	if vp.CurrentStopSequence != nil {
		v := int(vp.GetCurrentStopSequence())
		p.CurrentStopSequence = &v
	}
	if vp.CurrentStatus != nil {
		p.CurrentStatus = vp.CurrentStatus.String()
	}
	if vp.CongestionLevel != nil {
		p.CongestionLevel = vp.CongestionLevel.String()
	}
	if vp.OccupancyStatus != nil {
		p.OccupancyStatus = vp.OccupancyStatus.String()
	}
	if vp.OccupancyPercentage != nil {
		v := int(vp.GetOccupancyPercentage())
		p.OccupancyPercentage = &v
	}
	return p
}

func alertFromProto(a *pb.Alert) *Alert {
	p := &Alert{}
	for _, ap := range a.ActivePeriod {
		p.ActivePeriod = append(p.ActivePeriod, TimeRange{
			Start: int64(ap.GetStart()),
			End:   int64(ap.GetEnd()),
		})
	}
	for _, ie := range a.InformedEntity {
		es := EntitySelector{
			AgencyID: ie.GetAgencyId(),
			RouteID:  ie.GetRouteId(),
			StopID:   ie.GetStopId(),
		}
		if ie.RouteType != nil {
			v := int(ie.GetRouteType())
			es.RouteType = &v
		}
		if ie.DirectionId != nil {
			v := int(ie.GetDirectionId())
			es.DirectionID = &v
		}
		if ie.Trip != nil {
			es.Trip = tripDescriptorFromProto(ie.Trip)
		}
		p.InformedEntity = append(p.InformedEntity, es)
	}
	if a.Cause != nil {
		p.Cause = a.Cause.String()
	}
	if a.Effect != nil {
		p.Effect = a.Effect.String()
	}
	if a.Url != nil {
		p.URL = translatedStringFromProto(a.Url)
	}
	if a.HeaderText != nil {
		p.HeaderText = translatedStringFromProto(a.HeaderText)
	}
	if a.DescriptionText != nil {
		p.DescriptionText = translatedStringFromProto(a.DescriptionText)
	}
	if a.SeverityLevel != nil {
		p.SeverityLevel = a.SeverityLevel.String()
	}
	return p
}

func tripUpdateFromProto(tu *pb.TripUpdate) *TripUpdate {
	p := &TripUpdate{
		Trip:      tripDescriptorFromProto(tu.Trip),
		Vehicle:   vehicleDescriptorFromProto(tu.Vehicle),
		Timestamp: int64(tu.GetTimestamp()),
	}
	if tu.Delay != nil {
		v := tu.GetDelay()
		p.Delay = &v
	}
	for _, stu := range tu.StopTimeUpdate {
		s := StopTimeUpdate{
			StopSequence: int(stu.GetStopSequence()),
			StopID:       stu.GetStopId(),
		}
		if stu.Arrival != nil {
			s.Arrival = stopTimeEventFromProto(stu.Arrival)
		}
		if stu.Departure != nil {
			s.Departure = stopTimeEventFromProto(stu.Departure)
		}
		if stu.ScheduleRelationship != nil {
			s.ScheduleRelationship = stu.ScheduleRelationship.String()
		}
		p.StopTimeUpdate = append(p.StopTimeUpdate, s)
	}
	return p
}

func stopTimeEventFromProto(e *pb.TripUpdate_StopTimeEvent) *StopTimeEvent {
	p := &StopTimeEvent{
		Delay: e.GetDelay(),
		Time:  e.GetTime(),
	}
	if e.Uncertainty != nil {
		v := e.GetUncertainty()
		p.Uncertainty = &v
	}
	return p
}

func translatedStringFromProto(ts *pb.TranslatedString) *TranslatedString {
	if ts == nil {
		return nil
	}
	p := &TranslatedString{}
	for _, t := range ts.Translation {
		p.Translation = append(p.Translation, Translation{
			Text:     t.GetText(),
			Language: t.GetLanguage(),
		})
	}
	return p
}

// --- Enum mappings ---

func scheduleRelationshipToProto(s string) pb.TripDescriptor_ScheduleRelationship {
	switch s {
	case "SCHEDULED":
		return pb.TripDescriptor_SCHEDULED
	case "ADDED":
		return pb.TripDescriptor_ADDED
	case "UNSCHEDULED":
		return pb.TripDescriptor_UNSCHEDULED
	case "CANCELED":
		return pb.TripDescriptor_CANCELED
	case "REPLACEMENT":
		return pb.TripDescriptor_REPLACEMENT
	case "DUPLICATED":
		return pb.TripDescriptor_DUPLICATED
	case "DELETED":
		return pb.TripDescriptor_DELETED
	default:
		return pb.TripDescriptor_SCHEDULED
	}
}

func stuScheduleRelationshipToProto(s string) pb.TripUpdate_StopTimeUpdate_ScheduleRelationship {
	switch s {
	case "SCHEDULED":
		return pb.TripUpdate_StopTimeUpdate_SCHEDULED
	case "SKIPPED":
		return pb.TripUpdate_StopTimeUpdate_SKIPPED
	case "NO_DATA":
		return pb.TripUpdate_StopTimeUpdate_NO_DATA
	case "UNSCHEDULED":
		return pb.TripUpdate_StopTimeUpdate_UNSCHEDULED
	default:
		return pb.TripUpdate_StopTimeUpdate_SCHEDULED
	}
}

func currentStatusToProto(s string) pb.VehiclePosition_VehicleStopStatus {
	switch s {
	case "INCOMING_AT":
		return pb.VehiclePosition_INCOMING_AT
	case "STOPPED_AT":
		return pb.VehiclePosition_STOPPED_AT
	case "IN_TRANSIT_TO":
		return pb.VehiclePosition_IN_TRANSIT_TO
	default:
		return pb.VehiclePosition_IN_TRANSIT_TO
	}
}

func congestionLevelToProto(s string) pb.VehiclePosition_CongestionLevel {
	switch s {
	case "RUNNING_SMOOTHLY":
		return pb.VehiclePosition_RUNNING_SMOOTHLY
	case "STOP_AND_GO":
		return pb.VehiclePosition_STOP_AND_GO
	case "CONGESTION":
		return pb.VehiclePosition_CONGESTION
	case "SEVERE_CONGESTION":
		return pb.VehiclePosition_SEVERE_CONGESTION
	default:
		return pb.VehiclePosition_UNKNOWN_CONGESTION_LEVEL
	}
}

func occupancyStatusToProto(s string) pb.VehiclePosition_OccupancyStatus {
	switch s {
	case "EMPTY":
		return pb.VehiclePosition_EMPTY
	case "MANY_SEATS_AVAILABLE":
		return pb.VehiclePosition_MANY_SEATS_AVAILABLE
	case "FEW_SEATS_AVAILABLE":
		return pb.VehiclePosition_FEW_SEATS_AVAILABLE
	case "STANDING_ROOM_ONLY":
		return pb.VehiclePosition_STANDING_ROOM_ONLY
	case "CRUSHED_STANDING_ROOM_ONLY":
		return pb.VehiclePosition_CRUSHED_STANDING_ROOM_ONLY
	case "FULL":
		return pb.VehiclePosition_FULL
	case "NOT_ACCEPTING_PASSENGERS":
		return pb.VehiclePosition_NOT_ACCEPTING_PASSENGERS
	case "NO_DATA_AVAILABLE":
		return pb.VehiclePosition_NO_DATA_AVAILABLE
	case "NOT_BOARDABLE":
		return pb.VehiclePosition_NOT_BOARDABLE
	default:
		return pb.VehiclePosition_NO_DATA_AVAILABLE
	}
}

func causeToProto(s string) pb.Alert_Cause {
	switch s {
	case "UNKNOWN_CAUSE":
		return pb.Alert_UNKNOWN_CAUSE
	case "OTHER_CAUSE":
		return pb.Alert_OTHER_CAUSE
	case "TECHNICAL_PROBLEM":
		return pb.Alert_TECHNICAL_PROBLEM
	case "STRIKE":
		return pb.Alert_STRIKE
	case "DEMONSTRATION":
		return pb.Alert_DEMONSTRATION
	case "ACCIDENT":
		return pb.Alert_ACCIDENT
	case "HOLIDAY":
		return pb.Alert_HOLIDAY
	case "WEATHER":
		return pb.Alert_WEATHER
	case "MAINTENANCE":
		return pb.Alert_MAINTENANCE
	case "CONSTRUCTION":
		return pb.Alert_CONSTRUCTION
	case "POLICE_ACTIVITY":
		return pb.Alert_POLICE_ACTIVITY
	case "MEDICAL_EMERGENCY":
		return pb.Alert_MEDICAL_EMERGENCY
	default:
		return pb.Alert_UNKNOWN_CAUSE
	}
}

func effectToProto(s string) pb.Alert_Effect {
	switch s {
	case "NO_SERVICE":
		return pb.Alert_NO_SERVICE
	case "REDUCED_SERVICE":
		return pb.Alert_REDUCED_SERVICE
	case "SIGNIFICANT_DELAYS":
		return pb.Alert_SIGNIFICANT_DELAYS
	case "DETOUR":
		return pb.Alert_DETOUR
	case "ADDITIONAL_SERVICE":
		return pb.Alert_ADDITIONAL_SERVICE
	case "MODIFIED_SERVICE":
		return pb.Alert_MODIFIED_SERVICE
	case "OTHER_EFFECT":
		return pb.Alert_OTHER_EFFECT
	case "STOP_MOVED":
		return pb.Alert_STOP_MOVED
	case "NO_EFFECT":
		return pb.Alert_NO_EFFECT
	case "ACCESSIBILITY_ISSUE":
		return pb.Alert_ACCESSIBILITY_ISSUE
	default:
		return pb.Alert_UNKNOWN_EFFECT
	}
}

func severityToProto(s string) pb.Alert_SeverityLevel {
	switch s {
	case "INFO":
		return pb.Alert_INFO
	case "WARNING":
		return pb.Alert_WARNING
	case "SEVERE":
		return pb.Alert_SEVERE
	default:
		return pb.Alert_UNKNOWN_SEVERITY
	}
}
