package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/ping"
	"github.com/twsnmp/twsnmpfk/wol"
)

func (a *App) GetNode(id string) datastore.NodeEnt {
	n := datastore.GetNode(id)
	if n == nil {
		return datastore.NodeEnt{}
	}
	return *n
}

// addNode add node
func (a *App) addNode(n datastore.NodeEnt) bool {
	if err := datastore.AddNode(&n); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    "ノードを追加しました",
	})
	return true
}

// UpdateNode update node
func (a *App) UpdateNode(nu datastore.NodeEnt) bool {
	n := datastore.GetNode(nu.ID)
	if n == nil {
		if nu.ID != "" {
			log.Printf("node not found id=%s", nu.ID)
		}
		nu.ID = ""
		return a.addNode(nu)
	}
	n.Name = nu.Name
	n.Descr = nu.Descr
	n.IP = nu.IP
	n.Icon = nu.Icon
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
	n.Password = nu.Password
	n.PublicKey = nu.PublicKey
	n.URL = nu.URL
	n.AddrMode = nu.AddrMode
	n.AutoAck = nu.AutoAck
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    "ノードを更新しました",
	})
	return true
}

// DeleteNodes delete node
func (a *App) DeleteNodes(ids []string) {
	for _, id := range ids {
		n := datastore.GetNode(id)
		if n != nil {
			datastore.DeleteNode(id)
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeName: n.Name,
				NodeID:   n.ID,
				Event:    "ノードを削除しました",
			})
		}
	}
}

// CopyNode : copy ndde
func (a *App) CopyNode(id string) bool {
	ns := datastore.GetNode(id)
	if ns == nil {
		return false
	}
	n := datastore.NodeEnt{}
	n.ID = ""
	n.X = ns.X + 100
	n.Y = ns.Y
	n.Name = ns.Name + "-Copy"
	n.Descr = ns.Descr
	n.IP = ns.IP
	n.Icon = ns.Icon
	n.SnmpMode = ns.SnmpMode
	n.Community = ns.Community
	n.User = ns.User
	n.Password = ns.Password
	n.PublicKey = ns.PublicKey
	n.URL = ns.URL
	n.AddrMode = ns.AddrMode
	n.AutoAck = ns.AutoAck
	if !a.addNode(n) {
		log.Printf("fail to copy node id='%s'", id)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    "ノードをコピーしました",
	})
	return true
}

// WakeOnLan : send wake on lan packet
func (a *App) WakeOnLan(id string) bool {
	n := datastore.GetNode(id)
	if n == nil {
		log.Printf("WakeOnLan node not found")
		return false
	}
	mac := strings.SplitN(n.MAC, "(", 2)
	if len(mac) < 1 || mac[0] == "" {
		log.Printf("WakeOnLan no MAC")
		return false
	}
	if err := wol.SendWakeOnLanPacket(mac[0]); err != nil {
		log.Printf("WakeOnLan node not found")
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		NodeID:   n.ID,
		Event:    fmt.Sprintf("%sにWake ON LANパケットを送信しました", n.MAC),
	})
	return true
}

// PingReq はPingのリクエストです。
type PingReq struct {
	IP   string `json:"IP"`
	Size int    `json:"Size"`
	TTL  int    `json:"TTL"`
}

// PingRes はPingの実行結果です。
type PingRes struct {
	Stat      int    `json:"Stat"`
	TimeStamp int64  `json:"TimeStamp"`
	Time      int64  `json:"Time"`
	Size      int    `json:"Size"`
	SendTTL   int    `json:"SendTTL"`
	RecvTTL   int    `json:"RecvTTL"`
	RecvSrc   string `json:"RecvSrc"`
	Loc       string `json:"Loc"`
}

// Ping を実行します。
func (a *App) Ping(req PingReq) PingRes {
	ipreg := regexp.MustCompile(`^[0-9.]+$`)
	if !ipreg.MatchString(req.IP) {
		if ips, err := net.LookupIP(req.IP); err == nil {
			for _, ip := range ips {
				if ip.IsGlobalUnicast() {
					s := ip.To4().String()
					if ipreg.MatchString(s) {
						req.IP = s
						break
					}
				}
			}
		}
	}
	res := PingRes{}
	pe := ping.DoPing(req.IP, 2, 0, req.Size, req.TTL)
	res.Stat = int(pe.Stat)
	res.TimeStamp = time.Now().Unix()
	res.Time = pe.Time
	res.Size = pe.Size
	res.RecvSrc = pe.RecvSrc
	res.RecvTTL = pe.RecvTTL
	res.SendTTL = req.TTL
	if pe.RecvSrc != "" {
		res.Loc = datastore.GetLoc(pe.RecvSrc)
	}
	return res
}
