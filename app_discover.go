package main

import (
	"log"
	"net"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/discover"
)

// GetDiscoverConf retunrs discover config
func (a *App) GetDiscoverConf() datastore.DiscoverConfEnt {
	return datastore.DiscoverConf
}

// GetDiscoverAddressRange retunrs discover address
func (a *App) GetDiscoverAddressRange() []string {
	ret := []string{}
	ifs, err := net.Interfaces()
	if err != nil {
		log.Printf("GetDiscoverAddressRange err=%v", err)
		return ret
	}
	for _, i := range ifs {
		if (i.Flags&net.FlagLoopback) == net.FlagLoopback ||
			(i.Flags&net.FlagUp) != net.FlagUp ||
			(i.Flags&net.FlagPointToPoint) == net.FlagPointToPoint ||
			len(i.HardwareAddr) != 6 ||
			i.HardwareAddr[0]&0x02 == 0x02 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			cidr := a.String()
			ipTmp, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			ip := ipTmp.To4()
			if ip == nil {
				continue
			}
			start := ip.Mask(ipnet.Mask)
			mask := ipnet.Mask
			end := net.IP(make([]byte, 4))
			for i := range ip {
				end[i] = ip[i] | ^mask[i]
			}
			end[3] -= 1
			ret = append(ret, start.String())
			ret = append(ret, end.String())
		}
	}
	return ret
}

// GetDiscoverStats restunrs discover stats
func (a *App) GetDiscoverStats() discover.DiscoverStat {
	return discover.Stat
}

// StartDiscover start discover
func (a *App) StartDiscover(dc datastore.DiscoverConfEnt) bool {
	datastore.DiscoverConf = dc
	if err := datastore.SaveDiscoverConf(); err != nil {
		log.Println(err)
		return false
	}
	if err := discover.StartDiscover(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

// StopDiscover stop discover
func (a *App) StopDiscover() {
	discover.StopDiscover()
}
