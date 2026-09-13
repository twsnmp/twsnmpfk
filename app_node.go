package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/wol"
)

func (a *App) GetNode(id string) datastore.NodeEnt {
	n := datastore.GetNode(id)
	if n == nil {
		if strings.HasPrefix(id, "NET:") {
			nt := datastore.GetNetwork(id)
			if nt != nil {
				return datastore.NodeEnt{
					ID: id,
					IP: nt.IP,
				}
			}
		}
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
	if saved := datastore.GetNode(n.ID); saved != nil {
		logger.CheckNodeAddr(saved)
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
	n.SnmpPort = nu.SnmpPort
	n.Community = nu.Community
	n.User = nu.User
	n.SSHUser = nu.SSHUser
	n.Password = nu.Password
	n.GNMIUser = nu.GNMIUser
	n.GNMIPassword = nu.GNMIPassword
	n.GNMIPort = nu.GNMIPort
	n.GNMIEncoding = nu.GNMIEncoding
	n.PublicKey = nu.PublicKey
	n.URL = nu.URL
	n.AddrMode = nu.AddrMode
	n.MAC = nu.MAC
	n.AutoAck = nu.AutoAck
	logger.CheckNodeAddr(n)
	if err := datastore.UpdateNode(n); err != nil {
		log.Printf("UpdateNode save err=%v", err)
	}
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
	n.SnmpPort = ns.SnmpPort
	n.Community = ns.Community
	n.User = ns.User
	n.SSHUser = ns.SSHUser
	n.Password = ns.Password
	n.GNMIUser = ns.GNMIUser
	n.GNMIPassword = ns.GNMIPassword
	n.GNMIPort = ns.GNMIPort
	n.GNMIEncoding = ns.GNMIEncoding
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

// SaveNodeMemo saves a memo related to the node.
func (a *App) SaveNodeMemo(nodeID, memo string) bool {
	if err := datastore.SaveNodeMemo(nodeID, memo); err != nil {
		log.Printf("save node memo err=%v", err)
		return false
	}
	return true
}

// GetNodeMemo returns a memo related to the node.
func (a *App) GetNodeMemo(nodeID string) string {
	return datastore.GetNodeMemo(nodeID)
}

// DetectNodeType auto-detects the device category, OS, icon, and recommended sensor pollings for a node.
func (a *App) DetectNodeType(id string) *datastore.DetectResult {
	n := datastore.GetNode(id)
	if n == nil {
		return nil
	}
	input := &datastore.DetectInput{
		IP:     n.IP,
		Name:   n.Name,
		Vendor: n.Vendor,
	}
	if (input.Vendor == "" || input.Vendor == "Unknown") && n.MAC != "" {
		v := datastore.FindVendor(n.MAC)
		if v != "" && v != "Unknown" {
			input.Vendor = v
		}
	}
	if input.Vendor == "" || input.Vendor == "Unknown" {
		if arp := datastore.GetArpEnt(n.IP); arp != nil {
			if arp.Vendor != "" && arp.Vendor != "Unknown" {
				input.Vendor = arp.Vendor
			} else if arp.MAC != "" {
				v := datastore.FindVendor(arp.MAC)
				if v != "" && v != "Unknown" {
					input.Vendor = v
				}
			}
		}
	}
	if n.IP != "" {
		r := &net.Resolver{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()
		if names, err := r.LookupAddr(ctx, n.IP); err == nil && len(names) > 0 {
			input.HostName = names[0]
		}
	}

	agent := backend.GetSNMPAgent(n)
	if agent != nil {
		if err := agent.Connect(); err == nil {
			defer agent.Conn.Close()
			oids := []string{
				datastore.MIBDB.NameToOID("sysObjectID"),
				datastore.MIBDB.NameToOID("sysDescr"),
			}
			if res, err := agent.GetNext(oids); err == nil {
				for _, v := range res.Variables {
					name := datastore.MIBDB.OIDToName(v.Name)
					if strings.HasPrefix(name, "sysObjectID") {
						input.SysObjectID = fmt.Sprintf("%v", v.Value)
					} else if strings.HasPrefix(name, "sysDescr") {
						switch val := v.Value.(type) {
						case string:
							input.SysDescr = val
						case []byte:
							input.SysDescr = string(val)
						}
					}
				}
			}
		}
	}

	title, server, body := backend.FetchWebSignatures(n.IP, n.URL)
	input.HTTPTitle = title
	input.HTTPServer = server
	input.HTTPBody = body

	res := datastore.DetectNode(input)
	if res == nil {
		return nil
	}

	// Verify sensor pollings with agent if connected
	if agent != nil && len(res.SensorPollings) > 0 {
		var verified []datastore.SensorPollingDef
		for _, sp := range res.SensorPollings {
			if datastore.CheckSensorSupport(agent, sp.Params) {
				verified = append(verified, sp)
			}
		}
		res.SensorPollings = verified
	}

	return res
}

// ApplyNodeDetection applies auto-detected icon and sensor pollings to the node.
func (a *App) ApplyNodeDetection(id string, applyIcon bool, applyPolling bool) bool {
	res := a.DetectNodeType(id)
	if res == nil {
		return false
	}
	n := datastore.GetNode(id)
	if n == nil {
		return false
	}

	updated := false
	if applyIcon && res.Icon != "" {
		n.Icon = res.Icon
		if res.Name != "" && res.RuleID != "unknown" {
			if !strings.Contains(n.Descr, res.Name) {
				n.Descr += fmt.Sprintf(" [%s]", res.Name)
			}
		}
		if (n.Vendor == "" || n.Vendor == "Unknown") && n.MAC != "" {
			v := datastore.FindVendor(n.MAC)
			if v != "" && v != "Unknown" {
				n.Vendor = v
			}
		}
		updated = true
	}

	if applyPolling && len(res.SensorPollings) > 0 {
		for _, sp := range res.SensorPollings {
			level := sp.Level
			if level == "" {
				level = "low"
			}
			mode := sp.Mode
			if mode == "" {
				mode = "get"
			}
			p := &datastore.PollingEnt{
				NodeID:  n.ID,
				Name:    sp.Name,
				Type:    sp.Type,
				Mode:    mode,
				Params:  sp.Params,
				Script:  sp.Script,
				Level:   level,
				State:   "unknown",
				PollInt: datastore.MapConf.PollInt,
				Timeout: datastore.MapConf.Timeout,
				Retry:   datastore.MapConf.Retry,
			}
			if err := datastore.AddPollingWithDupCheck(p); err != nil {
				log.Printf("apply detection add polling err=%v", err)
			}
		}
	}

	if updated {
		if err := datastore.UpdateNode(n); err != nil {
			log.Printf("ApplyNodeDetection save err=%v", err)
		}
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:     "user",
			Level:    "info",
			NodeName: n.Name,
			NodeID:   n.ID,
			Event:    fmt.Sprintf(i18n.Trans("Auto detected as %s"), res.Name),
		})
	}
	return true
}

