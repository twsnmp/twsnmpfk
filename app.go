package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/discover"
	"github.com/twsnmp/twsnmpfk/logger"
	"github.com/twsnmp/twsnmpfk/notify"
	"github.com/twsnmp/twsnmpfk/ping"
	"github.com/twsnmp/twsnmpfk/polling"
)

// App struct
type App struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
	settings Settings
}

type Settings struct {
	Kiosk bool `json:"Kiosk"`
	Lock  bool `json:"Lock"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		settings: Settings{
			Kiosk: kiosk,
			Lock:  lock,
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.wg = &sync.WaitGroup{}
	a.ctx, a.cancel = context.WithCancel(ctx)
	log.Println("call datastore.Init")
	if err := datastore.Init(a.ctx, dataStorePath, a.wg); err != nil {
		log.Fatalf("init db err=%v", err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FK起動",
	})
	log.Println("call ping.Start")
	if err := ping.Start(a.ctx, a.wg, pingMode); err != nil {
		log.Fatalf("start ping err=%v", err)
	}
	log.Println("call logger.Start")
	if err := logger.Start(a.ctx, a.wg); err != nil {
		log.Fatalf("start logger err=%v", err)
	}
	log.Println("call polling.Start")
	if err := polling.Start(a.ctx, a.wg); err != nil {
		log.Fatalf("start polling err=%v", err)
	}
	log.Println("call backend.Start")
	if err := backend.Start(a.ctx, dataStorePath, version, a.wg); err != nil {
		log.Fatalf("start backend err=%v", err)
	}
	log.Println("call notify.Start")
	if err := notify.Start(a.ctx, a.wg); err != nil {
		log.Fatalf("start notify err=%v", err)
	}

}

func (a *App) shutdown(ctx context.Context) {
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: "TWSNMP FK停止",
	})
	if a.cancel != nil {
		log.Println("shutdown call cancel")
		a.cancel()
		if a.wg != nil {
			log.Println("shutdown wait start")
			a.wg.Wait()
			log.Println("shutdown wait end")
		}
	}
}

// GetVersin returns version
func (a *App) GetVersion() string {
	return fmt.Sprintf("%s(%s)", version, commit)
}

// GetSettings returns settings
func (a *App) GetSettings() Settings {
	return a.settings
}

// GetMapConf returns map config
func (a *App) GetMapConf() datastore.MapConfEnt {
	return datastore.MapConf
}

// SetMapConf save map config
func (a *App) SetMapConf(m datastore.MapConfEnt) bool {
	datastore.MapConf = m
	return datastore.SaveMapConf() == nil
}

// GetNotifyConf returns notify config
func (a *App) GetNotifyConf() datastore.NotifyConfEnt {
	return datastore.NotifyConf
}

// SetNotifyConf save notify config
func (a *App) SetNotifyConf(n datastore.NotifyConfEnt) bool {
	datastore.NotifyConf = n
	return datastore.SaveNotifyConf() == nil
}

// TestNotifyConf test notfiy
func (a *App) TestNotifyConf(n datastore.NotifyConfEnt) bool {
	return notify.SendTestMail(&n) == nil
}

// GetAIConf returns AI config
func (a *App) GetAIConf() datastore.AIConfEnt {
	return datastore.AIConf
}

// SetAIConf save AI config
func (a *App) SetAIConf(ai datastore.AIConfEnt) bool {
	datastore.AIConf = ai
	return datastore.SaveAIConf() == nil
}

// GetLastEventLogs retunrs last event logs
func (a *App) GetLastEventLogs(count int) []datastore.EventLogEnt {
	ret := []datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		ret = append(ret, *l)
		return len(ret) < count
	})
	return ret
}

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
		ret[i.ID] = *i
		return true
	})
	return ret
}

// GetDrawItems retunrs map backgrand image
func (a *App) GetBackImage() datastore.BackImageEnt {
	return datastore.BackImage
}

// GetDiscoverConf retunrs discover config
func (a *App) GetDiscoverConf() datastore.DiscoverConfEnt {
	return datastore.DiscoverConf
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
		Event:    "ノードを追加しました",
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
	n.SnmpMode = nu.SnmpMode
	n.Community = nu.Community
	n.User = nu.User
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
		Event:    "ノードを更新しました",
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
				Event:    "ノードを削除しました",
			})
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
		Event:    "ラインを追加しました",
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   lu.NodeID2,
		NodeName: getNodeName(lu.NodeID2),
		Event:    "ラインを追加しました",
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
		Event:    "ラインを更新しました",
	})
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   lu.NodeID2,
		NodeName: getNodeName(lu.NodeID2),
		Event:    "ラインを更新しました",
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
		Event: "描画アイテムを追加しました",
	})
	return true
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
		Event: "描画アイテムを更新しました",
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
		Event: fmt.Sprintf("描画を削除しました %d件", len(ids)),
	})
}

// GetPollings retunrs polling list
func (a *App) GetPollings(node string) []datastore.PollingEnt {
	ret := []datastore.PollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if node == "" || node == p.NodeID {
			ret = append(ret, *p)
		}
		return true
	})
	return ret
}
