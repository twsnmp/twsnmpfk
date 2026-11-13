package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logreport"
)

func setupMCPTestDB(t *testing.T) (context.Context, context.CancelFunc, *sync.WaitGroup) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	td := t.TempDir()
	datastore.MapConf = datastore.MapConfEnt{}
	datastore.AIConf = datastore.AIConfEnt{}
	if err := datastore.Init(ctx, td, wg); err != nil {
		cancel()
		t.Fatalf("datastore.Init failed: %v", err)
	}
	clearMCPTestEntities()

	t.Cleanup(func() {
		clearMCPTestEntities()
		cancel()
		wg.Wait()
		datastore.CloseDB()
		monitorMu.Lock()
		MonitorDataes = nil
		monitorMu.Unlock()
		pollingHooks.Store(nil)
	})
	return ctx, cancel, wg
}

func clearMCPTestEntities() {
	lineIDs := []string{}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		lineIDs = append(lineIDs, l.ID)
		return true
	})
	for _, id := range lineIDs {
		_ = datastore.DeleteLine(id)
	}
	pollingIDs := []string{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		pollingIDs = append(pollingIDs, p.ID)
		return true
	})
	if len(pollingIDs) > 0 {
		_ = datastore.DeletePollings(pollingIDs)
	}
	nodeIDs := []string{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		nodeIDs = append(nodeIDs, n.ID)
		return true
	})
	for _, id := range nodeIDs {
		_ = datastore.DeleteNode(id)
	}
	netIDs := []string{}
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		netIDs = append(netIDs, n.ID)
		return true
	})
	for _, id := range netIDs {
		_ = datastore.DeleteNetwork(id)
	}
	certIDs := []string{}
	datastore.ForEachCertMonitors(func(c *datastore.CertMonitorEnt) bool {
		certIDs = append(certIDs, c.ID)
		return true
	})
	for _, id := range certIDs {
		datastore.DeleteCertMonitor(id)
	}
}

func connectMCPTestSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "TWSNMP FK Test",
		Version: "v1.0.0",
	}, nil)
	addTools(server)
	addPrompts(server)

	t1, t2 := mcp.NewInMemoryTransports()
	ss, err := server.Connect(context.Background(), t1, nil)
	if err != nil {
		t.Fatalf("server.Connect failed: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-test-client",
		Version: "v1.0.0",
	}, nil)
	cs, err := client.Connect(context.Background(), t2, nil)
	if err != nil {
		_ = ss.Close()
		t.Fatalf("client.Connect failed: %v", err)
	}
	t.Cleanup(func() {
		_ = cs.Close()
		_ = ss.Close()
	})
	return cs
}

func callMCPToolText(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) (string, error) {
	t.Helper()
	if args == nil {
		args = map[string]any{}
	}
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", err
	}
	text := ""
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			text += tc.Text
		}
	}
	if res.IsError {
		return text, fmt.Errorf("tool %s returned error: %s", name, text)
	}
	return text, nil
}

func callMCPToolJSON[T any](t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) T {
	t.Helper()
	text, err := callMCPToolText(t, cs, name, args)
	if err != nil {
		t.Fatalf("tool %s failed unexpectedly: %v", name, err)
	}
	var out T
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("failed to unmarshal JSON from %s (%q): %v", name, text, err)
	}
	return out
}

func TestMCPServerToolsAndSchemas(t *testing.T) {
	setupMCPTestDB(t)
	cs := connectMCPTestSession(t)

	expectedRequired := map[string][]string{
		"get_node_list":              {},
		"get_network_list":           {},
		"get_polling_list":           {},
		"get_polling_log":            {"id"},
		"get_polling_log_data":       {"id"},
		"do_ping":                    {"target"},
		"get_mib_tree":               {},
		"snmpwalk":                   {"target", "mib_object_name"},
		"snmpset":                    {"target", "mib_object_name", "type", "value"},
		"add_node":                   {"name", "ip"},
		"update_node":                {"id"},
		"get_ip_address_list":        {},
		"get_resource_monitor_list":  {},
		"search_event_log":           {},
		"search_syslog":              {},
		"get_syslog_summary":         {},
		"search_snmp_trap_log":       {},
		"get_server_certificate_list": {},
		"add_event_log":               {"event"},
		"get_ip_address_info":         {"ip"},
		"get_mac_address_info":       {"mac"},
		"get_otel_metrics":           {},
		"get_otel_traces":            {},
		"get_sensor_list":            {},
		"get_report_summary":         {"report_type"},
		// Tier 1-3 new tools
		"get_node":                  {"id"},
		"delete_node":               {"id"},
		"get_polling":               {"id"},
		"add_polling":               {"node_id", "name", "type"},
		"update_polling":            {"id"},
		"delete_polling":            {"id"},
		"check_polling":             {},
		"get_line_list":             {},
		"add_line":                  {"node_id1", "node_id2"},
		"delete_line":               {"id"},
		"get_network":               {"id"},
		"add_network":               {"name", "ip"},
		"update_network":            {"id"},
		"delete_network":            {"id"},
		"check_network":             {},
		"update_server_certificate": {},
		"delete_server_certificate": {"id"},
		"get_system_status":         {},
		"search_netflow":            {},
		"search_sflow":              {},
		"search_sflow_counter":      {},
		"get_otel_trace_detail":     {"trace_id"},
		"get_ai_anomaly_list":       {},
		"diagnose_node":             {"target"},
		"search_mib":                {"query"},
		"get_mac_address_list":      {},
		"get_wifi_ap_list":          {},
		"get_bluetooth_device_list": {},
		"get_fdb_table":             {"id"},
		"auto_connect_lines":        {},
		"auto_layout":               {},
		"gnmi_capabilities":         {"node_id"},
		"gnmi_get":                  {"node_id", "path"},
	}

	gotTools := map[string]*mcp.Tool{}
	for tool, err := range cs.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("cs.Tools failed: %v", err)
		}
		gotTools[tool.Name] = tool
	}

	if len(gotTools) != len(expectedRequired) {
		t.Fatalf("registered tool count mismatch: got %d, want %d", len(gotTools), len(expectedRequired))
	}

	for name, wantReq := range expectedRequired {
		tool, ok := gotTools[name]
		if !ok {
			t.Fatalf("expected tool %q not registered", name)
		}
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatalf("marshal schema for %s: %v", name, err)
		}
		var schemaObj struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(raw, &schemaObj); err != nil {
			t.Fatalf("unmarshal schema for %s: %v", name, err)
		}
		if len(schemaObj.Required) != len(wantReq) {
			t.Fatalf("tool %s required fields mismatch: got %v, want %v (schema: %s)", name, schemaObj.Required, wantReq, string(raw))
		}
		gotMap := map[string]bool{}
		for _, r := range schemaObj.Required {
			gotMap[r] = true
		}
		for _, r := range wantReq {
			if !gotMap[r] {
				t.Fatalf("tool %s missing required field %q in %v", name, r, schemaObj.Required)
			}
		}
	}
}

