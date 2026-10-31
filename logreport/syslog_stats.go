package logreport

import (
	"time"
)

// ProcessSyslogStat aggregates statistics for host, tag, facility, and severity.
func (s *Session) ProcessSyslogStat(r Record, facility int) {
	if r.Host == "" && r.Tag == "" {
		return
	}
	now := time.Now().UnixNano()
	sum := getEnt[SyslogStatsSummary](s, KindSyslogStats, "summary")
	if sum == nil {
		sum = &SyslogStatsSummary{
			ID:          "summary",
			Hosts:       make(map[string]*SyslogHostStat),
			Tags:        make(map[string]*SyslogTagStat),
			Facilities:  make(map[int]int64),
			Severities:  make(map[int]int64),
			FirstTime:   r.Time,
			LastTime:    r.Time,
			UpdateTime:  now,
		}
	}
	if sum.Hosts == nil {
		sum.Hosts = make(map[string]*SyslogHostStat)
	}
	if sum.Tags == nil {
		sum.Tags = make(map[string]*SyslogTagStat)
	}
	if sum.Facilities == nil {
		sum.Facilities = make(map[int]int64)
	}
	if sum.Severities == nil {
		sum.Severities = make(map[int]int64)
	}

	sum.Total++
	sum.Facilities[facility]++
	sum.Severities[r.Severity]++
	if r.Time > sum.LastTime {
		sum.LastTime = r.Time
	}
	if sum.FirstTime == 0 || r.Time < sum.FirstTime {
		sum.FirstTime = r.Time
	}
	sum.UpdateTime = now

	// Severity category (0-3: Error, 4: Warn, 5-7: Normal)
	isErr := r.Severity <= 3
	isWarn := r.Severity == 4

	if isErr {
		sum.ErrorCount++
	} else if isWarn {
		sum.WarnCount++
	} else {
		sum.NormalCount++
	}

	// Host stats
	if r.Host != "" {
		hs, ok := sum.Hosts[r.Host]
		if !ok {
			node := s.lookupNode(r.Host)
			hs = &SyslogHostStat{
				Host:      r.Host,
				NodeName:  node.Name,
				FirstTime: r.Time,
				LastTime:  r.Time,
			}
			sum.Hosts[r.Host] = hs
		}
		hs.Count++
		if isErr {
			hs.ErrorCount++
		} else if isWarn {
			hs.WarnCount++
		} else {
			hs.NormalCount++
		}
		if r.Time > hs.LastTime {
			hs.LastTime = r.Time
		}
	}

	// Tag stats
	if r.Tag != "" {
		ts, ok := sum.Tags[r.Tag]
		if !ok {
			ts = &SyslogTagStat{
				Tag:       r.Tag,
				FirstTime: r.Time,
				LastTime:  r.Time,
			}
			sum.Tags[r.Tag] = ts
		}
		ts.Count++
		if r.Time > ts.LastTime {
			ts.LastTime = r.Time
		}
	}

	putEnt(s, KindSyslogStats, "summary", sum)
	s.Processed++
}
