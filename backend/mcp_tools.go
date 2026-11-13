package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gosnmp/gosnmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/twsnmp/rdap"
	"github.com/xhit/go-str2duration/v2"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/logreport"
	"github.com/twsnmp/twsnmpfk/ping"
)

// Add tools to MCP server
func addTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_list",
		Description: "get node list from TWSNMP",
	}, getNodeList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_network_list",
		Description: "get network list from TWSNMP",
	}, getNetworkList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling_list",
		Description: "get polling list from TWSNMP",
	}, getPollingList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling_log",
		Description: "get polling log from TWSNMP",
	}, getPollingLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling_log_data",
		Description: "get polling log data as CSV from TWSNMP",
	}, getPollingLogData)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "do_ping",
		Description: "do ping",
	}, doPing)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mib_tree",
		Description: "get MIB tree from TWSNMP (prefer search_mib for compact lookups)",
	}, getMIBTree)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "snmpwalk",
		Description: "SNMP walk tool",
	}, snmpWalk)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "snmpset",
		Description: "SNMP set tool",
	}, snmpSet)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_node",
		Description: "add node to TWSNMP",
	}, addNode)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_node",
		Description: "update node name, ip, position, description or icon",
	}, updateNode)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ip_address_list",
		Description: "get IP address list from TWSNMP ARP watch table",
	}, getIPAddressList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_resource_monitor_list",
		Description: "get resource monitor list from TWSNMP",
	}, getResourceMonitorList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_event_log",
		Description: "search event log from TWSNMP",
	}, searchEventLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_syslog",
		Description: "search syslog from TWSNMP",
	}, searchSyslog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_syslog_summary",
		Description: "get syslog pattern summary from TWSNMP",
	}, getSyslogSummary)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_snmp_trap_log",
		Description: "search SNMP trap log from TWSNMP",
	}, searchSnmpTrapLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_server_certificate_list",
		Description: "get server certificate monitor list from TWSNMP",
	}, getServerCertificateList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_event_log",
		Description: "add event log to TWSNMP",
	}, addEventLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ip_address_info",
		Description: "get ip address info.(DNS host,Managed node,Geo location,RDAP)",
	}, getIPInfo)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mac_address_info",
		Description: "get mac address info.(IP,Managed node,Vendor)",
	}, getMACInfo)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_otel_metrics",
		Description: "get OpenTelemetry metrics list or details from TWSNMP",
	}, getOTelMetrics)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_otel_traces",
		Description: "get OpenTelemetry trace list or error traces from TWSNMP",
	}, getOTelTraces)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_sensor_list",
		Description: "get connected sensor list and health status from TWSNMP",
	}, getSensorList)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_report_summary",
		Description: "get report summary (wifiAP, blueDevice, envMonitor, dnsq, winEventID, twAgentAlert, etc.) from TWSNMP",
	}, getReportSummary)

	addCRUDTools(s)
	addAnalyticsTools(s)
}

// Common helpers shared across MCP tool implementations

func mcpTextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

func mcpJSONResult(v any) (*mcp.CallToolResult, any, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	return mcpTextResult(string(j)), nil, nil
}

func compileRegexFilter(fieldName, pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid %s regular expression %q: %w", fieldName, pattern, err)
	}
	return re, nil
}

func clampLimit(val, minVal, maxVal, defaultVal int) int {
	if val <= 0 {
		return defaultVal
	}
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

func formatTimeUnix(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return time.Unix(sec, 0).Format(time.RFC3339)
}

func formatTimeUnixNano(nsec int64) string {
	if nsec <= 0 {
		return ""
	}
	return time.Unix(0, nsec).Format(time.RFC3339Nano)
}

// resolveNode locates a managed node by ID, exact Name, or (when allowIP is true) exact IP.
// When requireUniqueName is true, duplicate name or IP matches return an ambiguity error.
func resolveNode(ref string, allowIP, requireUniqueName bool) (*datastore.NodeEnt, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("node reference must not be empty")
	}
	if n := datastore.GetNode(ref); n != nil {
		return n, nil
	}
	var nameMatches []*datastore.NodeEnt
	var ipMatches []*datastore.NodeEnt
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if n.Name == ref {
			nameMatches = append(nameMatches, n)
		}
		if allowIP && n.IP == ref {
			ipMatches = append(ipMatches, n)
		}
		return true
	})
	if len(nameMatches) == 1 {
		return nameMatches[0], nil
	}
	if len(nameMatches) > 1 {
		if requireUniqueName {
			return nil, fmt.Errorf("ambiguous node name %q matches %d nodes; specify node ID instead", ref, len(nameMatches))
		}
		return nameMatches[0], nil
	}
	if allowIP && len(ipMatches) > 0 {
		if len(ipMatches) > 1 && requireUniqueName {
			return nil, fmt.Errorf("ambiguous node IP %q matches %d nodes; specify node ID instead", ref, len(ipMatches))
		}
		return ipMatches[0], nil
	}
	return nil, fmt.Errorf("node not found: %s", ref)
}

// resolveNetwork locates a network container by ID (with or without "NET:" prefix), unique Name, or IP.
func resolveNetwork(ref string) (*datastore.NetworkEnt, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("network reference must not be empty")
	}
	if n := datastore.GetNetwork(ref); n != nil {
		return n, nil
	}
	var nameMatches []*datastore.NetworkEnt
	var ipMatches []*datastore.NetworkEnt
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		if n.Name == ref {
			nameMatches = append(nameMatches, n)
		}
		if n.IP == ref {
			ipMatches = append(ipMatches, n)
		}
		return true
	})
	if len(nameMatches) == 1 {
		return nameMatches[0], nil
	}
	if len(nameMatches) > 1 {
		return nil, fmt.Errorf("ambiguous network name %q matches %d networks; specify network ID instead", ref, len(nameMatches))
	}
	if len(ipMatches) == 1 {
		return ipMatches[0], nil
	}
	if len(ipMatches) > 1 {
		return nil, fmt.Errorf("ambiguous network IP %q matches %d networks; specify network ID instead", ref, len(ipMatches))
	}
	return nil, fmt.Errorf("network not found: %s", ref)
}

func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

func addMCPEventLog(level, nodeID, nodeName, event string) {
	if level == "" {
		level = "info"
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:     time.Now().UnixNano(),
		Level:    level,
		Type:     "mcp",
		NodeID:   nodeID,
		NodeName: nodeName,
		Event:    event,
	})
}

func parseLogModeInt(s string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "none", "0":
		return datastore.LogModeNone, nil
	case "always", "1":
		return datastore.LogModeAlways, nil
	case "on_change", "onchange", "2":
		return datastore.LogModeOnChange, nil
	case "ai", "3":
		return datastore.LogModeAI, nil
	default:
		return 0, fmt.Errorf("invalid log_mode %q (expected none, always, on_change, ai)", s)
	}
}

func logModeString(m int) string {
	switch m {
	case datastore.LogModeAlways:
		return "always"
	case datastore.LogModeOnChange:
		return "on_change"
	case datastore.LogModeAI:
		return "ai"
	default:
		return "none"
	}
}

// get_node_list tool
type mcpNodeEnt struct {
	ID          string
	Name        string
	IP          string
	MAC         string
	State       string
	X           int
	Y           int
	Icon        string
	Description string
}

type getNodeListParams struct {
	NameFilter  string `json:"name_filter,omitempty" jsonschema:"name_filter specifies the search criteria for node names using regular expressions. If blank, all nodes are searched."`
	IPFilter    string `json:"ip_filter,omitempty" jsonschema:"ip_filter specifies the search criteria for node IP address using regular expressions. If blank, all nodes are searched."`
	StateFilter string `json:"state_filter,omitempty" jsonschema:"state_filter uses a regular expression to specify search criteria for node state names(normal,warn,low,high,repair,unknown). If blank, all nodes are searched."`
}

