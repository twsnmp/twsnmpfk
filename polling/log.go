package polling

// LOG監視ポーリング処理

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/vjeantet/grok"
)

func doPollingSyslog(pe *datastore.PollingEnt) {
	switch pe.Mode {
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
	addJavaScriptFunctions(pe, vm)
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
			values, err := grokExtractor.Parse("%%{TWSNMP}", msg)
			if err != nil {
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
				failed = true
				setPollingError("log", pe, fmt.Errorf("invalid script"))
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
		if !failed {
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

func doPollingArpLog(pe *datastore.PollingEnt) {
	var err error
	filter := pe.Filter
	script := pe.Script
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	count := 0
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		if l.Time < st {
			return false
		}
		if filter != "" && l.State != filter {
			return true
		}
		count++
		return true
	})
	pe.Result["lastTime"] = time.Now().UnixNano()
	pe.Result["count"] = float64(count)
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

func doPollingTrap(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	host := pe.Params
	filter := pe.Filter
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("log", pe, fmt.Errorf("invalid log watch format"))
			return
		}
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	vm := otto.New()
	addJavaScriptFunctions(pe, vm)
	count := 0
	datastore.ForEachLastTraps(func(l *datastore.TrapEnt) bool {
		if l.Time < st {
			return false
		}
		if host != "" && l.FromAddress != host {
			return true
		} else if host == "" && pe.NodeID != getNodeIDFromTrapFromAddr(l.FromAddress) {
			return true
		}
		msg := l.TrapType + " " + l.Variables
		if regexFilter != nil && !regexFilter.Match([]byte(msg)) {
			return true
		}
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
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

func getNodeIDFromTrapFromAddr(fa string) string {
	a := strings.Split(fa, "(")
	if len(a) < 1 {
		return ""
	}
	n := datastore.FindNodeFromIP(a[0])
	if n != nil {
		return n.ID
	}
	return ""
}
