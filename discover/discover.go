// Package discover : 自動発見
package discover

/* discover.go: 自動発見の処理
自動発見は、PINGを実行して、応答があるノードに関してSNMPの応答があるか確認する
*/

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gosnmp/gosnmp"
	"github.com/signalsciences/ipv4"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/ping"
)

// GRID : 自動発見時にノードを配置する間隔
const GRID = 90

var (
	Stat DiscoverStat
	Stop bool
	X    int
	Y    int
)

type DiscoverStat struct {
	Running   bool   `json:"Running"`
	Total     uint32 `json:"Total"`
	Sent      uint32 `json:"Sent"`
	Found     uint32 `json:"Found"`
	Snmp      uint32 `json:"Snmp"`
	Web       uint32 `json:"Web"`
	Mail      uint32 `json:"Mail"`
	SSH       uint32 `json:"SSH"`
	File      uint32 `json:"File"`
	RDP       uint32 `json:"RDP"`
	LDAP      uint32 `json:"LDAP"`
	Wait      int    `json:"Wait"`
	StartTime int64  `json:"StartTime"`
	Now       int64  `json:"Now"`
}

type discoverInfoEnt struct {
	IP          string
	HostName    string
	SysName     string
	SysObjectID string
	SysDescr    string
	MAC         string
	Vendor      string
	HTTPTitle   string
	HTTPServer  string
	HTTPBody    string
	IfMap       map[string]string
	ServerList  map[string]bool
	X           int
	Y           int
	SnmpConf    *datastore.SnmpConfEnt
}

// StopDiscover : 自動発見を停止する
func StopDiscover() {
	Stop = true
	st := time.Now()
	for Stat.Running {
		time.Sleep(time.Millisecond * 100)
		if time.Since(st) > 3*time.Second {
			log.Println("StopDiscover timeout waiting for discover to stop")
			break
		}
	}
}

func StartDiscover() error {
	if Stat.Running {
		return fmt.Errorf("discover already runnning")
	}
	return Discover()
}

