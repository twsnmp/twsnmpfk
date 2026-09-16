package polling

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/pkg/sigma"
	"github.com/vjeantet/grok"
)

func doPollingSyslog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingSyslogStats(pe)
	case "pri":
		doPollingSyslogPri(pe)
	case "sigma":
		doPollingSyslogSigma(pe)
	default:
		doPollingSyslogCount(pe)
	}
}

func doPollingSyslogPri(pe *datastore.PollingEnt) bool {
	var err error
	var regexFilter *regexp.Regexp
	filter := pe.Filter
	host := pe.Params
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
			return false
		}
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	count := 0
	priMap := make(map[int]int)
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		if l.Time < st {
			return false
		}
		if host != "" && host != l.Host {
			return true
		}
		msg := l.Type + " " + l.Tag + " " + l.Message
		if regexFilter != nil && !regexFilter.Match([]byte(msg)) {
			return true
		}
		count++
		pri := l.Facility*8 + l.Severity
		priMap[pri]++
		return true
	})
	pe.Result["lastTime"] = float64(time.Now().UnixNano())
	pe.Result["count"] = float64(count)
	for pri, c := range priMap {
		pe.Result[fmt.Sprintf("pri_%d", int(pri))] = float64(c)
	}
	setPollingState(pe, "normal")
	return true
}