func getNodeList(ctx context.Context, req *mcp.CallToolRequest, args getNodeListParams) (*mcp.CallToolResult, any, error) {
	name, err := compileRegexFilter("name_filter", args.NameFilter)
	if err != nil {
		return nil, nil, err
	}
	ip, err := compileRegexFilter("ip_filter", args.IPFilter)
	if err != nil {
		return nil, nil, err
	}
	state, err := compileRegexFilter("state_filter", args.StateFilter)
	if err != nil {
		return nil, nil, err
	}
	list := []mcpNodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if name != nil && !name.MatchString(n.Name) {
			return true
		}
		if ip != nil && !ip.MatchString(n.IP) {
			return true
		}
		if state != nil && !state.MatchString(n.State) {
			return true
		}
		list = append(list, mcpNodeEnt{
			ID:          n.ID,
			Name:        n.Name,
			IP:          n.IP,
			MAC:         n.MAC,
			X:           n.X,
			Y:           n.Y,
			Icon:        n.Icon,
			Description: n.Descr,
			State:       n.State,
		})
		return true
	})
	return mcpJSONResult(&list)
}

// get_network_list tool
type mcpNetworkEnt struct {
	ID          string
	Name        string
	IP          string
	Ports       []string
	X           int
	Y           int
	Description string
	Unmanaged   bool
	Error       string
}

type getNetworkListParams struct {
	NameFilter string `json:"name_filter,omitempty" jsonschema:"name_filter specifies the search criteria for network names using regular expressions. If blank, all networks are searched."`
	IPFilter   string `json:"ip_filter,omitempty" jsonschema:"ip_filter specifies the search criteria for network IP address using regular expressions. If blank, all networks are searched."`
}

func getNetworkList(ctx context.Context, req *mcp.CallToolRequest, args getNetworkListParams) (*mcp.CallToolResult, any, error) {
	name, err := compileRegexFilter("name_filter", args.NameFilter)
	if err != nil {
		return nil, nil, err
	}
	ip, err := compileRegexFilter("ip_filter", args.IPFilter)
	if err != nil {
		return nil, nil, err
	}
	list := []mcpNetworkEnt{}
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		if name != nil && !name.MatchString(n.Name) {
			return true
		}
		if ip != nil && !ip.MatchString(n.IP) {
			return true
		}
		ports := []string{}
		for _, p := range n.Ports {
			ports = append(ports, fmt.Sprintf("%s=%s", p.Name, p.State))
		}
		list = append(list, mcpNetworkEnt{
			ID:          n.ID,
			Name:        n.Name,
			IP:          n.IP,
			X:           n.X,
			Y:           n.Y,
			Description: n.Descr,
			Ports:       ports,
			Unmanaged:   n.Unmanaged,
			Error:       n.Error,
		})
		return true
	})
	return mcpJSONResult(&list)
}

// get_polling_list tool
type mcpPollingEnt struct {
	ID       string
	Name     string
	NodeID   string
	NodeName string
	Type     string
	Level    string
	State    string
	Logging  bool
	LastTime string
	Result   map[string]any
}

type getPollingListParams struct {
	TypeFilter     string `json:"type_filter,omitempty" jsonschema:"type_filter uses a regular expression to specify search criteria for polling type names. If blank, all pollings are searched. Type names can be ping,tcp,http,dns,twsnmp,syslog,stun,snmp,etc."`
	NameFilter     string `json:"name_filter,omitempty" jsonschema:"name_filter specifies the search criteria for polling names using regular expressions. If blank, all pollings are searched."`
	NodeNameFilter string `json:"node_name_filter,omitempty" jsonschema:"node_name_filter specifies the search criteria for node names of polling using regular expressions. If blank, all pollings are searched."`
	StateFilter    string `json:"state_filter,omitempty" jsonschema:"state_filter uses a regular expression to specify search criteria for polling state names. If blank, all pollings are searched. State names can be normal,warn,low,high,repair,unknown"`
	NodeID         string `json:"node_id,omitempty" jsonschema:"Optional exact node ID or unique node name to filter pollings for a specific node."`
}

func getPollingList(ctx context.Context, req *mcp.CallToolRequest, args getPollingListParams) (*mcp.CallToolResult, any, error) {
	name, err := compileRegexFilter("name_filter", args.NameFilter)
	if err != nil {
		return nil, nil, err
	}
	nodeName, err := compileRegexFilter("node_name_filter", args.NodeNameFilter)
	if err != nil {
		return nil, nil, err
	}
	typeFilter, err := compileRegexFilter("type_filter", args.TypeFilter)
	if err != nil {
		return nil, nil, err
	}
	state, err := compileRegexFilter("state_filter", args.StateFilter)
	if err != nil {
		return nil, nil, err
	}
	targetNodeID := ""
	if args.NodeID != "" {
		n, err := resolveNode(args.NodeID, false, false)
		if err != nil {
			return nil, nil, err
		}
		targetNodeID = n.ID
	}
	list := []mcpPollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if targetNodeID != "" && p.NodeID != targetNodeID {
			return true
		}
		if name != nil && !name.MatchString(p.Name) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(p.Type) {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		if nodeName != nil && !nodeName.MatchString(n.Name) {
			return true
		}
		if state != nil && !state.MatchString(p.State) {
			return true
		}
		list = append(list, mcpPollingEnt{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type,
			Logging:  p.LogMode > 0,
			NodeID:   p.NodeID,
			NodeName: n.Name,
			LastTime: formatTimeUnixNano(p.LastTime),
			Level:    p.Level,
			State:    p.State,
			Result:   p.Result,
		})
		return true
	})
	return mcpJSONResult(&list)
}

// get_polling_log / get_polling_log_data tools
type mcpPollingLogEnt struct {
	Time   string
	State  string
	Result map[string]any
}

type getPollingLogParams struct {
	ID    string `json:"id" jsonschema:"The ID of the polling to retrieve the polling log"`
	Limit int    `json:"limit,omitempty" jsonschema:"Limit on number of logs retrieved. min 1, max 2000 (default 100)"`
}

func getPollingLog(ctx context.Context, req *mcp.CallToolRequest, args getPollingLogParams) (*mcp.CallToolResult, any, error) {
	id := args.ID
	if id == "" {
		return nil, nil, fmt.Errorf("no id")
	}
	polling := datastore.GetPolling(id)
	if polling == nil {
		return nil, nil, fmt.Errorf("polling not found")
	}
	limit := clampLimit(args.Limit, 1, 2000, 100)
	list := []mcpPollingLogEnt{}
	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		list = append(list, mcpPollingLogEnt{
			Time:   formatTimeUnixNano(l.Time),
			State:  l.State,
			Result: l.Result,
		})
		return len(list) < limit
	})
	return mcpJSONResult(&list)
}

func getPollingLogData(ctx context.Context, req *mcp.CallToolRequest, args getPollingLogParams) (*mcp.CallToolResult, any, error) {
	id := args.ID
	if id == "" {
		return nil, nil, fmt.Errorf("no id")
	}
	polling := datastore.GetPolling(id)
	if polling == nil {
		return nil, nil, fmt.Errorf("polling not found")
	}
	limit := clampLimit(args.Limit, 1, 2000, 100)
	list := []mcpPollingLogEnt{}
	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		list = append(list, mcpPollingLogEnt{
			Time:   formatTimeUnixNano(l.Time),
			State:  l.State,
			Result: l.Result,
		})
		return len(list) < limit
	})
	if len(list) < 1 {
		return nil, nil, fmt.Errorf("polling log not found")
	}
	keySet := make(map[string]bool)
	for _, entry := range list {
		for k, v := range entry.Result {
			if k == "lastTime" {
				continue
			}
			if _, ok := v.(float64); ok {
				keySet[k] = true
			}
		}
	}
	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	csv := []string{"time,state," + strings.Join(keys, ",")}
	if len(keys) == 0 {
		csv = []string{"time,state"}
	}
	for _, l := range list {
		s := fmt.Sprintf("%s,%s", l.Time, l.State)
		for _, k := range keys {
			if v, ok := l.Result[k].(float64); ok {
				s += "," + fmt.Sprintf("%f", v)
			} else {
				s += ","
			}
		}
		csv = append(csv, s)
	}
	return mcpTextResult(strings.Join(csv, "\n")), nil, nil
}

// do_ping tool
type mcpPingEnt struct {
	Result       string `json:"Result"`
	Time         string `json:"Time"`
	RTT          string `json:"RTT"`
	RTTNano      int64  `json:"RTTNano"`
	Size         int    `json:"Size"`
	TTL          int    `json:"TTL"`
	ResponseFrom string `json:"ResponseFrom"`
	Location     string `json:"Location"`
}

