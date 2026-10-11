package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strings"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/logger"
)

type EventLogFilterEnt struct {
	NodeID    string `json:"NodeID"`
	Start     string `json:"Start"`
	End       string `json:"End"`
	EventType string `json:"EventType"`
	NodeName  string `json:"NodeName"`
	Event     string `json:"Event"`
	Level     int    `json:"Level"`
}

// GetEventLogs retunrs  event logs
func (a *App) GetEventLogs(filter EventLogFilterEnt) []*datastore.EventLogEnt {
	ret := []*datastore.EventLogEnt{}
	typeFilter := makeStringFilter(filter.EventType)
	nodeFilter := makeStringFilter(filter.NodeName)
	eventFilter := makeStringFilter(filter.Event)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if filter.NodeID != "" && filter.NodeID != l.NodeID {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.Type) {
			return true
		}
		if nodeFilter != nil && !nodeFilter.MatchString(l.NodeName) {
			return true
		}
		if eventFilter != nil && !eventFilter.MatchString(l.Event) {
			return true
		}
		if filter.Level != 0 && filter.Level > getLevelNum(l.Level) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

func getLevelNum(l string) int {
	switch l {
	case "high":
		return 3
	case "low":
		return 2
	case "warn":
		return 1
	}
	return 0
}

// GetMapEventLogs retunrs  event logs
func (a *App) GetMapEventLogs() []*datastore.EventLogEnt {
	ret := []*datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(func(l *datastore.EventLogEnt) bool {
		if l.Type == "user" || l.Type == "oprate" || l.Type == "arpwatch" {
			return true
		}
		ret = append(ret, l)
		return len(ret) < 100
	})
	return ret
}

type SyslogFilterEnt struct {
	Start    string `json:"Start"`
	End      string `json:"End"`
	Host     string `json:"Host"`
	Tag      string `json:"Tag"`
	Message  string `json:"Message"`
	Severity int    `json:"Severity"`
}

// GetSyslogs retunrs syslogs
func (a *App) GetSyslogs(filter SyslogFilterEnt) []*datastore.SyslogEnt {
	ret := []*datastore.SyslogEnt{}
	hostFilter := makeIPFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	msgFilter := makeStringFilter(filter.Message)
	st := makeTimeFilter(filter.Start, 1)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
		if filter.Severity < l.Severity {
			return true
		}
		if hostFilter != nil && !hostFilter.MatchString(l.Host) {
			return true
		}
		if tagFilter != nil && !tagFilter.MatchString(l.Tag) {
			return true
		}
		if msgFilter != nil && !msgFilter.MatchString(l.Message) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

// CalculateSyslogAnomaly calculates anomaly score for syslogs
func (a *App) CalculateSyslogAnomaly(logs []*datastore.SyslogEnt, algo, vmode string) *backend.SyslogAnomalyResponse {
	return backend.CalculateSyslogAnomaly(logs, algo, vmode)
}

type TrapFilterEnt struct {
	Start string `json:"Start"`
	End   string `json:"End"`
	From  string `json:"From"`
	Type  string `json:"Type"`
}

// GetTraps retunrs SNMP Trap log
func (a *App) GetTraps(filter TrapFilterEnt) []*datastore.TrapEnt {
	ret := []*datastore.TrapEnt{}
	fromFilter := makeIPFilter(filter.From)
	typeFilter := makeStringFilter(filter.Type)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachTraps(st, et, func(l *datastore.TrapEnt) bool {
		if fromFilter != nil && !fromFilter.MatchString(l.FromAddress) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.TrapType) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

type NetFlowFilterEnt struct {
	Start    string `json:"Start"`
	End      string `json:"End"`
	Single   bool   `json:"Single"`
	SrcAddr  string `json:"SrcAddr"`
	SrcPort  int    `json:"SrcPort"`
	SrcLoc   string `json:"SrcLoc"`
	SrcMAC   string `json:"SrcMAC"`
	DstAddr  string `json:"DstAddr"`
	DstPort  int    `json:"DstPort"`
	DstLoc   string `json:"DstLoc"`
	DstMAC   string `json:"DstMAC"`
	Protocol string `json:"Protocol"`
	TCPFlags string `json:"TCPFlags"`
}

// GetNetFlow returns NetFlow logs.
func (a *App) GetNetFlow(filter NetFlowFilterEnt) []*datastore.NetFlowEnt {
	ret := []*datastore.NetFlowEnt{}
	srcFilter := makeIPFilter(filter.SrcAddr)
	srcLocFilter := makeStringFilter(filter.SrcLoc)
	srcMACFilter := makeStringFilter(filter.SrcMAC)
	dstFilter := makeIPFilter(filter.DstAddr)
	dstLocFilter := makeStringFilter(filter.DstLoc)
	dstMACFilter := makeStringFilter(filter.DstMAC)
	tcpFlagsFilter := makeStringFilter(filter.TCPFlags)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.Start, 1)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachNetFlow(st, et, func(l *datastore.NetFlowEnt) bool {
		if filter.Single {
			if srcFilter != nil && (!srcFilter.MatchString(l.SrcAddr) && !srcFilter.MatchString(l.DstAddr)) {
				return true
			}
			if srcLocFilter != nil && (!srcLocFilter.MatchString(l.SrcLoc) && !srcLocFilter.MatchString(l.DstLoc)) {
				return true
			}
			if srcMACFilter != nil && (!srcMACFilter.MatchString(l.SrcMAC) && !srcMACFilter.MatchString(l.DstMAC)) {
				return true
			}
			if filter.SrcPort > 0 && (filter.SrcPort != l.SrcPort && filter.SrcPort != l.DstPort) {
				return true
			}
		} else {
			if srcFilter != nil && !srcFilter.MatchString(l.SrcAddr) {
				return true
			}
			if srcLocFilter != nil && !srcLocFilter.MatchString(l.SrcLoc) {
				return true
			}
			if srcMACFilter != nil && !srcLocFilter.MatchString(l.SrcMAC) {
				return true
			}
			if dstFilter != nil && !dstFilter.MatchString(l.DstAddr) {
				return true
			}
			if dstLocFilter != nil && !dstLocFilter.MatchString(l.DstLoc) {
				return true
			}
			if dstMACFilter != nil && !dstMACFilter.MatchString(l.DstMAC) {
				return true
			}
			if filter.SrcPort > 0 && filter.SrcPort != l.SrcPort {
				return true
			}
			if filter.DstPort > 0 && filter.DstPort != l.DstPort {
				return true
			}
		}
		if tcpFlagsFilter != nil && !tcpFlagsFilter.MatchString(l.TCPFlags) {
			return true
		}
		if protocolFilter != nil && !protocolFilter.MatchString(l.Protocol) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

type SFlowFilterEnt struct {
	Start    string `json:"Start"`
	End      string `json:"End"`
	Single   bool   `json:"Single"`
	SrcAddr  string `json:"SrcAddr"`
	SrcPort  int    `json:"SrcPort"`
	SrcLoc   string `json:"SrcLoc"`
	SrcMAC   string `json:"SrcMAC"`
	DstAddr  string `json:"DstAddr"`
	DstPort  int    `json:"DstPort"`
	DstLoc   string `json:"DstLoc"`
	DstMAC   string `json:"DstMAC"`
	Protocol string `json:"Protocol"`
	TCPFlags string `json:"TCPFlags"`
	Reason   int    `json:"Reason"`
}

// GetSFlow は sFlowログを返します
func (a *App) GetSFlow(filter SFlowFilterEnt) []*datastore.SFlowEnt {
	ret := []*datastore.SFlowEnt{}
	srcFilter := makeIPFilter(filter.SrcAddr)
	srcLocFilter := makeStringFilter(filter.SrcLoc)
	srcMACFilter := makeStringFilter(filter.SrcMAC)
	dstFilter := makeIPFilter(filter.DstAddr)
	dstLocFilter := makeStringFilter(filter.DstLoc)
	dstMACFilter := makeStringFilter(filter.DstMAC)
	tcpFlagsFilter := makeStringFilter(filter.TCPFlags)
	protocolFilter := makeStringFilter(filter.Protocol)
	st := makeTimeFilter(filter.Start, 6)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachSFlow(st, et, func(l *datastore.SFlowEnt) bool {
		if filter.Single {
			if srcFilter != nil && (!srcFilter.MatchString(l.SrcAddr) && !srcFilter.MatchString(l.DstAddr)) {
				return true
			}
			if srcLocFilter != nil && (!srcLocFilter.MatchString(l.SrcLoc) && !srcLocFilter.MatchString(l.DstLoc)) {
				return true
			}
			if srcMACFilter != nil && (!srcMACFilter.MatchString(l.SrcMAC) && !srcMACFilter.MatchString(l.DstMAC)) {
				return true
			}
			if filter.SrcPort > 0 && (filter.SrcPort != l.SrcPort && filter.SrcPort != l.DstPort) {
				return true
			}
		} else {
			if srcFilter != nil && !srcFilter.MatchString(l.SrcAddr) {
				return true
			}
			if srcLocFilter != nil && !srcLocFilter.MatchString(l.SrcLoc) {
				return true
			}
			if srcMACFilter != nil && !srcLocFilter.MatchString(l.SrcMAC) {
				return true
			}
			if dstFilter != nil && !dstFilter.MatchString(l.DstAddr) {
				return true
			}
			if dstLocFilter != nil && !dstLocFilter.MatchString(l.DstLoc) {
				return true
			}
			if dstMACFilter != nil && !dstMACFilter.MatchString(l.DstMAC) {
				return true
			}
			if filter.SrcPort > 0 && filter.SrcPort != l.SrcPort {
				return true
			}
			if filter.DstPort > 0 && filter.DstPort != l.DstPort {
				return true
			}
		}
		if tcpFlagsFilter != nil && !tcpFlagsFilter.MatchString(l.TCPFlags) {
			return true
		}
		if protocolFilter != nil && !protocolFilter.MatchString(l.Protocol) {
			return true
		}
		if filter.Reason > 0 && filter.Reason != l.Reason {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

type SFlowCounterFilterEnt struct {
	Start  string `json:"Start"`
	End    string `json:"End"`
	Type   string `json:"Type"`
	Remote string `json:"Remote"`
}

// GetSFlowCounter は sFlow Counterログを返します
func (a *App) GetSFlowCounter(filter SFlowCounterFilterEnt) []*datastore.SFlowCounterEnt {
	ret := []*datastore.SFlowCounterEnt{}
	remoteFilter := makeIPFilter(filter.Remote)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachSFlowCounter(st, et, func(l *datastore.SFlowCounterEnt) bool {
		if remoteFilter != nil && !remoteFilter.MatchString(l.Remote) {
			return true
		}
		if filter.Type != "" && filter.Type != l.Type {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

// GetArpTable returns the ARP Table.
func (a *App) GetArpTable() []*datastore.ArpEnt {
	ret := []*datastore.ArpEnt{}
	datastore.ForEachArp(func(l *datastore.ArpEnt) bool {
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

// IPAMSubnetBlock represents a /24 subnet block within a larger range.
type IPAMSubnetBlock struct {
	Subnet string  `json:"Subnet"`
	Size   int64   `json:"Size"`
	Used   int64   `json:"Used"`
	Usage  float64 `json:"Usage"`
}

// IPAMRangeEnt represents the IPAM utilization report for an address range.
type IPAMRangeEnt struct {
	Range   string             `json:"Range"`
	StartIP string             `json:"StartIP"`
	EndIP   string             `json:"EndIP"`
	Size    int64              `json:"Size"`
	Used    int64              `json:"Used"`
	Usage   float64            `json:"Usage"`
	UsedIP  []int              `json:"UsedIP"` // 100 slots for heatmap (0..99%)
	Subnets []*IPAMSubnetBlock `json:"Subnets,omitempty"`
}

// GetIPAM creates an IPAM report.
func (a *App) GetIPAM() []*IPAMRangeEnt {
	now := time.Now().Unix()
	timeoutSec := int64(datastore.MapConf.ArpTimeout * 3600)

	// Collect used IPs: uint32 -> name/MAC
	usedIPs := make(map[uint32]string)

	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if n.IP != "" {
			if ip := net.ParseIP(n.IP).To4(); ip != nil {
				usedIPs[ip2int(ip)] = n.Name
			}
		}
		return true
	})

	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		if timeoutSec > 0 && now-a.LastTime > timeoutSec {
			return true
		}
		if ip := net.ParseIP(a.IP).To4(); ip != nil {
			u := ip2int(ip)
			if _, exists := usedIPs[u]; !exists {
				usedIPs[u] = a.MAC
			}
		}
		return true
	})

	// Parse configured ranges
	rangeStrs := []string{}
	if datastore.MapConf.ArpWatchRange != "" {
		for _, r := range strings.Split(datastore.MapConf.ArpWatchRange, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				rangeStrs = append(rangeStrs, r)
			}
		}
	}

	// Fallback: If no range configured, infer /24 subnets from used IPs
	if len(rangeStrs) == 0 {
		subnetSet := make(map[string]bool)
		for u := range usedIPs {
			ip := int2ip(u)
			subnet := fmt.Sprintf("%d.%d.%d.0/24", ip[0], ip[1], ip[2])
			subnetSet[subnet] = true
		}
		for s := range subnetSet {
			rangeStrs = append(rangeStrs, s)
		}
		sort.Strings(rangeStrs)
		if len(rangeStrs) == 0 {
			rangeStrs = append(rangeStrs, "192.168.1.0/24")
		}
	}

	ret := []*IPAMRangeEnt{}

	for _, rStr := range rangeStrs {
		var sIP, eIP uint32
		var validStart, validEnd uint32
		var isCIDR bool

		parts := strings.SplitN(rStr, "-", 2)
		if len(parts) == 1 {
			// CIDR format
			ip, ipnet, err := net.ParseCIDR(rStr)
			if err != nil {
				continue
			}
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}
			isCIDR = true
			mask := binary.BigEndian.Uint32(ipnet.Mask)
			sIP = ip2int(ipnet.IP.To4())
			eIP = sIP | (^mask)

			ones, _ := ipnet.Mask.Size()
			if ones < 31 {
				validStart = sIP + 1
				validEnd = eIP - 1
			} else {
				validStart = sIP
				validEnd = eIP
			}
		} else {
			// Range format: A.B.C.D-E.F.G.H
			s := net.ParseIP(strings.TrimSpace(parts[0])).To4()
			e := net.ParseIP(strings.TrimSpace(parts[1])).To4()
			if s == nil || e == nil {
				continue
			}
			sIP = ip2int(s)
			eIP = ip2int(e)
			if sIP > eIP {
				sIP, eIP = eIP, sIP
			}
			validStart = sIP
			validEnd = eIP
		}

		if validStart > validEnd {
			continue
		}

		size := int64(validEnd - validStart + 1)
		ent := &IPAMRangeEnt{
			Range:   rStr,
			StartIP: int2ip(validStart).String(),
			EndIP:   int2ip(validEnd).String(),
			Size:    size,
			UsedIP:  make([]int, 100),
		}

		// Count used IPs in this range
		var usedCount int64
		blockMap := make(map[uint32]int64)

		for u := range usedIPs {
			if u >= validStart && u <= validEnd {
				usedCount++
				slot := 0
				if size > 1 {
					slot = int((int64(u-validStart) * 100) / size)
					if slot >= 100 {
						slot = 99
					}
					if slot < 0 {
						slot = 0
					}
				}
				ent.UsedIP[slot]++

				if size > 256 {
					blockNet := u & 0xffffff00
					blockMap[blockNet]++
				}
			}
		}

		ent.Used = usedCount
		if size > 0 {
			ent.Usage = (float64(usedCount) * 100.0) / float64(size)
		}

		// If large range, generate /24 subnet blocks
		if size > 256 {
			firstBlock := validStart & 0xffffff00
			lastBlock := validEnd & 0xffffff00
			totalBlocks := int64((lastBlock-firstBlock)/256 + 1)

			maxBlocks := int64(256)
			step := int64(1)
			if totalBlocks > maxBlocks {
				step = (totalBlocks + maxBlocks - 1) / maxBlocks
			}

			for b := int64(0); b < totalBlocks; b += step {
				blockNet := firstBlock + uint32(b*256)
				blockIP := int2ip(blockNet)
				bUsed := blockMap[blockNet]

				bStart := blockNet + 1
				bEnd := blockNet + 254
				if bStart < validStart {
					bStart = validStart
				}
				if bEnd > validEnd {
					bEnd = validEnd
				}
				bSize := int64(254)
				if bStart <= bEnd {
					bSize = int64(bEnd - bStart + 1)
				}

				bUsage := 0.0
				if bSize > 0 {
					bUsage = (float64(bUsed) * 100.0) / float64(bSize)
				}

				ent.Subnets = append(ent.Subnets, &IPAMSubnetBlock{
					Subnet: fmt.Sprintf("%d.%d.%d.0/24", blockIP[0], blockIP[1], blockIP[2]),
					Size:   bSize,
					Used:   bUsed,
					Usage:  bUsage,
				})
			}
		} else if isCIDR {
			ent.Subnets = append(ent.Subnets, &IPAMSubnetBlock{
				Subnet: rStr,
				Size:   size,
				Used:   usedCount,
				Usage:  ent.Usage,
			})
		}

		ret = append(ret, ent)
	}

	return ret
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nIP uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nIP)
	return ip
}

// ResetArpTable clears the ARP Table.
func (a *App) ResetArpTable() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm clear"),
		Message:       i18n.Trans("Do you want to clear?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	logger.ResetArpWatch = true
	if err := datastore.ResetArpTable(); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Clear all arp watch info"),
	})
	return true
}

func (a *App) DeleteArpEnt(ips []string) bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteArpEnt(ips); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Delete arp ent ip=%v"), ips),
	})
	return true
}

type ArpLogEnt struct {
	Time      int64  `json:"Time"`
	State     string `json:"State"`
	IP        string `json:"IP"`
	Node      string `json:"Node"`
	NewMAC    string `json:"NewMAC"`
	NewVendor string `json:"NewVendor"`
	OldMAC    string `json:"OldMAC"`
	OldVendor string `json:"OldVendor"`
}

// GetArpLogs returns the latest ARP logs.
func (a *App) GetArpLogs() []*ArpLogEnt {
	ret := []*ArpLogEnt{}
	nodeMap := make(map[string]string)
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		node := ""
		var ok bool
		if node, ok = nodeMap[l.IP]; !ok {
			if n := datastore.FindNodeFromIP(l.IP); n != nil {
				node = n.Name
				nodeMap[l.IP] = n.Name
			} else {
				nodeMap[l.IP] = ""
			}
		}
		ret = append(ret, &ArpLogEnt{
			Time:      l.Time,
			State:     l.State,
			IP:        l.IP,
			NewMAC:    l.NewMAC,
			OldMAC:    l.OldMAC,
			NewVendor: datastore.FindVendor(l.NewMAC),
			OldVendor: datastore.FindVendor(l.OldMAC),
			Node:      node,
		})
		return len(ret) < maxDispLog
	})
	return ret
}

// DeleteAllEventLogs deletes all event logs.
func (a *App) DeleteAllEventLogs() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("logs"); err != nil {
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all event logs"),
	})
	return true
}

// DeleteAllSyslog deletes all Syslog entries.
func (a *App) DeleteAllSyslog() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("syslog"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all syslog"),
	})
	return true
}

