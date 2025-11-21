package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gosnmp/gosnmp"
	"github.com/twsnmp/rdap"
	"github.com/xhit/go-str2duration/v2"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/twsnmp/twsnmpfk/datastore"
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
		Description: "get polling log data from TWSNMP",
	}, getPollingLogData)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "do_ping",
		Description: "do ping",
	}, doPing)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_mib_tree",
		Description: "get MIB tree from TWSNMP",
	}, getMIBTree)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "snmpwalk",
		Description: "SNMP walk tool",
	}, snmpWalk)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_node",
		Description: "add node to TWSNMP",
	}, addNode)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_node",
		Description: "update node name,ip, position,description or icon",
	}, updateNode)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ip_address_list",
		Description: "get IP address list from TWSNMP",
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
		Description: "get syslog summary from TWSNMP",
	}, getSyslogSummary)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_snmp_trap_log",
		Description: "search SNMP trap log from TWSNMP",
	}, searchSnmpTrapLog)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_server_certificate_list",
		Description: "get server certificate list from TWSNMP",
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
	NameFilter  string `json:"name_filter" jsonschema:"name_filter specifies the search criteria for node names using regular expressions.If blank, all nodes are searched."`
	IPFilter    string `json:"ip_filter" jsonschema:"ip_filter specifies the search criteria for node IP address using regular expressions.If blank, all nodes are searched."`
	StateFilter string `json:"state_filter" jsonschema:"state_filter uses a regular expression to specify search criteria for node state names(normal,warn,low,high,repair,unknown).If blank, all nodes are searched."`
}

func getNodeList(ctx context.Context, req *mcp.CallToolRequest, args getNodeListParams) (*mcp.CallToolResult, any, error) {
	name := makeRegexFilter(args.NameFilter)
	ip := makeRegexFilter(args.IPFilter)
	state := makeRegexFilter(args.StateFilter)
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
	j, err := json.Marshal(&list)
	if err != nil {
		j = []byte(err.Error())
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
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
}

type getNetworkListParams struct {
	NameFilter string `json:"name_filter" jsonschema:"name_filter specifies the search criteria for network names using regular expressions.If blank, all networks are searched."`
	IPFilter   string `json:"ip_filter" jsonschema:"ip_filter specifies the search criteria for network IP address using regular expressions.If blank, all networks are searched."`
}

func getNetworkList(ctx context.Context, req *mcp.CallToolRequest, args getNetworkListParams) (*mcp.CallToolResult, any, error) {
	name := makeRegexFilter(args.NameFilter)
	ip := makeRegexFilter(args.IPFilter)
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
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		j = []byte(err.Error())
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
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
	TypeFilter     string `json:"type_filter" jsonschema:"type_filter uses a regular expression to specify search criteria for polling type names.If blank, all pollings are searched.Type names can be ping,tcp,http,dns,twsnmp,syslog"`
	NameFilter     string `json:"name_filter" jsonschema:"name_filter specifies the search criteria for polling names using regular expressions.If blank, all pollings are searched."`
	NodeNameFilter string `json:"node_name_filter" jsonschema:"node_name_filter specifies the search criteria for node names of polling using regular expressions.If blank, all pollings are searched."`
	StateFilter    string `json:"state_filter" jsonschema:"state_filter uses a regular expression to specify search criteria for polling state names.If blank, all pollings are searched.State names can be normal,warn,low,high,repair,unknown"`
}

func getPollingList(ctx context.Context, req *mcp.CallToolRequest, args getPollingListParams) (*mcp.CallToolResult, any, error) {
	name := makeRegexFilter(args.NameFilter)
	nodeName := makeRegexFilter(args.NodeNameFilter)
	typeFilter := makeRegexFilter(args.TypeFilter)
	state := makeRegexFilter(args.StateFilter)
	list := []mcpPollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
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
			NodeName: n.Name,
			LastTime: time.Unix(0, p.LastTime).Format(time.RFC3339Nano),
			Level:    p.Level,
			State:    n.State,
			Result:   p.Result,
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_polling_log tool
type mcpPollingLogEnt struct {
	Time   string
	State  string
	Result map[string]any
}

type getPollingLogParams struct {
	ID    string `json:"id" jsonschema:"The ID of the polling to retrieve the polling log"`
	Limit int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 1, max 2000"`
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
	limit := args.Limit
	if limit < 1 || limit > 2000 {
		limit = 100
	}
	list := []mcpPollingLogEnt{}
	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		list = append(list, mcpPollingLogEnt{
			Time:   time.Unix(0, l.Time).Format(time.RFC3339),
			State:  l.State,
			Result: l.Result,
		})
		return len(list) <= limit
	})
	j, err := json.Marshal(&list)
	if err != nil {
		j = []byte(err.Error())
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
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
	limit := args.Limit
	if limit < 100 || limit > 2000 {
		limit = 100
	}
	list := []mcpPollingLogEnt{}
	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		list = append(list, mcpPollingLogEnt{
			Time:   time.Unix(0, l.Time).Format(time.RFC3339),
			State:  l.State,
			Result: l.Result,
		})
		return len(list) < limit
	})
	if len(list) < 1 {
		return nil, nil, fmt.Errorf("polling log not found")
	}
	keys := []string{}
	for k, v := range list[0].Result {
		if k == "lastTime" {
			continue
		}
		if _, ok := v.(float64); !ok {
			continue
		}
		keys = append(keys, k)
	}
	csv := []string{"time,state," + strings.Join(keys, ",")}
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
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: strings.Join(csv, "\n")},
		},
	}, nil, nil
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
	Target  string `json:"target" jsonschema:"ping target ip address or host name"`
	Size    int    `json:"size" jsonschema:"ping packet size"`
	TTL     int    `json:"ttl" jsonschema:"IP packet TTL"`
	Timeout int    `json:"timeout" jsonschema:"timeout sec of ping"`
}

