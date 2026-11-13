package backend

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/ping"
)

func addCRUDTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node",
		Description: "get detailed information and polling summary for a single node from TWSNMP (credentials masked)",
	}, getNodeTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_node",
		Description: "delete a node, its pollings, and its connected lines from TWSNMP",
	}, deleteNodeTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_polling",
		Description: "get full configuration and latest result of a single polling from TWSNMP",
	}, getPollingTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_polling",
		Description: "add a new polling check to a node in TWSNMP",
	}, addPollingTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_polling",
		Description: "update an existing polling configuration in TWSNMP",
	}, updatePollingTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_polling",
		Description: "delete a polling from TWSNMP",
	}, deletePollingTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "check_polling",
		Description: "trigger immediate execution of a polling, all pollings for a node, or all pollings in TWSNMP",
	}, checkPollingTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_line_list",
		Description: "get map lines connecting nodes and networks in TWSNMP",
	}, getLineListTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_line",
		Description: "add a map line between two nodes or networks in TWSNMP",
	}, addLineTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_line",
		Description: "delete a map line from TWSNMP",
	}, deleteLineTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_network",
		Description: "get detailed information and port table for a network container in TWSNMP (credentials masked)",
	}, getNetworkTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_network",
		Description: "add a network (switch/hub) container to TWSNMP map",
	}, addNetworkTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_network",
		Description: "update a network container in TWSNMP",
	}, updateNetworkTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_network",
		Description: "delete a network container and its attached map lines from TWSNMP",
	}, deleteNetworkTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "check_network",
		Description: "trigger immediate SNMP/LLDP status check for one or all network containers in TWSNMP",
	}, checkNetworkTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_server_certificate",
		Description: "add or update a TLS server certificate monitor entry in TWSNMP",
	}, updateServerCertificateTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_server_certificate",
		Description: "delete a TLS server certificate monitor entry from TWSNMP",
	}, deleteServerCertificateTool)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_system_status",
		Description: "get TWSNMP system status, runtime metrics, map counts, and receiver statuses",
	}, getSystemStatusTool)
}

// get_node tool
type getNodeParams struct {
	ID string `json:"id" jsonschema:"Node ID, unique node name, or unique IP address"`
}

type mcpNodeDetailEnt struct {
	ID           string
	Name         string
	IP           string
	MAC          string
	Vendor       string
	State        string
	X            int
	Y            int
	Icon         string
	Description  string
	SnmpMode     string
	SnmpPort     int
	Community    string
	User         string
	SSHUser      string
	Password     string
	GNMIPort     string
	GNMIEncoding string
	GNMIUser     string
	GNMIPassword string
	PublicKey    string
	URL          string
	AddrMode     string
	AutoAck      bool
	Loc          string
	Pollings     []mcpPollingEnt
}

func getNodeTool(ctx context.Context, req *mcp.CallToolRequest, args getNodeParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNode(args.ID, true, true)
	if err != nil {
		return nil, nil, err
	}
	pollings := []mcpPollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID != n.ID {
			return true
		}
		pollings = append(pollings, mcpPollingEnt{
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
	sort.Slice(pollings, func(i, j int) bool {
		return pollings[i].ID < pollings[j].ID
	})
	res := mcpNodeDetailEnt{
		ID:           n.ID,
		Name:         n.Name,
		IP:           n.IP,
		MAC:          n.MAC,
		Vendor:       n.Vendor,
		State:        n.State,
		X:            n.X,
		Y:            n.Y,
		Icon:         n.Icon,
		Description:  n.Descr,
		SnmpMode:     n.SnmpMode,
		SnmpPort:     n.SnmpPort,
		Community:    maskSecret(n.Community),
		User:         n.User,
		SSHUser:      n.SSHUser,
		Password:     maskSecret(n.Password),
		GNMIPort:     n.GNMIPort,
		GNMIEncoding: n.GNMIEncoding,
		GNMIUser:     n.GNMIUser,
		GNMIPassword: maskSecret(n.GNMIPassword),
		PublicKey:    maskSecret(n.PublicKey),
		URL:          n.URL,
		AddrMode:     n.AddrMode,
		AutoAck:      n.AutoAck,
		Loc:          n.Loc,
		Pollings:     pollings,
	}
	return mcpJSONResult(&res)
}

// delete_node tool
type deleteNodeParams struct {
	ID string `json:"id" jsonschema:"Node ID, unique node name, or unique IP address to delete"`
}

func deleteNodeTool(ctx context.Context, req *mcp.CallToolRequest, args deleteNodeParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNode(args.ID, true, true)
	if err != nil {
		return nil, nil, err
	}
	hooks := pollingHooks.Load()
	pollingCount := 0
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == n.ID {
			pollingCount++
			if hooks != nil && hooks.StopPolling != nil {
				hooks.StopPolling(p)
			}
		}
		return true
	})
	lineIDs := []string{}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 == n.ID || l.NodeID2 == n.ID {
			lineIDs = append(lineIDs, l.ID)
		}
		return true
	})
	if err := datastore.DeleteNode(n.ID); err != nil {
		return nil, nil, err
	}
	deletedLines := 0
	for _, lid := range lineIDs {
		if datastore.GetLine(lid) == nil {
			// Already cleaned up via DeletePollings inside DeleteNode.
			deletedLines++
			continue
		}
		if err := datastore.DeleteLine(lid); err == nil {
			deletedLines++
		}
	}
	addMCPEventLog("info", n.ID, n.Name, fmt.Sprintf("Delete node via MCP: %s (%s)", n.Name, n.IP))
	return mcpJSONResult(&struct {
		Deleted         bool   `json:"Deleted"`
		ID              string `json:"ID"`
		Name            string `json:"Name"`
		IP              string `json:"IP"`
		DeletedPollings int    `json:"DeletedPollings"`
		DeletedLines    int    `json:"DeletedLines"`
	}{
		Deleted:         true,
		ID:              n.ID,
		Name:            n.Name,
		IP:              n.IP,
		DeletedPollings: pollingCount,
		DeletedLines:    deletedLines,
	})
}

// Polling CRUD helpers & tools
type mcpPollingDetailEnt struct {
	ID           string         `json:"ID"`
	Name         string         `json:"Name"`
	NodeID       string         `json:"NodeID"`
	NodeName     string         `json:"NodeName"`
	Type         string         `json:"Type"`
	Mode         string         `json:"Mode"`
	Params       string         `json:"Params"`
	Filter       string         `json:"Filter"`
	Extractor    string         `json:"Extractor"`
	Script       string         `json:"Script"`
	Level        string         `json:"Level"`
	PollInt      int            `json:"PollInt"`
	Timeout      int            `json:"Timeout"`
	Retry        int            `json:"Retry"`
	LogMode      string         `json:"LogMode"`
	LastTime     string         `json:"LastTime"`
	NextTime     string         `json:"NextTime"`
	FailTime     string         `json:"FailTime"`
	State        string         `json:"State"`
	FailAction   string         `json:"FailAction"`
	RepairAction string         `json:"RepairAction"`
	AIMode       string         `json:"AIMode"`
	VectorCols   string         `json:"VectorCols"`
	MqttURL      string         `json:"MqttURL"`
	MqttTopic    string         `json:"MqttTopic"`
	MqttCols     string         `json:"MqttCols"`
	Result       map[string]any `json:"Result"`
}

