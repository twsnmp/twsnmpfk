package main

import (
	"context"
	"sync"
	"testing"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestUpdateMapConfMCPTransportChange(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	td := t.TempDir()
	if err := datastore.Init(ctx, td, wg); err != nil {
		cancel()
		t.Fatalf("datastore.Init failed: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		wg.Wait()
		datastore.CloseDB()
	})

	a := NewApp()
	m := datastore.MapConf
	m.MCPTransport = "stream"
	m.MCPEndpoint = "127.0.0.1:8181"
	if !a.UpdateMapConf(m) {
		t.Fatalf("initial UpdateMapConf failed")
	}
	seq := backend.MCPConfigChangedSeq()

	// Unchanged MCP settings must not emit a spurious reload signal.
	if !a.UpdateMapConf(datastore.MapConf) {
		t.Fatalf("no-op UpdateMapConf failed")
	}
	if got := backend.MCPConfigChangedSeq(); got != seq {
		t.Fatalf("expected MCPConfigChangedSeq=%d when MCP config unchanged, got %d", seq, got)
	}

	// Change ONLY MCPTransport while keeping endpoint, token, and allowlist identical.
	m2 := datastore.MapConf
	m2.MCPTransport = "sse"
	if !a.UpdateMapConf(m2) {
		t.Fatalf("UpdateMapConf on MCPTransport change failed")
	}
	if datastore.MapConf.MCPTransport != "sse" {
		t.Fatalf("expected datastore.MapConf.MCPTransport=%q, got %q", "sse", datastore.MapConf.MCPTransport)
	}
	if got := backend.MCPConfigChangedSeq(); got != seq+1 {
		t.Fatalf("expected MCPConfigChangedSeq=%d on MCPTransport change, got %d", seq+1, got)
	}

	// Also verify MCPFrom / MCPToken / MCPEndpoint trigger reloads individually.
	m3 := datastore.MapConf
	m3.MCPToken = "secret-token"
	if !a.UpdateMapConf(m3) {
		t.Fatalf("UpdateMapConf on MCPToken change failed")
	}
	if got := backend.MCPConfigChangedSeq(); got != seq+2 {
		t.Fatalf("expected MCPConfigChangedSeq=%d on MCPToken change, got %d", seq+2, got)
	}
}
