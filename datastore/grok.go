package datastore

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type GrokEnt struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Descr string `json:"Decr"`
	Pat   string `json:"Pat"`
	Ok    string `json:"Ok"`
}

var defGrockList = []GrokEnt{
	{
		ID:    "EPSLOGIN",
		Name:  "EPSの認証",
		Descr: "EPSで認証した時のユーザーID、クライアントを抽出",
		Pat:   `Login %{NOTSPACE:stat}: \[(host/)*%{USER:user}\].+cli %{MAC:client}`,
		Ok:    "OK",
	},
	{
		ID:    "FZLOGIN",
		Name:  "FileZenログイン",
		Descr: "FileZenにログインした時のユーザーID、クラアンとを抽出",
		Pat:   `FileZen: %{IP:client} %{USER:user} "Authentication %{NOTSPACE:stat}`,
		Ok:    "succeeded.",
	},
	{
		ID:    "NAOSLOGIN",
		Name:  "NAOSログイン",
		Descr: "NAOSのログイン",
		Pat:   `Login %{NOTSPACE:stat}: \[.+\] %{USER:user}`,
		Ok:    "Success",
	},
	{
		ID:    "DEVICE",
		Name:  "デバイス情報(ip)",
		Descr: "デバイス情報を取得mac=が先のケース",
		Pat:   `mac=%{MAC:mac}.+ip=%{IP:ip}`,
	},
	{
		ID:    "DEVICER",
		Name:  "デバイス情報(mac)",
		Descr: "デバイス情報を取得ip=が先のケース",
		Pat:   `ip=%{IP:ip}.+mac=%{MAC:mac}`,
	},
	{
		ID:    "WELFFLOW",
		Name:  "WELFフロー",
		Descr: "WELF形式のFWのログからフロー情報を取得",
		Pat:   `src=%{IP:src}:%{BASE10NUM:sport}:.+dst=%{IP:dst}:%{BASE10NUM:dport}:.+proto=%{WORD:prot}.+sent=%{BASE10NUM:sent}.+rcvd=%{BASE10NUM:rcvd}.+spkt=%{BASE10NUM:spkt}.+rpkt=%{BASE10NUM:rpkt}`,
	},
	{
		ID:    "OPENWEATHER",
		Name:  "気象情報",
		Descr: "Open Weatherのサイトから気象データを取得",
		Pat:   `"weather":.+"main":\s*"%{WORD:weather}".+"main":.+"temp":\s*%{BASE10NUM:temp}.+"feels_like":\s*%{BASE10NUM:feels_like}.+"temp_min":\s*%{BASE10NUM:temp_min}.+"temp_max":\s*%{BASE10NUM:temp_max}.+"pressure":\s*%{BASE10NUM:pressure}.+"humidity":\s*%{BASE10NUM:humidity}.+"wind":\s*{"speed":\s*%{BASE10NUM:wind}`,
	},
	{
		ID:    "UPTIME",
		Name:  "負荷(uptime)",
		Descr: "uptimeコマンドの出力から負荷を取得",
		Pat:   `load average: %{BASE10NUM:load1m}, %{BASE10NUM:load5m}, %{BASE10NUM:load15m}`,
	},
	{
		ID:    "SSHLOGIN",
		Name:  "SSHのログイン",
		Descr: "SSHでログインした時のユーザーID、クライアントを取得",
		Pat:   `%{NOTSPACE:stat} (password|publickey) for( invalid user | )%{USER:user} from %{IP:client}`,
		Ok:    "Accepted",
	},
	{
		ID:    "TWPCAP_STATS",
		Name:  "TWPCAPの統計情報",
		Descr: "TWPCAPで処理したパケット数などの統計情報を抽出",
		Pat:   `type=Stats,total=%{BASE10NUM:total},count=%{BASE10NUM:count},ps=%{BASE10NUM:ps}`,
	},
	{
		ID:    "TWPCAP_IPTOMAC",
		Name:  "TWPCAPのIPとMACアドレス",
		Descr: "TWPCAPで収集したIPとMACアドレスの情報を抽出",
		Pat:   `type=IPToMAC,ip=%{IP:ip},mac=%{MAC:mac},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},dhcp=%{BASE10NUM:dhcp}`,
	},
	{
		ID:    "TWPCAP_DNS",
		Name:  "TWPCAPのDNS問い合わせ",
		Descr: "TWPCAPで収集したDNSの問い合わせ情報を抽出",
		Pat:   `type=DNS,sv=%{IP:sv},DNSType=%{WORD:dnsType},Name=%{IPORHOST:name},count=%{BASE10NUM:count},change=%{BASE10NUM:chnage},lcl=%{IP:lastIP},lMAC=%{MAC:lastMAC}`,
	},
	{
		ID:    "TWPCAP_DHCP",
		Name:  "TWPCAPのDHCPサーバー情報",
		Descr: "TWPCAPで収集したDHCPサーバー情報を抽出",
		Pat:   `type=DHCP,sv=%{IP:sv},count=%{BASE10NUM:count},offer=%{BASE10NUM:offer},ack=%{BASE10NUM:ack},nak=%{BASE10NUM:nak}`,
	},
	{
		ID:    "TWPCAP_NTP",
		Name:  "TWPCAPのNTPサーバー情報",
		Descr: "TWPCAPで収集したNTPサーバー情報を抽出",
		Pat:   `type=NTP,sv=%{IP:sv},count=%{BASE10NUM:count},change=%{BASE10NUM:change},lcl=%{IP:client},version=%{BASE10NUM:version},stratum=%{BASE10NUM:stratum},refid=%{WORD:refid}`,
	},
	{
		ID:    "TWPCAP_RADIUS",
		Name:  "TWPCAPのRADIUS通信情報",
		Descr: "TWPCAPで収集したRADIUS通信情報を抽出",
		Pat:   `type=RADIUS,cl=%{IP:client},sv=%{IP:server},count=%{BASE10NUM:count},req=%{BASE10NUM:request},accept=%{BASE10NUM:accept},reject=%{BASE10NUM:reject},challenge=%{BASE10NUM:challenge}`,
	},
	{
		ID:    "TWPCAP_TLSFlow",
		Name:  "TWPCAPのTLS通信情報",
		Descr: "TWPCAPで収集したTLS通信情報を抽出",
		Pat:   `type=TLSFlow,cl=%{IP:client},sv=%{IP:server},serv=%{WORD:service},count=%{BASE10NUM:count},handshake=%{BASE10NUM:handshake},alert=%{BASE10NUM:alert},minver=%{DATA:minver},maxver=%{DATA:maxver},cipher=%{DATA:cipher},ft=`,
	},
}

