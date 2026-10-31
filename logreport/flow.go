package logreport

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FlowRecord represents a normalized flow event from NetFlow, IPFIX, or sFlow.
type FlowRecord struct {
	Time     int64
	SrcIP    string
	SrcPort  int
	DstIP    string
	DstPort  int
	Protocol string
	Prot     int
	Packets  int64
	Bytes    int64
	Duration float64
	TCPFlags string
}

// ProcessFlow updates FlowEnt, ServerEnt, and FumbleEnt in the session.
func (s *Session) ProcessFlow(fr *FlowRecord) {
	if fr == nil || fr.SrcIP == "" || fr.DstIP == "" {
		return
	}
	if fr.Prot == 0 {
		fr.Prot = parseProtocolNumber(fr.Protocol)
	}

	// 1. Check Fumble (TCP drops, ICMP errors)
	if s.checkFumble(fr) {
		s.Processed++
		return
	}

	// 2. Determine Client, Server, Service direction
	server, client, service := getFlowDir(fr)
	if server == "" || client == "" {
		return
	}

	// 3. Update Server Report
	s.checkServerReport(server, service, fr.Bytes, fr.Packets, fr.Time)

	// 4. Update Flow Report
	id := fmt.Sprintf("%s:%s", client, server)
	now := time.Now().UnixNano()
	f := getEnt[FlowEnt](s, KindFlow, id)
	if f != nil {
		if f.Services == nil {
			f.Services = make(map[string]int64)
		}
		if _, ok := f.Services[service]; ok {
			f.Services[service]++
		} else {
			f.Services[service] = 1
			setFlowPenalty(f)
		}
		if f.ServerLoc == "" {
			f.ServerLoc = s.store.GetLoc(f.Server)
		}
		if f.ClientLoc == "" {
			f.ClientLoc = s.store.GetLoc(f.Client)
		}
		f.Bytes += fr.Bytes
		f.Packets += fr.Packets
		f.Count++
		if fr.Duration > f.Duration {
			f.Duration = fr.Duration
		}
		if fr.Time > f.LastTime {
			f.LastTime = fr.Time
		}
		f.UpdateTime = now
		putEnt(s, KindFlow, id, f)
		s.Processed++
		return
	}

	f = &FlowEnt{
		ID:         id,
		Client:     client,
		Server:     server,
		Services:   make(map[string]int64),
		Count:      1,
		Bytes:      fr.Bytes,
		Packets:    fr.Packets,
		Duration:   fr.Duration,
		ServerLoc:  s.store.GetLoc(server),
		ClientLoc:  s.store.GetLoc(client),
		FirstTime:  fr.Time,
		LastTime:   fr.Time,
		UpdateTime: now,
	}
	cNode := s.lookupNode(client)
	f.ClientName = cNode.Name
	f.ClientNodeID = cNode.ID
	sNode := s.lookupNode(server)
	f.ServerName = sNode.Name
	f.ServerNodeID = sNode.ID

	f.Services[service] = 1
	setFlowPenalty(f)
	putEnt(s, KindFlow, id, f)
	s.Processed++
}

var fumbleAlertSuppressed sync.Map

func (s *Session) checkFumble(fr *FlowRecord) bool {
	now := time.Now().UnixNano()
	fumbleThreshold := int64(10)

	recordAlert := func(id string, tcpCount, icmpCount int64) {
		if tcpCount >= fumbleThreshold || icmpCount >= fumbleThreshold {
			lastAlert, exists := fumbleAlertSuppressed.Load(id)
			if !exists || now-lastAlert.(int64) > int64(1*time.Hour) {
				fumbleAlertSuppressed.Store(id, now)
			}
		}
	}

	// TCP flows with packet count <= 2 are treated as fumble/aborted attempts
	if fr.Prot == 6 && fr.Packets > 0 && fr.Packets <= 2 {
		id := fr.DstIP + "_" + fr.SrcIP
		f := getEnt[FumbleEnt](s, KindFumble, id)
		if f != nil {
			f.TCPCount++
			f.LastTime = now
			putEnt(s, KindFumble, id, f)
			recordAlert(id, f.TCPCount, f.IcmpCount)
			return true
		}
		id = fr.SrcIP + "_" + fr.DstIP
		f = getEnt[FumbleEnt](s, KindFumble, id)
		if f != nil {
			f.TCPCount++
			f.LastTime = now
			putEnt(s, KindFumble, id, f)
			recordAlert(id, f.TCPCount, f.IcmpCount)
			return true
		}
		putEnt(s, KindFumble, id, &FumbleEnt{
			ID:        id,
			TCPCount:  1,
			FirstTime: now,
			LastTime:  now,
		})
		recordAlert(id, 1, 0)
		return true
	} else if fr.Prot == 1 {
		// ICMP unreachable / time exceeded
		icmpType := fr.DstPort / 256
		switch icmpType {
		case 3, 4, 5, 11, 12:
			id := fr.DstIP + "_" + fr.SrcIP
			f := getEnt[FumbleEnt](s, KindFumble, id)
			if f != nil {
				f.IcmpCount++
				f.LastTime = now
				putEnt(s, KindFumble, id, f)
				recordAlert(id, f.TCPCount, f.IcmpCount)
				return true
			}
			putEnt(s, KindFumble, id, &FumbleEnt{
				ID:        id,
				IcmpCount: 1,
				FirstTime: now,
				LastTime:  now,
			})
			recordAlert(id, 0, 1)
			return true
		}
	}
	return false
}

