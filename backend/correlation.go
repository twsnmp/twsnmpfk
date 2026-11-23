package backend

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"gonum.org/v1/gonum/stat/distuv"
)

var getPollingLogsFunc func(string) []*datastore.PollingLogEnt

func fetchPollingLogsBatch(
	ctx context.Context,
	pollingIDs []string,
	stSec, etSec int64,
	progressCb func(done, total int),
) map[string][]*datastore.PollingLogEnt {
	if getPollingLogsFunc != nil {
		out := make(map[string][]*datastore.PollingLogEnt, len(pollingIDs))
		if progressCb != nil {
			progressCb(0, len(pollingIDs))
		}
		for i, id := range pollingIDs {
			if ctx != nil && ctx.Err() != nil {
				break
			}
			out[id] = getPollingLogsFunc(id)
			if progressCb != nil {
				progressCb(i+1, len(pollingIDs))
			}
		}
		return out
	}
	return datastore.GetPollingLogsBatch(ctx, pollingIDs, stSec, etSec, progressCb)
}

// CorrelationProgress reports progress during polling correlation discovery.
type CorrelationProgress struct {
	Phase   string `json:"Phase"`
	Current int    `json:"Current"`
	Total   int    `json:"Total"`
	Message string `json:"Message"`
}

const (
	corrAlgoPearson   = "pearson"
	corrAlgoSpearman  = "spearman"
	corrAlgoKendall   = "kendall"
	corrAlgoCrossCorr = "crosscorr"
	corrAlgoDCor      = "dcor"
)

// CorrelationCandidate describes a polling and its available numeric vector keys.
type CorrelationCandidate struct {
	PollingID   string   `json:"PollingID"`
	NodeID      string   `json:"NodeID"`
	NodeName    string   `json:"NodeName"`
	PollingName string   `json:"PollingName"`
	Type        string   `json:"Type"`
	Mode        string   `json:"Mode"`
	PollInt     int      `json:"PollInt"`
	LogMode     int      `json:"LogMode"`
	VectorCols  string   `json:"VectorCols"`
	Keys        []string `json:"Keys"`
}

// CorrelationPairReq specifies parameters for analyzing correlation between two pollings.
type CorrelationPairReq struct {
	PollingID1   string   `json:"PollingID1"`
	PollingID2   string   `json:"PollingID2"`
	Vars1        []string `json:"Vars1"`
	Vars2        []string `json:"Vars2"`
	IncludeState bool     `json:"IncludeState"`
	Algorithm    string   `json:"Algorithm"`
	BucketSec    int64    `json:"BucketSec"`
	StartTime    int64    `json:"StartTime"`
	EndTime      int64    `json:"EndTime"`
	MaxLag       int      `json:"MaxLag"`
}

// CorrelationAlgoScore holds the correlation result for one algorithm.
type CorrelationAlgoScore struct {
	Algorithm   string  `json:"Algorithm"`
	Score       float64 `json:"Score"`
	PValue      float64 `json:"PValue"`
	Lag         int     `json:"Lag"`
	LagSec      int64   `json:"LagSec"`
	SampleCount int     `json:"SampleCount"`
	Var1        string  `json:"Var1"`
	Var2        string  `json:"Var2"`
}

// CorrelationMatrixCell holds a correlation score for one variable pair.
type CorrelationMatrixCell struct {
	Var1   string  `json:"Var1"`
	Var2   string  `json:"Var2"`
	Score  float64 `json:"Score"`
	PValue float64 `json:"PValue"`
}

// CorrelationLagPoint represents cross-correlation score at a specific bucket lag.
type CorrelationLagPoint struct {
	Lag    int     `json:"Lag"`
	LagSec int64   `json:"LagSec"`
	Score  float64 `json:"Score"`
}

// CorrelationPairRes returns the detailed correlation analysis result for two pollings.
type CorrelationPairRes struct {
	PollingID1    string                  `json:"PollingID1"`
	NodeID1       string                  `json:"NodeID1"`
	NodeName1     string                  `json:"NodeName1"`
	PollingName1  string                  `json:"PollingName1"`
	SelectedVar1  string                  `json:"SelectedVar1"`
	PollingID2    string                  `json:"PollingID2"`
	NodeID2       string                  `json:"NodeID2"`
	NodeName2     string                  `json:"NodeName2"`
	PollingName2  string                  `json:"PollingName2"`
	SelectedVar2  string                  `json:"SelectedVar2"`
	Vars1         []string                `json:"Vars1"`
	Vars2         []string                `json:"Vars2"`
	BucketSec     int64                   `json:"BucketSec"`
	SampleCount   int                     `json:"SampleCount"`
	Scores        []CorrelationAlgoScore  `json:"Scores"`
	PrimaryScore  float64                 `json:"PrimaryScore"`
	PrimaryPValue float64                 `json:"PrimaryPValue"`
	BestLag       int                     `json:"BestLag"`
	BestLagSec    int64                   `json:"BestLagSec"`
	Slope         float64                 `json:"Slope"`
	Intercept     float64                 `json:"Intercept"`
	Timestamps    []int64                 `json:"Timestamps"`
	Series1       []float64               `json:"Series1"`
	Series2       []float64               `json:"Series2"`
	NormSeries1   []float64               `json:"NormSeries1"`
	NormSeries2   []float64               `json:"NormSeries2"`
	LagProfile    []CorrelationLagPoint   `json:"LagProfile"`
	VarMatrix     []CorrelationMatrixCell `json:"VarMatrix"`
	Error         string                  `json:"Error"`
}

