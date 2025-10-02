package polling

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/vjeantet/grok"
)

func doPollingSyslog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingSyslogStats(pe)
	case "pri":
		doPollingSyslogPri(pe)
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