func (s *Session) checkServerReport(server, service string, bytes, packets, t int64) {
	if service == "" {
		return
	}
	now := time.Now().UnixNano()
	id := server
	se := getEnt[ServerEnt](s, KindServer, id)
	if se != nil {
		if se.Services == nil {
			se.Services = make(map[string]int64)
		}
		if _, ok := se.Services[service]; ok {
			se.Services[service]++
		} else {
			se.Services[service] = 1
			setServerPenalty(se)
		}
		se.Count++
		se.Bytes += bytes
		se.Packets += packets
		if t > se.LastTime {
			se.LastTime = t
		}
		se.UpdateTime = now
		sNode := s.lookupNode(server)
		se.ServerName = sNode.Name
		se.ServerNodeID = sNode.ID
		putEnt(s, KindServer, id, se)
		return
	}

	se = &ServerEnt{
		ID:         id,
		Server:     server,
		Services:   make(map[string]int64),
		Loc:        s.store.GetLoc(server),
		Count:      1,
		Bytes:      bytes,
		Packets:    packets,
		FirstTime:  t,
		LastTime:   t,
		UpdateTime: now,
	}
	sNode := s.lookupNode(server)
	se.ServerName = sNode.Name
	se.ServerNodeID = sNode.ID
	se.Services[service] = 1
	setServerPenalty(se)
	putEnt(s, KindServer, id, se)
}

func getFlowDir(fr *FlowRecord) (server, client, service string) {
	guc1 := isGlobalUnicast(fr.SrcIP)
	guc2 := isGlobalUnicast(fr.DstIP)
	if !guc1 && !guc2 && !strings.HasPrefix(fr.SrcIP, "fe80::") &&
		!strings.HasPrefix(fr.SrcIP, "169.254.") {
		// Only unicast or link-local included
		return
	}
	if fr.Prot == 1 {
		// ICMP
		return getFlowDirICMP(fr)
	}
	if fr.Prot == 2 {
		server = fr.SrcIP
		client = fr.DstIP
		service = "igmp"
		return
	}
	s1, ok1 := getServiceName(fr.Prot, fr.SrcPort)
	s2, ok2 := getServiceName(fr.Prot, fr.DstPort)
	if ok1 {
		if ok2 {
			if !guc2 || isDstServer(fr) {
				server = fr.DstIP
				client = fr.SrcIP
				service = s2
			} else {
				server = fr.SrcIP
				client = fr.DstIP
				service = s1
			}
		} else {
			server = fr.SrcIP
			client = fr.DstIP
			service = s1
		}
	} else if ok2 {
		server = fr.DstIP
		client = fr.SrcIP
		service = s2
	} else {
		// Unknown port
		server = fr.DstIP
		client = fr.SrcIP
		service = getOtherProtName(fr.Prot, fr.DstPort)
	}
	return
}

func getFlowDirICMP(fr *FlowRecord) (server, client, service string) {
	icmpType := fr.DstPort / 256
	switch icmpType {
	case 0, 3, 11, 12:
		service = fmt.Sprintf("%d/icmp", icmpType)
		server = fr.SrcIP
		client = fr.DstIP
	case 5, 8, 13:
		service = fmt.Sprintf("%d/icmp", icmpType)
		server = fr.DstIP
		client = fr.SrcIP
	default:
		service = fmt.Sprintf("%d/icmp", icmpType)
		server = fr.DstIP
		client = fr.SrcIP
	}
	return
}

func isDstServer(fr *FlowRecord) bool {
	if strings.HasSuffix(fr.DstIP, ".255") {
		return true
	}
	if strings.HasSuffix(fr.SrcIP, ".255") {
		return false
	}
	srcP := isPrivateAddr(fr.SrcIP)
	dstP := isPrivateAddr(fr.DstIP)
	if srcP && !dstP {
		return true
	}
	if !srcP && dstP {
		return false
	}
	return fr.SrcPort >= fr.DstPort
}

var privateCIDRs []*net.IPNet

func initPrivateCIDRs() {
	if len(privateCIDRs) > 0 {
		return
	}
	for _, ps := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7"} {
		if _, cidr, err := net.ParseCIDR(ps); err == nil {
			privateCIDRs = append(privateCIDRs, cidr)
		}
	}
}

