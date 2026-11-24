package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openconfig/gnmic/pkg/api"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/logreport"
	"github.com/twsnmp/twsnmpfk/ping"
)

const maxAnalyticsLogScan = 50000

func addAnalyticsTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_netflow",
		Description: "search NetFlow records from TWSNMP",
	}, searchNetFlowTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_sflow",
		Description: "search sFlow flow sample records from TWSNMP",
	}, searchSFlowTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_sflow_counter",
		Description: "search sFlow interface counter sample records from TWSNMP",
	}, searchSFlowCounterTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_otel_trace_detail",
		Description: "get detailed spans of an OpenTelemetry trace by trace_id from TWSNMP",
	}, getOTelTraceDetailTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ai_anomaly_list",
		Description: "get AI anomaly detection scores and alert levels across pollings from TWSNMP",
	}, getAIAnomalyListTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "diagnose_node",
		Description: "run multi-layer live diagnostics (Ping, SNMP sysInfo, TCP ports, HTTP title/server) against a node or target",
	}, diagnoseNodeTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_mib",
		Description: "search loaded MIB objects by name, OID, or description in TWSNMP",
	}, searchMIBTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mac_address_list",
		Description: "get aggregated MAC address table (IPs, nodes, vendor, timestamps) from TWSNMP",
	}, getMACAddressListTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_wifi_ap_list",
		Description: "get detected Wi-Fi access points from TWSNMP report store",
	}, getWifiAPListTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_bluetooth_device_list",
		Description: "get detected Bluetooth devices from TWSNMP report store",
	}, getBluetoothDeviceListTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_fdb_table",
		Description: "get live SNMP switch Forwarding Database (FDB / MAC-to-port) table for a managed network or switch",
	}, getFDBTableTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "auto_connect_lines",
		Description: "automatically discover and add map lines from FDB/ARP topology data in TWSNMP",
	}, autoConnectLinesTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "auto_layout",
		Description: "automatically arrange node positions on the TWSNMP map (hierarchical, cluster, categorized) or undo",
	}, autoLayoutTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "gnmi_capabilities",
		Description: "query gNMI supported models, version, and encodings from a managed node",
	}, gnmiCapabilitiesTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "gnmi_get",
		Description: "query operational/config telemetry data via gNMI Get from a managed node",
	}, gnmiGetTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "analyze_polling_correlation",
		Description: "analyze correlation between two polling log time-series (same or different nodes) using multiple algorithms",
	}, analyzePollingCorrelationTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "discover_polling_correlations",
		Description: "automatically discover and rank correlated polling pairs across nodes",
	}, discoverPollingCorrelationsTool)
}

// search_netflow tool
type mcpNetFlowEnt struct {
	Time     string  `json:"Time"`
	SrcAddr  string  `json:"SrcAddr"`
	SrcPort  int     `json:"SrcPort"`
	SrcLoc   string  `json:"SrcLoc,omitempty"`
	SrcMAC   string  `json:"SrcMAC,omitempty"`
	DstAddr  string  `json:"DstAddr"`
	DstPort  int     `json:"DstPort"`
	DstLoc   string  `json:"DstLoc,omitempty"`
	DstMAC   string  `json:"DstMAC,omitempty"`
	Bytes    int     `json:"Bytes"`
	Packets  int     `json:"Packets"`
	TCPFlags string  `json:"TCPFlags,omitempty"`
	Protocol string  `json:"Protocol"`
	ToS      int     `json:"ToS,omitempty"`
	Start    string  `json:"Start,omitempty"`
	End      string  `json:"End,omitempty"`
	Dur      float64 `json:"Dur"`
}

type searchNetFlowParams struct {
	StartTime      string `json:"start_time,omitempty" jsonschema:"Start time (duration like '-1h' or timestamp, default '-1h')"`
	EndTime        string `json:"end_time,omitempty" jsonschema:"End time (duration, timestamp, or 'now', default 'now')"`
	SrcFilter      string `json:"src_filter,omitempty" jsonschema:"Regex filter on source IP address"`
	DstFilter      string `json:"dst_filter,omitempty" jsonschema:"Regex filter on destination IP address"`
	ProtocolFilter string `json:"protocol_filter,omitempty" jsonschema:"Regex filter on protocol name"`
	SrcPort        int    `json:"src_port,omitempty" jsonschema:"Exact source port filter (>0)"`
	DstPort        int    `json:"dst_port,omitempty" jsonschema:"Exact destination port filter (>0)"`
	Limit          int    `json:"limit,omitempty" jsonschema:"Maximum number of records to return (1-2000, default 200)"`
}

func searchNetFlowTool(ctx context.Context, req *mcp.CallToolRequest, args searchNetFlowParams) (*mcp.CallToolResult, any, error) {
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	srcRe, err := compileRegexFilter("src_filter", args.SrcFilter)
	if err != nil {
		return nil, nil, err
	}
	dstRe, err := compileRegexFilter("dst_filter", args.DstFilter)
	if err != nil {
		return nil, nil, err
	}
	protoRe, err := compileRegexFilter("protocol_filter", args.ProtocolFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 2000, 200)
	list := []mcpNetFlowEnt{}
	scanned := 0
	if err := datastore.ForEachNetFlow(st, et, func(e *datastore.NetFlowEnt) bool {
		if ctx.Err() != nil || scanned >= maxAnalyticsLogScan {
			return false
		}
		scanned++
		if srcRe != nil && !srcRe.MatchString(e.SrcAddr) {
			return true
		}
		if dstRe != nil && !dstRe.MatchString(e.DstAddr) {
			return true
		}
		if protoRe != nil && !protoRe.MatchString(e.Protocol) {
			return true
		}
		if args.SrcPort > 0 && e.SrcPort != args.SrcPort {
			return true
		}
		if args.DstPort > 0 && e.DstPort != args.DstPort {
			return true
		}
		list = append(list, mcpNetFlowEnt{
			Time:     formatTimeUnixNano(e.Time),
			SrcAddr:  e.SrcAddr,
			SrcPort:  e.SrcPort,
			SrcLoc:   e.SrcLoc,
			SrcMAC:   e.SrcMAC,
			DstAddr:  e.DstAddr,
			DstPort:  e.DstPort,
			DstLoc:   e.DstLoc,
			DstMAC:   e.DstMAC,
			Bytes:    e.Bytes,
			Packets:  e.Packets,
			TCPFlags: e.TCPFlags,
			Protocol: e.Protocol,
			ToS:      e.ToS,
			Start:    formatTimeUnixNano(e.Start),
			End:      formatTimeUnixNano(e.End),
			Dur:      e.Dur,
		})
		return len(list) < limit
	}); err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(&list)
}

// search_sflow tool
type mcpSFlowEnt struct {
	Time     string `json:"Time"`
	SrcAddr  string `json:"SrcAddr"`
	SrcPort  int    `json:"SrcPort"`
	SrcLoc   string `json:"SrcLoc,omitempty"`
	SrcMAC   string `json:"SrcMAC,omitempty"`
	DstAddr  string `json:"DstAddr"`
	DstPort  int    `json:"DstPort"`
	DstLoc   string `json:"DstLoc,omitempty"`
	DstMAC   string `json:"DstMAC,omitempty"`
	Bytes    int    `json:"Bytes"`
	TCPFlags string `json:"TCPFlags,omitempty"`
	Protocol string `json:"Protocol"`
	Reason   int    `json:"Reason,omitempty"`
}

type searchSFlowParams struct {
	StartTime      string `json:"start_time,omitempty" jsonschema:"Start time (duration like '-1h' or timestamp, default '-1h')"`
	EndTime        string `json:"end_time,omitempty" jsonschema:"End time (duration, timestamp, or 'now', default 'now')"`
	SrcFilter      string `json:"src_filter,omitempty" jsonschema:"Regex filter on source IP address"`
	DstFilter      string `json:"dst_filter,omitempty" jsonschema:"Regex filter on destination IP address"`
	ProtocolFilter string `json:"protocol_filter,omitempty" jsonschema:"Regex filter on protocol name"`
	SrcPort        int    `json:"src_port,omitempty" jsonschema:"Exact source port filter (>0)"`
	DstPort        int    `json:"dst_port,omitempty" jsonschema:"Exact destination port filter (>0)"`
	Limit          int    `json:"limit,omitempty" jsonschema:"Maximum number of records to return (1-2000, default 200)"`
}