func toMCPPollingDetail(p *datastore.PollingEnt, nodeName string) *mcpPollingDetailEnt {
	return &mcpPollingDetailEnt{
		ID:           p.ID,
		Name:         p.Name,
		NodeID:       p.NodeID,
		NodeName:     nodeName,
		Type:         p.Type,
		Mode:         p.Mode,
		Params:       p.Params,
		Filter:       p.Filter,
		Extractor:    p.Extractor,
		Script:       p.Script,
		Level:        p.Level,
		PollInt:      p.PollInt,
		Timeout:      p.Timeout,
		Retry:        p.Retry,
		LogMode:      logModeString(p.LogMode),
		LastTime:     formatTimeUnixNano(p.LastTime),
		NextTime:     formatTimeUnixNano(p.NextTime),
		FailTime:     formatTimeUnixNano(p.FailTime),
		State:        p.State,
		FailAction:   p.FailAction,
		RepairAction: p.RepairAction,
		AIMode:       p.AIMode,
		VectorCols:   p.VectorCols,
		MqttURL:      p.MqttURL,
		MqttTopic:    p.MqttTopic,
		MqttCols:     p.MqttCols,
		Result:       p.Result,
	}
}

func validatePollingLevel(level string) (string, error) {
	l := strings.ToLower(strings.TrimSpace(level))
	switch l {
	case "", "low", "warn", "high", "off":
		return l, nil
	default:
		return "", fmt.Errorf("invalid level %q (expected off, low, warn, high)", level)
	}
}

type getPollingParams struct {
	ID string `json:"id" jsonschema:"Polling ID"`
}

func getPollingTool(ctx context.Context, req *mcp.CallToolRequest, args getPollingParams) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(args.ID)
	if id == "" {
		return nil, nil, fmt.Errorf("polling id must not be empty")
	}
	p := datastore.GetPolling(id)
	if p == nil {
		return nil, nil, fmt.Errorf("polling not found: %s", id)
	}
	nodeName := ""
	if n := datastore.GetNode(p.NodeID); n != nil {
		nodeName = n.Name
	}
	return mcpJSONResult(toMCPPollingDetail(p, nodeName))
}

type addPollingParams struct {
	NodeID       string `json:"node_id" jsonschema:"Target node ID, unique node name, or unique IP address"`
	Name         string `json:"name" jsonschema:"Polling name"`
	Type         string `json:"type" jsonschema:"Polling type (ping, tcp, http, dns, tls, snmp, gnmi, mqtt, syslog, trap, netflow, sflow, twsnmp, ntp, ssh, stun, cmd, report, etc.)"`
	Mode         string `json:"mode,omitempty" jsonschema:"Polling sub-mode (e.g. get, walk, count, subscribe)"`
	Params       string `json:"params,omitempty" jsonschema:"Polling target parameter (port, URL, OID, hostname, etc.)"`
	Filter       string `json:"filter,omitempty" jsonschema:"Filter pattern for log/SNMP/HTTP pollings"`
	Extractor    string `json:"extractor,omitempty" jsonschema:"Grok/value extractor pattern"`
	Script       string `json:"script,omitempty" jsonschema:"Evaluation JavaScript for polling result"`
	Level        string `json:"level,omitempty" jsonschema:"Failure severity level: off, low (default), warn, or high"`
	PollInt      int    `json:"poll_int,omitempty" jsonschema:"Polling interval in seconds (default map PollInt or 60)"`
	Timeout      int    `json:"timeout,omitempty" jsonschema:"Polling timeout in seconds (default map Timeout or 3)"`
	Retry        int    `json:"retry,omitempty" jsonschema:"Polling retry count (default map Retry or 0)"`
	LogMode      string `json:"log_mode,omitempty" jsonschema:"Logging mode: none (default), always, on_change, or ai"`
	FailAction   string `json:"fail_action,omitempty" jsonschema:"Action definition executed on failure"`
	RepairAction string `json:"repair_action,omitempty" jsonschema:"Action definition executed on recovery"`
	AIMode       string `json:"ai_mode,omitempty" jsonschema:"AI anomaly detection mode"`
	VectorCols   string `json:"vector_cols,omitempty" jsonschema:"Comma-separated vector columns for AI analysis"`
	MqttURL      string `json:"mqtt_url,omitempty" jsonschema:"MQTT broker URL for mqtt polling"`
	MqttTopic    string `json:"mqtt_topic,omitempty" jsonschema:"MQTT subscription topic"`
	MqttCols     string `json:"mqtt_cols,omitempty" jsonschema:"MQTT JSON field extraction columns"`
}

func addPollingTool(ctx context.Context, req *mcp.CallToolRequest, args addPollingParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNode(args.NodeID, true, true)
	if err != nil {
		return nil, nil, err
	}
	name := strings.TrimSpace(args.Name)
	if name == "" {
		return nil, nil, fmt.Errorf("polling name must not be empty")
	}
	ptype := strings.TrimSpace(args.Type)
	if ptype == "" {
		return nil, nil, fmt.Errorf("polling type must not be empty")
	}
	level, err := validatePollingLevel(args.Level)
	if err != nil {
		return nil, nil, err
	}
	if level == "" {
		level = "low"
	}
	logMode, err := parseLogModeInt(args.LogMode)
	if err != nil {
		return nil, nil, err
	}
	pollInt := args.PollInt
	if pollInt <= 0 {
		pollInt = datastore.MapConf.PollInt
		if pollInt <= 0 {
			pollInt = 60
		}
	}
	timeout := args.Timeout
	if timeout <= 0 {
		timeout = datastore.MapConf.Timeout
		if timeout <= 0 {
			timeout = 3
		}
	}
	retry := args.Retry
	if retry < 0 {
		retry = 0
	} else if retry == 0 && datastore.MapConf.Retry > 0 {
		retry = datastore.MapConf.Retry
	}
	p := &datastore.PollingEnt{
		NodeID:       n.ID,
		Name:         name,
		Type:         ptype,
		Mode:         args.Mode,
		Params:       args.Params,
		Filter:       args.Filter,
		Extractor:    args.Extractor,
		Script:       args.Script,
		Level:        level,
		PollInt:      pollInt,
		Timeout:      timeout,
		Retry:        retry,
		LogMode:      logMode,
		State:        "unknown",
		NextTime:     0,
		FailAction:   args.FailAction,
		RepairAction: args.RepairAction,
		AIMode:       args.AIMode,
		VectorCols:   args.VectorCols,
		MqttURL:      args.MqttURL,
		MqttTopic:    args.MqttTopic,
		MqttCols:     args.MqttCols,
	}
	if err := datastore.AddPolling(p); err != nil {
		return nil, nil, err
	}
	datastore.SetNodeStateChanged(n.ID)
	addMCPEventLog("info", n.ID, n.Name, fmt.Sprintf("Add polling via MCP: %s (%s)", p.Name, p.Type))
	return mcpJSONResult(toMCPPollingDetail(p, n.Name))
}

