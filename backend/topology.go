package backend

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

// NeighborLineEnt extends datastore.LineEnt with confidence level and reasoning
type NeighborLineEnt struct {
	datastore.LineEnt
	Confidence string `json:"Confidence"` // "strict" | "speculative"
	Reason     string `json:"Reason"`     // "LLDP" | "CDP" | "STP" | "FDB-Edge" | "FDB-Heuristic" | "ARP/Subnet" | "AI"
}

// FindTopologyForNetwork searches neighbor networks and connecting lines using LLDP, CDP, STP, FDB and ARP.
func FindTopologyForNetwork(n *datastore.NetworkEnt) (*FindNeighborNetworksAndLinesResp, error) {
	ret := &FindNeighborNetworksAndLinesResp{
		Networks: []datastore.NetworkEnt{},
		Lines:    []NeighborLineEnt{},
	}

	agent := getSNMPAgentForNetwork(n)
	if agent == nil {
		n.Error = "Invalid SNMP config"
		return ret, fmt.Errorf("invalid SNMP config")
	}
	err := agent.Connect()
	if err != nil {
		n.Error = fmt.Sprintf("SNMP connect err=%s", err)
		return ret, err
	}
	defer agent.Conn.Close()

	// Port map by index and ID for local network
	portByIndex := make(map[string]*datastore.PortEnt)
	portByID := make(map[string]*datastore.PortEnt)
	for i := range n.Ports {
		p := &n.Ports[i]
		portByIndex[p.Index] = p
		portByID[p.ID] = p
	}

	// 1. Build bridge port to ifIndex mapping
	portToIFIndexMap := make(map[int]int)
	_ = agent.Walk(datastore.MIBDB.NameToOID("dot1dBasePortIfIndex"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		if idx, err := strconv.Atoi(a[1]); err == nil {
			portToIFIndexMap[idx] = int(gosnmp.ToBigInt(variable.Value).Int64())
		}
		return nil
	})

	// 2. LLDP Discovery
	findLLDPTopology(agent, n, ret, portByIndex)

	// 3. CDP (CISCO-CDP-MIB) Discovery
	findCDPTopology(agent, n, ret, portByIndex)

	// 4. STP (BRIDGE-MIB Spanning Tree) Discovery
	findSTPTopology(agent, n, ret, portByIndex, portToIFIndexMap)

	// 5. ARP Table for IP <-> MAC resolution
	ipToMac, macToIP := getNetworkARPTable(agent)

	// 6. FDB Discovery (Edge port & Cascade heuristic)
	findFDBTopology(agent, n, ret, portByIndex, portToIFIndexMap, ipToMac, macToIP)

	// 7. ARP Direct Discovery (Router/switch port to node connections)
	findARPTopology(agent, n, ret, portByIndex)

	return ret, nil
}

