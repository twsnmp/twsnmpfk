package datastore

import (
	"context"
	"sync"
	"testing"
)

func TestCertMonitorGetSaveDelete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	td := t.TempDir()
	if err := Init(ctx, td, wg); err != nil {
		cancel()
		t.Fatalf("Init failed: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		wg.Wait()
		CloseDB()
	})

	// Missing entry must return an empty struct without panicking.
	missing := GetCertMonitor("non-existent-id")
	if missing == nil || missing.ID != "" {
		t.Fatalf("expected non-nil empty CertMonitorEnt for missing ID, got %+v", missing)
	}

	c := &CertMonitorEnt{
		Target: "example.org",
		Port:   443,
		State:  "normal",
	}
	if err := SaveCertMonitor(c); err != nil {
		t.Fatalf("SaveCertMonitor failed: %v", err)
	}
	if c.ID == "" {
		t.Fatalf("expected generated ID from SaveCertMonitor")
	}

	got := GetCertMonitor(c.ID)
	if got == nil || got.ID != c.ID || got.Target != "example.org" || got.Port != 443 {
		t.Fatalf("GetCertMonitor returned unexpected entry: %+v", got)
	}

	DeleteCertMonitor(c.ID)
	afterDel := GetCertMonitor(c.ID)
	if afterDel == nil || afterDel.ID != "" {
		t.Fatalf("expected empty CertMonitorEnt after deletion, got %+v", afterDel)
	}
}