func Discover() error {
	sip, err := ipv4.FromDots(datastore.DiscoverConf.StartIP)
	if err != nil {
		return fmt.Errorf("discover start ip err=%v", err)
	}
	eip, err := ipv4.FromDots(datastore.DiscoverConf.EndIP)
	if err != nil {
		return fmt.Errorf("discover end ip err=%v", err)
	}
	if sip > eip {
		return fmt.Errorf("discover start ip > end ip")
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Start discover %s - %s"), datastore.DiscoverConf.StartIP, datastore.DiscoverConf.EndIP),
	})
	Stop = false
	Stat.Total = eip - sip + 1
	Stat.Sent = 0
	Stat.Found = 0
	Stat.Snmp = 0
	Stat.Web = 0
	Stat.Mail = 0
	Stat.SSH = 0
	Stat.File = 0
	Stat.RDP = 0
	Stat.Wait = 0
	Stat.Running = true
	Stat.StartTime = time.Now().Unix()
	Stat.Now = Stat.StartTime
	X = (1 + datastore.DiscoverConf.X/GRID) * GRID
	Y = (1 + datastore.DiscoverConf.Y/GRID) * GRID
	var mu sync.Mutex
	sem := make(chan bool, 16)
	portScanSem := make(chan bool, 2)
	pacer := time.NewTicker(time.Millisecond * 80)
	go func() {
		defer pacer.Stop()
		for ; sip <= eip && !Stop; sip++ {
			select {
			case <-pacer.C:
			}
			if Stop {
				break
			}
			sem <- true
			Stat.Sent++
			Stat.Now = time.Now().Unix()
			go func(ip uint32) {
				defer func() {
					<-sem
				}()
				if Stop {
					return
				}
				ipstr := ipv4.ToDots(ip)
				node := datastore.FindNodeFromIP(ipstr)
				if node != nil && !datastore.DiscoverConf.ReCheck {
					log.Printf("discover skip ip=%s", ipstr)
					return
				}
				r := ping.DoPing(ipstr, datastore.DiscoverConf.Timeout, datastore.DiscoverConf.Retry, 64, 0)
				if r.Stat == ping.PingOK {
					dent := discoverInfoEnt{
						IP:         ipstr,
						IfMap:      make(map[string]string),
						ServerList: make(map[string]bool),
					}
					if arp := datastore.GetArpEnt(ipstr); arp != nil {
						dent.MAC = arp.MAC
						dent.Vendor = arp.Vendor
						if dent.Vendor == "" && dent.MAC != "" {
							dent.Vendor = datastore.FindVendor(dent.MAC)
						}
					}
					if node != nil {
						if dent.MAC == "" && node.MAC != "" {
							dent.MAC = node.MAC
						}
						if (dent.Vendor == "" || dent.Vendor == "Unknown") && node.Vendor != "" {
							dent.Vendor = node.Vendor
						}
					}
					r := &net.Resolver{}
					ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
					defer cancel()
					if names, err := r.LookupAddr(ctx, ipstr); err == nil && len(names) > 0 {
						dent.HostName = names[0]
					}
					getSnmpInfo(ipstr, &dent)
					if datastore.DiscoverConf.PortScan {
						portScanSem <- true
						checkServer(&dent)
						<-portScanSem
					}
					mu.Lock()
					dent.X = X
					dent.Y = Y
					Stat.Found++
					if node == nil {
						X += GRID
						if X > GRID*10 {
							X = GRID
							Y += GRID
						}
					}
					if datastore.DiscoverConf.AddNetwork {
						if dent.ServerList["lldp"] || dent.ServerList["bridge"] {
							if datastore.FindNetworkByIP(ipstr) == nil {
								X = GRID
								Y += GRID
							}
						}
					}
					if dent.SysName != "" {
						Stat.Snmp++
					}
					if dent.ServerList["http"] || dent.ServerList["https"] {
						Stat.Web++
					}
					if dent.ServerList["cifs"] || dent.ServerList["nfs"] {
						Stat.File++
					}
					if dent.ServerList["rdp"] || dent.ServerList["vnc"] {
						Stat.RDP++
					}
					if dent.ServerList["ldap"] || dent.ServerList["ldaps"] || dent.ServerList["kerberos"] {
						Stat.LDAP++
					}
					if dent.ServerList["smtp"] || dent.ServerList["imap"] || dent.ServerList["pop3"] {
						Stat.Mail++
					}
					if dent.ServerList["ssh"] {
						Stat.SSH++
					}
					if node == nil {
						addFoundNode(&dent)
					} else {
						updateNode(node, &dent)
					}
					mu.Unlock()
				}
			}(sip)
		}
		for len(sem) > 0 {
			time.Sleep(time.Millisecond * 10)
			Stat.Now = time.Now().Unix()
			Stat.Wait = len(sem)
		}
		if datastore.DiscoverConf.AutoLine > datastore.AutoLineNone {
			if _, _, err := backend.AutoConnectLines(datastore.DiscoverConf.AutoLine); err != nil {
				log.Printf("auto connect lines err=%v", err)
			}
		}
		if datastore.DiscoverConf.AutoLayout > datastore.AutoLayoutNone {
			if _, err := backend.OptimizeLayout(datastore.DiscoverConf.AutoLayout); err != nil {
				log.Printf("auto layout err=%v", err)
			}
		}
		Stat.Running = false
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf(i18n.Trans("End discover %s - %s"), datastore.DiscoverConf.StartIP, datastore.DiscoverConf.EndIP),
		})
	}()
	return nil
}

func ClearStat() {
	if Stat.Running {
		return
	}
	Stat.Total = 0
	Stat.Sent = 0
	Stat.Found = 0
	Stat.Snmp = 0
	Stat.Web = 0
	Stat.Mail = 0
	Stat.SSH = 0
	Stat.File = 0
	Stat.RDP = 0
	Stat.StartTime = 0
	Stat.Now = 0
}

