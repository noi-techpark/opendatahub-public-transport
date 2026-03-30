// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/noi-techpark/opendatahub-public-transport/lib/go-gtfsrt/gtfsrt"
	"github.com/noi-techpark/opendatahub-public-transport/lib/go-siri/siri"
)

// testGTFSURL and testNeTExURL defined in staticdata_test.go

func TestSETransformPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skip in short mode")
	}

	sd, err := LoadStaticData(testNeTExURL, testGTFSURL)
	if err != nil {
		t.Fatalf("load static data: %v", err)
	}

	resp, err := http.Get("https://efa.sta.bz.it/siri-lite/situation-exchange")
	if err != nil {
		t.Fatalf("fetch SE feed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	se, err := siri.DeserializeSE(body, siri.FormatJSON)
	if err != nil {
		t.Fatalf("deserialize SE: %v", err)
	}

	situations := se.ServiceDelivery.SituationExchangeDelivery.Situations.PtSituationElement
	t.Logf("SIRI SE: %d situations", len(situations))

	resolver := sd.GetResolver()
	rt := ConvertSE(se, resolver)

	t.Logf("GTFS-RT: %d alert entities", len(rt.Entity))

	// Validate each alert
	for _, e := range rt.Entity {
		if e.Alert == nil {
			t.Errorf("entity %s has nil Alert", e.ID)
			continue
		}
		if e.ID == "" {
			t.Error("entity has empty ID")
		}
		if len(e.Alert.InformedEntity) == 0 {
			t.Logf("WARN: alert %s has no informed entities (no matching stops/routes in GTFS)", e.ID)
		}
		if e.Alert.Cause == "" {
			t.Errorf("alert %s has empty cause", e.ID)
		}
	}

	// Serialization
	pbData, err := rt.Serialize(gtfsrt.FormatProtobuf)
	if err != nil {
		t.Fatalf("serialize protobuf: %v", err)
	}
	jsonData, err := rt.Serialize(gtfsrt.FormatJSON)
	if err != nil {
		t.Fatalf("serialize JSON: %v", err)
	}
	t.Logf("Output: %d bytes protobuf, %d bytes JSON", len(pbData), len(jsonData))
}
