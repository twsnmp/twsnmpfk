package backend

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestCorrelationAlgorithms(t *testing.T) {
	// 1. Perfect positive linear: y = 2x + 1
	n := 30
	x := make([]float64, n)
	yPos := make([]float64, n)
	yNeg := make([]float64, n)
	yCubic := make([]float64, n)
	for i := 0; i < n; i++ {
		v := float64(i + 1)
		x[i] = v
		yPos[i] = 2*v + 1
		yNeg[i] = -3*v + 5
		yCubic[i] = v * v * v
	}

	r, p := calcPearson(x, yPos)
	if math.Abs(r-1.0) > 1e-9 || p > 1e-6 {
		t.Fatalf("Pearson positive failed: r=%v p=%v", r, p)
	}
	rho, pS := calcSpearman(x, yPos)
	if math.Abs(rho-1.0) > 1e-9 || pS > 1e-6 {
		t.Fatalf("Spearman positive failed: rho=%v p=%v", rho, pS)
	}
	tau, pK := calcKendallTauB(x, yPos)
	if math.Abs(tau-1.0) > 1e-9 || pK > 1e-6 {
		t.Fatalf("Kendall positive failed: tau=%v p=%v", tau, pK)
	}
	dcor, pD := calcDistanceCorrelation([][]float64{x}, [][]float64{yPos}, 200, true)
	if dcor < 0.99 || pD > 0.05 {
		t.Fatalf("dCor positive failed: dcor=%v p=%v", dcor, pD)
	}

	// 2. Perfect negative linear: y = -3x + 5
	r, _ = calcPearson(x, yNeg)
	if math.Abs(r+1.0) > 1e-9 {
		t.Fatalf("Pearson negative failed: r=%v", r)
	}
	rho, _ = calcSpearman(x, yNeg)
	if math.Abs(rho+1.0) > 1e-9 {
		t.Fatalf("Spearman negative failed: rho=%v", rho)
	}
	tau, _ = calcKendallTauB(x, yNeg)
	if math.Abs(tau+1.0) > 1e-9 {
		t.Fatalf("Kendall negative failed: tau=%v", tau)
	}

	// 3. Monotonic nonlinear: y = x^3
	rho, _ = calcSpearman(x, yCubic)
	tau, _ = calcKendallTauB(x, yCubic)
	if math.Abs(rho-1.0) > 1e-9 || math.Abs(tau-1.0) > 1e-9 {
		t.Fatalf("Rank correlations failed on monotonic x^3: rho=%v tau=%v", rho, tau)
	}

	// 4. Symmetric nonlinear: y = x^2 over [-15..15] -> Pearson ~ 0, dCor > 0.45
	m := 31
	xSym := make([]float64, m)
	yQuad := make([]float64, m)
	for i := 0; i < m; i++ {
		v := float64(i - 15)
		xSym[i] = v
		yQuad[i] = v * v
	}
	rSym, _ := calcPearson(xSym, yQuad)
	dCorSym, pSym := calcDistanceCorrelation([][]float64{xSym}, [][]float64{yQuad}, 200, true)
	if math.Abs(rSym) > 1e-9 {
		t.Fatalf("expected ~0 Pearson for symmetric x^2, got %v", rSym)
	}
	if dCorSym < 0.45 || pSym > 0.05 {
		t.Fatalf("expected strong dCor for x^2, got dcor=%v p=%v", dCorSym, pSym)
	}

	// 5. Constant series (zero variance): must return 0 without NaN/Inf
	c := make([]float64, n)
	for i := range c {
		c[i] = 42.0
	}
	r0, p0 := calcPearson(x, c)
	rho0, _ := calcSpearman(x, c)
	tau0, _ := calcKendallTauB(x, c)
	dc0, _ := calcDistanceCorrelation([][]float64{x}, [][]float64{c}, 200, false)
	if math.IsNaN(r0) || r0 != 0 || p0 != 1.0 || rho0 != 0 || tau0 != 0 || dc0 != 0 {
		t.Fatalf("constant input failed: r=%v p=%v rho=%v tau=%v dc=%v", r0, p0, rho0, tau0, dc0)
	}

	// 6. Multivariate dCor
	dMulti, _ := calcDistanceCorrelation([][]float64{x, yPos}, [][]float64{yNeg, yCubic}, 200, false)
	if dMulti < 0.95 {
		t.Fatalf("expected high multivariate dCor, got %v", dMulti)
	}
}