func TestMCPNodeAndPollingCRUDAndBugs(t *testing.T) {
	setupMCPTestDB(t)
	datastore.MapConf.PollInt = 45
	datastore.MapConf.Timeout = 2
	datastore.MapConf.Retry = 1
	cs := connectMCPTestSession(t)

	// B2: add_node creates node and default PING polling with valid interval/timeout/level/state.
	nodeRes := callMCPToolJSON[mcpNodeEnt](t, cs, "add_node", map[string]any{
		"name":        "core-router",
		"ip":          "192.0.2.10",
		"description": "test router",
	})
	if nodeRes.ID == "" || nodeRes.IP != "192.0.2.10" || nodeRes.Icon != "desktop" {
		t.Fatalf("unexpected add_node result: %+v", nodeRes)
	}
	if nodeRes.X != 64 || nodeRes.Y != 64 {
		t.Fatalf("expected default position (64,64), got (%d,%d)", nodeRes.X, nodeRes.Y)
	}
	// Populate secret credentials directly to verify get_node masking.
	nStored := datastore.GetNode(nodeRes.ID)
	if nStored == nil {
		t.Fatalf("stored node not found")
	}
	nStored.Community = "secret-comm"
	nStored.Password = "secret-pass"
	nStored.GNMIPassword = "secret-gnmi"
	nStored.PublicKey = "ssh-ed25519 SECRETKEY"
	nStored.State = "repair"
	if err := datastore.UpdateNode(nStored); err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}

	// B1: get_polling_list checks polling state (not node state) and filters by node_id.
	pollings := callMCPToolJSON[[]mcpPollingEnt](t, cs, "get_polling_list", map[string]any{
		"node_id":      "core-router",
		"state_filter": "^unknown$",
	})
	if len(pollings) != 1 {
		t.Fatalf("expected 1 polling matching state unknown on core-router (node state repair), got %d", len(pollings))
	}
	defaultPoll := datastore.GetPolling(pollings[0].ID)
	if defaultPoll.PollInt != 45 || defaultPoll.Timeout != 2 || defaultPoll.Retry != 1 || defaultPoll.Level != "low" {
		t.Fatalf("unexpected default PING polling params: %+v", defaultPoll)
	}

	// get_node masks credentials and returns pollings.
	detail := callMCPToolJSON[mcpNodeDetailEnt](t, cs, "get_node", map[string]any{
		"id": "core-router",
	})
	if detail.Community != "***" || detail.Password != "***" || detail.GNMIPassword != "***" || detail.PublicKey != "***" {
		t.Fatalf("expected credentials masked, got comm=%q pass=%q gnmi=%q pk=%q", detail.Community, detail.Password, detail.GNMIPassword, detail.PublicKey)
	}
	if strings.Contains(fmt.Sprintf("%+v", detail), "secret-") {
		t.Fatalf("raw secret leaked in get_node response: %+v", detail)
	}
	if len(detail.Pollings) != 1 {
		t.Fatalf("expected 1 polling inside get_node, got %d", len(detail.Pollings))
	}

	// B3: update_node updates IP cleanly and rejects invalid IP.
	if _, err := callMCPToolText(t, cs, "update_node", map[string]any{
		"id": nodeRes.ID,
		"ip": "not-an-ip",
	}); err == nil {
		t.Fatalf("expected update_node with invalid IP to fail")
	}
	updatedNode := callMCPToolJSON[mcpNodeEnt](t, cs, "update_node", map[string]any{
		"id":   "192.0.2.10",
		"name": "core-router-updated",
		"ip":   "192.0.2.20",
	})
	if updatedNode.IP != "192.0.2.20" || updatedNode.Name != "core-router-updated" {
		t.Fatalf("update_node did not update IP/Name: %+v", updatedNode)
	}
	stateChangedFound := false
	datastore.ForEachStateChangedNodes(func(id string) bool {
		if id == nodeRes.ID {
			stateChangedFound = true
		}
		return true
	})
	if !stateChangedFound {
		t.Fatalf("expected SetNodeStateChanged after update_node")
	}

	// Polling CRUD: add_polling, get_polling, update_polling (with StopPolling hook + negative retry), check_polling, delete_polling.
	var stoppedIDs []string
	var checkedPollingIDs []string
	var checkedNodeIDs []string
	allChecked := false
	SetPollingControlHooks(PollingControlHooks{
		CheckPolling: func(id string) bool {
			checkedPollingIDs = append(checkedPollingIDs, id)
			return true
		},
		CheckNode: func(nodeID string) {
			checkedNodeIDs = append(checkedNodeIDs, nodeID)
		},
		CheckAll: func() {
			allChecked = true
		},
		StopPolling: func(p *datastore.PollingEnt) {
			stoppedIDs = append(stoppedIDs, p.ID)
		},
	})

	addedPoll := callMCPToolJSON[mcpPollingDetailEnt](t, cs, "add_polling", map[string]any{
		"node_id":  nodeRes.ID,
		"name":     "gNMI Interface",
		"type":     "gnmi",
		"mode":     "subscribe",
		"params":   "/interfaces",
		"level":    "warn",
		"log_mode": "on_change",
		"retry":    3,
	})
	if addedPoll.ID == "" || addedPoll.LogMode != "on_change" || addedPoll.Retry != 3 {
		t.Fatalf("unexpected add_polling response: %+v", addedPoll)
	}

	fetchedPoll := callMCPToolJSON[mcpPollingDetailEnt](t, cs, "get_polling", map[string]any{
		"id": addedPoll.ID,
	})
	if fetchedPoll.Name != "gNMI Interface" || fetchedPoll.NodeName != "core-router-updated" {
		t.Fatalf("unexpected get_polling response: %+v", fetchedPoll)
	}

	updatedPoll := callMCPToolJSON[mcpPollingDetailEnt](t, cs, "update_polling", map[string]any{
		"id":       addedPoll.ID,
		"level":    "high",
		"poll_int": 120,
		"retry":    -1, // reset to 0
	})
	if updatedPoll.Level != "high" || updatedPoll.PollInt != 120 || updatedPoll.Retry != 0 {
		t.Fatalf("unexpected update_polling result: %+v", updatedPoll)
	}
	if len(stoppedIDs) != 1 || stoppedIDs[0] != addedPoll.ID {
		t.Fatalf("expected StopPolling hook for active gnmi subscribe, got %v", stoppedIDs)
	}

	checkResp := callMCPToolJSON[map[string]any](t, cs, "check_polling", map[string]any{"polling_id": addedPoll.ID})
	if checkResp["Status"] != "triggered" || len(checkedPollingIDs) != 1 {
		t.Fatalf("unexpected check_polling single response: %v, checked=%v", checkResp, checkedPollingIDs)
	}
	_ = callMCPToolJSON[map[string]any](t, cs, "check_polling", map[string]any{"node_id": "core-router-updated"})
	_ = callMCPToolJSON[map[string]any](t, cs, "check_polling", map[string]any{"all": true})
	if len(checkedNodeIDs) != 1 || !allChecked {
		t.Fatalf("expected node & all check hooks triggered: nodes=%v all=%v", checkedNodeIDs, allChecked)
	}

	// Test fallback check_polling when polling engine hooks are nil.
	SetPollingControlHooks(PollingControlHooks{})
	fallbackResp := callMCPToolJSON[map[string]any](t, cs, "check_polling", map[string]any{"polling_id": addedPoll.ID})
	if fallbackResp["Status"] != "scheduled" || fallbackResp["PollingEngineActive"] != false {
		t.Fatalf("unexpected fallback check_polling response: %v", fallbackResp)
	}

	delPoll := callMCPToolJSON[map[string]any](t, cs, "delete_polling", map[string]any{"id": addedPoll.ID})
	if delPoll["Deleted"] != true || datastore.GetPolling(addedPoll.ID) != nil {
		t.Fatalf("delete_polling failed: %v", delPoll)
	}

	// Verify delete_node cascades pollings AND lines immediately.
	node2 := callMCPToolJSON[mcpNodeEnt](t, cs, "add_node", map[string]any{"name": "leaf-switch", "ip": "192.0.2.21"})
	line := callMCPToolJSON[mcpLineEnt](t, cs, "add_line", map[string]any{
		"node_id1": nodeRes.ID,
		"node_id2": node2.ID,
		"port":     "eth0",
	})
	if line.ID == "" || datastore.GetLine(line.ID) == nil {
		t.Fatalf("add_line failed: %+v", line)
	}
	delNode := callMCPToolJSON[map[string]any](t, cs, "delete_node", map[string]any{"id": nodeRes.ID})
	if delNode["Deleted"] != true || datastore.GetNode(nodeRes.ID) != nil {
		t.Fatalf("delete_node failed: %v", delNode)
	}
	if datastore.GetPolling(defaultPoll.ID) != nil {
		t.Fatalf("expected node polling %s deleted with node", defaultPoll.ID)
	}
	if datastore.GetLine(line.ID) != nil {
		t.Fatalf("expected line %s deleted immediately on delete_node", line.ID)
	}
}

