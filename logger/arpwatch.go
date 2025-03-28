package logger

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/cmd"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/ping"
)

var arpTable = make(map[string]string)
var macToIPTable = make(map[string]string)
var localCheckAddrs []string

var ResetArpWatch = false

func resetArpTable() {
	arpTable = make(map[string]string)
	macToIPTable = make(map[string]string)
	localCheckAddrs = []string{}
}

func UpdateArpTable(ip, mac string) {
	updateArpTable(ip, mac)
}

func arpWatch(stopCh chan bool) {
	log.Println("start arp")
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Start ARP watch"),
	})
	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		arpTable[a.IP] = a.MAC
		return true
	})
	lastArpWatchRange := ""
	checkArpTable()
	timer := time.NewTicker(time.Second * 300)
	pinger := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-stopCh:
			timer.Stop()
			log.Println("stop arp")
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:  "system",
				Level: "info",
				Event: i18n.Trans("Stop ARP watch"),
			})
			return
		case <-pinger.C:
			if ResetArpWatch {
				resetArpTable()
				checkArpTable()
				ResetArpWatch = false
			}
			if lastArpWatchRange != datastore.MapConf.ArpWatchRange {
				// 変更されたら更新する
				localCheckAddrs = []string{}
				lastArpWatchRange = datastore.MapConf.ArpWatchRange
				makeLoacalCheckAddrs()
			}
			i := 0
			for len(localCheckAddrs) > 0 {
				i++
				a := localCheckAddrs[0]
				ping.DoPing(a, 1, 0, 64, 0)
				localCheckAddrs[0] = ""
				localCheckAddrs = localCheckAddrs[1:]
				if i > 50 {
					break
				}
			}
		case <-timer.C:
			checkArpTable()
			if len(localCheckAddrs) < 1 {
				makeLoacalCheckAddrs()
			} else {
				log.Printf("arp watch wait ip count=%d", len(localCheckAddrs))
			}
		}
	}
}

var lastAddressUsage = make(map[string]float64)

func makeLoacalCheckAddrs() {
	ipMap := make(map[string]bool)
	for _, r := range strings.Split(datastore.MapConf.ArpWatchRange, ",") {
		a := strings.SplitN(r, "-", 2)
		var sIP uint32
		var eIP uint32
		if len(a) == 1 {
			// CIDR
			ip, ipnet, err := net.ParseCIDR(r)
			if err != nil {
				continue
			}
			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}
			sIP = ip2int(ipv4)
			for eIP = sIP; ipnet.Contains(int2ip(eIP)); eIP++ {
			}
			eIP--
		} else {
			sIP = ip2int(net.ParseIP(a[0]))
			eIP = ip2int(net.ParseIP(a[1]))
		}
		if sIP >= eIP {
			continue
		}
		localIPCount := 0
		localHitCount := 0
		for nIP := sIP; nIP <= eIP; nIP++ {
			ip := int2ip(nIP)
			if !ip.IsGlobalUnicast() || ip.IsMulticast() {
				continue
			}
			sa := ip.String()
			localIPCount++
			if e := datastore.GetArpEnt(sa); e != nil {
				// ARPテーブルに存在する
				localHitCount++
				ipMap[sa] = true
				continue
			}
			if _, ok := ipMap[sa]; ok {
				// 同じIPアドレスを追加しない。
				continue
			}
			ipMap[sa] = true
			localCheckAddrs = append(localCheckAddrs, sa)
		}
		lau := 0.0
		if localIPCount > 0 {
			lau = 100.0 * float64(localHitCount) / float64(localIPCount)
		} else {
			continue
		}
		if au, ok := lastAddressUsage[r]; ok && au == lau {
			continue
		}
		lastAddressUsage[r] = lau
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "arpwatch",
			Level: "info",
			Event: fmt.Sprintf(i18n.Trans("ARP Watch range %s usage:%d/%d %.2f%%"), r, localHitCount, localIPCount, lau),
		})
	}
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

func checkArpTable() {
	if runtime.GOOS == "windows" {
		checkArpTableWindows()
	} else {
		checkArpTableUnix()
	}
	checkNodeMAC()
}

func checkArpTableWindows() {
	out, err := cmd.GetCmd("arp", []string{"-a"}).Output()
	if err != nil {
		log.Printf("check arp table err=%v", err)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		updateArpTable(fields[0], fields[1])
	}
}

func checkArpTableUnix() {
	out, err := cmd.GetCmd("arp", []string{"-an"}).Output()
	if err != nil {
		log.Printf("check arp table err=%v", err)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// strip brackets around IP
		ip := strings.Replace(fields[1], "(", "", -1)
		ip = strings.Replace(ip, ")", "", -1)
		updateArpTable(ip, fields[3])
	}
}

