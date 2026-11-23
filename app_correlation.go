package main

import (
	"context"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfk/backend"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	corrCancelMu sync.Mutex
	corrCancel   context.CancelFunc
)

// GetCorrelationCandidates returns all pollings with logging enabled and their numeric variable keys.
func (a *App) GetCorrelationCandidates() []backend.CorrelationCandidate {
	return backend.GetCorrelationCandidates()
}

// AnalyzePollingCorrelation performs correlation analysis between two pollings.
func (a *App) AnalyzePollingCorrelation(req backend.CorrelationPairReq) backend.CorrelationPairRes {
	return backend.AnalyzePollingCorrelation(req)
}

// DiscoverPollingCorrelations automatically discovers correlated polling pairs with progress events.
func (a *App) DiscoverPollingCorrelations(req backend.CorrelationDiscoverReq) backend.CorrelationDiscoverRes {
	baseCtx := a.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	corrCancelMu.Lock()
	if corrCancel != nil {
		corrCancel()
	}
	ctx, cancel := context.WithCancel(baseCtx)
	corrCancel = cancel
	corrCancelMu.Unlock()

	defer func() {
		cancel()
		corrCancelMu.Lock()
		corrCancel = nil
		corrCancelMu.Unlock()
	}()

	var emitMu sync.Mutex
	lastEmit := time.Time{}
	lastPhase := ""

	return backend.DiscoverPollingCorrelationsWithProgress(ctx, req, func(p backend.CorrelationProgress) {
		if a.ctx == nil {
			return
		}
		emitMu.Lock()
		now := time.Now()
		shouldEmit := p.Phase != lastPhase || p.Current >= p.Total || now.Sub(lastEmit) >= 120*time.Millisecond
		if shouldEmit {
			lastEmit = now
			lastPhase = p.Phase
		}
		emitMu.Unlock()
		if shouldEmit {
			wails.EventsEmit(a.ctx, "correlation_progress", p)
		}
	})
}

// CancelCorrelationDiscovery cancels an in-progress correlation discovery.
func (a *App) CancelCorrelationDiscovery() bool {
	corrCancelMu.Lock()
	defer corrCancelMu.Unlock()
	if corrCancel != nil {
		corrCancel()
		return true
	}
	return false
}
