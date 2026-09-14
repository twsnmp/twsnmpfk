package notify

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func setupTestDB(t *testing.T) (func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	td, err := os.MkdirTemp("", "twsnmpfk_dep_test")
	if err != nil {
		cancel()
		return nil, err
	}
	wg := &sync.WaitGroup{}
	err = datastore.Init(ctx, td, wg)
	if err != nil {
		cancel()
		os.RemoveAll(td)
		return nil, err
	}

	clearStore := func() {
		datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
			_ = datastore.DeleteNode(n.ID)
			return true
		})
		datastore.ForEachLines(func(l *datastore.LineEnt) bool {
			_ = datastore.DeleteLine(l.ID)
			return true
		})
	}
	clearStore()

	cleanup := func() {
		clearStore()
		cancel()
		datastore.CloseDB()
		os.RemoveAll(td)
	}
	return cleanup, nil
}

func TestBuildDependencyTree(t *testing.T) {
	cleanup, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("setupTestDB err=%v", err)
	}
	defer cleanup()

	// Setup nodes
	rootNode := &datastore.NodeEnt{Name: "RootRouter", Icon: "router"}
	swNode := &datastore.NodeEnt{Name: "Switch1", Icon: "switch"}
	srv1 := &datastore.NodeEnt{Name: "Server1", Icon: "server"}
	srv2 := &datastore.NodeEnt{Name: "Server2", Icon: "server"}

	_ = datastore.AddNode(rootNode)
	_ = datastore.AddNode(swNode)
	_ = datastore.AddNode(srv1)
	_ = datastore.AddNode(srv2)

	// Setup lines using assigned IDs: root -- sw1, sw1 -- srv1, sw1 -- srv2
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: rootNode.ID, NodeID2: swNode.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: swNode.ID, NodeID2: srv1.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: swNode.ID, NodeID2: srv2.ID})

	parentMap, depthMap := BuildDependencyTree(rootNode.ID)

	if depthMap[rootNode.ID] != 0 {
		t.Errorf("expected root depth 0, got %d", depthMap[rootNode.ID])
	}
	if depthMap[swNode.ID] != 1 {
		t.Errorf("expected sw1 depth 1, got %d", depthMap[swNode.ID])
	}
	if depthMap[srv1.ID] != 2 || depthMap[srv2.ID] != 2 {
		t.Errorf("expected srv1, srv2 depth 2, got %d, %d", depthMap[srv1.ID], depthMap[srv2.ID])
	}
	if parentMap[swNode.ID] != rootNode.ID {
		t.Errorf("expected sw1 parent root, got %s", parentMap[swNode.ID])
	}
	if parentMap[srv1.ID] != swNode.ID || parentMap[srv2.ID] != swNode.ID {
		t.Errorf("expected servers parent sw1, got %s, %s", parentMap[srv1.ID], parentMap[srv2.ID])
	}
}

func TestAnalyzeFailureDependencies(t *testing.T) {
	cleanup, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("setupTestDB err=%v", err)
	}
	defer cleanup()

	// Network topology:
	// root (IP: 192.168.1.1) -- sw1 -- srv1
	//                        |       \-- srv2
	//                        \-- srv3
	localIPs := getLocalIPs()
	rootIP := "192.168.1.1"
	if len(localIPs) > 0 {
		rootIP = localIPs[0]
	}

	rootNode := &datastore.NodeEnt{Name: "RootHost", IP: rootIP, Icon: "router"}
	swNode := &datastore.NodeEnt{Name: "Switch1", Icon: "switch"}
	srv1 := &datastore.NodeEnt{Name: "Server1", Icon: "server"}
	srv2 := &datastore.NodeEnt{Name: "Server2", Icon: "server"}
	srv3 := &datastore.NodeEnt{Name: "Server3", Icon: "server"}

	_ = datastore.AddNode(rootNode)
	_ = datastore.AddNode(swNode)
	_ = datastore.AddNode(srv1)
	_ = datastore.AddNode(srv2)
	_ = datastore.AddNode(srv3)

	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: rootNode.ID, NodeID2: swNode.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: swNode.ID, NodeID2: srv1.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: swNode.ID, NodeID2: srv2.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: rootNode.ID, NodeID2: srv3.ID})

	// Scenario 1: sw1, srv1, srv2 fail together, plus srv3 fails independently
	events := []*datastore.EventLogEnt{
		{Type: "polling", NodeID: swNode.ID, NodeName: "Switch1", Level: "high", Event: "ping timeout"},
		{Type: "polling", NodeID: srv1.ID, NodeName: "Server1", Level: "high", Event: "ping timeout"},
		{Type: "polling", NodeID: srv2.ID, NodeName: "Server2", Level: "high", Event: "ping timeout"},
		{Type: "polling", NodeID: srv3.ID, NodeName: "Server3", Level: "warn", Event: "snmp error"},
	}

	res := AnalyzeFailureDependencies(events)

	if len(res.RootCauses) != 2 {
		t.Fatalf("expected 2 root causes, got %d (%v)", len(res.RootCauses), res.RootCauses)
	}

	hasSW1 := false
	hasSRV3 := false
	for _, rc := range res.RootCauses {
		if rc == swNode.ID {
			hasSW1 = true
		}
		if rc == srv3.ID {
			hasSRV3 = true
		}
	}
	if !hasSW1 || !hasSRV3 {
		t.Errorf("expected root causes to contain sw1 and srv3, got %v", res.RootCauses)
	}

	if res.ImpactedBy[srv1.ID] != swNode.ID {
		t.Errorf("expected srv1 to be impacted by sw1, got %s", res.ImpactedBy[srv1.ID])
	}
	if res.ImpactedBy[srv2.ID] != swNode.ID {
		t.Errorf("expected srv2 to be impacted by sw1, got %s", res.ImpactedBy[srv2.ID])
	}
	if len(res.ImpactedMap[swNode.ID]) != 2 {
		t.Errorf("expected sw1 to impact 2 nodes, got %d", len(res.ImpactedMap[swNode.ID]))
	}
}