// findLLDPTopology gathers LLDP neighbors and links
func findLLDPTopology(agent *gosnmp.GoSNMP, n *datastore.NetworkEnt, ret *FindNeighborNetworksAndLinesResp, portByIndex map[string]*datastore.PortEnt) {
	remoteMap := make(map[string]*datastore.NetworkEnt)

	_ = agent.Walk(datastore.MIBDB.NameToOID("lldpRemoteSystemsData"), func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		switch a[0] {
		case "lldpRemChassisId":
			remoteMap[a[1]] = &datastore.NetworkEnt{
				SystemID: datastore.GetMIBValueString(a[0], &variable, false),
			}
		case "lldpRemPortId":
			if rn, ok := remoteMap[a[1]]; ok {
				b := strings.Split(a[1], ".")
				if len(b) >= 2 {
					id := datastore.GetMIBValueString(a[0], &variable, false)
					rn.Ports = append(rn.Ports, datastore.PortEnt{
						ID:    id,
						Index: b[1],
						Name:  id,
						X:     len(n.Ports),
					})
				}
			}
		case "lldpRemSysName":
			if rn, ok := remoteMap[a[1]]; ok {
				rn.Name = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpRemSysDesc":
			if rn, ok := remoteMap[a[1]]; ok {
				rn.Descr = datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpRemSysCapEnabled":
			if rn, ok := remoteMap[a[1]]; ok {
				rn.Descr += " " + datastore.GetMIBValueString(a[0], &variable, false)
			}
		case "lldpRemManAddrIfId", "lldpRemManAddrOID", "lldpRemManAddrIfSubtype":
			// Extract IPv4 management address from OID suffix: .1.4.x.x.x.x
			b := strings.Split(a[1], ".")
			for i := 0; i+5 <= len(b); i++ {
				if b[i] == "1" && b[i+1] == "4" { // subtype IPv4(1), length 4
					ip := strings.Join(b[i+2:i+6], ".")
					prefix := strings.Join(b[:i], ".")
					if rn, ok := remoteMap[prefix]; ok && rn.IP == "" {
						rn.IP = ip
					}
					break
				}
			}
		}
		return nil
	})

	for _, rn := range remoteMap {
		rnr := datastore.FindNetwork(rn.SystemID, rn.IP)
		if rnr == nil {
			// Check if this LLDP neighbor matches a regular node (e.g. Linux server, AP, router)
			node := datastore.FindNodeFromIP(rn.IP)
			if node == nil {
				node = datastore.FindNodeFromMAC(rn.SystemID)
			}
			if node == nil && rn.Name != "" {
				datastore.ForEachNodes(func(nd *datastore.NodeEnt) bool {
					if strings.EqualFold(nd.Name, rn.Name) {
						node = nd
						return false
					}
					return true
				})
			}

			if node != nil {
				// Matched a regular node via LLDP!
				for _, frp := range rn.Ports {
					if lp, ok := portByIndex[frp.Index]; ok {
						pid := findBestPollingIDForNode(node, frp.Index)
						l := NeighborLineEnt{
							LineEnt: datastore.LineEnt{
								NodeID1:    fmt.Sprintf("NET:%s", n.ID),
								PollingID1: lp.ID,
								NodeID2:    node.ID,
								PollingID2: pid,
								Width:      2,
								Info:       "LLDP",
							},
							Confidence: "strict",
							Reason:     "LLDP",
						}
						if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, false) {
							ret.Lines = append(ret.Lines, l)
						}
					}
				}
			} else {
				// Unregistered Network switch/router
				rn.SnmpMode = n.SnmpMode
				rn.Community = n.Community
				rn.Password = n.Password
				rn.User = n.User
				rn.HPorts = n.HPorts
				rn.Ports = []datastore.PortEnt{}
				rn.Y = n.Y + n.H
				rn.X = n.X
				ret.Networks = append(ret.Networks, *rn)
			}
		} else {
			// Registered Network switch
			for _, rp := range rnr.Ports {
				for _, frp := range rn.Ports {
					if frp.ID == rp.ID || frp.Name == rp.Name || frp.Index == rp.Index {
						if lp, ok := portByIndex[frp.Index]; ok {
							l := NeighborLineEnt{
								LineEnt: datastore.LineEnt{
									NodeID1:    fmt.Sprintf("NET:%s", n.ID),
									PollingID1: lp.ID,
									NodeID2:    fmt.Sprintf("NET:%s", rnr.ID),
									PollingID2: rp.ID,
									Width:      2,
									Info:       "LLDP",
								},
								Confidence: "strict",
								Reason:     "LLDP",
							}
							if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, true) {
								ret.Lines = append(ret.Lines, l)
							}
						}
					}
				}
			}
		}
	}
}