func searchSFlowTool(ctx context.Context, req *mcp.CallToolRequest, args searchSFlowParams) (*mcp.CallToolResult, any, error) {
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	srcRe, err := compileRegexFilter("src_filter", args.SrcFilter)
	if err != nil {
		return nil, nil, err
	}
	dstRe, err := compileRegexFilter("dst_filter", args.DstFilter)
	if err != nil {
		return nil, nil, err
	}
	protoRe, err := compileRegexFilter("protocol_filter", args.ProtocolFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 2000, 200)
	list := []mcpSFlowEnt{}
	scanned := 0
	if err := datastore.ForEachSFlow(st, et, func(e *datastore.SFlowEnt) bool {
		if ctx.Err() != nil || scanned >= maxAnalyticsLogScan {
			return false
		}
		scanned++
		if srcRe != nil && !srcRe.MatchString(e.SrcAddr) {
			return true
		}
		if dstRe != nil && !dstRe.MatchString(e.DstAddr) {
			return true
		}
		if protoRe != nil && !protoRe.MatchString(e.Protocol) {
			return true
		}
		if args.SrcPort > 0 && e.SrcPort != args.SrcPort {
			return true
		}
		if args.DstPort > 0 && e.DstPort != args.DstPort {
			return true
		}
		list = append(list, mcpSFlowEnt{
			Time:     formatTimeUnixNano(e.Time),
			SrcAddr:  e.SrcAddr,
			SrcPort:  e.SrcPort,
			SrcLoc:   e.SrcLoc,
			SrcMAC:   e.SrcMAC,
			DstAddr:  e.DstAddr,
			DstPort:  e.DstPort,
			DstLoc:   e.DstLoc,
			DstMAC:   e.DstMAC,
			Bytes:    e.Bytes,
			TCPFlags: e.TCPFlags,
			Protocol: e.Protocol,
			Reason:   e.Reason,
		})
		return len(list) < limit
	}); err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(&list)
}

// search_sflow_counter tool
type mcpSFlowCounterEnt struct {
	Time   string `json:"Time"`
	Remote string `json:"Remote"`
	Type   string `json:"Type"`
	Data   string `json:"Data"`
}

type searchSFlowCounterParams struct {
	StartTime    string `json:"start_time,omitempty" jsonschema:"Start time (duration like '-1h' or timestamp, default '-1h')"`
	EndTime      string `json:"end_time,omitempty" jsonschema:"End time (duration, timestamp, or 'now', default 'now')"`
	RemoteFilter string `json:"remote_filter,omitempty" jsonschema:"Regex filter on exporter/remote IP"`
	TypeFilter   string `json:"type_filter,omitempty" jsonschema:"Regex filter on counter type"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum number of records to return (1-2000, default 200)"`
}

func searchSFlowCounterTool(ctx context.Context, req *mcp.CallToolRequest, args searchSFlowCounterParams) (*mcp.CallToolResult, any, error) {
	st, et, err := resolveTimeRange(args.StartTime, args.EndTime, "-1h")
	if err != nil {
		return nil, nil, err
	}
	remoteRe, err := compileRegexFilter("remote_filter", args.RemoteFilter)
	if err != nil {
		return nil, nil, err
	}
	typeRe, err := compileRegexFilter("type_filter", args.TypeFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 2000, 200)
	list := []mcpSFlowCounterEnt{}
	scanned := 0
	if err := datastore.ForEachSFlowCounter(st, et, func(e *datastore.SFlowCounterEnt) bool {
		if ctx.Err() != nil || scanned >= maxAnalyticsLogScan {
			return false
		}
		scanned++
		if remoteRe != nil && !remoteRe.MatchString(e.Remote) {
			return true
		}
		if typeRe != nil && !typeRe.MatchString(e.Type) {
			return true
		}
		list = append(list, mcpSFlowCounterEnt{
			Time:   formatTimeUnixNano(e.Time),
			Remote: e.Remote,
			Type:   e.Type,
			Data:   e.Data,
		})
		return len(list) < limit
	}); err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(&list)
}

// get_otel_trace_detail tool
type mcpOTelSpanDetail struct {
	SpanID       string   `json:"SpanID"`
	ParentSpanID string   `json:"ParentSpanID,omitempty"`
	Host         string   `json:"Host"`
	Service      string   `json:"Service"`
	Scope        string   `json:"Scope,omitempty"`
	Name         string   `json:"Name"`
	Start        string   `json:"Start"`
	End          string   `json:"End"`
	Dur          float64  `json:"Dur"`
	Error        bool     `json:"Error"`
	Attributes   []string `json:"Attributes,omitempty"`
}

type mcpOTelTraceDetail struct {
	TraceID   string              `json:"TraceID"`
	Bucket    string              `json:"Bucket"`
	Start     string              `json:"Start"`
	End       string              `json:"End"`
	Last      string              `json:"Last"`
	Dur       float64             `json:"Dur"`
	NumSpans  int                 `json:"NumSpans"`
	NumErrors int                 `json:"NumErrors"`
	Spans     []mcpOTelSpanDetail `json:"Spans"`
}

type getOTelTraceDetailParams struct {
	TraceID string `json:"trace_id" jsonschema:"OpenTelemetry TraceID to look up"`
	Bucket  string `json:"bucket,omitempty" jsonschema:"Optional OTel bucket timestamp (searches all recent buckets if omitted)"`
}

func getOTelTraceDetailTool(ctx context.Context, req *mcp.CallToolRequest, args getOTelTraceDetailParams) (*mcp.CallToolResult, any, error) {
	tid := strings.TrimSpace(args.TraceID)
	if tid == "" {
		return nil, nil, fmt.Errorf("trace_id must not be empty")
	}
	var trace *datastore.OTelTraceEnt
	bucket := strings.TrimSpace(args.Bucket)
	if bucket != "" {
		trace = datastore.GetOTelTrace(bucket, tid)
	} else {
		buckets := datastore.GetOTelTraceBucketList()
		for i := len(buckets) - 1; i >= 0; i-- {
			if ctx.Err() != nil {
				break
			}
			if tr := datastore.GetOTelTrace(buckets[i], tid); tr != nil {
				trace = tr
				bucket = buckets[i]
				break
			}
		}
	}
	if trace == nil {
		return nil, nil, fmt.Errorf("otel trace not found: %s", tid)
	}
	spans := make([]mcpOTelSpanDetail, 0, len(trace.Spans))
	errCount := 0
	for _, sp := range trace.Spans {
		isErr := isOTelSpanError(sp)
		if isErr {
			errCount++
		}
		spans = append(spans, mcpOTelSpanDetail{
			SpanID:       sp.SpanID,
			ParentSpanID: sp.ParentSpanID,
			Host:         sp.Host,
			Service:      sp.Service,
			Scope:        sp.Scope,
			Name:         sp.Name,
			Start:        formatTimeUnixNano(sp.Start),
			End:          formatTimeUnixNano(sp.End),
			Dur:          sp.Dur,
			Error:        isErr,
			Attributes:   sp.Attributes,
		})
	}
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].Start < spans[j].Start
	})
	if trace.Bucket != "" {
		bucket = trace.Bucket
	}
	return mcpJSONResult(&mcpOTelTraceDetail{
		TraceID:   trace.TraceID,
		Bucket:    bucket,
		Start:     formatTimeUnixNano(trace.Start),
		End:       formatTimeUnixNano(trace.End),
		Last:      formatTimeUnixNano(trace.Last),
		Dur:       trace.Dur,
		NumSpans:  len(spans),
		NumErrors: errCount,
		Spans:     spans,
	})
}

