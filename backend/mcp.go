package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

var stopMCPCh = make(chan bool)
var mcpAllow sync.Map

func mcpServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	timer := time.NewTicker(time.Second * 5)
	var e *echo.Echo
	stopMCPServer := func() {
		if e == nil {
			return
		}
		log.Println("stop mcp server")
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: i18n.Trans("Stop MCP server"),
		})
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
			if datastore.MapConf.MCPTransport != "off" && e == nil {
				log.Println("start mcp server")
				setMCPAllow()
				e = startMCPServer()
				datastore.AddEventLog(&datastore.EventLogEnt{
					Type:  "system",
					Level: "info",
					Event: fmt.Sprintf(i18n.Trans("Start MCP server: transport=%s endpoint=%s"), datastore.MapConf.MCPTransport, datastore.MapConf.MCPEndpoint),
				})
			} else if datastore.MapConf.MCPTransport == "off" && e != nil {
				stopMCPServer()
			}
		}
	}
}

func NotifyMCPConfigChanged() {
	stopMCPCh <- true
}

func startMCPServer() *echo.Echo {
	// Create MCP Server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "TWSNMP FK MCP Server",
			Version: version,
		},
		nil)
	// Add tools to MCP server
	addTools(s)
	// Add prompts to MCP server
	addPrompts(s)
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
	if datastore.MapConf.MCPTransport == "sse" {
		handler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			return s
		}, nil)
		e.Any("/sse", func(c echo.Context) error {
			if !checkMCPACL(c) {
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/message", func(c echo.Context) error {
			if !checkMCPACL(c) {
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})

		log.Printf("sse mcp server listening on %s", datastore.MapConf.MCPEndpoint)
	} else {
		// Create the streamable HTTP handler.
		handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
			return s
		}, nil)
		e.Any("/mcp", func(c echo.Context) error {
			if !checkMCPACL(c) {
				log.Printf("mcp reject connection from %s", c.Request().RemoteAddr)
				return echo.ErrUnauthorized
			}
			handler.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		})
		log.Printf("streamable HTTP server listening on %s", datastore.MapConf.MCPEndpoint)
	}
	go func() {
		if err := e.StartServer(sv); err != nil {
			log.Printf("start mcp server err=%v", err)
		}
	}()
	return e
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
