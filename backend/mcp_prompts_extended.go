package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type promptArgSpec struct {
	Key      string
	Label    string
	Desc     string
	Required bool
}

type promptSpec struct {
	Name        string
	Title       string
	Description string
	Instruction string
	Args        []promptArgSpec
}

func formatOptionalPromptArgs(args map[string]string, pairs [][2]string) []string {
	lines := []string{}
	for _, p := range pairs {
		if val, ok := args[p[0]]; ok && strings.TrimSpace(val) != "" {
			lines = append(lines, fmt.Sprintf("- %s: %s", p[1], val))
		}
	}
	return lines
}

func registerSpecPrompt(s *mcp.Server, spec promptSpec) {
	pArgs := make([]*mcp.PromptArgument, 0, len(spec.Args))
	for _, a := range spec.Args {
		pArgs = append(pArgs, &mcp.PromptArgument{
			Name:        a.Key,
			Title:       a.Label,
			Description: a.Desc,
			Required:    a.Required,
		})
	}
	s.AddPrompt(&mcp.Prompt{
		Name:        spec.Name,
		Title:       spec.Title,
		Description: spec.Description,
		Arguments:   pArgs,
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		lines := []string{}
		for _, a := range spec.Args {
			val := ""
			if req != nil && req.Params != nil && req.Params.Arguments != nil {
				val = strings.TrimSpace(req.Params.Arguments[a.Key])
			}
			if a.Required && val == "" {
				return nil, fmt.Errorf("%s is required", a.Key)
			}
			if val != "" {
				lines = append(lines, fmt.Sprintf("- %s: %s", a.Label, val))
			}
		}
		text := spec.Instruction
		if len(lines) > 0 {
			text += "\n" + strings.Join(lines, "\n")
		}
		return &mcp.GetPromptResult{
			Description: spec.Description,
			Messages: []*mcp.PromptMessage{
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: text},
				},
			},
		}, nil
	})
}