func TestAnalyzeFailureDependencies_MultiTier(t *testing.T) {
	cleanup, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("setupTestDB err=%v", err)
	}
	defer cleanup()

	// Topology: root -- coreSW -- floorSW -- client1
	localIPs := getLocalIPs()
	rootIP := "10.0.0.1"
	if len(localIPs) > 0 {
		rootIP = localIPs[0]
	}

	rootNode := &datastore.NodeEnt{Name: "RootHost", IP: rootIP, Icon: "router"}
	coreSW := &datastore.NodeEnt{Name: "CoreSwitch", Icon: "switch"}
	floorSW := &datastore.NodeEnt{Name: "FloorSwitch", Icon: "switch"}
	client1 := &datastore.NodeEnt{Name: "Client1", Icon: "desktop"}

	_ = datastore.AddNode(rootNode)
	_ = datastore.AddNode(coreSW)
	_ = datastore.AddNode(floorSW)
	_ = datastore.AddNode(client1)

	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: rootNode.ID, NodeID2: coreSW.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: coreSW.ID, NodeID2: floorSW.ID})
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: floorSW.ID, NodeID2: client1.ID})

	// When coreSW, floorSW, client1 all fail
	events := []*datastore.EventLogEnt{
		{Type: "polling", NodeID: coreSW.ID, NodeName: "CoreSwitch", Level: "high"},
		{Type: "polling", NodeID: floorSW.ID, NodeName: "FloorSwitch", Level: "high"},
		{Type: "polling", NodeID: client1.ID, NodeName: "Client1", Level: "high"},
	}

	res := AnalyzeFailureDependencies(events)

	if len(res.RootCauses) != 1 || res.RootCauses[0] != coreSW.ID {
		t.Fatalf("expected only coreSW as root cause, got %v", res.RootCauses)
	}

	if res.ImpactedBy[floorSW.ID] != coreSW.ID {
		t.Errorf("expected floorSW impacted by coreSW, got %s", res.ImpactedBy[floorSW.ID])
	}
	if res.ImpactedBy[client1.ID] != coreSW.ID {
		t.Errorf("expected client1 impacted by coreSW, got %s", res.ImpactedBy[client1.ID])
	}
}

func TestGetNotifyData_WithDependency(t *testing.T) {
	cleanup, err := setupTestDB(t)
	if err != nil {
		t.Fatalf("setupTestDB err=%v", err)
	}
	defer cleanup()

	i18n.SetLang("ja")
	datastore.NotifyConf.Subject = "TWSNMP"
	datastore.NotifyConf.CheckDependency = true
	datastore.NotifyConf.Interval = 60

	nodeA := &datastore.NodeEnt{Name: "SwitchA", Icon: "switch"}
	nodeB := &datastore.NodeEnt{Name: "ServerB", Icon: "server"}
	_ = datastore.AddNode(nodeA)
	_ = datastore.AddNode(nodeB)
	_ = datastore.AddLine(&datastore.LineEnt{NodeID1: nodeA.ID, NodeID2: nodeB.ID})

	now := time.Now().UnixNano()
	// 1. Single failure case (SwitchA fails) with a system event mixed in
	singleEvent := []*datastore.EventLogEnt{
		{Type: "system", Level: "warn", Event: "version update available", Time: now},
		{Type: "polling", NodeID: nodeA.ID, NodeName: "SwitchA", Level: "high", Event: "ping timeout", Time: now},
	}
	dataSingle := getNotifyData(singleEvent, 2)
	if !strings.Contains(dataSingle.failureSubject, "SwitchA") || !strings.Contains(dataSingle.failureSubject, "主原因") {
		t.Errorf("expected single failure subject to contain SwitchA and root cause, got %s", dataSingle.failureSubject)
	}
	if !strings.Contains(dataSingle.failureBody, "[主原因] ping timeout") {
		t.Errorf("expected polling failure body to mention root cause, got %s", dataSingle.failureBody)
	}
	// System event must NOT have [主原因]
	if strings.Contains(dataSingle.failureBody, "[主原因] version update") {
		t.Errorf("non-polling event must NOT be annotated with root cause, got %s", dataSingle.failureBody)
	}

	// 2. Multiple failures (SwitchA and ServerB fail)
	multiEvents := []*datastore.EventLogEnt{
		{Type: "polling", NodeID: nodeA.ID, NodeName: "SwitchA", Level: "high", Event: "ping timeout", Time: now},
		{Type: "polling", NodeID: nodeB.ID, NodeName: "ServerB", Level: "high", Event: "ping timeout", Time: now},
	}
	dataMulti := getNotifyData(multiEvents, 0)
	if !strings.Contains(dataMulti.failureSubject, "他1台影響") {
		t.Errorf("expected multi failure subject to mention other 1 impacted, got %s", dataMulti.failureSubject)
	}
	if !strings.Contains(dataMulti.failureBody, "影響元") || !strings.Contains(dataMulti.failureBody, "SwitchA") {
		t.Errorf("expected multi failure body to mention impacted by SwitchA, got %s", dataMulti.failureBody)
	}
}