func TestCrossCorrelationLag(t *testing.T) {
	n := 60
	x := make([]float64, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		// Non-periodic pulse + oscillation so lag peak is unique
		x[i] = math.Sin(float64(i)*0.25) + 0.1*float64(i)/float64(n)
		if i == 15 || i == 35 {
			x[i] += 5.0
		}
	}
	// y[t+3] = x[t] => x leads y by +3 buckets
	for i := 0; i < n; i++ {
		if i >= 3 {
			y[i] = x[i-3]
		} else {
			y[i] = 0
		}
	}

	score, _, lag, profile := calcCrossCorrelation(x, y, 8, 60, 10)
	if lag != 3 || score < 0.9 {
		t.Fatalf("expected lag=+3 and high score, got lag=%d score=%v (profile len=%d)", lag, score, len(profile))
	}

	// Reverse: y leads x by +2 => lag = -2 when calling (x, y) where x[t+2] = y[t]
	scoreRev, _, lagRev, _ := calcCrossCorrelation(y, x, 8, 60, 10)
	if lagRev != -3 || scoreRev < 0.9 {
		t.Fatalf("expected lag=-3 and high score on reversed args, got lag=%d score=%v", lagRev, scoreRev)
	}
}

func setupCorrelationTestEntities(t *testing.T) (cloudPollID, routerPollID, routerAuxPollID, constPollID string, startSec, endSec int64) {
	t.Helper()
	nCloud := &datastore.NodeEnt{Name: "Cloud-GW", IP: "10.0.1.1"}
	nRouter := &datastore.NodeEnt{Name: "Core-Router", IP: "10.0.2.1"}
	if err := datastore.AddNode(nCloud); err != nil {
		t.Fatalf("AddNode cloud: %v", err)
	}
	if err := datastore.AddNode(nRouter); err != nil {
		t.Fatalf("AddNode router: %v", err)
	}

	pCloud := &datastore.PollingEnt{
		NodeID:     nCloud.ID,
		Name:       "Cloud PING",
		Type:       "ping",
		PollInt:    60,
		LogMode:    datastore.LogModeAlways,
		VectorCols: "rtt,hour",
		Result:     map[string]any{"rtt": 12.0, "successRate": 1.0},
	}
	pRouter := &datastore.PollingEnt{
		NodeID:  nRouter.ID,
		Name:    "Router PING",
		Type:    "ping",
		PollInt: 60,
		LogMode: datastore.LogModeAI,
		Result:  map[string]any{"rtt": 5.0, "jitter": 1.0},
	}
	pRouterAux := &datastore.PollingEnt{
		NodeID:  nRouter.ID,
		Name:    "Router CPU",
		Type:    "snmp",
		PollInt: 60,
		LogMode: datastore.LogModeAlways,
		Result:  map[string]any{"cpu": 20.0},
	}
	pConst := &datastore.PollingEnt{
		NodeID:  nRouter.ID,
		Name:    "Router Constant",
		Type:    "ping",
		PollInt: 60,
		LogMode: datastore.LogModeAlways,
		Result:  map[string]any{"rtt": 1.0},
	}
	for _, p := range []*datastore.PollingEnt{pCloud, pRouter, pRouterAux, pConst} {
		if err := datastore.AddPolling(p); err != nil {
			t.Fatalf("AddPolling %s: %v", p.Name, err)
		}
	}

	now := time.Now().Unix()
	base := (now/60)*60 - 60*60 // 60 minutes ago aligned to minute
	logsMap := map[string][]*datastore.PollingLogEnt{}

	for i := 0; i < 40; i++ {
		tsNano := (base + int64(i)*60) * 1e9
		rttCloud := 10.0 + float64(i)*0.5 + math.Sin(float64(i)*0.3)*2.0
		rttRouter := 2.0*rttCloud + 3.0 // strongly correlated with Cloud PING rtt
		cpuRouter := 80.0 - float64(i)*1.2 + math.Cos(float64(i)*0.3)*3.0

		// Cloud has slight 15s offset and skips bucket i==10 to test alignment
		if i != 10 {
			logsMap[pCloud.ID] = append(logsMap[pCloud.ID], &datastore.PollingLogEnt{
				Time:      tsNano + 15*1e9,
				PollingID: pCloud.ID,
				State:     "normal",
				Result:    map[string]any{"rtt": rttCloud, "successRate": 1.0},
			})
		}
		logsMap[pRouter.ID] = append(logsMap[pRouter.ID], &datastore.PollingLogEnt{
			Time:      tsNano + 40*1e9,
			PollingID: pRouter.ID,
			State:     "normal",
			Result:    map[string]any{"rtt": rttRouter, "jitter": float64(i % 3)},
		})
		logsMap[pRouterAux.ID] = append(logsMap[pRouterAux.ID], &datastore.PollingLogEnt{
			Time:      tsNano + 5*1e9,
			PollingID: pRouterAux.ID,
			State:     "normal",
			Result:    map[string]any{"cpu": cpuRouter},
		})
		logsMap[pConst.ID] = append(logsMap[pConst.ID], &datastore.PollingLogEnt{
			Time:      tsNano,
			PollingID: pConst.ID,
			State:     "normal",
			Result:    map[string]any{"rtt": 1.0},
		})
	}

	orig := getPollingLogsFunc
	getPollingLogsFunc = func(pollingID string) []*datastore.PollingLogEnt {
		return logsMap[pollingID]
	}
	t.Cleanup(func() {
		getPollingLogsFunc = orig
	})

	return pCloud.ID, pRouter.ID, pRouterAux.ID, pConst.ID, base - 60, base + 45*60
}