// CorrelationDiscoverReq specifies parameters for automatic correlation discovery.
type CorrelationDiscoverReq struct {
	BasePollingID string   `json:"BasePollingID"`
	NodeIDs       []string `json:"NodeIDs"`
	Algorithm     string   `json:"Algorithm"`
	BucketSec     int64    `json:"BucketSec"`
	StartTime     int64    `json:"StartTime"`
	EndTime       int64    `json:"EndTime"`
	MinAbsScore   float64  `json:"MinAbsScore"`
	MinSamples    int      `json:"MinSamples"`
	CrossNodeOnly bool     `json:"CrossNodeOnly"`
	IncludeState  bool     `json:"IncludeState"`
	MaxLag        int      `json:"MaxLag"`
	Limit         int      `json:"Limit"`
}

// CorrelationDiscoverItem describes a discovered correlated polling pair.
type CorrelationDiscoverItem struct {
	PollingID1   string  `json:"PollingID1"`
	NodeID1      string  `json:"NodeID1"`
	NodeName1    string  `json:"NodeName1"`
	PollingName1 string  `json:"PollingName1"`
	Var1         string  `json:"Var1"`
	PollingID2   string  `json:"PollingID2"`
	NodeID2      string  `json:"NodeID2"`
	NodeName2    string  `json:"NodeName2"`
	PollingName2 string  `json:"PollingName2"`
	Var2         string  `json:"Var2"`
	Algorithm    string  `json:"Algorithm"`
	Score        float64 `json:"Score"`
	AbsScore     float64 `json:"AbsScore"`
	PValue       float64 `json:"PValue"`
	Lag          int     `json:"Lag"`
	LagSec       int64   `json:"LagSec"`
	SampleCount  int     `json:"SampleCount"`
	CrossNode    bool    `json:"CrossNode"`
}

// CorrelationDiscoverRes holds the auto-discovery summary and discovered pairs.
type CorrelationDiscoverRes struct {
	Items          []CorrelationDiscoverItem `json:"Items"`
	TotalPollings  int                       `json:"TotalPollings"`
	EvaluatedPairs int                       `json:"EvaluatedPairs"`
	DurationMs     int64                     `json:"DurationMs"`
	BucketSec      int64                     `json:"BucketSec"`
	Error          string                    `json:"Error"`
}

// pollingBucketSeries stores bucket-averaged numeric variables for a polling.
type pollingBucketSeries struct {
	PollingID   string
	NodeID      string
	NodeName    string
	PollingName string
	Vars        []string
	Timestamps  []int64              // sorted ascending Unix seconds
	Data        map[string][]float64 // aligned to Timestamps
}

func sanitizeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0.0
	}
	return v
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, false
		}
		return val, true
	case float32:
		f := float64(val)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uint32:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// getCorrelationKeys extracts numeric keys for correlation using the same strategy as AI anomaly detection.
func getCorrelationKeys(p *datastore.PollingEnt, logs []*datastore.PollingLogEnt, requested []string, includeState bool) []string {
	if p == nil {
		return nil
	}
	// Discover all available numeric keys from polling Result (matching getAIDataKeys)
	availSet := make(map[string]bool)
	if p.Type == "syslog" && p.Mode == "pri" {
		for i := 0; i < 256; i++ {
			availSet[fmt.Sprintf("pri_%d", i)] = true
		}
	} else {
		for k, v := range p.Result {
			if k == "lastTime" {
				continue
			}
			if _, ok := toFloat64(v); ok {
				availSet[k] = true
			}
		}
		// Fallback to scanning logs if current Result has no numeric keys
		if len(availSet) == 0 && len(logs) > 0 {
			step := 1
			if len(logs) > 20 {
				step = len(logs) / 20
			}
			for i := 0; i < len(logs); i += step {
				for k, v := range logs[i].Result {
					if k == "lastTime" {
						continue
					}
					if _, ok := toFloat64(v); ok {
						availSet[k] = true
					}
				}
			}
		}
	}

	// Explicit requested variables take highest priority
	if len(requested) > 0 {
		out := []string{}
		seen := map[string]bool{}
		for _, r := range requested {
			r = strings.TrimSpace(r)
			if r == "" || r == "hour" || r == "weekday" || seen[r] {
				continue
			}
			if r == "state" || availSet[r] || len(availSet) == 0 {
				seen[r] = true
				out = append(out, r)
			}
		}
		if len(out) > 0 {
			return out
		}
	}

	// Use VectorCols configured on the polling (excluding time features hour/weekday)
	if strings.TrimSpace(p.VectorCols) != "" {
		out := []string{}
		seen := map[string]bool{}
		for _, c := range strings.Split(p.VectorCols, ",") {
			c = strings.TrimSpace(c)
			if c == "" || c == "hour" || c == "weekday" || seen[c] {
				continue
			}
			seen[c] = true
			out = append(out, c)
		}
		if len(out) > 0 {
			if includeState && !seen["state"] {
				out = append(out, "state")
			}
			return out
		}
	}

	// Default: sorted numeric keys
	keys := make([]string, 0, len(availSet))
	for k := range availSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(keys) == 0 || includeState {
		keys = append(keys, "state")
	}
	return keys
}

// GetCorrelationCandidates returns all pollings with logging enabled and their numeric variable keys.
func GetCorrelationCandidates() []CorrelationCandidate {
	ret := []CorrelationCandidate{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode == datastore.LogModeNone {
			return true
		}
		nodeName := ""
		if n := datastore.GetNode(p.NodeID); n != nil {
			nodeName = n.Name
		}
		keys := getCorrelationKeys(p, nil, nil, true)
		ret = append(ret, CorrelationCandidate{
			PollingID:   p.ID,
			NodeID:      p.NodeID,
			NodeName:    nodeName,
			PollingName: p.Name,
			Type:        p.Type,
			Mode:        p.Mode,
			PollInt:     p.PollInt,
			LogMode:     p.LogMode,
			VectorCols:  p.VectorCols,
			Keys:        keys,
		})
		return true
	})
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].NodeName == ret[j].NodeName {
			if ret[i].PollingName == ret[j].PollingName {
				return ret[i].PollingID < ret[j].PollingID
			}
			return ret[i].PollingName < ret[j].PollingName
		}
		return ret[i].NodeName < ret[j].NodeName
	})
	return ret
}