func TestMCPLineAndNetworkCRUD(t *testing.T) {
	setupMCPTestDB(t)
	cs := connectMCPTestSession(t)

	netCreated := callMCPToolJSON[mcpNetworkDetailEnt](t, cs, "add_network", map[string]any{
		"name":        "core-sw",
		"ip":          "198.51.100.1",
		"community":   "sw-secret",
		"password":    "v3-secret",
		"unmanaged":   true,
		"description": "access switch",
	})
	if netCreated.ID == "" || !netCreated.Unmanaged {
		t.Fatalf("unexpected add_network result: %+v", netCreated)
	}
	if netCreated.Community != "***" || netCreated.Password != "***" {
		t.Fatalf("expected network credentials masked, got %+v", netCreated)
	}

	netFetched := callMCPToolJSON[mcpNetworkDetailEnt](t, cs, "get_network", map[string]any{
		"id": "NET:" + netCreated.ID,
	})
	if netFetched.Name != "core-sw" || netFetched.Community != "***" || netFetched.Password != "***" {
		t.Fatalf("unexpected get_network result: %+v", netFetched)
	}

	// Update boolean string flags without clobbering others + test invalid bool rejection.
	if _, err := callMCPToolText(t, cs, "update_network", map[string]any{
		"id":       "core-sw",
		"lldp":     "notbool",
	}); err == nil {
		t.Fatalf("expected invalid lldp bool string error")
	}
	netUpdated := callMCPToolJSON[mcpNetworkDetailEnt](t, cs, "update_network", map[string]any{
		"id":        "core-sw",
		"unmanaged": "false",
		"lldp":      "true",
		"h_ports":   48,
	})
	if netUpdated.Unmanaged || !netUpdated.LLDP || netUpdated.HPorts != 48 {
		t.Fatalf("unexpected update_network flags/HPorts: %+v", netUpdated)
	}

	chkNet := callMCPToolJSON[map[string]any](t, cs, "check_network", map[string]any{"id": netCreated.ID})
	if chkNet["Status"] != "triggered" {
		t.Fatalf("unexpected check_network: %v", chkNet)
	}

	nodeA := callMCPToolJSON[mcpNodeEnt](t, cs, "add_node", map[string]any{"name": "app-01", "ip": "198.51.100.10"})
	nodeB := callMCPToolJSON[mcpNodeEnt](t, cs, "add_node", map[string]any{"name": "app-02", "ip": "198.51.100.11"})

	l1 := callMCPToolJSON[mcpLineEnt](t, cs, "add_line", map[string]any{
		"node_id1": nodeA.ID,
		"node_id2": "NET:" + netCreated.ID,
		"info":     "uplink",
	})
	l2 := callMCPToolJSON[mcpLineEnt](t, cs, "add_line", map[string]any{
		"node_id1": nodeA.Name,
		"node_id2": nodeB.Name,
	})
	// Duplicate line should error cleanly.
	if _, err := callMCPToolText(t, cs, "add_line", map[string]any{
		"node_id1": nodeB.ID,
		"node_id2": nodeA.ID,
	}); err == nil {
		t.Fatalf("expected duplicate add_line error")
	}

	lines := callMCPToolJSON[[]mcpLineEnt](t, cs, "get_line_list", map[string]any{
		"node_id": "NET:" + netCreated.ID,
	})
	if len(lines) != 1 || lines[0].ID != l1.ID {
		t.Fatalf("expected 1 line matching network, got %+v", lines)
	}

	delNet := callMCPToolJSON[map[string]any](t, cs, "delete_network", map[string]any{"id": netCreated.ID})
	if delNet["Deleted"] != true || datastore.GetNetwork(netCreated.ID) != nil {
		t.Fatalf("delete_network failed: %v", delNet)
	}
	if datastore.GetLine(l1.ID) != nil {
		t.Fatalf("expected network line %s deleted immediately on delete_network", l1.ID)
	}
	if datastore.GetLine(l2.ID) == nil {
		t.Fatalf("node-to-node line %s should remain after network delete", l2.ID)
	}
	_ = callMCPToolJSON[map[string]any](t, cs, "delete_line", map[string]any{"id": l2.ID})
	if datastore.GetLine(l2.ID) != nil {
		t.Fatalf("delete_line failed for %s", l2.ID)
	}
}

