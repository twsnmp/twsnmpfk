package logreport

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

// Default retention of report entities.
const (
	DefaultRetentionDays = 30
	DefaultMaxEntries    = 10000
)

type cleanMeta struct {
	ID          string  `json:"ID"`
	Count       int     `json:"Count"`
	LastTime    int64   `json:"LastTime"`
	AddressType string  `json:"AddressType"`
	Score       float64 `json:"Score"`
}

// limited kinds are also trimmed to maxEntries; the others only expire by age.
var limitedKinds = map[string]bool{
	KindWifiAP: true, KindBlueDevice: true, KindDNSQ: true, KindRADIUSFlow: true, KindTLSFlow: true,
	KindFlow: true, KindServer: true, KindFumble: true,
	KindTwAgentAlert: true, KindTwAgentFIM: true, KindTwAgentSCA: true, KindTwAgentTCP: true,
}

// Cleanup removes expired entities of the kinds fed by source.
// Entities not seen for days are deleted. For limited kinds the oldest
// (least counted) entities are removed beyond maxEntries.
// Randomized Bluetooth addresses are dropped after one day.
func Cleanup(ctx context.Context, store ReportStore, source string, days, maxEntries int) error {
	if days <= 0 {
		days = DefaultRetentionDays
	}
	if maxEntries <= 0 {
		maxEntries = DefaultMaxEntries
	}
	now := time.Now()
	delOld := now.AddDate(0, 0, -days).UnixNano()
	delRandom := now.AddDate(0, 0, -1).UnixNano()
	var firstErr error

	kinds := KindsOf(source)
	if len(kinds) == 0 && source == "" {
		kinds = AllKinds()
	}

	for _, kind := range kinds {
		items, err := store.ListLogReportData(ctx, kind)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		var ids []string
		var keep []cleanMeta
		for id, b := range items {
			var m cleanMeta
			if json.Unmarshal(b, &m) != nil {
				ids = append(ids, id)
				continue
			}
			m.ID = id
			switch {
			case m.LastTime < delOld:
				ids = append(ids, id)
			case kind == KindBlueDevice && m.LastTime < delRandom && strings.Contains(m.AddressType, " Random"):
				ids = append(ids, id)
			default:
				keep = append(keep, m)
			}
		}
		if limitedKinds[kind] && len(keep) > maxEntries {
			if kind == KindTLSFlow || kind == KindFlow || kind == KindServer || kind == KindTwAgentTCP {
				// Safer flows / servers (higher score) are dropped first.
				sort.Slice(keep, func(i, j int) bool {
					return keep[i].Score-float64(keep[i].LastTime-delOld)/float64(24*time.Hour) >
						keep[j].Score-float64(keep[j].LastTime-delOld)/float64(24*time.Hour)
				})
			} else {
				sort.Slice(keep, func(i, j int) bool {
					if keep[i].LastTime == keep[j].LastTime {
						return keep[i].Count < keep[j].Count
					}
					return keep[i].LastTime < keep[j].LastTime
				})
			}
			for i := 0; i < len(keep)-maxEntries; i++ {
				ids = append(ids, keep[i].ID)
			}
		}
		if err := store.DeleteLogReportData(ctx, kind, ids); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// AllSources returns all known report sources.
func AllSources() []string {
	return []string{SourceWifiScan, SourceBlueScan, SourcePcap, SourceWinLog, SourceTwAgent}
}

// CleanupAll cleans up expired entities for all known sources and kinds.
func CleanupAll(ctx context.Context, store ReportStore, days, maxEntries int) error {
	if store != nil && (days <= 0 || maxEntries <= 0) {
		d, l := store.GetRetentionConfig(ctx)
		if days <= 0 && d > 0 {
			days = d
		}
		if maxEntries <= 0 && l > 0 {
			maxEntries = l
		}
	}
	return Cleanup(ctx, store, "", days, maxEntries)
}