// findCDPTopology gathers Cisco CDP neighbors and links
func findCDPTopology(agent *gosnmp.GoSNMP, n *datastore.NetworkEnt, ret *FindNeighborNetworksAndLinesResp, portByIndex map[string]*datastore.PortEnt) {
	type cdpNeighbor struct {
		LocalIfIndex string
		DeviceID     string
		Address      string
		DevicePort   string
		Platform     string
	}
	cdpNeighbors := make(map[string]*cdpNeighbor)

	cdpOID := datastore.MIBDB.NameToOID("cdpCacheEntry")
	_ = agent.Walk(cdpOID, func(variable gosnmp.SnmpPDU) error {
		name := datastore.MIBDB.OIDToName(variable.Name)
		a := strings.SplitN(name, ".", 2)
		if len(a) != 2 {
			return nil
		}
		parts := strings.Split(a[1], ".")
		if len(parts) < 2 {
			return nil
		}
		ifIndex := parts[0]
		key := a[1]
		if _, ok := cdpNeighbors[key]; !ok {
			cdpNeighbors[key] = &cdpNeighbor{LocalIfIndex: ifIndex}
		}
		nb := cdpNeighbors[key]

		switch a[0] {
		case "cdpCacheDeviceId":
			nb.DeviceID = datastore.GetMIBValueString(a[0], &variable, false)
		case "cdpCacheDevicePort":
			nb.DevicePort = datastore.GetMIBValueString(a[0], &variable, false)
		case "cdpCachePlatform":
			nb.Platform = datastore.GetMIBValueString(a[0], &variable, false)
		case "cdpCacheAddress":
			switch val := variable.Value.(type) {
			case []byte:
				if len(val) == 4 {
					nb.Address = net.IP(val).String()
				}
			case string:
				if len(val) == 4 {
					nb.Address = net.IP([]byte(val)).String()
				}
			}
		}
		return nil
	})

	for _, cdp := range cdpNeighbors {
		if cdp.DeviceID == "" && cdp.Address == "" {
			continue
		}
		lp, ok := portByIndex[cdp.LocalIfIndex]
		if !ok {
			continue
		}

		// Try to match network node first
		rnr := datastore.FindNetwork("", cdp.Address)
		if rnr == nil && cdp.DeviceID != "" {
			datastore.ForEachNetworks(func(netEnt *datastore.NetworkEnt) bool {
				if strings.EqualFold(netEnt.Name, cdp.DeviceID) {
					rnr = netEnt
					return false
				}
				return true
			})
		}

		if rnr != nil {
			rPortID := ""
			for _, rp := range rnr.Ports {
				if strings.EqualFold(rp.Name, cdp.DevicePort) || rp.Index == cdp.DevicePort {
					rPortID = rp.ID
					break
				}
			}
			l := NeighborLineEnt{
				LineEnt: datastore.LineEnt{
					NodeID1:    fmt.Sprintf("NET:%s", n.ID),
					PollingID1: lp.ID,
					NodeID2:    fmt.Sprintf("NET:%s", rnr.ID),
					PollingID2: rPortID,
					Width:      2,
					Info:       "CDP",
				},
				Confidence: "strict",
				Reason:     "CDP",
			}
			if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, true) {
				ret.Lines = append(ret.Lines, l)
			}
			continue
		}

		// Try to match regular node (e.g. Cisco AP, server, IP Phone)
		var node *datastore.NodeEnt
		if cdp.Address != "" {
			node = datastore.FindNodeFromIP(cdp.Address)
		}
		if node == nil && cdp.DeviceID != "" {
			datastore.ForEachNodes(func(nd *datastore.NodeEnt) bool {
				if strings.EqualFold(nd.Name, cdp.DeviceID) {
					node = nd
					return false
				}
				return true
			})
		}
		if node != nil {
			pid := findBestPollingIDForNode(node, cdp.LocalIfIndex)
			l := NeighborLineEnt{
				LineEnt: datastore.LineEnt{
					NodeID1:    fmt.Sprintf("NET:%s", n.ID),
					PollingID1: lp.ID,
					NodeID2:    node.ID,
					PollingID2: pid,
					Width:      2,
					Info:       "CDP",
				},
				Confidence: "strict",
				Reason:     "CDP",
			}
			if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, false) {
				ret.Lines = append(ret.Lines, l)
			}
		}
	}
}

