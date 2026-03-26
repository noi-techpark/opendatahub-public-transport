package gtfsrt

import (
	"fmt"

	pb "github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/pb"
	"google.golang.org/protobuf/proto"
)

// --- Serialize to protobuf ---

func serializeProtobuf(fm *FeedMessage) ([]byte, error) {
	msg := toProto(fm)
	return proto.Marshal(msg)
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

func vehiclePositionToProto(vp *VehiclePosition) *pb.VehiclePosition {
	p := &pb.VehiclePosition{}
	p.Trip = tripDescriptorToProto(vp.Trip)
	if vp.Vehicle != nil {
		p.Vehicle = &pb.VehicleDescriptor{}
		if vp.Vehicle.ID != "" {
			p.Vehicle.Id = proto.String(vp.Vehicle.ID)
		}
		if vp.Vehicle.Label != "" {
			p.Vehicle.Label = proto.String(vp.Vehicle.Label)
		}
	}
	if vp.Position != nil {
		p.Position = &pb.Position{
			Latitude:  proto.Float32(vp.Position.Latitude),
			Longitude: proto.Float32(vp.Position.Longitude),
		}
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
		if ie.RouteID != "" {
			es.RouteId = proto.String(ie.RouteID)
		}
		if ie.StopID != "" {
			es.StopId = proto.String(ie.StopID)
		}
		if ie.DirectionID != nil {
			es.DirectionId = proto.Uint32(uint32(*ie.DirectionID))
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
	if tu.Timestamp > 0 {
		p.Timestamp = proto.Uint64(uint64(tu.Timestamp))
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
			pstu.Arrival = &pb.TripUpdate_StopTimeEvent{}
			if stu.Arrival.Delay != 0 {
				pstu.Arrival.Delay = proto.Int32(stu.Arrival.Delay)
			}
			if stu.Arrival.Time > 0 {
				pstu.Arrival.Time = proto.Int64(stu.Arrival.Time)
			}
		}
		if stu.Departure != nil {
			pstu.Departure = &pb.TripUpdate_StopTimeEvent{}
			if stu.Departure.Delay != 0 {
				pstu.Departure.Delay = proto.Int32(stu.Departure.Delay)
			}
			if stu.Departure.Time > 0 {
				pstu.Departure.Time = proto.Int64(stu.Departure.Time)
			}
		}
		p.StopTimeUpdate = append(p.StopTimeUpdate, pstu)
	}
	return p
}

func translatedStringToProto(ts *TranslatedString) *pb.TranslatedString {
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
	fe := FeedEntity{ID: e.GetId()}
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

func vehiclePositionFromProto(vp *pb.VehiclePosition) *VehiclePosition {
	p := &VehiclePosition{
		StopID:    vp.GetStopId(),
		Timestamp: int64(vp.GetTimestamp()),
	}
	p.Trip = tripDescriptorFromProto(vp.Trip)
	if vp.Vehicle != nil {
		p.Vehicle = &VehicleDescriptor{
			ID:    vp.Vehicle.GetId(),
			Label: vp.Vehicle.GetLabel(),
		}
	}
	if vp.Position != nil {
		p.Position = &Position{
			Latitude:  vp.Position.GetLatitude(),
			Longitude: vp.Position.GetLongitude(),
		}
	}
	if vp.CurrentStatus != nil {
		p.CurrentStatus = vp.CurrentStatus.String()
	}
	if vp.CongestionLevel != nil {
		p.CongestionLevel = vp.CongestionLevel.String()
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
			RouteID: ie.GetRouteId(),
			StopID:  ie.GetStopId(),
		}
		if ie.DirectionId != nil {
			v := int(ie.GetDirectionId())
			es.DirectionID = &v
		}
		p.InformedEntity = append(p.InformedEntity, es)
	}
	if a.Cause != nil {
		p.Cause = a.Cause.String()
	}
	if a.Effect != nil {
		p.Effect = a.Effect.String()
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
		Timestamp: int64(tu.GetTimestamp()),
	}
	for _, stu := range tu.StopTimeUpdate {
		s := StopTimeUpdate{
			StopSequence: int(stu.GetStopSequence()),
			StopID:       stu.GetStopId(),
		}
		if stu.Arrival != nil {
			s.Arrival = &StopTimeEvent{
				Delay: stu.Arrival.GetDelay(),
				Time:  stu.Arrival.GetTime(),
			}
		}
		if stu.Departure != nil {
			s.Departure = &StopTimeEvent{
				Delay: stu.Departure.GetDelay(),
				Time:  stu.Departure.GetTime(),
			}
		}
		p.StopTimeUpdate = append(p.StopTimeUpdate, s)
	}
	return p
}

func translatedStringFromProto(ts *pb.TranslatedString) *TranslatedString {
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
	default:
		return pb.TripDescriptor_SCHEDULED
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