func addExtendedPrompts(s *mcp.Server) {
	specs := []promptSpec{
		{
			Name:        "troubleshoot_network",
			Title:       "Investigate network or node issues",
			Description: "Guided multi-step troubleshooting workflow using TWSNMP tools.",
			Instruction: "Investigate network and monitoring health in TWSNMP by checking get_system_status, abnormal nodes (get_node_list with state_filter), failing pollings (get_polling_list), lines (get_line_list), recent events (search_event_log), syslogs (search_syslog), and live diagnostics (diagnose_node) if a target is given:",
			Args: []promptArgSpec{
				{Key: "target", Label: "Target node or IP", Desc: "Optional target node ID, name, or IP to focus on"},
				{Key: "start_time", Label: "Start time", Desc: "Optional lookback window (default -1h)"},
			},
		},
		{
			Name:        "diagnose_node",
			Title:       "Run live diagnostics on a node or target",
			Description: "Execute multi-layer Ping, SNMP, TCP port, and HTTP diagnostics via diagnose_node.",
			Instruction: "Run live diagnostics using the diagnose_node tool (and inspect node details via get_node if managed):",
			Args: []promptArgSpec{
				{Key: "target", Label: "Target", Desc: "Node ID, name, IP, or hostname", Required: true},
				{Key: "ports", Label: "TCP ports", Desc: "Optional comma-separated TCP ports"},
				{Key: "timeout", Label: "Timeout", Desc: "Optional timeout in seconds (1-10)"},
			},
		},
		{
			Name:        "get_system_status",
			Title:       "Get TWSNMP system status",
			Description: "Inspect TWSNMP runtime, receivers, DB size, and node/polling summaries.",
			Instruction: "Retrieve TWSNMP system status and receiver health by using the get_system_status tool.",
		},
		{
			Name:        "get_node",
			Title:       "Get single node details",
			Description: "Retrieve full node details and pollings via get_node.",
			Instruction: "Get detailed node information using the get_node tool:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Node ID/Name/IP", Desc: "Node ID, unique name, or IP", Required: true},
			},
		},
		{
			Name:        "delete_node",
			Title:       "Delete a node",
			Description: "Delete a node, its pollings, and its lines via delete_node.",
			Instruction: "Verify the target node with get_node first, then delete it using delete_node:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Node ID/Name/IP", Desc: "Node ID, unique name, or IP", Required: true},
			},
		},
		{
			Name:        "get_polling",
			Title:       "Get polling configuration",
			Description: "Get full polling configuration and result via get_polling.",
			Instruction: "Get polling configuration and latest result using get_polling:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Polling ID", Desc: "Polling ID", Required: true},
			},
		},
		{
			Name:        "add_polling",
			Title:       "Add a polling check",
			Description: "Create a new polling on a node via add_polling.",
			Instruction: "Create a new polling check on TWSNMP using add_polling:",
			Args: []promptArgSpec{
				{Key: "node_id", Label: "Node", Desc: "Target node ID, name, or IP", Required: true},
				{Key: "name", Label: "Name", Desc: "Polling name", Required: true},
				{Key: "type", Label: "Type", Desc: "Polling type (ping, tcp, http, dns, snmp, etc.)", Required: true},
				{Key: "params", Label: "Params", Desc: "Optional polling parameter (port, URL, OID)"},
				{Key: "level", Label: "Level", Desc: "Optional severity level (off, low, warn, high)"},
			},
		},
		{
			Name:        "update_polling",
			Title:       "Update a polling configuration",
			Description: "Modify an existing polling via update_polling.",
			Instruction: "Update polling configuration using update_polling:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Polling ID", Desc: "Polling ID to update", Required: true},
				{Key: "level", Label: "Level", Desc: "Optional level (off, low, warn, high)"},
				{Key: "poll_int", Label: "Interval", Desc: "Optional interval in seconds"},
				{Key: "params", Label: "Params", Desc: "Optional parameter override"},
			},
		},
		{
			Name:        "delete_polling",
			Title:       "Delete a polling",
			Description: "Delete a polling check via delete_polling.",
			Instruction: "Delete polling check using delete_polling:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Polling ID", Desc: "Polling ID to delete", Required: true},
			},
		},
		{
			Name:        "check_polling",
			Title:       "Run polling checks immediately",
			Description: "Trigger immediate execution of a polling, node pollings, or all pollings via check_polling.",
			Instruction: "Trigger immediate polling execution using check_polling:",
			Args: []promptArgSpec{
				{Key: "polling_id", Label: "Polling ID", Desc: "Specific polling ID"},
				{Key: "node_id", Label: "Node ID/Name/IP", Desc: "Check all pollings for a node"},
				{Key: "all", Label: "Check all", Desc: "Set true to check all pollings"},
			},
		},
		{
			Name:        "get_polling_log_data",
			Title:       "Export polling log data as CSV",
			Description: "Fetch polling log metrics in CSV format via get_polling_log_data.",
			Instruction: "Retrieve polling log time-series CSV using get_polling_log_data:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Polling ID", Desc: "Polling ID", Required: true},
				{Key: "limit", Label: "Limit", Desc: "Max rows to return"},
				{Key: "start_time", Label: "Start time", Desc: "Optional start time"},
				{Key: "end_time", Label: "End time", Desc: "Optional end time"},
			},
		},
		{
			Name:        "get_line_list",
			Title:       "Get map line list",
			Description: "List map lines between nodes and networks via get_line_list.",
			Instruction: "Get map lines using get_line_list:",
			Args: []promptArgSpec{
				{Key: "node_id", Label: "Node/Network", Desc: "Optional node or network filter"},
				{Key: "state_filter", Label: "State filter", Desc: "Optional state regex"},
			},
		},
		{
			Name:        "add_line",
			Title:       "Add map line",
			Description: "Connect two nodes or networks with a line via add_line.",
			Instruction: "Add a map line using add_line:",
			Args: []promptArgSpec{
				{Key: "node_id1", Label: "Endpoint 1", Desc: "First node or network ID/name/IP", Required: true},
				{Key: "node_id2", Label: "Endpoint 2", Desc: "Second node or network ID/name/IP", Required: true},
				{Key: "polling_id1", Label: "Polling/Port 1", Desc: "Optional polling or port ID"},
				{Key: "polling_id2", Label: "Polling/Port 2", Desc: "Optional polling or port ID"},
			},
		},
		{
			Name:        "delete_line",
			Title:       "Delete map line",
			Description: "Delete a map line by ID via delete_line.",
			Instruction: "Delete a map line using delete_line:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Line ID", Desc: "Line ID to delete", Required: true},
			},
		},
		{
			Name:        "get_network",
			Title:       "Get network container details",
			Description: "Get network container details and port table via get_network.",
			Instruction: "Get detailed network container info using get_network:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Network ID/Name/IP", Desc: "Network ID, name, or IP", Required: true},
			},
		},
		{
			Name:        "add_network",
			Title:       "Add network container",
			Description: "Add a network/switch container via add_network.",
			Instruction: "Add a network container using add_network:",
			Args: []promptArgSpec{
				{Key: "name", Label: "Name", Desc: "Network name", Required: true},
				{Key: "ip", Label: "IP", Desc: "Management IP", Required: true},
				{Key: "unmanaged", Label: "Unmanaged", Desc: "Optional true/false"},
			},
		},
		{
			Name:        "update_network",
			Title:       "Update network container",
			Description: "Update network container fields via update_network.",
			Instruction: "Update a network container using update_network:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Network ID/Name/IP", Desc: "Network ID, name, or IP", Required: true},
				{Key: "name", Label: "Name", Desc: "Optional new name"},
				{Key: "ip", Label: "IP", Desc: "Optional new IP"},
			},
		},
		{
			Name:        "delete_network",
			Title:       "Delete network container",
			Description: "Delete a network container and connected lines via delete_network.",
			Instruction: "Delete a network container using delete_network:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Network ID/Name/IP", Desc: "Network ID, name, or IP", Required: true},
			},
		},
		{
			Name:        "check_network",
			Title:       "Check network status",
			Description: "Trigger SNMP/LLDP check for one or all network containers via check_network.",
			Instruction: "Trigger a network container check using check_network:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Network ID/Name/IP", Desc: "Specific network"},
				{Key: "all", Label: "Check all", Desc: "Set true to check all networks"},
			},
		},
		{
			Name:        "update_server_certificate",
			Title:       "Add or update TLS server certificate monitor",
			Description: "Add or update a TLS certificate monitor entry via update_server_certificate.",
			Instruction: "Save a server certificate monitor using update_server_certificate:",
			Args: []promptArgSpec{
				{Key: "id", Label: "ID", Desc: "Existing ID to update (omit to add)"},
				{Key: "target", Label: "Target", Desc: "Hostname or IP"},
				{Key: "port", Label: "Port", Desc: "TLS port (default 443)"},
			},
		},
		{
			Name:        "delete_server_certificate",
			Title:       "Delete TLS server certificate monitor",
			Description: "Remove a server certificate monitor entry via delete_server_certificate.",
			Instruction: "Delete server certificate monitor using delete_server_certificate:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Certificate monitor ID", Desc: "Entry ID", Required: true},
			},
		},
		{
			Name:        "snmpset",
			Title:       "Set SNMP MIB object value",
			Description: "Modify a writable SNMP MIB object via snmpset.",
			Instruction: "Perform an SNMP SET operation using snmpset:",
			Args: []promptArgSpec{
				{Key: "target", Label: "Target", Desc: "IP, node name, or node ID", Required: true},
				{Key: "mib_object_name", Label: "MIB object", Desc: "MIB object name or OID", Required: true},
				{Key: "type", Label: "Value type", Desc: "integer or string", Required: true},
				{Key: "value", Label: "Value", Desc: "Value to write", Required: true},
			},
		},
		{
			Name:        "search_netflow",
			Title:       "Search NetFlow records",
			Description: "Query NetFlow traffic flows via search_netflow.",
			Instruction: "Search NetFlow records using search_netflow:",
			Args: []promptArgSpec{
				{Key: "src_filter", Label: "Source filter", Desc: "Source IP regex"},
				{Key: "dst_filter", Label: "Destination filter", Desc: "Destination IP regex"},
				{Key: "start_time", Label: "Start time", Desc: "Start time (default -1h)"},
				{Key: "limit", Label: "Limit", Desc: "Max records"},
			},
		},
		{
			Name:        "search_sflow",
			Title:       "Search sFlow records",
			Description: "Query sFlow flow samples via search_sflow.",
			Instruction: "Search sFlow records using search_sflow:",
			Args: []promptArgSpec{
				{Key: "src_filter", Label: "Source filter", Desc: "Source IP regex"},
				{Key: "dst_filter", Label: "Destination filter", Desc: "Destination IP regex"},
				{Key: "start_time", Label: "Start time", Desc: "Start time (default -1h)"},
				{Key: "limit", Label: "Limit", Desc: "Max records"},
			},
		},
		{
			Name:        "search_sflow_counter",
			Title:       "Search sFlow counter samples",
			Description: "Query sFlow interface counter samples via search_sflow_counter.",
			Instruction: "Search sFlow counter records using search_sflow_counter:",
			Args: []promptArgSpec{
				{Key: "remote_filter", Label: "Remote filter", Desc: "Exporter IP regex"},
				{Key: "type_filter", Label: "Type filter", Desc: "Counter type regex"},
				{Key: "start_time", Label: "Start time", Desc: "Start time (default -1h)"},
			},
		},
		{
			Name:        "get_otel_metrics",
			Title:       "Get OpenTelemetry metrics",
			Description: "Query OpenTelemetry metrics via get_otel_metrics.",
			Instruction: "Get OpenTelemetry metrics using get_otel_metrics:",
			Args: []promptArgSpec{
				{Key: "host_filter", Label: "Host filter", Desc: "Host regex"},
				{Key: "service_filter", Label: "Service filter", Desc: "Service regex"},
				{Key: "name_filter", Label: "Metric filter", Desc: "Metric name regex"},
			},
		},
		{
			Name:        "get_otel_traces",
			Title:       "Get OpenTelemetry traces",
			Description: "List recent or failing OpenTelemetry traces via get_otel_traces.",
			Instruction: "Get OpenTelemetry traces using get_otel_traces:",
			Args: []promptArgSpec{
				{Key: "service_filter", Label: "Service filter", Desc: "Service regex"},
				{Key: "error_only", Label: "Errors only", Desc: "Set true to list traces with errors"},
			},
		},
		{
			Name:        "get_otel_trace_detail",
			Title:       "Get OpenTelemetry trace spans",
			Description: "Inspect all spans of an OTel trace via get_otel_trace_detail.",
			Instruction: "Get OpenTelemetry trace spans using get_otel_trace_detail:",
			Args: []promptArgSpec{
				{Key: "trace_id", Label: "Trace ID", Desc: "TraceID to inspect", Required: true},
				{Key: "bucket", Label: "Bucket", Desc: "Optional bucket key"},
			},
		},
		{
			Name:        "get_ai_anomaly_list",
			Title:       "Get AI anomaly detection scores",
			Description: "List latest AI anomaly scores across pollings via get_ai_anomaly_list.",
			Instruction: "Get AI anomaly detection results using get_ai_anomaly_list:",
			Args: []promptArgSpec{
				{Key: "node_id", Label: "Node", Desc: "Optional node ID, name, or IP"},
				{Key: "min_score", Label: "Min score", Desc: "Optional minimum score threshold"},
			},
		},
		{
			Name:        "search_mib",
			Title:       "Search MIB definitions",
			Description: "Search loaded SNMP MIB objects via search_mib.",
			Instruction: "Search MIB database using search_mib:",
			Args: []promptArgSpec{
				{Key: "query", Label: "Query", Desc: "MIB name, OID, or description regex", Required: true},
				{Key: "limit", Label: "Limit", Desc: "Max objects to return"},
			},
		},
		{
			Name:        "get_sensor_list",
			Title:       "Get sensor health list",
			Description: "List registered telemetry sensors via get_sensor_list.",
			Instruction: "Get connected sensor status using get_sensor_list:",
			Args: []promptArgSpec{
				{Key: "host_filter", Label: "Host filter", Desc: "Host regex"},
				{Key: "type_filter", Label: "Type filter", Desc: "Sensor type regex"},
			},
		},
		{
			Name:        "get_report_summary",
			Title:       "Get log analytics report summary",
			Description: "Query aggregated log reports via get_report_summary.",
			Instruction: "Get log analytics report using get_report_summary:",
			Args: []promptArgSpec{
				{Key: "report_type", Label: "Report type", Desc: "Report kind (wifiAP, blueDevice, envMonitor, dnsq, twAgentAlert, etc.)", Required: true},
				{Key: "filter", Label: "Filter", Desc: "Optional regex filter"},
			},
		},
		{
			Name:        "get_mac_address_list",
			Title:       "Get MAC address list",
			Description: "Get aggregated MAC addresses via get_mac_address_list.",
			Instruction: "Get aggregated MAC address table using get_mac_address_list:",
			Args: []promptArgSpec{
				{Key: "mac_filter", Label: "MAC filter", Desc: "MAC regex"},
				{Key: "vendor_filter", Label: "Vendor filter", Desc: "Vendor regex"},
			},
		},
		{
			Name:        "get_wifi_ap_list",
			Title:       "Get detected Wi-Fi AP list",
			Description: "List detected Wi-Fi access points via get_wifi_ap_list.",
			Instruction: "Get Wi-Fi access points using get_wifi_ap_list:",
			Args: []promptArgSpec{
				{Key: "ssid_filter", Label: "SSID filter", Desc: "SSID regex"},
				{Key: "bssid_filter", Label: "BSSID filter", Desc: "BSSID regex"},
			},
		},
		{
			Name:        "get_bluetooth_device_list",
			Title:       "Get detected Bluetooth device list",
			Description: "List detected Bluetooth/BLE devices via get_bluetooth_device_list.",
			Instruction: "Get Bluetooth devices using get_bluetooth_device_list:",
			Args: []promptArgSpec{
				{Key: "name_filter", Label: "Name filter", Desc: "Device name regex"},
				{Key: "public_address_only", Label: "Public addresses only", Desc: "true/false"},
			},
		},
		{
			Name:        "get_fdb_table",
			Title:       "Get SNMP switch FDB table",
			Description: "Fetch live switch FDB (MAC-to-port) entries via get_fdb_table.",
			Instruction: "Fetch switch FDB table using get_fdb_table:",
			Args: []promptArgSpec{
				{Key: "id", Label: "Network ID/Name/IP", Desc: "Network or switch reference", Required: true},
			},
		},
		{
			Name:        "auto_connect_lines",
			Title:       "Auto-connect map lines",
			Description: "Infer and add map topology lines via auto_connect_lines.",
			Instruction: "Auto-connect map lines using auto_connect_lines:",
			Args: []promptArgSpec{
				{Key: "mode", Label: "Mode", Desc: "strict (default) or speculative"},
			},
		},
		{
			Name:        "auto_layout",
			Title:       "Auto-arrange map layout",
			Description: "Arrange node coordinates or undo previous layout via auto_layout.",
			Instruction: "Optimize map layout using auto_layout:",
			Args: []promptArgSpec{
				{Key: "mode", Label: "Mode", Desc: "hierarchical, cluster, or categorized"},
				{Key: "undo", Label: "Undo", Desc: "true to revert last layout"},
			},
		},
		{
			Name:        "gnmi_capabilities",
			Title:       "Get gNMI capabilities",
			Description: "Query gNMI capabilities and models via gnmi_capabilities.",
			Instruction: "Get gNMI capabilities using gnmi_capabilities:",
			Args: []promptArgSpec{
				{Key: "node_id", Label: "Node ID/Name/IP", Desc: "Managed node reference", Required: true},
				{Key: "target", Label: "Target host:port", Desc: "Optional target override"},
			},
		},
		{
			Name:        "gnmi_get",
			Title:       "Query gNMI telemetry path",
			Description: "Execute gNMI Get on a telemetry path via gnmi_get.",
			Instruction: "Query gNMI path using gnmi_get:",
			Args: []promptArgSpec{
				{Key: "node_id", Label: "Node ID/Name/IP", Desc: "Managed node reference", Required: true},
				{Key: "path", Label: "gNMI path", Desc: "Telemetry path (e.g. /interfaces)", Required: true},
				{Key: "encoding", Label: "Encoding", Desc: "Optional encoding (json_ietf, json, etc.)"},
			},
		},
	}

	for _, spec := range specs {
		registerSpecPrompt(s, spec)
	}
}
