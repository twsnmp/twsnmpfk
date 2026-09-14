package backend

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func setupTestDB(t *testing.T) (context.CancelFunc, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	td, err := os.MkdirTemp("", "twsnmpfk_layout_test")
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	datastore.Init(ctx, td, wg)

	cleanup := func() {
		cancel()
		os.RemoveAll(td)
	}
	return cancel, cleanup
}

func TestLayoutHierarchical(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	// Add Gateway
	gw := &datastore.NodeEnt{
		Name: "Gateway",
		IP:   "192.168.1.1",
		Icon: "router",
		X:    10,
		Y:    10,
	}
	_ = datastore.AddNode(gw)

	// Add Switch
	sw := &datastore.NetworkEnt{
		Name: "CoreSwitch",
		IP:   "192.168.1.2",
		X:    20,
		Y:    20,
		Ports: []datastore.PortEnt{
			{ID: "p1", Name: "port1", X: 0, Y: 0},
			{ID: "p2", Name: "port2", X: 1, Y: 0},
		},
	}
	_ = datastore.AddNetwork(sw)

	// Add Connected Node 1
	n1 := &datastore.NodeEnt{
		Name: "Server1",
		IP:   "192.168.1.10",
		Icon: "server",
		X:    30,
		Y:    30,
	}
	_ = datastore.AddNode(n1)

	// Add Connected Node 2
	n2 := &datastore.NodeEnt{
		Name: "PC1",
		IP:   "192.168.1.50",
		Icon: "desktop",
		X:    40,
		Y:    40,
	}
	_ = datastore.AddNode(n2)

	// Connect lines
	_ = datastore.AddLine(&datastore.LineEnt{
		NodeID1:    "NET:" + sw.ID,
		PollingID1: "p1",
		NodeID2:    n1.ID,
	})
	_ = datastore.AddLine(&datastore.LineEnt{
		NodeID1:    "NET:" + sw.ID,
		PollingID1: "p2",
		NodeID2:    n2.ID,
	})

	// Run Hierarchical Layout
	count, err := OptimizeLayout(datastore.AutoLayoutHierarchical)
	if err != nil {
		t.Fatalf("OptimizeLayout error: %v", err)
	}
	if count < 4 {
		t.Fatalf("expected at least 4 items laid out, got %d", count)
	}

	// Verify Gateway is at Layer 0 (Y=100)
	savedGW := datastore.GetNode(gw.ID)
	if savedGW.Y != 100 {
		t.Errorf("expected gateway Y=100, got %d", savedGW.Y)
	}

	// Verify Switch is at Layer 1 (Y=280)
	savedSW := datastore.GetNetwork(sw.ID)
	if savedSW.Y != 280 {
		t.Errorf("expected switch Y=280, got %d", savedSW.Y)
	}

	// Verify n1 and n2 are positioned below the switch (Y > savedSW.Y)
	savedN1 := datastore.GetNode(n1.ID)
	savedN2 := datastore.GetNode(n2.ID)
	if savedN1.Y <= savedSW.Y || savedN2.Y <= savedSW.Y {
		t.Errorf("expected connected nodes below switch Y=%d, got n1.Y=%d, n2.Y=%d", savedSW.Y, savedN1.Y, savedN2.Y)
	}

	// Verify Undo
	if !HasUndoLayout() {
		t.Fatal("expected HasUndoLayout to be true")
	}

	restored, err := UndoLayout()
	if err != nil {
		t.Fatalf("UndoLayout error: %v", err)
	}
	if restored == 0 {
		t.Errorf("expected items restored, got 0")
	}

	// Verify positions restored to initial values
	restoredGW := datastore.GetNode(gw.ID)
	if restoredGW.X != 10 || restoredGW.Y != 10 {
		t.Errorf("expected restored GW at (10, 10), got (%d, %d)", restoredGW.X, restoredGW.Y)
	}
}

func TestLayoutClusterAndCategorized(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	sw := &datastore.NetworkEnt{
		Name: "Switch1",
		IP:   "192.168.1.2",
		X:    0,
		Y:    0,
	}
	_ = datastore.AddNetwork(sw)

	n1 := &datastore.NodeEnt{
		Name: "PC1",
		IP:   "192.168.1.10",
		Icon: "windows",
		X:    0,
		Y:    0,
	}
	_ = datastore.AddNode(n1)

	// Test Cluster Layout
	_, err := OptimizeLayout(datastore.AutoLayoutCluster)
	if err != nil {
		t.Fatalf("layoutCluster error: %v", err)
	}

	// Test Categorized Layout
	_, err = OptimizeLayout(datastore.AutoLayoutCategorized)
	if err != nil {
		t.Fatalf("layoutCategorized error: %v", err)
	}
}
