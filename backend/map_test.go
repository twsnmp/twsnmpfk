package backend

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestUpdateLineState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td, err := os.MkdirTemp("", "twsnmpfk_map_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)

	wg := &sync.WaitGroup{}
	datastore.Init(ctx, td, wg)

	// 1. ノードの作成
	node1 := &datastore.NodeEnt{
		Name:  "Node1",
		State: "normal",
	}
	if err := datastore.AddNode(node1); err != nil {
		t.Fatal(err)
	}

	node2 := &datastore.NodeEnt{
		Name:  "Node2",
		State: "repair",
	}
	if err := datastore.AddNode(node2); err != nil {
		t.Fatal(err)
	}

	// 2. ポーリング未指定のラインのテスト
	line1 := &datastore.LineEnt{
		NodeID1: node1.ID,
		NodeID2: node2.ID,
	}
	if err := datastore.AddLine(line1); err != nil {
		t.Fatal(err)
	}

	UpdateLineState()

	l1 := datastore.GetLine(line1.ID)
	if l1 == nil {
		t.Fatal("line1 not found")
	}
	if l1.State1 != "normal" {
		t.Errorf("expected line1.State1 to be 'normal', got '%s'", l1.State1)
	}
	if l1.State2 != "repair" {
		t.Errorf("expected line1.State2 to be 'repair', got '%s'", l1.State2)
	}

	// 3. ノード再確認時 (node2.State = "unknown")
	node2.State = "unknown"
	UpdateLineState()

	l1 = datastore.GetLine(line1.ID)
	if l1.State2 != "unknown" {
		t.Errorf("expected line1.State2 to reflect node unknown, got '%s'", l1.State2)
	}

	// 4. ポーリング指定ありのラインのテスト
	poll1 := &datastore.PollingEnt{
		NodeID: node1.ID,
		Name:   "PING",
		State:  "normal",
	}
	if err := datastore.AddPolling(poll1); err != nil {
		t.Fatal(err)
	}

	line2 := &datastore.LineEnt{
		NodeID1:    node1.ID,
		PollingID1: poll1.ID,
		NodeID2:    node2.ID,
	}
	if err := datastore.AddLine(line2); err != nil {
		t.Fatal(err)
	}

	UpdateLineState()
	l2 := datastore.GetLine(line2.ID)
	if l2 == nil {
		t.Fatal("line2 not found")
	}
	if l2.State1 != "normal" {
		t.Errorf("expected line2.State1 to be 'normal', got '%s'", l2.State1)
	}

	// node1 を unknown にした場合、ポーリング指定があっても unknown が反映されること
	node1.State = "unknown"
	UpdateLineState()
	l2 = datastore.GetLine(line2.ID)
	if l2.State1 != "unknown" {
		t.Errorf("expected line2.State1 to be 'unknown' when node is unknown, got '%s'", l2.State1)
	}
}
