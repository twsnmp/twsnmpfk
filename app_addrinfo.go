package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/twsnmp/rdap"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

type AddrInfoEnt struct {
	Title string `json:"Title"`
	Level string `json:"Level"`
	Value string `json:"Value"`
}

// GetAddressInfo retunrs address info
func (a *App) GetAddressInfo(addr string) []AddrInfoEnt {
	if addr == "" {
		return []AddrInfoEnt{}
	}
	if _, err := net.ParseMAC(addr); err == nil {
		return a.getMACInfo(addr)
	}
	if ip := net.ParseIP(addr); ip != nil {
		return a.getIPInfo(addr)
	}
	return a.getDomainInfo(addr)
}

func (a *App) getMACInfo(addr string) []AddrInfoEnt {
	mac := normMACAddr(addr)
	ret := []AddrInfoEnt{}
	ret = append(ret, AddrInfoEnt{Title: i18n.Trans("MAC Address"), Value: addr})
	if n := datastore.FindNodeFromMAC(mac); n != nil {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Node"), Value: n.Name})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("IP Address"), Value: n.IP})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Descr"), Value: n.Descr})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("Managed Node"), Value: i18n.Trans("No")})
	}
	ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Vendor"), Value: datastore.FindVendor(mac)})
	ip := findIPFromArp(mac)
	if ip == "" {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("ARP Watch"), Value: i18n.Trans("Not found")})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("ARP Watch"), Value: ip})
	}
	return ret
}

func normMACAddr(m string) string {
	if hw, err := net.ParseMAC(m); err == nil {
		m = strings.ToUpper(hw.String())
		return m
	}
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

func findIPFromArp(mac string) string {
	ip := ""
	datastore.ForEachArp(func(a *datastore.ArpEnt) bool {
		if a.MAC == mac {
			ip = a.IP
			return false
		}
		return true
	})
	return ip
}

func (a *App) getIPInfo(ip string) []AddrInfoEnt {
	ret := []AddrInfoEnt{}
	ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("IP Address"), Value: ip})
	if n := datastore.FindNodeFromIP(ip); n != nil {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Managed Node"), Value: n.Name})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("Managed Node"), Value: i18n.Trans("No")})
	}
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	if names, err := r.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
		for _, n := range names {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("DNS Host"), Value: n})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("DNS Host"), Value: i18n.Trans("Unknown")})
	}
	loc := datastore.GetLoc(ip)
	ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Location"), Value: loc})
	if strings.Contains(loc, "LOCAL,") {
		return ret
	}
	client := &rdap.Client{}
	ri, err := client.QueryIP(ip)
	if err != nil {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: "RDAP:error", Value: fmt.Sprintf("%v", err)})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:IP Version", Value: ri.IPVersion}) //IPバージョン
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Type", Value: ri.Type})            // 種類
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Handle", Value: ri.Handle})       //範囲
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Name", Value: ri.Name})            // 所有者
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Country", Value: ri.Country})      // 国
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Whois Server", Value: ri.Port43})  // Whoisの情報源
	}
	return ret
}

func (a *App) getDomainInfo(domain string) []AddrInfoEnt {
	ret := []AddrInfoEnt{}
	ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Domain"), Value: domain})
	r := &net.Resolver{}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	if ips, err := r.LookupHost(ctx, domain); err == nil && len(ips) > 0 {
		for _, ip := range ips {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("IP Address"), Value: ip})
			loc := datastore.GetLoc(ip)
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Location"), Value: loc})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("IP Address"), Value: i18n.Trans("Unknown")})
	}
	if list, err := r.LookupNS(ctx, domain); err == nil && len(list) > 0 {
		for _, e := range list {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Name server"), Value: e.Host})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("Name server"), Value: i18n.Trans("Unknown")})
	}
	if list, err := r.LookupMX(ctx, domain); err == nil && len(list) > 0 {
		for _, e := range list {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("Mail server"), Value: e.Host})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("Mail server"), Value: i18n.Trans("Unknown")})
	}
	if list, err := r.LookupTXT(ctx, domain); err == nil && len(list) > 0 {
		for _, e := range list {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("TXT record"), Value: e})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "warn", Title: i18n.Trans("TXT record"), Value: i18n.Trans("Unknown")})
	}
	if cname, list, err := r.LookupSRV(ctx, "https", "tcp", domain); err == nil && len(list) > 0 {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "SRV(CNAME)", Value: cname})
		for _, e := range list {
			ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("SRV record"), Value: fmt.Sprintf("%s:%d %d %d", e.Target, e.Port, e.Priority, e.Weight)})
		}
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: i18n.Trans("SRV record"), Value: i18n.Trans("Unknown")})
	}
	client := &rdap.Client{}
	ri, err := client.QueryDomain(domain)
	if err != nil {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:error", Value: fmt.Sprintf("%v", err)})
	} else {
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Handle", Value: ri.Handle})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:LDHName", Value: ri.LDHName})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:UnicodeName", Value: ri.UnicodeName})
		ret = append(ret, AddrInfoEnt{Level: "info", Title: "RDAP:Whois Server", Value: ri.Port43}) // Whoisの情報源
	}
	return ret
}