// findSTPTopology gathers Spanning Tree designated bridge links
func findSTPTopology(agent *gosnmp.GoSNMP, n *datastore.NetworkEnt, ret *FindNeighborNetworksAndLinesResp, portByIndex map[string]*datastore.PortEnt, portToIFIndexMap map[int]int) {
	stpOID := datastore.MIBDB.NameToOID("dot1dStpPortTable")
	if stpOID == "" {
		return
	}

	_ = agent.Walk(stpOID, func(variable gosnmp.SnmpPDU) error {
		name := datastore.MIBDB.OIDToName(variable.Name)
		a := strings.SplitN(name, ".", 2)
		if len(a) != 2 {
			return nil
		}
		if a[0] != "dot1dStpPortDesignatedBridge" {
			return nil
		}

		portNum, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		ifIndex := portNum
		if mapped, ok := portToIFIndexMap[portNum]; ok {
			ifIndex = mapped
		}
		lp, ok := portByIndex[fmt.Sprintf("%d", ifIndex)]
		if !ok {
			return nil
		}

		// Designated bridge is 8 bytes: 2 bytes priority + 6 bytes MAC address
		var mac string
		switch val := variable.Value.(type) {
		case []byte:
			if len(val) >= 8 {
				mac = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", val[2], val[3], val[4], val[5], val[6], val[7])
			} else if len(val) == 6 {
				mac = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", val[0], val[1], val[2], val[3], val[4], val[5])
			}
		case string:
			b := []byte(val)
			if len(b) >= 8 {
				mac = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", b[2], b[3], b[4], b[5], b[6], b[7])
			}
		}

		if mac == "" || strings.EqualFold(mac, n.SystemID) {
			return nil
		}

		// Find opposing network switch
		rnr := datastore.FindNetwork(mac, "")
		if rnr != nil && rnr.ID != n.ID {
			l := NeighborLineEnt{
				LineEnt: datastore.LineEnt{
					NodeID1:    fmt.Sprintf("NET:%s", n.ID),
					PollingID1: lp.ID,
					NodeID2:    fmt.Sprintf("NET:%s", rnr.ID),
					PollingID2: "",
					Width:      2,
					Info:       "STP",
				},
				Confidence: "strict",
				Reason:     "STP",
			}
			if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, true) {
				ret.Lines = append(ret.Lines, l)
			}
		}
		return nil
	})
}

// findFDBTopology analyzes FDB (Forwarding Database) with edge port distinction and cascade heuristics
func findFDBTopology(agent *gosnmp.GoSNMP, n *datastore.NetworkEnt, ret *FindNeighborNetworksAndLinesResp, portByIndex map[string]*datastore.PortEnt, portToIFIndexMap map[int]int, ipToMac, macToIP map[string]string) {
	fdbList := getEnhancedFDB(agent, portToIFIndexMap)

	// Count MACs learned per port
	portMacCount := make(map[int]int)
	for _, e := range fdbList {
		portMacCount[e.IfIndex]++
	}

	// Identify switch/router MACs to avoid treating them as simple endpoints
	knownSwitchMACs := make(map[string]bool)
	datastore.ForEachNetworks(func(netEnt *datastore.NetworkEnt) bool {
		if netEnt.SystemID != "" {
			knownSwitchMACs[strings.ToUpper(netEnt.SystemID)] = true
		}
		return true
	})

	for _, e := range fdbList {
		normMAC := strings.ToUpper(e.MAC)
		if knownSwitchMACs[normMAC] {
			continue // Handled by LLDP/CDP/STP
		}

		// Find node matching this MAC or corresponding IP
		node := datastore.FindNodeFromMAC(e.MAC)
		if node == nil {
			if ip, ok := macToIP[e.MAC]; ok {
				node = datastore.FindNodeFromIP(ip)
			}
		}
		if node == nil {
			continue
		}

		idxStr := fmt.Sprintf("%d", e.IfIndex)
		lp, ok := portByIndex[idxStr]
		if !ok {
			continue
		}

		macCount := portMacCount[e.IfIndex]
		confidence := "strict"
		reason := "FDB-Edge"

		if macCount > 2 {
			confidence = "speculative"
			reason = fmt.Sprintf("FDB-Heuristic (%d MACs)", macCount)
		}

		pid := findBestPollingIDForNode(node, idxStr)
		l := NeighborLineEnt{
			LineEnt: datastore.LineEnt{
				NodeID1:    fmt.Sprintf("NET:%s", n.ID),
				PollingID1: lp.ID,
				NodeID2:    node.ID,
				PollingID2: pid,
				Width:      2,
				Info:       reason,
			},
			Confidence: confidence,
			Reason:     reason,
		}

		if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, false) {
			ret.Lines = append(ret.Lines, l)
		}
	}
}