// get_ai_anomaly_list tool
type mcpAIAnomalyEnt struct {
	PollingID   string      `json:"PollingID"`
	PollingName string      `json:"PollingName"`
	PollingType string      `json:"PollingType"`
	NodeID      string      `json:"NodeID"`
	NodeName    string      `json:"NodeName"`
	LatestScore float64     `json:"LatestScore"`
	Level       string      `json:"Level"`
	Count       int         `json:"Count"`
	LastTime    string      `json:"LastTime"`
	ScoreData   [][]float64 `json:"ScoreData,omitempty"`
}

type mcpAIAnomalyListResp struct {
	Thresholds datastore.AIConfEnt `json:"Thresholds"`
	Items      []mcpAIAnomalyEnt   `json:"Items"`
}

type getAIAnomalyListParams struct {
	NodeID         string  `json:"node_id,omitempty" jsonschema:"Optional node ID, unique name, or IP to filter"`
	MinScore       float64 `json:"min_score,omitempty" jsonschema:"Minimum latest anomaly score threshold"`
	IncludeHistory bool    `json:"include_history,omitempty" jsonschema:"Include full score time-series array (default false)"`
	Limit          int     `json:"limit,omitempty" jsonschema:"Maximum number of items to return (1-1000, default 100)"`
}

func classifyAILevel(score float64) string {
	if datastore.AIConf.HighThreshold > 0 && score > datastore.AIConf.HighThreshold {
		return "high"
	}
	if datastore.AIConf.LowThreshold > 0 && score > datastore.AIConf.LowThreshold {
		return "low"
	}
	if datastore.AIConf.WarnThreshold > 0 && score > datastore.AIConf.WarnThreshold {
		return "warn"
	}
	return "normal"
}

func getAIAnomalyListTool(ctx context.Context, req *mcp.CallToolRequest, args getAIAnomalyListParams) (*mcp.CallToolResult, any, error) {
	filterNodeID := ""
	if strings.TrimSpace(args.NodeID) != "" {
		n, err := resolveNode(args.NodeID, true, true)
		if err != nil {
			return nil, nil, err
		}
		filterNodeID = n.ID
	}
	limit := clampLimit(args.Limit, 1, 1000, 100)
	items := []mcpAIAnomalyEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		if filterNodeID != "" && p.NodeID != filterNodeID {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) == 0 {
			return true
		}
		latest := air.ScoreData[len(air.ScoreData)-1][1]
		if args.MinScore > 0 && latest < args.MinScore {
			return true
		}
		ent := mcpAIAnomalyEnt{
			PollingID:   p.ID,
			PollingName: p.Name,
			PollingType: p.Type,
			NodeID:      n.ID,
			NodeName:    n.Name,
			LatestScore: latest,
			Level:       classifyAILevel(latest),
			Count:       len(air.ScoreData),
			LastTime:    formatTimeUnix(air.LastTime),
		}
		if args.IncludeHistory {
			ent.ScoreData = air.ScoreData
		}
		items = append(items, ent)
		return true
	})
	sort.Slice(items, func(i, j int) bool {
		if items[i].LatestScore == items[j].LatestScore {
			return items[i].PollingID < items[j].PollingID
		}
		return items[i].LatestScore > items[j].LatestScore
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return mcpJSONResult(&mcpAIAnomalyListResp{
		Thresholds: datastore.AIConf,
		Items:      items,
	})
}

// diagnose_node tool
type diagnoseNodeParams struct {
	Target  string `json:"target" jsonschema:"Node ID, node name, IP address, or hostname to diagnose"`
	Ports   string `json:"ports,omitempty" jsonschema:"Comma-separated TCP ports to check (default 22,80,443,8080,8443)"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"Probe timeout in seconds (1-10, default 3)"`
}

type diagnoseTCPPortResult struct {
	Port   int    `json:"Port"`
	Open   bool   `json:"Open"`
	RTT    string `json:"RTT,omitempty"`
	Error  string `json:"Error,omitempty"`
	Banner string `json:"Banner,omitempty"`
}

type diagnoseSNMPResult struct {
	Reachable   bool   `json:"Reachable"`
	SysDescr    string `json:"SysDescr,omitempty"`
	SysObjectID string `json:"SysObjectID,omitempty"`
	SysUpTime   string `json:"SysUpTime,omitempty"`
	SysName     string `json:"SysName,omitempty"`
	SysLocation string `json:"SysLocation,omitempty"`
	Error       string `json:"Error,omitempty"`
}

type diagnoseHTTPResult struct {
	Checked bool   `json:"Checked"`
	URL     string `json:"URL,omitempty"`
	Title   string `json:"Title,omitempty"`
	Server  string `json:"Server,omitempty"`
}

type diagnoseNodeResult struct {
	Target          string                  `json:"Target"`
	ResolvedIP      string                  `json:"ResolvedIP"`
	Node            *mcpNodeEnt             `json:"Node,omitempty"`
	FailingPollings []mcpPollingEnt         `json:"FailingPollings,omitempty"`
	Ping            mcpPingEnt              `json:"Ping"`
	SNMP            diagnoseSNMPResult      `json:"SNMP"`
	TCPPorts        []diagnoseTCPPortResult `json:"TCPPorts"`
	HTTP            diagnoseHTTPResult      `json:"HTTP"`
}

func parseDiagPorts(spec string) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return []int{22, 80, 443, 8080, 8443}, nil
	}
	seen := map[int]bool{}
	ports := []int{}
	for _, part := range strings.Split(spec, ",") {
		pStr := strings.TrimSpace(part)
		if pStr == "" {
			continue
		}
		p, err := strconv.Atoi(pStr)
		if err != nil || p < 1 || p > 65535 {
			return nil, fmt.Errorf("invalid TCP port %q in ports", pStr)
		}
		if !seen[p] {
			seen[p] = true
			ports = append(ports, p)
		}
	}
	if len(ports) > 32 {
		return nil, fmt.Errorf("too many ports specified (%d > 32)", len(ports))
	}
	return ports, nil
}

func formatSNMPDiagValue(pdu gosnmp.SnmpPDU) string {
	switch pdu.Type {
	case gosnmp.OctetString:
		if b, ok := pdu.Value.([]byte); ok {
			return strings.TrimSpace(string(b))
		}
	case gosnmp.ObjectIdentifier:
		oid := fmt.Sprintf("%v", pdu.Value)
		if datastore.MIBDB != nil {
			if name := datastore.MIBDB.OIDToName(oid); name != "" && name != oid {
				return fmt.Sprintf("%s (%s)", name, oid)
			}
		}
		return oid
	case gosnmp.TimeTicks:
		ticks := gosnmp.ToBigInt(pdu.Value).Int64()
		return (time.Duration(ticks*10) * time.Millisecond).String()
	}
	return fmt.Sprintf("%v", pdu.Value)
}

