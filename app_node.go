package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
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
		Event:    i18n.Trans("Add Node"),
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
	n.Image = nu.Image
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
	n.SSHUser = nu.SSHUser
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
		Event:    i18n.Trans("Update Node"),
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
				Event:    i18n.Trans("Delete Node"),
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
	n.Image = ns.Image
	n.SnmpMode = ns.SnmpMode
	n.Community = ns.Community
	n.User = ns.User
	n.SSHUser = ns.SSHUser
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
		Event:    i18n.Trans("Copy Node"),
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
		Event:    fmt.Sprintf(i18n.Trans("Send Wake on LAN Packet to %s"), n.MAC),
	})
	return true
}

// GetHostResource は、ノードからホストリソースMIBを取得して返します。
func (a *App) GetHostResource(id string) *backend.HostResourceEnt {
	n := datastore.GetNode(id)
	if n == nil {
		log.Printf("host resorce node not found id=%s", id)
		return nil
	}
	return backend.GetHostResource(n)
}