// findARPTopology analyzes ARP table to link router/switch ports to nodes
func findARPTopology(agent *gosnmp.GoSNMP, n *datastore.NetworkEnt, ret *FindNeighborNetworksAndLinesResp, portByIndex map[string]*datastore.PortEnt) {
	type arpEntry struct {
		IfIndex string
		IP      string
		MAC     string
	}
	var arpEntries []arpEntry
	portArpCount := make(map[string]int)

	arpOID := datastore.MIBDB.NameToOID("ipNetToMediaPhysAddress")
	err := agent.Walk(arpOID, func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		parts := strings.Split(a[1], ".")
		if len(parts) >= 5 {
			ifIndex := parts[0]
			ip := strings.Join(parts[len(parts)-4:], ".")
			mac := datastore.GetMIBValueString(a[0], &variable, false)
			if ip != "" {
				arpEntries = append(arpEntries, arpEntry{
					IfIndex: ifIndex,
					IP:      ip,
					MAC:     mac,
				})
				portArpCount[ifIndex]++
			}
		}
		return nil
	})
	if err != nil || len(arpEntries) == 0 {
		_ = agent.Walk(datastore.MIBDB.NameToOID("atPhysAddress"), func(variable gosnmp.SnmpPDU) error {
			a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
			if len(a) != 2 {
				return nil
			}
			parts := strings.Split(a[1], ".")
			if len(parts) >= 5 {
				ifIndex := parts[0]
				ip := strings.Join(parts[len(parts)-4:], ".")
				mac := datastore.GetMIBValueString(a[0], &variable, false)
				if ip != "" {
					arpEntries = append(arpEntries, arpEntry{
						IfIndex: ifIndex,
						IP:      ip,
						MAC:     mac,
					})
					portArpCount[ifIndex]++
				}
			}
			return nil
		})
	}

	for _, e := range arpEntries {
		node := datastore.FindNodeFromIP(e.IP)
		if node == nil && e.MAC != "" {
			node = datastore.FindNodeFromMAC(e.MAC)
		}
		if node == nil {
			continue
		}
		lp, ok := portByIndex[e.IfIndex]
		if !ok {
			continue
		}

		confidence := "strict"
		reason := "ARP"
		if portArpCount[e.IfIndex] > 2 {
			confidence = "speculative"
			reason = fmt.Sprintf("ARP-Multi (%d)", portArpCount[e.IfIndex])
		}

		pid := findBestPollingIDForNode(node, e.IfIndex)
		l := NeighborLineEnt{
			LineEnt: datastore.LineEnt{
				NodeID1:    fmt.Sprintf("NET:%s", n.ID),
				PollingID1: lp.ID,
				NodeID2:    node.ID,
				PollingID2: pid,
				Width:      2,
				Info:       reason,
			},
			Confidence: confidence,
			Reason:     reason,
		}
		if !hasNeighborLine(ret.Lines, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, false) {
			ret.Lines = append(ret.Lines, l)
		}
	}
}