func resolveAutoBucketSec(startSec, endSec int64, minPollInt int) int64 {
	var span int64
	if startSec > 0 && endSec > startSec {
		span = endSec - startSec
	} else if startSec > 0 {
		span = time.Now().Unix() - startSec
	}
	var bucket int64 = 300
	if span > 0 {
		switch {
		case span <= 24*3600:
			bucket = 60
		case span <= 7*24*3600:
			bucket = 300
		case span <= 30*24*3600:
			bucket = 900
		default:
			bucket = 3600
		}
	}
	if minPollInt > 0 {
		var snapped int64
		switch {
		case minPollInt <= 60:
			snapped = 60
		case minPollInt <= 300:
			snapped = 300
		case minPollInt <= 900:
			snapped = 900
		default:
			snapped = 3600
		}
		if snapped > bucket {
			bucket = snapped
		}
	}
	if bucket < 1 {
		bucket = 60
	}
	return bucket
}

func buildBucketSeriesFromLogs(
	p *datastore.PollingEnt,
	nodeName string,
	logs []*datastore.PollingLogEnt,
	vars []string,
	bucketSec int64,
	startSec, endSec int64,
) *pollingBucketSeries {
	if p == nil || len(logs) == 0 || len(vars) == 0 || bucketSec <= 0 {
		return nil
	}

	type bucketAcc struct {
		sums  map[string]float64
		count float64
	}
	buckets := make(map[int64]*bucketAcc)
	for _, l := range logs {
		if l == nil {
			continue
		}
		tsSec := l.Time / 1e9
		if startSec > 0 && tsSec < startSec {
			continue
		}
		if endSec > 0 && tsSec > endSec {
			continue
		}
		b := (tsSec / bucketSec) * bucketSec
		acc, ok := buckets[b]
		if !ok {
			acc = &bucketAcc{
				sums: make(map[string]float64, len(vars)),
			}
			buckets[b] = acc
		}
		acc.count += 1.0
		for _, k := range vars {
			if k == "state" {
				acc.sums[k] += getStateNum(l.State)
				continue
			}
			if l.Result != nil {
				if v, ok := l.Result[k]; ok {
					if fv, ok := toFloat64(v); ok {
						acc.sums[k] += fv
					}
				}
			}
		}
	}
	if len(buckets) == 0 {
		return nil
	}
	timestamps := make([]int64, 0, len(buckets))
	for ts := range buckets {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	data := make(map[string][]float64, len(vars))
	for _, k := range vars {
		data[k] = make([]float64, len(timestamps))
	}
	for idx, ts := range timestamps {
		acc := buckets[ts]
		for _, k := range vars {
			data[k][idx] = acc.sums[k] / acc.count
		}
	}

	// Filter out non-requested zero-variance keys? Keep vars intact for pair matrix,
	// individual correlation routines check variance cleanly.
	return &pollingBucketSeries{
		PollingID:   p.ID,
		NodeID:      p.NodeID,
		NodeName:    nodeName,
		PollingName: p.Name,
		Vars:        vars,
		Timestamps:  timestamps,
		Data:        data,
	}
}

func alignTwoBucketSeries(s1, s2 *pollingBucketSeries) ([]int64, map[string][]float64, map[string][]float64) {
	if s1 == nil || s2 == nil || len(s1.Timestamps) == 0 || len(s2.Timestamps) == 0 {
		return nil, nil, nil
	}
	tsOut := make([]int64, 0, min(len(s1.Timestamps), len(s2.Timestamps)))
	d1 := make(map[string][]float64, len(s1.Vars))
	d2 := make(map[string][]float64, len(s2.Vars))
	for _, k := range s1.Vars {
		d1[k] = []float64{}
	}
	for _, k := range s2.Vars {
		d2[k] = []float64{}
	}

	i, j := 0, 0
	for i < len(s1.Timestamps) && j < len(s2.Timestamps) {
		t1, t2 := s1.Timestamps[i], s2.Timestamps[j]
		if t1 == t2 {
			tsOut = append(tsOut, t1)
			for _, k := range s1.Vars {
				d1[k] = append(d1[k], s1.Data[k][i])
			}
			for _, k := range s2.Vars {
				d2[k] = append(d2[k], s2.Data[k][j])
			}
			i++
			j++
		} else if t1 < t2 {
			i++
		} else {
			j++
		}
	}
	return tsOut, d1, d2
}

func minMaxNormalize(xs []float64) []float64 {
	n := len(xs)
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	minV, maxV := xs[0], xs[0]
	for _, v := range xs[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	diff := maxV - minV
	if diff < 1e-12 {
		for i := range out {
			out[i] = 0.5
		}
		return out
	}
	for i, v := range xs {
		out[i] = sanitizeFloat((v - minV) / diff)
	}
	return out
}

func calcLinearRegression(x, y []float64) (slope, intercept float64) {
	n := len(x)
	if n < 2 || len(y) != n {
		return 0, 0
	}
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)
	var num, den float64
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		num += dx * dy
		den += dx * dx
	}
	if den < 1e-18 {
		return 0, sanitizeFloat(meanY)
	}
	slope = num / den
	intercept = meanY - slope*meanX
	return sanitizeFloat(slope), sanitizeFloat(intercept)
}

func pearsonPValue(r float64, n int) float64 {
	if n <= 2 {
		return 1.0
	}
	absR := math.Abs(r)
	if absR >= 1.0-1e-12 {
		return 0.0
	}
	tVal := absR * math.Sqrt(float64(n-2)/(1.0-absR*absR))
	tDist := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: float64(n - 2)}
	p := 2.0 * (1.0 - tDist.CDF(tVal))
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return sanitizeFloat(p)
}

