package polling

import (
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func doPollingArpLog(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingArpLogStats(pe)
	default:
		doPollingArpLogCount(pe)
	}
}

func doPollingArpLogCount(pe *datastore.PollingEnt) {
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
	setVMFuncAndValues(pe, vm)
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

func doPollingArpLogStats(pe *datastore.PollingEnt) {
	script := pe.Script
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	patternMap := make(map[string]int)
	stateMap := make(map[string]int)
	ipMap := make(map[string]int)
	macMap := make(map[string]int)
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		if l.Time < st {
			return false
		}
		state := l.State
		ip := l.IP
		mac := l.NewMAC
		patternMap[ip+mac]++
		stateMap[state]++
		ipMap[ip]++
		macMap[mac]++
		count++
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	pe.Result["pattern"] = float64(len(patternMap))
	pe.Result["states"] = float64(len(stateMap))
	pe.Result["IPs"] = float64(len(ipMap))
	pe.Result["MACs"] = float64(len(macMap))
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
		setPollingError("arplog", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}
