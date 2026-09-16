package main

import (
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestApp_AnalyzeSigmaLogs(t *testing.T) {
	app := &App{}

	now := time.Now().UnixNano()
	logs := []*datastore.SyslogEnt{
		{
			Time:     now - 3000,
			Host:     "192.168.1.50",
			Tag:      "sshd",
			Message:  "Failed password for invalid user admin from 192.168.1.100 port 55432 ssh2",
			Severity: 4,
			Facility: 4,
		},
		{
			Time:     now - 2000,
			Host:     "192.168.1.50",
			Tag:      "sudo",
			Message:  "pam_unix(sudo:auth): authentication failure; logname=alice uid=1001 euid=0 tty=/dev/pts/1 ruser=alice rhost= user=alice",
			Severity: 4,
			Facility: 10,
		},
		{
			Time:     now - 1000,
			Host:     "192.168.1.10",
			Tag:      "kernel",
			Message:  "normal kernel message",
			Severity: 6,
			Facility: 0,
		},
	}

	// 1. Analyze with all packs
	res, err := app.AnalyzeSigmaLogs(logs, []string{}, []string{})
	if err != nil {
		t.Fatalf("AnalyzeSigmaLogs error: %v", err)
	}

	if res.Stats.TotalLogs != 3 {
		t.Errorf("expected TotalLogs=3, got %d", res.Stats.TotalLogs)
	}
	if res.Stats.TotalDetections < 1 {
		t.Errorf("expected at least 1 detection, got %d", res.Stats.TotalDetections)
	}
	if res.Stats.HitLogs < 1 {
		t.Errorf("expected at least 1 hit log, got %d", res.Stats.HitLogs)
	}

	// 2. Analyze with specific rule filter
	resFiltered, err := app.AnalyzeSigmaLogs(logs, []string{"linux-auth"}, []string{"lnx_sshd_failed_login"})
	if err != nil {
		t.Fatalf("AnalyzeSigmaLogs filtered error: %v", err)
	}
	if resFiltered.Stats.ActiveRules != 1 {
		t.Errorf("expected 1 active rule, got %d", resFiltered.Stats.ActiveRules)
	}

	// 3. Check GetSigmaPacks
	packs := app.GetSigmaPacks()
	if len(packs) < 10 {
		t.Errorf("expected at least 10 packs, got %d", len(packs))
	}
}
