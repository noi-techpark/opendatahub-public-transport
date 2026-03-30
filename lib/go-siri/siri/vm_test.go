// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package siri

import (
	"os"
	"testing"
)

func TestVM_ArrayFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/vm_array.json")
	if err != nil {
		t.Fatal(err)
	}
	vm, err := DeserializeVM(data, FormatJSON)
	if err != nil {
		t.Fatalf("deserialize: %v", err)
	}
	acts := vm.ServiceDelivery.VehicleMonitoringDelivery.VehicleActivity
	if len(acts) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(acts))
	}
	if acts[0].MonitoredVehicleJourney.VehicleRef != "1001" {
		t.Errorf("first VehicleRef = %q, want 1001", acts[0].MonitoredVehicleJourney.VehicleRef)
	}
	if acts[1].MonitoredVehicleJourney.VehicleRef != "1002" {
		t.Errorf("second VehicleRef = %q, want 1002", acts[1].MonitoredVehicleJourney.VehicleRef)
	}
	if acts[1].MonitoredVehicleJourney.MonitoredCall.VehicleAtStop != "true" {
		t.Error("second vehicle should be at stop")
	}
}

func TestVM_SingleObjectFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/vm_single.json")
	if err != nil {
		t.Fatal(err)
	}
	vm, err := DeserializeVM(data, FormatJSON)
	if err != nil {
		t.Fatalf("deserialize single: %v", err)
	}
	acts := vm.ServiceDelivery.VehicleMonitoringDelivery.VehicleActivity
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}
	if acts[0].MonitoredVehicleJourney.VehicleRef != "1001" {
		t.Errorf("VehicleRef = %q, want 1001", acts[0].MonitoredVehicleJourney.VehicleRef)
	}
	if acts[0].MonitoredVehicleJourney.Delay != "PT1M30S" {
		t.Errorf("Delay = %q, want PT1M30S", acts[0].MonitoredVehicleJourney.Delay)
	}
}

func TestVM_RoundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/vm_array.json")
	if err != nil {
		t.Fatal(err)
	}
	vm, err := DeserializeVM(data, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	out, err := vm.Serialize(FormatJSON)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	vm2, err := DeserializeVM(out, FormatJSON)
	if err != nil {
		t.Fatalf("re-deserialize: %v", err)
	}
	if len(vm2.ServiceDelivery.VehicleMonitoringDelivery.VehicleActivity) != 2 {
		t.Error("round-trip lost activities")
	}
}
