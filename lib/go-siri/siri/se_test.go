// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"os"
	"testing"
)

func TestSE_ArrayFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/se_array.json")
	if err != nil {
		t.Fatal(err)
	}
	se, err := DeserializeSE(data, FormatJSON)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	elems := se.ServiceDelivery.SituationExchangeDelivery.Situations.PtSituationElement
	if len(elems) != 2 {
		t.Fatalf("expected 2 situations, got %d", len(elems))
	}

	// First situation: array polymorphism in ValidityPeriod, ReasonName, AffectedStopPoint, Consequence
	s1 := elems[0]
	if s1.SituationNumber != "SIT001" {
		t.Errorf("s1 SituationNumber = %q", s1.SituationNumber)
	}

	vp := ParseValidityPeriods(s1.ValidityPeriod)
	if len(vp) != 1 {
		t.Fatalf("expected 1 validity period (from array), got %d", len(vp))
	}
	if vp[0].StartTime != "2026-03-30T06:00:00Z" {
		t.Errorf("validity start = %q", vp[0].StartTime)
	}

	rn := ParseReasonNames(s1.ReasonName)
	if len(rn) != 2 {
		t.Fatalf("expected 2 reason names (from array), got %d", len(rn))
	}
	if rn[0].Text != "Baustelle" {
		t.Errorf("first reason = %q", rn[0].Text)
	}

	if s1.Affects.StopPoints == nil {
		t.Fatal("StopPoints is nil")
	}
	stops := s1.Affects.StopPoints.AffectedStopPoint
	if len(stops) != 2 {
		t.Fatalf("expected 2 affected stops (from array), got %d", len(stops))
	}

	cons := ParseConsequences(s1.Consequences.Consequence)
	if len(cons) != 1 {
		t.Fatalf("expected 1 consequence (from array), got %d", len(cons))
	}
	if cons[0].Condition != "delayed" {
		t.Errorf("condition = %q", cons[0].Condition)
	}

	// Second situation: single-object polymorphism in ValidityPeriod, ReasonName, Network, Consequence
	s2 := elems[1]
	if s2.SituationNumber != "SIT002" {
		t.Errorf("s2 SituationNumber = %q", s2.SituationNumber)
	}

	vp2 := ParseValidityPeriods(s2.ValidityPeriod)
	if len(vp2) != 1 {
		t.Fatalf("expected 1 validity period (from single object), got %d", len(vp2))
	}

	rn2 := ParseReasonNames(s2.ReasonName)
	if len(rn2) != 1 {
		t.Fatalf("expected 1 reason name (from single object), got %d", len(rn2))
	}
	if rn2[0].Text != "Streik" {
		t.Errorf("reason = %q", rn2[0].Text)
	}

	networks := ParseAffectedNetworks(s2.Affects.Networks.AffectedNetwork)
	if len(networks) != 1 {
		t.Fatalf("expected 1 network (from single object), got %d", len(networks))
	}
	lines := ParseAffectedLines(networks[0].AffectedLine)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (from single object), got %d", len(lines))
	}
	if lines[0].LineRef != "240" {
		t.Errorf("LineRef = %q", lines[0].LineRef)
	}

	cons2 := ParseConsequences(s2.Consequences.Consequence)
	if len(cons2) != 1 {
		t.Fatalf("expected 1 consequence (from single object), got %d", len(cons2))
	}
	if cons2[0].Severity != "severe" {
		t.Errorf("severity = %q", cons2[0].Severity)
	}
}

func TestSE_SingleObjectFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/se_single.json")
	if err != nil {
		t.Fatal(err)
	}
	se, err := DeserializeSE(data, FormatJSON)
	if err != nil {
		t.Fatalf("deserialize single: %v", err)
	}

	// Single PtSituationElement (not array)
	elems := se.ServiceDelivery.SituationExchangeDelivery.Situations.PtSituationElement
	if len(elems) != 1 {
		t.Fatalf("expected 1 situation from single object, got %d", len(elems))
	}
	if elems[0].SituationNumber != "SIT001" {
		t.Errorf("SituationNumber = %q", elems[0].SituationNumber)
	}

	// Single AffectedStopPoint (not array)
	stops := elems[0].Affects.StopPoints.AffectedStopPoint
	if len(stops) != 1 {
		t.Fatalf("expected 1 stop from single object, got %d", len(stops))
	}
	if stops[0].StopPointRef != "stop_1" {
		t.Errorf("StopPointRef = %q", stops[0].StopPointRef)
	}
}

func TestSE_RoundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/se_array.json")
	if err != nil {
		t.Fatal(err)
	}
	se, err := DeserializeSE(data, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	out, err := se.Serialize(FormatJSON)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	se2, err := DeserializeSE(out, FormatJSON)
	if err != nil {
		t.Fatalf("re-deserialize: %v", err)
	}
	if len(se2.ServiceDelivery.SituationExchangeDelivery.Situations.PtSituationElement) != 2 {
		t.Error("round-trip lost situations")
	}
}
