// Package discover : 自動発見
package discover

/* discover.go: 自動発見の処理
自動発見は、PINGを実行して、応答があるノードに関してSNMPの応答があるか確認する
*/

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/signalsciences/ipv4"
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
	IfMap       map[string]string
	ServerList  map[string]bool
	X           int
	Y           int
}

// StopDiscover : 自動発見を停止する
func StopDiscover() {
	for Stat.Running {
		Stop = true
		time.Sleep(time.Millisecond * 100)
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
	sem := make(chan bool, 256)
	go func() {
		for ; sip <= eip && !Stop; sip++ {
			sem <- true
			Stat.Sent++
			Stat.Now = time.Now().Unix()
			go func(ip uint32) {
				defer func() {
					<-sem
				}()
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
					r := &net.Resolver{}
					ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
					defer cancel()
					if names, err := r.LookupAddr(ctx, ipstr); err == nil && len(names) > 0 {
						dent.HostName = names[0]
					}
					getSnmpInfo(ipstr, &dent)
					if datastore.DiscoverConf.PortScan {
						checkServer(&dent)
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
						if _, ok := dent.ServerList["lldp"]; ok {
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

func getSnmpInfo(t string, dent *discoverInfoEnt) {
	agent := &gosnmp.GoSNMP{
		Target:    t,
		Port:      161,
		Transport: "udp",
		Community: datastore.MapConf.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(datastore.DiscoverConf.Timeout) * time.Second,
		Retries:   datastore.DiscoverConf.Retry,
		MaxOids:   gosnmp.MaxOids,
	}
	switch datastore.MapConf.SnmpMode {
	case "v3auth":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthNoPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
		}
	case "v3authpriv":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	case "v3authprivex":
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		agent.MsgFlags = gosnmp.AuthPriv
		agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 datastore.MapConf.SnmpUser,
			AuthenticationProtocol:   gosnmp.SHA256,
			AuthenticationPassphrase: datastore.MapConf.SnmpPassword,
			PrivacyProtocol:          gosnmp.AES256,
			PrivacyPassphrase:        datastore.MapConf.SnmpPassword,
		}
	}
	err := agent.Connect()
	if err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	defer agent.Conn.Close()
	oids := []string{datastore.MIBDB.NameToOID("sysName"), datastore.MIBDB.NameToOID("sysObjectID")}
	result, err := agent.GetNext(oids)
	if err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for _, variable := range result.Variables {
		if datastore.MIBDB.OIDToName(variable.Name) == "sysName.0" {
			dent.SysName = getMIBStringVal(variable.Value)
		} else if datastore.MIBDB.OIDToName(variable.Name) == "sysObjectID.0" {
			dent.SysObjectID = getMIBStringVal(variable.Value)
		}
	}
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
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) == 2 {
			dent.ServerList["lldp"] = true
		}
		return fmt.Errorf("checkend")
	})
}

func addFoundNode(dent *discoverInfoEnt) {
	funcList := []string{}
	n := datastore.NodeEnt{
		Name:  dent.HostName,
		IP:    dent.IP,
		Icon:  "desktop",
		X:     dent.X,
		Y:     dent.Y,
		Descr: fmt.Sprintf(i18n.Trans("Found at %s"), time.Now().Format("2006/01/02")),
	}
	if n.Name == "" {
		if dent.SysName != "" {
			n.Name = dent.SysName
		} else {
			n.Name = dent.IP
		}
	}
	if dent.SysObjectID != "" {
		n.SnmpMode = datastore.MapConf.SnmpMode
		n.User = datastore.MapConf.SnmpUser
		n.Password = datastore.MapConf.SnmpPassword
		n.Community = datastore.MapConf.Community
		n.Icon = "hdd"
		funcList = append(funcList, "snmp")
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
		Event:    i18n.Trans("Add by dicover"),
	})
	if datastore.DiscoverConf.AddNetwork && datastore.FindNetworkByIP(n.IP) == nil {
		if _, ok := dent.ServerList["lldp"]; ok {
			datastore.AddNetwork(&datastore.NetworkEnt{
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
				Descr:     time.Now().Format("2006/01/02") + "に発見",
			})
		}
	}
	if !datastore.DiscoverConf.AddPolling {
		return
	}
	addPolling(dent, &n)
}
func updateNode(n *datastore.NodeEnt, dent *discoverInfoEnt) {
	if n.Name == n.IP {
		if dent.SysName != "" {
			n.Name = dent.SysName
		}
	}
	if dent.SysObjectID != "" && n.User == "" && n.Community == "" {
		n.SnmpMode = datastore.MapConf.SnmpMode
		n.User = datastore.MapConf.SnmpUser
		n.Password = datastore.MapConf.SnmpPassword
		n.Community = datastore.MapConf.Community
		if n.Icon == "desktop" {
			n.Icon = "hdd"
			n.Descr += " / snmp対応"
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "discover",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    "自動発見により更新",
	})
	if datastore.DiscoverConf.AddNetwork && datastore.FindNetworkByIP(n.IP) == nil {
		if _, ok := dent.ServerList["lldp"]; ok {
			datastore.AddNetwork(&datastore.NetworkEnt{
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
				Descr:     time.Now().Format("2006/01/02") + "に発見",
			})
		}
	}
	if !datastore.DiscoverConf.AddPolling {
		return
	}
	addPolling(dent, n)

}

func addPolling(dent *discoverInfoEnt, n *datastore.NodeEnt) {
	p := &datastore.PollingEnt{
		NodeID:  n.ID,
		Name:    "PING",
		Type:    "ping",
		Level:   "low",
		State:   "unknown",
		PollInt: datastore.MapConf.PollInt,
		Timeout: datastore.MapConf.Timeout,
		Retry:   datastore.MapConf.Retry,
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
			NodeID:  n.ID,
			Name:    name,
			Type:    ptype,
			Mode:    mode,
			Params:  params,
			Level:   level,
			State:   "unknown",
			PollInt: datastore.MapConf.PollInt,
			Timeout: datastore.MapConf.Timeout,
			Retry:   datastore.MapConf.Retry,
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
		NodeID:  n.ID,
		Name:    "sysUptime",
		Type:    "snmp",
		Mode:    "sysUpTime",
		Level:   "off",
		State:   "unknown",
		PollInt: datastore.MapConf.PollInt,
		Timeout: datastore.MapConf.Timeout,
		Retry:   datastore.MapConf.Retry,
	}
	if err := datastore.AddPollingWithDupCheck(p); err != nil {
		log.Printf("discover err=%v", err)
		return
	}
	for i, name := range dent.IfMap {
		p = &datastore.PollingEnt{
			NodeID:  n.ID,
			Type:    "snmp",
			Name:    fmt.Sprintf("%s(%s)", name, i),
			Mode:    "ifOperStatus",
			Params:  i,
			Level:   "off",
			State:   "unknown",
			PollInt: datastore.MapConf.PollInt,
			Timeout: datastore.MapConf.Timeout,
			Retry:   datastore.MapConf.Retry,
		}
		if err := datastore.AddPollingWithDupCheck(p); err != nil {
			log.Printf("discover err=%v", err)
			return
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
		time.Sleep(time.Second)
		if doTCPConnect(dent.IP + ":" + p) {
			dent.ServerList[s] = true
		}
	}
}

func doTCPConnect(dst string) bool {
	conn, err := net.DialTimeout("tcp", dst, time.Duration(datastore.DiscoverConf.Timeout)*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
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
