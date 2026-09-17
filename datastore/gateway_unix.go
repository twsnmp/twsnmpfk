//go:build !windows

package datastore

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/twsnmp/twsnmpfk/cmd"
)

func discoverDefaultGateway() string {
	switch runtime.GOOS {
	case "darwin":
		return getDarwinGateway()
	case "linux":
		return getLinuxGateway()
	default:
		return ""
	}
}

// getDarwinGateway parses `netstat -rn -f inet` output on macOS.
func getDarwinGateway() string {
	c := cmd.GetCmd("netstat", []string{"-rn", "-f", "inet"})
	out, err := c.Output()
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
	c := cmd.GetCmd("ip", []string{"route"})
	out, err := c.Output()
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