func diagnoseNodeTool(ctx context.Context, req *mcp.CallToolRequest, args diagnoseNodeParams) (*mcp.CallToolResult, any, error) {
	ports, err := parseDiagPorts(args.Ports)
	if err != nil {
		return nil, nil, err
	}
	timeout := clampLimit(args.Timeout, 1, 10, 3)

	var nodeSummary *mcpNodeEnt
	var resolvedNode *datastore.NodeEnt
	if n, err := resolveNode(args.Target, true, false); err == nil && n != nil {
		resolvedNode = n
		nodeSummary = &mcpNodeEnt{
			ID:          n.ID,
			Name:        n.Name,
			IP:          n.IP,
			MAC:         n.MAC,
			State:       n.State,
			X:           n.X,
			Y:           n.Y,
			Icon:        n.Icon,
			Description: n.Descr,
		}
	}
	ip := ""
	if resolvedNode != nil && resolvedNode.IP != "" {
		ip = resolvedNode.IP
	} else {
		ip = getTargetIP(args.Target)
	}
	if ip == "" {
		return nil, nil, fmt.Errorf("target not found: %s", args.Target)
	}

	res := diagnoseNodeResult{
		Target:     args.Target,
		ResolvedIP: ip,
		Node:       nodeSummary,
		TCPPorts:   make([]diagnoseTCPPortResult, 0, len(ports)),
	}

	if resolvedNode != nil {
		failing := []mcpPollingEnt{}
		datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
			if p.NodeID == resolvedNode.ID && p.State != "normal" {
				failing = append(failing, mcpPollingEnt{
					ID:       p.ID,
					Name:     p.Name,
					NodeID:   p.NodeID,
					NodeName: resolvedNode.Name,
					Type:     p.Type,
					Level:    p.Level,
					State:    p.State,
					Logging:  p.LogMode > 0,
					LastTime: formatTimeUnixNano(p.LastTime),
					Result:   p.Result,
				})
			}
			return true
		})
		res.FailingPollings = failing
	}

	// 1. Ping check
	pe := ping.DoPing(ip, timeout, 0, 64, 254)
	res.Ping = mcpPingEnt{
		Result:       pe.Stat.String(),
		Time:         time.Now().Format(time.RFC3339),
		RTT:          time.Duration(pe.Time).String(),
		RTTNano:      pe.Time,
		Size:         pe.Size,
		TTL:          pe.RecvTTL,
		ResponseFrom: pe.RecvSrc,
	}
	if pe.RecvSrc != "" {
		res.Ping.Location = datastore.GetLoc(pe.RecvSrc)
	}

	// 2. SNMP sysInfo check
	agent, _, err := buildSNMPAgentFromParams(args.Target, "", "", "", "", 0)
	if err != nil {
		res.SNMP.Error = err.Error()
	} else {
		agent.Timeout = time.Duration(timeout) * time.Second
		agent.Retries = 0
		if err := agent.Connect(); err != nil {
			res.SNMP.Error = err.Error()
		} else {
			defer agent.Conn.Close()
			oids := []string{
				"1.3.6.1.2.1.1.1.0", // sysDescr
				"1.3.6.1.2.1.1.2.0", // sysObjectID
				"1.3.6.1.2.1.1.3.0", // sysUpTime
				"1.3.6.1.2.1.1.5.0", // sysName
				"1.3.6.1.2.1.1.6.0", // sysLocation
			}
			pkt, err := agent.Get(oids)
			if err != nil {
				res.SNMP.Error = err.Error()
			} else if pkt != nil {
				res.SNMP.Reachable = true
				for _, v := range pkt.Variables {
					if v.Type == gosnmp.NoSuchObject || v.Type == gosnmp.NoSuchInstance {
						continue
					}
					val := formatSNMPDiagValue(v)
					switch strings.TrimPrefix(v.Name, ".") {
					case "1.3.6.1.2.1.1.1.0":
						res.SNMP.SysDescr = val
					case "1.3.6.1.2.1.1.2.0":
						res.SNMP.SysObjectID = val
					case "1.3.6.1.2.1.1.3.0":
						res.SNMP.SysUpTime = val
					case "1.3.6.1.2.1.1.5.0":
						res.SNMP.SysName = val
					case "1.3.6.1.2.1.1.6.0":
						res.SNMP.SysLocation = val
					}
				}
			}
		}
	}

	// 3. TCP port probes
	dialer := &net.Dialer{Timeout: time.Duration(timeout) * time.Second}
	openWebPort := 0
	for _, port := range ports {
		if ctx.Err() != nil {
			break
		}
		start := time.Now()
		conn, derr := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
		pr := diagnoseTCPPortResult{Port: port}
		if derr != nil {
			pr.Error = derr.Error()
		} else {
			pr.Open = true
			pr.RTT = time.Since(start).Round(time.Microsecond).String()
			_ = conn.Close()
			if openWebPort == 0 && (port == 80 || port == 443 || port == 8080 || port == 8443) {
				openWebPort = port
			}
		}
		res.TCPPorts = append(res.TCPPorts, pr)
	}

	// 4. Web signature check if HTTP/HTTPS port is open or node URL configured
	webURL := ""
	if resolvedNode != nil && resolvedNode.URL != "" {
		webURL = resolvedNode.URL
	} else if openWebPort != 0 {
		scheme := "http"
		if openWebPort == 443 || openWebPort == 8443 {
			scheme = "https"
		}
		webURL = fmt.Sprintf("%s://%s:%d", scheme, ip, openWebPort)
	}
	if webURL != "" && ctx.Err() == nil {
		title, server, _ := FetchWebSignatures(ip, webURL)
		res.HTTP = diagnoseHTTPResult{
			Checked: true,
			URL:     webURL,
			Title:   title,
			Server:  server,
		}
	}

	return mcpJSONResult(&res)
}

// search_mib tool
type searchMIBParams struct {
	Query string `json:"query" jsonschema:"Case-insensitive regex or substring matching MIB name, OID, or description"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of MIB objects to return (1-500, default 50)"`
}

type mcpMIBSearchEnt struct {
	OID         string `json:"OID"`
	Name        string `json:"Name"`
	Type        string `json:"Type,omitempty"`
	Status      string `json:"Status,omitempty"`
	Units       string `json:"Units,omitempty"`
	Enum        string `json:"Enum,omitempty"`
	Hint        string `json:"Hint,omitempty"`
	Description string `json:"Description,omitempty"`
}