// getEnhancedFDB reads dot1qTpFdbPort with fallback to dot1dTpFdbPort
func getEnhancedFDB(agent *gosnmp.GoSNMP, portToIFIndexMap map[int]int) []*FDBTableEnt {
	ret := []*FDBTableEnt{}

	// Try Q-BRIDGE-MIB dot1qTpFdbPort
	dot1qOID := datastore.MIBDB.NameToOID("dot1qTpFdbPort")
	_ = agent.Walk(dot1qOID, func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
		if len(a) != 1+1+6 {
			return nil
		}
		vlan, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		mac, err := indexToMacAddress(a[2:])
		if err != nil {
			return nil
		}
		port := int(gosnmp.ToBigInt(variable.Value).Int64())
		if port == 0 {
			return nil
		}
		ifIndex := port
		if mapped, ok := portToIFIndexMap[port]; ok {
			ifIndex = mapped
		}

		nodeName := ""
		if nn := datastore.FindNodeFromMAC(mac); nn != nil {
			nodeName = nn.Name
		}
		ret = append(ret, &FDBTableEnt{
			MAC:     mac,
			VLanID:  vlan,
			Port:    port,
			IfIndex: ifIndex,
			Node:    nodeName,
			Vendor:  datastore.FindVendor(mac),
		})
		return nil
	})

	// Fallback to standard BRIDGE-MIB dot1dTpFdbPort if Q-BRIDGE was empty
	if len(ret) == 0 {
		dot1dOID := datastore.MIBDB.NameToOID("dot1dTpFdbPort")
		_ = agent.Walk(dot1dOID, func(variable gosnmp.SnmpPDU) error {
			a := strings.Split(datastore.MIBDB.OIDToName(variable.Name), ".")
			if len(a) != 1+6 {
				return nil
			}
			mac, err := indexToMacAddress(a[1:])
			if err != nil {
				return nil
			}
			port := int(gosnmp.ToBigInt(variable.Value).Int64())
			if port == 0 {
				return nil
			}
			ifIndex := port
			if mapped, ok := portToIFIndexMap[port]; ok {
				ifIndex = mapped
			}

			nodeName := ""
			if nn := datastore.FindNodeFromMAC(mac); nn != nil {
				nodeName = nn.Name
			}
			ret = append(ret, &FDBTableEnt{
				MAC:     mac,
				VLanID:  0,
				Port:    port,
				IfIndex: ifIndex,
				Node:    nodeName,
				Vendor:  datastore.FindVendor(mac),
			})
			return nil
		})
	}

	return ret
}

// getNetworkARPTable retrieves ARP mappings from router/switch SNMP agent
func getNetworkARPTable(agent *gosnmp.GoSNMP) (map[string]string, map[string]string) {
	ipToMac := make(map[string]string)
	macToIP := make(map[string]string)

	arpOID := datastore.MIBDB.NameToOID("ipNetToMediaPhysAddress")
	err := agent.Walk(arpOID, func(variable gosnmp.SnmpPDU) error {
		a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
		if len(a) != 2 {
			return nil
		}
		parts := strings.Split(a[1], ".")
		if len(parts) >= 5 {
			ip := strings.Join(parts[len(parts)-4:], ".")
			mac := datastore.GetMIBValueString(a[0], &variable, false)
			if ip != "" && mac != "" {
				ipToMac[ip] = mac
				macToIP[mac] = ip
			}
		}
		return nil
	})
	if err != nil || len(ipToMac) == 0 {
		_ = agent.Walk(datastore.MIBDB.NameToOID("atPhysAddress"), func(variable gosnmp.SnmpPDU) error {
			a := strings.SplitN(datastore.MIBDB.OIDToName(variable.Name), ".", 2)
			if len(a) != 2 {
				return nil
			}
			parts := strings.Split(a[1], ".")
			if len(parts) >= 5 {
				ip := strings.Join(parts[len(parts)-4:], ".")
				mac := datastore.GetMIBValueString(a[0], &variable, false)
				if ip != "" && mac != "" {
					ipToMac[ip] = mac
					macToIP[mac] = ip
				}
			}
			return nil
		})
	}
	return ipToMac, macToIP
}

// findBestPollingIDForNode selects the most suitable polling ID for a node connected to a port
func findBestPollingIDForNode(node *datastore.NodeEnt, ifIndex string) string {
	pid := ""
	pcmp := fmt.Sprintf("ifOperStatus.%s", ifIndex)
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == node.ID {
			if p.Type == "snmp" && p.Params == pcmp {
				pid = p.ID
				return false
			}
			if pid == "" {
				pid = p.ID
			} else if p.Type == "ping" {
				pid = p.ID
			}
		}
		return true
	})
	return pid
}