type doPingParams struct {
	Target  string `json:"target" jsonschema:"ping target ip address, node ID, node name, or host name"`
	Size    int    `json:"size,omitempty" jsonschema:"ping packet size (1-1500, default 64)"`
	TTL     int    `json:"ttl,omitempty" jsonschema:"IP packet TTL (1-255, default 254)"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"timeout sec of ping (1-10, default 3)"`
}

func doPing(ctx context.Context, req *mcp.CallToolRequest, args doPingParams) (*mcp.CallToolResult, any, error) {
	target := getTargetIP(args.Target)
	if target == "" {
		return nil, nil, fmt.Errorf("target ip not found")
	}
	timeout := clampLimit(args.Timeout, 1, 10, 3)
	size := clampLimit(args.Size, 1, 1500, 64)
	ttl := clampLimit(args.TTL, 1, 255, 254)
	pe := ping.DoPing(target, timeout, 0, size, ttl)
	res := mcpPingEnt{
		Result:       pe.Stat.String(),
		Time:         time.Now().Format(time.RFC3339),
		RTT:          time.Duration(pe.Time).String(),
		Size:         pe.Size,
		ResponseFrom: pe.RecvSrc,
		TTL:          pe.RecvTTL,
		RTTNano:      pe.Time,
	}
	if pe.RecvSrc != "" {
		res.Location = datastore.GetLoc(pe.RecvSrc)
	}
	return mcpJSONResult(&res)
}

// getTargetIP resolves an IP literal, node ID/name, or DNS hostname into an IP string.
func getTargetIP(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	if ip := net.ParseIP(target); ip != nil {
		return ip.String()
	}
	if n := datastore.GetNode(target); n != nil {
		return n.IP
	}
	if n := datastore.FindNodeFromName(target); n != nil {
		return n.IP
	}
	if ips, err := net.LookupIP(target); err == nil {
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				return v4.String()
			}
		}
		for _, ip := range ips {
			if !ip.IsUnspecified() {
				return ip.String()
			}
		}
	}
	return ""
}

// get_mib_tree tool
type getMIBTreeParams struct {
	RootOID  string `json:"root_oid,omitempty" jsonschema:"Optional MIB object name or OID subtree root to filter (e.g. .1.3.6.1.2.1.1 or system). If omitted, returns top tree."`
	MaxDepth int    `json:"max_depth,omitempty" jsonschema:"Optional maximum recursion depth when filtering or browsing the MIB tree (0 = unlimited)."`
}

func cloneMIBSubtree(n *datastore.MIBTreeEnt, depth, maxDepth int) *datastore.MIBTreeEnt {
	cp := &datastore.MIBTreeEnt{
		OID:     n.OID,
		Name:    n.Name,
		MIBInfo: n.MIBInfo,
	}
	if maxDepth > 0 && depth >= maxDepth {
		return cp
	}
	for _, ch := range n.Children {
		cp.Children = append(cp.Children, cloneMIBSubtree(ch, depth+1, maxDepth))
	}
	return cp
}

func findMIBSubtree(nodes []*datastore.MIBTreeEnt, targetOID, targetName string) []*datastore.MIBTreeEnt {
	for _, n := range nodes {
		if (targetOID != "" && n.OID == targetOID) || (targetName != "" && n.Name == targetName) {
			return []*datastore.MIBTreeEnt{n}
		}
		if found := findMIBSubtree(n.Children, targetOID, targetName); len(found) > 0 {
			return found
		}
	}
	return nil
}

func getMIBTree(ctx context.Context, req *mcp.CallToolRequest, args getMIBTreeParams) (*mcp.CallToolResult, any, error) {
	roots := datastore.MIBTree
	if args.RootOID != "" {
		oid := nameToOID(args.RootOID)
		roots = findMIBSubtree(datastore.MIBTree, oid, args.RootOID)
		if len(roots) == 0 {
			roots = []*datastore.MIBTreeEnt{}
		}
	}
	if args.MaxDepth > 0 || args.RootOID != "" {
		out := make([]*datastore.MIBTreeEnt, 0, len(roots))
		for _, r := range roots {
			out = append(out, cloneMIBSubtree(r, 1, args.MaxDepth))
		}
		return mcpJSONResult(&out)
	}
	return mcpJSONResult(&roots)
}

type mcpMIBEnt struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type snmpWalkParams struct {
	Target        string `json:"target" jsonschema:"Target IP address, hostname, node ID, or node name"`
	MIBObjectName string `json:"mib_object_name" jsonschema:"MIB object name or numeric OID"`
	Community     string `json:"community,omitempty" jsonschema:"community name for snmp v2c mode"`
	User          string `json:"user,omitempty" jsonschema:"User name for snmp v3 mode"`
	Password      string `json:"password,omitempty" jsonschema:"Password for snmp v3 mode"`
	SnmpMode      string `json:"snmp_mode,omitempty" jsonschema:"snmp mode (v2c,v3auth,v3authpriv,v3authprivex,v3sha256aes128,v3sha512aes256)"`
	Port          int    `json:"port,omitempty" jsonschema:"SNMP UDP port (default node SnmpPort or 161)"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of variables to return (1-50000, default unlimited)"`
}

func buildSNMPAgentFromParams(target, community, user, password, snmpMode string, port int) (*gosnmp.GoSNMP, *datastore.NodeEnt, error) {
	if target == "" {
		return nil, nil, fmt.Errorf("no target")
	}
	var resolvedNode *datastore.NodeEnt
	resolvedIP := ""
	if n, err := resolveNode(target, true, false); err == nil && n != nil {
		resolvedNode = n
		if community == "" {
			community = n.Community
		}
		if user == "" {
			user = n.User
		}
		if password == "" {
			password = n.Password
		}
		if snmpMode == "" {
			snmpMode = n.SnmpMode
		}
		if port <= 0 && n.SnmpPort > 0 {
			port = n.SnmpPort
		}
		resolvedIP = n.IP
	} else {
		resolvedIP = getTargetIP(target)
		if resolvedIP == "" {
			return nil, nil, fmt.Errorf("target not found")
		}
	}
	if port <= 0 || port > 65535 {
		port = 161
	}
	timeout := datastore.MapConf.Timeout
	if timeout <= 0 {
		timeout = 3
	}
	agent := &gosnmp.GoSNMP{
		Target:    resolvedIP,
		Port:      uint16(port),
		Transport: "udp",
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(timeout) * time.Second,
		Retries:   datastore.MapConf.Retry,
		MaxOids:   gosnmp.MaxOids,
	}
	switch snmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: password,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        password,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        password,
		}
	case "v3sha256aes128":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: password,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        password,
		}
	case "v3sha512aes256":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 user,
			AuthenticationProtocol:   gosnmp.SHA512,
			AuthenticationPassphrase: password,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        password,
		}
	}
	return agent, resolvedNode, nil
}

func snmpWalk(ctx context.Context, req *mcp.CallToolRequest, args snmpWalkParams) (*mcp.CallToolResult, any, error) {
	name := args.MIBObjectName
	if name == "" {
		return nil, nil, fmt.Errorf("no mib_object_name")
	}
	agent, _, err := buildSNMPAgentFromParams(args.Target, args.Community, args.User, args.Password, args.SnmpMode, args.Port)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit > 50000 {
		limit = 50000
	}
	res := []mcpMIBEnt{}
	if err := agent.Connect(); err != nil {
		return nil, nil, err
	}
	defer agent.Conn.Close()
	errLimitReached := fmt.Errorf("walk limit reached")
	err = agent.Walk(nameToOID(name), func(variable gosnmp.SnmpPDU) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		varName := datastore.MIBDB.OIDToName(variable.Name)
		value := datastore.GetMIBValueString(varName, &variable, false)
		res = append(res, mcpMIBEnt{
			Name:  varName,
			Value: value,
		})
		if limit > 0 && len(res) >= limit {
			return errLimitReached
		}
		return nil
	})
	if err != nil && err != errLimitReached {
		return nil, nil, err
	}
	return mcpJSONResult(&res)
}