func TestAlignAndDiscover(t *testing.T) {
	setupMCPTestDB(t)
	cloudID, routerID, routerAuxID, constID, st, et := setupCorrelationTestEntities(t)

	// 1. Candidates check
	cands := GetCorrelationCandidates()
	if len(cands) != 4 {
		t.Fatalf("expected 4 candidates, got %d", len(cands))
	}

	// 2. Pair analysis (Cloud PING vs Router PING across different nodes)
	pairRes := AnalyzePollingCorrelation(CorrelationPairReq{
		PollingID1: cloudID,
		PollingID2: routerID,
		Algorithm:  "pearson",
		BucketSec:  60,
		StartTime:  st,
		EndTime:    et,
	})
	if pairRes.Error != "" {
		t.Fatalf("AnalyzePollingCorrelation unexpected error: %s", pairRes.Error)
	}
	if pairRes.SampleCount != 39 {
		t.Fatalf("expected 39 overlapping samples (1 skipped out of 40), got %d", pairRes.SampleCount)
	}
	if pairRes.SelectedVar1 != "rtt" || pairRes.SelectedVar2 != "rtt" {
		t.Fatalf("expected selected vars (rtt, rtt), got (%s, %s)", pairRes.SelectedVar1, pairRes.SelectedVar2)
	}
	if math.Abs(pairRes.PrimaryScore-1.0) > 1e-6 {
		t.Fatalf("expected Pearson ~1.0, got %v", pairRes.PrimaryScore)
	}
	if math.Abs(pairRes.Slope-2.0) > 1e-4 || math.Abs(pairRes.Intercept-3.0) > 1e-4 {
		t.Fatalf("expected regression y=2x+3, got slope=%v intercept=%v", pairRes.Slope, pairRes.Intercept)
	}
	if len(pairRes.Scores) != 5 {
		t.Fatalf("expected 5 algorithm scores, got %d", len(pairRes.Scores))
	}
	if len(pairRes.NormSeries1) != pairRes.SampleCount || len(pairRes.NormSeries2) != pairRes.SampleCount {
		t.Fatalf("normalized series length mismatch")
	}

	// 3. Auto discovery: CrossNodeOnly = true
	discCross := DiscoverPollingCorrelations(context.Background(), CorrelationDiscoverReq{
		Algorithm:     "pearson",
		BucketSec:     60,
		StartTime:     st,
		EndTime:       et,
		MinAbsScore:   0.8,
		MinSamples:    10,
		CrossNodeOnly: true,
	})
	if discCross.Error != "" {
		t.Fatalf("DiscoverPollingCorrelations error: %s", discCross.Error)
	}
	if len(discCross.Items) == 0 {
		t.Fatalf("expected cross-node correlations to be discovered")
	}
	for _, it := range discCross.Items {
		if !it.CrossNode || it.NodeID1 == it.NodeID2 {
			t.Fatalf("CrossNodeOnly=true returned same-node pair: %+v", it)
		}
		if it.PollingID1 == constID || it.PollingID2 == constID {
			t.Fatalf("constant polling should not produce high correlation: %+v", it)
		}
	}

	// 4. Auto discovery: 1-to-N with BasePollingID = routerID, CrossNodeOnly = false (should include routerAuxID)
	discBase := DiscoverPollingCorrelations(context.Background(), CorrelationDiscoverReq{
		BasePollingID: routerID,
		Algorithm:     "spearman",
		BucketSec:     60,
		StartTime:     st,
		EndTime:       et,
		MinAbsScore:   0.8,
		MinSamples:    10,
		CrossNodeOnly: false,
	})
	foundCloud := false
	foundRouterAux := false
	for _, it := range discBase.Items {
		if it.PollingID2 == cloudID || it.PollingID1 == cloudID {
			foundCloud = true
		}
		if it.PollingID2 == routerAuxID || it.PollingID1 == routerAuxID {
			foundRouterAux = true
		}
	}
	if !foundCloud || !foundRouterAux {
		t.Fatalf("expected both cloud (%v) and routerAux (%v) in 1-to-N discovery, got %+v", foundCloud, foundRouterAux, discBase.Items)
	}
}