// calcPearson computes Pearson correlation coefficient and two-tailed p-value.
func calcPearson(x, y []float64) (r float64, pValue float64) {
	n := len(x)
	if n < 3 || len(y) != n {
		return 0, 1.0
	}
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)
	var cov, varX, varY float64
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		cov += dx * dy
		varX += dx * dx
		varY += dy * dy
	}
	if varX < 1e-18 || varY < 1e-18 {
		return 0, 1.0
	}
	r = cov / math.Sqrt(varX*varY)
	if r > 1.0 {
		r = 1.0
	} else if r < -1.0 {
		r = -1.0
	}
	return sanitizeFloat(r), pearsonPValue(r, n)
}

func rankWithTies(xs []float64) []float64 {
	n := len(xs)
	ranks := make([]float64, n)
	if n == 0 {
		return ranks
	}
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool {
		return xs[idx[a]] < xs[idx[b]]
	})
	i := 0
	for i < n {
		j := i + 1
		for j < n && math.Abs(xs[idx[j]]-xs[idx[i]]) <= 1e-12 {
			j++
		}
		avgRank := 0.5*float64(i+j-1) + 1.0
		for k := i; k < j; k++ {
			ranks[idx[k]] = avgRank
		}
		i = j
	}
	return ranks
}

// calcSpearman computes Spearman rank correlation coefficient and two-tailed p-value.
func calcSpearman(x, y []float64) (rho float64, pValue float64) {
	if len(x) < 3 || len(x) != len(y) {
		return 0, 1.0
	}
	rx := rankWithTies(x)
	ry := rankWithTies(y)
	return calcPearson(rx, ry)
}

func uniformSubsampleIndices(n, maxN int) []int {
	if n <= maxN || maxN <= 0 {
		idx := make([]int, n)
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	idx := make([]int, maxN)
	for i := 0; i < maxN; i++ {
		idx[i] = i * (n - 1) / (maxN - 1)
	}
	return idx
}

func subsamplePair(x, y []float64, maxN int) ([]float64, []float64) {
	n := len(x)
	if n <= maxN {
		return x, y
	}
	idx := uniformSubsampleIndices(n, maxN)
	sx := make([]float64, len(idx))
	sy := make([]float64, len(idx))
	for i, k := range idx {
		sx[i] = x[k]
		sy[i] = y[k]
	}
	return sx, sy
}

// calcKendallTauB computes Kendall's tau-b with tie correction and normal-approximated p-value.
func calcKendallTauB(x, y []float64) (tau float64, pValue float64) {
	if len(x) < 3 || len(x) != len(y) {
		return 0, 1.0
	}
	sx, sy := subsamplePair(x, y, 1000)
	n := len(sx)

	var concordant, discordant, tiesX, tiesY int64
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			dx := sx[i] - sx[j]
			dy := sy[i] - sy[j]
			tx := math.Abs(dx) <= 1e-12
			ty := math.Abs(dy) <= 1e-12
			if tx && ty {
				continue
			}
			if tx {
				tiesX++
				continue
			}
			if ty {
				tiesY++
				continue
			}
			if (dx > 0 && dy > 0) || (dx < 0 && dy < 0) {
				concordant++
			} else {
				discordant++
			}
		}
	}

	denX := float64(concordant + discordant + tiesX)
	denY := float64(concordant + discordant + tiesY)
	if denX <= 0 || denY <= 0 {
		return 0, 1.0
	}
	tau = float64(concordant-discordant) / math.Sqrt(denX*denY)
	if tau > 1.0 {
		tau = 1.0
	} else if tau < -1.0 {
		tau = -1.0
	}

	// Normal approximation for Kendall's tau p-value
	if n >= 4 {
		// Standard formula: S = C - D, Var(S) = n(n-1)(2n+5)/18
		v0 := float64(n*(n-1)*(2*n+5)) / 18.0
		if v0 <= 0 {
			return sanitizeFloat(tau), 1.0
		}
		z := float64(concordant-discordant) / math.Sqrt(v0)
		pValue = math.Erfc(math.Abs(z) / math.Sqrt2)
		if pValue > 1 {
			pValue = 1
		}
	} else {
		pValue = 1.0
	}
	return sanitizeFloat(tau), sanitizeFloat(pValue)
}