type snmpSetParams struct {
	Target        string `json:"target" jsonschema:"Target IP address, hostname, node ID, or node name"`
	MIBObjectName string `json:"mib_object_name" jsonschema:"mib object name or numeric OID"`
	Community     string `json:"community,omitempty" jsonschema:"community name for snmp v2c mode"`
	User          string `json:"user,omitempty" jsonschema:"User name for snmp v3 mode"`
	Password      string `json:"password,omitempty" jsonschema:"Password for snmp v3 mode"`
	SnmpMode      string `json:"snmp_mode,omitempty" jsonschema:"snmp mode (v2c,v3auth,v3authpriv,v3authprivex,v3sha256aes128,v3sha512aes256)"`
	Port          int    `json:"port,omitempty" jsonschema:"SNMP UDP port (default node SnmpPort or 161)"`
	Type          string `json:"type" jsonschema:"Type of set value (integer or string)"`
	Value         string `json:"value" jsonschema:"Set value"`
}

func snmpSet(ctx context.Context, req *mcp.CallToolRequest, args snmpSetParams) (*mcp.CallToolResult, any, error) {
	name := args.MIBObjectName
	if name == "" {
		return nil, nil, fmt.Errorf("no mib_object_name")
	}
	agent, resolvedNode, err := buildSNMPAgentFromParams(args.Target, args.Community, args.User, args.Password, args.SnmpMode, args.Port)
	if err != nil {
		return nil, nil, err
	}
	setPDU := []gosnmp.SnmpPDU{}
	switch strings.ToLower(args.Type) {
	case "integer":
		i, err := strconv.Atoi(args.Value)
		if err != nil {
			return nil, nil, err
		}
		setPDU = append(setPDU, gosnmp.SnmpPDU{
			Name:  nameToOID(name),
			Type:  gosnmp.Integer,
			Value: i,
		})
	case "string", "":
		setPDU = append(setPDU, gosnmp.SnmpPDU{
			Name:  nameToOID(name),
			Type:  gosnmp.OctetString,
			Value: []byte(args.Value),
		})
	default:
		return nil, nil, fmt.Errorf("unsupported snmp set type %q (expected integer or string)", args.Type)
	}
	if err := agent.Connect(); err != nil {
		return nil, nil, err
	}
	defer agent.Conn.Close()
	r, err := agent.Set(setPDU)
	if err != nil {
		return nil, nil, err
	}
	if r.Error != gosnmp.NoError {
		return nil, nil, fmt.Errorf("snmp set %s", r.Error.String())
	}
	res := []mcpMIBEnt{}
	for _, variable := range r.Variables {
		varName := datastore.MIBDB.OIDToName(variable.Name)
		value := datastore.GetMIBValueString(varName, &variable, false)
		res = append(res, mcpMIBEnt{
			Name:  varName,
			Value: value,
		})
	}
	nodeID, nodeName := "", args.Target
	if resolvedNode != nil {
		nodeID = resolvedNode.ID
		nodeName = resolvedNode.Name
	}
	addMCPEventLog("info", nodeID, nodeName, fmt.Sprintf("snmpset %s (%s=%s)", name, args.Type, args.Value))
	return mcpJSONResult(&res)
}

// add_node tool
type addNodeParams struct {
	Name        string `json:"name" jsonschema:"node name"`
	IP          string `json:"ip" jsonschema:"node ip address"`
	Icon        string `json:"icon,omitempty" jsonschema:"icon of node (default desktop)"`
	Description string `json:"description,omitempty" jsonschema:"description of node"`
	X           int    `json:"x,omitempty" jsonschema:"x position of node (1-1000, default 64)"`
	Y           int    `json:"y,omitempty" jsonschema:"y position of node (1-1000, default 64)"`
}

func addNode(ctx context.Context, req *mcp.CallToolRequest, args addNodeParams) (*mcp.CallToolResult, any, error) {
	icon := args.Icon
	if icon == "" {
		icon = "desktop"
	}
	name := strings.TrimSpace(args.Name)
	if name == "" {
		return nil, nil, fmt.Errorf("node name is empty")
	}
	ip := strings.TrimSpace(args.IP)
	if ip == "" {
		return nil, nil, fmt.Errorf("IP is empty")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, nil, fmt.Errorf("invalid IP address: %s", ip)
	}
	ip = parsedIP.String()
	x := args.X
	if x < 1 || x > 1000 {
		x = 64
	}
	y := args.Y
	if y < 1 || y > 1000 {
		y = 64
	}
	n := &datastore.NodeEnt{
		Name:  name,
		IP:    ip,
		Icon:  icon,
		X:     x,
		Y:     y,
		Descr: args.Description,
		State: "unknown",
	}
	if err := datastore.AddNode(n); err != nil {
		return nil, nil, err
	}
	pollInt := datastore.MapConf.PollInt
	if pollInt <= 0 {
		pollInt = 60
	}
	timeout := datastore.MapConf.Timeout
	if timeout <= 0 {
		timeout = 1
	}
	retry := datastore.MapConf.Retry
	if retry < 0 {
		retry = 0
	}
	if err := datastore.AddPolling(&datastore.PollingEnt{
		Name:    "PING",
		Type:    "ping",
		NodeID:  n.ID,
		PollInt: pollInt,
		Timeout: timeout,
		Retry:   retry,
		State:   "unknown",
		Level:   "low",
	}); err != nil {
		log.Printf("mcp add_node add default polling err=%v", err)
	}
	addMCPEventLog("info", n.ID, n.Name, fmt.Sprintf("Add node via MCP: %s (%s)", n.Name, n.IP))
	return mcpJSONResult(&mcpNodeEnt{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Descr,
		IP:          n.IP,
		MAC:         n.MAC,
		State:       n.State,
		X:           n.X,
		Y:           n.Y,
		Icon:        n.Icon,
	})
}

// update_node tool
type updateNodeParams struct {
	ID          string `json:"id" jsonschema:"node id or current name or current ip"`
	Name        string `json:"name,omitempty" jsonschema:"new node name or empty"`
	IP          string `json:"ip,omitempty" jsonschema:"new ip address or empty"`
	Icon        string `json:"icon,omitempty" jsonschema:"new icon or empty"`
	Description string `json:"description,omitempty" jsonschema:"description of node"`
	X           int    `json:"x,omitempty" jsonschema:"x position of node"`
	Y           int    `json:"y,omitempty" jsonschema:"y position of node"`
}

func updateNode(ctx context.Context, req *mcp.CallToolRequest, args updateNodeParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNode(args.ID, true, true)
	if err != nil {
		return nil, nil, err
	}
	if args.X > 0 {
		n.X = args.X
	}
	if args.Y > 0 {
		n.Y = args.Y
	}
	if args.Icon != "" {
		n.Icon = args.Icon
	}
	if args.Description != "" {
		n.Descr = args.Description
	}
	if args.Name != "" {
		n.Name = strings.TrimSpace(args.Name)
	}
	ipChanged := false
	if args.IP != "" {
		parsedIP := net.ParseIP(strings.TrimSpace(args.IP))
		if parsedIP == nil {
			return nil, nil, fmt.Errorf("invalid IP address: %s", args.IP)
		}
		newIP := parsedIP.String()
		if n.IP != newIP {
			n.IP = newIP
			ipChanged = true
		}
	}
	if err := datastore.UpdateNode(n); err != nil {
		return nil, nil, err
	}
	if ipChanged {
		datastore.SetNodeStateChanged(n.ID)
	}
	addMCPEventLog("info", n.ID, n.Name, fmt.Sprintf("Update node via MCP: %s (%s)", n.Name, n.IP))
	return mcpJSONResult(&mcpNodeEnt{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Descr,
		IP:          n.IP,
		MAC:         n.MAC,
		State:       n.State,
		X:           n.X,
		Y:           n.Y,
		Icon:        n.Icon,
	})
}

// get_ip_address_list tool
type mcpIPEnt struct {
	IP        string
	MAC       string
	Node      string
	Vendor    string
	FirstTime string
	LastTime  string
}

type getIPAddressListParams struct {
	IPFilter     string `json:"ip_filter,omitempty" jsonschema:"Regex filter for IP address. If blank, all entries are returned."`
	MACFilter    string `json:"mac_filter,omitempty" jsonschema:"Regex filter for MAC address. If blank, all entries are returned."`
	VendorFilter string `json:"vendor_filter,omitempty" jsonschema:"Regex filter for vendor name. If blank, all entries are returned."`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum number of entries to return (1-2000, default 500)."`
}

