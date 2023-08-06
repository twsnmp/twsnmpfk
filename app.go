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