// calcCrossCorrelation evaluates Pearson correlation across lags in [-maxLag, +maxLag].
// Positive lag k means x leads y by k buckets: correlation(x[t], y[t+k]).
func calcCrossCorrelation(x, y []float64, maxLag int, bucketSec int64, minOverlap int) (bestScore float64, bestP float64, bestLag int, profile []CorrelationLagPoint) {
	n := len(x)
	if n < 3 || len(y) != n {
		return 0, 1.0, 0, []CorrelationLagPoint{}
	}
	if maxLag < 0 {
		maxLag = 0
	}
	if maxLag > 60 {
		maxLag = 60
	}
	if minOverlap < 5 {
		minOverlap = 5
	}
	if maxLag > n-minOverlap {
		maxLag = max(0, n-minOverlap)
	}

	profile = make([]CorrelationLagPoint, 0, 2*maxLag+1)
	bestAbs := -1.0
	bestP = 1.0

	for lag := -maxLag; lag <= maxLag; lag++ {
		var xs, ys []float64
		if lag >= 0 {
			xs = x[:n-lag]
			ys = y[lag:]
		} else {
			k := -lag
			xs = x[k:]
			ys = y[:n-k]
		}
		if len(xs) < minOverlap {
			continue
		}
		r, p := calcPearson(xs, ys)
		profile = append(profile, CorrelationLagPoint{
			Lag:    lag,
			LagSec: int64(lag) * bucketSec,
			Score:  r,
		})
		absR := math.Abs(r)
		// Prefer larger |r|; break ties with smaller |lag|
		if absR > bestAbs+1e-12 || (math.Abs(absR-bestAbs) <= 1e-12 && absInt(lag) < absInt(bestLag)) {
			bestAbs = absR
			bestScore = r
			bestP = p
			bestLag = lag
		}
	}
	if bestAbs < 0 {
		return 0, 1.0, 0, profile
	}
	return sanitizeFloat(bestScore), sanitizeFloat(bestP), bestLag, profile
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func standardizeColumns(cols [][]float64, indices []int) [][]float64 {
	n := len(indices)
	if n == 0 || len(cols) == 0 {
		return nil
	}
	validCols := make([][]float64, 0, len(cols))
	for _, col := range cols {
		if len(col) == 0 {
			continue
		}
		var sum float64
		for _, idx := range indices {
			sum += col[idx]
		}
		mean := sum / float64(n)
		var sumSq float64
		for _, idx := range indices {
			d := col[idx] - mean
			sumSq += d * d
		}
		sd := math.Sqrt(sumSq / float64(n))
		if sd < 1e-12 {
			continue
		}
		norm := make([]float64, n)
		for i, idx := range indices {
			norm[i] = (col[idx] - mean) / sd
		}
		validCols = append(validCols, norm)
	}
	if len(validCols) == 0 {
		return nil
	}
	rows := make([][]float64, n)
	for i := 0; i < n; i++ {
		row := make([]float64, len(validCols))
		for c := 0; c < len(validCols); c++ {
			row[c] = validCols[c][i]
		}
		rows[i] = row
	}
	return rows
}

func doubleCenteredDistMatrix(rows [][]float64) ([]float64, float64) {
	n := len(rows)
	if n == 0 {
		return nil, 0
	}
	dim := len(rows[0])
	dMat := make([]float64, n*n)
	rowMeans := make([]float64, n)
	var grandSum float64

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			var sq float64
			for c := 0; c < dim; c++ {
				diff := rows[i][c] - rows[j][c]
				sq += diff * diff
			}
			dist := math.Sqrt(sq)
			dMat[i*n+j] = dist
			dMat[j*n+i] = dist
			rowMeans[i] += dist
			rowMeans[j] += dist
			grandSum += 2.0 * dist
		}
	}
	for i := 0; i < n; i++ {
		rowMeans[i] /= float64(n)
	}
	grandMean := grandSum / float64(n*n)

	var dVar float64
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			centered := dMat[i*n+j] - rowMeans[i] - rowMeans[j] + grandMean
			dMat[i*n+j] = centered
			dVar += centered * centered
		}
	}
	dVar /= float64(n * n)
	return dMat, dVar
}

// calcDistanceCorrelation computes multivariate Distance Correlation (0..1) and optional permutation p-value.
func calcDistanceCorrelation(colsX, colsY [][]float64, maxSamples int, withPValue bool) (dCor float64, pValue float64) {
	if len(colsX) == 0 || len(colsY) == 0 {
		return 0, 1.0
	}
	nRaw := len(colsX[0])
	if nRaw < 3 || len(colsY[0]) != nRaw {
		return 0, 1.0
	}
	if maxSamples <= 0 {
		maxSamples = 500
	}
	indices := uniformSubsampleIndices(nRaw, maxSamples)
	rowsX := standardizeColumns(colsX, indices)
	rowsY := standardizeColumns(colsY, indices)
	if len(rowsX) < 3 || len(rowsY) < 3 {
		return 0, 1.0
	}
	n := len(rowsX)
	matA, varX := doubleCenteredDistMatrix(rowsX)
	matB, varY := doubleCenteredDistMatrix(rowsY)
	if varX <= 1e-18 || varY <= 1e-18 {
		return 0, 1.0
	}
	var covSq float64
	for i := 0; i < n*n; i++ {
		covSq += matA[i] * matB[i]
	}
	covSq /= float64(n * n)
	if covSq <= 0 {
		return 0, 1.0
	}
	dCorSq := covSq / math.Sqrt(varX*varY)
	if dCorSq > 1.0 {
		dCorSq = 1.0
	}
	dCor = math.Sqrt(dCorSq)

	pValue = 1.0
	if withPValue {
		// Deterministic permutation test reusing double-centered matrices A and B
		const perms = 99
		rng := rand.New(rand.NewSource(42))
		perm := make([]int, n)
		for i := range perm {
			perm[i] = i
		}
		exceed := 0
		for p := 0; p < perms; p++ {
			rng.Shuffle(n, func(i, j int) {
				perm[i], perm[j] = perm[j], perm[i]
			})
			var pCovSq float64
			for i := 0; i < n; i++ {
				pi := perm[i]
				rowA := i * n
				rowB := pi * n
				for j := 0; j < n; j++ {
					pCovSq += matA[rowA+j] * matB[rowB+perm[j]]
				}
			}
			pCovSq /= float64(n * n)
			if pCovSq >= covSq-1e-15 {
				exceed++
			}
		}
		pValue = float64(exceed+1) / float64(perms+1)
	}
	return sanitizeFloat(dCor), sanitizeFloat(pValue)
}

func normalizeAlgorithm(algo string) string {
	switch strings.ToLower(strings.TrimSpace(algo)) {
	case "", "pearson", "all":
		return corrAlgoPearson
	case "spearman":
		return corrAlgoSpearman
	case "kendall":
		return corrAlgoKendall
	case "crosscorr", "cross", "lag":
		return corrAlgoCrossCorr
	case "dcor", "distance":
		return corrAlgoDCor
	default:
		return corrAlgoPearson
	}
}

func evalUnivariatePair(algo string, x, y []float64, maxLag int, bucketSec int64) (score, pValue float64, lag int, profile []CorrelationLagPoint) {
	switch algo {
	case corrAlgoSpearman:
		s, p := calcSpearman(x, y)
		return s, p, 0, nil
	case corrAlgoKendall:
		s, p := calcKendallTauB(x, y)
		return s, p, 0, nil
	case corrAlgoCrossCorr:
		return calcCrossCorrelation(x, y, maxLag, bucketSec, 10)
	default:
		s, p := calcPearson(x, y)
		return s, p, 0, nil
	}
}

