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
	"github.com/twsnmp/twsnmpfk/ping"
)

func doPollingPing(pe *datastore.PollingEnt) {
	if pe.Mode == "line" {
		doPollingCheckLineCond(pe)
		return
	} else if pe.Mode == "smoke" {
		doPollingPingSmoke(pe)
		return
	}
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("ping", pe, fmt.Errorf("node not found"))
		return
	}
	size := 64
	ttl := 0
	if pe.Params != "" {
		if strings.Contains(pe.Params, "=") {
			a := strings.Split(pe.Params, ",")
			for _, s := range a {
				b := strings.SplitN(s, "=", 2)
				if len(b) == 2 {
					if i, err := strconv.Atoi(b[1]); err == nil {
						if b[0] == "ttl" && i > 0 && i < 256 {
							ttl = i
						} else if b[0] == "size" && i >= 0 && i < 3000 {
							size = i
						}
					}
				}
			}
		} else {
			if i, err := strconv.Atoi(pe.Params); err == nil {
				size = i
			}
		}
	}
	r := ping.DoPing(n.IP, pe.Timeout, pe.Retry, size, ttl)
	if r.Stat == ping.PingOK {
		pe.Result["rtt"] = float64(r.Time)
		pe.Result["ttl"] = float64(r.RecvTTL)
		delete(pe.Result, "error")
		setPollingState(pe, "normal")
	} else {
		pe.Result["rtt"] = 0.0
		pe.Result["ttl"] = 0.0
		pe.Result["error"] = fmt.Sprintf("%v", r.Error)
		setPollingState(pe, pe.Level)
	}
}

func doPollingCheckLineCond(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("ping", pe, fmt.Errorf("node not found"))
		return
	}
	lastError := ""
	speed := []float64{}
	rtt := []float64{}
	fail := 0
	ttl := 0
	for i := 0; i < 20; i++ {
		r64 := ping.DoPing(n.IP, pe.Timeout, pe.Retry, 64, 0)
		if r64.Stat != ping.PingOK {
			lastError = fmt.Sprintf("%v", r64.Error)
			fail += 1
			continue
		}
		r1364 := ping.DoPing(n.IP, pe.Timeout, pe.Retry, 1364, 0)
		if r1364.Stat != ping.PingOK {
			lastError = fmt.Sprintf("%v", r1364.Error)
			fail += 1
			continue
		}
		if r64.Time == r1364.Time {
			fail += 1
			continue
		}
		ttl = r64.RecvTTL
		a := float64(64.0-1364.0) / float64(r64.Time-r1364.Time)
		b := float64(r64.Time) - a*float64(64.0)
		s := a * (8.0 * 1000.0) //Mbps
		if s > 0.0 && s < 1000.0 && b > 0.0 {
			rtt = append(rtt, b)
			speed = append(speed, s)
			if len(speed) >= 5 {
				break
			}
		} else {
			fail += 1
		}
	}
	if len(speed) < 3 {
		pe.Result["error"] = lastError
		pe.Result["rtt"] = 0.0
		pe.Result["ttl"] = ttl
		pe.Result["rtt_cv"] = 0.0
		pe.Result["speed"] = 0.0
		pe.Result["speed_cv"] = 0.0
		pe.Result["fail"] = float64(fail)
		setPollingState(pe, pe.Level)
		return
	}
	// 5回の測定から平均値と変動係数を計算
	rm, rcv := calcMeanCV(rtt)
	pe.Result["rtt"] = rm
	pe.Result["rtt_cv"] = rcv
	pe.Result["ttl"] = ttl
	sm, scv := calcMeanCV(speed)
	pe.Result["speed"] = sm
	pe.Result["speed_cv"] = scv
	pe.Result["fail"] = float64(fail)
	delete(pe.Result, "error")
	setPollingState(pe, "normal")
}