func updateArpTable(ip, mac string) {
	if !strings.Contains(ip, ".") || !strings.ContainsAny(mac, ":-") {
		return
	}
	mac = normMACAddr(mac)
	if strings.HasPrefix(mac, "FF") || strings.HasPrefix(mac, "01") {
		return
	}
	datastore.UpdateArpEnt(ip, mac)
	m, ok := arpTable[ip]
	if !ok {
		// New
		arpTable[ip] = mac
		logCh <- &datastore.LogEnt{
			Time: time.Now().UnixNano(),
			Type: "arplog",
			Log:  fmt.Sprintf("New,%s,%s", ip, mac),
		}
		log.Printf("New %s %s", ip, mac)
		return
	}
	if mac != m {
		// Change
		arpTable[ip] = mac
		logCh <- &datastore.LogEnt{
			Time: time.Now().UnixNano(),
			Type: "arplog",
			Log:  fmt.Sprintf("Change,%s,%s,%s", ip, mac, m),
		}
		log.Printf("Change %s %s -> %s", ip, m, mac)
		return
	}
	// No Change
	// Check MAC to IP Table
	ipot, ok := macToIPTable[mac]
	if !ok {
		macToIPTable[mac] = ip
	} else if ip != ipot {
		macToIPTable[mac] = ip
	}
}

func normMACAddr(m string) string {
	m = strings.Replace(m, "-", ":", -1)
	a := strings.Split(m, ":")
	r := ""
	for _, e := range a {
		if r != "" {
			r += ":"
		}
		if len(e) == 1 {
			r += "0"
		}
		r += e
	}
	return strings.ToUpper(r)
}

// ノードリストのMACアドレスをチェックする
func checkNodeMAC() {
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if n.AddrMode == "mac" {
			checkFixMACMode(n)
			return true
		}
		if n.AddrMode == "host" {
			checkFixHostMode(n)
			return true
		}
		if mac, ok := arpTable[n.IP]; ok {
			if n.MAC != mac {
				v := datastore.FindVendor(mac)
				if v != "" {
					n.Vendor = v
				}
				if n.MAC == "" {
					datastore.AddEventLog(&datastore.EventLogEnt{
						Type:     "arpwatch",
						Level:    "info",
						NodeID:   n.ID,
						NodeName: n.Name,
						Event:    fmt.Sprintf(i18n.Trans("Set MAC Address %s"), mac),
					})
				} else {
					datastore.AddEventLog(&datastore.EventLogEnt{
						Type:     "arpwatch",
						Level:    "warn",
						NodeID:   n.ID,
						NodeName: n.Name,
						Event:    fmt.Sprintf(i18n.Trans("Change MAC Address %s -> %s"), n.MAC, mac),
					})
				}
				n.MAC = mac
			}
		}
		return true
	})
}

func checkFixMACMode(n *datastore.NodeEnt) {
	if n.MAC == "" {
		if mac, ok := arpTable[n.IP]; ok {
			v := datastore.FindVendor(mac)
			if v != "" {
				mac += fmt.Sprintf("(%s)", v)
			}
			n.MAC = mac
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "system",
				Level:    "warn",
				NodeID:   n.ID,
				NodeName: n.Name,
				Event:    fmt.Sprintf(i18n.Trans("Fixed MAC address node MAC is %s"), n.MAC),
			})
		}
		return
	}
	a := strings.Split(n.MAC, "(")
	if len(a) < 1 {
		return
	}
	key := strings.TrimSpace(a[0])
	if ip, ok := macToIPTable[key]; ok {
		if ip != n.IP {
			oldIP := n.IP
			n.IP = ip
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "system",
				Level:    "warn",
				NodeID:   n.ID,
				NodeName: n.Name,
				Event:    fmt.Sprintf(i18n.Trans("Fixed MAC address node '%s' Chnage IP address from '%s' to '%s'"), n.MAC, oldIP, ip),
			})
		}
	}
}

func checkFixHostMode(n *datastore.NodeEnt) {
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*50)
	defer cancel()
	ips, err := r.LookupHost(ctx, n.Name)
	if err != nil {
		log.Printf("check fixed host err=%v", err)
		return
	}
	hitIP := ""
	for _, ip := range ips {
		if n.IP == ip {
			return
		}
		if strings.Contains(ip, ":") || hitIP != "" {
			continue
		}
		nIP := net.ParseIP(ip)
		if nIP.IsGlobalUnicast() {
			hitIP = ip
		}
	}
	if hitIP == "" {
		log.Printf("check fixed host no ip address")
		return
	}
	oldIP := n.IP
	n.IP = hitIP
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "system",
		Level:    "warn",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    fmt.Sprintf(i18n.Trans("Fixed host name node '%s' Chnage IP from '%s' to ''%s"), n.Name, oldIP, hitIP),
	})
}
