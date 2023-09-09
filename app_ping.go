package main

import (
	"net"
	"regexp"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/ping"
)

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