func trySnmp(t string, snmpConf *datastore.SnmpConfEnt, dent *discoverInfoEnt) bool {
	agent := &gosnmp.GoSNMP{
		Target:    t,
		Port:      161,
		Transport: "udp",
		Community: snmpConf.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(datastore.DiscoverConf.Timeout) * time.Second,
		Retries:   datastore.DiscoverConf.Retry,
		MaxOids:   gosnmp.MaxOids,
	}
	switch snmpConf.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 snmpConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: snmpConf.SnmpPassword,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 snmpConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: snmpConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        snmpConf.SnmpPassword,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 snmpConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: snmpConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        snmpConf.SnmpPassword,
		}
	case "v3sha256aes128":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 snmpConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: snmpConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        snmpConf.SnmpPassword,
		}
	case "v3sha512aes256":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 snmpConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA512,
			AuthenticationPassphrase: snmpConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        snmpConf.SnmpPassword,
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("discover snmp connect err=%v", err)
		return false
	}
	defer agent.Conn.Close()
	oids := []string{
		datastore.MIBDB.NameToOID("sysName"),
		datastore.MIBDB.NameToOID("sysObjectID"),
		datastore.MIBDB.NameToOID("sysDescr"),
	}
	result, err := agent.GetNext(oids)
	if err != nil {
		return false
	}
	hasInfo := false
	for _, variable := range result.Variables {
		name := datastore.MIBDB.OIDToName(variable.Name)
		if name == "sysName.0" {
			dent.SysName = getMIBStringVal(variable.Value)
			hasInfo = true
		} else if name == "sysObjectID.0" {
			dent.SysObjectID = getMIBStringVal(variable.Value)
			hasInfo = true
		} else if name == "sysDescr.0" {
			dent.SysDescr = getMIBStringVal(variable.Value)
			hasInfo = true
		}
	}
	if !hasInfo {
		return false
	}
	dent.SnmpConf = snmpConf
	agent.Walk(datastore.MIBDB.NameToOID("ifType"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) == 2 &&
			a[0] == "ifType" &&
			gosnmp.ToBigInt(variable.Value).Int64() == 6 {
			dent.IfMap[a[1]] = fmt.Sprintf("#%s", a[1])
		}
		return nil
	})
	agent.Walk(datastore.MIBDB.NameToOID("ifName"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) == 2 {
			if _, ok := dent.IfMap[a[1]]; ok {
				dent.IfMap[a[1]] = datastore.GetMIBValueString(a[0], &variable, false)
			}
		}
		return nil
	})
	agent.Walk(datastore.MIBDB.NameToOID("lldpLocalSystemData"), func(variable gosnmp.SnmpPDU) error {
		dent.ServerList["lldp"] = true
		return fmt.Errorf("checkend")
	})
	if !dent.ServerList["lldp"] {
		agent.Walk(datastore.MIBDB.NameToOID("dot1dBaseBridgeAddress"), func(variable gosnmp.SnmpPDU) error {
			dent.ServerList["bridge"] = true
			return fmt.Errorf("checkend")
		})
	}
	return true
}

func getSnmpInfo(t string, dent *discoverInfoEnt) {
	for _, sc := range datastore.GetDiscoverSnmpConfigs() {
		confCopy := sc
		if trySnmp(t, &confCopy, dent) {
			break
		}
	}
}