type updatePollingParams struct {
	ID           string `json:"id" jsonschema:"Polling ID to update"`
	NodeID       string `json:"node_id,omitempty" jsonschema:"Optional new node ID, unique node name, or IP"`
	Name         string `json:"name,omitempty" jsonschema:"Optional new polling name"`
	Type         string `json:"type,omitempty" jsonschema:"Optional new polling type"`
	Mode         string `json:"mode,omitempty" jsonschema:"Optional new polling sub-mode"`
	Params       string `json:"params,omitempty" jsonschema:"Optional new target params"`
	Filter       string `json:"filter,omitempty" jsonschema:"Optional new filter pattern"`
	Extractor    string `json:"extractor,omitempty" jsonschema:"Optional new extractor pattern"`
	Script       string `json:"script,omitempty" jsonschema:"Optional new evaluation script"`
	Level        string `json:"level,omitempty" jsonschema:"Optional new failure level (off, low, warn, high)"`
	PollInt      int    `json:"poll_int,omitempty" jsonschema:"Optional new interval in seconds (>0)"`
	Timeout      int    `json:"timeout,omitempty" jsonschema:"Optional new timeout in seconds (>0)"`
	Retry        int    `json:"retry,omitempty" jsonschema:"Optional new retry count (>0 to set, negative e.g. -1 resets to 0)"`
	LogMode      string `json:"log_mode,omitempty" jsonschema:"Optional new log mode (none, always, on_change, ai)"`
	FailAction   string `json:"fail_action,omitempty" jsonschema:"Optional new failure action"`
	RepairAction string `json:"repair_action,omitempty" jsonschema:"Optional new repair action"`
	AIMode       string `json:"ai_mode,omitempty" jsonschema:"Optional new AI mode"`
	VectorCols   string `json:"vector_cols,omitempty" jsonschema:"Optional new AI vector columns"`
	MqttURL      string `json:"mqtt_url,omitempty" jsonschema:"Optional new MQTT broker URL"`
	MqttTopic    string `json:"mqtt_topic,omitempty" jsonschema:"Optional new MQTT subscription topic"`
	MqttCols     string `json:"mqtt_cols,omitempty" jsonschema:"Optional new MQTT columns"`
}

func updatePollingTool(ctx context.Context, req *mcp.CallToolRequest, args updatePollingParams) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(args.ID)
	if id == "" {
		return nil, nil, fmt.Errorf("polling id must not be empty")
	}
	p := datastore.GetPolling(id)
	if p == nil {
		return nil, nil, fmt.Errorf("polling not found: %s", id)
	}
	// Stop active push polling subscriptions before mutating type/mode/topic.
	if hooks := pollingHooks.Load(); hooks != nil && hooks.StopPolling != nil {
		if (p.Type == "gnmi" && p.Mode == "subscribe") || p.Type == "mqtt" {
			hooks.StopPolling(p)
			time.Sleep(20 * time.Millisecond)
		}
	}
	if args.NodeID != "" {
		n, err := resolveNode(args.NodeID, true, true)
		if err != nil {
			return nil, nil, err
		}
		if p.NodeID != n.ID {
			if p.NodeID != "" {
				datastore.SetNodeStateChanged(p.NodeID)
			}
			datastore.SetNodeStateChanged(n.ID)
			p.NodeID = n.ID
		}
	}
	if args.Name != "" {
		p.Name = strings.TrimSpace(args.Name)
	}
	if args.Type != "" {
		p.Type = strings.TrimSpace(args.Type)
	}
	if args.Mode != "" {
		p.Mode = args.Mode
	}
	if args.Params != "" {
		p.Params = args.Params
	}
	if args.Filter != "" {
		p.Filter = args.Filter
	}
	if args.Extractor != "" {
		p.Extractor = args.Extractor
	}
	if args.Script != "" {
		p.Script = args.Script
	}
	if args.Level != "" {
		level, err := validatePollingLevel(args.Level)
		if err != nil {
			return nil, nil, err
		}
		p.Level = level
	}
	if args.PollInt > 0 {
		p.PollInt = args.PollInt
	}
	if args.Timeout > 0 {
		p.Timeout = args.Timeout
	}
	if args.Retry < 0 {
		p.Retry = 0
	} else if args.Retry > 0 {
		p.Retry = args.Retry
	}
	if args.LogMode != "" {
		lm, err := parseLogModeInt(args.LogMode)
		if err != nil {
			return nil, nil, err
		}
		p.LogMode = lm
	}
	if args.FailAction != "" {
		p.FailAction = args.FailAction
	}
	if args.RepairAction != "" {
		p.RepairAction = args.RepairAction
	}
	if args.AIMode != "" {
		p.AIMode = args.AIMode
	}
	if args.VectorCols != "" {
		p.VectorCols = args.VectorCols
	}
	if args.MqttURL != "" {
		p.MqttURL = args.MqttURL
	}
	if args.MqttTopic != "" {
		p.MqttTopic = args.MqttTopic
	}
	if args.MqttCols != "" {
		p.MqttCols = args.MqttCols
	}
	p.State = "unknown"
	p.NextTime = 0
	if err := datastore.UpdatePolling(p, true); err != nil {
		return nil, nil, err
	}
	datastore.SetNodeStateChanged(p.NodeID)
	nodeName := ""
	if n := datastore.GetNode(p.NodeID); n != nil {
		nodeName = n.Name
	}
	addMCPEventLog("info", p.NodeID, nodeName, fmt.Sprintf("Update polling via MCP: %s (%s)", p.Name, p.Type))
	return mcpJSONResult(toMCPPollingDetail(p, nodeName))
}

type deletePollingParams struct {
	ID string `json:"id" jsonschema:"Polling ID to delete"`
}

func deletePollingTool(ctx context.Context, req *mcp.CallToolRequest, args deletePollingParams) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(args.ID)
	if id == "" {
		return nil, nil, fmt.Errorf("polling id must not be empty")
	}
	p := datastore.GetPolling(id)
	if p == nil {
		return nil, nil, fmt.Errorf("polling not found: %s", id)
	}
	if hooks := pollingHooks.Load(); hooks != nil && hooks.StopPolling != nil {
		hooks.StopPolling(p)
	}
	if err := datastore.DeletePollings([]string{id}); err != nil {
		return nil, nil, err
	}
	nodeName := ""
	if n := datastore.GetNode(p.NodeID); n != nil {
		nodeName = n.Name
	}
	addMCPEventLog("info", p.NodeID, nodeName, fmt.Sprintf("Delete polling via MCP: %s (%s)", p.Name, p.ID))
	return mcpJSONResult(&struct {
		Deleted bool   `json:"Deleted"`
		ID      string `json:"ID"`
		Name    string `json:"Name"`
		NodeID  string `json:"NodeID"`
	}{
		Deleted: true,
		ID:      p.ID,
		Name:    p.Name,
		NodeID:  p.NodeID,
	})
}

type checkPollingParams struct {
	PollingID string `json:"polling_id,omitempty" jsonschema:"Specific polling ID to run immediately"`
	NodeID    string `json:"node_id,omitempty" jsonschema:"Node ID, unique node name, or IP to run all pollings for that node"`
	All       bool   `json:"all,omitempty" jsonschema:"Set true to run all pollings and network checks immediately"`
}