func searchMIBTool(ctx context.Context, req *mcp.CallToolRequest, args searchMIBParams) (*mcp.CallToolResult, any, error) {
	q := strings.TrimSpace(args.Query)
	if q == "" {
		return nil, nil, fmt.Errorf("query must not be empty")
	}
	pattern := q
	if !strings.HasPrefix(pattern, "(?") {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid query regular expression %q: %w", q, err)
	}
	limit := clampLimit(args.Limit, 1, 500, 50)
	matches := []mcpMIBSearchEnt{}
	for oid, info := range datastore.MIBInfoMap {
		if ctx.Err() != nil {
			break
		}
		if info == nil {
			continue
		}
		name := oid
		if datastore.MIBDB != nil {
			if resolved := datastore.MIBDB.OIDToName(oid); resolved != "" {
				name = resolved
			}
		}
		if re.MatchString(oid) || re.MatchString(name) || re.MatchString(info.Description) {
			matches = append(matches, mcpMIBSearchEnt{
				OID:         oid,
				Name:        name,
				Type:        info.Type,
				Status:      info.Status,
				Units:       info.Units,
				Enum:        info.Enum,
				Hint:        info.Hint,
				Description: info.Description,
			})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].OID < matches[j].OID
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return mcpJSONResult(&matches)
}

// get_mac_address_list tool
type mcpMACListEnt struct {
	MAC       string   `json:"MAC"`
	IPs       []string `json:"IPs"`
	Nodes     []string `json:"Nodes"`
	Vendor    string   `json:"Vendor"`
	FirstTime string   `json:"FirstTime,omitempty"`
	LastTime  string   `json:"LastTime,omitempty"`
}

type getMACAddressListParams struct {
	MACFilter    string `json:"mac_filter,omitempty" jsonschema:"Regex filter on MAC address"`
	IPFilter     string `json:"ip_filter,omitempty" jsonschema:"Regex filter matching any associated IP address"`
	VendorFilter string `json:"vendor_filter,omitempty" jsonschema:"Regex filter on vendor name"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum number of entries to return (1-2000, default 500)"`
}

func getMACAddressListTool(ctx context.Context, req *mcp.CallToolRequest, args getMACAddressListParams) (*mcp.CallToolResult, any, error) {
	macRe, err := compileRegexFilter("mac_filter", args.MACFilter)
	if err != nil {
		return nil, nil, err
	}
	ipRe, err := compileRegexFilter("ip_filter", args.IPFilter)
	if err != nil {
		return nil, nil, err
	}
	vendorRe, err := compileRegexFilter("vendor_filter", args.VendorFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 2000, 500)

	type macAgg struct {
		ips       map[string]bool
		nodes     map[string]bool
		vendor    string
		firstTime int64
		lastTime  int64
	}
	byMAC := map[string]*macAgg{}
	getOrCreate := func(mac string) *macAgg {
		norm := normalizeMACAddr(mac)
		agg, ok := byMAC[norm]
		if !ok {
			agg = &macAgg{
				ips:   map[string]bool{},
				nodes: map[string]bool{},
			}
			byMAC[norm] = agg
		}
		return agg
	}

	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		if strings.TrimSpace(a.MAC) == "" {
			return true
		}
		agg := getOrCreate(a.MAC)
		if a.IP != "" {
			agg.ips[a.IP] = true
		}
		if a.Vendor != "" && agg.vendor == "" {
			agg.vendor = a.Vendor
		}
		if a.NodeID != "" {
			if n := datastore.GetNode(a.NodeID); n != nil {
				agg.nodes[n.Name] = true
			}
		}
		if a.FirstTime > 0 && (agg.firstTime == 0 || a.FirstTime < agg.firstTime) {
			agg.firstTime = a.FirstTime
		}
		if a.LastTime > agg.lastTime {
			agg.lastTime = a.LastTime
		}
		return true
	})

	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if ctx.Err() != nil {
			return false
		}
		if strings.TrimSpace(n.MAC) == "" {
			return true
		}
		agg := getOrCreate(n.MAC)
		if n.IP != "" {
			agg.ips[n.IP] = true
		}
		if n.Name != "" {
			agg.nodes[n.Name] = true
		}
		if n.Vendor != "" && agg.vendor == "" {
			agg.vendor = n.Vendor
		}
		return true
	})

	list := []mcpMACListEnt{}
	for mac, agg := range byMAC {
		if macRe != nil && !macRe.MatchString(mac) {
			continue
		}
		vendor := agg.vendor
		if vendor == "" {
			vendor = datastore.FindVendor(mac)
		}
		if vendorRe != nil && !vendorRe.MatchString(vendor) {
			continue
		}
		ips := make([]string, 0, len(agg.ips))
		ipMatched := ipRe == nil
		for ip := range agg.ips {
			ips = append(ips, ip)
			if ipRe != nil && ipRe.MatchString(ip) {
				ipMatched = true
			}
		}
		if !ipMatched {
			continue
		}
		sort.Strings(ips)
		nodes := make([]string, 0, len(agg.nodes))
		for n := range agg.nodes {
			nodes = append(nodes, n)
		}
		sort.Strings(nodes)
		list = append(list, mcpMACListEnt{
			MAC:       mac,
			IPs:       ips,
			Nodes:     nodes,
			Vendor:    vendor,
			FirstTime: formatTimeUnix(agg.firstTime),
			LastTime:  formatTimeUnix(agg.lastTime),
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].MAC < list[j].MAC
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return mcpJSONResult(&list)
}

// get_wifi_ap_list tool
type mcpWifiAPEnt struct {
	ID             string              `json:"ID"`
	Host           string              `json:"Host"`
	BSSID          string              `json:"BSSID"`
	SSID           string              `json:"SSID"`
	Channel        string              `json:"Channel"`
	Info           string              `json:"Info"`
	Vendor         string              `json:"Vendor"`
	Count          int                 `json:"Count"`
	Change         int                 `json:"Change"`
	RSSICount      int                 `json:"RSSICount"`
	LatestRSSI     int                 `json:"LatestRSSI,omitempty"`
	LatestRSSITime string              `json:"LatestRSSITime,omitempty"`
	RSSI           []logreport.RSSIEnt `json:"RSSI,omitempty"`
	FirstTime      string              `json:"FirstTime"`
	LastTime       string              `json:"LastTime"`
	lastTimeRaw    int64
}

type getWifiAPListParams struct {
	SSIDFilter     string `json:"ssid_filter,omitempty" jsonschema:"Regex filter on SSID"`
	BSSIDFilter    string `json:"bssid_filter,omitempty" jsonschema:"Regex filter on BSSID"`
	HostFilter     string `json:"host_filter,omitempty" jsonschema:"Regex filter on sensor host"`
	VendorFilter   string `json:"vendor_filter,omitempty" jsonschema:"Regex filter on AP vendor"`
	IncludeHistory bool   `json:"include_history,omitempty" jsonschema:"Include full RSSI time-series history (default false)"`
	Limit          int    `json:"limit,omitempty" jsonschema:"Maximum number of APs to return (1-500, default 100)"`
}

func getWifiAPListTool(ctx context.Context, req *mcp.CallToolRequest, args getWifiAPListParams) (*mcp.CallToolResult, any, error) {
	ssidRe, err := compileRegexFilter("ssid_filter", args.SSIDFilter)
	if err != nil {
		return nil, nil, err
	}
	bssidRe, err := compileRegexFilter("bssid_filter", args.BSSIDFilter)
	if err != nil {
		return nil, nil, err
	}
	hostRe, err := compileRegexFilter("host_filter", args.HostFilter)
	if err != nil {
		return nil, nil, err
	}
	vendorRe, err := compileRegexFilter("vendor_filter", args.VendorFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 500, 100)
	logger.FlushReportEngine()
	rawItems, err := datastore.ListLogReportData(logreport.KindWifiAP)
	if err != nil {
		return nil, nil, err
	}
	list := []mcpWifiAPEnt{}
	for _, b := range rawItems {
		if ctx.Err() != nil {
			break
		}
		var ap logreport.WifiAPEnt
		if err := json.Unmarshal(b, &ap); err != nil {
			continue
		}
		if ssidRe != nil && !ssidRe.MatchString(ap.SSID) {
			continue
		}
		if bssidRe != nil && !bssidRe.MatchString(ap.BSSID) {
			continue
		}
		if hostRe != nil && !hostRe.MatchString(ap.Host) {
			continue
		}
		if vendorRe != nil && !vendorRe.MatchString(ap.Vendor) {
			continue
		}
		ent := mcpWifiAPEnt{
			ID:          ap.ID,
			Host:        ap.Host,
			BSSID:       ap.BSSID,
			SSID:        ap.SSID,
			Channel:     ap.Channel,
			Info:        ap.Info,
			Vendor:      ap.Vendor,
			Count:       ap.Count,
			Change:      ap.Change,
			RSSICount:   len(ap.RSSI),
			FirstTime:   formatTimeUnixNano(ap.FirstTime),
			LastTime:    formatTimeUnixNano(ap.LastTime),
			lastTimeRaw: ap.LastTime,
		}
		if len(ap.RSSI) > 0 {
			last := ap.RSSI[len(ap.RSSI)-1]
			ent.LatestRSSI = last.Value
			ent.LatestRSSITime = formatTimeUnixNano(last.Time)
		}
		if args.IncludeHistory {
			ent.RSSI = ap.RSSI
		}
		list = append(list, ent)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].lastTimeRaw == list[j].lastTimeRaw {
			return list[i].BSSID < list[j].BSSID
		}
		return list[i].lastTimeRaw > list[j].lastTimeRaw
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return mcpJSONResult(&list)
}

// get_bluetooth_device_list tool
type mcpBluetoothDeviceEnt struct {
	ID             string              `json:"ID"`
	Host           string              `json:"Host"`
	Address        string              `json:"Address"`
	AddressType    string              `json:"AddressType"`
	Name           string              `json:"Name"`
	Vendor         string              `json:"Vendor"`
	Info           string              `json:"Info"`
	Count          int                 `json:"Count"`
	RSSICount      int                 `json:"RSSICount"`
	LatestRSSI     int                 `json:"LatestRSSI,omitempty"`
	LatestRSSITime string              `json:"LatestRSSITime,omitempty"`
	RSSI           []logreport.RSSIEnt `json:"RSSI,omitempty"`
	FirstTime      string              `json:"FirstTime"`
	LastTime       string              `json:"LastTime"`
	lastTimeRaw    int64
}

type getBluetoothDeviceListParams struct {
	NameFilter        string `json:"name_filter,omitempty" jsonschema:"Regex filter on device name"`
	AddressFilter     string `json:"address_filter,omitempty" jsonschema:"Regex filter on Bluetooth MAC address"`
	HostFilter        string `json:"host_filter,omitempty" jsonschema:"Regex filter on sensor host"`
	VendorFilter      string `json:"vendor_filter,omitempty" jsonschema:"Regex filter on vendor name"`
	PublicAddressOnly bool   `json:"public_address_only,omitempty" jsonschema:"Only return devices with Public Bluetooth addresses"`
	IncludeHistory    bool   `json:"include_history,omitempty" jsonschema:"Include full RSSI time-series history (default false)"`
	Limit             int    `json:"limit,omitempty" jsonschema:"Maximum number of devices to return (1-500, default 100)"`
}

func getBluetoothDeviceListTool(ctx context.Context, req *mcp.CallToolRequest, args getBluetoothDeviceListParams) (*mcp.CallToolResult, any, error) {
	nameRe, err := compileRegexFilter("name_filter", args.NameFilter)
	if err != nil {
		return nil, nil, err
	}
	addrRe, err := compileRegexFilter("address_filter", args.AddressFilter)
	if err != nil {
		return nil, nil, err
	}
	hostRe, err := compileRegexFilter("host_filter", args.HostFilter)
	if err != nil {
		return nil, nil, err
	}
	vendorRe, err := compileRegexFilter("vendor_filter", args.VendorFilter)
	if err != nil {
		return nil, nil, err
	}
	limit := clampLimit(args.Limit, 1, 500, 100)
	logger.FlushReportEngine()
	rawItems, err := datastore.ListLogReportData(logreport.KindBlueDevice)
	if err != nil {
		return nil, nil, err
	}
	list := []mcpBluetoothDeviceEnt{}
	for _, b := range rawItems {
		if ctx.Err() != nil {
			break
		}
		var dev logreport.BlueDeviceEnt
		if err := json.Unmarshal(b, &dev); err != nil {
			continue
		}
		if args.PublicAddressOnly && !strings.EqualFold(strings.TrimSpace(dev.AddressType), "public") {
			continue
		}
		if nameRe != nil && !nameRe.MatchString(dev.Name) {
			continue
		}
		if addrRe != nil && !addrRe.MatchString(dev.Address) {
			continue
		}
		if hostRe != nil && !hostRe.MatchString(dev.Host) {
			continue
		}
		if vendorRe != nil && !vendorRe.MatchString(dev.Vendor) {
			continue
		}
		ent := mcpBluetoothDeviceEnt{
			ID:          dev.ID,
			Host:        dev.Host,
			Address:     dev.Address,
			AddressType: dev.AddressType,
			Name:        dev.Name,
			Vendor:      dev.Vendor,
			Info:        dev.Info,
			Count:       dev.Count,
			RSSICount:   len(dev.RSSI),
			FirstTime:   formatTimeUnixNano(dev.FirstTime),
			LastTime:    formatTimeUnixNano(dev.LastTime),
			lastTimeRaw: dev.LastTime,
		}
		if len(dev.RSSI) > 0 {
			last := dev.RSSI[len(dev.RSSI)-1]
			ent.LatestRSSI = last.Value
			ent.LatestRSSITime = formatTimeUnixNano(last.Time)
		}
		if args.IncludeHistory {
			ent.RSSI = dev.RSSI
		}
		list = append(list, ent)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].lastTimeRaw == list[j].lastTimeRaw {
			return list[i].Address < list[j].Address
		}
		return list[i].lastTimeRaw > list[j].lastTimeRaw
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return mcpJSONResult(&list)
}

// get_fdb_table tool
type getFDBTableParams struct {
	ID string `json:"id" jsonschema:"Network ID (with or without NET: prefix), network name, or IP"`
}

func getFDBTableTool(ctx context.Context, req *mcp.CallToolRequest, args getFDBTableParams) (*mcp.CallToolResult, any, error) {
	if datastore.MIBDB == nil {
		return nil, nil, fmt.Errorf("MIB database is not loaded")
	}
	n, err := resolveNetwork(args.ID)
	if err != nil {
		if node, nerr := resolveNode(args.ID, true, true); nerr == nil && node != nil {
			n = datastore.FindNetworkByIP(node.IP)
		}
		if n == nil {
			return nil, nil, err
		}
	}
	table := GetFDBTable(n.ID)
	if table == nil {
		table = []*FDBTableEnt{}
	}
	return mcpJSONResult(&table)
}

// auto_connect_lines tool
type autoConnectLinesParams struct {
	Mode string `json:"mode,omitempty" jsonschema:"Auto-connect mode: 'strict' (default) or 'speculative'"`
}

func autoConnectLinesTool(ctx context.Context, req *mcp.CallToolRequest, args autoConnectLinesParams) (*mcp.CallToolResult, any, error) {
	mode := datastore.AutoLineStrict
	modeName := "strict"
	switch strings.ToLower(strings.TrimSpace(args.Mode)) {
	case "", "strict":
		mode = datastore.AutoLineStrict
		modeName = "strict"
	case "speculative":
		mode = datastore.AutoLineSpeculative
		modeName = "speculative"
	default:
		return nil, nil, fmt.Errorf("invalid mode %q (expected strict or speculative)", args.Mode)
	}
	added, total, err := AutoConnectLines(mode)
	if err != nil {
		return nil, nil, err
	}
	addMCPEventLog("info", "", "", fmt.Sprintf("Auto connect lines via MCP: mode=%s added=%d total=%d", modeName, added, total))
	return mcpJSONResult(&struct {
		Mode       string `json:"Mode"`
		AddedLines int    `json:"AddedLines"`
		TotalLines int    `json:"TotalLines"`
	}{
		Mode:       modeName,
		AddedLines: added,
		TotalLines: total,
	})
}

// auto_layout tool
type autoLayoutParams struct {
	Mode string `json:"mode,omitempty" jsonschema:"Layout mode: 'hierarchical' (default), 'cluster', or 'categorized'"`
	Undo bool   `json:"undo,omitempty" jsonschema:"Set true to undo the previous auto layout operation"`
}

func autoLayoutTool(ctx context.Context, req *mcp.CallToolRequest, args autoLayoutParams) (*mcp.CallToolResult, any, error) {
	if args.Undo {
		if !HasUndoLayout() {
			return nil, nil, fmt.Errorf("no layout snapshot available to undo")
		}
		restored, err := UndoLayout()
		if err != nil {
			return nil, nil, err
		}
		addMCPEventLog("info", "", "", fmt.Sprintf("Undo map layout via MCP: restored=%d", restored))
		return mcpJSONResult(&struct {
			Mode          string `json:"Mode"`
			MovedNodes    int    `json:"MovedNodes"`
			UndoAvailable bool   `json:"UndoAvailable"`
		}{
			Mode:          "undo",
			MovedNodes:    restored,
			UndoAvailable: HasUndoLayout(),
		})
	}
	modeInt := datastore.AutoLayoutHierarchical
	modeName := "hierarchical"
	switch strings.ToLower(strings.TrimSpace(args.Mode)) {
	case "", "hierarchical":
		modeInt = datastore.AutoLayoutHierarchical
		modeName = "hierarchical"
	case "cluster":
		modeInt = datastore.AutoLayoutCluster
		modeName = "cluster"
	case "categorized":
		modeInt = datastore.AutoLayoutCategorized
		modeName = "categorized"
	default:
		return nil, nil, fmt.Errorf("invalid layout mode %q (expected hierarchical, cluster, categorized)", args.Mode)
	}
	moved, err := OptimizeLayout(modeInt)
	if err != nil {
		return nil, nil, err
	}
	addMCPEventLog("info", "", "", fmt.Sprintf("Auto layout map via MCP: mode=%s moved=%d", modeName, moved))
	return mcpJSONResult(&struct {
		Mode          string `json:"Mode"`
		MovedNodes    int    `json:"MovedNodes"`
		UndoAvailable bool   `json:"UndoAvailable"`
	}{
		Mode:          modeName,
		MovedNodes:    moved,
		UndoAvailable: HasUndoLayout(),
	})
}

// gNMI tools
type gnmiCapabilitiesParams struct {
	NodeID string `json:"node_id" jsonschema:"Node ID, unique node name, or unique IP address"`
	Target string `json:"target,omitempty" jsonschema:"Optional gNMI target host:port override (defaults to node IP and GNMIPort or 57400)"`
}

type mcpGNMIModelData struct {
	Name         string `json:"Name"`
	Organization string `json:"Organization"`
	Version      string `json:"Version"`
}

type mcpGNMICapabilitiesResp struct {
	NodeID    string             `json:"NodeID"`
	NodeName  string             `json:"NodeName"`
	Target    string             `json:"Target"`
	Version   string             `json:"Version"`
	Encodings string             `json:"Encodings"`
	Models    []mcpGNMIModelData `json:"Models"`
}

func resolveGNMITarget(n *datastore.NodeEnt, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	if n.IP == "" {
		return "", fmt.Errorf("node %s has no IP address", n.Name)
	}
	port := strings.TrimSpace(n.GNMIPort)
	if port == "" {
		port = "57400"
	}
	if strings.Contains(port, ":") {
		return port, nil
	}
	return net.JoinHostPort(n.IP, port), nil
}

func gnmiCapabilitiesTool(ctx context.Context, req *mcp.CallToolRequest, args gnmiCapabilitiesParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNode(args.NodeID, true, true)
	if err != nil {
		return nil, nil, err
	}
	target, err := resolveGNMITarget(n, args.Target)
	if err != nil {
		return nil, nil, err
	}
	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(target),
		api.Username(n.GNMIUser),
		api.Password(n.GNMIPassword),
		api.SkipVerify(true),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gnmi target error: %w", err)
	}
	gctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := tg.CreateGNMIClient(gctx); err != nil {
		return nil, nil, fmt.Errorf("gnmi connect error (%s): %w", target, err)
	}
	defer tg.Close()
	capResp, err := tg.Capabilities(gctx)
	if err != nil {
		return nil, nil, fmt.Errorf("gnmi capabilities error (%s): %w", target, err)
	}
	models := []mcpGNMIModelData{}
	for _, m := range capResp.GetSupportedModels() {
		models = append(models, mcpGNMIModelData{
			Name:         m.GetName(),
			Organization: m.GetOrganization(),
			Version:      m.GetVersion(),
		})
	}
	encs := []string{}
	for _, e := range capResp.GetSupportedEncodings() {
		encs = append(encs, e.String())
	}
	return mcpJSONResult(&mcpGNMICapabilitiesResp{
		NodeID:    n.ID,
		NodeName:  n.Name,
		Target:    target,
		Version:   capResp.GetGNMIVersion(),
		Encodings: strings.Join(encs, ","),
		Models:    models,
	})
}

