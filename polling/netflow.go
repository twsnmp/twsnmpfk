package polling

// LOG監視ポーリング処理

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func doPollingNetFlow(pe *datastore.PollingEnt) {
	switch pe.Mode {
	case "stats":
		doPollingNetflowStats(pe)
	case "traffic":
		doPollingNetflowTraffic(pe)
	default:
		doPollingNetflowCount(pe)
	}
}

func doPollingNetflowTraffic(pe *datastore.PollingEnt) {
	var err error
	var filterSrc *regexp.Regexp
	var filterIP *regexp.Regexp
	var filterDst *regexp.Regexp
	var filterProtocol *regexp.Regexp
	var filterPort int
	if pe.Filter != "" {
		fs := strings.Split(pe.Filter, ",")
		for _, fe := range fs {
			f := strings.Split(fe, "=")
			if len(f) != 2 {
				continue
			}
			switch f[0] {
			case "src":
				filterSrc = makeRegexpIPFilter(f[1])
			case "dst":
				filterDst = makeRegexpIPFilter(f[1])
			case "ip":
				filterIP = makeRegexpIPFilter(f[1])
			case "port":
				filterPort, _ = strconv.Atoi(f[1])
			case "prot":
				filterProtocol = makeRegexpFilter(f[1])
			default:
				setPollingError("log", pe, fmt.Errorf("invalid filter format"))
				return
			}
		}
	}
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	var totalBytes float64
	var totalPackets float64
	var totalDur float64
	datastore.ForEachNetFlow(st, et, func(l *datastore.NetFlowEnt) bool {
		// Filter
		if filterPort > 0 {
			if filterPort != int(l.SrcPort) && filterPort != int(l.DstPort) {
				return true
			}
		}
		if filterIP != nil {
			if !filterIP.Match([]byte(l.SrcAddr)) && !filterIP.Match([]byte(l.DstAddr)) {
				return true
			}
		}
		if filterSrc != nil {
			if !filterIP.Match([]byte(l.SrcAddr)) {
				return true
			}
		}
		if filterDst != nil {
			if !filterIP.Match([]byte(l.DstAddr)) {
				return true
			}
		}
		if filterProtocol != nil {
			if !filterProtocol.Match([]byte(l.Protocol)) {
				return true
			}
		}
		totalBytes += float64(l.Bytes)
		totalPackets += float64(l.Packets)
		totalDur += float64(l.Dur)
		return true
	})
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	pe.Result["lastTime"] = float64(et)
	pe.Result["bytes"] = totalBytes
	pe.Result["packets"] = totalPackets
	pe.Result["duration"] = totalDur
	if totalDur > 0 {
		pe.Result["bps"] = totalBytes / totalDur
		pe.Result["pps"] = totalPackets / totalDur
	} else {
		pe.Result["bps"] = float64(0)
		pe.Result["pps"] = float64(0)
	}
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	value, err := vm.Run(pe.Script)
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

func doPollingNetflowCount(pe *datastore.PollingEnt) {
	var err error
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	var dstMap = make(map[string]bool)
	var srcMap = make(map[string]bool)
	var flowMap = make(map[string]bool)

	var totalBytes float64
	var totalPackets float64
	var totalDur float64
	var srcCount float64
	var dstCount float64
	var flowCount float64
	var icmpPackets float64
	var icmpBytes float64
	var otherProtPackets float64
	var otherProtBytes float64
	var otherTCPPackets float64
	var otherTCPBytes float64
	var otherUDPPackets float64
	var otherUDPBytes float64
	var httpPackets float64
	var httpBytes float64
	var httpsPackets float64
	var httpsBytes float64
	var dnsPackets float64
	var dnsBytes float64
	var mailPackets float64
	var mailBytes float64
	var sshPackets float64
	var sshBytes float64
	var filePackets float64
	var fileBytes float64
	var rdpPackets float64
	var rdpBytes float64
	var dhcpPackets float64
	var dhcpBytes float64
	var ntpPackets float64
	var ntpBytes float64
	var snmpPackets float64
	var snmpBytes float64
	var count float64

	datastore.ForEachNetFlow(st, et, func(l *datastore.NetFlowEnt) bool {
		count++
		k := l.SrcAddr + l.DstAddr
		if _, ok := flowMap[k]; !ok {
			flowCount++
			flowMap[k] = true
		}
		if _, ok := srcMap[l.SrcAddr]; !ok {
			srcCount++
			srcMap[l.SrcAddr] = true
		}
		if _, ok := dstMap[l.DstAddr]; !ok {
			dstCount++
			dstMap[l.DstAddr] = true
		}
		switch l.Protocol {
		case "tcp":
			if portCheck(l, []int{80}) {
				httpBytes += float64(l.Bytes)
				httpPackets += float64(l.Packets)
			} else if portCheck(l, []int{443}) {
				httpsBytes += float64(l.Bytes)
				httpsPackets += float64(l.Packets)
			} else if portCheck(l, []int{22}) {
				sshBytes += float64(l.Bytes)
				sshPackets += float64(l.Packets)
			} else if portCheck(l, []int{25, 587, 143, 110, 995, 993, 465}) {
				mailBytes += float64(l.Bytes)
				mailPackets += float64(l.Packets)
			} else if portCheck(l, []int{53}) {
				dnsBytes += float64(l.Bytes)
				dnsPackets += float64(l.Packets)
			} else if portCheck(l, []int{3389}) {
				rdpBytes += float64(l.Bytes)
				rdpPackets += float64(l.Packets)
			} else if portCheck(l, []int{137, 139, 445, 2049}) {
				fileBytes += float64(l.Bytes)
				filePackets += float64(l.Packets)
			} else {
				otherTCPBytes += float64(l.Bytes)
				otherTCPPackets += float64(l.Packets)
			}
		case "udp":
			if portCheck(l, []int{53}) {
				dnsBytes += float64(l.Bytes)
				dnsPackets += float64(l.Packets)
			} else if portCheck(l, []int{67, 68}) {
				dhcpBytes += float64(l.Bytes)
				dhcpPackets += float64(l.Packets)
			} else if portCheck(l, []int{123}) {
				ntpBytes += float64(l.Bytes)
				ntpPackets += float64(l.Packets)
			} else if portCheck(l, []int{161, 162}) {
				snmpBytes += float64(l.Bytes)
				snmpPackets += float64(l.Packets)
			} else {
				otherUDPBytes += float64(l.Bytes)
				otherUDPPackets += float64(l.Packets)
			}
		case "icmp", "icmpv6":
			icmpPackets += float64(l.Packets)
			icmpBytes += float64(l.Bytes)
		default:
			otherProtPackets += float64(l.Packets)
			otherProtBytes += float64(l.Packets)
		}
		totalBytes += float64(l.Bytes)
		totalPackets += float64(l.Packets)
		totalDur += float64(l.Dur)
		return true
	})
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	pe.Result["lastTime"] = float64(et)
	pe.Result["bytes"] = totalBytes
	pe.Result["count"] = count
	pe.Result["packets"] = totalPackets
	pe.Result["duration"] = totalDur
	if totalDur > 0 {
		pe.Result["bps"] = totalBytes / totalDur
		pe.Result["pps"] = totalPackets / totalDur
	} else {
		pe.Result["bps"] = float64(0)
		pe.Result["pps"] = float64(0)
	}
	pe.Result["srcCount"] = srcCount
	pe.Result["dstCount"] = dstCount
	pe.Result["flowCount"] = flowCount
	pe.Result["icmpPackets"] = icmpPackets
	pe.Result["icmpBytes"] = icmpBytes
	pe.Result["otherProtPackets"] = otherProtPackets
	pe.Result["otherProtBytes"] = otherProtBytes
	pe.Result["otherTCPPackets"] = otherTCPPackets
	pe.Result["otherTCPBytes"] = otherTCPBytes
	pe.Result["otherUDPPackets"] = otherUDPPackets
	pe.Result["otherUDPBytes"] = otherUDPBytes
	pe.Result["httpPackets"] = httpPackets
	pe.Result["httpBytes"] = httpBytes
	pe.Result["httpsPackets"] = httpsPackets
	pe.Result["httpsBytes"] = httpsBytes
	pe.Result["dnsPackets"] = dnsPackets
	pe.Result["dnsBytes"] = dnsBytes
	pe.Result["mailPackets"] = mailPackets
	pe.Result["mailBytes"] = mailBytes
	pe.Result["sshPackets"] = sshPackets
	pe.Result["sshBytes"] = sshBytes
	pe.Result["filePackets"] = filePackets
	pe.Result["fileBytes"] = fileBytes
	pe.Result["rdpPackets"] = rdpPackets
	pe.Result["rdpBytes"] = rdpBytes
	pe.Result["dhcpPackets"] = dhcpPackets
	pe.Result["dhcpBytes"] = dhcpBytes
	pe.Result["ntpPackets"] = ntpPackets
	pe.Result["ntpBytes"] = ntpBytes
	pe.Result["snmpPackets"] = snmpPackets
	pe.Result["snmpBytes"] = snmpBytes
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	value, err := vm.Run(pe.Script)
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

func portCheck(l *datastore.NetFlowEnt, ports []int) bool {
	for _, p := range ports {
		if p == l.SrcPort {
			return true
		}
		if p == l.DstPort {
			return true
		}
	}
	return false
}

func makeRegexpIPFilter(f string) *regexp.Regexp {
	if ip := net.ParseIP(f); ip != nil {
		f = regexp.QuoteMeta(f)
	}
	reg, err := regexp.Compile(f)
	if err != nil {
		return nil
	}
	return reg
}

func doPollingNetflowStats(pe *datastore.PollingEnt) {
	st := time.Now().Add(-time.Second * time.Duration(pe.PollInt)).UnixNano()
	if v, ok := pe.Result["lastTime"]; ok {
		if vf, ok := v.(float64); ok {
			st = int64(vf)
		}
	}
	et := time.Now().UnixNano()
	count := 0
	totalPacktes := float64(0)
	totalBytes := float64(0)
	fumbles := float64(0)
	ipMap := make(map[string]int)
	macMap := make(map[string]int)
	flowMap := make(map[string]int)
	protMap := make(map[string]int)
	fumbleSrcMap := make(map[string]int)
	fumbleFlowMap := make(map[string]int)
	datastore.ForEachNetFlow(st, et, func(l *datastore.NetFlowEnt) bool {
		sa := l.SrcAddr
		da := l.DstAddr
		sp := l.SrcPort
		dp := l.DstPort
		pi := 0
		switch l.Protocol {
		case "tcp":
			pi = 6
		case "udp":
			pi = 17
		case "icmp", "icmpv6":
			pi = 1
		default:
			return true
		}
		var flowKey string
		var prot string
		var ok bool
		if sa > da {
			flowKey = da + ":" + sa
		} else {
			flowKey = sa + ":" + da
		}
		if prot, ok = datastore.GetServiceName(pi, int(sp)); !ok {
			prot, _ = datastore.GetServiceName(pi, int(dp))
		}
		protMap[prot]++
		ipMap[sa]++
		flowMap[flowKey]++
		// TCP short
		if pi == 6 && l.Packets < 4 {
			fumbleFlowMap[flowKey]++
			fumbleSrcMap[sa]++
			fumbles++
		}
		// ICMP error
		if pi == 1 {
			switch sp {
			case 3, 4, 5, 11, 12:
				fumbleFlowMap[flowKey]++
				fumbleSrcMap[da]++
				fumbles++
			}
		}
		count++
		totalBytes += float64(l.Bytes)
		totalPacktes += float64(l.Packets)
		return true
	})
	pe.Result["lastTime"] = et
	pe.Result["count"] = float64(count)
	pe.Result["bytes"] = totalBytes
	pe.Result["packets"] = totalPacktes
	pe.Result["IPs"] = len(ipMap)
	pe.Result["MACs"] = len(macMap)
	pe.Result["flows"] = len(flowMap)
	pe.Result["fumbleFlows"] = len(fumbleFlowMap)
	pe.Result["fumbleSrc"] = len(fumbleSrcMap)
	pe.Result["fumbles"] = fumbles
	if pe.Script == "" {
		setPollingState(pe, "normal")
		return
	}
	vm := otto.New()
	setVMFuncAndValues(pe, vm)
	vm.Set("interval", pe.PollInt)
	for k, v := range pe.Result {
		vm.Set(k, v)
	}
	value, err := vm.Run(pe.Script)
	if err != nil {
		setPollingError("netflow", pe, err)
		return
	}
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(pe, "normal")
	} else {
		setPollingState(pe, pe.Level)
	}
}

func makeRegexpFilter(f string) *regexp.Regexp {
	reg, err := regexp.Compile(f)
	if err != nil {
		return nil
	}
	return reg
}
