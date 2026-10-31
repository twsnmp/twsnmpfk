package logreport

import (
	"strings"
	"time"
)

// TrapRecord represents a normalized SNMP trap event for report aggregation.
type TrapRecord struct {
	Time        int64
	FromAddress string
	TrapType    string
	Enterprise  string
	Variables   string
	Level       string
}

// ProcessTrapStat aggregates SNMP trap statistics for host, type, enterprise, and severity.
func (s *Session) ProcessTrapStat(tr *TrapRecord) {
	if tr == nil || (tr.FromAddress == "" && tr.TrapType == "") {
		return
	}
	now := time.Now().UnixNano()
	sum := getEnt[TrapStatsSummary](s, KindTrapStats, "summary")
	if sum == nil {
		sum = &TrapStatsSummary{
			ID:          "summary",
			Hosts:       make(map[string]*TrapHostStat),
			Types:       make(map[string]*TrapTypeStat),
			Enterprises: make(map[string]int64),
			FirstTime:   tr.Time,
			LastTime:    tr.Time,
			UpdateTime:  now,
		}
	}
	if sum.Hosts == nil {
		sum.Hosts = make(map[string]*TrapHostStat)
	}
	if sum.Types == nil {
		sum.Types = make(map[string]*TrapTypeStat)
	}
	if sum.Enterprises == nil {
		sum.Enterprises = make(map[string]int64)
	}

	sum.Total++
	if tr.Enterprise != "" {
		sum.Enterprises[tr.Enterprise]++
	}
	if tr.Time > sum.LastTime {
		sum.LastTime = tr.Time
	}
	if sum.FirstTime == 0 || tr.Time < sum.FirstTime {
		sum.FirstTime = tr.Time
	}
	sum.UpdateTime = now

	// Severity evaluation
	isErr, isWarn := isTrapErrorOrWarn(tr.Level, tr.TrapType, tr.Variables)
	if isErr {
		sum.ErrorCount++
	} else if isWarn {
		sum.WarnCount++
	} else {
		sum.NormalCount++
	}

	// Host stats
	if tr.FromAddress != "" {
		hs, ok := sum.Hosts[tr.FromAddress]
		if !ok {
			node := s.lookupNode(tr.FromAddress)
			hs = &TrapHostStat{
				Host:      tr.FromAddress,
				NodeName:  node.Name,
				FirstTime: tr.Time,
				LastTime:  tr.Time,
			}
			sum.Hosts[tr.FromAddress] = hs
		}
		hs.Count++
		if isErr {
			hs.ErrorCount++
		} else if isWarn {
			hs.WarnCount++
		} else {
			hs.NormalCount++
		}
		if tr.Time > hs.LastTime {
			hs.LastTime = tr.Time
		}
	}

	// Type / OID stats
	if tr.TrapType != "" {
		ts, ok := sum.Types[tr.TrapType]
		if !ok {
			ts = &TrapTypeStat{
				TrapType:  tr.TrapType,
				FirstTime: tr.Time,
				LastTime:  tr.Time,
			}
			sum.Types[tr.TrapType] = ts
		}
		ts.Count++
		if tr.Time > ts.LastTime {
			ts.LastTime = tr.Time
		}
	}

	putEnt(s, KindTrapStats, "summary", sum)
	s.Processed++
}

func isTrapErrorOrWarn(level, trapType, vars string) (isErr bool, isWarn bool) {
	l := strings.ToLower(level)
	t := strings.ToLower(trapType)
	v := strings.ToLower(vars)

	if l == "high" || l == "low" || l == "error" || l == "crit" || l == "emergency" || l == "alert" || l == "down" {
		return true, false
	}
	if strings.Contains(t, "linkdown") || strings.Contains(t, "fail") || strings.Contains(t, "error") || strings.Contains(t, "crit") || strings.Contains(t, "down") {
		return true, false
	}
	if strings.Contains(v, "down") || strings.Contains(v, "error") || strings.Contains(v, "fail") || strings.Contains(v, "crit") {
		return true, false
	}

	if l == "warn" || l == "warning" {
		return false, true
	}
	if strings.Contains(t, "authenticationfailure") || strings.Contains(t, "coldstart") || strings.Contains(t, "warmstart") || strings.Contains(t, "warn") {
		return false, true
	}
	return false, false
}