type gnmiGetParams struct {
	NodeID   string `json:"node_id" jsonschema:"Node ID, unique node name, or unique IP address"`
	Path     string `json:"path" jsonschema:"gNMI telemetry path (e.g. /interfaces/interface or /system)"`
	Target   string `json:"target,omitempty" jsonschema:"Optional gNMI target host:port override"`
	Encoding string `json:"encoding,omitempty" jsonschema:"gNMI encoding (json_ietf, json, proto, bytes, ascii; default node GNMIEncoding or json_ietf)"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum number of leaf path entries to return (1-5000, default 500)"`
}

type mcpGNMIGetEnt struct {
	Path  string `json:"Path"`
	Index string `json:"Index,omitempty"`
	Value string `json:"Value"`
}

func extractGNMIPathValues(d any, path, index string, inArray bool) []mcpGNMIGetEnt {
	r := []mcpGNMIGetEnt{}
	switch v := d.(type) {
	case string:
		r = append(r, mcpGNMIGetEnt{Path: path, Value: v, Index: index})
	case float64:
		r = append(r, mcpGNMIGetEnt{Path: path, Value: strconv.FormatFloat(v, 'f', -1, 64), Index: index})
	case bool:
		r = append(r, mcpGNMIGetEnt{Path: path, Value: strconv.FormatBool(v), Index: index})
	case map[string]any:
		nameKey := ""
		if in, ok := v["name"]; ok {
			if sn, ok := in.(string); ok {
				nameKey = sn
			}
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			vv := v[k]
			if inArray && nameKey != "" {
				r = append(r, extractGNMIPathValues(vv, fmt.Sprintf("%s[name=%s]/%s", path, nameKey, k), index, false)...)
			} else {
				r = append(r, extractGNMIPathValues(vv, path+"/"+k, index, false)...)
			}
		}
	case []any:
		for i, vv := range v {
			r = append(r, extractGNMIPathValues(vv, path, strconv.Itoa(i), true)...)
		}
	case nil:
	default:
		r = append(r, mcpGNMIGetEnt{Path: path, Value: fmt.Sprintf("%v", v), Index: index})
	}
	return r
}

