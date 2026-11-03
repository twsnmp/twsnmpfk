package main

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestCheckDrawItemTitles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td, err := os.MkdirTemp("", "twsnmpfk_drawitem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	datastore.Init(ctx, td, wg)

	node := &datastore.NodeEnt{
		Name:  "TestNode",
		State: "normal",
	}
	if err := datastore.AddNode(node); err != nil {
		t.Fatal(err)
	}

	p := &datastore.PollingEnt{
		NodeID: node.ID,
		Name:   "CPU Utilization",
		Type:   "snmp",
	}
	if err := datastore.AddPolling(p); err != nil {
		t.Fatal(err)
	}
	p.Result["load"] = 42.5

	// 1. Classic Gauge (Type 5) with empty text
	diGauge := &datastore.DrawItemEnt{
		Type:      datastore.DrawItemTypePollingGauge,
		PollingID: p.ID,
		VarName:   "load",
		Text:      "",
	}
	checkDrawItem(diGauge)
	if diGauge.Text != "CPU Utilization (load)" {
		t.Errorf("expected diGauge.Text to be 'CPU Utilization (load)', got %q", diGauge.Text)
	}

	// 2. New Gauge (Type 6) with custom text
	diNewGauge := &datastore.DrawItemEnt{
		Type:      datastore.DrawItemTypePollingNewGauge,
		PollingID: p.ID,
		VarName:   "load",
		Text:      "Custom Title",
	}
	checkDrawItem(diNewGauge)
	if diNewGauge.Text != "Custom Title" {
		t.Errorf("expected diNewGauge.Text to be 'Custom Title', got %q", diNewGauge.Text)
	}

	// 3. Line (Type 8) with empty text and no VarName
	diLine := &datastore.DrawItemEnt{
		Type:      datastore.DrawItemTypePollingLine,
		PollingID: p.ID,
		Text:      "",
	}
	checkDrawItem(diLine)
	if diLine.Text != "CPU Utilization" {
		t.Errorf("expected diLine.Text to be 'CPU Utilization', got %q", diLine.Text)
	}

	// 4. Polling Text (Type 4) should keep metric text
	diText := &datastore.DrawItemEnt{
		Type:      datastore.DrawItemTypePollingText,
		PollingID: p.ID,
		VarName:   "load",
		Text:      "",
	}
	checkDrawItem(diText)
	if diText.Text != "42.500000" {
		t.Errorf("expected diText.Text to be '42.500000', got %q", diText.Text)
	}
}
