package main

import (
	"context"

	"github.com/twsnmp/twsnmpfk/backend"
)

// GetCorrelationCandidates returns all pollings with logging enabled and their numeric variable keys.
func (a *App) GetCorrelationCandidates() []backend.CorrelationCandidate {
	return backend.GetCorrelationCandidates()
}

// AnalyzePollingCorrelation performs correlation analysis between two pollings.
func (a *App) AnalyzePollingCorrelation(req backend.CorrelationPairReq) backend.CorrelationPairRes {
	return backend.AnalyzePollingCorrelation(req)
}

// DiscoverPollingCorrelations automatically discovers correlated polling pairs.
func (a *App) DiscoverPollingCorrelations(req backend.CorrelationDiscoverReq) backend.CorrelationDiscoverRes {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return backend.DiscoverPollingCorrelations(ctx, req)
}