var grokMap = make(map[string]*GrokEnt)

func loadGrokMap() {
	loadGrokFromDB()
	if len(grokMap) < 1 {
		LoadDefGrokEnt()
	}
}

func loadGrokFromDB() error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var g GrokEnt
				if err := json.Unmarshal(v, &g); err == nil {
					grokMap[g.ID] = &g
				}
				return nil
			})
		}
		return nil
	})
}

func GetGrokEnt(id string) *GrokEnt {
	if r, ok := grokMap[id]; ok {
		return r
	}
	return nil
}

// UpdateGrokEnt : Add or Replace GrokEnt
func UpdateGrokEnt(g *GrokEnt) error {
	s, err := json.Marshal(g)
	if err != nil {
		return err
	}
	st := time.Now()
	err = db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		return b.Put([]byte(g.ID), s)
	})
	if err != nil {
		return err
	}
	grokMap[g.ID] = g
	log.Printf("UpdateGrokEnt dur=%v", time.Since(st))
	return nil
}

func DeleteGrokEnt(id string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	err := db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("grok"))
		return b.Delete([]byte(id))
	})
	delete(grokMap, id)
	if err != nil {
		return err
	}
	log.Printf("DeleteGrokEnt dur=%v", time.Since(st))
	return nil
}

func ForEachGrokEnt(f func(*GrokEnt) bool) {
	for _, g := range grokMap {
		if !f(g) {
			break
		}
	}
}

func LoadDefGrokEnt() {
	for i := range defGrockList {
		UpdateGrokEnt(&defGrockList[i])
	}
}
