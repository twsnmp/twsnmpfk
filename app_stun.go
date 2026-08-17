package main

import (
	"fmt"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/stun"
)

type StunInfoResult struct {
	IP        string        `json:"IP"`
	Port      int           `json:"Port"`
	Hostname  string        `json:"Hostname"`
	LocalIP   string        `json:"LocalIP"`
	LocalPort int           `json:"LocalPort"`
	RTT       string        `json:"RTT"`
	RTTNano   int64         `json:"RTTNano"`
	Location  string        `json:"Location"`
	Server    string        `json:"Server"`
	Protocol  string        `json:"Protocol"`
	Error     string        `json:"Error"`
	Entries   []AddrInfoEnt `json:"Entries"`
}

// GetStunInfo retrieves public mapped IP and connection details using STUN
func (a *App) GetStunInfo(server string, proto string, timeoutSec int) *StunInfoResult {
	if server == "" {
		server = stun.DefaultServer
	}
	network := "udp4"
	if proto == "ipv6" || proto == "udp6" {
		network = "udp6"
	}
	timeout := time.Duration(timeoutSec) * time.Second
	if timeoutSec <= 0 {
		timeout = stun.DefaultTimeout
	}

	res, err := stun.Query(server, network, timeout)
	if err != nil {
		return &StunInfoResult{
			Server:   server,
			Protocol: network,
			Error:    err.Error(),
			Entries: []AddrInfoEnt{
				{Level: "high", Title: i18n.Trans("Error"), Value: err.Error()},
				{Level: "info", Title: i18n.Trans("STUN Server"), Value: server},
				{Level: "info", Title: i18n.Trans("Protocol"), Value: network},
			},
		}
	}

	loc := datastore.GetLoc(res.IP)
	rttStr := fmt.Sprintf("%.2f ms", float64(res.RTT.Microseconds())/1000.0)

	entries := []AddrInfoEnt{
		{Level: "info", Title: i18n.Trans("Global IP Address"), Value: res.IP},
		{Level: "info", Title: i18n.Trans("Reverse DNS Host"), Value: func() string {
			if res.Hostname != "" {
				return res.Hostname
			}
			return i18n.Trans("Unknown")
		}()},
		{Level: "info", Title: i18n.Trans("Mapped Port"), Value: fmt.Sprintf("%d", res.Port)},
		{Level: "info", Title: i18n.Trans("Local Address"), Value: fmt.Sprintf("%s:%d", res.LocalIP, res.LocalPort)},
		{Level: "info", Title: i18n.Trans("Response Time"), Value: rttStr},
		{Level: "info", Title: i18n.Trans("STUN Server"), Value: res.Server},
		{Level: "info", Title: i18n.Trans("Protocol"), Value: res.Network},
		{Level: "info", Title: i18n.Trans("Location"), Value: loc},
	}

	return &StunInfoResult{
		IP:        res.IP,
		Port:      int(res.Port),
		Hostname:  res.Hostname,
		LocalIP:   res.LocalIP,
		LocalPort: int(res.LocalPort),
		RTT:       rttStr,
		RTTNano:   res.RTTNano,
		Location:  loc,
		Server:    res.Server,
		Protocol:  res.Network,
		Entries:   entries,
	}
}