func getIPAddressList(ctx context.Context, req *mcp.CallToolRequest, args getIPAddressListParams) (*mcp.CallToolResult, any, error) {
	ipF, err := compileRegexFilter("ip_filter", args.IPFilter)
	if err != nil {
		return nil, nil, err
	}
	macF, err := compileRegexFilter("mac_filter", args.MACFilter)
	if err != nil {
		return nil, nil, err
	}
	vendorF, err := compileRegexFilter("vendor_filter", args.VendorFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 2000, 500)
	list := []mcpIPEnt{}
	datastore.ForEachArp(func(l *datastore.ArpEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		vendor := l.Vendor
		if vendor == "" {
			vendor = datastore.FindVendor(l.MAC)
		}
		if ipF != nil && !ipF.MatchString(l.IP) {
			return true
		}
		if macF != nil && !macF.MatchString(l.MAC) {
			return true
		}
		if vendorF != nil && !vendorF.MatchString(vendor) {
			return true
		}
		node := ""
		if l.NodeID != "" {
			if n := datastore.GetNode(l.NodeID); n != nil {
				node = n.Name
			}
		}
		list = append(list, mcpIPEnt{
			IP:        l.IP,
			MAC:       l.MAC,
			Node:      node,
			Vendor:    vendor,
			FirstTime: formatTimeUnix(l.FirstTime),
			LastTime:  formatTimeUnix(l.LastTime),
		})
		return true
	})
	sort.Slice(list, func(i, j int) bool {
		return list[i].IP < list[j].IP
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return mcpJSONResult(&list)
}

// get_resource_monitor_list tool
type mcpResourceMonitorEnt struct {
	Time        string
	CPUUsage    string
	MemoryUsage string
	SwapUsage   string
	DiskUsage   string
	Load        string
}

type getResourceMonitorListParams struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum number of snapshots to return (1-500, default 100)."`
}

func getResourceMonitorList(ctx context.Context, req *mcp.CallToolRequest, args getResourceMonitorListParams) (*mcp.CallToolResult, any, error) {
	list := []mcpResourceMonitorEnt{}
	monList := GetMonitorDataList()
	skip := 30
	if len(monList) < 120 {
		skip = 5
	}
	limit := clampLimit(args.Limit, 1, 500, 100)
	for i, m := range monList {
		if i%skip != 0 {
			continue
		}
		list = append(list, mcpResourceMonitorEnt{
			Time:        formatTimeUnixNano(m.Time),
			CPUUsage:    fmt.Sprintf("%.02f%%", m.CPU),
			MemoryUsage: fmt.Sprintf("%.02f%%", m.Mem),
			SwapUsage:   fmt.Sprintf("%.02f%%", m.Swap),
			DiskUsage:   fmt.Sprintf("%.02f%%", m.Disk),
			Load:        fmt.Sprintf("%.02f", m.Load),
		})
	}
	if len(list) > limit {
		list = list[len(list)-limit:]
	}
	return mcpJSONResult(&list)
}

// search_event_log tool
type mcpEventLogEnt struct {
	Time  string
	Type  string
	Level string
	Node  string
	Event string
}

type searchEventLogParams struct {
	NodeFilter  string `json:"node_filter,omitempty" jsonschema:"node_filter specifies the search criteria for node names using regular expressions. If blank, no filter."`
	TypeFilter  string `json:"type_filter,omitempty" jsonschema:"type_filter specifies the search criteria for type names using regular expressions. If blank, no filter."`
	LevelFilter string `json:"level_filter,omitempty" jsonschema:"level_filter specifies the search criteria for level names using regular expressions. If blank, no filter. Level names can be warn,low,high,debug,info"`
	EventFilter string `json:"event_filter,omitempty" jsonschema:"event_filter specifies the search criteria for events using regular expressions. If blank, no filter."`
	StartTime   string `json:"start_time,omitempty" jsonschema:"start date and time of logs to search or duration from now (default -1h)."`
	EndTime     string `json:"end_time,omitempty" jsonschema:"end date and time of logs to search or duration from now. empty or now is current time."`
	Limit       int    `json:"limit,omitempty" jsonschema:"Limit on number of logs retrieved (1-10000, default 100)."`
}

func searchEventLog(ctx context.Context, req *mcp.CallToolRequest, args searchEventLogParams) (*mcp.CallToolResult, any, error) {
	node, err := compileRegexFilter("node_filter", args.NodeFilter)
	if err != nil {
		return nil, nil, err
	}
	typeFilter, err := compileRegexFilter("type_filter", args.TypeFilter)
	if err != nil {
		return nil, nil, err
	}
	level, err := compileRegexFilter("level_filter", args.LevelFilter)
	if err != nil {
		return nil, nil, err
	}
	event, err := compileRegexFilter("event_filter", args.EventFilter)
	if err != nil {
		return nil, nil, err
	}
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 10000, 100)
	log.Printf("mcp search_event_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpEventLogEnt{}
	_ = datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		if event != nil && !event.MatchString(l.Event) {
			return true
		}
		if level != nil && !level.MatchString(l.Level) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.Type) {
			return true
		}
		if node != nil && !node.MatchString(l.NodeName) {
			return true
		}
		list = append(list, mcpEventLogEnt{
			Time:  formatTimeUnixNano(l.Time),
			Type:  l.Type,
			Level: l.Level,
			Node:  l.NodeName,
			Event: l.Event,
		})
		return len(list) < limit
	})
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(&list)
}

// search_syslog tool
type mcpSyslogEnt struct {
	Time     string
	Level    string
	Host     string
	Type     string
	Tag      string
	Message  string
	Severity int
	Facility int
}

type searchSyslogParams struct {
	LevelFilter   string `json:"level_filter,omitempty" jsonschema:"level_filter specifies the search criteria for level names using regular expressions. If blank, no filter. Level names can be warn,low,high,debug,info"`
	HostFilter    string `json:"host_filter,omitempty" jsonschema:"host_filter specifies the search criteria for host names using regular expressions. If blank, no filter."`
	TagFilter     string `json:"tag_filter,omitempty" jsonschema:"tag_filter specifies the search criteria for tag names using regular expressions. If blank, no filter."`
	MessageFilter string `json:"message_filter,omitempty" jsonschema:"message_filter specifies the search criteria for messages using regular expressions. If blank, no filter."`
	StartTime     string `json:"start_time,omitempty" jsonschema:"start date and time of logs to search or duration from now (default -1h)."`
	EndTime       string `json:"end_time,omitempty" jsonschema:"end date and time of logs to search or duration from now. empty or now is current time."`
	Limit         int    `json:"limit,omitempty" jsonschema:"Limit on number of logs retrieved (1-10000, default 100)."`
}

func searchSyslog(ctx context.Context, req *mcp.CallToolRequest, args searchSyslogParams) (*mcp.CallToolResult, any, error) {
	host, err := compileRegexFilter("host_filter", args.HostFilter)
	if err != nil {
		return nil, nil, err
	}
	tag, err := compileRegexFilter("tag_filter", args.TagFilter)
	if err != nil {
		return nil, nil, err
	}
	level, err := compileRegexFilter("level_filter", args.LevelFilter)
	if err != nil {
		return nil, nil, err
	}
	message, err := compileRegexFilter("message_filter", args.MessageFilter)
	if err != nil {
		return nil, nil, err
	}
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 10000, 100)
	log.Printf("mcp search_syslog limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpSyslogEnt{}
	_ = datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		if message != nil && !message.MatchString(l.Message) {
			return true
		}
		if tag != nil && !tag.MatchString(l.Tag) {
			return true
		}
		if level != nil && !level.MatchString(l.Level) {
			return true
		}
		if host != nil && !host.MatchString(l.Host) {
			return true
		}
		list = append(list, mcpSyslogEnt{
			Time:     formatTimeUnixNano(l.Time),
			Host:     l.Host,
			Level:    l.Level,
			Type:     l.Type,
			Tag:      l.Tag,
			Facility: l.Facility,
			Severity: l.Severity,
			Message:  l.Message,
		})
		return len(list) < limit
	})
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(&list)
}