func checkPollingTool(ctx context.Context, req *mcp.CallToolRequest, args checkPollingParams) (*mcp.CallToolResult, any, error) {
	hooks := pollingHooks.Load()
	switch {
	case strings.TrimSpace(args.PollingID) != "":
		p := datastore.GetPolling(strings.TrimSpace(args.PollingID))
		if p == nil {
			return nil, nil, fmt.Errorf("polling not found: %s", args.PollingID)
		}
		status := "scheduled"
		engineActive := false
		if hooks != nil && hooks.CheckPolling != nil {
			engineActive = true
			if !hooks.CheckPolling(p.ID) {
				status = "already_queued_or_skipped"
			} else {
				status = "triggered"
			}
		} else {
			p.NextTime = 0
			_ = datastore.UpdatePolling(p, false)
		}
		nodeName := ""
		if n := datastore.GetNode(p.NodeID); n != nil {
			nodeName = n.Name
		}
		addMCPEventLog("info", p.NodeID, nodeName, fmt.Sprintf("Check polling via MCP: %s", p.Name))
		return mcpJSONResult(&struct {
			Status              string `json:"Status"`
			Scope               string `json:"Scope"`
			PollingID           string `json:"PollingID"`
			PollingName         string `json:"PollingName"`
			NodeID              string `json:"NodeID"`
			PollingEngineActive bool   `json:"PollingEngineActive"`
		}{
			Status:              status,
			Scope:               "polling",
			PollingID:           p.ID,
			PollingName:         p.Name,
			NodeID:              p.NodeID,
			PollingEngineActive: engineActive,
		})
	case strings.TrimSpace(args.NodeID) != "":
		n, err := resolveNode(args.NodeID, true, true)
		if err != nil {
			return nil, nil, err
		}
		status := "scheduled"
		engineActive := false
		if hooks != nil && hooks.CheckNode != nil {
			engineActive = true
			hooks.CheckNode(n.ID)
			UpdateLineState()
			status = "triggered"
		} else {
			datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
				if p.NodeID == n.ID {
					p.NextTime = 0
					_ = datastore.UpdatePolling(p, false)
				}
				return true
			})
		}
		addMCPEventLog("info", n.ID, n.Name, fmt.Sprintf("Check node pollings via MCP: %s", n.Name))
		return mcpJSONResult(&struct {
			Status              string `json:"Status"`
			Scope               string `json:"Scope"`
			NodeID              string `json:"NodeID"`
			NodeName            string `json:"NodeName"`
			PollingEngineActive bool   `json:"PollingEngineActive"`
		}{
			Status:              status,
			Scope:               "node",
			NodeID:              n.ID,
			NodeName:            n.Name,
			PollingEngineActive: engineActive,
		})
	case args.All:
		status := "scheduled"
		engineActive := false
		if hooks != nil && hooks.CheckAll != nil {
			engineActive = true
			hooks.CheckAll()
			CheckAllNetworks()
			UpdateLineState()
			status = "triggered"
		} else {
			datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
				p.NextTime = 0
				_ = datastore.UpdatePolling(p, false)
				return true
			})
		}
		addMCPEventLog("info", "", "", "Check all pollings via MCP")
		return mcpJSONResult(&struct {
			Status              string `json:"Status"`
			Scope               string `json:"Scope"`
			PollingEngineActive bool   `json:"PollingEngineActive"`
		}{
			Status:              status,
			Scope:               "all",
			PollingEngineActive: engineActive,
		})
	default:
		return nil, nil, fmt.Errorf("one of polling_id, node_id, or all=true is required")
	}
}

// Line CRUD helpers & tools
type mcpLineEnt struct {
	ID         string `json:"ID"`
	NodeID1    string `json:"NodeID1"`
	NodeName1  string `json:"NodeName1"`
	PollingID1 string `json:"PollingID1"`
	State1     string `json:"State1"`
	NodeID2    string `json:"NodeID2"`
	NodeName2  string `json:"NodeName2"`
	PollingID2 string `json:"PollingID2"`
	State2     string `json:"State2"`
	PollingID  string `json:"PollingID"`
	State      string `json:"State"`
	Width      int    `json:"Width"`
	Port       string `json:"Port"`
	Info       string `json:"Info"`
}

func resolveLineEndpointDisplay(id string) string {
	if strings.HasPrefix(id, "NET:") {
		if n := datastore.GetNetwork(id); n != nil {
			return n.Name
		}
		return id
	}
	if n := datastore.GetNode(id); n != nil {
		return n.Name
	}
	return ""
}

func toMCPLineEnt(l *datastore.LineEnt) mcpLineEnt {
	return mcpLineEnt{
		ID:         l.ID,
		NodeID1:    l.NodeID1,
		NodeName1:  resolveLineEndpointDisplay(l.NodeID1),
		PollingID1: l.PollingID1,
		State1:     l.State1,
		NodeID2:    l.NodeID2,
		NodeName2:  resolveLineEndpointDisplay(l.NodeID2),
		PollingID2: l.PollingID2,
		State2:     l.State2,
		PollingID:  l.PollingID,
		State:      l.State,
		Width:      l.Width,
		Port:       l.Port,
		Info:       l.Info,
	}
}

func resolveLineEndpointRef(ref string) (string, string, bool, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", false, fmt.Errorf("endpoint reference must not be empty")
	}
	if strings.HasPrefix(ref, "NET:") {
		netEnt, err := resolveNetwork(ref)
		if err != nil {
			return "", "", false, err
		}
		return "NET:" + netEnt.ID, netEnt.Name, true, nil
	}
	if n, err := resolveNode(ref, true, true); err == nil && n != nil {
		return n.ID, n.Name, false, nil
	}
	if netEnt, err := resolveNetwork(ref); err == nil && netEnt != nil {
		return "NET:" + netEnt.ID, netEnt.Name, true, nil
	}
	return "", "", false, fmt.Errorf("node or network not found: %s", ref)
}

func validateEndpointPolling(endpointID, pollingID string, isNet bool) error {
	if pollingID == "" {
		return nil
	}
	if isNet {
		n := datastore.GetNetwork(endpointID)
		if n == nil {
			return fmt.Errorf("network not found: %s", endpointID)
		}
		for _, port := range n.Ports {
			if port.ID == pollingID {
				return nil
			}
		}
		return fmt.Errorf("port %q not found on network %s", pollingID, endpointID)
	}
	p := datastore.GetPolling(pollingID)
	if p == nil {
		return fmt.Errorf("polling not found: %s", pollingID)
	}
	if p.NodeID != endpointID {
		return fmt.Errorf("polling %s does not belong to node %s", pollingID, endpointID)
	}
	return nil
}

func combineLineState(s1, s2 string) string {
	rank := map[string]int{
		"high":    5,
		"low":     4,
		"warn":    3,
		"repair":  2,
		"normal":  1,
		"unknown": 0,
	}
	if s1 == "" {
		s1 = "unknown"
	}
	if s2 == "" {
		s2 = "unknown"
	}
	if rank[s1] >= rank[s2] {
		return s1
	}
	return s2
}

type getLineListParams struct {
	NodeID      string `json:"node_id,omitempty" jsonschema:"Optional node or network ID/name/IP filter"`
	StateFilter string `json:"state_filter,omitempty" jsonschema:"Optional regex filter for overall line state"`
}

