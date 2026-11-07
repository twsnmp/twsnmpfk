package polling

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func doPollingOTel(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "trace":
		doPollingOTelTrace(pe)
	case "heartbeat", "last":
		doPollingOTelHeartbeat(pe)
	default: // "metric"
		doPollingOTelMetric(pe)
	}
}

func parseOTelParams(params string) map[string]string {
	ret := make(map[string]string)
	if params == "" {
		return ret
	}
	parts := strings.Split(params, ",")
	isKV := false
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) == 2 {
			ret[strings.ToLower(strings.TrimSpace(kv[0]))] = strings.TrimSpace(kv[1])
			isKV = true
		}
	}
	if !isKV {
		// Positional format: host,service,scope,name
		fields := []string{"host", "service", "scope", "name"}
		for i, p := range parts {
			if i < len(fields) {
				ret[fields[i]] = strings.TrimSpace(p)
			}
		}
	}
	return ret
}

func doPollingOTelMetric(pe *datastore.PollingEnt) {
	pMap := parseOTelParams(pe.Params)
	host := pMap["host"]
	service := pMap["service"]
	scope := pMap["scope"]
	name := pMap["name"]

	var targetMetric *datastore.OTelMetricEnt
	if host != "" && service != "" && scope != "" && name != "" {
		targetMetric = datastore.FindOTelMetric(host, service, scope, name)
	}
	if targetMetric == nil {
		datastore.ForEachOTelMetric(func(m *datastore.OTelMetricEnt) bool {
			if host != "" && m.Host != host {
				return true
			}
			if service != "" && m.Service != service {
				return true
			}
			if scope != "" && m.Scope != scope {
				return true
			}
			if name != "" && m.Name != name {
				return true
			}
			targetMetric = m
			return false
		})
	}

	if targetMetric == nil {
		setPollingError("otel", pe, fmt.Errorf("metric not found"))
		return
	}

	now := time.Now().UnixNano()
	var gaugeVal, sumVal, minVal, maxVal, countVal float64
	var dpTime int64

	if len(targetMetric.DataPoints) > 0 {
		dp := targetMetric.DataPoints[len(targetMetric.DataPoints)-1]
		dpTime = dp.Time
		gaugeVal = dp.Gauge
		sumVal = dp.Sum
		minVal = dp.Min
		maxVal = dp.Max
		if dp.Count > 0 {
			countVal = float64(dp.Count)
		} else {
			countVal = float64(targetMetric.Count)
		}
	} else {
		dpTime = targetMetric.Last
		countVal = float64(targetMetric.Count)
	}

	elapsedSec := float64(now-dpTime) / 1e9
	if elapsedSec < 0 {
		elapsedSec = 0
	}

	pe.Result["gauge"] = gaugeVal
	pe.Result["sum"] = sumVal
	pe.Result["min"] = minVal
	pe.Result["max"] = maxVal
	pe.Result["count"] = countVal
	pe.Result["lastTime"] = float64(dpTime)
	pe.Result["elapsed"] = elapsedSec

	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}

	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)

	val, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("otel", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := val.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingOTelTrace(pe *datastore.PollingEnt) {
	pMap := parseOTelParams(pe.Params)
	serviceFilter := pMap["service"]
	scopeFilter := pMap["scope"]
	nameFilter := pMap["name"]

	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok && int64(vf) > 0 {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()

	// Buckets to inspect: current minute and previous minutes covered by interval
	buckets := datastore.GetOTelTraceBucketList()
	reqBuckets := make(map[string]bool)
	for t := st; t <= et; t += int64(time.Minute) {
		reqBuckets[time.Unix(0, t).Format("2006-01-02T15:04")] = true
	}
	reqBuckets[time.Unix(0, et).Format("2006-01-02T15:04")] = true

	var totalSpans int
	var errorSpans int
	var durList []float64
	var maxDur float64
	var durSum float64

	for _, b := range buckets {
		if !reqBuckets[b] {
			continue
		}
		datastore.ForEachOTelTrace(b, func(t *datastore.OTelTraceEnt) bool {
			if t.End < st || t.Start > et {
				return true
			}
			for _, span := range t.Spans {
				if span.End < st || span.Start > et {
					continue
				}
				if serviceFilter != "" && span.Service != serviceFilter {
					continue
				}
				if scopeFilter != "" && span.Scope != scopeFilter {
					continue
				}
				if nameFilter != "" && span.Name != nameFilter {
					continue
				}

				totalSpans++
				durList = append(durList, span.Dur)
				durSum += span.Dur
				if span.Dur > maxDur {
					maxDur = span.Dur
				}

				// Check error condition
				if isOTelSpanError(span) {
					errorSpans++
				}
			}
			return true
		})
	}

	var errorRate float64
	var avgDur float64
	var p95Dur float64

	if totalSpans > 0 {
		errorRate = (float64(errorSpans) / float64(totalSpans)) * 100.0
		avgDur = durSum / float64(totalSpans)
		sort.Float64s(durList)
		idx95 := int(math.Ceil(float64(totalSpans)*0.95)) - 1
		if idx95 >= 0 && idx95 < len(durList) {
			p95Dur = durList[idx95]
		}
	}

	pe.Result["count"] = float64(totalSpans)
	pe.Result["error"] = float64(errorSpans)
	pe.Result["errorRate"] = errorRate
	pe.Result["avgDur"] = avgDur
	pe.Result["maxDur"] = maxDur
	pe.Result["p95Dur"] = p95Dur
	pe.Result["lastTime"] = float64(et)

	if pe.Script == "" {
		if errorSpans > 0 {
			setPollingState(pe, pe.Level)
		} else {
			setPollingState(pe, "normal")
		}
		return
	}

	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)

	val, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("otel", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := val.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func isOTelSpanError(span datastore.OTelTraceSpanEnt) bool {
	for _, attr := range span.Attributes {
		lower := strings.ToLower(attr)
		if strings.Contains(lower, "error=true") ||
			strings.Contains(lower, "status_code=status_code_error") ||
			strings.Contains(lower, "status.code=status_code_error") ||
			strings.Contains(lower, "status.code=error") {
			return true
		}
		if strings.HasPrefix(lower, "http.status_code=") || strings.HasPrefix(lower, "http.response.status_code=") {
			parts := strings.Split(lower, "=")
			if len(parts) == 2 {
				if code, err := strconv.Atoi(parts[1]); err == nil && code >= 400 {
					return true
				}
			}
		}
		if strings.HasPrefix(lower, "rpc.grpc.status_code=") {
			parts := strings.Split(lower, "=")
			if len(parts) == 2 {
				if code, err := strconv.Atoi(parts[1]); err == nil && code != 0 {
					return true
				}
			}
		}
	}
	return false
}

func doPollingOTelHeartbeat(pe *datastore.PollingEnt) {
	pMap := parseOTelParams(pe.Params)
	serviceFilter := pMap["service"]
	hostFilter := pMap["host"]

	var latestTime int64
	now := time.Now().UnixNano()

	// Check metrics
	datastore.ForEachOTelMetric(func(m *datastore.OTelMetricEnt) bool {
		if serviceFilter != "" && m.Service != serviceFilter {
			return true
		}
		if hostFilter != "" && m.Host != hostFilter {
			return true
		}
		if m.Last > latestTime {
			latestTime = m.Last
		}
		return true
	})

	// Check traces from recent buckets
	buckets := datastore.GetOTelTraceBucketList()
	if len(buckets) > 0 {
		for i := len(buckets) - 1; i >= 0 && i >= len(buckets)-3; i-- {
			datastore.ForEachOTelTrace(buckets[i], func(t *datastore.OTelTraceEnt) bool {
				for _, span := range t.Spans {
					if serviceFilter != "" && span.Service != serviceFilter {
						continue
					}
					if hostFilter != "" && span.Host != hostFilter {
						continue
					}
					if span.End > latestTime {
						latestTime = span.End
					}
				}
				return true
			})
		}
	}

	if latestTime == 0 {
		setPollingError("otel", pe, fmt.Errorf("no otel telemetry received"))
		return
	}

	elapsedSec := float64(now-latestTime) / 1e9
	if elapsedSec < 0 {
		elapsedSec = 0
	}

	pe.Result["elapsed"] = elapsedSec
	pe.Result["lastTime"] = float64(latestTime)

	thresholdSec := float64(pe.PollInt * 2)
	if pe.Filter != "" {
		if v, err := strconv.ParseFloat(pe.Filter, 64); err == nil && v > 0 {
			thresholdSec = v
		}
	}

	if pe.Script == "" {
		if elapsedSec > thresholdSec {
			setPollingState(pe, pe.Level)
		} else {
			setPollingState(pe, "normal")
		}
		return
	}

	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("threshold", thresholdSec)
	vm.Set("interval", pe.PollInt)

	val, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("otel", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := val.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}