// AnalyzePollingCorrelation performs comprehensive correlation analysis between two pollings.
func AnalyzePollingCorrelation(req CorrelationPairReq) CorrelationPairRes {
	res := CorrelationPairRes{
		PollingID1: req.PollingID1,
		PollingID2: req.PollingID2,
		Vars1:      []string{},
		Vars2:      []string{},
		Scores:     []CorrelationAlgoScore{},
		Timestamps: []int64{},
		Series1:    []float64{},
		Series2:    []float64{},
		NormSeries1: []float64{},
		NormSeries2: []float64{},
		LagProfile: []CorrelationLagPoint{},
		VarMatrix:  []CorrelationMatrixCell{},
	}

	p1 := datastore.GetPolling(req.PollingID1)
	p2 := datastore.GetPolling(req.PollingID2)
	if p1 == nil || p2 == nil {
		res.Error = "polling not found"
		return res
	}
	res.NodeID1 = p1.NodeID
	res.PollingName1 = p1.Name
	if n1 := datastore.GetNode(p1.NodeID); n1 != nil {
		res.NodeName1 = n1.Name
	}
	res.NodeID2 = p2.NodeID
	res.PollingName2 = p2.Name
	if n2 := datastore.GetNode(p2.NodeID); n2 != nil {
		res.NodeName2 = n2.Name
	}

	ids := []string{p1.ID}
	if p2.ID != p1.ID {
		ids = append(ids, p2.ID)
	}
	logsMap := fetchPollingLogsBatch(context.Background(), ids, req.StartTime, req.EndTime, nil)
	logs1 := logsMap[p1.ID]
	logs2 := logsMap[p2.ID]
	if len(logs1) == 0 || len(logs2) == 0 {
		res.Error = "insufficient polling logs"
		return res
	}

	vars1 := getCorrelationKeys(p1, logs1, req.Vars1, req.IncludeState)
	vars2 := getCorrelationKeys(p2, logs2, req.Vars2, req.IncludeState)
	res.Vars1 = vars1
	res.Vars2 = vars2
	if len(vars1) == 0 || len(vars2) == 0 {
		res.Error = "no numeric variables found"
		return res
	}

	bucketSec := req.BucketSec
	if bucketSec <= 0 {
		minPollInt := p1.PollInt
		if p2.PollInt > minPollInt {
			minPollInt = p2.PollInt
		}
		st, et := req.StartTime, req.EndTime
		if st == 0 && et == 0 {
			s1Start, s1End := logs1[0].Time/1e9, logs1[len(logs1)-1].Time/1e9
			s2Start, s2End := logs2[0].Time/1e9, logs2[len(logs2)-1].Time/1e9
			st = max(s1Start, s2Start)
			et = min(s1End, s2End)
		}
		bucketSec = resolveAutoBucketSec(st, et, minPollInt)
	}
	res.BucketSec = bucketSec

	maxLag := req.MaxLag
	if maxLag <= 0 {
		maxLag = 12
	}
	if maxLag > 60 {
		maxLag = 60
	}

	bs1 := buildBucketSeriesFromLogs(p1, res.NodeName1, logs1, vars1, bucketSec, req.StartTime, req.EndTime)
	bs2 := buildBucketSeriesFromLogs(p2, res.NodeName2, logs2, vars2, bucketSec, req.StartTime, req.EndTime)
	ts, d1, d2 := alignTwoBucketSeries(bs1, bs2)
	res.SampleCount = len(ts)
	if len(ts) < 5 {
		res.Error = fmt.Sprintf("insufficient overlapping samples (%d < 5)", len(ts))
		return res
	}

	primaryAlgo := normalizeAlgorithm(req.Algorithm)
	matrixAlgo := primaryAlgo
	if matrixAlgo == corrAlgoDCor {
		matrixAlgo = corrAlgoPearson
	}

	// Evaluate all variable pairs for VarMatrix and select best pair
	bestVar1, bestVar2 := vars1[0], vars2[0]
	bestMatrixAbs := -1.0
	for _, v1 := range vars1 {
		for _, v2 := range vars2 {
			// Skip identical variable if analyzing the exact same polling
			if p1.ID == p2.ID && v1 == v2 && (len(vars1) > 1 || len(vars2) > 1) {
				continue
			}
			s, p, _, _ := evalUnivariatePair(matrixAlgo, d1[v1], d2[v2], maxLag, bucketSec)
			res.VarMatrix = append(res.VarMatrix, CorrelationMatrixCell{
				Var1:   v1,
				Var2:   v2,
				Score:  s,
				PValue: p,
			})
			if math.Abs(s) > bestMatrixAbs {
				bestMatrixAbs = math.Abs(s)
				bestVar1 = v1
				bestVar2 = v2
			}
		}
	}

	x := d1[bestVar1]
	y := d2[bestVar2]
	res.SelectedVar1 = bestVar1
	res.SelectedVar2 = bestVar2
	res.Timestamps = ts
	res.Series1 = x
	res.Series2 = y
	res.NormSeries1 = minMaxNormalize(x)
	res.NormSeries2 = minMaxNormalize(y)
	res.Slope, res.Intercept = calcLinearRegression(x, y)

	// Compute all 5 algorithms
	pScore, pP := calcPearson(x, y)
	sScore, sP := calcSpearman(x, y)
	kScore, kP := calcKendallTauB(x, y)
	cScore, cP, cLag, lagProf := calcCrossCorrelation(x, y, maxLag, bucketSec, 10)
	res.LagProfile = lagProf

	cols1 := make([][]float64, 0, len(vars1))
	for _, k := range vars1 {
		cols1 = append(cols1, d1[k])
	}
	cols2 := make([][]float64, 0, len(vars2))
	for _, k := range vars2 {
		cols2 = append(cols2, d2[k])
	}
	dScore, dP := calcDistanceCorrelation(cols1, cols2, 200, true)

	res.Scores = []CorrelationAlgoScore{
		{
			Algorithm:   corrAlgoPearson,
			Score:       pScore,
			PValue:      pP,
			SampleCount: len(ts),
			Var1:        bestVar1,
			Var2:        bestVar2,
		},
		{
			Algorithm:   corrAlgoSpearman,
			Score:       sScore,
			PValue:      sP,
			SampleCount: len(ts),
			Var1:        bestVar1,
			Var2:        bestVar2,
		},
		{
			Algorithm:   corrAlgoKendall,
			Score:       kScore,
			PValue:      kP,
			SampleCount: len(ts),
			Var1:        bestVar1,
			Var2:        bestVar2,
		},
		{
			Algorithm:   corrAlgoCrossCorr,
			Score:       cScore,
			PValue:      cP,
			Lag:         cLag,
			LagSec:      int64(cLag) * bucketSec,
			SampleCount: len(ts),
			Var1:        bestVar1,
			Var2:        bestVar2,
		},
		{
			Algorithm:   corrAlgoDCor,
			Score:       dScore,
			PValue:      dP,
			SampleCount: len(ts),
			Var1:        strings.Join(vars1, ","),
			Var2:        strings.Join(vars2, ","),
		},
	}

	for _, sc := range res.Scores {
		if sc.Algorithm == primaryAlgo {
			res.PrimaryScore = sc.Score
			res.PrimaryPValue = sc.PValue
			res.BestLag = sc.Lag
			res.BestLagSec = sc.LagSec
			break
		}
	}
	return res
}

