// Package polling : ポーリング処理
package polling

/*
polling.go :ポーリング処理を行う
ポーリングの種類は
(1)能動的なポーリング
 ping
 snmp - sysUptime,ifOperStatus,
 http
 https
 tls
 dns
（２）受動的なポーリング
 syslog
 snmp trap

*/

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

const maxPolling = 300

var (
	doPollingCh  chan string
	busyPollings sync.Map
)

func Start(ctx context.Context, wg *sync.WaitGroup) error {
	doPollingCh = make(chan string, maxPolling)
	wg.Add(1)
	go pollingBackend(ctx, wg)
	return nil
}

func GetAutoPollings(n *datastore.NodeEnt, pt *datastore.PollingTemplateEnt) []*datastore.PollingEnt {
	switch pt.Type {
	case "snmp":
		return getAutoSnmpPollings(n, pt)
	default:
		log.Printf("polling not supported type=%s", pt.Type)
	}
	return nil
}

func PollNowNode(nodeID string) {
	n := datastore.GetNode(nodeID)
	if n == nil {
		return
	}
	n.State = "unknown"
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.NodeID == nodeID && pe.State != "normal" && pe.Level != "off" {
			pe.State = "unknown"
			pe.NextTime = 0
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    i18n.Trans("re check polling:") + pe.Name,
			})
			datastore.UpdatePolling(pe, false)
			doPollingCh <- pe.ID
		}
		return true
	})
	datastore.SetNodeStateChanged(n.ID)
}

func CheckAllPoll() {
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.State != "normal" && pe.Level != "off" {
			pe.State = "unknown"
			pe.NextTime = 0
			n := datastore.GetNode(pe.NodeID)
			if n == nil {
				return true
			}
			n.State = "unknown"
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    i18n.Trans("re check polling:") + pe.Name,
			})
			datastore.SetNodeStateChanged(n.ID)
			datastore.UpdatePolling(pe, false)
			doPollingCh <- pe.ID
		}
		return true
	})
}

// pollingBackend :  ポーリングのバックグランド処理
func pollingBackend(ctx context.Context, wg *sync.WaitGroup) {
	log.Println("start polling")
	defer wg.Done()
	time.Sleep(time.Millisecond * 100)
	timer := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			log.Println("stop polling")
			return
		case <-timer.C:
			checkPolling()
		case id := <-doPollingCh:
			pe := datastore.GetPolling(id)
			if pe != nil && pe.NextTime <= time.Now().UnixNano() {
				if _, busy := busyPollings.Load(id); !busy {
					busyPollings.Store(id, pe)
					go doPolling(pe)
				}
			}
		}
	}
}

func checkPolling() {
	now := time.Now().UnixNano()
	list := []*datastore.PollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.Level != "off" && p.NextTime <= now {
			if _, busy := busyPollings.Load(p.ID); !busy {
				list = append(list, p)
			}
		}
		return true
	})
	if len(list) < 1 {
		return
	}
	log.Printf("check polling len=%d NumGoroutine=%d", len(list), runtime.NumGoroutine())
	sort.Slice(list, func(i, j int) bool {
		return list[i].NextTime < list[j].NextTime
	})
	for i := 0; i < len(list) && i < maxPolling; i++ {
		doPollingCh <- list[i].ID
	}
}

func doPolling(pe *datastore.PollingEnt) {
	defer func() {
		busyPollings.Delete(pe.ID)
		pe.NextTime = time.Now().UnixNano() + (int64(pe.PollInt) * 1000 * 1000 * 1000)
	}()
	oldState := pe.State
	switch pe.Type {
	case "ping":
		doPollingPing(pe)
	case "snmp":
		doPollingSnmp(pe)
	case "tcp":
		doPollingTCP(pe)
	case "http":
		doPollingHTTP(pe)
	case "tls":
		doPollingTLS(pe)
	case "dns":
		doPollingDNS(pe)
	case "ntp":
		doPollingNTP(pe)
	case "syslog":
		doPollingSyslog(pe)
	case "trap":
		doPollingTrap(pe)
	case "arplog":
		doPollingArpLog(pe)
	case "cmd":
		doPollingCmd(pe)
	case "ssh":
		doPollingSSH(pe)
	case "twsnmp":
		doPollingTWSNMP(pe)
	case "lxi":
		doPollingLxi(pe)
	}
	datastore.UpdatePolling(pe, false)
	if pe.LogMode == datastore.LogModeAlways || pe.LogMode == datastore.LogModeAI || (pe.LogMode == datastore.LogModeOnChange && oldState != pe.State) {
		if err := datastore.AddPollingLog(pe); err != nil {
			log.Printf("add polling log err=%v %#v", err, pe)
		}
	}
}

func setPollingState(pe *datastore.PollingEnt, newState string) {
	sendEvent := false
	oldState := pe.State
	delete(pe.Result, "error")
	if v, ok := pe.Result["_level"]; ok {
		if l, ok := v.(string); ok {
			log.Printf("setPollingState set level from JavaScript %s to %s", newState, l)
			newState = l
		}
		delete(pe.Result, "_level")
	}
	switch newState {
	case "normal":
		if pe.State != "normal" && pe.State != "repair" {
			if pe.State == "unknown" ||
				pe.Type == "syslog" || pe.Type == "trap" ||
				pe.Type == "arplog" {
				pe.State = "normal"
			} else {
				pe.State = "repair"
			}
			sendEvent = true
		}
	case "unknown":
		if pe.State != "unknown" {
			pe.State = "unknown"
			sendEvent = true
		}
	default:
		if pe.State != newState {
			pe.State = newState
			sendEvent = true
		}
	}
	if sendEvent {
		nodeName := "unknown"
		if n := datastore.GetNode(pe.NodeID); n != nil {
			nodeName = n.Name
		}
		datastore.SetNodeStateChanged(pe.NodeID)
		l := &datastore.EventLogEnt{
			Type:      "polling",
			Level:     pe.State,
			NodeID:    pe.NodeID,
			NodeName:  nodeName,
			LastLevel: oldState,
			Event:     fmt.Sprintf(i18n.Trans("Change polling state:%s(%s)"), pe.Name, pe.Type),
		}
		datastore.AddEventLog(l)
	}
}

func setPollingError(s string, pe *datastore.PollingEnt, err error) {
	pe.Result["error"] = fmt.Sprintf("%v", err)
	setPollingState(pe, "unknown")
}

func hasSameNamePolling(nodeID, name string) bool {
	r := false
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == nodeID && p.Name == name {
			r = true
			return false
		}
		return true
	})
	return r
}

func addJavaScriptFunctions(pe *datastore.PollingEnt, vm *otto.Otto) {
	vm.Set("setResult", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			n := call.Argument(0).String()
			if call.Argument(1).IsNumber() {
				if v, err := call.Argument(1).ToFloat(); err == nil {
					pe.Result[n] = v
				}
			} else if call.Argument(1).IsString() {
				pe.Result[n] = call.Argument(1).String()
			}
		}
		return otto.Value{}
	})
	vm.Set("getResult", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			k := call.Argument(0).String()
			if v, ok := pe.Result[k]; ok {
				if ov, err := otto.ToValue(v); err == nil {
					return ov
				}
			}
		}
		return otto.UndefinedValue()
	})
	vm.Set("setLevel", func(call otto.FunctionCall) otto.Value {
		if call.Argument(0).IsString() {
			level := call.Argument(0).String()
			pe.Result["_level"] = level
		}
		return otto.Value{}
	})
}