func getLineListTool(ctx context.Context, req *mcp.CallToolRequest, args getLineListParams) (*mcp.CallToolResult, any, error) {
	stateRe, err := compileRegexFilter("state_filter", args.StateFilter)
	if err != nil {
		return nil, nil, err
	}
	targetEndpoint := ""
	if strings.TrimSpace(args.NodeID) != "" {
		ep, _, _, err := resolveLineEndpointRef(args.NodeID)
		if err != nil {
			return nil, nil, err
		}
		targetEndpoint = ep
	}
	list := []mcpLineEnt{}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if targetEndpoint != "" && l.NodeID1 != targetEndpoint && l.NodeID2 != targetEndpoint {
			return true
		}
		if stateRe != nil && !stateRe.MatchString(l.State) {
			return true
		}
		list = append(list, toMCPLineEnt(l))
		return true
	})
	sort.Slice(list, func(i, j int) bool {
		return list[i].ID < list[j].ID
	})
	return mcpJSONResult(&list)
}

type addLineParams struct {
	NodeID1    string `json:"node_id1" jsonschema:"First endpoint node ID/name/IP or network ID/name (e.g. NET:<id>)"`
	NodeID2    string `json:"node_id2" jsonschema:"Second endpoint node ID/name/IP or network ID/name (e.g. NET:<id>)"`
	PollingID1 string `json:"polling_id1,omitempty" jsonschema:"Optional polling ID on endpoint 1 (or port ID if endpoint 1 is a network)"`
	PollingID2 string `json:"polling_id2,omitempty" jsonschema:"Optional polling ID on endpoint 2 (or port ID if endpoint 2 is a network)"`
	PollingID  string `json:"polling_id,omitempty" jsonschema:"Optional line traffic polling ID"`
	Width      int    `json:"width,omitempty" jsonschema:"Line width (1-10, default 2)"`
	Port       string `json:"port,omitempty" jsonschema:"Port label shown on the line"`
	Info       string `json:"info,omitempty" jsonschema:"Additional line info text"`
}

func addLineTool(ctx context.Context, req *mcp.CallToolRequest, args addLineParams) (*mcp.CallToolResult, any, error) {
	ep1, name1, isNet1, err := resolveLineEndpointRef(args.NodeID1)
	if err != nil {
		return nil, nil, err
	}
	ep2, name2, isNet2, err := resolveLineEndpointRef(args.NodeID2)
	if err != nil {
		return nil, nil, err
	}
	if ep1 == ep2 && args.PollingID1 == args.PollingID2 {
		return nil, nil, fmt.Errorf("cannot connect an endpoint to itself with the same polling/port")
	}
	if err := validateEndpointPolling(ep1, args.PollingID1, isNet1); err != nil {
		return nil, nil, err
	}
	if err := validateEndpointPolling(ep2, args.PollingID2, isNet2); err != nil {
		return nil, nil, err
	}
	if args.PollingID != "" && datastore.GetPolling(args.PollingID) == nil {
		return nil, nil, fmt.Errorf("line traffic polling not found: %s", args.PollingID)
	}
	s1, ok1 := getLineEndState(ep1, args.PollingID1)
	if !ok1 && isNet1 && args.PollingID1 != "" {
		return nil, nil, fmt.Errorf("network port %s not found on %s", args.PollingID1, ep1)
	}
	s2, ok2 := getLineEndState(ep2, args.PollingID2)
	if !ok2 && isNet2 && args.PollingID2 != "" {
		return nil, nil, fmt.Errorf("network port %s not found on %s", args.PollingID2, ep2)
	}
	l := &datastore.LineEnt{
		NodeID1:    ep1,
		PollingID1: args.PollingID1,
		State1:     s1,
		NodeID2:    ep2,
		PollingID2: args.PollingID2,
		State2:     s2,
		PollingID:  args.PollingID,
		Width:      clampLimit(args.Width, 1, 10, 2),
		Port:       args.Port,
		Info:       args.Info,
		State:      combineLineState(s1, s2),
	}
	if datastore.HasLine(l, isNet1 || isNet2) {
		return nil, nil, fmt.Errorf("line already exists between %s and %s", ep1, ep2)
	}
	if err := datastore.AddLine(l); err != nil {
		return nil, nil, err
	}
	UpdateLineState()
	eventNodeID := ep1
	if isNet1 && !isNet2 {
		eventNodeID = ep2
	} else if isNet1 && isNet2 {
		eventNodeID = ""
	}
	addMCPEventLog("info", eventNodeID, name1, fmt.Sprintf("Add line via MCP: %s <-> %s", name1, name2))
	res := toMCPLineEnt(l)
	return mcpJSONResult(&res)
}

type deleteLineParams struct {
	ID string `json:"id" jsonschema:"Line ID to delete"`
}

func deleteLineTool(ctx context.Context, req *mcp.CallToolRequest, args deleteLineParams) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(args.ID)
	if id == "" {
		return nil, nil, fmt.Errorf("line id must not be empty")
	}
	l := datastore.GetLine(id)
	if l == nil {
		return nil, nil, fmt.Errorf("line not found: %s", id)
	}
	if err := datastore.DeleteLine(id); err != nil {
		return nil, nil, err
	}
	addMCPEventLog("info", "", "", fmt.Sprintf("Delete line via MCP: %s (%s <-> %s)", id, resolveLineEndpointDisplay(l.NodeID1), resolveLineEndpointDisplay(l.NodeID2)))
	return mcpJSONResult(&struct {
		Deleted bool   `json:"Deleted"`
		ID      string `json:"ID"`
		NodeID1 string `json:"NodeID1"`
		NodeID2 string `json:"NodeID2"`
	}{
		Deleted: true,
		ID:      id,
		NodeID1: l.NodeID1,
		NodeID2: l.NodeID2,
	})
}

// Network CRUD tools
type mcpNetworkDetailEnt struct {
	ID          string              `json:"ID"`
	Name        string              `json:"Name"`
	IP          string              `json:"IP"`
	Description string              `json:"Description"`
	SnmpMode    string              `json:"SnmpMode"`
	SnmpPort    int                 `json:"SnmpPort"`
	Community   string              `json:"Community"`
	User        string              `json:"User"`
	Password    string              `json:"Password"`
	URL         string              `json:"URL"`
	ArpWatch    bool                `json:"ArpWatch"`
	Unmanaged   bool                `json:"Unmanaged"`
	LLDP        bool                `json:"LLDP"`
	HPorts      int                 `json:"HPorts"`
	X           int                 `json:"X"`
	Y           int                 `json:"Y"`
	W           int                 `json:"W"`
	H           int                 `json:"H"`
	SystemID    string              `json:"SystemID"`
	Error       string              `json:"Error"`
	Ports       []datastore.PortEnt `json:"Ports"`
}

func toMCPNetworkDetail(n *datastore.NetworkEnt) *mcpNetworkDetailEnt {
	ports := n.Ports
	if ports == nil {
		ports = []datastore.PortEnt{}
	}
	return &mcpNetworkDetailEnt{
		ID:          n.ID,
		Name:        n.Name,
		IP:          n.IP,
		Description: n.Descr,
		SnmpMode:    n.SnmpMode,
		SnmpPort:    n.SnmpPort,
		Community:   maskSecret(n.Community),
		User:        n.User,
		Password:    maskSecret(n.Password),
		URL:         n.URL,
		ArpWatch:    n.ArpWatch,
		Unmanaged:   n.Unmanaged,
		LLDP:        n.LLDP,
		HPorts:      n.HPorts,
		X:           n.X,
		Y:           n.Y,
		W:           n.W,
		H:           n.H,
		SystemID:    n.SystemID,
		Error:       n.Error,
		Ports:       ports,
	}
}