func evaluatePairForDiscovery(bs1, bs2 *pollingBucketSeries, algo string, bucketSec int64, maxLag int, minSamples int) (CorrelationDiscoverItem, bool) {
	ts, d1, d2 := alignTwoBucketSeries(bs1, bs2)
	if len(ts) < minSamples {
		return CorrelationDiscoverItem{}, false
	}

	item := CorrelationDiscoverItem{
		PollingID1:   bs1.PollingID,
		NodeID1:      bs1.NodeID,
		NodeName1:    bs1.NodeName,
		PollingName1: bs1.PollingName,
		PollingID2:   bs2.PollingID,
		NodeID2:      bs2.NodeID,
		NodeName2:    bs2.NodeName,
		PollingName2: bs2.PollingName,
		Algorithm:    algo,
		SampleCount:  len(ts),
		CrossNode:    bs1.NodeID != bs2.NodeID,
	}

	if algo == corrAlgoDCor {
		cols1 := make([][]float64, 0, len(bs1.Vars))
		for _, k := range bs1.Vars {
			cols1 = append(cols1, d1[k])
		}
		cols2 := make([][]float64, 0, len(bs2.Vars))
		for _, k := range bs2.Vars {
			cols2 = append(cols2, d2[k])
		}
		s, p := calcDistanceCorrelation(cols1, cols2, 200, false)
		item.Var1 = strings.Join(bs1.Vars, ",")
		item.Var2 = strings.Join(bs2.Vars, ",")
		item.Score = s
		item.AbsScore = math.Abs(s)
		item.PValue = p
		return item, true
	}

	bestAbs := -1.0
	bestScore := 0.0
	bestP := 1.0
	bestLag := 0
	bestV1, bestV2 := "", ""
	for _, v1 := range bs1.Vars {
		for _, v2 := range bs2.Vars {
			s, p, lag, _ := evalUnivariatePair(algo, d1[v1], d2[v2], maxLag, bucketSec)
			absS := math.Abs(s)
			if absS > bestAbs {
				bestAbs = absS
				bestScore = s
				bestP = p
				bestLag = lag
				bestV1 = v1
				bestV2 = v2
			}
		}
	}
	if bestAbs < 0 || bestV1 == "" {
		return CorrelationDiscoverItem{}, false
	}
	item.Var1 = bestV1
	item.Var2 = bestV2
	item.Score = bestScore
	item.AbsScore = bestAbs
	item.PValue = bestP
	item.Lag = bestLag
	item.LagSec = int64(bestLag) * bucketSec
	return item, true
}

// DiscoverPollingCorrelations scans pollings in parallel and ranks correlated pairs.
func DiscoverPollingCorrelations(ctx context.Context, req CorrelationDiscoverReq) CorrelationDiscoverRes {
	return DiscoverPollingCorrelationsWithProgress(ctx, req, nil)
}

