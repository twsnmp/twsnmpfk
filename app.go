package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

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
	if err := logger.Start(a.ctx, a.wg, syslogPort, trapPort); err != nil {
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

// SelectFile returns select local file
func (a *App) SelectFile(title string, image bool) string {
	filter := []wails.FileFilter{}
	if image {
		filter = append(filter, wails.FileFilter{
			DisplayName: "Image File(*.png,*.jpg)",
			Pattern:     "*.png;*.jpg;",
		})
	}
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title:   title,
		Filters: filter,
	})
	if err != nil {
		log.Printf("SelectFile err=%v", err)
	}
	return file
}

// GetImage returns image data
func (a *App) GetImage(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return ""
	}
	t := "png"
	if filepath.Ext(path) == "jpg" {
		t = "jpeg"
	}
	return fmt.Sprintf("data:image/%s;base64,%s", t, base64.StdEncoding.EncodeToString(b))
}
