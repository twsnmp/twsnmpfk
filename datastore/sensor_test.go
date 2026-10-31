package datastore

import (
	"context"
	"os"
	"sync"
	"testing"
)

func TestSensorDataStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	td, err := os.MkdirTemp("", "twsnmpfk_sensor_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	Init(ctx, td, wg)
	defer CloseDB()

	// 1. UpdateSensor
	UpdateSensor("192.168.1.50", "twWifiScan", "wlan0", 5)
	UpdateSensor("192.168.1.50", "twWifiScan", "wlan0", 3)

	sensorsList := ListSensors()
	if len(sensorsList) != 1 {
		t.Fatalf("expected 1 sensor, got %d", len(sensorsList))
	}
	sn := sensorsList[0]
	if sn.Host != "192.168.1.50" || sn.Type != "twWifiScan" || sn.Total != 8 || sn.Send != 2 {
		t.Fatalf("unexpected sensor values: %+v", sn)
	}

	// 2. CheckSensorStats
	CheckSensorStats("192.168.1.50", "twWifiScan", "wlan0", 10, 1, 18, 2.5)
	sn, err = GetSensor(sn.ID)
	if err != nil {
		t.Fatalf("GetSensor failed: %v", err)
	}
	if sn.StatsLen != 1 || len(sn.Stats) != 1 {
		t.Fatalf("expected 1 stat entry, got %d", sn.StatsLen)
	}
	if sn.Stats[0].PS != 2.5 {
		t.Fatalf("expected PS 2.5, got %f", sn.Stats[0].PS)
	}

	// 3. CheckSensorMonitor
	CheckSensorMonitor("192.168.1.50", "twWifiScan", "wlan0", 15.5, 45.0, 0.8, 1200.0, 5600.0, 100, 200, 42)
	sn, err = GetSensor(sn.ID)
	if err != nil {
		t.Fatalf("GetSensor failed: %v", err)
	}
	if sn.MonitorsLen != 1 || len(sn.Monitors) != 1 {
		t.Fatalf("expected 1 monitor entry, got %d", sn.MonitorsLen)
	}
	if sn.Monitors[0].CPU != 15.5 {
		t.Fatalf("expected CPU 15.5, got %f", sn.Monitors[0].CPU)
	}

	// 4. ToggleSensorIgnore
	if err := ToggleSensorIgnore(sn.ID); err != nil {
		t.Fatalf("ToggleSensorIgnore failed: %v", err)
	}
	sn, _ = GetSensor(sn.ID)
	if !sn.Ignore || sn.State != "off" {
		t.Fatalf("expected ignored sensor state 'off', got %s", sn.State)
	}

	// 5. DeleteSensors
	if err := DeleteSensors([]string{sn.ID}); err != nil {
		t.Fatalf("DeleteSensors failed: %v", err)
	}
	if len(ListSensors()) != 0 {
		t.Fatalf("expected 0 sensors after deletion")
	}
}

func TestLogReportDataStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	td, err := os.MkdirTemp("", "twsnmpfk_logreport_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	Init(ctx, td, wg)
	defer CloseDB()

	kind := "wifiAP"
	items := map[string][]byte{
		"ap-1": []byte(`{"SSID":"Office-5G","BSSID":"00:11:22:33:44:55"}`),
		"ap-2": []byte(`{"SSID":"Guest-2G","BSSID":"66:77:88:99:AA:BB"}`),
	}

	// Save
	if err := SaveLogReportData(kind, items); err != nil {
		t.Fatalf("SaveLogReportData failed: %v", err)
	}

	// List
	list, err := ListLogReportData(kind)
	if err != nil {
		t.Fatalf("ListLogReportData failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}

	// Get single
	data, err := GetLogReportData(kind, "ap-1")
	if err != nil || data == nil {
		t.Fatalf("GetLogReportData failed: %v", err)
	}

	// Delete single
	if err := DeleteLogReportData(kind, []string{"ap-1"}); err != nil {
		t.Fatalf("DeleteLogReportData failed: %v", err)
	}
	list, _ = ListLogReportData(kind)
	if len(list) != 1 {
		t.Fatalf("expected 1 item after delete, got %d", len(list))
	}

	// Reset
	if err := ResetLogReportData(kind); err != nil {
		t.Fatalf("ResetLogReportData failed: %v", err)
	}
	list, _ = ListLogReportData(kind)
	if len(list) != 0 {
		t.Fatalf("expected 0 items after reset, got %d", len(list))
	}
}
