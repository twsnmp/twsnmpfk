package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
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
	Kiosk bool   `json:"Kiosk"`
	Lock  string `json:"Lock"`
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
	if dataStorePath != "" {
		a.startTWSNMP()
	}
	a.setMenu()
}

func (a *App) shutdown(ctx context.Context) {
	if dataStorePath == "" {
		return
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Stop TWSNMP"),
	})
	if a.cancel != nil {
		log.Println("shutdown call cancel")
		a.cancel()
		if a.wg != nil {
			log.Println("shutdown wait start")
			a.wg.Wait()
			log.Println("shutdown wait end")
			datastore.CloseDB()
		}
	}
}

// Menu
func (a *App) setMenu() {
	myMenu := menu.NewMenu()
	if runtime.GOOS == "darwin" {
		myMenu.Append(menu.AppMenu())
		myMenu.Append(menu.EditMenu())
		if !kiosk {
			winMenu := myMenu.AddSubmenu("Window")
			winMenu.AddText("Minimize", keys.CmdOrCtrl("m"), func(cd *menu.CallbackData) {
				wails.WindowMinimise(a.ctx)
			})
			winMenu.AddCheckbox("Zoom", wails.WindowIsMaximised(a.ctx), nil, func(cd *menu.CallbackData) {
				if wails.WindowIsMaximised(a.ctx) {
					wails.WindowUnmaximise(a.ctx)
				} else {
					wails.WindowMaximise(a.ctx)
				}
			})
			winMenu.AddCheckbox("Full Screen", wails.WindowIsFullscreen(a.ctx), keys.CmdOrCtrl("f"), func(cd *menu.CallbackData) {
				if wails.WindowIsFullscreen(a.ctx) {
					wails.WindowUnfullscreen(a.ctx)
				} else {
					wails.WindowFullscreen(a.ctx)
				}
			})
			winMenu.AddSeparator()
			winMenu.AddText("Reload", nil, func(cd *menu.CallbackData) {
				wails.WindowReload(a.ctx)
			})
			winMenu.AddSeparator()
			winMenu.AddText("TWSNMP FK -"+datastore.MapConf.MapName, nil, func(cd *menu.CallbackData) {
				wails.WindowUnminimise(a.ctx)
			})
		}
		wails.MenuSetApplicationMenu(a.ctx, myMenu)
		wails.MenuUpdateApplicationMenu(a.ctx)
	} else {
		fileMenu := myMenu.AddSubmenu("File")
		fileMenu.AddText("About", nil, func(cd *menu.CallbackData) {
			wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
				Type:    wails.InfoDialog,
				Title:   "About TWSNMP FK",
				Message: fmt.Sprintf("TWSNMP FK %s(%s)", version, commit),
			})
		})
		fileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
			wails.Quit(a.ctx)
		})
		if !kiosk {
			winMenu := myMenu.AddSubmenu("Window")
			winMenu.AddText("Minimize", keys.CmdOrCtrl("m"), func(cd *menu.CallbackData) {
				wails.WindowMinimise(a.ctx)
			})
			winMenu.AddCheckbox("Zoom", wails.WindowIsMaximised(a.ctx), nil, func(cd *menu.CallbackData) {
				if wails.WindowIsMaximised(a.ctx) {
					wails.WindowUnmaximise(a.ctx)
				} else {
					wails.WindowMaximise(a.ctx)
				}
			})
			winMenu.AddCheckbox("Full Screen", wails.WindowIsFullscreen(a.ctx), keys.CmdOrCtrl("f"), func(cd *menu.CallbackData) {
				if wails.WindowIsFullscreen(a.ctx) {
					wails.WindowUnfullscreen(a.ctx)
				} else {
					wails.WindowFullscreen(a.ctx)
				}
			})
			winMenu.AddSeparator()
			winMenu.AddText("Reload", nil, func(cd *menu.CallbackData) {
				wails.WindowReload(a.ctx)
			})
		}
	}
	wails.MenuSetApplicationMenu(a.ctx, myMenu)
	wails.MenuUpdateApplicationMenu(a.ctx)
	if datastore.MapConf.MapName != "" {
		wails.WindowSetTitle(a.ctx, "TWSNMP FK - "+datastore.MapConf.MapName)
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

// GetImageIconList return IconMap
func (a *App) GetImageIconList() []string {
	ret := []string{}
	if dataStorePath == "" {
		return ret
	}
	if files, err := filepath.Glob(filepath.Join(dataStorePath, "icons", "*.png")); err == nil {
		for _, p := range files {
			ret = append(ret, filepath.Base(p))
		}
	}
	if files, err := filepath.Glob(filepath.Join(dataStorePath, "icons", "*.jpg")); err == nil {
		for _, p := range files {
			ret = append(ret, filepath.Base(p))
		}
	}
	return ret
}

// GetImageIcon returns icon image data
func (a *App) GetImageIcon(icon string) string {
	return a.GetImage(filepath.Join(dataStorePath, "icons", icon))
}

// GetLangは言語を返します。
func (a *App) GetLang() string {
	return i18n.GetLang()
}

// IsDarkは,Darkモードの状態を返します。
func (a *App) IsDark() bool {
	return datastore.Dark
}

// IsLatestは,TWSNMPが最新版であることを返します。
func (a *App) IsLatest() bool {
	return backend.IsLatest()
}

// SetDarkは,Darkモードの状態を設定します。
func (a *App) SetDark(d bool) {
	if err := datastore.SetDark(d); err != nil {
		log.Println(err)
	}
}

// HasDatastoreは,データストアの選択状態を返します。
func (a *App) HasDatastore() bool {
	return dataStorePath != ""
}

// GetMonitorDatasは、システムリソースのモニター情報を返します。
func (a *App) GetMonitorDatas() []*backend.MonitorDataEnt {
	return backend.MonitorDataes
}

// SelectDatastore は、データストアのディレクトリを選択してサービスを起動します。
func (a *App) SelectDatastore() bool {
	if dataStorePath != "" {
		return true
	}
	p, _ := os.UserHomeDir()
	var err error
	dataStorePath, err = wails.OpenDirectoryDialog(a.ctx,
		wails.OpenDialogOptions{
			Title:                i18n.Trans("Select data store path"),
			DefaultDirectory:     p,
			CanCreateDirectories: true,
		})
	if err != nil {
		log.Println(err)
		return false
	}
	if dataStorePath != "" {
		a.startTWSNMP()
		a.setMenu()
		return true
	}
	return false
}

// startTWSNMP は、データストアのディレクトリを設定します。
func (a *App) startTWSNMP() {
	log.Println("call datastore.Init")
	if err := datastore.Init(a.ctx, dataStorePath, a.wg); err != nil {
		log.Fatalf("init db err=%v", err)
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: i18n.Trans("Start TWSNMP"),
	})
	log.Println("call ping.Start")
	if err := ping.Start(a.ctx, a.wg, pingMode); err != nil {
		log.Fatalf("start ping err=%v", err)
	}
	log.Println("call logger.Start")
	if err := logger.Start(a.ctx, a.wg, syslogPort, trapPort, sshdPort, netflowPort, sFlowPort); err != nil {
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

// GetAudio returns image data
func (a *App) GetAudio(path string) string {
	if s, err := os.Stat(path); err != nil || s.Size() > 1024*1024 || s.Size() < 100 {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return ""
	}
	t := "mp3"
	if filepath.Ext(path) == "wav" {
		t = "wav"
	}
	return fmt.Sprintf("data:audio/%s;base64,%s", t, base64.StdEncoding.EncodeToString(b))
}

// SelectAudioFile returns select local file
func (a *App) SelectAudioFile(title string) string {
	filter := []wails.FileFilter{
		{DisplayName: "Audio File(*.mp3,*.wav)", Pattern: "*.mp3;*.wav"},
	}
	file, err := wails.OpenFileDialog(a.ctx, wails.OpenDialogOptions{
		Title:   title,
		Filters: filter,
	})
	if err != nil {
		log.Printf("SelectAudioFile err=%v", err)
	}
	return file
}

// SendFeedback send feedback to twsnmp
func (a *App) SendFeedback(message string) bool {
	msg := message
	msg += fmt.Sprintf("\n-----\nTWSNMP FK\nGOOS=%s,GOARCH=%s\n", runtime.GOOS, runtime.GOARCH)
	if len(backend.MonitorDataes) > 0 {
		i := len(backend.MonitorDataes) - 1
		msg += fmt.Sprintf("CPU=%f,Mem=%f,Load=%f,Disk=%f\n",
			backend.MonitorDataes[i].CPU,
			backend.MonitorDataes[i].Mem,
			backend.MonitorDataes[i].Load,
			backend.MonitorDataes[i].Disk,
		)
	}
	values := url.Values{}
	values.Set("msg", msg)
	values.Add("hash", calcHash(msg))

	req, err := http.NewRequest(
		"POST",
		"https://lhx98.linkclub.jp/twise.co.jp/cgi-bin/twsnmpfb.cgi",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		log.Println(err)
		return false
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}
	if string(r) != "OK" {
		log.Println(r)
		return false
	}
	return true
}

func calcHash(msg string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(msg + time.Now().Format("2006/01/02T15"))); err != nil {
		log.Printf("calc hash err=%v", err)
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
