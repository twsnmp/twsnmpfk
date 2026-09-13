package datastore

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	cachedGateway   string
	cachedGatewayAt time.Time
	gatewayLock     sync.RWMutex
)

// GetDefaultGateway returns the IPv4 default gateway address of the host.
func GetDefaultGateway() string {
	gatewayLock.RLock()
	if cachedGateway != "" && time.Since(cachedGatewayAt) < time.Minute {
		gw := cachedGateway
		gatewayLock.RUnlock()
		return gw
	}
	gatewayLock.RUnlock()

	gatewayLock.Lock()
	defer gatewayLock.Unlock()

	// Double check
	if cachedGateway != "" && time.Since(cachedGatewayAt) < time.Minute {
		return cachedGateway
	}

	gw := discoverDefaultGateway()
	if gw != "" {
		cachedGateway = gw
		cachedGatewayAt = time.Now()
	}
	return gw
}

func discoverDefaultGateway() string {
	switch runtime.GOOS {
	case "darwin":
		return getDarwinGateway()
	case "linux":
		return getLinuxGateway()
	case "windows":
		return getWindowsGateway()
	default:
		return ""
	}
}

// getDarwinGateway parses `netstat -rn -f inet` output on macOS.
func getDarwinGateway() string {
	cmd := exec.Command("netstat", "-rn", "-f", "inet")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "default" {
			gw := fields[1]
			if ip := net.ParseIP(gw); ip != nil && ip.To4() != nil {
				return gw
			}
		}
	}
	return ""
}

// getLinuxGateway parses /proc/net/route or `ip route`.
func getLinuxGateway() string {
	if f, err := os.Open("/proc/net/route"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 3 && fields[1] == "00000000" { // Destination 0.0.0.0
				gwHex := fields[2]
				if len(gwHex) == 8 {
					b, err := hex.DecodeString(gwHex)
					if err == nil && len(b) == 4 {
						// /proc/net/route stores IP in little-endian
						ip := net.IPv4(b[3], b[2], b[1], b[0])
						if !ip.IsUnspecified() {
							return ip.String()
						}
					}
				}
			}
		}
	}
	// Fallback to `ip route`
	cmd := exec.Command("ip", "route")
	out, err := cmd.Output()
	if err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) >= 3 && fields[0] == "default" && fields[1] == "via" {
				if ip := net.ParseIP(fields[2]); ip != nil && ip.To4() != nil {
					return fields[2]
				}
			}
		}
	}
	return ""
}

// getWindowsGateway parses `route print 0.0.0.0` on Windows.
func getWindowsGateway() string {
	cmd := exec.Command("route", "print", "0.0.0.0")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 5 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
			gw := fields[2]
			if ip := net.ParseIP(gw); ip != nil && ip.To4() != nil && !ip.IsUnspecified() {
				return gw
			}
		}
	}
	return ""
}

// SetDefaultGatewayForTest sets cached gateway for unit tests.
func SetDefaultGatewayForTest(gw string) {
	gatewayLock.Lock()
	defer gatewayLock.Unlock()
	cachedGateway = gw
	cachedGatewayAt = time.Now()
}