func TestMCPServerCertificateCRUD(t *testing.T) {
	setupMCPTestDB(t)
	cs := connectMCPTestSession(t)

	// Missing entry should not fail with old inverted check.
	if c := datastore.GetCertMonitor("does-not-exist"); c == nil || c.ID != "" {
		t.Fatalf("expected empty CertMonitorEnt for missing ID, got %+v", c)
	}

	created := callMCPToolJSON[mcpServerCertificateEnt](t, cs, "update_server_certificate", map[string]any{
		"target": "example.com",
		"port":   8443,
	})
	if created.ID == "" || created.Server != "example.com" || created.Port != 8443 {
		t.Fatalf("unexpected update_server_certificate create response: %+v", created)
	}

	// Simulate populated certificate info and verify updating target resets it.
	stored := datastore.GetCertMonitor(created.ID)
	if stored == nil || stored.ID == "" {
		t.Fatalf("GetCertMonitor failed to load saved cert monitor %s", created.ID)
	}
	stored.Subject = "CN=example.com"
	stored.State = "normal"
	_ = datastore.SaveCertMonitor(stored)

	updated := callMCPToolJSON[mcpServerCertificateEnt](t, cs, "update_server_certificate", map[string]any{
		"id":     created.ID,
		"target": "new.example.com",
	})
	if updated.Server != "new.example.com" || updated.Subject != "" || updated.State != "unknown" {
		t.Fatalf("expected target change to reset cert metadata, got %+v", updated)
	}

	list := callMCPToolJSON[[]mcpServerCertificateEnt](t, cs, "get_server_certificate_list", nil)
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected 1 cert monitor entry, got %+v", list)
	}

	// Nonexistent update/delete ID returns error.
	if _, err := callMCPToolText(t, cs, "update_server_certificate", map[string]any{"id": "missing", "target": "a.com"}); err == nil {
		t.Fatalf("expected error updating nonexistent cert monitor")
	}
	if _, err := callMCPToolText(t, cs, "delete_server_certificate", map[string]any{"id": "missing"}); err == nil {
		t.Fatalf("expected error deleting nonexistent cert monitor")
	}

	del := callMCPToolJSON[map[string]any](t, cs, "delete_server_certificate", map[string]any{"id": created.ID})
	if del["Deleted"] != true {
		t.Fatalf("delete_server_certificate failed: %v", del)
	}
}