func doPing(ctx context.Context, req *mcp.CallToolRequest, args doPingParams) (*mcp.CallToolResult, any, error) {
	target := getTargetIP(args.Target)
	if target == "" {
		return nil, nil, fmt.Errorf("target ip not found")
	}
	timeout := args.Timeout
	if timeout < 1 || timeout > 10 {
		timeout = 3
	}

	size := args.Size
	if size < 1 || size > 1500 {
		size = 64
	}
	ttl := args.TTL
	if ttl < 1 || ttl > 255 {
		ttl = 254
	}
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
	j, err := json.Marshal(&res)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// getTargetIP: targetからIPアドレスを取得する、targetはノード名、ホスト名、IPアドレス
func getTargetIP(target string) string {
	ipreg := regexp.MustCompile(`^[0-9.]+$`)
	if ipreg.MatchString(target) {
		return target
	}
	n := datastore.FindNodeFromName(target)
	if n != nil {
		return n.IP
	}
	if ips, err := net.LookupIP(target); err == nil {
		for _, ip := range ips {
			if ip.IsGlobalUnicast() {
				s := ip.To4().String()
				if ipreg.MatchString(s) {
					return s
				}
			}
		}
	}
	return ""
}

// get_mib_tree tool
func getMIBTree(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	j, err := json.Marshal(&datastore.MIBTree)
	if err != nil {
		j = []byte(err.Error())
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpMIBEnt struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}
type snmpWalkParams struct {
	Target        string `json:"target"`
	MIBObjectName string `json:"mib_object_name" jsonschema:"mib object name"`
	Community     string `json:"community" jsonschema:"community name for snmp v2c mode"`
	User          string `json:"user" jsonschema:"User name for snmp v3 mode"`
	Password      string `json:"password" jsonschema:"Password for snmp v3 mode"`
	SnmpMode      string `json:"snmp_mode" jsonschema:"snmp mode (v2c,v3auth,v3authpriv,v3authprivex)"`
}

func snmpWalk(ctx context.Context, req *mcp.CallToolRequest, args snmpWalkParams) (*mcp.CallToolResult, any, error) {
	community := args.Community
	user := args.User
	password := args.Password
	snmpMode := args.SnmpMode
	name := args.MIBObjectName
	if name == "" {
		return nil, nil, fmt.Errorf("no mib_object_name")
	}
	target := args.Target
	if target == "" {
		return nil, nil, fmt.Errorf("no target")
	}
	if n := datastore.FindNodeFromName(target); n != nil {
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
		target = n.IP
	} else {
		target = getTargetIP(target)
		if target == "" {
			return nil, nil, fmt.Errorf("target not found")
		}
	}
	agent := &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Transport: "udp",
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(datastore.MapConf.Timeout) * time.Second,
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
	}
	res := []mcpMIBEnt{}
	err := agent.Connect()
	if err != nil {
		return nil, nil, err
	}
	defer agent.Conn.Close()
	err = agent.Walk(nameToOID(name), func(variable gosnmp.SnmpPDU) error {
		name := datastore.MIBDB.OIDToName(variable.Name)
		value := datastore.GetMIBValueString(name, &variable, false)
		res = append(res, mcpMIBEnt{
			Name:  name,
			Value: value,
		})
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	j, err := json.Marshal(&res)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type addNodeParams struct {
	Name        string `json:"name" jsonschema:"node name"`
	IP          string `json:"ip" jsonschema:"node ip address"`
	Icon        string `json:"icon" jsonschema:"icon of node"`
	Description string `json:"description" jsonschema:"description of node"`
	X           int    `json:"x" jsonschema:"x position of node"`
	Y           int    `json:"y" jsonschema:"y position of node"`
}

func addNode(ctx context.Context, req *mcp.CallToolRequest, args addNodeParams) (*mcp.CallToolResult, any, error) {
	icon := args.Icon
	if icon == "" {
		icon = "desktop"
	}
	descr := args.Description
	name := args.Name
	if name == "" {
		return nil, nil, fmt.Errorf("node name is empty")
	}
	ip := args.IP
	if ip == "" {
		return nil, nil, fmt.Errorf("IP is empty")
	}
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
		Descr: descr,
		State: "unknown",
	}
	if err := datastore.AddNode(n); err != nil {
		return nil, nil, err
	}
	datastore.AddPolling(&datastore.PollingEnt{
		Name:   "PING",
		Type:   "ping",
		NodeID: n.ID})
	j, err := json.Marshal(&mcpNodeEnt{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Descr,
		IP:          n.IP,
		State:       n.State,
		X:           n.X,
		Y:           n.Y,
		Icon:        n.Icon,
	})
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// update_node

type updateNodeParams struct {
	ID          string `json:"id" jsonschema:"node id or current name or current ip"`
	Name        string `json:"name" jsonschema:"new node name or empty"`
	IP          string `json:"ip" jsonschema:"new ip address or empty"`
	Icon        string `json:"icon" jsonschema:"new icon or empty"`
	Description string `json:"description" jsonschema:"description of node"`
	X           int    `json:"x" jsonschema:"x position of node"`
	Y           int    `json:"y" jsonschema:"y position of node"`
}

func updateNode(ctx context.Context, req *mcp.CallToolRequest, args updateNodeParams) (*mcp.CallToolResult, any, error) {
	id := args.ID
	n := datastore.GetNode(id)
	if n == nil {
		n = datastore.FindNodeFromName(id)
		if n == nil {
			n = datastore.FindNodeFromIP(id)
			if n == nil {
				return nil, nil, fmt.Errorf("node not found")
			}
		}
	}
	icon := args.Icon
	descr := args.Description
	name := args.Name
	x := args.X
	y := args.Y
	if x > 0 {
		n.X = x
	}
	if y > 0 {
		n.Y = y
	}
	if icon != "" {
		n.Icon = icon
	}
	if descr != "" {
		n.Descr = descr
	}
	if name != "" {
		n.Name = name
	}
	j, err := json.Marshal(&mcpNodeEnt{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Descr,
		IP:          n.IP,
		State:       n.State,
		X:           n.X,
		Y:           n.Y,
		Icon:        n.Icon,
	})
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_ip_address_list
type mcpIPEnt struct {
	IP        string
	MAC       string
	Node      string
	Vendor    string
	FirstTime string
	LastTime  string
}

func getIPAddressList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	list := []mcpIPEnt{}
	datastore.ForEachArp(func(l *datastore.ArpEnt) bool {
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
			Vendor:    l.Vendor,
			FirstTime: time.Unix(0, l.FirstTime).Format(time.RFC3339Nano),
			LastTime:  time.Unix(0, l.LastTime).Format(time.RFC3339Nano),
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

type mcpResourceMonitorEnt struct {
	Time        string
	CPUUsage    string
	MemoryUsage string
	SwapUsage   string
	DiskUsage   string
	Load        string
}

func getResourceMonitorList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	list := []mcpResourceMonitorEnt{}
	skip := 30
	if len(MonitorDataes) < 120 {
		skip = 5
	}
	for i, m := range MonitorDataes {
		if i%skip != 0 {
			continue
		}
		list = append(list, mcpResourceMonitorEnt{
			Time:        time.Unix(m.Time, 0).Format(time.RFC3339),
			CPUUsage:    fmt.Sprintf("%.02f%%", m.CPU),
			MemoryUsage: fmt.Sprintf("%.02f%%", m.Mem),
			SwapUsage:   fmt.Sprintf("%.02f%%", m.Swap),
			DiskUsage:   fmt.Sprintf("%.02f%%", m.Disk),
			Load:        fmt.Sprintf("%.02f", m.Load),
		})
	}

	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil

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
	NodeFilter  string `json:"node_filter" jsonschema:"node_filter specifies the search criteria for node names using regular expressions. If blank, no filter."`
	TypeFilter  string `json:"type_filter" jsonschema:"type_filter specifies the search criteria for type names using regular expressions.If blank, no filter."`
	LevelFilter string `json:"level_filter" jsonschema:"level_filter specifies the search criteria for level names using regular expressions. If blank, no filter.Level names can be warn,low,high,debug,info"`
	EventFilter string `json:"event_filter" jsonschema:"event_filter specifies the search criteria for events using regular expressions.If blank, no filter."`
	StartTime   string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now."`
	EndTime     string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	Limit       int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 100,max 10000"`
}

func searchEventLog(ctx context.Context, req *mcp.CallToolRequest, args searchEventLogParams) (*mcp.CallToolResult, any, error) {
	node := makeRegexFilter(args.NodeFilter)
	typeFilter := makeRegexFilter(args.TypeFilter)
	level := makeRegexFilter(args.LevelFilter)
	event := makeRegexFilter(args.EventFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit < 100 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	log.Printf("mcp search_event_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpEventLogEnt{}
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
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
			Time:  time.Unix(0, l.Time).Format(time.RFC3339Nano),
			Type:  l.Type,
			Level: l.Level,
			Node:  l.NodeName,
			Event: l.Event,
		})
		return len(list) < limit
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil

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
	LevelFilter   string `json:"level_filter" jsonschema:"level_filter specifies the search criteria for level names using regular expressions. If blank, no filter.Level names can be warn,low,high,debug,info"`
	HostFilter    string `json:"host_filter" jsonschema:"host_filter specifies the search criteria for host names using regular expressions. If blank, no filter."`
	TagFilter     string `json:"tag_filter" jsonschema:"tag_filter specifies the search criteria for tag names using regular expressions.If blank, no filter."`
	MessageFilter string `json:"message_filter" jsonschema:"message_filter specifies the search criteria for messages using regular expressions.If blank, no filter."`
	StartTime     string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now."`
	EndTime       string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	Limit         int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 100,max 10000"`
}

func searchSyslog(ctx context.Context, req *mcp.CallToolRequest, args searchSyslogParams) (*mcp.CallToolResult, any, error) {
	host := makeRegexFilter(args.HostFilter)
	tag := makeRegexFilter(args.TagFilter)
	level := makeRegexFilter(args.LevelFilter)
	message := makeRegexFilter(args.MessageFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit < 100 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	log.Printf("mcp search_syslog limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpSyslogEnt{}
	datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
		e := mcpSyslogEnt{
			Time:     time.Unix(0, l.Time).Format(time.RFC3339Nano),
			Host:     l.Host,
			Level:    l.Level,
			Type:     l.Type,
			Tag:      l.Tag,
			Facility: l.Facility,
			Severity: l.Severity,
			Message:  l.Message,
		}
		if message != nil && !message.MatchString(e.Message) {
			return true
		}
		if tag != nil && !tag.MatchString(e.Tag) {
			return true
		}
		if level != nil && !level.MatchString(e.Level) {
			return true
		}
		if host != nil && !host.MatchString(e.Host) {
			return true
		}
		list = append(list, e)
		return len(list) < limit
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_syslog_summary tool
type mcpSyslogSummaryEnt struct {
	Pattern string
	Count   int
}

type getSyslogSummaryParams struct {
	LevelFilter   string `json:"level_filter" jsonschema:"level_filter specifies the search criteria for level names using regular expressions. If blank, no filter.Level names can be warn,low,high,debug,info"`
	HostFilter    string `json:"host_filter" jsonschema:"host_filter specifies the search criteria for host names using regular expressions. If blank, no filter."`
	TagFilter     string `json:"tag_filter" jsonschema:"tag_filter specifies the search criteria for tag names using regular expressions.If blank, no filter."`
	MessageFilter string `json:"message_filter" jsonschema:"message_filter specifies the search criteria for messages using regular expressions.If blank, no filter."`
	StartTime     string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now."`
	EndTime       string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	TopN          int    `json:"top_n" jsonschema:"Top n syslog pattern. min 5,max 500"`
}

func getSyslogSummary(ctx context.Context, req *mcp.CallToolRequest, args getSyslogSummaryParams) (*mcp.CallToolResult, any, error) {
	host := makeRegexFilter(args.HostFilter)
	tag := makeRegexFilter(args.TagFilter)
	level := makeRegexFilter(args.LevelFilter)
	message := makeRegexFilter(args.MessageFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	topN := args.TopN
	if topN < 5 {
		topN = 5
	}
	if topN > 500 {
		topN = 500
	}
	log.Printf("mcp get_syslog_summary topn=%d st=%v et=%v", topN, time.Unix(0, st), time.Unix(0, et))
	patternMap := make(map[string]int)
	datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
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
	list := []mcpSyslogSummaryEnt{}
	for p, c := range patternMap {
		list = append(list, mcpSyslogSummaryEnt{
			Pattern: p,
			Count:   c,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})
	if len(list) > topN {
		list = list[:topN]
	}
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil

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
	FromFilter     string `json:"from_filter" jsonschema:"from_filter specifies the search criteria for trap sender address using regular expressions.If blank, no filter."`
	TrapTypeFilter string `json:"trap_type_filter" jsonschema:"trap_type_filter specifies the search criteria for SNMP trap types using regular expressions.If blank, no filter."`
	VariableFilter string `json:"variable_filter" jsonschema:"variable_filter specifies the search criteria for SNMP trap variables using regular expressions.If blank, no filter."`
	StartTime      string `json:"start_time" jsonschema:"start date and time of logs to search or duration from now."`
	EndTime        string `json:"end_time" jsonschema:"end date and time of logs to search.empty or now is current time."`
	Limit          int    `json:"limit" jsonschema:"Limit on number of logs retrieved. min 100,max 10000"`
}

func searchSnmpTrapLog(ctx context.Context, req *mcp.CallToolRequest, args searchSnmpTrapLogParams) (*mcp.CallToolResult, any, error) {
	from := makeRegexFilter(args.FromFilter)
	trapType := makeRegexFilter(args.TrapTypeFilter)
	variable := makeRegexFilter(args.VariableFilter)
	start := args.StartTime
	if start == "" {
		start = "-1h"
	}
	end := args.EndTime
	st, et, err := getTimeRange(start, end)
	if err != nil {
		return nil, nil, err
	}
	limit := args.Limit
	if limit < 100 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	log.Printf("mcp search_snmp_trap_log limit=%d st=%v et=%v", limit, time.Unix(0, st), time.Unix(0, et))
	list := []mcpSNMPTrapLogEnt{}
	datastore.ForEachTraps(st, et, func(l *datastore.TrapEnt) bool {
		e := mcpSNMPTrapLogEnt{
			Time:        time.Unix(0, l.Time).Format(time.RFC3339Nano),
			FromAddress: l.FromAddress,
			TrapType:    l.TrapType,
			Variables:   l.Variables,
		}
		if from != nil && !from.MatchString(e.FromAddress) {
			return true
		}
		if variable != nil && !variable.MatchString(e.Variables) {
			return true
		}
		if trapType != nil && !trapType.MatchString(e.TrapType) {
			return true
		}
		list = append(list, e)
		return len(list) < limit
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_server_certificate_list
type mcpServerCertificateEnt struct {
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

func getServerCertificateList(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
	list := []mcpServerCertificateEnt{}
	datastore.ForEachCertMonitors(func(c *datastore.CertMonitorEnt) bool {
		list = append(list, mcpServerCertificateEnt{
			State:        c.State,
			Server:       c.Target,
			Port:         c.Port,
			Subject:      c.Subject,
			Issuer:       c.Issuer,
			SerialNumber: c.SerialNumber,
			Verify:       c.Verify,
			NotBefore:    time.Unix(c.NotBefore, 0).Format(time.RFC3339),
			NotAfter:     time.Unix(c.NotAfter, 0).Format(time.RFC3339),
			Error:        c.Error,
			FirstTime:    time.Unix(0, c.FirstTime).Format(time.RFC3339Nano),
			LastTime:     time.Unix(0, c.LastTime).Format(time.RFC3339Nano),
		})
		return true
	})
	j, err := json.Marshal(&list)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// add_event_log tool
type addEventLogParams struct {
	Level string `json:"level" jsonschema:"Level of event (info,normal,warn,low,high)"`
	Node  string `json:"node" jsonschema:"Node name associated with the event"`
	Event string `json:"event" jsonschema:"Event log contents"`
}

func addEventLog(ctx context.Context, req *mcp.CallToolRequest, args addEventLogParams) (*mcp.CallToolResult, any, error) {
	level := args.Level
	if level == "" {
		level = "info"
	}
	event := args.Event
	node := args.Node
	nodeID := ""
	if node != "" {
		if n := datastore.FindNodeFromName(node); n != nil {
			nodeID = n.ID
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Time:     time.Now().UnixNano(),
		Level:    level,
		Type:     "mcp",
		Event:    event,
		NodeName: node,
		NodeID:   nodeID,
	})
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "ok"},
		},
	}, nil, nil

}

// get_ip_address_info
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
	ip := args.IP
	info := new(mcpIPInfoEnt)
	info.IP = ip
	if n := datastore.FindNodeFromIP(ip); n != nil {
		info.Node = n.Name
	}
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancel()
	if names, err := r.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
		info.DNSNames = names
	}
	info.Location = datastore.GetLoc(ip)
	if !strings.HasPrefix(info.Location, "LOCAL") {
		client := &rdap.Client{}
		if ri, err := client.QueryIP(ip); err == nil {
			info.RDAPIPVersion = ri.IPVersion
			info.RDAPName = ri.Name
			info.RDAPCountry = ri.Country
			info.RDAPWhoisServer = ri.Port43
			info.RDAPHandle = ri.Handle
			info.RDAPType = ri.Type
		}
	}
	j, err := json.Marshal(info)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

// get_mac_address_info
type getMACInfoParams struct {
	MAC string `json:"mac" jsonschema:"MAC adddress"`
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
	j, err := json.Marshal(info)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(j)},
		},
	}, nil, nil
}

func normalizeMACAddr(m string) string {
	if hw, err := net.ParseMAC(m); err == nil {
		m = strings.ToUpper(hw.String())
		return m
	}
	m = strings.Replace(m, "-", ":", -1)
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
		if a.MAC == mac {
			ip = a.IP
			return false
		}
		return true
	})
	return ip
}

// getTimeRange
func getTimeRange(start, end string) (int64, int64, error) {
	var st time.Time
	var err error
	et := time.Now()
	if start == "" {
		return 0, 0, fmt.Errorf("start_time must not be empty")
	}
	if d, err := str2duration.ParseDuration(start); err == nil {
		st = et.Add(d)
	} else if st, err = dateparse.ParseLocal(start); err != nil {
		return 0, 0, err
	}
	if end != "" && end != "now" {
		if et, err = dateparse.ParseLocal(end); err != nil {
			return 0, 0, err
		}
	}
	if st.UnixNano() > et.UnixNano() {
		return 0, 0, fmt.Errorf("start_time must be before end_time")
	}
	return st.UnixNano(), et.UnixNano(), nil
}

func nameToOID(name string) string {
	oid := datastore.MIBDB.NameToOID(name)
	if oid == ".1" {
		oid = ".1.3"
	}
	if oid == ".0.0" {
		if matched, _ := regexp.MatchString(`\\.[0-9.]+`, name); matched {
			return name
		}
	}
	return oid
}

// makeRegexFilter
func makeRegexFilter(s string) *regexp.Regexp {
	if s != "" {
		if f, err := regexp.Compile(s); err == nil && f != nil {
			return f
		}
	}
	return nil
}
