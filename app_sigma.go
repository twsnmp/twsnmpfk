package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/pkg/sigma"
)

// SigmaRuleCount stores hit count for a specific rule
type SigmaRuleCount struct {
	ID     string   `json:"ID"`
	Title  string   `json:"Title"`
	Level  string   `json:"Level"`
	Source string   `json:"Source"`
	Count  int      `json:"Count"`
	Tags   []string `json:"Tags"`
}

// SigmaTagCount stores hit count for a tag
type SigmaTagCount struct {
	Tag      string `json:"Tag"`
	Category string `json:"Category"` // "mitre", "compliance", or "general"
	Count    int    `json:"Count"`
}

// SigmaLogSourceCount stores hit count for a logsource product:service
type SigmaLogSourceCount struct {
	LogSource string `json:"LogSource"`
	Count     int    `json:"Count"`
}

// SigmaTimelinePoint represents threat count in a time bucket
type SigmaTimelinePoint struct {
	Time     int64 `json:"Time"`
	Count    int   `json:"Count"`
	Critical int   `json:"Critical"`
	High     int   `json:"High"`
	Medium   int   `json:"Medium"`
	Low      int   `json:"Low"`
	Info     int   `json:"Info"`
}

// SigmaReportStats aggregates summary metrics for Sigma report
type SigmaReportStats struct {
	TotalLogs       int                   `json:"TotalLogs"`
	HitLogs         int                   `json:"HitLogs"`
	TotalDetections int                   `json:"TotalDetections"`
	ActiveRules     int                   `json:"ActiveRules"`
	Critical        int                   `json:"Critical"`
	High            int                   `json:"High"`
	Medium          int                   `json:"Medium"`
	Low             int                   `json:"Low"`
	Informational   int                   `json:"Informational"`
	ComplianceHits  int                   `json:"ComplianceHits"`
	TopRules        []SigmaRuleCount      `json:"TopRules"`
	TopTags         []SigmaTagCount       `json:"TopTags"`
	LogSources      []SigmaLogSourceCount `json:"LogSources"`
	Timeline        []SigmaTimelinePoint  `json:"Timeline"`
}

// SigmaHitItem represents a single matched event
type SigmaHitItem struct {
	Time         int64    `json:"Time"`
	RuleID       string   `json:"RuleID"`
	Title        string   `json:"Title"`
	Level        string   `json:"Level"`
	Source       string   `json:"Source"`
	LogSource    string   `json:"LogSource"`
	Tags         []string `json:"Tags"`
	Host         string   `json:"Host"`
	Tag          string   `json:"Tag"`
	Message      string   `json:"Message"`
	Log          string   `json:"Log"`
	IsCompliance bool     `json:"IsCompliance"`
}

// SigmaReportResult is returned to frontend
type SigmaReportResult struct {
	Stats SigmaReportStats `json:"Stats"`
	Items []*SigmaHitItem  `json:"Items"`
}

// GetSigmaPacks returns list of all available embedded rule packs
func (a *App) GetSigmaPacks() []*sigma.SigmaPackInfo {
	return sigma.GetAllSigmaPacksInfo()
}

func isComplianceRule(entry *sigma.SigmaRuleEntry) bool {
	if strings.Contains(strings.ToLower(entry.Source), "compliance") {
		return true
	}
	for _, tag := range entry.Evaluator.Rule.Tags {
		t := strings.ToLower(tag)
		if strings.HasPrefix(t, "compliance.") ||
			strings.HasPrefix(t, "pci_dss") ||
			strings.HasPrefix(t, "nist") ||
			strings.HasPrefix(t, "gdpr") ||
			strings.HasPrefix(t, "cis") ||
			strings.HasPrefix(t, "hipaa") ||
			strings.HasPrefix(t, "tsc_") {
			return true
		}
	}
	return false
}

func classifyTag(tag string) string {
	t := strings.ToLower(tag)
	if strings.HasPrefix(t, "attack.") {
		return "mitre"
	}
	if strings.HasPrefix(t, "compliance.") ||
		strings.HasPrefix(t, "pci_dss") ||
		strings.HasPrefix(t, "nist") ||
		strings.HasPrefix(t, "gdpr") ||
		strings.HasPrefix(t, "cis") ||
		strings.HasPrefix(t, "hipaa") ||
		strings.HasPrefix(t, "tsc_") {
		return "compliance"
	}
	return "general"
}

// getCustomSigmaDir returns custom rules directory in datastore if exists
func getCustomSigmaDir() (string, string) {
	dsPath := datastore.GetDataStorePath()
	if dsPath == "" {
		return "", ""
	}
	customDir := filepath.Join(dsPath, "sigma")
	if fi, err := os.Stat(customDir); err == nil && fi.IsDir() {
		confDir := filepath.Join(customDir, "config")
		if cfi, err := os.Stat(confDir); err == nil && cfi.IsDir() {
			return customDir, confDir
		}
		return customDir, ""
	}
	return "", ""
}