// get_syslog_summary tool
type mcpSyslogSummaryEnt struct {
	Pattern string
	Count   int
}

type getSyslogSummaryParams struct {
	LevelFilter   string `json:"level_filter,omitempty" jsonschema:"level_filter specifies the search criteria for level names using regular expressions. If blank, no filter. Level names can be warn,low,high,debug,info"`
	HostFilter    string `json:"host_filter,omitempty" jsonschema:"host_filter specifies the search criteria for host names using regular expressions. If blank, no filter."`
	TagFilter     string `json:"tag_filter,omitempty" jsonschema:"tag_filter specifies the search criteria for tag names using regular expressions. If blank, no filter."`
	MessageFilter string `json:"message_filter,omitempty" jsonschema:"message_filter specifies the search criteria for messages using regular expressions. If blank, no filter."`
	StartTime     string `json:"start_time,omitempty" jsonschema:"start date and time of logs to search or duration from now (default -1h)."`
	EndTime       string `json:"end_time,omitempty" jsonschema:"end date and time of logs to search or duration from now. empty or now is current time."`
	TopN          int    `json:"top_n,omitempty" jsonschema:"Top n syslog patterns to return (5-500, default 5)."`
	MaxScan       int    `json:"max_scan,omitempty" jsonschema:"Maximum number of syslog records to scan (1000-2000000, default 200000)."`
}

func getSyslogSummary(ctx context.Context, req *mcp.CallToolRequest, args getSyslogSummaryParams) (*mcp.CallToolResult, any, error) {
	host, err := compileRegexFilter("host_filter", args.HostFilter)
	if err != nil {
		return nil, nil, err
	}
	tag, err := compileRegexFilter("tag_filter", args.TagFilter)
	if err != nil {
		return nil, nil, err
	}
	level, err := compileRegexFilter("level_filter", args.LevelFilter)
	if err != nil {
		return nil, nil, err
	}
	message, err := compileRegexFilter("message_filter", args.MessageFilter)
	if err != nil {
		return nil, nil, err
	}
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	topN := clampLimit(args.TopN, 5, 500, 5)
	maxScan := clampLimit(args.MaxScan, 1000, 2000000, 200000)
	log.Printf("mcp get_syslog_summary topn=%d max_scan=%d st=%v et=%v", topN, maxScan, time.Unix(0, st), time.Unix(0, et))
	patternMap := make(map[string]int)
	scanned := 0
	_ = datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		scanned++
		if scanned > maxScan {
			return false
		}
		if message != nil && !message.MatchString(l.Message) {
			return true
		}
		if tag != nil && !tag.MatchString(l.Tag) {
			return true
		}
		if level != nil && !level.MatchString(l.Level) {
			return true
		}
		if host != nil && !host.MatchString(l.Host) {
			return true
		}
		patternMap[normalizeLog(fmt.Sprintf("%s %s %s %s", l.Host, l.Type, l.Tag, l.Message))]++
		return true
	})
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	list := []mcpSyslogSummaryEnt{}
	for p, c := range patternMap {
		list = append(list, mcpSyslogSummaryEnt{
			Pattern: p,
			Count:   c,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].Pattern < list[j].Pattern
		}
		return list[i].Count > list[j].Count
	})
	if len(list) > topN {
		list = list[:topN]
	}
	return mcpJSONResult(&list)
}