func isPrivateAddr(a string) bool {
	initPrivateCIDRs()
	ip := net.ParseIP(a)
	if ip == nil {
		return false
	}
	for _, cidr := range privateCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func isGlobalUnicast(a string) bool {
	ip := net.ParseIP(a)
	if ip == nil {
		return false
	}
	return ip.IsGlobalUnicast()
}

func getOtherProtName(prot, port int) string {
	switch prot {
	case 6:
		if port > 0 {
			return fmt.Sprintf("%d/tcp", port)
		}
		return "other/tcp"
	case 17:
		if port > 0 {
			return fmt.Sprintf("%d/udp", port)
		}
		return "other/udp"
	default:
		return fmt.Sprintf("other/%d", prot)
	}
}

func parseProtocolNumber(p string) int {
	switch strings.ToLower(p) {
	case "icmp":
		return 1
	case "igmp":
		return 2
	case "tcp":
		return 6
	case "egp":
		return 8
	case "udp":
		return 17
	case "ipv6-icmp":
		return 58
	default:
		if n, err := strconv.Atoi(p); err == nil {
			return n
		}
		return 0
	}
}

var knownServices = map[string]string{
	"20/tcp":   "ftp-data",
	"21/tcp":   "ftp",
	"22/tcp":   "ssh",
	"23/tcp":   "telnet",
	"25/tcp":   "smtp",
	"53/tcp":   "dns",
	"53/udp":   "dns",
	"67/udp":   "dhcps",
	"68/udp":   "dhcpc",
	"69/udp":   "tftp",
	"80/tcp":   "http",
	"88/tcp":   "kerberos",
	"88/udp":   "kerberos",
	"110/tcp":  "pop3",
	"123/udp":  "ntp",
	"137/udp":  "netbios-ns",
	"138/udp":  "netbios-dgm",
	"139/tcp":  "netbios-ssn",
	"143/tcp":  "imap",
	"161/udp":  "snmp",
	"162/udp":  "snmptrap",
	"389/tcp":  "ldap",
	"389/udp":  "ldap",
	"443/tcp":  "https",
	"443/udp":  "quic",
	"445/tcp":  "microsoft-ds",
	"514/udp":  "syslog",
	"514/tcp":  "syslog-tcp",
	"636/tcp":  "ldaps",
	"993/tcp":  "imaps",
	"995/tcp":  "pop3s",
	"1433/tcp": "ms-sql-s",
	"1521/tcp": "oracle",
	"1812/udp": "radius",
	"1813/udp": "radius-acct",
	"1883/tcp": "mqtt",
	"3306/tcp": "mysql",
	"3389/tcp": "rdp",
	"5432/tcp": "postgresql",
	"6379/tcp": "redis",
	"8080/tcp": "http-alt",
	"8443/tcp": "https-alt",
	"8883/tcp": "mqtts",
}

func getServiceName(prot, port int) (string, bool) {
	var key string
	switch prot {
	case 6:
		key = fmt.Sprintf("%d/tcp", port)
	case 17:
		key = fmt.Sprintf("%d/udp", port)
	default:
		key = fmt.Sprintf("%d/%d", port, prot)
	}
	if name, ok := knownServices[key]; ok {
		return name, true
	}
	if port > 0 && port < 1024 {
		return key, true
	}
	return key, false
}

func setFlowPenalty(f *FlowEnt) {
	f.Penalty = 0
	if !isSafeCountry(f.ServerLoc) {
		f.Penalty += 2
	}
	for sv := range f.Services {
		if !isSafeService(sv) {
			f.Penalty++
		}
	}
	if f.ServerName == "" || f.ServerName == f.Server {
		f.Penalty++
	}
}

func setServerPenalty(s *ServerEnt) {
	s.Penalty = 0
	if !isSafeCountry(s.Loc) {
		s.Penalty += 2
	}
	for sv := range s.Services {
		if !isSafeService(sv) {
			s.Penalty++
		}
	}
	if s.ServerName == "" || s.ServerName == s.Server {
		s.Penalty++
	}
}

func isSafeCountry(loc string) bool {
	if loc == "" || loc == "LOCAL" || loc == "JP" || loc == "US" || loc == "GB" || loc == "DE" || loc == "FR" {
		return true
	}
	// Suspicious/high risk flags or unknown countries outside safe list
	switch loc {
	case "RU", "CN", "KP", "IR", "SY":
		return false
	default:
		return true
	}
}

func isSafeService(svc string) bool {
	s := strings.ToLower(svc)
	if s == "telnet" || s == "23/tcp" || s == "ms-sql-s" || s == "1433/tcp" || s == "microsoft-ds" || s == "445/tcp" || s == "rdp" || s == "3389/tcp" {
		return false
	}
	return true
}