func doPollingSyslogCount(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	var grokExtractor *grok.Grok
	host := pe.Params
	filter := pe.Filter
	extractor := pe.Extractor
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
			return
		}
	}
	if extractor != "" {
		grokExtractor, err = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err != nil {
			setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
			return
		}
		if err = grokExtractor.AddPattern("TWSNMP", extractor); err != nil {
			setPollingError("log", pe, fmt.Errorf("no extractor pattern"))
			return
		}
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	count := 0
	failed := false
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		if l.Time < st {
			return false
		}
		if host != "" && host != l.Host {
			return true
		}
		msg := l.Type + " " + l.Tag + " " + l.Message
		if regexFilter != nil && !regexFilter.Match([]byte(msg)) {
			return true
		}
		if grokExtractor != nil {
			values, err := grokExtractor.Parse("%{TWSNMP}", msg)
			if err != nil {
				log.Println(err)
				return true
			}
			if len(values) < 1 {
				return true
			}
			count++
			for k, v := range values {
				vm.Set(k, v)
				pe.Result[k] = v
			}
			value, err := vm.Run(script)
			if err == nil {
				if ok, _ := value.ToBoolean(); !ok {
					failed = true
					setPollingState(pe, pe.Level)
					return false
				}
			} else {
				log.Println(err)
				failed = true
				setPollingError("log", pe, err)
				return false
			}
		} else {
			count++
		}
		return true
	})
	pe.Result["lastTime"] = time.Now().UnixNano()
	pe.Result["count"] = float64(count)
	if extractor != "" {
		if !failed && count > 0 {
			setPollingState(pe, "normal")
		}
		return
	}
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm.Set("count", count)
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("log", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func doPollingSyslogStats(pe *datastore.PollingEnt) {
	script := pe.Script
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	normal := 0
	warns := 0
	errors := 0
	patternMap := make(map[string]int)
	errorPatternMap := make(map[string]int)
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		if l.Time < st {
			return false
		}
		host := l.Host
		tag := l.Tag
		message := l.Message
		sv := l.Severity
		msg := host + " " + tag + " " + message
		nl := normalizeSyslog(msg)
		patternMap[nl]++
		switch {
		case sv < 4:
			errorPatternMap[nl]++
			errors++
		case sv == 4:
			warns++
		default:
			normal++
		}
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	pe.Result["error"] = float64(errors)
	pe.Result["warn"] = float64(warns)
	pe.Result["normal"] = float64(normal)
	pe.Result["patterns"] = float64(len(patternMap))
	pe.Result["errorPatterns"] = float64(len(errorPatternMap))
	if script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("interval", pe.PollInt)
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("syslog", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

var regNum = regexp.MustCompile(`\b-?\d+(\.\d+)?\b`)
var regUUDI = regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
var regEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
var regIP = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
var regMAC = regexp.MustCompile(`\b(?:[0-9a-fA-F]{2}[:-]){5}(?:[0-9a-fA-F]{2})\b`)

func normalizeSyslog(msg string) string {
	normalized := msg
	normalized = regUUDI.ReplaceAllString(normalized, "#UUID#")
	normalized = regEmail.ReplaceAllString(normalized, "#EMAIL#")
	normalized = regIP.ReplaceAllString(normalized, "#IP#")
	normalized = regMAC.ReplaceAllString(normalized, "#MAC#")
	normalized = regNum.ReplaceAllString(normalized, "#NUM#")
	return normalized
}

func doPollingSyslogSigma(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	host := pe.Params
	filter := pe.Filter
	extractor := pe.Extractor
	script := pe.Script

	if extractor != "" {
		if regexFilter, err = regexp.Compile(extractor); err != nil {
			setPollingError("syslog", pe, fmt.Errorf("invalid extractor regex: %v", err))
			return
		}
	}

	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}

	// Parse packs and rules from filter
	packs, rules := parseSigmaFilter(filter)

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
		setPollingError("syslog", pe, fmt.Errorf("sigma engine init error: %v", err))
		return
	}

	totalDetections := 0
	hitLogs := 0
	critical := 0
	high := 0
	medium := 0
	low := 0
	info := 0
	compliance := 0
	scanned := 0
	lastRule := ""

	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		if l.Time < st {
			return false
		}
		if host != "" && host != l.Host {
			return true
		}

		rawMsg := l.Message
		if l.Tag != "" {
			rawMsg = l.Tag + ": " + l.Message
		}
		if regexFilter != nil && !regexFilter.MatchString(rawMsg) {
			return true
		}

		scanned++

		extra := map[string]interface{}{
			"host":     l.Host,
			"tag":      l.Tag,
			"facility": l.Facility,
			"severity": l.Severity,
			"client":   l.Host,
			"srcip":    l.Host,
			"message":  l.Message,
		}

		matched := eng.MatchAllWithExtra(rawMsg, extra, l.Time)
		if len(matched) == 0 {
			return true
		}

		hitLogs++
		totalDetections += len(matched)

		for _, entry := range matched {
			ev := entry.Evaluator
			level := strings.ToLower(ev.Level)
			switch level {
			case "critical":
				critical++
			case "high":
				high++
			case "medium":
				medium++
			case "low":
				low++
			default:
				info++
			}
			if isComplianceRule(entry) {
				compliance++
			}
			lastRule = ev.Title
		}

		return true
	})

	pe.Result["lastTime"] = float64(time.Now().UnixNano())
	pe.Result["count"] = float64(totalDetections)
	pe.Result["hitLogs"] = float64(hitLogs)
	pe.Result["critical"] = float64(critical)
	pe.Result["high"] = float64(high)
	pe.Result["medium"] = float64(medium)
	pe.Result["low"] = float64(low)
	pe.Result["info"] = float64(info)
	pe.Result["compliance"] = float64(compliance)
	pe.Result["scanned"] = float64(scanned)
	if lastRule != "" {
		pe.Result["lastRule"] = lastRule
	} else {
		delete(pe.Result, "lastRule")
	}

	if script == "" {
		if totalDetections == 0 {
			setPollingState(pe, "normal")
		} else {
			setPollingState(pe, pe.Level)
		}
		return
	}

	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	vm.Set("count", float64(totalDetections))
	vm.Set("hitLogs", float64(hitLogs))
	vm.Set("critical", float64(critical))
	vm.Set("high", float64(high))
	vm.Set("medium", float64(medium))
	vm.Set("low", float64(low))
	vm.Set("info", float64(info))
	vm.Set("compliance", float64(compliance))
	vm.Set("scanned", float64(scanned))
	vm.Set("interval", float64(pe.PollInt))

	value, err := vm.Run(script)
	if err != nil {
		setPollingError("syslog", pe, fmt.Errorf("invalid script err=%v", err))
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func parseSigmaFilter(filter string) ([]string, []string) {
	filter = strings.TrimSpace(filter)
	if filter == "" || strings.EqualFold(filter, "all") {
		return nil, nil
	}

	availPackMap := make(map[string]bool)
	for _, p := range sigma.GetAvailableSigmaPacks() {
		availPackMap[strings.ToLower(p)] = true
	}

	var packs []string
	var rules []string

	tokens := strings.FieldsFunc(filter, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	})

	for _, token := range tokens {
		t := strings.TrimSpace(token)
		if t == "" {
			continue
		}
		tl := strings.ToLower(t)
		if strings.HasPrefix(tl, "pack:") || strings.HasPrefix(tl, "packs:") || strings.HasPrefix(tl, "pack=") {
			parts := strings.SplitN(t, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(t, "=", 2)
			}
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				packs = append(packs, strings.TrimSpace(parts[1]))
			}
		} else if strings.HasPrefix(tl, "rule:") || strings.HasPrefix(tl, "rules:") || strings.HasPrefix(tl, "rule=") {
			parts := strings.SplitN(t, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(t, "=", 2)
			}
			if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
				rules = append(rules, strings.TrimSpace(parts[1]))
			}
		} else if availPackMap[tl] || tl == "all" {
			packs = append(packs, t)
		} else {
			rules = append(rules, t)
		}
	}

	return packs, rules
}

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