func TestMCPBugsFixedFormattingAndLimits(t *testing.T) {
	ctx, cancel, wg := setupMCPTestDB(t)
	cs := connectMCPTestSession(t)

	// B4: ArpEnt stores FirstTime/LastTime in seconds -> output must be modern year RFC3339, not 1970!
	datastore.UpdateArpEnt("192.0.2.50", "aa:bb:cc:dd:ee:01")
	n := &datastore.NodeEnt{Name: "mac-node", IP: "192.0.2.51", MAC: "aa-bb-cc-dd-ee-02", Vendor: "TestVendor"}
	_ = datastore.AddNode(n)

	ips := callMCPToolJSON[[]mcpIPEnt](t, cs, "get_ip_address_list", nil)
	if len(ips) != 1 || !strings.HasPrefix(ips[0].FirstTime, "202") || !strings.HasPrefix(ips[0].LastTime, "202") {
		t.Fatalf("expected current-era RFC3339 ARP timestamps, got %+v", ips)
	}
	macs := callMCPToolJSON[[]mcpMACListEnt](t, cs, "get_mac_address_list", nil)
	if len(macs) != 2 || macs[0].MAC != "AA:BB:CC:DD:EE:01" || macs[1].MAC != "AA:BB:CC:DD:EE:02" {
		t.Fatalf("unexpected get_mac_address_list result: %+v", macs)
	}

	// B5: Resource monitor timestamps & snapshot helper.
	now := time.Now()
	monitorMu.Lock()
	MonitorDataes = append(MonitorDataes, &MonitorDataEnt{
		Time:         now.UnixNano(),
		CPU:          12.5,
		Mem:          44.0,
		NumGoroutine: 10,
	})
	monitorMu.Unlock()
	mons := callMCPToolJSON[[]mcpResourceMonitorEnt](t, cs, "get_resource_monitor_list", nil)
	if len(mons) != 1 {
		t.Fatalf("expected 1 resource monitor sample, got %d", len(mons))
	}
	parsedMonTime, err := time.Parse(time.RFC3339Nano, mons[0].Time)
	if err != nil || parsedMonTime.Year() < 2024 {
		t.Fatalf("invalid resource monitor timestamp %q: %v", mons[0].Time, err)
	}

	// B8: nameToOID preserves numeric OIDs with canonical leading dot.
	if got := nameToOID("1.3.6.1.2.1.1.1.0"); got != ".1.3.6.1.2.1.1.1.0" {
		t.Fatalf("nameToOID numeric mismatch: got %q", got)
	}
	if got := nameToOID(".1.3.6.1.2.1.1.5.0"); got != ".1.3.6.1.2.1.1.5.0" {
		t.Fatalf("nameToOID leading-dot numeric mismatch: got %q", got)
	}

	// B10: getTimeRange parses duration in both start_time and end_time, and invalid regex errors.
	st, et, err := getTimeRange("-2h", "-1h")
	if err != nil || st >= et {
		t.Fatalf("expected valid relative time window (-2h..-1h), got st=%d et=%d err=%v", st, et, err)
	}
	if _, _, err := getTimeRange("-1h", "-2h"); err == nil {
		t.Fatalf("expected error when start_time > end_time")
	}
	if _, err := callMCPToolText(t, cs, "get_node_list", map[string]any{"name_filter": "[invalid"}); err == nil {
		t.Fatalf("expected invalid regex error from get_node_list")
	}

	// B9: get_polling_log respects exact limit and get_polling_log_data sorts CSV headers deterministically.
	p := &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "latency",
		Type:    "ping",
		State:   "normal",
		LogMode: datastore.LogModeAlways,
		Result:  map[string]any{"rtt": 1.5, "z_metric": 9.0, "a_metric": 2.0},
	}
	_ = datastore.AddPolling(p)
	p.Result = map[string]any{"rtt": 1.5, "z_metric": 9.0, "a_metric": 2.0}
	for i := 0; i < 5; i++ {
		_ = datastore.AddPollingLog(p)
		time.Sleep(2 * time.Millisecond)
	}
	// Cancel context so eventLogger flushes pollingLogCh synchronously while DB remains open.
	cancel()
	wg.Wait()

	logs := callMCPToolJSON[[]mcpPollingLogEnt](t, cs, "get_polling_log", map[string]any{
		"id":    p.ID,
		"limit": 3,
	})
	if len(logs) != 3 {
		t.Fatalf("expected exact limit 3 polling logs (not limit+1), got %d", len(logs))
	}
	csvText, err := callMCPToolText(t, cs, "get_polling_log_data", map[string]any{
		"id":    p.ID,
		"limit": 2,
	})
	if err != nil {
		t.Fatalf("get_polling_log_data failed: %v", err)
	}
	csvLines := strings.Split(strings.TrimSpace(csvText), "\n")
	if len(csvLines) != 3 {
		t.Fatalf("expected header + 2 CSV rows, got %d lines: %q", len(csvLines), csvText)
	}
	if csvLines[0] != "time,state,a_metric,rtt,z_metric" {
		t.Fatalf("expected deterministic sorted CSV header, got %q", csvLines[0])
	}
	_ = ctx
}