var regNum = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
var regUUID = regexp.MustCompile(`\b[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}\b`)
var regEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
var regIP = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
var regMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b`)

func normalizeLog(s string) string {
	s = regUUID.ReplaceAllString(s, "#UUID#")
	s = regEmail.ReplaceAllString(s, "#EMAIL#")
	s = regIP.ReplaceAllString(s, "#IP#")
	s = regMAC.ReplaceAllString(s, "#MAC#")
	s = regNum.ReplaceAllString(s, "#NUM#")
	return s
}

// search_snmp_trap_log tool
type mcpSNMPTrapLogEnt struct {
	Time        string
	FromAddress string
	TrapType    string
	Variables   string
}

type searchSnmpTrapLogParams struct {
	FromFilter     string `json:"from_filter,omitempty" jsonschema:"from_filter specifies the search criteria for trap sender address using regular expressions. If blank, no filter."`
	TrapTypeFilter string `json:"trap_type_filter,omitempty" jsonschema:"trap_type_filter specifies the search criteria for SNMP trap types using regular expressions. If blank, no filter."`
	VariableFilter string `json:"variable_filter,omitempty" jsonschema:"variable_filter specifies the search criteria for SNMP trap variables using regular expressions. If blank, no filter."`
	StartTime      string `json:"start_time,omitempty" jsonschema:"start date and time of logs to search or duration from now (default -1h)."`
	EndTime        string `json:"end_time,omitempty" jsonschema:"end date and time of logs to search or duration from now. empty or now is current time."`
	Limit          int    `json:"limit,omitempty" jsonschema:"Limit on number of logs retrieved (1-10000, default 100)."`
}

func searchSnmpTrapLog(ctx context.Context, req *mcp.CallToolRequest, args searchSnmpTrapLogParams) (*mcp.CallToolResult, any, error) {
	from, err := compileRegexFilter("from_filter", args.FromFilter)
	if err != nil {
		return nil, nil, err
	}
	trapType, err := compileRegexFilter("trap_type_filter", args.TrapTypeFilter)
	if err != nil {
		return nil, nil, err
	}
	variable, err := compileRegexFilter("variable_filter", args.VariableFilter)
	if err != nil {
		return nil, nil, err
	}
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 10000, 100)
	log.Printf("mcp search_snmp_trap_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpSNMPTrapLogEnt{}
	_ = datastore.ForEachTraps(st, et, func(l *datastore.TrapEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		if from != nil && !from.MatchString(l.FromAddress) {
			return true
		}
		if variable != nil && !variable.MatchString(l.Variables) {
			return true
		}
		if trapType != nil && !trapType.MatchString(l.TrapType) {
			return true
		}
		list = append(list, mcpSNMPTrapLogEnt{
			Time:        formatTimeUnixNano(l.Time),
			FromAddress: l.FromAddress,
			TrapType:    l.TrapType,
			Variables:   l.Variables,
		})
		return len(list) < limit
	})
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(&list)
}

// get_server_certificate_list tool
type mcpServerCertificateEnt struct {
	ID           string
	State        string
	Server       string
	Port         uint16
	Subject      string
	Issuer       string
	SerialNumber string
	Verify       bool
	NotAfter     string
	NotBefore    string
	Error        string
	FirstTime    string
	LastTime     string
}

type getServerCertificateListParams struct{}

func getServerCertificateList(ctx context.Context, req *mcp.CallToolRequest, _ getServerCertificateListParams) (*mcp.CallToolResult, any, error) {
	list := []mcpServerCertificateEnt{}
	datastore.ForEachCertMonitors(func(c *datastore.CertMonitorEnt) bool {
		list = append(list, mcpServerCertificateEnt{
			ID:           c.ID,
			State:        c.State,
			Server:       c.Target,
			Port:         c.Port,
			Subject:      c.Subject,
			Issuer:       c.Issuer,
			SerialNumber: c.SerialNumber,
			Verify:       c.Verify,
			NotBefore:    formatTimeUnix(c.NotBefore),
			NotAfter:     formatTimeUnix(c.NotAfter),
			Error:        c.Error,
			FirstTime:    formatTimeUnixNano(c.FirstTime),
			LastTime:     formatTimeUnixNano(c.LastTime),
		})
		return true
	})
	return mcpJSONResult(&list)
}

// add_event_log tool
type addEventLogParams struct {
	Level string `json:"level,omitempty" jsonschema:"Level of event (info,normal,warn,low,high; default info)"`
	Node  string `json:"node,omitempty" jsonschema:"Node ID or node name associated with the event"`
	Event string `json:"event" jsonschema:"Event log contents"`
}

func addEventLog(ctx context.Context, req *mcp.CallToolRequest, args addEventLogParams) (*mcp.CallToolResult, any, error) {
	level := strings.TrimSpace(args.Level)
	if level == "" {
		level = "info"
	}
	event := args.Event
	if event == "" {
		return nil, nil, fmt.Errorf("event is required")
	}
	node := strings.TrimSpace(args.Node)
	nodeID := ""
	if node != "" {
		if n, err := resolveNode(node, false, false); err == nil && n != nil {
			nodeID = n.ID
			node = n.Name
		}
	}
	addMCPEventLog(level, nodeID, node, event)
	return mcpTextResult("ok"), nil, nil
}

// get_ip_address_info tool
type getIPInfoParams struct {
	IP string `json:"ip" jsonschema:"IP address"`
}

type mcpIPInfoEnt struct {
	IP              string
	Node            string
	DNSNames        []string
	Location        string
	RDAPIPVersion   string
	RDAPType        string
	RDAPHandle      string
	RDAPName        string
	RDAPCountry     string
	RDAPWhoisServer string
}

func getIPInfo(ctx context.Context, req *mcp.CallToolRequest, args getIPInfoParams) (*mcp.CallToolResult, any, error) {
	ipStr := strings.TrimSpace(args.IP)
	parsed := net.ParseIP(ipStr)
	if parsed == nil {
		return nil, nil, fmt.Errorf("invalid IP address: %s", args.IP)
	}
	ipStr = parsed.String()
	info := new(mcpIPInfoEnt)
	info.IP = ipStr
	if n := datastore.FindNodeFromIP(ipStr); n != nil {
		info.Node = n.Name
	}
	r := &net.Resolver{}
	dnsCtx, dnsCancel := context.WithTimeout(ctx, 2*time.Second)
	defer dnsCancel()
	if names, err := r.LookupAddr(dnsCtx, ipStr); err == nil && len(names) > 0 {
		info.DNSNames = names
	}
	info.Location = datastore.GetLoc(ipStr)
	if info.Location != "" && !strings.HasPrefix(info.Location, "LOCAL") {
		rdapCtx, rdapCancel := context.WithTimeout(ctx, 3*time.Second)
		defer rdapCancel()
		type rdapRes struct {
			ip  *rdap.IPNetwork
			err error
		}
		ch := make(chan rdapRes, 1)
		go func() {
			client := &rdap.Client{}
			ri, err := client.QueryIP(ipStr)
			ch <- rdapRes{ip: ri, err: err}
		}()
		select {
		case res := <-ch:
			if res.err == nil && res.ip != nil {
				info.RDAPIPVersion = res.ip.IPVersion
				info.RDAPName = res.ip.Name
				info.RDAPCountry = res.ip.Country
				info.RDAPWhoisServer = res.ip.Port43
				info.RDAPHandle = res.ip.Handle
				info.RDAPType = res.ip.Type
			}
		case <-rdapCtx.Done():
		}
	}
	return mcpJSONResult(info)
}

// get_mac_address_info tool
type getMACInfoParams struct {
	MAC string `json:"mac" jsonschema:"MAC address"`
}

type mcpMACInfoEnt struct {
	MAC    string
	Node   string
	IP     string
	Vendor string
}

func getMACInfo(ctx context.Context, req *mcp.CallToolRequest, args getMACInfoParams) (*mcp.CallToolResult, any, error) {
	mac := normalizeMACAddr(args.MAC)
	info := new(mcpMACInfoEnt)
	info.MAC = mac
	if n := datastore.FindNodeFromMAC(mac); n != nil {
		info.Node = n.Name
		info.IP = n.IP
	} else {
		info.IP = findIPFromArp(mac)
	}
	info.Vendor = datastore.FindVendor(mac)
	return mcpJSONResult(info)
}

func normalizeMACAddr(m string) string {
	if hw, err := net.ParseMAC(strings.TrimSpace(m)); err == nil {
		return strings.ToUpper(hw.String())
	}
	m = strings.ReplaceAll(m, "-", ":")
	a := strings.Split(m, ":")
	r := ""
	for _, e := range a {
		if r != "" {
			r += ":"
		}
		if len(e) == 1 {
			r += "0"
		}
		r += e
	}
	return strings.ToUpper(r)
}

func findIPFromArp(mac string) string {
	ip := ""
	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		if strings.EqualFold(a.MAC, mac) {
			ip = a.IP
			return false
		}
		return true
	})
	return ip
}

func parseTimeExpr(expr string, now time.Time) (time.Time, error) {
	if expr == "now" {
		return now, nil
	}
	if d, err := str2duration.ParseDuration(expr); err == nil {
		return now.Add(d), nil
	}
	t, err := dateparse.ParseLocal(expr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func resolveTimeRange(start, end, defaultStart string) (int64, int64, error) {
	if start == "" {
		start = defaultStart
	}
	return getTimeRange(start, end)
}

// getTimeRange converts start and end time strings (duration like "-1h" or timestamps) to UnixNano bounds.
func getTimeRange(start, end string) (int64, int64, error) {
	if start == "" {
		return 0, 0, fmt.Errorf("start_time must not be empty")
	}
	now := time.Now()
	st, err := parseTimeExpr(start, now)
	if err != nil {
		return 0, 0, err
	}
	et := now
	if end != "" && end != "now" {
		et, err = parseTimeExpr(end, now)
		if err != nil {
			return 0, 0, err
		}
	}
	if st.UnixNano() > et.UnixNano() {
		return 0, 0, fmt.Errorf("start_time must be before end_time")
	}
	return st.UnixNano(), et.UnixNano(), nil
}

var (
	regNumericOID       = regexp.MustCompile(`^[0-9]+(\.[0-9]+)+$`)
	regLeadingDotNumOID = regexp.MustCompile(`^\.[0-9]+(\.[0-9]+)*$`)
)

func nameToOID(name string) string {
	name = strings.TrimSpace(name)
	if regLeadingDotNumOID.MatchString(name) {
		return name
	}
	if regNumericOID.MatchString(name) {
		return "." + name
	}
	oid := datastore.MIBDB.NameToOID(name)
	if oid == ".1" {
		oid = ".1.3"
	}
	if oid == ".0.0" {
		if regLeadingDotNumOID.MatchString(name) {
			return name
		}
		if regNumericOID.MatchString(name) {
			return "." + name
		}
	}
	return oid
}

// get_otel_metrics tool
type mcpOTelMetricEnt struct {
	Host        string  `json:"Host"`
	Service     string  `json:"Service"`
	Scope       string  `json:"Scope"`
	Name        string  `json:"Name"`
	Type        string  `json:"Type"`
	Count       int     `json:"Count"`
	LatestGauge float64 `json:"LatestGauge,omitempty"`
	LatestSum   float64 `json:"LatestSum,omitempty"`
	First       int64   `json:"First"`
	Last        int64   `json:"Last"`
}

type getOTelMetricsParams struct {
	HostFilter    string `json:"host_filter,omitempty" jsonschema:"Host name regex filter. If blank, all hosts are searched."`
	ServiceFilter string `json:"service_filter,omitempty" jsonschema:"Service name regex filter. If blank, all services are searched."`
	NameFilter    string `json:"name_filter,omitempty" jsonschema:"Metric name regex filter. If blank, all metric names are searched."`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of metrics to return (1-500, default 100)."`
}

func getOTelMetrics(ctx context.Context, req *mcp.CallToolRequest, args getOTelMetricsParams) (*mcp.CallToolResult, any, error) {
	hostF, err := compileRegexFilter("host_filter", args.HostFilter)
	if err != nil {
		return nil, nil, err
	}
	svcF, err := compileRegexFilter("service_filter", args.ServiceFilter)
	if err != nil {
		return nil, nil, err
	}
	nameF, err := compileRegexFilter("name_filter", args.NameFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 500, 100)

	list := []mcpOTelMetricEnt{}
	datastore.ForEachOTelMetric(func(m *datastore.OTelMetricEnt) bool {
		if hostF != nil && !hostF.MatchString(m.Host) {
			return true
		}
		if svcF != nil && !svcF.MatchString(m.Service) {
			return true
		}
		if nameF != nil && !nameF.MatchString(m.Name) {
			return true
		}
		ent := mcpOTelMetricEnt{
			Host:    m.Host,
			Service: m.Service,
			Scope:   m.Scope,
			Name:    m.Name,
			Type:    m.Type,
			Count:   m.Count,
			First:   m.First,
			Last:    m.Last,
		}
		if len(m.DataPoints) > 0 {
			dp := m.DataPoints[len(m.DataPoints)-1]
			ent.LatestGauge = dp.Gauge
			ent.LatestSum = dp.Sum
		}
		list = append(list, ent)
		return len(list) < limit
	})
	return mcpJSONResult(&list)
}

