package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/discover"
	"github.com/twsnmp/twsnmpfk/i18n"
)

// GetNodes retunrs map nodes
func (a *App) GetNodes() map[string]datastore.NodeEnt {
	ret := make(map[string]datastore.NodeEnt)
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		ret[n.ID] = *n
		return true
	})
	return ret
}

// GetLines retunrs map lines
func (a *App) GetLines() []datastore.LineEnt {
	ret := []datastore.LineEnt{}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		ret = append(ret, *l)
		return true
	})
	return ret
}

// GetDrawItems retunrs map draw items
func (a *App) GetDrawItems() map[string]datastore.DrawItemEnt {
	ret := make(map[string]datastore.DrawItemEnt)
	datastore.ForEachItems(func(i *datastore.DrawItemEnt) bool {
		checkDrawItem(i)
		ret[i.ID] = *i
		return true
	})
	return ret
}

func checkDrawItem(di *datastore.DrawItemEnt) {
	if di.Type < 4 || di.PollingID == "" {
		return
	}
	if di.Type == 4 {
		di.Text = "No Value"
	}
	if di.Type == 5 {
		di.Value = 0.0
	}
	p := datastore.GetPolling(di.PollingID)
	if p == nil {
		return
	}
	varName, format, scale := autoGetPollingSetting(di, p)
	i, ok := p.Result[varName]
	if !ok {
		return
	}
	text := ""
	val := 0.0
	switch v := i.(type) {
	case string:
		if format == "" {
			text = v
		} else {
			text = fmt.Sprintf(format, v)
		}
	case float64:
		v *= scale
		if format == "" {
			text = fmt.Sprintf("%f", v)
		} else if strings.Contains(format, "BPS") {
			bps := humanize.Bytes(uint64(v)) + "PS"
			text = strings.Replace(format, "BPS", bps, 1)
		} else if strings.Contains(format, "PPS") {
			pps := humanize.Commaf(v) + "PPS"
			text = strings.Replace(format, "PPS", pps, 1)
		} else {
			text = fmt.Sprintf(format, v)
		}
		val = v
	}
	if text == "" {
		text = "No Value"
	}
	switch di.Type {
	case datastore.DrawItemTypePollingGauge:
		if val > 100.0 {
			val = 100.0
		}
		if val > 90.0 {
			di.Color = "#e31a1c"
		} else if val > 80.0 {
			di.Color = "#dfdf22"
		} else {
			di.Color = "#1f78b4"
		}
		di.Value = val
	case datastore.DrawItemTypePollingText:
		di.Text = text
		switch p.State {
		case "high":
			di.Color = "#e31a1c"
		case "low":
			di.Color = "#fb9a99"
		case "warn":
			di.Color = "#dfdf22"
		default:
			di.Color = "#eee"
		}
		di.Value = val
	}
}

func autoGetPollingSetting(di *datastore.DrawItemEnt, p *datastore.PollingEnt) (varName, format string, scale float64) {
	varName = di.VarName
	format = di.Format
	scale = di.Scale
	if scale == 0.0 {
		scale = 1.0
	}
	// ポーリングだけ選択して変数が空欄なら自動で設定する
	if varName != "" {
		return
	}
	// 値があるものを優先的に返す
	if _, ok := p.Result["bps"]; ok {
		varName = "bps"
		if format == "" {
			format = "BPS"
		}
		scale = 1.0
		return
	}
	if _, ok := p.Result["rtt"]; ok {
		varName = "rtt"
		if format == "" {
			format = "RTT=%.3fSec"
		}
		scale = 0.000000001
		return
	}
	if _, ok := p.Result["state"]; ok {
		varName = "state"
		format = "%s"
		return
	}
	if _, ok := p.Result["avg"]; ok {
		varName = "avg"
		if format == "" {
			format = "AVG=%.2f"
		}
		return
	}
	if _, ok := p.Result["count"]; ok {
		varName = "count"
		if format == "" {
			format = "COUNT=%.0f"
		}
		return
	}
	// 自動選択できないものは、値なしを表示する
	return
}

// GetDrawItems retunrs map backgrand image
func (a *App) GetBackImage() datastore.BackImageEnt {
	return datastore.BackImage
}

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

type UpdatePosEnt struct {
	ID string `json:"ID"`
	X  int    `json:"X"`
	Y  int    `json:"Y"`
}

// UpdateNodePos update node positons
func (a *App) UpdateNodePos(list []UpdatePosEnt) {
	for _, e := range list {
		n := datastore.GetNode(e.ID)
		if n != nil {
			n.X = e.X
			n.Y = e.Y
		}
	}
}

// UpdateDrawItemPos update node positons
func (a *App) UpdateDrawItemPos(list []UpdatePosEnt) {
	for _, e := range list {
		n := datastore.GetDrawItem(e.ID)
		if n != nil {
			n.X = e.X
			n.Y = e.Y
		}
	}
}