type getNetworkParams struct {
	ID string `json:"id" jsonschema:"Network ID (with or without NET: prefix), unique name, or unique IP"`
}

func getNetworkTool(ctx context.Context, req *mcp.CallToolRequest, args getNetworkParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNetwork(args.ID)
	if err != nil {
		return nil, nil, err
	}
	return mcpJSONResult(toMCPNetworkDetail(n))
}

type addNetworkParams struct {
	Name        string `json:"name" jsonschema:"Network name"`
	IP          string `json:"ip" jsonschema:"Management IP address"`
	Description string `json:"description,omitempty" jsonschema:"Network description"`
	SnmpMode    string `json:"snmp_mode,omitempty" jsonschema:"SNMP mode (v2c, v3auth, v3authpriv, v3authprivex, etc.)"`
	Community   string `json:"community,omitempty" jsonschema:"SNMP v2c community"`
	User        string `json:"user,omitempty" jsonschema:"SNMP v3 username"`
	Password    string `json:"password,omitempty" jsonschema:"SNMP v3 password"`
	SnmpPort    int    `json:"snmp_port,omitempty" jsonschema:"SNMP UDP port (default 161)"`
	URL         string `json:"url,omitempty" jsonschema:"Management URL"`
	ArpWatch    bool   `json:"arp_watch,omitempty" jsonschema:"Enable ARP watch monitoring"`
	Unmanaged   bool   `json:"unmanaged,omitempty" jsonschema:"Mark network as unmanaged (skip SNMP checks)"`
	LLDP        bool   `json:"lldp,omitempty" jsonschema:"Enable LLDP topology discovery"`
	HPorts      int    `json:"h_ports,omitempty" jsonschema:"Horizontal port layout count (default 24)"`
	X           int    `json:"x,omitempty" jsonschema:"Map X coordinate (1-2000, default 64)"`
	Y           int    `json:"y,omitempty" jsonschema:"Map Y coordinate (1-2000, default 64)"`
}

func addNetworkTool(ctx context.Context, req *mcp.CallToolRequest, args addNetworkParams) (*mcp.CallToolResult, any, error) {
	name := strings.TrimSpace(args.Name)
	if name == "" {
		return nil, nil, fmt.Errorf("network name must not be empty")
	}
	ipStr := strings.TrimSpace(args.IP)
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return nil, nil, fmt.Errorf("invalid IP address: %s", args.IP)
	}
	ipStr = parsedIP.String()
	snmpPort := args.SnmpPort
	if snmpPort <= 0 || snmpPort > 65535 {
		snmpPort = 161
	}
	x := args.X
	if x < 1 || x > 2000 {
		x = 64
	}
	y := args.Y
	if y < 1 || y > 2000 {
		y = 64
	}
	hPorts := args.HPorts
	if hPorts <= 0 {
		hPorts = 24
	}
	n := &datastore.NetworkEnt{
		Name:      name,
		IP:        ipStr,
		Descr:     args.Description,
		SnmpMode:  args.SnmpMode,
		Community: args.Community,
		User:      args.User,
		Password:  args.Password,
		SnmpPort:  snmpPort,
		URL:       args.URL,
		ArpWatch:  args.ArpWatch,
		Unmanaged: args.Unmanaged,
		LLDP:      args.LLDP,
		HPorts:    hPorts,
		X:         x,
		Y:         y,
		SystemID:  ipStr,
	}
	if err := datastore.AddNetwork(n); err != nil {
		return nil, nil, err
	}
	if !n.Unmanaged {
		CheckNetwork(n.ID)
	}
	addMCPEventLog("info", "", n.Name, fmt.Sprintf("Add network via MCP: %s (%s)", n.Name, n.IP))
	return mcpJSONResult(toMCPNetworkDetail(n))
}

func parseOptionalBoolString(field, val string) (set bool, b bool, err error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return false, false, nil
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false, false, fmt.Errorf("invalid boolean value %q for %s (use true or false)", val, field)
	}
	return true, parsed, nil
}

type updateNetworkParams struct {
	ID          string `json:"id" jsonschema:"Network ID (with or without NET: prefix), unique name, or unique IP"`
	Name        string `json:"name,omitempty" jsonschema:"Optional new network name"`
	IP          string `json:"ip,omitempty" jsonschema:"Optional new management IP address"`
	Description string `json:"description,omitempty" jsonschema:"Optional new description"`
	SnmpMode    string `json:"snmp_mode,omitempty" jsonschema:"Optional new SNMP mode"`
	Community   string `json:"community,omitempty" jsonschema:"Optional new SNMP v2c community"`
	User        string `json:"user,omitempty" jsonschema:"Optional new SNMP v3 username"`
	Password    string `json:"password,omitempty" jsonschema:"Optional new SNMP v3 password"`
	SnmpPort    int    `json:"snmp_port,omitempty" jsonschema:"Optional new SNMP UDP port (1-65535)"`
	URL         string `json:"url,omitempty" jsonschema:"Optional new management URL"`
	ArpWatch    string `json:"arp_watch,omitempty" jsonschema:"Set ARP watch: 'true', 'false', or empty to keep unchanged"`
	Unmanaged   string `json:"unmanaged,omitempty" jsonschema:"Set unmanaged: 'true', 'false', or empty to keep unchanged"`
	LLDP        string `json:"lldp,omitempty" jsonschema:"Set LLDP: 'true', 'false', or empty to keep unchanged"`
	HPorts      int    `json:"h_ports,omitempty" jsonschema:"Optional new horizontal port count (>0)"`
	X           int    `json:"x,omitempty" jsonschema:"Optional new X position (>0)"`
	Y           int    `json:"y,omitempty" jsonschema:"Optional new Y position (>0)"`
}

func updateNetworkTool(ctx context.Context, req *mcp.CallToolRequest, args updateNetworkParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNetwork(args.ID)
	if err != nil {
		return nil, nil, err
	}
	if setArp, bArp, err := parseOptionalBoolString("arp_watch", args.ArpWatch); err != nil {
		return nil, nil, err
	} else if setArp {
		n.ArpWatch = bArp
	}
	if setUnm, bUnm, err := parseOptionalBoolString("unmanaged", args.Unmanaged); err != nil {
		return nil, nil, err
	} else if setUnm {
		n.Unmanaged = bUnm
	}
	if setLLDP, bLLDP, err := parseOptionalBoolString("lldp", args.LLDP); err != nil {
		return nil, nil, err
	} else if setLLDP {
		n.LLDP = bLLDP
	}
	if args.Name != "" {
		n.Name = strings.TrimSpace(args.Name)
	}
	if args.IP != "" {
		parsed := net.ParseIP(strings.TrimSpace(args.IP))
		if parsed == nil {
			return nil, nil, fmt.Errorf("invalid IP address: %s", args.IP)
		}
		n.IP = parsed.String()
	}
	if args.Description != "" {
		n.Descr = args.Description
	}
	if args.SnmpMode != "" {
		n.SnmpMode = args.SnmpMode
	}
	if args.Community != "" {
		n.Community = args.Community
	}
	if args.User != "" {
		n.User = args.User
	}
	if args.Password != "" {
		n.Password = args.Password
	}
	if args.SnmpPort > 0 && args.SnmpPort <= 65535 {
		n.SnmpPort = args.SnmpPort
	}
	if args.URL != "" {
		n.URL = args.URL
	}
	if args.HPorts > 0 {
		n.HPorts = args.HPorts
	}
	if args.X > 0 && args.X <= 2000 {
		n.X = args.X
	}
	if args.Y > 0 && args.Y <= 2000 {
		n.Y = args.Y
	}
	if err := datastore.UpdateNetwork(n); err != nil {
		return nil, nil, err
	}
	if !n.Unmanaged {
		CheckNetwork(n.ID)
	}
	addMCPEventLog("info", "", n.Name, fmt.Sprintf("Update network via MCP: %s (%s)", n.Name, n.IP))
	return mcpJSONResult(toMCPNetworkDetail(n))
}

