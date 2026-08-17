package polling

import (
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/stun"
)

func doPollingSTUN(pe *datastore.PollingEnt) {
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		setPollingError("stun", pe, fmt.Errorf("node not found"))
		return
	}

	server := pe.Params
	if server == "" {
		server = stun.DefaultServer
	}

	network := "udp4"
	if pe.Mode == "ipv6" || pe.Mode == "udp6" {
		network = "udp6"
	}

	timeoutSec := pe.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	timeoutDur := time.Duration(timeoutSec) * time.Second

	var res *stun.Result
	var err error
	ok := false

	for i := 0; !ok && i <= pe.Retry; i++ {
		res, err = stun.Query(server, network, timeoutDur)
		if err != nil {
			pe.Result["error"] = fmt.Sprintf("%v", err)
			continue
		}
		ok = true
	}

	if !ok || res == nil {
		pe.Result["rtt"] = 0.0
		pe.Result["ip"] = ""
		setPollingState(pe, pe.Level)
		return
	}

	oldIP := ""
	if v, ok := pe.Result["ip"]; ok {
		if s, ok := v.(string); ok {
			oldIP = s
		}
	}

	pe.Result["ip"] = res.IP
	pe.Result["port"] = float64(res.Port)
	pe.Result["host"] = res.Hostname
	pe.Result["local"] = fmt.Sprintf("%s:%d", res.LocalIP, res.LocalPort)
	pe.Result["rtt"] = float64(res.RTTNano)
	delete(pe.Result, "error")

	// IP address change event log
	if oldIP != "" && oldIP != res.IP {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "polling",
			Level:    pe.Level,
			NodeID:   pe.NodeID,
			NodeName: n.Name,
			Event:    fmt.Sprintf("%s: %s -> %s (%s)", i18n.Trans("STUN IP changed"), oldIP, res.IP, pe.Name),
		})
	}

	script := pe.Script
	if script == "" {
		if oldIP != "" && oldIP != res.IP {
			setPollingState(pe, pe.Level)
			return
		}
		setPollingState(pe, "normal")
		return
	}

	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	vm.Set("ip", res.IP)
	vm.Set("oldip", oldIP)
	vm.Set("port", res.Port)
	vm.Set("host", res.Hostname)
	vm.Set("local", res.LocalIP)
	vm.Set("rtt", res.RTTNano)

	value, err := vm.Run(script)
	if err != nil {
		setPollingError("stun", pe, fmt.Errorf("%v", err))
		return
	}
	if pass, _ := value.ToBoolean(); pass {
		setPollingState(pe, "normal")
		return
	}
	setPollingState(pe, pe.Level)
}