func addFoundNode(dent *discoverInfoEnt) {
	funcList := []string{}
	if (dent.Vendor == "" || dent.Vendor == "Unknown") && dent.MAC != "" {
		dent.Vendor = datastore.FindVendor(dent.MAC)
	}
	n := datastore.NodeEnt{
		Name:   dent.HostName,
		IP:     dent.IP,
		MAC:    dent.MAC,
		Vendor: dent.Vendor,
		Icon:   "desktop",
		X:      dent.X,
		Y:      dent.Y,
		Descr:  fmt.Sprintf(i18n.Trans("Found at %s"), time.Now().Format("2006/01/02")),
	}
	if n.Name == "" {
		if dent.SysName != "" {
			n.Name = dent.SysName
		} else {
			n.Name = dent.IP
		}
	}
	if dent.SysObjectID != "" {
		if dent.SnmpConf != nil {
			n.SnmpMode = dent.SnmpConf.SnmpMode
			n.User = dent.SnmpConf.SnmpUser
			n.Password = dent.SnmpConf.SnmpPassword
			n.Community = dent.SnmpConf.Community
		} else {
			n.SnmpMode = datastore.MapConf.SnmpMode
			n.User = datastore.MapConf.SnmpUser
			n.Password = datastore.MapConf.SnmpPassword
			n.Community = datastore.MapConf.Community
		}
		n.Icon = "hdd"
		funcList = append(funcList, "snmp")
	}
	var detectRes *datastore.DetectResult
	if datastore.DiscoverConf.AutoDetect {
		if dent.HTTPTitle == "" && dent.HTTPServer == "" && dent.HTTPBody == "" {
			dent.HTTPTitle, dent.HTTPServer, dent.HTTPBody = backend.FetchWebSignatures(dent.IP, "")
		}
		input := &datastore.DetectInput{
			IP:          dent.IP,
			Name:        dent.SysName,
			HostName:    dent.HostName,
			SysObjectID: dent.SysObjectID,
			SysDescr:    dent.SysDescr,
			HTTPTitle:   dent.HTTPTitle,
			HTTPServer:  dent.HTTPServer,
			HTTPBody:    dent.HTTPBody,
			Vendor:      dent.Vendor,
		}
		detectRes = datastore.DetectNode(input)
		if detectRes != nil && detectRes.Icon != "" && detectRes.RuleID != "unknown" {
			n.Icon = detectRes.Icon
			if detectRes.Name != "" {
				n.Descr += fmt.Sprintf(" [%s]", detectRes.Name)
			}
		}
	}
	if len(dent.ServerList) > 0 {
		for _, s := range []string{
			"http", "https", "pop3", "imap", "smtp", "ssh", "cifs", "nfs",
			"vnc", "rdp", "ldap", "ldaps", "kerberos", "lldp",
		} {
			if dent.ServerList[s] {
				funcList = append(funcList, s)
			}
		}
	}
	if len(funcList) > 0 {
		n.Descr += " "
		n.Descr += i18n.Trans("Protocol:") + strings.Join(funcList, ",")
	}
	if err := datastore.AddNode(&n); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "discover",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    i18n.Trans("Add by discover"),
	})
	hasSnmp := dent.SysObjectID != ""
	isNetwork := hasSnmp && (dent.ServerList["lldp"] || dent.ServerList["bridge"] ||
		(detectRes != nil && (detectRes.Category == "switch" || detectRes.Icon == "switch")))
	if datastore.DiscoverConf.AddNetwork && isNetwork && datastore.FindNetworkByIP(n.IP) == nil {
		net := &datastore.NetworkEnt{
			Name:      n.Name,
			IP:        n.IP,
			X:         n.X + GRID,
			Y:         n.Y,
			SnmpMode:  n.SnmpMode,
			Community: n.Community,
			User:      n.User,
			Password:  n.Password,
			HPorts:    24,
			Ports:     []datastore.PortEnt{},
		}
		if dent.SnmpConf != nil {
			net.SnmpMode = dent.SnmpConf.SnmpMode
			net.Community = dent.SnmpConf.Community
			net.User = dent.SnmpConf.SnmpUser
			net.Password = dent.SnmpConf.SnmpPassword
		}
		datastore.AddNetwork(net)
	}
	if !datastore.DiscoverConf.AddPolling {
		return
	}
	addPolling(dent, &n, detectRes)
}