// DiscoverPollingCorrelationsWithProgress scans pollings in a single log pass and reports progress.
func DiscoverPollingCorrelationsWithProgress(
	ctx context.Context,
	req CorrelationDiscoverReq,
	progressCb func(CorrelationProgress),
) CorrelationDiscoverRes {
	startTime := time.Now()
	if ctx == nil {
		ctx = context.Background()
	}
	emit := func(phase string, current, total int, msg string) {
		if progressCb != nil {
			progressCb(CorrelationProgress{
				Phase:   phase,
				Current: current,
				Total:   total,
				Message: msg,
			})
		}
	}

	algo := normalizeAlgorithm(req.Algorithm)
	minAbs := req.MinAbsScore
	if minAbs < 0 {
		minAbs = 0
	}
	minSamples := req.MinSamples
	if minSamples < 5 {
		minSamples = 10
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	maxLag := req.MaxLag
	if maxLag <= 0 {
		maxLag = 6
	}
	if maxLag > 60 {
		maxLag = 60
	}

	bucketSec := req.BucketSec
	if bucketSec <= 0 {
		bucketSec = resolveAutoBucketSec(req.StartTime, req.EndTime, 0)
	}

	res := CorrelationDiscoverRes{
		Items:     []CorrelationDiscoverItem{},
		BucketSec: bucketSec,
	}

	nodeFilter := make(map[string]bool)
	for _, nid := range req.NodeIDs {
		nid = strings.TrimSpace(nid)
		if nid != "" {
			nodeFilter[nid] = true
		}
	}

	// Gather target pollings
	var pollings []*datastore.PollingEnt
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode == datastore.LogModeNone {
			return true
		}
		if len(nodeFilter) > 0 && !nodeFilter[p.NodeID] && p.ID != req.BasePollingID {
			return true
		}
		pollings = append(pollings, p)
		return true
	})

	sort.Slice(pollings, func(i, j int) bool {
		return pollings[i].ID < pollings[j].ID
	})

	pollingIDs := make([]string, 0, len(pollings))
	for _, p := range pollings {
		pollingIDs = append(pollingIDs, p.ID)
	}

	emit("loading", 0, len(pollingIDs), "")
	logsMap := fetchPollingLogsBatch(ctx, pollingIDs, req.StartTime, req.EndTime, func(done, total int) {
		emit("loading", done, total, "")
	})
	if ctx.Err() != nil {
		res.Error = ctx.Err().Error()
		res.DurationMs = time.Since(startTime).Milliseconds()
		emit("canceled", 0, 0, res.Error)
		return res
	}

	seriesList := make([]*pollingBucketSeries, 0, len(pollings))
	seriesMap := make(map[string]*pollingBucketSeries, len(pollings))
	for _, p := range pollings {
		if ctx.Err() != nil {
			res.Error = ctx.Err().Error()
			res.DurationMs = time.Since(startTime).Milliseconds()
			emit("canceled", 0, 0, res.Error)
			return res
		}
		logs := logsMap[p.ID]
		if len(logs) < minSamples {
			continue
		}
		vars := getCorrelationKeys(p, logs, nil, req.IncludeState)
		if len(vars) == 0 {
			continue
		}
		nodeName := ""
		if n := datastore.GetNode(p.NodeID); n != nil {
			nodeName = n.Name
		}
		bs := buildBucketSeriesFromLogs(p, nodeName, logs, vars, bucketSec, req.StartTime, req.EndTime)
		if bs == nil || len(bs.Timestamps) < minSamples {
			continue
		}
		seriesList = append(seriesList, bs)
		seriesMap[p.ID] = bs
	}
	res.TotalPollings = len(seriesList)

	type pairTask struct {
		s1 *pollingBucketSeries
		s2 *pollingBucketSeries
	}
	var pairs []pairTask
	if req.BasePollingID != "" {
		base, ok := seriesMap[req.BasePollingID]
		if !ok {
			res.Error = "base polling has insufficient logs"
			res.DurationMs = time.Since(startTime).Milliseconds()
			emit("done", 0, 0, res.Error)
			return res
		}
		for _, other := range seriesList {
			if other.PollingID == base.PollingID {
				continue
			}
			if req.CrossNodeOnly && base.NodeID == other.NodeID {
				continue
			}
			pairs = append(pairs, pairTask{s1: base, s2: other})
		}
	} else {
		for i := 0; i < len(seriesList); i++ {
			for j := i + 1; j < len(seriesList); j++ {
				if req.CrossNodeOnly && seriesList[i].NodeID == seriesList[j].NodeID {
					continue
				}
				pairs = append(pairs, pairTask{s1: seriesList[i], s2: seriesList[j]})
			}
		}
	}
	res.EvaluatedPairs = len(pairs)
	if len(pairs) == 0 {
		res.DurationMs = time.Since(startTime).Milliseconds()
		emit("done", 0, 0, "")
		return res
	}

	emit("calculating", 0, len(pairs), "")
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if workers > len(pairs) {
		workers = len(pairs)
	}

	var mu sync.Mutex
	var found []CorrelationDiscoverItem
	var wg sync.WaitGroup
	var donePairs atomic.Int64
	totalPairs := len(pairs)
	chunkSize := (totalPairs + workers - 1) / workers

	for w := 0; w < workers; w++ {
		startIdx := w * chunkSize
		endIdx := startIdx + chunkSize
		if endIdx > totalPairs {
			endIdx = totalPairs
		}
		if startIdx >= endIdx {
			continue
		}
		wg.Add(1)
		go func(slice []pairTask) {
			defer wg.Done()
			local := make([]CorrelationDiscoverItem, 0)
			for _, pair := range slice {
				if ctx.Err() != nil {
					break
				}
				item, ok := evaluatePairForDiscovery(pair.s1, pair.s2, algo, bucketSec, maxLag, minSamples)
				if ok && item.AbsScore >= minAbs {
					local = append(local, item)
				}
				c := int(donePairs.Add(1))
				if c == totalPairs || c%5 == 0 {
					emit("calculating", c, totalPairs, "")
				}
			}
			if len(local) > 0 {
				mu.Lock()
				found = append(found, local...)
				mu.Unlock()
			}
		}(pairs[startIdx:endIdx])
	}
	wg.Wait()

	if ctx.Err() != nil {
		res.Error = ctx.Err().Error()
		res.DurationMs = time.Since(startTime).Milliseconds()
		emit("canceled", int(donePairs.Load()), totalPairs, res.Error)
		return res
	}

	sort.Slice(found, func(i, j int) bool {
		if math.Abs(found[i].AbsScore-found[j].AbsScore) > 1e-12 {
			return found[i].AbsScore > found[j].AbsScore
		}
		if found[i].SampleCount != found[j].SampleCount {
			return found[i].SampleCount > found[j].SampleCount
		}
		if found[i].PollingID1 != found[j].PollingID1 {
			return found[i].PollingID1 < found[j].PollingID1
		}
		return found[i].PollingID2 < found[j].PollingID2
	})

	if len(found) > limit {
		found = found[:limit]
	}
	res.Items = found
	res.DurationMs = time.Since(startTime).Milliseconds()
	emit("done", totalPairs, totalPairs, "")
	return res
}
