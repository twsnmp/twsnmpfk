package backend

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gosnmp/gosnmp"
	"github.com/labstack/echo/v4"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/xhit/go-str2duration/v2"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/ping"
)

var stopMCPCh = make(chan bool)
var mcpAllow sync.Map

func mcpServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	timer := time.NewTicker(time.Second * 5)
	var mcpsv any = nil
	var e *echo.Echo
	stopMCPServer := func() {
		if mcpsv == nil {
			return
		}
		log.Println("stop mcp server")
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: i18n.Trans("Stop MCP server"),
		})
		switch m := mcpsv.(type) {
		case *server.SSEServer:
			m.Shutdown(ctx)
		case *server.StreamableHTTPServer:
			m.Shutdown(ctx)
		}
		mcpsv = nil
		if e != nil {
			e.Shutdown(ctx)
			e = nil
		}
	}
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			stopMCPServer()
			return
		case <-stopMCPCh:
			stopMCPServer()
		case <-timer.C:
			if datastore.MapConf.MCPTransport != "off" && mcpsv == nil {
				log.Println("start mcp server")
				setMCPAllow()
				mcpsv, e = startMCPServer()
				datastore.AddEventLog(&datastore.EventLogEnt{
					Type:  "system",
					Level: "info",
					Event: fmt.Sprintf(i18n.Trans("Start MCP server: transport=%s endpoint=%s"), datastore.MapConf.MCPTransport, datastore.MapConf.MCPEndpoint),
				})
			} else if datastore.MapConf.MCPTransport == "off" && mcpsv != nil {
				stopMCPServer()
			}
		}
	}
}

func NotifyMCPConfigChanged() {
	stopMCPCh <- true
}