func updateNode(n *datastore.NodeEnt, dent *discoverInfoEnt) {
	if n.Name == n.IP {
		if dent.SysName != "" {
			n.Name = dent.SysName
		}
	}
	// Synchronize MAC
	if n.MAC == "" && dent.MAC != "" {
		n.MAC = dent.MAC
	} else if dent.MAC == "" && n.MAC != "" {
		dent.MAC = n.MAC
	}
	// Synchronize Vendor
	if dent.Vendor == "" || dent.Vendor == "Unknown" {
		if n.Vendor != "" && n.Vendor != "Unknown" {
			dent.Vendor = n.Vendor
		} else if dent.MAC != "" {
			dent.Vendor = datastore.FindVendor(dent.MAC)
		} else if arp := datastore.GetArpEnt(n.IP); arp != nil {
			if arp.Vendor != "" && arp.Vendor != "Unknown" {
				dent.Vendor = arp.Vendor
			} else if arp.MAC != "" {
				dent.Vendor = datastore.FindVendor(arp.MAC)
			}
		}
	}
	if n.Vendor == "" || n.Vendor == "Unknown" {
		n.Vendor = dent.Vendor
	}

	if dent.SysObjectID != "" {
		if dent.SnmpConf != nil {
			n.SnmpMode = dent.SnmpConf.SnmpMode
			n.User = dent.SnmpConf.SnmpUser
			n.Password = dent.SnmpConf.SnmpPassword
			n.Community = dent.SnmpConf.Community
			if n.Icon == "desktop" {
				n.Icon = "hdd"
				n.Descr += " / snmp対応"
			}
		} else if n.User == "" && n.Community == "" {
			n.SnmpMode = datastore.MapConf.SnmpMode
			n.User = datastore.MapConf.SnmpUser
			n.Password = datastore.MapConf.SnmpPassword
			n.Community = datastore.MapConf.Community
			if n.Icon == "desktop" {
				n.Icon = "hdd"
				n.Descr += " / snmp対応"
			}
		}
	}

	// If SNMP info was not captured in dent, try existing node's SNMP config
	if dent.SysObjectID == "" && n.SnmpMode != "" && n.SnmpMode != "none" {
		agent := backend.GetSNMPAgent(n)
		if agent != nil {
			if err := agent.Connect(); err == nil {
				defer agent.Conn.Close()
				oids := []string{
					datastore.MIBDB.NameToOID("sysObjectID"),
					datastore.MIBDB.NameToOID("sysDescr"),
					datastore.MIBDB.NameToOID("sysName"),
				}
				if res, err := agent.GetNext(oids); err == nil {
					for _, v := range res.Variables {
						name := datastore.MIBDB.OIDToName(v.Name)
						if strings.HasPrefix(name, "sysObjectID") {
							dent.SysObjectID = fmt.Sprintf("%v", v.Value)
						} else if strings.HasPrefix(name, "sysDescr") {
							switch val := v.Value.(type) {
							case string:
								dent.SysDescr = val
							case []byte:
								dent.SysDescr = string(val)
							}
						} else if strings.HasPrefix(name, "sysName") {
							dent.SysName = getMIBStringVal(v.Value)
						}
					}
				}
			}
		}
	}

	// If HTTP info was not captured in dent, try fetching web signatures
	if dent.HTTPTitle == "" && dent.HTTPServer == "" && dent.HTTPBody == "" {
		dent.HTTPTitle, dent.HTTPServer, dent.HTTPBody = backend.FetchWebSignatures(n.IP, n.URL)
	}

	var detectRes *datastore.DetectResult
	if datastore.DiscoverConf.AutoDetect {
		input := &datastore.DetectInput{
			IP:          dent.IP,
			Name:        n.Name,
			HostName:    dent.HostName,
			SysObjectID: dent.SysObjectID,
			SysDescr:    dent.SysDescr,
			HTTPTitle:   dent.HTTPTitle,
			HTTPServer:  dent.HTTPServer,
			HTTPBody:    dent.HTTPBody,
			Vendor:      dent.Vendor,
		}
		detectRes = datastore.DetectNode(input)
		if detectRes != nil && detectRes.Icon != "" && detectRes.RuleID != "unknown" {
			isAutoTagged := strings.Contains(n.Descr, "[") && strings.Contains(n.Descr, "]")
			if n.Icon == "desktop" || n.Icon == "hdd" || n.Icon == "" || datastore.DiscoverConf.ReCheck || isAutoTagged {
				n.Icon = detectRes.Icon
			}
			if detectRes.Name != "" {
				reTag := regexp.MustCompile(`\s*\[[^\]]+\]`)
				cleanDescr := strings.TrimSpace(reTag.ReplaceAllString(n.Descr, ""))
				if cleanDescr != "" {
					n.Descr = cleanDescr + fmt.Sprintf(" [%s]", detectRes.Name)
				} else {
					n.Descr = fmt.Sprintf("[%s]", detectRes.Name)
				}
			}
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "discover",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    i18n.Trans("Update by discover"),
	})
	hasSnmp := dent.SysObjectID != ""
	isNetwork := hasSnmp && (dent.ServerList["lldp"] || dent.ServerList["bridge"] ||
		(detectRes != nil && (detectRes.Category == "switch" || detectRes.Icon == "switch")))
	if datastore.DiscoverConf.AddNetwork && isNetwork {
		net := datastore.FindNetworkByIP(n.IP)
		if net == nil {
			newNet := &datastore.NetworkEnt{
				Name:      n.Name,
				IP:        n.IP,
				X:         n.X + GRID,
				Y:         n.Y,
				SnmpMode:  n.SnmpMode,
				Community: n.Community,
				User:      n.User,
				Password:  n.Password,
				HPorts:    24,
				Ports:     []datastore.PortEnt{},
			}
			if dent.SnmpConf != nil {
				newNet.SnmpMode = dent.SnmpConf.SnmpMode
				newNet.Community = dent.SnmpConf.Community
				newNet.User = dent.SnmpConf.SnmpUser
				newNet.Password = dent.SnmpConf.SnmpPassword
			}
			datastore.AddNetwork(newNet)
		} else if dent.SnmpConf != nil && (len(net.Ports) < 1 || net.Error != "") {
			net.SnmpMode = dent.SnmpConf.SnmpMode
			net.Community = dent.SnmpConf.Community
			net.User = dent.SnmpConf.SnmpUser
			net.Password = dent.SnmpConf.SnmpPassword
			net.Error = ""
			_ = datastore.UpdateNetwork(net)
		}
	}
	if err := datastore.UpdateNode(n); err != nil {
		log.Printf("discover updateNode save err=%v", err)
	}
	if !datastore.DiscoverConf.AddPolling {
		return
	}
	addPolling(dent, n, detectRes)
}

