package siri

import "fmt"

// --- Vehicle Monitoring structs ---

type VMFeed struct {
	ServiceDelivery struct {
		ResponseTimestamp         string                    `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
		ProducerRef               string                    `json:"ProducerRef" xml:"ProducerRef"`
		VehicleMonitoringDelivery VehicleMonitoringDelivery `json:"VehicleMonitoringDelivery" xml:"VehicleMonitoringDelivery"`
	} `json:"ServiceDelivery" xml:"ServiceDelivery"`
}

type VehicleMonitoringDelivery struct {
	ResponseTimestamp string            `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
	VehicleActivity   []VehicleActivity `json:"VehicleActivity" xml:"VehicleActivity"`
}

type VehicleActivity struct {
	RecordedAtTime          string                  `json:"RecordedAtTime" xml:"RecordedAtTime"`
	ValidUntilTime          string                  `json:"ValidUntilTime" xml:"ValidUntilTime"`
	VehicleMonitoringRef    string                  `json:"VehicleMonitoringRef" xml:"VehicleMonitoringRef"`
	MonitoredVehicleJourney MonitoredVehicleJourney `json:"MonitoredVehicleJourney" xml:"MonitoredVehicleJourney"`
}

type MonitoredVehicleJourney struct {
	LineRef                 string                  `json:"LineRef" xml:"LineRef"`
	DirectionRef            string                  `json:"DirectionRef" xml:"DirectionRef"`
	FramedVehicleJourneyRef FramedVehicleJourneyRef `json:"FramedVehicleJourneyRef" xml:"FramedVehicleJourneyRef"`
	PublishedLineName       string                  `json:"PublishedLineName" xml:"PublishedLineName"`
	DirectionName           string                  `json:"DirectionName" xml:"DirectionName"`
	OperatorRef             string                  `json:"OperatorRef" xml:"OperatorRef"`
	ProductCategoryRef      string                  `json:"ProductCategoryRef" xml:"ProductCategoryRef"`
	Monitored               string                  `json:"Monitored" xml:"Monitored"`
	InCongestion            string                  `json:"InCongestion" xml:"InCongestion"`
	VehicleLocation         VehicleLocation         `json:"VehicleLocation" xml:"VehicleLocation"`
	Delay                   string                  `json:"Delay" xml:"Delay"`
	VehicleRef              string                  `json:"VehicleRef" xml:"VehicleRef"`
	MonitoredCall           MonitoredCall           `json:"MonitoredCall" xml:"MonitoredCall"`
}

type FramedVehicleJourneyRef struct {
	DataFrameRef           string `json:"DataFrameRef" xml:"DataFrameRef"`
	DatedVehicleJourneyRef string `json:"DatedVehicleJourneyRef" xml:"DatedVehicleJourneyRef"`
}

type VehicleLocation struct {
	Longitude string `json:"Longitude" xml:"Longitude"`
	Latitude  string `json:"Latitude" xml:"Latitude"`
}

type MonitoredCall struct {
	StopPointRef  string `json:"StopPointRef" xml:"StopPointRef"`
	StopPointName string `json:"StopPointName" xml:"StopPointName"`
	VehicleAtStop string `json:"VehicleAtStop" xml:"VehicleAtStop"`
}

// --- Deserialize ---

// DeserializeVM decodes bytes into a VMFeed.
func DeserializeVM(data []byte, format Format) (*VMFeed, error) {
	var feed VMFeed
	if err := deserialize(data, format, &feed); err != nil {
		return nil, fmt.Errorf("deserialize VM: %w", err)
	}
	return &feed, nil
}

// LoadVM reads a file and deserializes it into a VMFeed.
func LoadVM(path string, format Format) (*VMFeed, error) {
	var feed VMFeed
	if err := loadFromFile(path, format, &feed); err != nil {
		return nil, fmt.Errorf("load VM: %w", err)
	}
	return &feed, nil
}

// --- Serialize ---

// Serialize encodes the VM feed to bytes.
func (f *VMFeed) Serialize(format Format) ([]byte, error) {
	return serialize(f, format)
}

// Dump serializes and writes to a file.
func (f *VMFeed) Dump(path string, format Format) error {
	return dumpToFile(path, format, f)
}