// hasNeighborLine checks if a line already exists in the candidate slice
func hasNeighborLine(lines []NeighborLineEnt, l1 *datastore.LineEnt) bool {
	for _, l := range lines {
		if l.NodeID1 == l1.NodeID1 && l.PollingID1 == l1.PollingID1 &&
			l.NodeID2 == l1.NodeID2 && l.PollingID2 == l1.PollingID2 {
			return true
		}
		if l.NodeID1 == l1.NodeID2 && l.PollingID1 == l1.PollingID2 &&
			l.NodeID2 == l1.NodeID1 && l.PollingID2 == l1.PollingID1 {
			return true
		}
	}
	return false
}

// FindNodeConnection searches candidate switch connections for a specific regular node
func FindNodeConnection(nodeID string) ([]NeighborLineEnt, error) {
	nodeID = strings.TrimPrefix(nodeID, "NODE:")
	node := datastore.GetNode(nodeID)
	if node == nil {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	var candidates []NeighborLineEnt
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		resp, err := FindTopologyForNetwork(n)
		if err != nil {
			return true
		}
		for _, l := range resp.Lines {
			if l.NodeID2 == node.ID || l.NodeID1 == node.ID {
				candidates = append(candidates, l)
			}
		}
		return true
	})

	// If no connections found via SNMP MIBs, fallback to subnet heuristics
	if len(candidates) == 0 {
		var networks []*datastore.NetworkEnt
		datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
			networks = append(networks, n)
			return true
		})
		subLines := inferWithSubnetHeuristics([]*datastore.NodeEnt{node}, networks)
		for _, l := range subLines {
			if !hasNeighborLine(candidates, &l.LineEnt) && !datastore.HasLine(&l.LineEnt, false) {
				candidates = append(candidates, l)
			}
		}
	}

	return candidates, nil
}

// AutoConnectLines automatically detects and creates lines across all network nodes
func AutoConnectLines(mode int) (int, int, error) {
	if mode <= datastore.AutoLineNone {
		return 0, 0, nil
	}

	strictCount := 0
	speculativeCount := 0
	var allSpeculativeLines []NeighborLineEnt

	// 1. Gather lines from all switches
	datastore.ForEachNetworks(func(n *datastore.NetworkEnt) bool {
		resp, err := FindTopologyForNetwork(n)
		if err != nil {
			log.Printf("AutoConnectLines network=%s err=%v", n.Name, err)
			return true
		}

		for _, l := range resp.Lines {
			if l.Confidence == "strict" {
				if !datastore.HasLine(&l.LineEnt, false) && !datastore.HasLine(&l.LineEnt, true) {
					if err := datastore.AddLine(&l.LineEnt); err == nil {
						strictCount++
					}
				}
			} else if mode == datastore.AutoLineSpeculative {
				allSpeculativeLines = append(allSpeculativeLines, l)
			}
		}
		return true
	})

	// 2. Connect speculative lines if requested
	if mode == datastore.AutoLineSpeculative {
		for _, l := range allSpeculativeLines {
			if !datastore.HasLine(&l.LineEnt, false) && !datastore.HasLine(&l.LineEnt, true) {
				if err := datastore.AddLine(&l.LineEnt); err == nil {
					speculativeCount++
				}
			}
		}

		// 3. Trigger AI topology inference for remaining unattached nodes if LLM is enabled
		aiLines, err := InferRemainingLinesWithAI()
		if err == nil {
			for _, l := range aiLines {
				if !datastore.HasLine(&l.LineEnt, false) && !datastore.HasLine(&l.LineEnt, true) {
					if err := datastore.AddLine(&l.LineEnt); err == nil {
						speculativeCount++
					}
				}
			}
		}
	}

	if strictCount > 0 || speculativeCount > 0 {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf(i18n.Trans("Auto connected lines: %d strict, %d speculative"), strictCount, speculativeCount),
		})
	}

	return strictCount, speculativeCount, nil
}