func setLineState(l *datastore.LineEnt) {
	l.State1 = "unknown"
	if l.PollingID1 != "" {
		if p := datastore.GetPolling(l.PollingID1); p != nil {
			l.State1 = p.State
		}
	}
	l.State2 = l.State1
	if l.PollingID2 != "" {
		if p := datastore.GetPolling(l.PollingID2); p != nil {
			l.State2 = p.State
		}
	}
}

// GetLine retunrs line
func (a *App) GetLine(node1, node2 string) datastore.LineEnt {
	ret := datastore.LineEnt{
		NodeID1: node1,
		NodeID2: node2,
		Width:   2,
	}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 == node1 && l.NodeID2 == node2 {
			ret = *l
			return false
		}
		if l.NodeID2 == node1 && l.NodeID1 == node2 {
			ret = *l
			return false
		}
		return true
	})
	return ret
}

// addLine add line
func (a *App) addLine(lu datastore.LineEnt) bool {
	if err := datastore.AddLine(&lu); err != nil {
		log.Printf("post line err=%v", err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   lu.NodeID1,
		NodeName: getNodeName(lu.NodeID1),
		Event:    fmt.Sprintf(i18n.Trans("Add line to %s"), getNodeName(lu.NodeID2)),
	})
	return true
}

// UpdateLine update line
func (a *App) UpdateLine(lu datastore.LineEnt) bool {
	setLineState(&lu)
	l := datastore.GetLine(lu.ID)
	if l == nil {
		if lu.ID != "" {
			log.Printf("line not found id=%s", lu.ID)
		}
		return a.addLine(lu)
	}
	l.NodeID1 = lu.NodeID1
	l.NodeID2 = lu.NodeID2
	l.PollingID1 = lu.PollingID1
	l.PollingID2 = lu.PollingID2
	l.State1 = lu.State1
	l.State2 = lu.State2
	l.Info = lu.Info
	l.PollingID = lu.PollingID
	l.Width = lu.Width
	l.Port = lu.Port
	if err := datastore.UpdateLine(l); err != nil {
		log.Printf("post line err=%v", err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   lu.NodeID1,
		NodeName: getNodeName(lu.NodeID1),
		Event:    fmt.Sprintf(i18n.Trans("Update line to %s"), getNodeName(lu.NodeID2)),
	})
	return true
}

// UpdateLine upadte line
func (a *App) DeleteLine(id string) bool {
	if err := datastore.DeleteLine(id); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func getNodeName(id string) string {
	if n := datastore.GetNode(id); n != nil {
		return n.Name
	}
	return ""
}

// addDrawItem add draw item
func (a *App) addDrawItem(di datastore.DrawItemEnt) bool {
	if err := datastore.AddDrawItem(&di); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Add draw item"),
	})
	return true
}

// GetDrawItem returns draw item
func (a *App) GetDrawItem(id string) datastore.DrawItemEnt {
	di := datastore.GetDrawItem(id)
	if di == nil {
		return datastore.DrawItemEnt{
			W:     100,
			H:     32,
			Size:  24,
			Scale: 1.0,
			Color: "#888",
		}
	}
	return *di
}

// UpdateDrawItem update draw item
func (a *App) UpdateDrawItem(di datastore.DrawItemEnt) bool {
	odi := datastore.GetDrawItem(di.ID)
	if odi == nil {
		if di.ID != "" {
			log.Printf("draw item not found id=%s", di.ID)
		}
		return a.addDrawItem(di)
	}
	odi.Type = di.Type
	odi.W = di.W
	odi.H = di.H
	odi.Path = di.Path
	odi.Text = di.Text
	odi.Size = di.Size
	odi.Color = di.Color
	odi.Format = di.Format
	odi.VarName = di.VarName
	odi.PollingID = di.PollingID
	odi.Scale = di.Scale
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Update draw item"),
	})
	return true
}

// DeleteDrawItems delete draw items
func (a *App) DeleteDrawItems(ids []string) {
	for _, id := range ids {
		datastore.DeleteDrawItem(id)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Delete draw item(%d)"), len(ids)),
	})
}

// CopyDrawItem : copy ndde
func (a *App) CopyDrawItem(id string) bool {
	ds := datastore.GetDrawItem(id)
	if ds == nil {
		return false
	}
	di := datastore.DrawItemEnt{}
	di.ID = ""
	di.X = ds.X + 100
	di.Y = ds.Y
	di.Type = ds.Type
	di.W = ds.W
	di.H = ds.H
	di.Path = ds.Path
	di.Text = ds.Text
	di.Size = ds.Size
	di.Color = ds.Color
	di.Format = ds.Format
	di.VarName = ds.VarName
	di.PollingID = ds.PollingID
	di.Scale = ds.Scale
	if !a.addDrawItem(di) {
		log.Printf("fail to copy draw item id=%s", id)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Copy draw item"),
	})
	return true
}

// GetVPanelPorts returns port info of node
func (a *App) GetVPanelPorts(id string) []backend.VPanelPortEnt {
	return backend.GetVPanelPorts(id)
}

// GetVPanelPowerInfo returns power info of node
func (a *App) GetVPanelPowerInfo(id string) bool {
	return backend.GetVPanelPowerInfo(id)
}
