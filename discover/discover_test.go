package discover

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/ping"
)

func TestDiscover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ping.Start(ctx, &sync.WaitGroup{})
	defer cancel()
	time.Sleep(time.Second * 1)
	td, err := os.MkdirTemp("", "twsnmpfk_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	datastore.Init(ctx, td, &sync.WaitGroup{})
	datastore.MapConf.MapName = "Test123"
	if err := datastore.SaveMapConf(); err != nil {
		t.Fatal(err)
	}
	datastore.MapConf.Community = "public"
	datastore.DiscoverConf.StartIP = "192.168.1.1"
	datastore.DiscoverConf.EndIP = "192.168.1.2"
	datastore.DiscoverConf.Retry = 1
	datastore.DiscoverConf.Timeout = 2
	err = StartDiscover()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 15)
	t.Log("Done")
}

func TestUpdateNodeOnRecheck(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td, err := os.MkdirTemp("", "twsnmpfk_recheck_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	datastore.Init(ctx, td, &sync.WaitGroup{})
	datastore.MapConf.MapName = "ReCheckTest"
	_ = datastore.SaveMapConf()

	datastore.DiscoverConf.AutoDetect = true
	datastore.DiscoverConf.ReCheck = true

	// Node was previously detected as generic linux with mdi-linux icon
	node := &datastore.NodeEnt{
		ID:     "test-esp32-node",
		Name:   "192.168.1.6",
		IP:     "192.168.1.6",
		MAC:    "64:B7:08:12:34:56",
		Vendor: "Espressif Inc.",
		Icon:   "mdi-linux",
		Descr:  "Found at 2026/09/13 [Linux Server]",
	}
	if err := datastore.AddNode(node); err != nil {
		t.Fatal(err)
	}

	dent := &discoverInfoEnt{
		IP:         "192.168.1.6",
		ServerList: make(map[string]bool),
		IfMap:      make(map[string]string),
	}
	// Note: dent.Vendor and dent.MAC start empty (like ARP table missing entry during discover ping)
	updateNode(node, dent)

	updatedNode := datastore.GetNode(node.ID)
	if updatedNode == nil {
		t.Fatal("node not found after update")
	}
	if updatedNode.Icon != "mdi-developer-board" {
		t.Errorf("expected icon mdi-developer-board, got %s", updatedNode.Icon)
	}
	if updatedNode.Vendor != "Espressif Inc." {
		t.Errorf("expected vendor Espressif Inc., got %s", updatedNode.Vendor)
	}
	if !strings.Contains(updatedNode.Descr, "[ESP32 / ESP8266 (IoT Board)]") {
		t.Errorf("expected descr to contain [ESP32 / ESP8266 (IoT Board)], got %s", updatedNode.Descr)
	}
	if strings.Contains(updatedNode.Descr, "[Linux Server]") {
		t.Errorf("expected old tag [Linux Server] to be removed, got %s", updatedNode.Descr)
	}
}

