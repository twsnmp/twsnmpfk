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
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/notify"
	"github.com/twsnmp/twsnmpfk/wol"
)

const maxPolling = 300

var (
	doPollingCh  chan string
	busyPollings sync.Map
)
var stopPolling = false
var checkingPolling = false

func Start(ctx context.Context, wg *sync.WaitGroup) error {
	doPollingCh = make(chan string, maxPolling)
	wg.Add(1)
	go pollingBackend(ctx, wg)
	return nil
}

func GetAutoPollings(n *datastore.NodeEnt, pt *datastore.PollingTemplateEnt) []*datastore.PollingEnt {
	switch pt.Type {
	case "snmp":
		if pt.AutoParam == "ifIndex" {
			return getAutoSnmpIFPollings(n, pt)
		}
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
			if pe.Type == "gnmi" && pe.Mode == "subscribe" {
				GNMIStopSubscription(pe.ID)
				time.Sleep(time.Millisecond * 20)
			}
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
			if pe.Type == "gnmi" && pe.Mode == "subscribe" {
				GNMIStopSubscription(pe.ID)
				time.Sleep(time.Millisecond * 20)
			}
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
			gNMIStopAllSubscription()
			stopPolling = true
			log.Println("stop polling")
			return
		case <-timer.C:
			if !checkingPolling {
				go checkPolling()
			} else {
				log.Println("skip polling check")
			}
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
	checkingPolling = true
	defer func() {
		checkingPolling = false
	}()
	now := time.Now().UnixNano()
	list := []*datastore.PollingEnt{}
	total := 0
	st := time.Now()
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.Level != "off" && p.NextTime <= now {
			if _, busy := busyPollings.Load(p.ID); !busy {
				list = append(list, p)
			}
			total++
		}
		return true
	})
	if len(list) < 1 {
		return
	}
	log.Printf("check polling len=%d total=%d NumGoroutine=%d dur=%v", len(list), total, runtime.NumGoroutine(), time.Since(st))
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
	case "netflow":
		doPollingNetFlow(pe)
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
	case "gnmi":
		doPollingGNMI(pe)
		if pe.Mode == "subscribe" {
			return
		}
	case "twlogeye":
		doPollingTwLogEye(pe)
	case "pihole":
		docPollingPiHole(pe)
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
	if v, ok := pe.Result["_level"]; ok {
		if l, ok := v.(string); ok {
			log.Printf("setPollingState set level from JavaScript %s to %s", newState, l)
			newState = l
		}
		delete(pe.Result, "_level")
	}
	switch newState {
	case "normal":
		delete(pe.Result, "error")
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
		go doAction(pe)
	}
}

func doAction(pe *datastore.PollingEnt) {
	if pe.State == "unknown" {
		return
	}
	action := pe.FailAction
	if pe.State == "repair" || pe.State == "normal" {
		action = pe.RepairAction
	}
	if action == "" {
		return
	}
	for _, a := range strings.Split(action, "\n") {
		a = strings.TrimSpace(a)
		al := strings.Split(a, " ")
		if !doOneAction(al) {
			// アクションをwaitの条件で途中で終了できる
			break
		}
	}
}

func doOneAction(alin []string) bool {
	al := []string{}
	for _, a := range alin {
		if a != "" {
			al = append(al, a)
		}
	}
	if len(al) < 2 {
		return true
	}
	log.Printf("doOneAction %v", al)
	switch al[0] {
	case "wol":
		{
			mac := al[1]
			if n := datastore.FindNodeFromName(al[1]); n != nil {
				mac = n.MAC
			} else if n := datastore.FindNodeFromIP(al[1]); n != nil {
				mac = n.MAC
			}
			if strings.Contains(mac, ":") {
				wol.SendWakeOnLanPacket(mac)
			}
		}
	case "mail":
		{
			subject := al[1]
			body := subject
			if len(al) > 2 {
				body = al[2]
			}
			if subject != "" {
				notify.SendMail(subject, body)
			}
		}
	case "wait":
		{
			if to, err := strconv.Atoi(al[1]); err == nil {
				node := ""
				state := ""
				if len(al) > 3 {
					node = al[2]
					state = al[3]
				}
				if !doWait(to, node, state) {
					return false
				}
			}
		}
	case "cmd":
		doActionCmd(al[1:])
	}
	return true
}

func doWait(to int, node, state string) bool {
	for i := 0; i < to && !stopPolling; i++ {
		if node != "" {
			if n := datastore.FindNodeFromName(node); n != nil {
				if state == "up" && (n.State == "normal" || n.State == "repair") {
					return false
				} else if state == "down" && (n.State == "low" || n.State == "high" || n.State == "warn") {
					return false
				}
			}
		}
		time.Sleep(time.Second)
	}
	return true
}

func doActionCmd(cl []string) {
	tio := &timeout.Timeout{
		Duration:  30 * time.Second,
		KillAfter: 5 * time.Second,
	}
	if filepath.Ext(cl[0]) == ".sh" {
		cl[0] = filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
		tio.Cmd = exec.Command("/bin/sh", "-c", strings.Join(cl, " "))
	} else {
		exe := filepath.Join(datastore.GetDataStorePath(), "cmd", filepath.Base(cl[0]))
		if len(cl) == 1 {
			tio.Cmd = exec.Command(exe)
		} else {
			tio.Cmd = exec.Command(exe, cl[1:]...)
		}
	}
	if _, _, _, err := tio.Run(); err != nil {
		log.Printf("doActionCmd err=%v", err)
	}
}

func setPollingError(s string, pe *datastore.PollingEnt, err error) {
	pe.Result["error"] = fmt.Sprintf("%s err=%v", s, err)
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

func setVMFuncAndValues(pe *datastore.PollingEnt, vm *otto.Otto) {
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
	if len(pe.Result) > 0 {
		for k, v := range pe.Result {
			if k != "error" {
				vm.Set(k+"_last", v)
			}
		}
	}
	vm.Set("iterval", pe.PollInt)
}
