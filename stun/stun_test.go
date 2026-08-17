package stun

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestParseSTUNResponseIPv4XORMapped(t *testing.T) {
	magicCookie := []byte{0x21, 0x12, 0xA4, 0x42}
	transactionID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	// 20 bytes STUN header + 12 bytes XOR-MAPPED-ADDRESS attribute
	buf := make([]byte, 32)
	binary.BigEndian.PutUint16(buf[0:2], 0x0101) // Success response
	binary.BigEndian.PutUint16(buf[2:4], 12)     // Attribute length
	copy(buf[4:8], magicCookie)
	copy(buf[8:20], transactionID)

	// XOR-MAPPED-ADDRESS attribute (0x0020)
	binary.BigEndian.PutUint16(buf[20:22], 0x0020)
	binary.BigEndian.PutUint16(buf[22:24], 8) // Attr data len
	buf[24] = 0x00                           // Reserved
	buf[25] = 0x01                           // IPv4

	// Port 54321 XOR 0x2112
	port := uint16(54321)
	binary.BigEndian.PutUint16(buf[26:28], port^binary.BigEndian.Uint16(magicCookie[0:2]))

	// IP 203.0.113.195 XOR MagicCookie
	targetIP := net.ParseIP("203.0.113.195").To4()
	buf[28] = targetIP[0] ^ magicCookie[0]
	buf[29] = targetIP[1] ^ magicCookie[1]
	buf[30] = targetIP[2] ^ magicCookie[2]
	buf[31] = targetIP[3] ^ magicCookie[3]

	ip, parsedPort, err := parseSTUNResponse(buf, magicCookie, transactionID)
	if err != nil {
		t.Fatalf("parseSTUNResponse failed: %v", err)
	}

	if ip != "203.0.113.195" {
		t.Errorf("expected ip 203.0.113.195, got %s", ip)
	}
	if parsedPort != port {
		t.Errorf("expected port %d, got %d", port, parsedPort)
	}
}

func TestParseSTUNResponseIPv4Mapped(t *testing.T) {
	magicCookie := []byte{0x21, 0x12, 0xA4, 0x42}
	transactionID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	buf := make([]byte, 32)
	binary.BigEndian.PutUint16(buf[0:2], 0x0101)
	binary.BigEndian.PutUint16(buf[2:4], 12)
	copy(buf[4:8], magicCookie)
	copy(buf[8:20], transactionID)

	// MAPPED-ADDRESS attribute (0x0001)
	binary.BigEndian.PutUint16(buf[20:22], 0x0001)
	binary.BigEndian.PutUint16(buf[22:24], 8)
	buf[24] = 0x00
	buf[25] = 0x01 // IPv4

	port := uint16(12345)
	binary.BigEndian.PutUint16(buf[26:28], port)

	targetIP := net.ParseIP("198.51.100.1").To4()
	copy(buf[28:32], targetIP)

	ip, parsedPort, err := parseSTUNResponse(buf, magicCookie, transactionID)
	if err != nil {
		t.Fatalf("parseSTUNResponse failed: %v", err)
	}

	if ip != "198.51.100.1" {
		t.Errorf("expected ip 198.51.100.1, got %s", ip)
	}
	if parsedPort != port {
		t.Errorf("expected port %d, got %d", port, parsedPort)
	}
}
