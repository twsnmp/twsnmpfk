package main

import (
	"fmt"
	"log"
	"strings"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

// GetNodes retunrs map nodes
func (a *App) GetNodes() map[string]datastore.NodeEnt {
	wails.LogDebug(a.ctx, "GetNodes")
	ret := make(map[string]datastore.NodeEnt)
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		ret[n.ID] = *n
		return true
	})
	return ret
}

// GetLines retunrs map lines
func (a *App) GetLines() []datastore.LineEnt {
	backend.UpdateLineState()
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
		if strings.Contains(i.Text, "\t") {
			i.Text = strings.Split(i.Text, "\t")[0]
		}
		item := *i
		checkDrawItem(&item)
		ret[item.ID] = item
		return true
	})
	return ret
}

// GetNetworks retunrs map networks
func (a *App) GetNetworks() map[string]datastore.NetworkEnt {
	ret := make(map[string]datastore.NetworkEnt)
	datastore.ForEachNetworks(func(i *datastore.NetworkEnt) bool {
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
	if di.Type >= 5 {
		di.Value = 0.0
	}
	p := datastore.GetPolling(di.PollingID)
	if p == nil {
		return
	}
	if di.Type == datastore.DrawItemTypePollingLine {
		switch p.State {
		case "high":
			di.Color = "#e31a1c"
		case "low":
			di.Color = "#fb9a99"
		case "warn":
			di.Color = "#dfdf22"
		default:
			di.Color = "#1f78b4"
		}
		setPollingLogValuesForLine(di)
		return
	}
	varName, format, scale := autoGetPollingSetting(di, p)
	i, ok := p.Result[varName]
	text := ""
	val := 0.0
	if ok {
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
	}
	if text == "" {
		text = "No Value"
	}
	switch di.Type {
	case datastore.DrawItemTypePollingGauge, datastore.DrawItemTypePollingNewGauge, datastore.DrawItemTypePollingBar:
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
	case datastore.DrawItemTypePollingKPI:
		title := di.Text
		if strings.Contains(title, "\t") {
			title = strings.Split(title, "\t")[0]
		}
		if title == "" {
			if di.VarName != "" {
				title = fmt.Sprintf("%s (%s)", p.Name, di.VarName)
			} else {
				title = p.Name
			}
		}
		di.Text = title
		di.FormattedText = text
		di.Value = val
		switch p.State {
		case "high":
			di.Color = "#ef4444"
		case "low":
			di.Color = "#f87171"
		case "warn":
			di.Color = "#f59e0b"
		default:
			di.Color = "#00d2ff"
		}
		setPollingLogValuesForLine(di)
	}
}

func setPollingLogValuesForLine(di *datastore.DrawItemEnt) {
	di.Values = []float64{}
	datastore.ForEachLastPollingLog(di.PollingID, func(l *datastore.PollingLogEnt) bool {
		if v, ok := l.Result[di.VarName]; ok {
			if val, ok := v.(float64); ok {
				di.Values = append(di.Values, val)
			}
		}
		return len(di.Values) < 60*4
	})
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
		if format == "" && varName == "rtt" {
			if v, ok := p.Result["rtt"].(float64); ok && v > 1000 {
				format = "%.2f ms"
				if scale == 1.0 {
					scale = 0.000001
				}
			}
		}
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

// GetBackImage returns the map background image.
func (a *App) GetBackImage() datastore.BackImageEnt {
	return datastore.BackImage
}

// SetBackImage save map backgrand image
func (a *App) SetBackImage(backImage datastore.BackImageEnt) bool {
	datastore.BackImage = backImage
	return datastore.SaveBackImage() == nil
}

type UpdatePosEnt struct {
	ID string `json:"ID"`
	X  int    `json:"X"`
	Y  int    `json:"Y"`
}

// UpdateNodePos update node positons
func (a *App) UpdateNodePos(list []UpdatePosEnt) {
	var updated []*datastore.NodeEnt
	for _, e := range list {
		n := datastore.GetNode(e.ID)
		if n != nil {
			n.X = e.X
			n.Y = e.Y
			updated = append(updated, n)
		}
	}
	if len(updated) > 0 {
		_ = datastore.SaveNodes(updated)
	}
}

// UpdateNodeLoc update node location
func (a *App) UpdateNodeLoc(id, loc string) {
	n := datastore.GetNode(id)
	if n != nil {
		n.Loc = loc
		_ = datastore.UpdateNode(n)
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

// UpdateNetworkPos update node positons
func (a *App) UpdateNetworkPos(pe UpdatePosEnt) {
	if n := datastore.GetNetwork(pe.ID); n != nil {
		n.X = pe.X
		n.Y = pe.Y
	}
}

func setLineState(l *datastore.LineEnt) {
	l.State1 = "unknown"
	if strings.HasPrefix(l.NodeID1, "NET:") {
		if n := datastore.GetNetwork(l.NodeID1); n != nil {
			for _, p := range n.Ports {
				if p.ID == l.PollingID1 {
					l.State1 = p.State
					break
				}
			}
		}
	} else if node := datastore.GetNode(l.NodeID1); node != nil {
		if node.State == "unknown" {
			l.State1 = "unknown"
		} else if l.PollingID1 != "" {
			if p := datastore.GetPolling(l.PollingID1); p != nil {
				l.State1 = p.State
			} else {
				l.State1 = node.State
			}
		} else {
			l.State1 = node.State
		}
	}
	l.State2 = "unknown"
	if strings.HasPrefix(l.NodeID2, "NET:") {
		if n := datastore.GetNetwork(l.NodeID2); n != nil {
			for _, p := range n.Ports {
				if p.ID == l.PollingID2 {
					l.State2 = p.State
					break
				}
			}
		}
	} else if node := datastore.GetNode(l.NodeID2); node != nil {
		if node.State == "unknown" {
			l.State2 = "unknown"
		} else if l.PollingID2 != "" {
			if p := datastore.GetPolling(l.PollingID2); p != nil {
				l.State2 = p.State
			} else {
				l.State2 = node.State
			}
		} else {
			l.State2 = node.State
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
	// 片側ネットワークの場合は追加のみにする
	if !strings.HasPrefix(node1, "NET:") && !strings.HasPrefix(node2, "NET:") {
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
	}
	return ret
}

// GetLineByID retunrs line
func (a *App) GetLineByID(id string) datastore.LineEnt {
	return *datastore.GetLine(id)
}

// GetLinesByNode retunrs lines conneted to node
func (a *App) GetLinesByNode(id string) []datastore.LineEnt {
	ret := []datastore.LineEnt{}
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 == id || l.NodeID2 == id {
			ret = append(ret, *l)
		}
		return true
	})
	return ret
}

// FindNeighborNetworksAndLines returns neighbor networks and lines
func (a *App) FindNeighborNetworksAndLines(id string) backend.FindNeighborNetworksAndLinesResp {
	if strings.HasPrefix(id, "NODE:") {
		nodeID := strings.TrimPrefix(id, "NODE:")
		lines, err := backend.FindNodeConnection(nodeID)
		if err != nil {
			log.Printf("FindNodeConnection err=%v", err)
			return backend.FindNeighborNetworksAndLinesResp{
				Networks: []datastore.NetworkEnt{},
				Lines:    []backend.NeighborLineEnt{},
			}
		}
		return backend.FindNeighborNetworksAndLinesResp{
			Networks: []datastore.NetworkEnt{},
			Lines:    lines,
		}
	}

	netID := strings.TrimPrefix(id, "NET:")
	n := datastore.GetNetwork(netID)
	if n == nil {
		// Fallback: Check if it's a regular node
		if node := datastore.GetNode(netID); node != nil {
			lines, _ := backend.FindNodeConnection(node.ID)
			return backend.FindNeighborNetworksAndLinesResp{
				Networks: []datastore.NetworkEnt{},
				Lines:    lines,
			}
		}
		return backend.FindNeighborNetworksAndLinesResp{
			Networks: []datastore.NetworkEnt{},
			Lines:    []backend.NeighborLineEnt{},
		}
	}
	return backend.FindNeighborNetworksAndLines(n)
}

// FindNeighborNetworksAndLinesWithAI returns neighbor networks and lines including AI inference
func (a *App) FindNeighborNetworksAndLinesWithAI(id string) backend.FindNeighborNetworksAndLinesResp {
	isNode := strings.HasPrefix(id, "NODE:")
	cleanID := id
	if isNode {
		cleanID = strings.TrimPrefix(id, "NODE:")
	} else {
		cleanID = strings.TrimPrefix(id, "NET:")
	}

	if isNode || datastore.GetNode(cleanID) != nil {
		lines, _ := backend.FindNodeConnection(cleanID)
		aiLines, err := backend.InferRemainingLinesWithAI()
		if err == nil {
			for _, l := range aiLines {
				if l.NodeID1 == cleanID || l.NodeID2 == cleanID {
					lines = append(lines, l)
				}
			}
		}
		return backend.FindNeighborNetworksAndLinesResp{
			Networks: []datastore.NetworkEnt{},
			Lines:    lines,
		}
	}

	res := a.FindNeighborNetworksAndLines(cleanID)
	aiLines, err := backend.InferRemainingLinesWithAI()
	if err == nil {
		for _, l := range aiLines {
			if l.NodeID1 == fmt.Sprintf("NET:%s", cleanID) || l.NodeID2 == fmt.Sprintf("NET:%s", cleanID) {
				res.Lines = append(res.Lines, l)
			}
		}
	}
	return res
}

// FindNodeConnection searches candidate switch connections for a specific regular node
func (a *App) FindNodeConnection(nodeID string) []backend.NeighborLineEnt {
	nodeID = strings.TrimPrefix(nodeID, "NODE:")
	lines, err := backend.FindNodeConnection(nodeID)
	if err != nil {
		log.Printf("FindNodeConnection err=%v", err)
		return []backend.NeighborLineEnt{}
	}
	return lines
}

// AutoConnectLines automatically connects lines across the entire network
func (a *App) AutoConnectLines(mode int) map[string]int {
	strict, speculative, err := backend.AutoConnectLines(mode)
	if err != nil {
		log.Printf("AutoConnectLines err=%v", err)
	}
	return map[string]int{
		"strict":      strict,
		"speculative": speculative,
	}
}

// AutoLayout optimizes the layout of nodes and networks on the map
func (a *App) AutoLayout(mode int) bool {
	count, err := backend.OptimizeLayout(mode)
	if err != nil {
		log.Printf("AutoLayout err=%v", err)
		return false
	}
	return count > 0
}

// UndoAutoLayout restores the previous positions of nodes and networks
func (a *App) UndoAutoLayout() bool {
	count, err := backend.UndoLayout()
	if err != nil {
		log.Printf("UndoAutoLayout err=%v", err)
		return false
	}
	return count > 0
}

// HasUndoAutoLayout checks if an undo snapshot is available
func (a *App) HasUndoAutoLayout() bool {
	return backend.HasUndoLayout()
}

// ConnectLines connects multiple lines at once
func (a *App) ConnectLines(lines []datastore.LineEnt) int {
	count := 0
	for i := range lines {
		l := lines[i]
		if !datastore.HasLine(&l, false) && !datastore.HasLine(&l, true) {
			if err := datastore.AddLine(&l); err == nil {
				count++
			}
		}
	}
	return count
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

// DeleteLine delete line
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
	odi.Cond = di.Cond
	_ = datastore.UpdateDrawItem(odi)
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
	di.Cond = ds.Cond
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

// GetVPanelPorts returns port info of node or network
func (a *App) GetVPanelPorts(id string) []*backend.VPanelPortEnt {
	return backend.GetVPanelPorts(id)
}

// GetVPanelPowerInfo returns power info of node or network
func (a *App) GetVPanelPowerInfo(id string) bool {
	return backend.GetVPanelPowerInfo(id)
}

// GetFDBTable returns FDB Table of node or network
func (a *App) GetFDBTable(id string) []*backend.FDBTableEnt {
	return backend.GetFDBTable(id)
}

// addNetwork add network
func (a *App) addNetwork(n datastore.NetworkEnt) bool {
	if err := datastore.AddNetwork(&n); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		Event:    i18n.Trans("Add Network"),
	})
	return true
}

// GetNetwork returns draw network
func (a *App) GetNetwork(id string) datastore.NetworkEnt {
	n := datastore.GetNetwork(id)
	if n == nil {
		return datastore.NetworkEnt{}
	}
	return *n
}

// UpdateNetwork update draw network
func (a *App) UpdateNetwork(n datastore.NetworkEnt) bool {
	rn := datastore.GetNetwork(n.ID)
	if rn == nil {
		if n.ID != "" {
			log.Printf("netwrok not found id=%s", n.ID)
		}
		return a.addNetwork(n)
	}
	datastore.UpdateNetwork(&n)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		Event:    i18n.Trans("Update Network"),
	})
	return true
}

// DeleteNetwork deletes a network.
func (a *App) DeleteNetwork(id string) {
	n := datastore.GetNetwork(id)
	if n == nil {
		return
	}
	datastore.DeleteNetwork(id)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeName: n.Name,
		Event:    i18n.Trans("Delete Network"),
	})
}

// CheckNetwork check network state
func (a *App) CheckNetwork(id string) {
	backend.CheckNetwork(id)
}
