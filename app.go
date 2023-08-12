package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
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
