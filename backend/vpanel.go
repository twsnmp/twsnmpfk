package backend

import (
	"sort"
	"strconv"
	"strings"

	"github.com/twsnmp/twsnmpfk/datastore"
)

type VPanelPortEnt struct {
	Index        int64  `json:"Index"`
	State        string `json:"State"`
	Name         string `json:"Name"`
	Speed        int64  `json:"Speed"`
	OutPacktes   int64  `json:"OutPacktes"`
	OutBytes     int64  `json:"OutBytes"`
	OutError     int64  `json:"OutError"`
	InPacktes    int64  `json:"InPacktes"`
	InBytes      int64  `json:"InBytes"`
	InError      int64  `json:"InError"`
	Type         int64  `json:"Type"`
	Admin        int64  `json:"Admin"`
	Oper         int64  `json:"Oper"`
	MAC          string `json:"MAC"`
	pollingIndex string
}

// GetVPanelPowerInfo : パネルの電源状態を取得
func GetVPanelPowerInfo(id string) bool {
	n := datastore.GetNode(id)
	// まずはノードの状態を反映
	state := n.State
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == id && p.Type == "ping" {
			// PINGの状態を反映
			state = p.State
			return false
		}
		return true
	})
	return state == "normal" || state == "repair"
}

// GetVPanelPorts : パネルに表示するポートの情報を取得する
// 優先順位は
// 1.ポーリングの設定
// 2.SNMPから取得
// 3.ラインの設定
func GetVPanelPorts(id string) []VPanelPortEnt {
	// ポーリングから取得
	if ports := getPortsFromPolling(id); len(ports) > 0 {
		return ports
	}
	// SNMPで取得
	if ports := getPortsBySNMP(id); len(ports) > 0 {
		return ports
	}
	// ラインから取得
	return getPortsFromLine(id)
}

func getPortsFromPolling(id string) []VPanelPortEnt {
	ports := []VPanelPortEnt{}
	traffPollings := make(map[string]*datastore.PollingEnt)
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == id && p.Type == "snmp" {
			if p.Mode == "ifOperStatus" && strings.Contains(p.Filter, ":") {
				a := strings.Split(p.Filter, ":")
				if len(a) != 2 {
					return true
				}
				i, err := strconv.ParseInt(a[0], 10, 64)
				if err != nil {
					return true
				}
				state := "down"
				switch p.State {
				case "normal", "repair":
					state = "up"
				case "unknown":
					state = "off"
				}
				ports = append(ports, VPanelPortEnt{
					Index:        i,
					Name:         a[1],
					pollingIndex: p.Params,
					State:        state,
					Type:         6,
				})
			} else if p.Mode == "traffic" {
				traffPollings[p.Params] = p
			}
		}
		return true
	})
	for _, e := range ports {
		if p, ok := traffPollings[e.pollingIndex]; ok {
			e.InBytes = getTraffData("bytes", p)
			e.InPacktes = getTraffData("packets", p)
			e.InError = getTraffData("errors", p)
			e.OutBytes = getTraffData("outBytes", p)
			e.OutPacktes = getTraffData("outPackets", p)
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Index < ports[j].Index
	})
	return ports
}

func getTraffData(k string, p *datastore.PollingEnt) int64 {
	if d, ok := p.Result[k]; ok {
		if v, ok := d.(float64); ok {
			return int64(v)
		}
	}
	return 0
}

func getPortsFromLine(id string) []VPanelPortEnt {
	ports := []VPanelPortEnt{}
	max := int64(0)
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 != id && l.NodeID2 != id {
			return true
		}
		name := ""
		i := int64(0)
		a := strings.Split(l.Port, ":")
		if len(a) == 2 {
			i, _ = strconv.ParseInt(a[0], 10, 64)
			if max < i {
				max = i
			}
			name = a[1]
		} else if l.NodeID1 == id {
			np := datastore.GetNode(l.NodeID2)
			if np == nil {
				return true
			}
			name = np.Name
		} else {
			np := datastore.GetNode(l.NodeID1)
			if np == nil {
				return true
			}
			name = np.Name
		}
		s := l.State1
		if l.NodeID1 == id {
			s = l.State2
		}
		state := "down"
		switch s {
		case "normal", "repair":
			state = "up"
		case "unknown":
			state = "off"
		}
		ports = append(ports, VPanelPortEnt{
			Index: i,
			Name:  name,
			State: state,
			Type:  6,
		})
		return true
	})
	max++
	for _, e := range ports {
		if e.Index == 0 {
			e.Index = max
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].Index == ports[j].Index {
			return strings.Compare(ports[i].Name, ports[j].Name) < 0
		}
		return ports[i].Index < ports[j].Index
	})
	for i, e := range ports {
		e.Index = int64(i + 1)
	}
	return ports
}
