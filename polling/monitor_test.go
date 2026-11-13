package polling

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestPollingControlHooks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	td := t.TempDir()
	if err := datastore.Init(ctx, td, wg); err != nil {
		cancel()
		t.Fatalf("datastore.Init failed: %v", err)
	}
	origCh := doPollingCh
	doPollingCh = nil
	t.Cleanup(func() {
		doPollingCh = origCh
		cancel()
		wg.Wait()
		datastore.CloseDB()
	})

	// nil doPollingCh and nil polling entry must be safe no-ops.
	enqueuePolling("nonexistent")
	StopPollingSubscription(nil)
	StopPollingSubscription(&datastore.PollingEnt{ID: "missing", Type: "gnmi", Mode: "subscribe"})
	StopPollingSubscription(&datastore.PollingEnt{ID: "missing", Type: "mqtt", Mode: ""})

	// Buffered channel must enqueue without blocking when full.
	doPollingCh = make(chan string, 1)
	enqueuePolling("first")
	enqueuePolling("overflow")
	if got := <-doPollingCh; got != "first" {
		t.Fatalf("expected queued id %q, got %q", "first", got)
	}

	n := &datastore.NodeEnt{Name: "poll-node", IP: "192.0.2.50", State: "normal"}
	if err := datastore.AddNode(n); err != nil {
		t.Fatalf("AddNode failed: %v", err)
	}
	p := &datastore.PollingEnt{
		NodeID:   n.ID,
		Name:     "ping",
		Type:     "ping",
		State:    "normal",
		NextTime: time.Now().Add(time.Hour).UnixNano(),
	}
	if err := datastore.AddPolling(p); err != nil {
		t.Fatalf("AddPolling failed: %v", err)
	}

	if PollNow("missing-id") {
		t.Fatalf("expected PollNow on missing polling id to return false")
	}
	if !PollNow(p.ID) {
		t.Fatalf("expected PollNow on existing polling id to return true")
	}
	gotP := datastore.GetPolling(p.ID)
	if gotP == nil || gotP.State != "unknown" || gotP.NextTime != 0 {
		t.Fatalf("expected polling reset to unknown with NextTime=0, got %+v", gotP)
	}
	gotN := datastore.GetNode(n.ID)
	if gotN == nil || gotN.State != "unknown" {
		t.Fatalf("expected node state reset to unknown, got %+v", gotN)
	}
	select {
	case queued := <-doPollingCh:
		if queued != p.ID {
			t.Fatalf("expected %q on doPollingCh, got %q", p.ID, queued)
		}
	default:
		t.Fatalf("expected PollNow to enqueue polling ID onto doPollingCh")
	}
}
