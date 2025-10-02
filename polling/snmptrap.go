package polling

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func doPollingTrap(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingSnmpTrapStats(pe)
	default:
		doPollingTrapCount(pe)
	}
}

func doPollingTrapCount(pe *datastore.PollingEnt) {
	var err error
	var regexFilter *regexp.Regexp
	host := pe.Params
	filter := pe.Filter
	script := pe.Script
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			setPollingError("trap", pe, fmt.Errorf("invalid log watch format"))
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
	setVMFuncAndValues(pe, vm)
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
		setPollingError("trap", pe, fmt.Errorf("invalid script err=%v", err))
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

func doPollingSnmpTrapStats(pe *datastore.PollingEnt) {
	script := pe.Script
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	typeMap := make(map[string]int)
	fromMap := make(map[string]int)
	typeFromMap := make(map[string]int)
	datastore.ForEachLastTraps(func(l *datastore.TrapEnt) bool {
		if l.Time < st {
			return false
		}
		count++
		trapType := l.TrapType
		fa := l.FromAddress
		typeMap[trapType]++
		fromMap[fa]++
		typeFromMap[trapType+":"+fa]++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["types"] = float64(len(typeMap))
	pe.Result["froms"] = float64(len(fromMap))
	pe.Result["typeFroms"] = float64(len(typeFromMap))
	pe.Result["count"] = float64(count)
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
		setPollingError("trap", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}