// DeleteAllTraps deletes all TRAP logs.
func (a *App) DeleteAllTraps() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("trap"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all TRAP logs"),
	})
	return true
}

// DeleteAllNetFlow deletes all NetFlow logs.
func (a *App) DeleteAllNetFlow() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("netflow"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all NetFlow logs"),
	})
	return true
}

// DeleteAllSFlow deletes all sFlow logs.
func (a *App) DeleteAllSFlow() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("sflow"); err != nil {
		log.Println(err)
		return false
	}
	if err := datastore.DeleteAllLogs("sflowCounter"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all sFlow and sFlow Counter logs"),
	})
	return true
}

// DeleteAllPollingLogs deletes all polling logs.
func (a *App) DeleteAllPollingLogs() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("pollingLogs"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all polling logs"),
	})
	return true
}

func makeTimeFilter(dt string, oh int) int64 {
	if dt == "" {
		return time.Now().Add(-time.Hour * time.Duration(oh)).UnixNano()
	}
	zone, _ := time.Now().Zone()
	var t time.Time
	var err error
	if t, err = time.Parse("2006-01-02T15:04 MST", dt+" "+zone); err != nil {
		log.Println(err)
		t = time.Now().Add(-time.Hour * time.Duration(oh))
	}
	return t.UnixNano()
}

func makeStringFilter(f string) *regexp.Regexp {
	if f == "" {
		return nil
	}
	r, err := regexp.Compile(f)
	if err != nil {
		return nil
	}
	return r
}

func makeIPFilter(f string) *regexp.Regexp {
	if f == "" {
		return nil
	}
	if ip := net.ParseIP(f); ip != nil {
		f = regexp.QuoteMeta(f)
	}
	r, err := regexp.Compile(f)
	if err != nil {
		return nil
	}
	return r
}