// get_otel_traces tool
type mcpOTelTraceEnt struct {
	TraceID    string   `json:"TraceID"`
	Bucket     string   `json:"Bucket"`
	Start      int64    `json:"Start"`
	End        int64    `json:"End"`
	Dur        float64  `json:"Dur"`
	NumSpans   int      `json:"NumSpans"`
	NumErrors  int      `json:"NumErrors"`
	Services   string   `json:"Services"`
	Hosts      string   `json:"Hosts"`
	ErrorSpans []string `json:"ErrorSpans,omitempty"`
}

type getOTelTracesParams struct {
	ServiceFilter string `json:"service_filter,omitempty" jsonschema:"Service name regex filter. If blank, all services are searched."`
	ErrorOnly     bool   `json:"error_only,omitempty" jsonschema:"If true, only return traces containing errors."`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of traces to return (1-200, default 50)."`
}

func isOTelSpanError(span datastore.OTelTraceSpanEnt) bool {
	for _, attr := range span.Attributes {
		lower := strings.ToLower(attr)
		if strings.Contains(lower, "error=true") ||
			strings.Contains(lower, "status_code=status_code_error") ||
			strings.Contains(lower, "status.code=status_code_error") ||
			strings.Contains(lower, "status.code=error") {
			return true
		}
		if strings.HasPrefix(lower, "http.status_code=") || strings.HasPrefix(lower, "http.response.status_code=") {
			parts := strings.SplitN(lower, "=", 2)
			if len(parts) == 2 {
				if code, err := strconv.Atoi(parts[1]); err == nil && code >= 400 {
					return true
				}
			}
		}
	}
	return false
}

func getOTelTraces(ctx context.Context, req *mcp.CallToolRequest, args getOTelTracesParams) (*mcp.CallToolResult, any, error) {
	svcF, err := compileRegexFilter("service_filter", args.ServiceFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 200, 50)

	buckets := datastore.GetOTelTraceBucketList()
	list := []mcpOTelTraceEnt{}

	// Search from most recent buckets
	for i := len(buckets) - 1; i >= 0 && len(list) < limit; i-- {
		datastore.ForEachOTelTrace(buckets[i], func(t *datastore.OTelTraceEnt) bool {
			services := make(map[string]bool)
			hosts := make(map[string]bool)
			errorSpans := []string{}
			hasServiceMatch := svcF == nil

			for _, span := range t.Spans {
				services[span.Service] = true
				hosts[span.Host] = true
				if svcF != nil && svcF.MatchString(span.Service) {
					hasServiceMatch = true
				}
				if isOTelSpanError(span) {
					errorSpans = append(errorSpans, fmt.Sprintf("%s:%s", span.Service, span.Name))
				}
			}

			if !hasServiceMatch {
				return true
			}
			if args.ErrorOnly && len(errorSpans) == 0 {
				return true
			}

			svcList := make([]string, 0, len(services))
			for s := range services {
				svcList = append(svcList, s)
			}
			sort.Strings(svcList)
			hostList := make([]string, 0, len(hosts))
			for h := range hosts {
				hostList = append(hostList, h)
			}
			sort.Strings(hostList)

			list = append(list, mcpOTelTraceEnt{
				TraceID:    t.TraceID,
				Bucket:     t.Bucket,
				Start:      t.Start,
				End:        t.End,
				Dur:        t.Dur,
				NumSpans:   len(t.Spans),
				NumErrors:  len(errorSpans),
				Services:   strings.Join(svcList, ", "),
				Hosts:      strings.Join(hostList, ", "),
				ErrorSpans: errorSpans,
			})
			return len(list) < limit
		})
	}

	return mcpJSONResult(&list)
}

// get_sensor_list tool
type getSensorListParams struct {
	TypeFilter     string `json:"type_filter,omitempty" jsonschema:"Sensor type regex filter (e.g. twWifiScan, twBlueScan, twpcap, twwinlog, twagentneo). If blank, all sensors are returned."`
	StateFilter    string `json:"state_filter,omitempty" jsonschema:"Sensor state regex filter (normal, warn, error). If blank, all sensors are returned."`
	IncludeHistory bool   `json:"include_history,omitempty" jsonschema:"Include full Stats and Monitors time-series arrays (default false to return compact summaries)."`
	Limit          int    `json:"limit,omitempty" jsonschema:"Maximum number of sensors to return (1-500, default 100)."`
}

func getSensorList(ctx context.Context, req *mcp.CallToolRequest, args getSensorListParams) (*mcp.CallToolResult, any, error) {
	typeF, err := compileRegexFilter("type_filter", args.TypeFilter)
	if err != nil {
		return nil, nil, err
	}
	stateF, err := compileRegexFilter("state_filter", args.StateFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 500, 100)

	all := datastore.ListSensors()
	filtered := []*datastore.SensorEnt{}
	for _, sn := range all {
		if typeF != nil && !typeF.MatchString(sn.Type) {
			continue
		}
		if stateF != nil && !stateF.MatchString(sn.State) {
			continue
		}
		if !args.IncludeHistory {
			sn.Stats = nil
			sn.Monitors = nil
		}
		filtered = append(filtered, sn)
		if len(filtered) >= limit {
			break
		}
	}
	return mcpJSONResult(&filtered)
}

// get_report_summary tool
type getReportSummaryParams struct {
	ReportType     string `json:"report_type" jsonschema:"Report kind name (e.g. wifiAP, blueDevice, envMonitor, powerMonitor, motionSensor, etherType, dnsq, radiusFlow, tlsFlow, winEventID, winLogon, winAccount, winKerberos, winPrivilege, winProcess, winTask, flow, server, fumble, syslogStats, trapStats, twAgentInfo, twAgentAlert, twAgentFIM, twAgentSCA, twAgentTCP)."`
	Filter         string `json:"filter,omitempty" jsonschema:"Optional regex filter applied to JSON representation of each report record."`
	IncludeHistory bool   `json:"include_history,omitempty" jsonschema:"Include full historical time-series arrays such as RSSI, EnvData, Data (default false)."`
	Limit          int    `json:"limit,omitempty" jsonschema:"Maximum number of records to return (1-200, default 50)."`
}

func compactReportObject(obj map[string]any) {
	for _, key := range []string{"RSSI", "EnvData", "Data"} {
		if rawSlice, ok := obj[key].([]any); ok {
			obj[key+"Count"] = len(rawSlice)
			if len(rawSlice) > 0 {
				obj["Latest"+key] = rawSlice[len(rawSlice)-1]
			}
			delete(obj, key)
		}
	}
}

func getReportSummary(ctx context.Context, req *mcp.CallToolRequest, args getReportSummaryParams) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(args.ReportType) == "" {
		return mcpTextResult("report_type is required (e.g. wifiAP, blueDevice, envMonitor, dnsq, winEventID, twAgentAlert, twAgentInfo)"), nil, nil
	}
	reportType := logreport.NormalizeKind(strings.TrimSpace(args.ReportType))
	if !logreport.IsKind(reportType) {
		return nil, nil, fmt.Errorf("unknown report_type %q; valid kinds: %s", args.ReportType, strings.Join(logreport.AllKinds(), ", "))
	}
	filterRe, err := compileRegexFilter("filter", args.Filter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 200, 50)

	logger.FlushReportEngine()
	rawItems, err := datastore.ListLogReportData(reportType)
	if err != nil {
		return mcpTextResult(fmt.Sprintf("failed to get report data: %v", err)), nil, nil
	}

	keys := make([]string, 0, len(rawItems))
	for k := range rawItems {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	list := make([]any, 0, limit)
	for _, k := range keys {
		b := rawItems[k]
		if filterRe != nil && !filterRe.Match(b) {
			continue
		}
		var obj any
		if err := json.Unmarshal(b, &obj); err == nil {
			if !args.IncludeHistory {
				if m, ok := obj.(map[string]any); ok {
					compactReportObject(m)
				}
			}
			list = append(list, obj)
			if len(list) >= limit {
				break
			}
		}
	}
	return mcpJSONResult(&list)
}
