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

func TestSnmpConfigs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td, err := os.MkdirTemp("", "twsnmpfk_snmp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	datastore.Init(ctx, td, &sync.WaitGroup{})
	datastore.MapConf.MapName = "SnmpConfigTest"
	datastore.MapConf.SnmpMode = "v2c"
	datastore.MapConf.Community = "public"
	_ = datastore.SaveMapConf()

	// 1. When DiscoverConf.SnmpConfigs is empty, GetDiscoverSnmpConfigs returns MapConf setting
	datastore.DiscoverConf.SnmpConfigs = []datastore.SnmpConfEnt{}
	configs := datastore.GetDiscoverSnmpConfigs()
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
	if configs[0].SnmpMode != "v2c" || configs[0].Community != "public" {
		t.Errorf("unexpected default config: %+v", configs[0])
	}

	// 2. When DiscoverConf.SnmpConfigs has additional entries, MapConf is first, followed by additions
	datastore.DiscoverConf.SnmpConfigs = []datastore.SnmpConfEnt{
		{SnmpMode: "v2c", Community: "public"}, // duplicate of MapConf, should be ignored
		{SnmpMode: "v2c", Community: "private"},
		{SnmpMode: "v3auth", SnmpUser: "admin", SnmpPassword: "password123"},
	}
	configs = datastore.GetDiscoverSnmpConfigs()
	if len(configs) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(configs))
	}
	if configs[0].Community != "public" {
		t.Errorf("expected configs[0] to be public, got %s", configs[0].Community)
	}
	if configs[1].Community != "private" {
		t.Errorf("expected configs[1] to be private, got %s", configs[1].Community)
	}
	if configs[2].SnmpMode != "v3auth" || configs[2].SnmpUser != "admin" {
		t.Errorf("expected configs[2] to be v3auth admin, got %+v", configs[2])
	}

	// 3. Test addFoundNode with custom SnmpConf and AddNetwork
	datastore.DiscoverConf.AddNetwork = true
	customDent := &discoverInfoEnt{
		IP:          "192.168.1.50",
		SysName:     "switch-50",
		SysObjectID: "1.3.6.1.4.1.9.1.1",
		IfMap:       make(map[string]string),
		ServerList:  map[string]bool{"lldp": true},
		SnmpConf: &datastore.SnmpConfEnt{
			SnmpMode:     "v3auth",
			SnmpUser:     "admin",
			SnmpPassword: "password123",
		},
	}
	addFoundNode(customDent)
	node := datastore.FindNodeFromIP("192.168.1.50")
	if node == nil {
		t.Fatal("node 192.168.1.50 not found after addFoundNode")
	}
	if node.SnmpMode != "v3auth" || node.User != "admin" || node.Password != "password123" {
		t.Errorf("node SNMP settings mismatch: Mode=%s, User=%s, Pass=%s", node.SnmpMode, node.User, node.Password)
	}
	net := datastore.FindNetworkByIP("192.168.1.50")
	if net == nil {
		t.Fatal("network node 192.168.1.50 not found after addFoundNode with AddNetwork=true")
	}
	if net.SnmpMode != "v3auth" || net.User != "admin" || net.Password != "password123" {
		t.Errorf("network SNMP settings mismatch: Mode=%s, User=%s, Pass=%s", net.SnmpMode, net.User, net.Password)
	}

	// 4. Test updateNode with bridge MIB and custom SnmpConf
	customDentBridge := &discoverInfoEnt{
		IP:          "192.168.1.51",
		SysName:     "switch-51",
		SysObjectID: "1.3.6.1.4.1.9.1.2",
		IfMap:       make(map[string]string),
		ServerList:  map[string]bool{"bridge": true},
		SnmpConf: &datastore.SnmpConfEnt{
			SnmpMode:     "v3authpriv",
			SnmpUser:     "privuser",
			SnmpPassword: "privpassword",
		},
	}
	// Add node initially with wrong/empty credentials
	initialNode := &datastore.NodeEnt{
		Name:      "switch-51",
		IP:        "192.168.1.51",
		Community: "public",
		SnmpMode:  "v2c",
	}
	_ = datastore.AddNode(initialNode)
	updateNode(initialNode, customDentBridge)

	updatedNode := datastore.FindNodeFromIP("192.168.1.51")
	if updatedNode.SnmpMode != "v3authpriv" || updatedNode.User != "privuser" {
		t.Errorf("updatedNode SNMP settings mismatch: Mode=%s, User=%s", updatedNode.SnmpMode, updatedNode.User)
	}
	bridgeNet := datastore.FindNetworkByIP("192.168.1.51")
	if bridgeNet == nil {
		t.Fatal("network node 192.168.1.51 not found after updateNode with bridge MIB")
	}
	if bridgeNet.SnmpMode != "v3authpriv" || bridgeNet.User != "privuser" {
		t.Errorf("bridgeNet SNMP settings mismatch: Mode=%s, User=%s", bridgeNet.SnmpMode, bridgeNet.User)
	}

	// 5. Test that a device without SNMP support (SysObjectID == "") is NOT added as network node
	datastore.DiscoverConf.AutoDetect = true
	unmanagedDent := &discoverInfoEnt{
		IP:          "192.168.1.52",
		HostName:    "unmanaged-switch",
		SysObjectID: "", // No SNMP response!
		IfMap:       make(map[string]string),
		ServerList:  make(map[string]bool),
		Vendor:      "Allied Telesis",
	}
	addFoundNode(unmanagedDent)
	unmanagedNode := datastore.FindNodeFromIP("192.168.1.52")
	if unmanagedNode == nil {
		t.Fatal("node 192.168.1.52 should be added as normal node")
	}
	unmanagedNet := datastore.FindNetworkByIP("192.168.1.52")
	if unmanagedNet != nil {
		t.Errorf("network node should NOT be added for device without SNMP support, got %+v", unmanagedNet)
	}
}