// AnalyzeSigmaLogs runs Sigma threat detection and compliance audit against syslogs
func (a *App) AnalyzeSigmaLogs(logs []*datastore.SyslogEnt, packs []string, rules []string) (*SigmaReportResult, error) {
	result := &SigmaReportResult{
		Stats: SigmaReportStats{
			TotalLogs: len(logs),
		},
		Items: []*SigmaHitItem{},
	}

	customRulesDir, customConfigDir := getCustomSigmaDir()

	cfg := sigma.Config{
		RulesPath: customRulesDir,
		ConfigDir: customConfigDir,
		Packs:     packs,
		Rules:     rules,
		Strict:    false,
	}

	eng, err := sigma.NewEngine(cfg)
	if err != nil {
		return result, fmt.Errorf("failed to initialize Sigma engine: %w", err)
	}

	result.Stats.ActiveRules = len(eng.RuleEntries)
	if len(logs) == 0 {
		return result, nil
	}

	// Sort logs chronologically for correlation sliding-window accuracy
	sortedLogs := make([]*datastore.SyslogEnt, len(logs))
	copy(sortedLogs, logs)
	sort.Slice(sortedLogs, func(i, j int) bool {
		return sortedLogs[i].Time < sortedLogs[j].Time
	})

	ruleHitCount := make(map[string]*SigmaRuleCount)
	tagHitCount := make(map[string]int)
	logSourceHitCount := make(map[string]int)

	// Timeline bucket map: 1 hour buckets
	timelineMap := make(map[int64]*SigmaTimelinePoint)

	for _, l := range sortedLogs {
		rawMsg := l.Message
		if l.Tag != "" {
			rawMsg = l.Tag + ": " + l.Message
		}
		fullLog := fmt.Sprintf("%s %s %s", l.Host, l.Tag, l.Message)

		extra := map[string]interface{}{
			"host":     l.Host,
			"tag":      l.Tag,
			"facility": l.Facility,
			"severity": l.Severity,
			"client":   l.Host,
			"srcip":    l.Host,
			"message":  l.Message,
		}

		matchedEntries := eng.MatchAllWithExtra(rawMsg, extra, l.Time)
		if len(matchedEntries) == 0 {
			continue
		}

		result.Stats.HitLogs++
		result.Stats.TotalDetections += len(matchedEntries)

		bucketTime := (l.Time / (int64(time.Hour))) * int64(time.Hour)
		tp, exists := timelineMap[bucketTime]
		if !exists {
			tp = &SigmaTimelinePoint{Time: bucketTime}
			timelineMap[bucketTime] = tp
		}

		for _, entry := range matchedEntries {
			ev := entry.Evaluator
			level := strings.ToLower(ev.Level)
			if level == "" {
				level = "medium"
			}
			isComp := isComplianceRule(entry)
			if isComp {
				result.Stats.ComplianceHits++
			}

			tp.Count++
			switch level {
			case "critical":
				result.Stats.Critical++
				tp.Critical++
			case "high":
				result.Stats.High++
				tp.High++
			case "medium":
				result.Stats.Medium++
				tp.Medium++
			case "low":
				result.Stats.Low++
				tp.Low++
			default:
				result.Stats.Informational++
				tp.Info++
			}

			// Rule count
			rc, ok := ruleHitCount[ev.ID]
			if !ok {
				rc = &SigmaRuleCount{
					ID:     ev.ID,
					Title:  ev.Title,
					Level:  ev.Level,
					Source: entry.Source,
					Count:  0,
					Tags:   ev.Rule.Tags,
				}
				ruleHitCount[ev.ID] = rc
			}
			rc.Count++

			// Tags count
			for _, tag := range ev.Rule.Tags {
				tagHitCount[tag]++
			}

			// LogSource count
			lsKey := fmt.Sprintf("%s:%s", ev.Logsource.Product, ev.Logsource.Service)
			if ev.Logsource.Category != "" {
				lsKey = fmt.Sprintf("%s:%s:%s", ev.Logsource.Product, ev.Logsource.Category, ev.Logsource.Service)
			}
			logSourceHitCount[lsKey]++

			// Add hit item
			result.Items = append(result.Items, &SigmaHitItem{
				Time:         l.Time,
				RuleID:       ev.ID,
				Title:        ev.Title,
				Level:        level,
				Source:       entry.Source,
				LogSource:    lsKey,
				Tags:         ev.Rule.Tags,
				Host:         l.Host,
				Tag:          l.Tag,
				Message:      l.Message,
				Log:          fullLog,
				IsCompliance: isComp,
			})
		}
	}

	// Prepare TopRules
	topRules := make([]SigmaRuleCount, 0, len(ruleHitCount))
	for _, rc := range ruleHitCount {
		topRules = append(topRules, *rc)
	}
	sort.Slice(topRules, func(i, j int) bool {
		return topRules[i].Count > topRules[j].Count
	})
	result.Stats.TopRules = topRules

	// Prepare TopTags
	topTags := make([]SigmaTagCount, 0, len(tagHitCount))
	for tag, cnt := range tagHitCount {
		topTags = append(topTags, SigmaTagCount{
			Tag:      tag,
			Category: classifyTag(tag),
			Count:    cnt,
		})
	}
	sort.Slice(topTags, func(i, j int) bool {
		return topTags[i].Count > topTags[j].Count
	})
	result.Stats.TopTags = topTags

	// Prepare LogSources
	logSources := make([]SigmaLogSourceCount, 0, len(logSourceHitCount))
	for ls, cnt := range logSourceHitCount {
		logSources = append(logSources, SigmaLogSourceCount{
			LogSource: ls,
			Count:     cnt,
		})
	}
	sort.Slice(logSources, func(i, j int) bool {
		return logSources[i].Count > logSources[j].Count
	})
	result.Stats.LogSources = logSources

	// Prepare Timeline
	timeline := make([]SigmaTimelinePoint, 0, len(timelineMap))
	for _, tp := range timelineMap {
		timeline = append(timeline, *tp)
	}
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Time < timeline[j].Time
	})
	result.Stats.Timeline = timeline

	return result, nil
}
