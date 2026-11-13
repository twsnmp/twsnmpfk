package backend

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

// PollingControlHooks decouples backend MCP polling control from the polling package to avoid circular imports.
type PollingControlHooks struct {
	CheckPolling func(id string) bool
	CheckNode    func(nodeID string)
	CheckAll     func()
	StopPolling  func(pe *datastore.PollingEnt)
}

var (
	stopMCPCh           = make(chan bool, 1)
	mcpAllow            sync.Map
	pollingHooks        atomic.Pointer[PollingControlHooks]
	mcpConfigChangedSeq atomic.Uint64
)

// MCPConfigChangedSeq returns the monotonically increasing count of NotifyMCPConfigChanged calls.
func MCPConfigChangedSeq() uint64 {
	return mcpConfigChangedSeq.Load()
}

// SetPollingControlHooks registers callbacks to trigger immediate polling checks from backend MCP tools.
func SetPollingControlHooks(h PollingControlHooks) {
	pollingHooks.Store(&h)
}

type mcpStatus struct {
	status string
	errStr string
}

var (
	mcpStatusLock sync.RWMutex
	currMCPStatus mcpStatus
)

// GetMCPStatus returns the current runtime status of the MCP server ("running", "error", "stopped") and error string if any.
func GetMCPStatus() (string, string) {
	mcpStatusLock.RLock()
	defer mcpStatusLock.RUnlock()
	return currMCPStatus.status, currMCPStatus.errStr
}

func setMCPStatus(status, errStr string) {
	mcpStatusLock.Lock()
	defer mcpStatusLock.Unlock()
	currMCPStatus = mcpStatus{
		status: status,
		errStr: errStr,
	}
}

func mcpServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	timer := time.NewTicker(time.Second * 5)
	var e *echo.Echo
	stopMCPServer := func() {
		setMCPStatus("stopped", "")
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
				setMCPAllow()
				e = startMCPServer()
				if e != nil {
					log.Println("start mcp server")
					datastore.AddEventLog(&datastore.EventLogEnt{
						Type:  "system",
						Level: "info",
						Event: fmt.Sprintf(i18n.Trans("Start MCP server: transport=%s endpoint=%s"), datastore.MapConf.MCPTransport, datastore.MapConf.MCPEndpoint),
					})
				}
			} else if datastore.MapConf.MCPTransport == "off" && e != nil {
				stopMCPServer()
			}
		}
	}
}

func NotifyMCPConfigChanged() {
	mcpConfigChangedSeq.Add(1)
	select {
	case stopMCPCh <- true:
	default:
	}
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
	sv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
	}
	sv.Addr = datastore.MapConf.MCPEndpoint

	// Test endpoint listen
	ln, err := net.Listen("tcp", sv.Addr)
	if err != nil {
		log.Printf("start mcp server bind err=%v", err)
		setMCPStatus("error", err.Error())
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "warn",
			Event: fmt.Sprintf("MCP server start failed: %v", err),
		})
		return nil
	}
	ln.Close()
	setMCPStatus("running", "")

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
		if err := e.StartServer(sv); err != nil && err != http.ErrServerClosed {
			log.Printf("start mcp server err=%v", err)
			setMCPStatus("error", err.Error())
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
	mcpAllow.Clear()
	for _, ip := range strings.Split(datastore.MapConf.MCPFrom, ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			mcpAllow.Store(ip, true)
		}
	}
}

func checkMCPACL(c echo.Context) bool {
	if datastore.MapConf.MCPToken != "" {
		t := strings.TrimSpace(c.Request().Header.Get("Authorization"))
		if len(t) >= 7 && strings.EqualFold(t[:7], "Bearer ") {
			t = strings.TrimSpace(t[7:])
		}
		if subtle.ConstantTimeCompare([]byte(t), []byte(datastore.MapConf.MCPToken)) != 1 {
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