func TestMCPCompactPayloadsAndAnalyticsTools(t *testing.T) {
	setupMCPTestDB(t)
	cs := connectMCPTestSession(t)

	// Sensor compact vs include_history
	_ = datastore.SaveSensor(&datastore.SensorEnt{
		ID:       "s1",
		Host:     "edge-01",
		Type:     "netflow",
		State:    "normal",
		Stats:    []datastore.SensorStatsEnt{{Time: 1, Count: 5}},
		Monitors: []datastore.SensorMonitorEnt{{Time: 1, CPU: 10}},
	})
	sensorsCompact := callMCPToolJSON[[]map[string]any](t, cs, "get_sensor_list", nil)
	if len(sensorsCompact) != 1 || sensorsCompact[0]["Stats"] != nil || sensorsCompact[0]["Monitors"] != nil {
		t.Fatalf("expected Stats/Monitors omitted by default, got %+v", sensorsCompact)
	}
	sensorsFull := callMCPToolJSON[[]map[string]any](t, cs, "get_sensor_list", map[string]any{"include_history": true})
	if sensorsFull[0]["Stats"] == nil || sensorsFull[0]["Monitors"] == nil {
		t.Fatalf("expected Stats/Monitors included when include_history=true")
	}

	// WiFi & Bluetooth report entries + compact report_summary
	nowNano := time.Now().UnixNano()
	wifiJSON, _ := json.Marshal(logreport.WifiAPEnt{
		ID: "w1", Host: "ap-sensor", BSSID: "00:11:22:33:44:55", SSID: "CorpNet",
		RSSI: []logreport.RSSIEnt{{Value: -60, Time: nowNano - 1000}, {Value: -52, Time: nowNano}},
		FirstTime: nowNano - 2000, LastTime: nowNano,
	})
	_ = datastore.SaveLogReportData(logreport.KindWifiAP, map[string][]byte{"w1": wifiJSON})

	blePub, _ := json.Marshal(logreport.BlueDeviceEnt{
		ID: "b1", Host: "ble-sensor", Address: "AA:11:22:33:44:55", AddressType: "Public", Name: "BeaconA",
		RSSI: []logreport.RSSIEnt{{Value: -70, Time: nowNano}}, FirstTime: nowNano, LastTime: nowNano,
	})
	bleRand, _ := json.Marshal(logreport.BlueDeviceEnt{
		ID: "b2", Host: "ble-sensor", Address: "FF:11:22:33:44:55", AddressType: "Static Random", Name: "Phone",
		RSSI: []logreport.RSSIEnt{{Value: -80, Time: nowNano}}, FirstTime: nowNano, LastTime: nowNano,
	})
	_ = datastore.SaveLogReportData(logreport.KindBlueDevice, map[string][]byte{"b1": blePub, "b2": bleRand})

	repCompact := callMCPToolJSON[[]map[string]any](t, cs, "get_report_summary", map[string]any{
		"report_type": "wifiAP",
	})
	if len(repCompact) != 1 || repCompact[0]["RSSI"] != nil || repCompact[0]["RSSICount"] != float64(2) {
		t.Fatalf("expected compact RSSICount without RSSI slice, got %+v", repCompact)
	}

	aps := callMCPToolJSON[[]mcpWifiAPEnt](t, cs, "get_wifi_ap_list", map[string]any{"ssid_filter": "Corp"})
	if len(aps) != 1 || aps[0].LatestRSSI != -52 || aps[0].RSSICount != 2 || len(aps[0].RSSI) != 0 {
		t.Fatalf("unexpected get_wifi_ap_list compact result: %+v", aps)
	}
	blePublicOnly := callMCPToolJSON[[]mcpBluetoothDeviceEnt](t, cs, "get_bluetooth_device_list", map[string]any{
		"public_address_only": true,
		"include_history":     true,
	})
	if len(blePublicOnly) != 1 || blePublicOnly[0].Address != "AA:11:22:33:44:55" || len(blePublicOnly[0].RSSI) != 1 {
		t.Fatalf("unexpected get_bluetooth_device_list result: %+v", blePublicOnly)
	}

	// NetFlow / sFlow / sFlowCounter logs
	nfPayload, _ := json.Marshal(datastore.NetFlowEnt{SrcAddr: "10.0.0.1", SrcPort: 12345, DstAddr: "10.0.0.2", DstPort: 443, Protocol: "tcp", Bytes: 1500, Packets: 10})
	sfPayload, _ := json.Marshal(datastore.SFlowEnt{SrcAddr: "10.0.0.3", SrcPort: 5000, DstAddr: "10.0.0.4", DstPort: 53, Protocol: "udp", Bytes: 120})
	sfcPayload, _ := json.Marshal(datastore.SFlowCounterEnt{Remote: "192.0.2.1", Type: "IfCounters", Data: "{}"})
	datastore.SaveLogBuffer([]*datastore.LogEnt{
		{Time: nowNano, Type: "netflow", Log: string(nfPayload)},
		{Time: nowNano, Type: "sflow", Log: string(sfPayload)},
		{Time: nowNano, Type: "sflowCounter", Log: string(sfcPayload)},
	})

	nfRes := callMCPToolJSON[[]mcpNetFlowEnt](t, cs, "search_netflow", map[string]any{"dst_port": 443})
	if len(nfRes) != 1 || nfRes[0].SrcAddr != "10.0.0.1" {
		t.Fatalf("unexpected search_netflow: %+v", nfRes)
	}
	sfRes := callMCPToolJSON[[]mcpSFlowEnt](t, cs, "search_sflow", map[string]any{"protocol_filter": "udp"})
	if len(sfRes) != 1 || sfRes[0].DstPort != 53 {
		t.Fatalf("unexpected search_sflow: %+v", sfRes)
	}
	sfcRes := callMCPToolJSON[[]mcpSFlowCounterEnt](t, cs, "search_sflow_counter", map[string]any{"type_filter": "IfCounters"})
	if len(sfcRes) != 1 || sfcRes[0].Remote != "192.0.2.1" {
		t.Fatalf("unexpected search_sflow_counter: %+v", sfcRes)
	}

	// OTel trace detail
	_ = datastore.UpdateOTelTrace([]*datastore.OTelTraceEnt{{
		Bucket:  "2026-11-14T00:00",
		TraceID: "trace-abc-123",
		Start:   nowNano - 5000,
		End:     nowNano,
		Last:    nowNano,
		Dur:     0.5,
		Spans: []datastore.OTelTraceSpanEnt{
			{SpanID: "s1", Host: "h1", Service: "api", Name: "GET /ok", Start: nowNano - 5000, End: nowNano - 2000},
			{SpanID: "s2", ParentSpanID: "s1", Host: "h1", Service: "db", Name: "SELECT", Start: nowNano - 2000, End: nowNano, Attributes: []string{"error=true"}},
		},
	}})
	traceDetail := callMCPToolJSON[mcpOTelTraceDetail](t, cs, "get_otel_trace_detail", map[string]any{
		"trace_id": "trace-abc-123",
	})
	if traceDetail.NumSpans != 2 || traceDetail.NumErrors != 1 || traceDetail.Bucket != "2026-11-14T00:00" {
		t.Fatalf("unexpected get_otel_trace_detail: %+v", traceDetail)
	}

	// AI anomaly scores
	node := &datastore.NodeEnt{Name: "ai-node", IP: "192.0.2.77"}
	_ = datastore.AddNode(node)
	poll := &datastore.PollingEnt{NodeID: node.ID, Name: "cpu", Type: "snmp", State: "normal"}
	_ = datastore.AddPolling(poll)
	datastore.AIConf = datastore.AIConfEnt{WarnThreshold: 50, LowThreshold: 70, HighThreshold: 90}
	_ = datastore.SaveAIResult(&datastore.AIResult{
		PollingID: poll.ID,
		LastTime:  time.Now().Unix(),
		ScoreData: [][]float64{{1, 20}, {2, 95.5}},
	})
	aiResp := callMCPToolJSON[mcpAIAnomalyListResp](t, cs, "get_ai_anomaly_list", map[string]any{"min_score": 80.0})
	if len(aiResp.Items) != 1 || aiResp.Items[0].Level != "high" || aiResp.Items[0].LatestScore != 95.5 {
		t.Fatalf("unexpected get_ai_anomaly_list: %+v", aiResp)
	}
}