func calcMeanCV(a []float64) (float64, float64) {
	if len(a) < 1 {
		return 0.0, 0.0
	}
	n := float64(len(a))
	m := float64(0.0)
	for _, d := range a {
		m += d
	}
	m /= n
	if m == 0.0 {
		return 0.0, 0.0
	}
	v := float64(0.0)
	for _, d := range a {
		v += (d - m) * (d - m)
	}
	v /= n
	sigma := math.Sqrt(v)
	return m, sigma / m
}

func doPollingPingSmoke(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("ping", pe, fmt.Errorf("node not found"))
		return
	}
	size := 64
	ttl := 0
	count := 10
	if pe.Params != "" {
		if strings.Contains(pe.Params, "=") {
			a := strings.Split(pe.Params, ",")
			for _, s := range a {
				b := strings.SplitN(s, "=", 2)
				if len(b) == 2 {
					if i, err := strconv.Atoi(b[1]); err == nil {
						if b[0] == "ttl" && i > 0 && i < 256 {
							ttl = i
						} else if b[0] == "size" && i >= 0 && i < 3000 {
							size = i
						} else if b[0] == "count" && i > 0 && i <= 100 {
							count = i
						}
					}
				}
			}
		} else {
			if i, err := strconv.Atoi(pe.Params); err == nil {
				size = i
			}
		}
	}

	valids := []float64{}
	lossCount := 0
	lastTTL := 0
	lastErr := ""

	for i := 0; i < count; i++ {
		if i > 0 {
			time.Sleep(200 * time.Millisecond)
		}
		r := ping.DoPing(n.IP, pe.Timeout, pe.Retry, size, ttl)
		if r.Stat == ping.PingOK {
			valids = append(valids, float64(r.Time))
			lastTTL = r.RecvTTL
		} else {
			lossCount++
			if r.Error != nil {
				lastErr = fmt.Sprintf("%v", r.Error)
			} else {
				lastErr = "ping failed"
			}
		}
	}

	totalCount := float64(count)
	lossRate := (float64(lossCount) / totalCount) * 100.0

	var minRtt, maxRtt, avgRtt, medianRtt, jitterRtt float64
	if len(valids) > 0 {
		sort.Float64s(valids)
		minRtt = valids[0]
		maxRtt = valids[len(valids)-1]
		sum := 0.0
		for _, v := range valids {
			sum += v
		}
		avgRtt = sum / float64(len(valids))
		midIdx := len(valids) / 2
		if len(valids)%2 == 0 {
			medianRtt = (valids[midIdx-1] + valids[midIdx]) / 2.0
		} else {
			medianRtt = valids[midIdx]
		}
		jitterRtt = maxRtt - minRtt
	}

	pe.Result["rtt"] = avgRtt
	pe.Result["min"] = minRtt
	pe.Result["max"] = maxRtt
	pe.Result["avg"] = avgRtt
	pe.Result["mean"] = avgRtt
	pe.Result["median"] = medianRtt
	pe.Result["jitter"] = jitterRtt
	pe.Result["loss"] = lossRate
	pe.Result["fail"] = float64(lossCount)
	pe.Result["lossCount"] = float64(lossCount)
	pe.Result["count"] = totalCount
	pe.Result["ttl"] = float64(lastTTL)

	if len(valids) == 0 {
		pe.Result["error"] = lastErr
	} else {
		delete(pe.Result, "error")
	}

	script := pe.Script
	if script != "" {
		vm := otto.New()
		for k, v := range pe.Result {
			vm.Set(k, v)
		}
		vm.Set("interval", pe.PollInt)
		value, err := vm.Run(script)
		if err != nil {
			setPollingError("ping", pe, fmt.Errorf("invalid script err=%v", err))
			return
		}
		if ok, _ := value.ToBoolean(); ok {
			setPollingState(pe, "normal")
		} else {
			setPollingState(pe, pe.Level)
		}
	} else {
		if lossRate < 100.0 {
			setPollingState(pe, "normal")
		} else {
			setPollingState(pe, pe.Level)
		}
	}
}
