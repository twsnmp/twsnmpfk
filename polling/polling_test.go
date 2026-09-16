package polling

import (
	"testing"
)

func TestFormatDowntime(t *testing.T) {
	tests := []struct {
		sec      int64
		expected string
	}{
		{0, "0s"},
		{15, "15s"},
		{59, "59s"},
		{60, "1m"},
		{320, "5m 20s"},
		{3600, "1h"},
		{3660, "1h 1m"},
		{86400, "1d"},
		{90000, "1d 1h"},
	}

	for _, tt := range tests {
		got := formatDowntime(tt.sec)
		if got != tt.expected {
			t.Errorf("formatDowntime(%d) = %q; want %q", tt.sec, got, tt.expected)
		}
	}
}

func TestParseSigmaFilter(t *testing.T) {
	// Empty or all -> nil, nil
	packs, rules := parseSigmaFilter("")
	if len(packs) != 0 || len(rules) != 0 {
		t.Errorf("expected empty filter to return nil/nil, got packs=%v rules=%v", packs, rules)
	}
	packs, rules = parseSigmaFilter("all")
	if len(packs) != 0 || len(rules) != 0 {
		t.Errorf("expected 'all' to return nil/nil, got packs=%v rules=%v", packs, rules)
	}

	// Known packs
	packs, rules = parseSigmaFilter("windows-essential, linux-auth")
	if len(packs) != 2 || len(rules) != 0 {
		t.Errorf("expected 2 packs and 0 rules, got packs=%v rules=%v", packs, rules)
	}

	// Specific rules
	packs, rules = parseSigmaFilter("lnx_sshd_failed_login, win_security_failed_logons")
	if len(packs) != 0 || len(rules) != 2 {
		t.Errorf("expected 0 packs and 2 rules, got packs=%v rules=%v", packs, rules)
	}

	// Mixed with prefixes
	packs, rules = parseSigmaFilter("pack:network-threats, rule:my_custom_rule")
	if len(packs) != 1 || len(rules) != 1 || packs[0] != "network-threats" || rules[0] != "my_custom_rule" {
		t.Errorf("expected 1 pack and 1 rule with prefixes, got packs=%v rules=%v", packs, rules)
	}
}

