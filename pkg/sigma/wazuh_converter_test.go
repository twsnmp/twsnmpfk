package sigma

import (
	"testing"

	"github.com/bradleyjkemp/sigma-go"
)

const sampleWazuhSSHXML = `
<group name="syslog,sshd,">
  <rule id="5700" level="0" noalert="1">
    <decoded_as>sshd</decoded_as>
    <description>SSHD messages grouped.</description>
  </rule>

  <rule id="5710" level="5">
    <if_sid>5700</if_sid>
    <match>illegal user|invalid user</match>
    <description>sshd: Attempt to login using a non-existent user</description>
    <mitre>
      <id>T1110</id>
    </mitre>
    <group>authentication_failed,pci_dss_10.2.4,</group>
  </rule>

  <rule id="5712" level="10" frequency="6" timeframe="120">
    <if_matched_sid>5710</if_matched_sid>
    <same_source_ip />
    <description>sshd: brute force trying to get access to the system.</description>
    <mitre>
      <id>T1110</id>
    </mitre>
    <group>authentication_failures,</group>
  </rule>
</group>
`

const sampleWazuhDecoderXML = `
<decoder name="sshd">
  <program_name>^sshd</program_name>
</decoder>

<decoder name="sshd-success">
  <parent>sshd</parent>
  <prematch>^Accepted</prematch>
  <regex offset="after_prematch">^ \S+ for (\S+) from (\S+) port (\d+)</regex>
  <order>user, srcip, srcport</order>
</decoder>
`

func TestParseWazuhRulesXML(t *testing.T) {
	rules, err := ParseWazuhRulesXML([]byte(sampleWazuhSSHXML))
	if err != nil {
		t.Fatalf("ParseWazuhRulesXML error: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	if rules[0].ID != "5700" || rules[1].ID != "5710" || rules[2].ID != "5712" {
		t.Errorf("unexpected rule IDs: %+v", rules)
	}
}

func TestConvertWazuhRuleToSigma(t *testing.T) {
	rules, err := ParseWazuhRulesXML([]byte(sampleWazuhSSHXML))
	if err != nil {
		t.Fatal(err)
	}

	resolved := ResolveRuleHierarchy(rules, "sshd")
	if len(resolved) != 3 {
		t.Fatalf("expected 3 resolved rules, got %d", len(resolved))
	}

	// 1. Rule 5700 has noalert="1", should return nil
	out5700, err := ConvertWazuhRuleToSigma(resolved[0], WazuhConvertOptions{MinLevel: 1})
	if err != nil || out5700 != nil {
		t.Errorf("expected nil for 5700 (noalert), got %+v, err=%v", out5700, err)
	}

	// 2. Rule 5710 has level=5 -> medium, match="illegal user|invalid user"
	out5710, err := ConvertWazuhRuleToSigma(resolved[1], WazuhConvertOptions{MinLevel: 1})
	if err != nil || out5710 == nil {
		t.Fatalf("expected valid rule for 5710, got err=%v", err)
	}
	if out5710.ID != "wazuh-5710" {
		t.Errorf("expected ID wazuh-5710, got %s", out5710.ID)
	}
	if out5710.Level != "medium" {
		t.Errorf("expected level medium, got %s", out5710.Level)
	}
	if out5710.Correlation != nil {
		t.Errorf("expected no correlation for 5710, got %+v", out5710.Correlation)
	}

	// Verify it serializes to valid Sigma YAML that sigma-go can parse
	yamlData, err := FormatSigmaYAML(out5710)
	if err != nil {
		t.Fatalf("FormatSigmaYAML error: %v", err)
	}
	parsedRule, err := sigma.ParseRule(yamlData)
	if err != nil {
		t.Fatalf("sigma.ParseRule failed on generated YAML: %v\nYAML:\n%s", err, string(yamlData))
	}
	if parsedRule.ID != "wazuh-5710" {
		t.Errorf("parsed rule ID mismatch: %s", parsedRule.ID)
	}

	// 3. Rule 5712 has frequency=6, timeframe=120, same_source_ip, inherits match from 5710
	out5712, err := ConvertWazuhRuleToSigma(resolved[2], WazuhConvertOptions{MinLevel: 1})
	if err != nil || out5712 == nil {
		t.Fatalf("expected valid rule for 5712, got err=%v", err)
	}
	if out5712.Level != "high" {
		t.Errorf("expected level high for level 10, got %s", out5712.Level)
	}
	if out5712.Correlation == nil {
		t.Fatalf("expected correlation metadata for 5712, got nil")
	}
	if out5712.Correlation.Frequency != 6 || out5712.Correlation.Timeframe != "120s" {
		t.Errorf("unexpected correlation: %+v", out5712.Correlation)
	}
	if len(out5712.Correlation.GroupBy) != 1 || out5712.Correlation.GroupBy[0] != "client" {
		t.Errorf("expected group_by: [client], got %+v", out5712.Correlation.GroupBy)
	}

	// Verify that 5712 also parses via sigma-go and preserves correlation in AdditionalFields
	yaml5712, err := FormatSigmaYAML(out5712)
	if err != nil {
		t.Fatalf("FormatSigmaYAML error: %v", err)
	}
	parsed5712, err := sigma.ParseRule(yaml5712)
	if err != nil {
		t.Fatalf("sigma.ParseRule failed on 5712: %v", err)
	}
	if parsed5712.AdditionalFields["correlation"] == nil {
		t.Errorf("expected correlation in AdditionalFields, got nil")
	}
}

func TestConvertWazuhDecoderToRegex(t *testing.T) {
	decoders, err := ParseWazuhDecodersXML([]byte(sampleWazuhDecoderXML))
	if err != nil {
		t.Fatalf("ParseWazuhDecodersXML error: %v", err)
	}
	if len(decoders) != 2 {
		t.Fatalf("expected 2 decoders, got %d", len(decoders))
	}

	reStr, err := ConvertDecoderToNamedRegex(decoders[1])
	if err != nil {
		t.Fatalf("ConvertDecoderToNamedRegex error: %v", err)
	}

	expectedPrefix := "^ \\S+ for (?P<user>\\S+) from (?P<client>\\S+) port (?P<srcport>\\d+)"
	if reStr != expectedPrefix {
		t.Errorf("expected regex %q, got %q", expectedPrefix, reStr)
	}
}
