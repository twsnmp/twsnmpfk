package main

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
)

func TestGetSystemInfo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td, err := os.MkdirTemp("", "twsnmpfk_system_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	if err := datastore.Init(ctx, td, wg); err != nil {
		t.Fatal(err)
	}
	dataStorePath = td
	defer func() {
		dataStorePath = ""
	}()

	// Configure some test data
	node := &datastore.NodeEnt{
		Name:  "SysTestNode",
		State: "normal",
	}
	if err := datastore.AddNode(node); err != nil {
		t.Fatal(err)
	}

	p := &datastore.PollingEnt{
		NodeID: node.ID,
		Name:   "Ping",
		Type:   "ping",
	}
	if err := datastore.AddPolling(p); err != nil {
		t.Fatal(err)
	}

	datastore.MapConf.EnableSyslogd = true
	datastore.MapConf.EnableTrapd = false
	datastore.MapConf.EnableNetflowd = true
	datastore.MapConf.EnableSFlowd = false
	datastore.MapConf.EnableOTel = true
	datastore.MapConf.EnableMqtt = true
	datastore.MapConf.MCPTransport = "sse"
	datastore.MapConf.MCPEndpoint = "127.0.0.1:8089"
	datastore.MapConf.EnableArpWatch = true
	datastore.MapConf.ArpWatchRange = "192.168.1.0/24"
	datastore.SyslogPort = 514
	datastore.TrapPort = 162
	datastore.NetFlowPort = 2055

	app := NewApp()
	info := app.GetSystemInfo()
	if info == nil {
		t.Fatal("GetSystemInfo returned nil")
	}

	if info.NodeCount != 1 {
		t.Errorf("expected NodeCount 1, got %d", info.NodeCount)
	}
	if info.PollCount != 1 {
		t.Errorf("expected PollCount 1, got %d", info.PollCount)
	}
	if info.Status != "ok" {
		t.Errorf("expected Status 'ok', got '%s'", info.Status)
	}
	if info.PingMode == "" {
		t.Error("expected non-empty PingMode")
	}
	if info.Uptime == "" {
		t.Error("expected non-empty Uptime")
	}

	// Receivers check - initial state before logger registers them (stopped or disabled)
	if syslog, ok := info.Receivers["syslog"]; !ok || syslog.Status != "stopped" {
		t.Errorf("expected syslog receiver stopped initially, got %+v", syslog)
	}
	if trap, ok := info.Receivers["trap"]; !ok || trap.Status != "disabled" {
		t.Errorf("expected trap receiver disabled, got %+v", trap)
	}

	// Simulate running receiver
	logger.SetReceiverStatus("syslog", logger.ReceiverStatus{
		Status: "running",
		Port:   "UDP/TCP :514",
	})
	info = app.GetSystemInfo()
	if syslog, ok := info.Receivers["syslog"]; !ok || syslog.Status != "running" {
		t.Errorf("expected syslog receiver running, got %+v", syslog)
	}
	if info.Status != "ok" {
		t.Errorf("expected system status ok, got %s", info.Status)
	}

	// Simulate port bind error (address already in use)
	logger.SetReceiverStatus("syslog", logger.ReceiverStatus{
		Status: "error",
		Port:   "UDP/TCP :514",
		Error:  "listen udp 0.0.0.0:514: bind: address already in use",
	})
	info = app.GetSystemInfo()
	if syslog, ok := info.Receivers["syslog"]; !ok || syslog.Status != "error" || syslog.Error == "" {
		t.Errorf("expected syslog receiver error with message, got %+v", syslog)
	}
	if info.Status != "error" {
		t.Errorf("expected system status error when receiver fails, got %s", info.Status)
	}

	// Clean up
	logger.ClearReceiverStatus("syslog")
}
