// Package stun : STUN client implementation for resolving mapped public IP/port
package stun

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultServer  = "stun.cloudflare.com:3478"
	DefaultPort    = "3478"
	DefaultTimeout = 5 * time.Second
)

// Result represents the result of a STUN lookup
type Result struct {
	IP        string        `json:"ip"`
	Port      uint16        `json:"port"`
	Hostname  string        `json:"hostname"`
	LocalIP   string        `json:"localIP"`
	LocalPort uint16        `json:"localPort"`
	RTT       time.Duration `json:"rtt"`
	RTTNano   int64         `json:"rttNano"`
	Server    string        `json:"server"`
	Network   string        `json:"network"`
}

// Query performs a STUN Binding Request and returns the mapped address information.
func Query(server string, network string, timeout time.Duration) (*Result, error) {
	if server == "" {
		server = DefaultServer
	}
	if !strings.Contains(server, ":") {
		server = net.JoinHostPort(server, DefaultPort)
	}
	if network == "" {
		network = "udp4"
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	conn, err := net.DialTimeout(network, server, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to STUN server (%s): %w", server, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	var localIP string
	var localPort uint16
	if lAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		localIP = lAddr.IP.String()
		localPort = uint16(lAddr.Port)
	}

	req := make([]byte, 20)
	binary.BigEndian.PutUint16(req[0:2], 0x0001)     // Binding Request
	binary.BigEndian.PutUint16(req[2:4], 0x0000)     // Message Length (0 attributes for request)
	binary.BigEndian.PutUint32(req[4:8], 0x2112A442) // Magic Cookie
	if _, err := rand.Read(req[8:20]); err != nil {
		return nil, fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	startTime := time.Now()
	if _, err := conn.Write(req); err != nil {
		return nil, fmt.Errorf("failed to send STUN request: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read STUN response: %w", err)
	}
	rtt := time.Since(startTime)

	ip, port, err := parseSTUNResponse(buf[:n], req[4:8], req[8:20])
	if err != nil {
		return nil, err
	}

	hostname := lookupHostname(ip, timeout)

	return &Result{
		IP:        ip,
		Port:      port,
		Hostname:  hostname,
		LocalIP:   localIP,
		LocalPort: localPort,
		RTT:       rtt,
		RTTNano:   rtt.Nanoseconds(),
		Server:    server,
		Network:   network,
	}, nil
}

func parseSTUNResponse(buf []byte, magicCookie []byte, transactionID []byte) (string, uint16, error) {
	if len(buf) < 20 {
		return "", 0, fmt.Errorf("STUN response too short (%d bytes)", len(buf))
	}

	msgType := binary.BigEndian.Uint16(buf[0:2])
	if msgType != 0x0101 { // Binding Response (success)
		return "", 0, fmt.Errorf("unexpected STUN message type: 0x%04x", msgType)
	}

	msgLen := int(binary.BigEndian.Uint16(buf[2:4]))
	if len(buf) < 20+msgLen {
		return "", 0, fmt.Errorf("invalid STUN message length header")
	}

	pos := 20
	end := 20 + msgLen

	var mappedIP string
	var mappedPort uint16
	var found bool

	for pos+4 <= end {
		attrType := binary.BigEndian.Uint16(buf[pos : pos+2])
		attrLen := int(binary.BigEndian.Uint16(buf[pos+2 : pos+4]))
		pos += 4

		if pos+attrLen > end {
			break
		}

		attrData := buf[pos : pos+attrLen]

		// 0x0020: XOR-MAPPED-ADDRESS (RFC 5389)
		if attrType == 0x0020 && attrLen >= 4 {
			family := attrData[1]
			port := binary.BigEndian.Uint16(attrData[2:4]) ^ binary.BigEndian.Uint16(magicCookie[0:2])

			if family == 0x01 && attrLen >= 8 { // IPv4
				ip := net.IPv4(
					attrData[4]^magicCookie[0],
					attrData[5]^magicCookie[1],
					attrData[6]^magicCookie[2],
					attrData[7]^magicCookie[3],
				)
				return ip.String(), port, nil
			} else if family == 0x02 && attrLen >= 20 { // IPv6
				xorKey := append(magicCookie, transactionID...)
				ipBytes := make([]byte, 16)
				for i := 0; i < 16; i++ {
					ipBytes[i] = attrData[4+i] ^ xorKey[i]
				}
				return net.IP(ipBytes).String(), port, nil
			}
		}

		// 0x0001: MAPPED-ADDRESS (RFC 3489) - fallback if XOR-MAPPED-ADDRESS is not present
		if attrType == 0x0001 && attrLen >= 4 && !found {
			family := attrData[1]
			port := binary.BigEndian.Uint16(attrData[2:4])
			if family == 0x01 && attrLen >= 8 { // IPv4
				mappedIP = net.IP(attrData[4:8]).String()
				mappedPort = port
				found = true
			} else if family == 0x02 && attrLen >= 20 { // IPv6
				mappedIP = net.IP(attrData[4:20]).String()
				mappedPort = port
				found = true
			}
		}

		pad := (4 - (attrLen % 4)) % 4
		pos += attrLen + pad
	}

	if found {
		return mappedIP, mappedPort, nil
	}

	return "", 0, fmt.Errorf("mapped address attribute not found in response")
}

func lookupHostname(ip string, timeout time.Duration) string {
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	names, err := r.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	var cleaned []string
	for _, n := range names {
		cleaned = append(cleaned, strings.TrimSuffix(n, "."))
	}
	return strings.Join(cleaned, ", ")
}