func TestMCPSystemStatusTopologyMIBDiagnoseAndGNMIValidation(t *testing.T) {
	setupMCPTestDB(t)
	cs := connectMCPTestSession(t)

	// Populate MIBInfoMap entry cleanly.
	oldMIBInfo := datastore.MIBInfoMap
	datastore.MIBInfoMap = map[string]*datastore.MIBInfo{
		".1.3.6.1.2.1.1.1": {
			OID:         ".1.3.6.1.2.1.1.1",
			Type:        "DisplayString",
			Status:      "current",
			Description: "A textual description of the entity (sysDescr).",
		},
	}
	defer func() { datastore.MIBInfoMap = oldMIBInfo }()

	mibRes := callMCPToolJSON[[]mcpMIBSearchEnt](t, cs, "search_mib", map[string]any{"query": "sysdescr"})
	if len(mibRes) != 1 || mibRes[0].OID != ".1.3.6.1.2.1.1.1" {
		t.Fatalf("unexpected search_mib result: %+v", mibRes)
	}
	if _, err := callMCPToolText(t, cs, "search_mib", map[string]any{"query": "[bad"}); err == nil {
		t.Fatalf("expected invalid regex error in search_mib")
	}

	// Nodes for system_status + auto_layout + auto_connect_lines
	datastore.MapConf.MCPTransport = "stream"
	datastore.MapConf.MCPEndpoint = "127.0.0.1:8089"
	_ = datastore.AddNode(&datastore.NodeEnt{Name: "gw", IP: "192.0.2.1", Icon: "router", State: "normal", X: 10, Y: 10})
	_ = datastore.AddNode(&datastore.NodeEnt{Name: "srv", IP: "192.0.2.2", Icon: "server", State: "warn", X: 20, Y: 20})

	sysStatus := callMCPToolJSON[mcpSystemStatus](t, cs, "get_system_status", nil)
	if sysStatus.NodeCount != 2 || sysStatus.NumCPU < 1 || sysStatus.NumGoroutine < 1 || len(sysStatus.Receivers) < 9 || sysStatus.MCPTransport != "stream" || sysStatus.MCPEndpoint != "127.0.0.1:8089" {
		t.Fatalf("unexpected get_system_status: %+v", sysStatus)
	}

	layoutRes := callMCPToolJSON[map[string]any](t, cs, "auto_layout", map[string]any{"mode": "cluster"})
	if layoutRes["Mode"] != "cluster" || layoutRes["UndoAvailable"] != true {
		t.Fatalf("unexpected auto_layout result: %v", layoutRes)
	}
	undoRes := callMCPToolJSON[map[string]any](t, cs, "auto_layout", map[string]any{"undo": true})
	if undoRes["Mode"] != "undo" {
		t.Fatalf("unexpected auto_layout undo result: %v", undoRes)
	}

	connRes := callMCPToolJSON[map[string]any](t, cs, "auto_connect_lines", map[string]any{"mode": "speculative"})
	if connRes["Mode"] != "speculative" {
		t.Fatalf("unexpected auto_connect_lines result: %v", connRes)
	}
	if _, err := callMCPToolText(t, cs, "auto_connect_lines", map[string]any{"mode": "invalid"}); err == nil {
		t.Fatalf("expected invalid auto_connect_lines mode error")
	}

	// diagnose_node against local mock HTTP/TCP listener (ping service intentionally unstarted to verify no deadlock)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen failed: %v", err)
	}
	defer ln.Close()
	go func() {
		srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", "MockWeb/1.0")
			_, _ = w.Write([]byte("<html><head><title>Diag OK</title></head></html>"))
		})}
		_ = srv.Serve(ln)
	}()
	tcpPort := ln.Addr().(*net.TCPAddr).Port
	diagNode := &datastore.NodeEnt{
		Name: "diag-local",
		IP:   "127.0.0.1",
		URL:  fmt.Sprintf("http://127.0.0.1:%d/", tcpPort),
	}
	_ = datastore.AddNode(diagNode)

	diagRes := callMCPToolJSON[diagnoseNodeResult](t, cs, "diagnose_node", map[string]any{
		"target":  diagNode.ID,
		"ports":   fmt.Sprintf("%d", tcpPort),
		"timeout": 1,
	})
	if diagRes.ResolvedIP != "127.0.0.1" || len(diagRes.TCPPorts) != 1 {
		t.Fatalf("unexpected diagnose_node result: %+v", diagRes)
	}
	if diagRes.TCPPorts[0].Open {
		if !diagRes.HTTP.Checked || diagRes.HTTP.Title != "Diag OK" || diagRes.HTTP.Server != "MockWeb/1.0" {
			t.Fatalf("unexpected diagnose_node HTTP signature result: %+v", diagRes.HTTP)
		}
	} else if !strings.Contains(diagRes.TCPPorts[0].Error, "operation not permitted") {
		t.Fatalf("unexpected diagnose_node TCP probe failure: %+v", diagRes.TCPPorts[0])
	}

	// gNMI & FDB validation checks
	if _, err := callMCPToolText(t, cs, "gnmi_capabilities", map[string]any{"node_id": "missing"}); err == nil {
		t.Fatalf("expected error on missing node for gnmi_capabilities")
	}
	if _, err := callMCPToolText(t, cs, "gnmi_get", map[string]any{"node_id": diagNode.ID, "path": ""}); err == nil {
		t.Fatalf("expected error on empty path for gnmi_get")
	}
	if _, err := callMCPToolText(t, cs, "get_fdb_table", map[string]any{"id": "missing"}); err == nil {
		t.Fatalf("expected error for missing FDB target / unloaded MIBDB")
	}
}

