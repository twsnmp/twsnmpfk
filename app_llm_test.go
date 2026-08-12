package main

import (
	"testing"
)

func TestCleanPollingName(t *testing.T) {
	tests := []struct {
		event    string
		expected string
	}{
		{"Change polling state: Ping(ping)", "Ping"},
		{"ポーリング状態変化: Ping(ping)", "Ping"},
		{"Change polling state: Ping(ping) [Downtime: 5m 20s]", "Ping"},
		{"ポーリング状態変化: Ping(ping) [ダウンタイム: 5m 20s]", "Ping"},
		{"Change polling state: SNMP Poll(snmp) [Downtime: 1h]", "SNMP Poll"},
	}

	for _, tt := range tests {
		got := cleanPollingName(tt.event)
		if got != tt.expected {
			t.Errorf("cleanPollingName(%q) = %q; want %q", tt.event, got, tt.expected)
		}
	}
}