func getStaggeredNextTime(pollInt int) int64 {
	if pollInt < 10 {
		pollInt = 10
	}
	offset := 5 + rand.Intn(pollInt-4)
	return time.Now().UnixNano() + int64(offset)*1e9
}

func addPolling(dent *discoverInfoEnt, n *datastore.NodeEnt, detectRes *datastore.DetectResult) {
	p := &datastore.PollingEnt{
		NodeID:   n.ID,
		Name:     "PING",
		Type:     "ping",
		Level:    "low",
		State:    "unknown",
		PollInt:  datastore.MapConf.PollInt,
		Timeout:  datastore.MapConf.Timeout,
		Retry:    datastore.MapConf.Retry,
		NextTime: getStaggeredNextTime(datastore.MapConf.PollInt),
	}
	if err := datastore.AddPollingWithDupCheck(p); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for s := range dent.ServerList {
		name := ""
		ptype := ""
		params := ""
		mode := ""
		level := "off"
		switch s {
		case "http":
			name = "HTTP Server"
			ptype = "http"
			params = "http://" + n.IP
		case "https":
			name = "HTTPS Server"
			ptype = "http"
			mode = "https"
			params = "https://" + n.IP
		case "smtp":
			name = "SMTP Server"
			ptype = "tcp"
			params = "25"
			level = "low"
		case "pop3":
			name = "POP3 Server"
			ptype = "tcp"
			params = "110"
		case "imap":
			name = "IMAP Server"
			ptype = "tcp"
			params = "143"
			level = "low"
		case "ssh":
			name = "SSH Server"
			ptype = "tcp"
			params = "22"
		case "cifs":
			name = "CIFS Server"
			ptype = "tcp"
			params = "445"
		case "nfs":
			name = "NFS Server"
			ptype = "tcp"
			params = "2049"
		case "vnc":
			name = "VNC Server"
			ptype = "tcp"
			params = "5900"
		case "rdp":
			name = "RDP Server"
			ptype = "tcp"
			params = "3389"
		case "kerberos":
			name = "AD(kerberos) Server"
			ptype = "tcp"
			params = "88"
		case "ldap":
			name = "LDAP Server"
			ptype = "tcp"
			params = "389"
		case "ldaps":
			name = "LDAPS Server"
			ptype = "tcp"
			params = "636"
		default:
			continue
		}
		p = &datastore.PollingEnt{
			NodeID:   n.ID,
			Name:     name,
			Type:     ptype,
			Mode:     mode,
			Params:   params,
			Level:    level,
			State:    "unknown",
			PollInt:  datastore.MapConf.PollInt,
			Timeout:  datastore.MapConf.Timeout,
			Retry:    datastore.MapConf.Retry,
			NextTime: getStaggeredNextTime(datastore.MapConf.PollInt),
		}
		if err := datastore.AddPollingWithDupCheck(p); err != nil {
			log.Printf("discover err=%v", err)
			return
		}
	}
	if dent.SysObjectID == "" {
		return
	}
	p = &datastore.PollingEnt{
		NodeID:   n.ID,
		Name:     "sysUptime",
		Type:     "snmp",
		Mode:     "sysUpTime",
		Level:    "off",
		State:    "unknown",
		PollInt:  datastore.MapConf.PollInt,
		Timeout:  datastore.MapConf.Timeout,
		Retry:    datastore.MapConf.Retry,
		NextTime: getStaggeredNextTime(datastore.MapConf.PollInt),
	}
	if err := datastore.AddPollingWithDupCheck(p); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for i, name := range dent.IfMap {
		p = &datastore.PollingEnt{
			NodeID:   n.ID,
			Type:     "snmp",
			Name:     fmt.Sprintf("%s(%s)", name, i),
			Mode:     "ifOperStatus",
			Params:   i,
			Level:    "off",
			State:    "unknown",
			PollInt:  datastore.MapConf.PollInt,
			Timeout:  datastore.MapConf.Timeout,
			Retry:    datastore.MapConf.Retry,
			NextTime: getStaggeredNextTime(datastore.MapConf.PollInt),
		}
		if err := datastore.AddPollingWithDupCheck(p); err != nil {
			log.Printf("discover err=%v", err)
			return
		}
	}
	if detectRes != nil && len(detectRes.SensorPollings) > 0 {
		agent := backend.GetSNMPAgent(n)
		if agent != nil {
			if err := agent.Connect(); err == nil {
				defer agent.Conn.Close()
				for _, sp := range detectRes.SensorPollings {
					if datastore.CheckSensorSupport(agent, sp.Params) {
						level := sp.Level
						if level == "" {
							level = "off"
						}
						mode := sp.Mode
						if mode == "" {
							mode = "get"
						}
						p := &datastore.PollingEnt{
							NodeID:   n.ID,
							Name:     sp.Name,
							Type:     sp.Type,
							Mode:     mode,
							Params:   sp.Params,
							Script:   sp.Script,
							Level:    level,
							State:    "unknown",
							PollInt:  datastore.MapConf.PollInt,
							Timeout:  datastore.MapConf.Timeout,
							Retry:    datastore.MapConf.Retry,
							NextTime: getStaggeredNextTime(datastore.MapConf.PollInt),
						}
						if err := datastore.AddPollingWithDupCheck(p); err != nil {
							log.Printf("discover add sensor polling err=%v", err)
						}
					}
				}
			}
		}
	}
}