func startMCPServer() (any, *echo.Echo) {
	// Create MCP Server
	s := server.NewMCPServer(
		"TWSNMP FK MCP Server",
		"1.26.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)
	// Add tools to MCP server
	addGetNodeListTool(s)
	addGetNetworkListTool(s)
	addGetPollingListTool(s)
	addGetPollingLogTool(s)
	addDoPingtTool(s)
	addGetMIBTreeTool(s)
	addSNMPWalkTool(s)
	addAddNodeTool(s)
	addUpdateNodeTool(s)
	addGetIPAddressListTool(s)
	addGetResourceMonitorListTool(s)
	addSearchEventLogTool(s)
	addSearchSyslogTool(s)
	addGetSyslogSummaryTool(s)
	addSearchSNMPTrapLogTool(s)
	addGetServerCertificateListTool(s)
	addAddEventLogTool(s)
	sv := &http.Server{}
	sv.Addr = datastore.MapConf.MCPEndpoint
	if cert, err := getMCPServerCert(); err == nil {
		if cert != nil {
			sv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{*cert},
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
				},
				MinVersion: tls.VersionTLS13,
			}
		}
	} else {
		log.Printf("getMCPServerCert err=%v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	var mcpsv any = nil

	if datastore.MapConf.MCPTransport == "sse" {
		sseServer := server.NewSSEServer(s)
		e.Any("/sse", func(c echo.Context) error {
			if !checkMCPACL(c) {
				return echo.ErrUnauthorized
			}
			sseServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/message", func(c echo.Context) error {
			if !checkMCPACL(c) {
				return echo.ErrUnauthorized
			}
			sseServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})

		log.Printf("sse mcp server listening on %s", datastore.MapConf.MCPEndpoint)
		mcpsv = sseServer
	} else {
		streamServer := server.NewStreamableHTTPServer(s)
		e.Any("/mcp", func(c echo.Context) error {
			if !checkMCPACL(c) {
				return echo.ErrUnauthorized
			}
			streamServer.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
		log.Printf("streamable HTTP server listening on %s", datastore.MapConf.MCPEndpoint)
		mcpsv = streamServer
	}
	go func() {
		if err := e.StartServer(sv); err != nil {
			log.Printf("start mcp server err=%v", err)
		}
	}()
	return mcpsv, e
}

func getMCPServerCert() (*tls.Certificate, error) {
	if datastore.MCPCert == "" || datastore.MCPKey == "" {
		return nil, nil
	}
	keyPem, err := os.ReadFile(datastore.MCPKey)
	if err == nil {
		certPem, err := os.ReadFile(datastore.MCPCert)
		if err == nil {
			cert, err := tls.X509KeyPair(certPem, keyPem)
			if err == nil {
				return &cert, nil
			}
		}
	}
	return nil, err
}

func setMCPAllow() {
	for _, ip := range strings.Split(datastore.MapConf.MCPFrom, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			mcpAllow.Store(ip, true)
		}
	}
}

func checkMCPACL(c echo.Context) bool {
	if datastore.MapConf.MCPToken != "" {
		t := c.Request().Header.Get("Authorization")
		log.Printf("checkMCPACL token=%+v", t)
		if !strings.Contains(t, datastore.MapConf.MCPToken) {
			return false
		}
	}
	if datastore.MapConf.MCPFrom == "" {
		return true
	}
	if ip, _, err := net.SplitHostPort(c.Request().RemoteAddr); err == nil {
		if _, ok := mcpAllow.Load(ip); ok {
			return true
		}
	}
	if _, ok := mcpAllow.Load(c.RealIP()); ok {
		return true
	}
	return false
}

// get_node_list tool
type mcpNodeEnt struct {
	ID         string
	Name       string
	IP         string
	MAC        string
	State      string
	X          int
	Y          int
	Icon       string
	Descrption string
}

func addGetNodeListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_node_list",
		mcp.WithDescription("get node list from TWSNMP"),
		mcp.WithString("name_filter",
			mcp.Description(
				`name_filter specifies the search criteria for node names using regular expressions.
If blank, all nodes are searched.
`),
		),
		mcp.WithString("ip_filter",
			mcp.Description(
				`ip_filter specifies the search criteria for node IP address using regular expressions.
If blank, all nodes are searched.
`),
		),
		mcp.WithString("state_filter",
			mcp.Description(
				`state_filter uses a regular expression to specify search criteria for node state names.
If blank, all nodes are searched.
State names can be "normal","warn","low","high","repair","unknown"
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := makeRegexFilter(request.GetString("name_filter", ""))
		ip := makeRegexFilter(request.GetString("ip_filter", ""))
		state := makeRegexFilter(request.GetString("state_filter", ""))
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
				ID:         n.ID,
				Name:       n.Name,
				IP:         n.IP,
				MAC:        n.MAC,
				X:          n.X,
				Y:          n.Y,
				Icon:       n.Icon,
				Descrption: n.Descr,
				State:      n.State,
			})
			return true
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// get_network_list tool
type mcpNetworkEnt struct {
	ID         string
	Name       string
	IP         string
	Ports      []string
	X          int
	Y          int
	Descrption string
}

func addGetNetworkListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_network_list",
		mcp.WithDescription("get network list from TWSNMP"),
		mcp.WithString("name_filter",
			mcp.Description(
				`name_filter specifies the search criteria for network names using regular expressions.
If blank, all networks are searched.
`),
		),
		mcp.WithString("ip_filter",
			mcp.Description(
				`ip_filter specifies the search criteria for network IP address using regular expressions.
If blank, all networks are searched.
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := makeRegexFilter(request.GetString("name_filter", ""))
		ip := makeRegexFilter(request.GetString("ip_filter", ""))
		list := []mcpNetworkEnt{}
		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			if name != nil && name.MatchString(n.Name) {
				return true
			}
			if ip != nil && ip.MatchString(n.IP) {
				return true
			}
			ports := []string{}
			for _, p := range n.Ports {
				ports = append(ports, fmt.Sprintf("%s=%s", p.Name, p.State))
			}
			list = append(list, mcpNetworkEnt{
				ID:         n.ID,
				Name:       n.Name,
				IP:         n.IP,
				X:          n.X,
				Y:          n.Y,
				Descrption: n.Descr,
				Ports:      ports,
			})
			return true
		})
		j, err := json.Marshal(&list)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
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

func addGetPollingListTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("get_polling_list",
		mcp.WithDescription("get polling list from TWSNMP"),
		mcp.WithString("type_filter",
			mcp.Description(
				`type_filter uses a regular expression to specify search criteria for polling type names.
If blank, all pollings are searched.
Type names can be "ping","tcp","http","dns","twsnmp","syslog"
`),
		),
		mcp.WithString("state_filter",
			mcp.Description(
				`state_filter uses a regular expression to specify search criteria for polling state names.
If blank, all pollings are searched.
State names can be "normal","warn","low","high","repair","unknown"
`),
		),
		mcp.WithString("name_filter",
			mcp.Description(
				`name_filter specifies the search criteria for polling names using regular expressions.
If blank, all pollings are searched.
`),
		),
		mcp.WithString("node_name_filter",
			mcp.Description(
				`node_name_filter specifies the search criteria for node names of polling using regular expressions.
If blank, all pollings are searched.
`),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := makeRegexFilter(request.GetString("name_filter", ""))
		nodeName := makeRegexFilter(request.GetString("node_name_filter", ""))
		typeFilter := makeRegexFilter(request.GetString("type_filter", ""))
		state := makeRegexFilter(request.GetString("state_filter", ""))
		list := []mcpPollingEnt{}
		datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
			if name != nil && name.MatchString(p.Name) {
				return true
			}
			if typeFilter != nil && typeFilter.MatchString(p.Type) {
				return true
			}
			n := datastore.GetNode(p.NodeID)
			if n == nil {
				return true
			}
			if nodeName != nil && nodeName.MatchString(n.Name) {
				return true
			}
			if state != nil && state.MatchString(p.State) {
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// get_polling_log tool
type mcpPollingLogEnt struct {
	Time   string
	State  string
	Result map[string]any
}

func addGetPollingLogTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("get_polling_log",
		mcp.WithDescription("get polling log from TWSNMP"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description(`The ID of the polling to retrieve the polling log`),
		),
		mcp.WithNumber("limit",
			mcp.DefaultNumber(100),
			mcp.Max(2000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 1,max 2000"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		polling := datastore.GetPolling(id)
		if polling == nil {
			return mcp.NewToolResultText("polling not found"), nil
		}
		limit := request.GetInt("limit", 100)
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
		return mcp.NewToolResultText(string(j)), nil
	})
}

// do_ping tool
type mcpPingEnt struct {
	Result       string `json:"Result"`
	Time         string `json:"Time"`
	RTT          string `json:"RTT"`
	RTTNano      int64  `json:"RTTNano"`
	Size         int    `json:"Size"`
	TTL          int    `json:"TTL"`
	ResponceFrom string `json:"ResponceFrom"`
	Location     string `json:"Location"`
}

func addDoPingtTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("do_ping",
		mcp.WithDescription("do ping"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("ping target ip address or host name"),
		),
		mcp.WithNumber("size",
			mcp.DefaultNumber(64),
			mcp.Max(1500),
			mcp.Min(64),
			mcp.Description("ping packate size"),
		),
		mcp.WithNumber("ttl",
			mcp.DefaultNumber(254),
			mcp.Max(254),
			mcp.Min(1),
			mcp.Description("ip packet TTL"),
		),
		mcp.WithNumber("timeout",
			mcp.DefaultNumber(2),
			mcp.Max(10),
			mcp.Min(1),
			mcp.Description("timeout sec of ping"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		target, err := request.RequireString("target")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		target = getTragetIP(target)
		if target == "" {
			return mcp.NewToolResultText("target ip not found"), nil
		}
		timeout := request.GetInt("timeout", 3)
		size := request.GetInt("size", 64)
		ttl := request.GetInt("ttl", 254)
		pe := ping.DoPing(target, timeout, 0, size, ttl)
		res := mcpPingEnt{
			Result:       pe.Stat.String(),
			Time:         time.Now().Format(time.RFC3339),
			RTT:          time.Duration(pe.Time).String(),
			Size:         pe.Size,
			ResponceFrom: pe.RecvSrc,
			TTL:          pe.RecvTTL,
			RTTNano:      pe.Time,
		}
		if pe.RecvSrc != "" {
			res.Location = datastore.GetLoc(pe.RecvSrc)
		}
		j, err := json.Marshal(&res)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// getTragetIP: targetからIPアドレスを取得する、targetはノード名、ホスト名、IPアドレス
func getTragetIP(target string) string {
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

func addGetMIBTreeTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("get_MIB_tree",
		mcp.WithDescription("get MIB tree from TWSNMP"),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		j, err := json.Marshal(&datastore.MIBTree)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

type mcpMIBEnt struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

func addSNMPWalkTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("snmpwalk",
		mcp.WithDescription("SNMP walk tool"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("snmpwalk target ip address or host name or node name"),
		),
		mcp.WithString("mib_object_name",
			mcp.Required(),
			mcp.Description("snmpwak mib object name"),
		),
		mcp.WithString("community",
			mcp.Description("snmp v2c comminity name"),
		),
		mcp.WithString("user",
			mcp.Description("snmp v3 user name"),
		),
		mcp.WithString("password",
			mcp.Description("snmp v3 password"),
		),
		mcp.WithString("snmpmode",
			mcp.Enum("", "v2c", "v3auth", "v3authpriv", "v3authprivex"),
			mcp.Description(
				`snmp mode
v2c : SNMP v2 (default)
v3auth: SNMP v3 authentication protocol is SHA1,privacy protocol is none.
v3authpriv: SNMP v3 authentication protocol is SHA1,privacy protocol is AES.
v3authprivex: SNMP v3 authentication protocol is SHA1,privacy protocol is AES256.
`),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		community := request.GetString("community", "")
		user := request.GetString("community", "")
		password := request.GetString("password", "")
		snmpMode := request.GetString("snmpmode", "")
		name, err := request.RequireString("mib_object_name")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		target, err := request.RequireString("target")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
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
			target = getTragetIP(target)
			if target == "" {
				return mcp.NewToolResultText("target ip not found"), nil
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
		err = agent.Connect()
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
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
			return mcp.NewToolResultText(err.Error()), nil
		}
		j, err := json.Marshal(&res)
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// add_node tool
func addAddNodeTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("add_node",
		mcp.WithDescription("add node to TWSNMP"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("node name"),
		),
		mcp.WithString("ip",
			mcp.Required(),
			mcp.Description("node ip address"),
		),
		mcp.WithString("icon",
			mcp.Description("icon of node"),
		),
		mcp.WithString("description",
			mcp.Description("description of node"),
		),
		mcp.WithNumber("x",
			mcp.Max(1000),
			mcp.Min(64),
			mcp.Description("x positon of node"),
		),
		mcp.WithNumber("y",
			mcp.Max(1000),
			mcp.Min(64),
			mcp.Description("y positon of node"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		icon := request.GetString("icon", "desktop")
		descr := request.GetString("description", "")
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		ip, err := request.RequireString("ip")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		x := request.GetInt("x", 64)
		y := request.GetInt("y", 64)
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
			return mcp.NewToolResultText(err.Error()), nil
		}
		datastore.AddPolling(&datastore.PollingEnt{
			Name:   "PING",
			Type:   "ping",
			NodeID: n.ID})
		j, err := json.Marshal(&mcpNodeEnt{
			ID:         n.ID,
			Name:       n.Name,
			Descrption: n.Descr,
			IP:         n.IP,
			State:      n.State,
			X:          n.X,
			Y:          n.Y,
			Icon:       n.Icon,
		})
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// update_node
func addUpdateNodeTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("update_node",
		mcp.WithDescription("update node name,ip, positon,description or icon"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("node id to update"),
		),
		mcp.WithString("name",
			mcp.Description("node name"),
		),
		mcp.WithString("ip",
			mcp.Description("node ip address"),
		),
		mcp.WithString("icon",
			mcp.Description("icon of node"),
		),
		mcp.WithString("description",
			mcp.Description("description of node"),
		),
		mcp.WithNumber("x",
			mcp.Description("x positon of node"),
		),
		mcp.WithNumber("y",
			mcp.Description("y positon of node"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		icon := request.GetString("icon", "")
		descr := request.GetString("description", "")
		name := request.GetString("name", "")
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		x := request.GetInt("x", -1)
		y := request.GetInt("y", -1)
		n := datastore.GetNode(id)
		if n == nil {
			return mcp.NewToolResultText("node not found"), nil
		}
		if x >= 0 {
			n.X = x
		}
		if y >= 0 {
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
			ID:         n.ID,
			Name:       n.Name,
			Descrption: n.Descr,
			IP:         n.IP,
			State:      n.State,
			X:          n.X,
			Y:          n.Y,
			Icon:       n.Icon,
		})
		if err != nil {
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
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

func addGetIPAddressListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_ip_address_list",
		mcp.WithDescription("get IP address list from TWSNMP"),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

type mcpResourceMonitorEnt struct {
	Time        string
	CPUUsage    string
	MemoryUsage string
	SwapUsage   string
	DiskUsage   string
	Load        string
}

func addGetResourceMonitorListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_resource_monitor_list",
		mcp.WithDescription("get resource monitor list from TWSNMP"),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// search_event_log tool
type mcpEventLogEnt struct {
	Time  string
	Type  string
	Level string
	Node  string
	Event string
}

func addSearchEventLogTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_event_log",
		mcp.WithDescription("search event log from TWSNMP"),
		mcp.WithString("node_filter",
			mcp.Description(
				`node_filter specifies the search criteria for node names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("type_filter",
			mcp.Description(
				`type_filter specifies the search criteria for type names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("level_filter",
			mcp.Description(
				`level_filter specifies the search criteria for level names using regular expressions.
If blank, no filter.
Level names can be "warn","low","high","debug","info" 
`),
		),
		mcp.WithString("event_filter",
			mcp.Description(
				`event_filter specifies the search criteria for events using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("limit_log_count",
			mcp.DefaultNumber(100),
			mcp.Max(10000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 100,max 10000"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "-300ms", "-1.5h" or "-2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		node := makeRegexFilter(request.GetString("node_filter", ""))
		typeFilter := makeRegexFilter(request.GetString("type_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		event := makeRegexFilter(request.GetString("event_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := request.GetInt("limit_log_count", 100)
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
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

func addSearchSyslogTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_syslog",
		mcp.WithDescription("search syslog from TWSNMP"),
		mcp.WithString("host_filter",
			mcp.Description(
				`host_filter specifies the search criteria for host names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("tag_filter",
			mcp.Description(
				`tag_filter specifies the search criteria for tag names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("level_filter",
			mcp.Description(
				`level_filter specifies the search criteria for level names using regular expressions.
If blank, no filter.
Level names can be "warn","low","high","debug","info" 
`),
		),
		mcp.WithString("message_filter",
			mcp.Description(
				`message_filter specifies the search criteria for messages using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("limit_log_count",
			mcp.DefaultNumber(100),
			mcp.Max(10000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 100,max 10000"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host := makeRegexFilter(request.GetString("host_filter", ""))
		tag := makeRegexFilter(request.GetString("tag_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		message := makeRegexFilter(request.GetString("message_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := request.GetInt("limit_log_count", 100)
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// get_syslog_summary tool
type mcpSyslogSummaryEnt struct {
	Pattern string
	Count   int
}

func addGetSyslogSummaryTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_syslog_summary",
		mcp.WithDescription("get syslog summary from TWSNMP"),
		mcp.WithString("host_filter",
			mcp.Description(
				`host_filter specifies the search criteria for host names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("tag_filter",
			mcp.Description(
				`tag_filter specifies the search criteria for tag names using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("level_filter",
			mcp.Description(
				`level_filter specifies the search criteria for level names using regular expressions.
If blank, no filter.
Level names can be "warn","low","high","debug","info" 
`),
		),
		mcp.WithString("message_filter",
			mcp.Description(
				`message_filter specifies the search criteria for messages using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("top_n",
			mcp.DefaultNumber(10),
			mcp.Max(100),
			mcp.Min(5),
			mcp.Description("Top n syslog pattern. min 5,max 100"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		host := makeRegexFilter(request.GetString("host_filter", ""))
		tag := makeRegexFilter(request.GetString("tag_filter", ""))
		level := makeRegexFilter(request.GetString("level_filter", ""))
		message := makeRegexFilter(request.GetString("message_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		topN := request.GetInt("top_n", 10)
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

var regNum = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
var regUUDI = regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
var regEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
var regIP = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
var regMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b`)

func normalizeLog(s string) string {
	s = regUUDI.ReplaceAllString(s, "#UUID#")
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

func addSearchSNMPTrapLogTool(s *server.MCPServer) {
	tool := mcp.NewTool("search_snmp_trap_log",
		mcp.WithDescription("search SNMP trap log from TWSNMP"),
		mcp.WithString("from_filter",
			mcp.Description(
				`from_filter specifies the search criteria for trap sender address using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("trap_type_filter",
			mcp.Description(
				`trap_type_filter specifies the search criteria for SNMP trap types using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithString("variable_filter",
			mcp.Description(
				`variable_filter specifies the search criteria for SNMP trap variables using regular expressions.
If blank, no filter.
`),
		),
		mcp.WithNumber("limit_log_count",
			mcp.DefaultNumber(100),
			mcp.Max(10000),
			mcp.Min(1),
			mcp.Description("Limit on number of logs retrieved. min 100,max 10000"),
		),
		mcp.WithString("start_time",
			mcp.DefaultString("-1h"),
			mcp.Description(
				`start date and time of logs to search
or duration from now

A duration string is a possibly signed sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "-1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".

Example:
 2025/05/07 05:59:00
 -1h
`),
		),
		mcp.WithString("end_time",
			mcp.DefaultString(""),
			mcp.Description(
				`end date and time of logs to search.
empty or "now" is current time.

Example:
 2025/05/07 06:59:00
 now
`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		from := makeRegexFilter(request.GetString("host_filter", ""))
		trapType := makeRegexFilter(request.GetString("trap_type_filter", ""))
		variable := makeRegexFilter(request.GetString("variable_filter", ""))
		start := request.GetString("start_time", "-1h")
		end := request.GetString("end_time", "")
		st, et, err := getTimeRange(start, end)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}
		limit := request.GetInt("limit_log_count", 100)
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
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

func addGetServerCertificateListTool(s *server.MCPServer) {
	tool := mcp.NewTool("get_server_certificate_list",
		mcp.WithDescription("get server certificate list from TWSNMP"),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			j = []byte(err.Error())
		}
		return mcp.NewToolResultText(string(j)), nil
	})
}

// add_event_log tool
func addAddEventLogTool(s *server.MCPServer) {
	searchTool := mcp.NewTool("add_event_log",
		mcp.WithDescription("add event log to TWSNMP"),
		mcp.WithString("level",
			mcp.Enum("info", "normal", "warn", "low", "high"),
			mcp.Description("Level of event (info,normal,warn,low,high)"),
		),
		mcp.WithString("node",
			mcp.Description("Node name associated with the event"),
		),
		mcp.WithString("event",
			mcp.Description("Event log contents"),
		),
	)
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		level := request.GetString("level", "info")
		event := request.GetString("event", "")
		node := request.GetString("node", "")
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
		return mcp.NewToolResultText("ok"), nil
	})
}

// getTimeRange
func getTimeRange(start, end string) (int64, int64, error) {
	var st time.Time
	var err error
	et := time.Now()
	if start == "" {
		return 0, 0, fmt.Errorf("start_time must not empty")
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
		return 0, 0, fmt.Errorf("start_time must before end_time")
	}
	return st.UnixNano(), et.UnixNano(), nil
}

func nameToOID(name string) string {
	oid := datastore.MIBDB.NameToOID(name)
	if oid == ".1" {
		oid = ".1.3"
	}
	if oid == ".0.0" {
		if matched, _ := regexp.MatchString(`\.[0-9.]+`, name); matched {
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