func gnmiGetTool(ctx context.Context, req *mcp.CallToolRequest, args gnmiGetParams) (*mcp.CallToolResult, any, error) {
	pathArg := strings.TrimSpace(args.Path)
	if pathArg == "" {
		return nil, nil, fmt.Errorf("path must not be empty")
	}
	n, err := resolveNode(args.NodeID, true, true)
	if err != nil {
		return nil, nil, err
	}
	target, err := resolveGNMITarget(n, args.Target)
	if err != nil {
		return nil, nil, err
	}
	enc := strings.TrimSpace(args.Encoding)
	if enc == "" {
		enc = n.GNMIEncoding
	}
	if enc == "" {
		enc = "json_ietf"
	}
	limit := clampLimit(args.Limit, 1, 5000, 500)

	tg, err := api.NewTarget(
		api.Name(n.Name),
		api.Address(target),
		api.Username(n.GNMIUser),
		api.Password(n.GNMIPassword),
		api.SkipVerify(true),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gnmi target error: %w", err)
	}
	gctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := tg.CreateGNMIClient(gctx); err != nil {
		return nil, nil, fmt.Errorf("gnmi connect error (%s): %w", target, err)
	}
	defer tg.Close()

	getReq, err := api.NewGetRequest(api.Path(pathArg), api.Encoding(enc))
	if err != nil {
		return nil, nil, fmt.Errorf("gnmi request error: %w", err)
	}
	getResp, err := tg.Get(gctx, getReq)
	if err != nil {
		return nil, nil, fmt.Errorf("gnmi get error (%s %s): %w", target, pathArg, err)
	}
	entries := []mcpGNMIGetEnt{}
	for _, not := range getResp.GetNotification() {
		for _, u := range not.GetUpdate() {
			elems := make([]string, 0, len(u.GetPath().GetElem()))
			for _, elem := range u.GetPath().GetElem() {
				elems = append(elems, elem.GetName())
			}
			basePath := ""
			if len(elems) > 0 {
				basePath = "/" + strings.Join(elems, "/")
			}
			if val := u.GetVal(); val != nil {
				rawJSON := val.GetJsonIetfVal()
				if len(rawJSON) == 0 {
					rawJSON = val.GetJsonVal()
				}
				if len(rawJSON) > 0 {
					var d any
					if err := json.Unmarshal(rawJSON, &d); err == nil {
						entries = append(entries, extractGNMIPathValues(d, basePath, "", false)...)
					}
				} else {
					entries = append(entries, mcpGNMIGetEnt{
						Path:  basePath,
						Value: val.String(),
					})
				}
			}
		}
	}
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return mcpJSONResult(&entries)
}

