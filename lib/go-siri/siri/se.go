package siri

import (
	"encoding/json"
	"fmt"
)

// --- Situation Exchange structs ---

type SEFeed struct {
	ServiceDelivery struct {
		ResponseTimestamp         string                    `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
		ProducerRef               string                    `json:"ProducerRef" xml:"ProducerRef"`
		SituationExchangeDelivery SituationExchangeDelivery `json:"SituationExchangeDelivery" xml:"SituationExchangeDelivery"`
	} `json:"ServiceDelivery" xml:"ServiceDelivery"`
}

type SituationExchangeDelivery struct {
	ResponseTimestamp string     `json:"ResponseTimestamp" xml:"ResponseTimestamp"`
	Situations        Situations `json:"Situations" xml:"Situations"`
}

type Situations struct {
	PtSituationElement []PtSituationElement `json:"PtSituationElement" xml:"PtSituationElement"`
}

type PtSituationElement struct {
	CreationTime    string          `json:"CreationTime" xml:"CreationTime"`
	ParticipantRef  string          `json:"ParticipantRef" xml:"ParticipantRef"`
	SituationNumber string          `json:"SituationNumber" xml:"SituationNumber"`
	Version         string          `json:"Version" xml:"Version"`
	Progress        string          `json:"Progress" xml:"Progress"`
	ValidityPeriod  json.RawMessage `json:"ValidityPeriod" xml:"-"`
	AlertCause      string          `json:"AlertCause" xml:"AlertCause"`
	ReasonName      json.RawMessage `json:"ReasonName" xml:"-"`
	ScopeType       string          `json:"ScopeType" xml:"ScopeType"`
	Planned         string          `json:"Planned" xml:"Planned"`
	Language        string          `json:"Language" xml:"Language"`
	Affects         Affects         `json:"Affects" xml:"Affects"`
	Consequences    Consequences    `json:"Consequences" xml:"Consequences"`
}

type Affects struct {
	StopPoints *AffectedStopPoints `json:"StopPoints,omitempty" xml:"StopPoints,omitempty"`
	Networks   *AffectedNetworks   `json:"Networks,omitempty" xml:"Networks,omitempty"`
}

type AffectedStopPoints struct {
	AffectedStopPoint []AffectedStopPoint `json:"AffectedStopPoint" xml:"AffectedStopPoint"`
}

type AffectedStopPoint struct {
	StopPointRef string `json:"StopPointRef" xml:"StopPointRef"`
	PlaceName    string `json:"PlaceName" xml:"PlaceName"`
}

type AffectedNetworks struct {
	AffectedNetwork json.RawMessage `json:"AffectedNetwork" xml:"-"`
}

type AffectedNetwork struct {
	AffectedLine json.RawMessage `json:"AffectedLine" xml:"-"`
}

type AffectedLine struct {
	LineRef           string `json:"LineRef" xml:"LineRef"`
	PublishedLineName string `json:"PublishedLineName" xml:"PublishedLineName"`
}

type Consequences struct {
	Consequence json.RawMessage `json:"Consequence" xml:"-"`
}

type Consequence struct {
	Condition string `json:"Condition" xml:"Condition"`
	Severity  string `json:"Severity" xml:"Severity"`
}

type SIRITimeRange struct {
	StartTime string `json:"StartTime" xml:"StartTime"`
	EndTime   string `json:"EndTime" xml:"EndTime"`
}

type ReasonText struct {
	Lang string `json:"@xml:lang"`
	Text string `json:"#text"`
}

// --- Deserialize ---

func DeserializeSE(data []byte, format Format) (*SEFeed, error) {
	var feed SEFeed
	if err := deserialize(data, format, &feed); err != nil {
		return nil, fmt.Errorf("deserialize SE: %w", err)
	}
	return &feed, nil
}

func LoadSE(path string, format Format) (*SEFeed, error) {
	var feed SEFeed
	if err := loadFromFile(path, format, &feed); err != nil {
		return nil, fmt.Errorf("load SE: %w", err)
	}
	return &feed, nil
}

// --- Serialize ---

func (f *SEFeed) Serialize(format Format) ([]byte, error) { return serialize(f, format) }
func (f *SEFeed) Dump(path string, format Format) error    { return dumpToFile(path, format, f) }

// --- Polymorphic parsers ---

func ParseValidityPeriods(raw json.RawMessage) []SIRITimeRange {
	if len(raw) == 0 {
		return nil
	}
	var arr []SIRITimeRange
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		return arr
	}
	var single SIRITimeRange
	if err := json.Unmarshal(raw, &single); err == nil && single.StartTime != "" {
		return []SIRITimeRange{single}
	}
	return nil
}

func ParseReasonNames(raw json.RawMessage) []ReasonText {
	if len(raw) == 0 {
		return nil
	}
	var arr []ReasonText
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		return arr
	}
	var single ReasonText
	if err := json.Unmarshal(raw, &single); err == nil && single.Text != "" {
		return []ReasonText{single}
	}
	return nil
}

func ParseAffectedNetworks(raw json.RawMessage) []AffectedNetwork {
	if len(raw) == 0 {
		return nil
	}
	var arr []AffectedNetwork
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	var single AffectedNetwork
	if err := json.Unmarshal(raw, &single); err == nil {
		return []AffectedNetwork{single}
	}
	return nil
}

func ParseAffectedLines(raw json.RawMessage) []AffectedLine {
	if len(raw) == 0 {
		return nil
	}
	var arr []AffectedLine
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	var single AffectedLine
	if err := json.Unmarshal(raw, &single); err == nil {
		return []AffectedLine{single}
	}
	return nil
}

func ParseConsequences(raw json.RawMessage) []Consequence {
	if len(raw) == 0 {
		return nil
	}
	var arr []Consequence
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	var single Consequence
	if err := json.Unmarshal(raw, &single); err == nil {
		return []Consequence{single}
	}
	return nil
}