// サーバーの確認
func checkServer(dent *discoverInfoEnt) {
	checkList := map[string]string{
		"http":     "80",
		"https":    "443",
		"pop3":     "110",
		"imap":     "143",
		"smtp":     "25",
		"ssh":      "22",
		"cifs":     "445",
		"nfs":      "2049",
		"vnc":      "5900",
		"rdp":      "3389",
		"ldap":     "389",
		"ldaps":    "636",
		"kerberos": "88",
	}
	for s, p := range checkList {
		if Stop {
			return
		}
		time.Sleep(time.Millisecond * 100)
		if doTCPConnect(dent.IP + ":" + p) {
			dent.ServerList[s] = true
		}
	}
	if dent.ServerList["http"] || dent.ServerList["https"] {
		checkWebInfo(dent)
	}
}

func checkWebInfo(dent *discoverInfoEnt) {
	schemes := []string{}
	if dent.ServerList["http"] {
		schemes = append(schemes, "http")
	}
	if dent.ServerList["https"] {
		schemes = append(schemes, "https")
	}
	if len(schemes) == 0 {
		return
	}
	client := &http.Client{
		Timeout: time.Duration(datastore.DiscoverConf.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, scheme := range schemes {
		if Stop {
			return
		}
		url := fmt.Sprintf("%s://%s", scheme, dent.IP)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TWSNMP-FK/1.0)")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		rawBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
		rawStr := string(rawBytes)
		dent.HTTPServer = resp.Header.Get("Server")
		if realm := resp.Header.Get("WWW-Authenticate"); realm != "" {
			if dent.HTTPServer != "" {
				dent.HTTPServer += " " + realm
			} else {
				dent.HTTPServer = realm
			}
		}
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawStr)); err == nil {
			dent.HTTPTitle = strings.TrimSpace(doc.Find("title").First().Text())
		}
		cleanText := strings.Join(strings.Fields(rawStr), " ")
		if len(cleanText) > 8192 {
			cleanText = cleanText[:8192]
		}
		dent.HTTPBody = cleanText
		if dent.HTTPTitle != "" || dent.HTTPServer != "" || dent.HTTPBody != "" {
			break
		}
	}
}

func doTCPConnect(dst string) bool {
	timeout := time.Duration(datastore.DiscoverConf.Timeout) * time.Second
	if timeout <= 0 || timeout > time.Second {
		timeout = time.Second
	}
	conn, err := net.DialTimeout("tcp", dst, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func getMIBStringVal(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case []uint8:
		return string(v)
	case int, int64, uint, uint64:
		return fmt.Sprintf("%d", v)
	}
	return ""
}