// analyze_polling_correlation tool
type analyzePollingCorrelationParams struct {
	PollingID1   string `json:"polling_id_1" jsonschema:"First polling ID"`
	PollingID2   string `json:"polling_id_2" jsonschema:"Second polling ID"`
	Vars1        string `json:"vars_1,omitempty" jsonschema:"Comma-separated numeric variable keys for polling 1 (empty for auto)"`
	Vars2        string `json:"vars_2,omitempty" jsonschema:"Comma-separated numeric variable keys for polling 2 (empty for auto)"`
	Algorithm    string `json:"algorithm,omitempty" jsonschema:"Primary algorithm: all (default), pearson, spearman, kendall, crosscorr, dcor, mlp, lstm, or knn"`
	BucketSec    int64  `json:"bucket_sec,omitempty" jsonschema:"Aggregation time bucket width in seconds (0 for auto)"`
	PeriodHours  int    `json:"period_hours,omitempty" jsonschema:"Lookback period in hours (default 168, -1 for all logs)"`
	MaxLag       int    `json:"max_lag,omitempty" jsonschema:"Maximum lag buckets for cross-correlation (default 12, max 60)"`
	IncludeState bool   `json:"include_state,omitempty" jsonschema:"Include polling state numeric variable (1.0=normal, 0.5=unknown, 0.0=error)"`
}

type mcpCorrelationPairSummary struct {
	PollingID1    string                  `json:"PollingID1"`
	NodeName1     string                  `json:"NodeName1"`
	PollingName1  string                  `json:"PollingName1"`
	SelectedVar1  string                  `json:"SelectedVar1"`
	PollingID2    string                  `json:"PollingID2"`
	NodeName2     string                  `json:"NodeName2"`
	PollingName2  string                  `json:"PollingName2"`
	SelectedVar2  string                  `json:"SelectedVar2"`
	Vars1         []string                `json:"Vars1"`
	Vars2         []string                `json:"Vars2"`
	BucketSec     int64                   `json:"BucketSec"`
	SampleCount   int                     `json:"SampleCount"`
	PrimaryScore  float64                 `json:"PrimaryScore"`
	PrimaryPValue float64                 `json:"PrimaryPValue"`
	BestLag       int                     `json:"BestLag"`
	BestLagSec    int64                   `json:"BestLagSec"`
	Slope         float64                 `json:"Slope"`
	Intercept     float64                 `json:"Intercept"`
	Scores             []CorrelationAlgoScore         `json:"Scores"`
	VarMatrix          []CorrelationMatrixCell        `json:"VarMatrix,omitempty"`
	LagProfile         []CorrelationLagPoint          `json:"LagProfile,omitempty"`
	FeatureImportances []CorrelationFeatureImportance `json:"FeatureImportances,omitempty"`
	RegimeResult       *CorrelationRegimeResult       `json:"RegimeResult,omitempty"`
}

func splitCSVKeys(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func analyzePollingCorrelationTool(ctx context.Context, req *mcp.CallToolRequest, args analyzePollingCorrelationParams) (*mcp.CallToolResult, any, error) {
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	id1 := strings.TrimSpace(args.PollingID1)
	id2 := strings.TrimSpace(args.PollingID2)
	if id1 == "" || id2 == "" {
		return nil, nil, fmt.Errorf("polling_id_1 and polling_id_2 must not be empty")
	}
	var startTime, endTime int64
	if args.PeriodHours >= 0 {
		hours := args.PeriodHours
		if hours == 0 {
			hours = 168
		}
		endTime = time.Now().Unix()
		startTime = endTime - int64(hours)*3600
	}
	algo := strings.TrimSpace(args.Algorithm)
	if algo == "" {
		algo = "all"
	}
	res := AnalyzePollingCorrelation(CorrelationPairReq{
		PollingID1:   id1,
		PollingID2:   id2,
		Vars1:        splitCSVKeys(args.Vars1),
		Vars2:        splitCSVKeys(args.Vars2),
		IncludeState: args.IncludeState,
		Algorithm:    algo,
		BucketSec:    args.BucketSec,
		StartTime:    startTime,
		EndTime:      endTime,
		MaxLag:       args.MaxLag,
	})
	if res.Error != "" {
		return nil, nil, fmt.Errorf("%s", res.Error)
	}
	summary := mcpCorrelationPairSummary{
		PollingID1:    res.PollingID1,
		NodeName1:     res.NodeName1,
		PollingName1:  res.PollingName1,
		SelectedVar1:  res.SelectedVar1,
		PollingID2:    res.PollingID2,
		NodeName2:     res.NodeName2,
		PollingName2:  res.PollingName2,
		SelectedVar2:  res.SelectedVar2,
		Vars1:         res.Vars1,
		Vars2:         res.Vars2,
		BucketSec:     res.BucketSec,
		SampleCount:   res.SampleCount,
		PrimaryScore:  res.PrimaryScore,
		PrimaryPValue: res.PrimaryPValue,
		BestLag:       res.BestLag,
		BestLagSec:    res.BestLagSec,
		Slope:         res.Slope,
		Intercept:     res.Intercept,
		Scores:             res.Scores,
		VarMatrix:          res.VarMatrix,
		LagProfile:         res.LagProfile,
		FeatureImportances: res.FeatureImportances,
		RegimeResult:       res.RegimeResult,
	}
	return mcpJSONResult(&summary)
}

// discover_polling_correlations tool
type discoverPollingCorrelationsParams struct {
	BasePollingID string  `json:"base_polling_id,omitempty" jsonschema:"Optional polling ID to find correlations against (empty for all-vs-all)"`
	Algorithm     string  `json:"algorithm,omitempty" jsonschema:"Correlation algorithm: pearson (default), spearman, kendall, crosscorr, dcor, mlp, or knn"`
	BucketSec     int64   `json:"bucket_sec,omitempty" jsonschema:"Aggregation time bucket width in seconds (0 for auto)"`
	PeriodHours   int     `json:"period_hours,omitempty" jsonschema:"Lookback period in hours (default 168, -1 for all logs)"`
	MinAbsScore   float64 `json:"min_abs_score,omitempty" jsonschema:"Minimum absolute correlation threshold (default 0.7)"`
	MinSamples    int     `json:"min_samples,omitempty" jsonschema:"Minimum overlapping samples required (default 10)"`
	CrossNodeOnly bool    `json:"cross_node_only,omitempty" jsonschema:"Only return correlated pairs on different nodes"`
	IncludeState  bool    `json:"include_state,omitempty" jsonschema:"Include polling state numeric variable"`
	MaxLag        int     `json:"max_lag,omitempty" jsonschema:"Maximum lag buckets when algorithm is crosscorr (default 6)"`
	Limit         int     `json:"limit,omitempty" jsonschema:"Maximum number of pairs to return (1-500, default 30)"`
}

func discoverPollingCorrelationsTool(ctx context.Context, req *mcp.CallToolRequest, args discoverPollingCorrelationsParams) (*mcp.CallToolResult, any, error) {
	var startTime, endTime int64
	if args.PeriodHours >= 0 {
		hours := args.PeriodHours
		if hours == 0 {
			hours = 168
		}
		endTime = time.Now().Unix()
		startTime = endTime - int64(hours)*3600
	}
	limit := clampLimit(args.Limit, 1, 500, 30)
	res := DiscoverPollingCorrelations(ctx, CorrelationDiscoverReq{
		BasePollingID: strings.TrimSpace(args.BasePollingID),
		Algorithm:     args.Algorithm,
		BucketSec:     args.BucketSec,
		StartTime:     startTime,
		EndTime:       endTime,
		MinAbsScore:   args.MinAbsScore,
		MinSamples:    args.MinSamples,
		CrossNodeOnly: args.CrossNodeOnly,
		IncludeState:  args.IncludeState,
		MaxLag:        args.MaxLag,
		Limit:         limit,
	})
	if res.Error != "" {
		return nil, nil, fmt.Errorf("%s", res.Error)
	}
	return mcpJSONResult(&res)
}