type deleteNetworkParams struct {
	ID string `json:"id" jsonschema:"Network ID (with or without NET: prefix), unique name, or unique IP"`
}

func deleteNetworkTool(ctx context.Context, req *mcp.CallToolRequest, args deleteNetworkParams) (*mcp.CallToolResult, any, error) {
	n, err := resolveNetwork(args.ID)
	if err != nil {
		return nil, nil, err
	}
	netRef := "NET:" + n.ID
	lineIDs := []string{}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 == netRef || l.NodeID2 == netRef {
			lineIDs = append(lineIDs, l.ID)
		}
		return true
	})
	deletedLines := 0
	for _, lid := range lineIDs {
		if err := datastore.DeleteLine(lid); err == nil {
			deletedLines++
		}
	}
	if err := datastore.DeleteNetwork(n.ID); err != nil {
		return nil, nil, err
	}
	addMCPEventLog("info", "", n.Name, fmt.Sprintf("Delete network via MCP: %s (%s)", n.Name, n.ID))
	return mcpJSONResult(&struct {
		Deleted      bool   `json:"Deleted"`
		ID           string `json:"ID"`
		Name         string `json:"Name"`
		DeletedLines int    `json:"DeletedLines"`
	}{
		Deleted:      true,
		ID:           n.ID,
		Name:         n.Name,
		DeletedLines: deletedLines,
	})
}

type checkNetworkParams struct {
	ID  string `json:"id,omitempty" jsonschema:"Network ID (with or without NET: prefix), unique name, or unique IP"`
	All bool   `json:"all,omitempty" jsonschema:"Set true to check all network containers"`
}

func checkNetworkTool(ctx context.Context, req *mcp.CallToolRequest, args checkNetworkParams) (*mcp.CallToolResult, any, error) {
	if args.All {
		CheckAllNetworks()
		addMCPEventLog("info", "", "", "Check all networks via MCP")
		return mcpJSONResult(&struct {
			Status string `json:"Status"`
			Scope  string `json:"Scope"`
		}{
			Status: "triggered",
			Scope:  "all",
		})
	}
	if strings.TrimSpace(args.ID) == "" {
		return nil, nil, fmt.Errorf("id or all=true is required")
	}
	n, err := resolveNetwork(args.ID)
	if err != nil {
		return nil, nil, err
	}
	CheckNetwork(n.ID)
	addMCPEventLog("info", "", n.Name, fmt.Sprintf("Check network via MCP: %s", n.Name))
	return mcpJSONResult(&struct {
		Status string `json:"Status"`
		Scope  string `json:"Scope"`
		ID     string `json:"ID"`
		Name   string `json:"Name"`
	}{
		Status: "triggered",
		Scope:  "network",
		ID:     n.ID,
		Name:   n.Name,
	})
}

// Server certificate monitor CRUD tools
type updateServerCertificateParams struct {
	ID     string `json:"id,omitempty" jsonschema:"Existing certificate monitor ID (leave empty to create a new entry)"`
	Target string `json:"target,omitempty" jsonschema:"Target hostname or IP address (required when adding new entry)"`
	Port   int    `json:"port,omitempty" jsonschema:"TLS port (1-65535, default 443 when adding new entry)"`
}

func updateServerCertificateTool(ctx context.Context, req *mcp.CallToolRequest, args updateServerCertificateParams) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(args.ID)
	target := strings.TrimSpace(args.Target)
	if args.Port < 0 || args.Port > 65535 {
		return nil, nil, fmt.Errorf("invalid port %d (must be 1-65535)", args.Port)
	}
	var c *datastore.CertMonitorEnt
	if id != "" {
		c = datastore.GetCertMonitor(id)
		if c == nil || c.ID == "" {
			return nil, nil, fmt.Errorf("server certificate monitor not found: %s", id)
		}
		if target == "" {
			target = c.Target
		}
		port := c.Port
		if args.Port > 0 {
			port = uint16(args.Port)
		}
		if c.Target != target || c.Port != port {
			c.Subject = ""
			c.Issuer = ""
			c.SerialNumber = ""
			c.State = "unknown"
			c.NotAfter = 0
			c.NotBefore = 0
			c.FirstTime = 0
			c.LastTime = 0
			c.Verify = false
			c.Error = ""
		}
		c.Target = target
		c.Port = port
	} else {
		if target == "" {
			return nil, nil, fmt.Errorf("target is required when creating a new server certificate monitor")
		}
		port := uint16(443)
		if args.Port > 0 {
			port = uint16(args.Port)
		}
		c = &datastore.CertMonitorEnt{
			State:  "unknown",
			Target: target,
			Port:   port,
		}
	}
	if err := datastore.SaveCertMonitor(c); err != nil {
		return nil, nil, err
	}
	DoCehckCertMonitor()
	addMCPEventLog("info", "", c.Target, fmt.Sprintf("Save server certificate monitor via MCP: %s:%d", c.Target, c.Port))
	return mcpJSONResult(&mcpServerCertificateEnt{
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
}

type deleteServerCertificateParams struct {
	ID string `json:"id" jsonschema:"Server certificate monitor ID to delete"`
}

func deleteServerCertificateTool(ctx context.Context, req *mcp.CallToolRequest, args deleteServerCertificateParams) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(args.ID)
	if id == "" {
		return nil, nil, fmt.Errorf("certificate monitor id must not be empty")
	}
	c := datastore.GetCertMonitor(id)
	if c == nil || c.ID == "" {
		return nil, nil, fmt.Errorf("server certificate monitor not found: %s", id)
	}
	datastore.DeleteCertMonitor(id)
	addMCPEventLog("info", "", c.Target, fmt.Sprintf("Delete server certificate monitor via MCP: %s (%s:%d)", id, c.Target, c.Port))
	return mcpJSONResult(&struct {
		Deleted bool   `json:"Deleted"`
		ID      string `json:"ID"`
		Target  string `json:"Target"`
		Port    uint16 `json:"Port"`
	}{
		Deleted: true,
		ID:      c.ID,
		Target:  c.Target,
		Port:    c.Port,
	})
}

// get_system_status tool
type getSystemStatusParams struct{}

type mcpReceiverStatus struct {
	Enabled bool   `json:"Enabled"`
	Status  string `json:"Status"`
	Port    string `json:"Port"`
	Range   string `json:"Range,omitempty"`
	Error   string `json:"Error,omitempty"`
}