func TestMCPACLAllowlistAndPrompts(t *testing.T) {
	setupMCPTestDB(t)
	e := echo.New()

	origFrom := datastore.MapConf.MCPFrom
	origToken := datastore.MapConf.MCPToken
	defer func() {
		datastore.MapConf.MCPFrom = origFrom
		datastore.MapConf.MCPToken = origToken
		setMCPAllow()
	}()

	// setMCPAllow must clear stale entries when MCPFrom changes.
	datastore.MapConf.MCPFrom = "192.0.2.10"
	datastore.MapConf.MCPToken = "secret-token"
	setMCPAllow()

	reqOld := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	reqOld.RemoteAddr = "192.0.2.10:12345"
	reqOld.Header.Set("Authorization", "Bearer secret-token")
	if !checkMCPACL(e.NewContext(reqOld, httptest.NewRecorder())) {
		t.Fatalf("expected authorized request from 192.0.2.10 to pass ACL")
	}

	// Exact token comparison must reject substrings / prefixes and support lowercase bearer.
	reqSub := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	reqSub.RemoteAddr = "192.0.2.10:12345"
	reqSub.Header.Set("Authorization", "Bearer secret-tok")
	if checkMCPACL(e.NewContext(reqSub, httptest.NewRecorder())) {
		t.Fatalf("expected partial token to be rejected")
	}
	reqLower := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	reqLower.RemoteAddr = "192.0.2.10:12345"
	reqLower.Header.Set("Authorization", "bearer secret-token")
	if !checkMCPACL(e.NewContext(reqLower, httptest.NewRecorder())) {
		t.Fatalf("expected case-insensitive Bearer prefix to pass ACL")
	}

	// Update MCPFrom to another IP -> previous IP must be cleared!
	datastore.MapConf.MCPFrom = "192.0.2.20"
	setMCPAllow()
	if checkMCPACL(e.NewContext(reqOld, httptest.NewRecorder())) {
		t.Fatalf("expected stale IP 192.0.2.10 to be rejected after setMCPAllow update")
	}

	// Prompt regression checks: filter prompts must preserve instruction prefix (+= not =) and new prompts resolve.
	cs := connectMCPTestSession(t)
	promptsRes, err := cs.ListPrompts(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListPrompts failed: %v", err)
	}
	if len(promptsRes.Prompts) < 40 {
		t.Fatalf("expected extended prompts registered, got %d", len(promptsRes.Prompts))
	}

	for _, tc := range []struct {
		name    string
		args    map[string]string
		contain string
	}{
		{"get_node_list", map[string]string{"state_filter": "high"}, "get_node_list"},
		{"get_network_list", map[string]string{"name_filter": "sw"}, "get_network_list"},
		{"get_polling_list", map[string]string{"state_filter": "warn", "node_id": "n1"}, "get_polling_list"},
		{"search_event_log", map[string]string{"level_filter": "high"}, "search_event_log"},
		{"search_syslog", map[string]string{"host_filter": "gw"}, "search_syslog"},
		{"get_syslog_summary", map[string]string{"top_n": "10"}, "get_syslog_summary"},
		{"search_snmp_trap_log", map[string]string{"from_filter": "10.0.0.1"}, "search_snmp_trap_log"},
		{"troubleshoot_network", map[string]string{"target": "gw"}, "diagnose_node"},
		{"diagnose_node", map[string]string{"target": "127.0.0.1"}, "diagnose_node"},
		{"get_system_status", nil, "get_system_status"},
	} {
		gp, err := cs.GetPrompt(context.Background(), &mcp.GetPromptParams{
			Name:      tc.name,
			Arguments: tc.args,
		})
		if err != nil {
			t.Fatalf("GetPrompt(%s) failed: %v", tc.name, err)
		}
		if len(gp.Messages) == 0 {
			t.Fatalf("GetPrompt(%s) returned no messages", tc.name)
		}
		tcText, ok := gp.Messages[0].Content.(*mcp.TextContent)
		if !ok || !strings.Contains(tcText.Text, tc.contain) {
			t.Fatalf("GetPrompt(%s) text %q missing %q", tc.name, gp.Messages[0].Content, tc.contain)
		}
	}
}
