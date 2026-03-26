// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import "fmt"

// --- Estimated Timetable structs ---

type ETFeed struct {
	ServiceDelivery struct {
		ResponseTimestamp          string                         `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
		ProducerRef                string                         `json:"ProducerRef" xml:"ProducerRef"`
		EstimatedTimetableDelivery EstimatedTimetableDelivery     `json:"EstimatedTimetableDelivery" xml:"EstimatedTimetableDelivery"`
	} `json:"ServiceDelivery" xml:"ServiceDelivery"`
}

type EstimatedTimetableDelivery struct {
	ResponseTimestamp            string                          `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
	EstimatedJourneyVersionFrame []EstimatedJourneyVersionFrame  `json:"EstimatedJourneyVersionFrame" xml:"EstimatedJourneyVersionFrame"`
}

type EstimatedJourneyVersionFrame struct {
	RecordedAtTime          string                    `json:"RecordedAtTime" xml:"RecordedAtTime"`
	EstimatedVehicleJourney []EstimatedVehicleJourney `json:"EstimatedVehicleJourney" xml:"EstimatedVehicleJourney"`
}

type EstimatedVehicleJourney struct {
	RecordedAtTime              string                  `json:"RecordedAtTime" xml:"RecordedAtTime"`
	LineRef                     string                  `json:"LineRef" xml:"LineRef"`
	DirectionRef                string                  `json:"DirectionRef" xml:"DirectionRef"`
	FramedVehicleJourneyRef     FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef" xml:"FramedVehicleJourneyRef"`
	PublishedLineName           string                  `json:"PublishedLineName" xml:"PublishedLineName"`
	DirectionName               string                  `json:"DirectionName" xml:"DirectionName"`
	OperatorRef                 string                  `json:"OperatorRef" xml:"OperatorRef"`
	ProductCategoryRef          string                  `json:"ProductCategoryRef" xml:"ProductCategoryRef"`
	OriginRef                   string                  `json:"OriginRef" xml:"OriginRef"`
	DestinationRef              string                  `json:"DestinationRef" xml:"DestinationRef"`
	OriginAimedDepartureTime    string                  `json:"OriginAimedDepartureTime" xml:"OriginAimedDepartureTime"`
	DestinationAimedArrivalTime string                  `json:"DestinationAimedArrivalTime" xml:"DestinationAimedArrivalTime"`
	Monitored                   string                  `json:"Monitored" xml:"Monitored"`
	MonitoringError             string                  `json:"MonitoringError,omitempty" xml:"MonitoringError,omitempty"`
	BlockRef                    string                  `json:"BlockRef,omitempty" xml:"BlockRef,omitempty"`
	EstimatedCalls              EstimatedCalls          `json:"EstimatedCalls" xml:"EstimatedCalls"`
	IsCompleteStopSequence      string                  `json:"IsCompleteStopSequence,omitempty" xml:"IsCompleteStopSequence,omitempty"`
}

type EstimatedCalls struct {
	EstimatedCall []EstimatedCall `json:"EstimatedCall" xml:"EstimatedCall"`
}

type EstimatedCall struct {
	StopPointRef              string `json:"StopPointRef" xml:"StopPointRef"`
	VisitNumber               string `json:"VisitNumber,omitempty" xml:"VisitNumber,omitempty"`
	StopPointName             string `json:"StopPointName,omitempty" xml:"StopPointName,omitempty"`
	DestinationDisplay        string `json:"DestinationDisplay,omitempty" xml:"DestinationDisplay,omitempty"`
	AimedArrivalTime          string `json:"AimedArrivalTime,omitempty" xml:"AimedArrivalTime,omitempty"`
	ExpectedArrivalTime       string `json:"ExpectedArrivalTime,omitempty" xml:"ExpectedArrivalTime,omitempty"`
	ArrivalStatus             string `json:"ArrivalStatus,omitempty" xml:"ArrivalStatus,omitempty"`
	ArrivalPlatformName       string `json:"ArrivalPlatformName,omitempty" xml:"ArrivalPlatformName,omitempty"`
	ArrivalBoardingActivity   string `json:"ArrivalBoardingActivity,omitempty" xml:"ArrivalBoardingActivity,omitempty"`
	AimedDepartureTime        string `json:"AimedDepartureTime,omitempty" xml:"AimedDepartureTime,omitempty"`
	ExpectedDepartureTime     string `json:"ExpectedDepartureTime,omitempty" xml:"ExpectedDepartureTime,omitempty"`
	DepartureStatus           string `json:"DepartureStatus,omitempty" xml:"DepartureStatus,omitempty"`
	DeparturePlatformName     string `json:"DeparturePlatformName,omitempty" xml:"DeparturePlatformName,omitempty"`
	DepartureBoardingActivity string `json:"DepartureBoardingActivity,omitempty" xml:"DepartureBoardingActivity,omitempty"`
}

// --- Deserialize ---

func DeserializeET(data []byte, format Format) (*ETFeed, error) {
	var feed ETFeed
	if err := deserialize(data, format, &feed); err != nil {
		return nil, fmt.Errorf("deserialize ET: %w", err)
	}
	return &feed, nil
}

func LoadET(path string, format Format) (*ETFeed, error) {
	var feed ETFeed
	if err := loadFromFile(path, format, &feed); err != nil {
		return nil, fmt.Errorf("load ET: %w", err)
	}
	return &feed, nil
}

// --- Serialize ---

func (f *ETFeed) Serialize(format Format) ([]byte, error) { return serialize(f, format) }
func (f *ETFeed) Dump(path string, format Format) error    { return dumpToFile(path, format, f) }

// AllEstimatedVehicleJourneys returns all journeys across all frames.
func (f *ETFeed) AllEstimatedVehicleJourneys() []EstimatedVehicleJourney {
	var all []EstimatedVehicleJourney
	for _, frame := range f.ServiceDelivery.EstimatedTimetableDelivery.EstimatedJourneyVersionFrame {
		all = append(all, frame.EstimatedVehicleJourney...)
	}
	return all
}