type mcpSystemMonitorSnapshot struct {
	Time         string  `json:"Time"`
	CPU          float64 `json:"CPU"`
	Mem          float64 `json:"Mem"`
	MyCPU        float64 `json:"MyCPU"`
	MyMem        float64 `json:"MyMem"`
	Swap         float64 `json:"Swap"`
	Disk         float64 `json:"Disk"`
	Load         float64 `json:"Load"`
	Bytes        float64 `json:"Bytes"`
	Net          float64 `json:"Net"`
	Conn         int     `json:"Conn"`
	Proc         int     `json:"Proc"`
	DBSize       int64   `json:"DBSize"`
	HeapAlloc    int64   `json:"HeapAlloc"`
	Sys          int64   `json:"Sys"`
	NumGoroutine int     `json:"NumGoroutine"`
}

type mcpSystemStatus struct {
	Version            string                       `json:"Version"`
	StartTime          string                       `json:"StartTime"`
	Uptime             string                       `json:"Uptime"`
	NumCPU             int                          `json:"NumCPU"`
	NumGoroutine       int                          `json:"NumGoroutine"`
	DBSize             int64                        `json:"DBSize"`
	DatastorePath      string                       `json:"DatastorePath"`
	PingMode           string                       `json:"PingMode"`
	MCPTransport       string                       `json:"MCPTransport"`
	MCPEndpoint        string                       `json:"MCPEndpoint"`
	MCPStatus          string                       `json:"MCPStatus"`
	MCPError           string                       `json:"MCPError,omitempty"`
	NodeCount          int                          `json:"NodeCount"`
	NodeStateCounts    map[string]int               `json:"NodeStateCounts"`
	PollingCount       int                          `json:"PollingCount"`
	PollingStateCounts map[string]int               `json:"PollingStateCounts"`
	LineCount          int                          `json:"LineCount"`
	NetworkCount       int                          `json:"NetworkCount"`
	CertMonitorCount   int                          `json:"CertMonitorCount"`
	Monitor            *mcpSystemMonitorSnapshot    `json:"Monitor,omitempty"`
	Receivers          map[string]mcpReceiverStatus `json:"Receivers"`
}

func buildReceiverEntry(name string, enabled bool, defaultPort string, defaultRange string) mcpReceiverStatus {
	st, ok := logger.GetReceiverStatus(name)
	res := mcpReceiverStatus{
		Enabled: enabled,
		Status:  st.Status,
		Port:    st.Port,
		Range:   st.Range,
		Error:   st.Error,
	}
	if !ok {
		if enabled {
			res.Status = "configured_not_started"
		} else {
			res.Status = "disabled"
		}
	}
	if res.Port == "" {
		res.Port = defaultPort
	}
	if res.Range == "" {
		res.Range = defaultRange
	}
	return res
}

func getSystemStatusTool(ctx context.Context, req *mcp.CallToolRequest, _ getSystemStatusParams) (*mcp.CallToolResult, any, error) {
	nodeStates := map[string]int{}
	nodeCount := 0
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		nodeCount++
		state := n.State
		if state == "" {
			state = "unknown"
		}
		nodeStates[state]++
		return true
	})

	pollingStates := map[string]int{}
	pollingCount := 0
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		pollingCount++
		state := p.State
		if state == "" {
			state = "unknown"
		}
		pollingStates[state]++
		return true
	})

	lineCount := 0
	datastore.ForEachLines(func(_ *datastore.LineEnt) bool {
		lineCount++
		return true
	})

	networkCount := 0
	datastore.ForEachNetworks(func(_ *datastore.NetworkEnt) bool {
		networkCount++
		return true
	})

	certCount := 0
	datastore.ForEachCertMonitors(func(_ *datastore.CertMonitorEnt) bool {
		certCount++
		return true
	})

	var monitorSnap *mcpSystemMonitorSnapshot
	if m := GetLatestMonitorData(); m != nil {
		monitorSnap = &mcpSystemMonitorSnapshot{
			Time:         formatTimeUnixNano(m.Time),
			CPU:          m.CPU,
			Mem:          m.Mem,
			MyCPU:        m.MyCPU,
			MyMem:        m.MyMem,
			Swap:         m.Swap,
			Disk:         m.Disk,
			Load:         m.Load,
			Bytes:        m.Bytes,
			Net:          m.Net,
			Conn:         m.Conn,
			Proc:         m.Proc,
			DBSize:       m.DBSize,
			HeapAlloc:    m.HeapAlloc,
			Sys:          m.Sys,
			NumGoroutine: m.NumGoroutine,
		}
	}

	mcpStatus, mcpErr := GetMCPStatus()
	receivers := map[string]mcpReceiverStatus{
		"syslogd":  buildReceiverEntry("syslogd", datastore.MapConf.EnableSyslogd, fmt.Sprintf("UDP:%d", datastore.SyslogPort), ""),
		"trapd":    buildReceiverEntry("trapd", datastore.MapConf.EnableTrapd, fmt.Sprintf("UDP:%d", datastore.TrapPort), ""),
		"netflowd": buildReceiverEntry("netflowd", datastore.MapConf.EnableNetflowd, fmt.Sprintf("UDP:%d", datastore.NetFlowPort), ""),
		"sflowd":   buildReceiverEntry("sflowd", datastore.MapConf.EnableSFlowd, fmt.Sprintf("UDP:%d", datastore.SFlowPort), ""),
		"otel":     buildReceiverEntry("otel", datastore.MapConf.EnableOTel, fmt.Sprintf("gRPC:%d,HTTP:%d", datastore.OTelgRPCPort, datastore.OTelHTTPPort), datastore.MapConf.OTelFrom),
		"mqtt":     buildReceiverEntry("mqtt", datastore.MapConf.EnableMqtt, fmt.Sprintf("TCP:%d,WS:%d", datastore.MqttTCPPort, datastore.MqttWSPort), ""),
		"arpwatch": buildReceiverEntry("arpwatch", datastore.MapConf.EnableArpWatch, "", datastore.MapConf.ArpWatchRange),
		"sshd":     buildReceiverEntry("sshd", datastore.MapConf.EnableSshd, fmt.Sprintf("TCP:%d", datastore.SSHdPort), ""),
		"tcpd":     buildReceiverEntry("tcpd", datastore.MapConf.EnableTcpd, fmt.Sprintf("TCP:%d", datastore.TCPPort), ""),
	}

	uptime := ""
	startTimeStr := ""
	if !backendStartTime.IsZero() {
		startTimeStr = backendStartTime.Format(time.RFC3339)
		uptime = time.Since(backendStartTime).Round(time.Second).String()
	}

	res := mcpSystemStatus{
		Version:            version,
		StartTime:          startTimeStr,
		Uptime:             uptime,
		NumCPU:             runtime.NumCPU(),
		NumGoroutine:       runtime.NumGoroutine(),
		DBSize:             datastore.GetDBSize(),
		DatastorePath:      dspath,
		PingMode:           ping.GetPingMode(),
		MCPTransport:       datastore.MapConf.MCPTransport,
		MCPEndpoint:        datastore.MapConf.MCPEndpoint,
		MCPStatus:          mcpStatus,
		MCPError:           mcpErr,
		NodeCount:          nodeCount,
		NodeStateCounts:    nodeStates,
		PollingCount:       pollingCount,
		PollingStateCounts: pollingStates,
		LineCount:          lineCount,
		NetworkCount:       networkCount,
		CertMonitorCount:   certCount,
		Monitor:            monitorSnap,
		Receivers:          receivers,
	}
	return mcpJSONResult(&res)
}