func TestMCPCorrelation(t *testing.T) {
	setupMCPTestDB(t)
	cloudID, routerID, _, _, _, _ := setupCorrelationTestEntities(t)
	cs := connectMCPTestSession(t)

	// 1. analyze_polling_correlation tool
	pairSummary := callMCPToolJSON[mcpCorrelationPairSummary](t, cs, "analyze_polling_correlation", map[string]any{
		"polling_id_1": cloudID,
		"polling_id_2": routerID,
		"bucket_sec":   60,
		"period_hours": 2,
		"algorithm":    "all",
	})
	if pairSummary.SampleCount != 39 || math.Abs(pairSummary.PrimaryScore-1.0) > 1e-6 {
		t.Fatalf("unexpected MCP analyze_polling_correlation result: %+v", pairSummary)
	}
	if len(pairSummary.Scores) != 5 {
		t.Fatalf("expected 5 scores from MCP analyze_polling_correlation, got %d", len(pairSummary.Scores))
	}

	// 2. discover_polling_correlations tool
	discRes := callMCPToolJSON[CorrelationDiscoverRes](t, cs, "discover_polling_correlations", map[string]any{
		"algorithm":       "pearson",
		"bucket_sec":      60,
		"period_hours":    2,
		"min_abs_score":   0.8,
		"min_samples":     10,
		"cross_node_only": true,
		"limit":           10,
	})
	if len(discRes.Items) == 0 {
		t.Fatalf("expected MCP discover_polling_correlations to find items")
	}

	// 3. Error case for missing/unknown polling
	if _, err := callMCPToolText(t, cs, "analyze_polling_correlation", map[string]any{
		"polling_id_1": "missing",
		"polling_id_2": routerID,
	}); err == nil {
		t.Fatalf("expected error for missing polling ID")
	}
}

func TestDiscoverProgressAndCancel(t *testing.T) {
	setupMCPTestDB(t)
	_, _, _, _, st, et := setupCorrelationTestEntities(t)

	// 1. Progress phases + MinAbsScore=0 ranking
	phases := map[string]bool{}
	res := DiscoverPollingCorrelationsWithProgress(context.Background(), CorrelationDiscoverReq{
		Algorithm:     "pearson",
		BucketSec:     60,
		StartTime:     st,
		EndTime:       et,
		MinAbsScore:   0,
		MinSamples:    10,
		CrossNodeOnly: true,
		Limit:         10,
	}, func(p CorrelationProgress) {
		phases[p.Phase] = true
	})
	if res.Error != "" {
		t.Fatalf("unexpected error: %s", res.Error)
	}
	if !phases["loading"] || !phases["calculating"] || !phases["done"] {
		t.Fatalf("expected loading/calculating/done phases, got %+v", phases)
	}
	if len(res.Items) == 0 {
		t.Fatalf("expected ranked items with MinAbsScore=0")
	}

	// 2. Cancellation
	cctx, cancel := context.WithCancel(context.Background())
	cancel()
	cancelPhases := map[string]bool{}
	resCancel := DiscoverPollingCorrelationsWithProgress(cctx, CorrelationDiscoverReq{
		Algorithm: "pearson",
		BucketSec: 60,
		StartTime: st,
		EndTime:   et,
	}, func(p CorrelationProgress) {
		cancelPhases[p.Phase] = true
	})
	if resCancel.Error == "" || !cancelPhases["canceled"] {
		t.Fatalf("expected canceled error and phase, got err=%q phases=%+v", resCancel.Error, cancelPhases)
	}
}

func TestGetPollingLogsBatch(t *testing.T) {
	_, cancel, wg := setupMCPTestDB(t)
	n := &datastore.NodeEnt{Name: "BatchNode", IP: "10.0.9.1"}
	if err := datastore.AddNode(n); err != nil {
		t.Fatalf("AddNode: %v", err)
	}
	p1 := &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "p1",
		Type:    "ping",
		LogMode: datastore.LogModeAlways,
		Result:  map[string]any{"rtt": 10.0},
	}
	p2 := &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "p2",
		Type:    "ping",
		LogMode: datastore.LogModeAlways,
		Result:  map[string]any{"rtt": 20.0},
	}
	_ = datastore.AddPolling(p1)
	_ = datastore.AddPolling(p2)
	_ = datastore.AddPollingLog(p1)
	_ = datastore.AddPollingLog(p2)
	time.Sleep(10 * time.Millisecond)

	// Flush eventLogger synchronously before querying bbolt
	cancel()
	wg.Wait()

	now := time.Now().Unix()
	batch := datastore.GetPollingLogsBatch(context.Background(), []string{p1.ID, p2.ID}, now-60, now+60, nil)
	if len(batch[p1.ID]) != 1 || len(batch[p2.ID]) != 1 {
		t.Fatalf("expected 1 log per polling in batch, got p1=%d p2=%d", len(batch[p1.ID]), len(batch[p2.ID]))
	}
}
