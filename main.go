package main

import (
	"embed"
	"flag"
	"fmt"
	"log"

	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var version = "vx.x.x"
var commit = ""

var dataStorePath = ""
var pingMode string
var kiosk = false
var lock = false
var trapPort = 162
var syslogPort = 514
var maxDispLog = 10000
var lang = ""

func init() {
	flag.StringVar(&dataStorePath, "datastore", "", "Path to Data Store directory")
	flag.BoolVar(&kiosk, "kiosk", false, "Kisok mode(Frameless and Full screen)")
	flag.BoolVar(&lock, "lock", false, "Lock mode edit")
	flag.IntVar(&trapPort, "trapPort", 162, "SNMP TRAP port")
	flag.IntVar(&syslogPort, "syslogPort", 514, "Syslog port")
	flag.IntVar(&maxDispLog, "maxDispLog", 10000, "Max log size to diplay")
	flag.StringVar(&pingMode, "ping", "", "ping mode icmp or udp")
	flag.StringVar(&lang, "lang", "", "Language")
	flag.Parse()
}

func main() {
	flag.VisitAll(func(f *flag.Flag) {
		log.Printf("args %s=%s", f.Name, f.Value)
	})

	if lang != "" {
		i18n.SetLang(lang)
	}

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:      "TWSNMP FK",
		Width:      1600,
		Height:     900,
		Fullscreen: kiosk,
		Frameless:  kiosk,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       false,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "TWSNMP " + fmt.Sprintf("%s(%s)", version, commit),
				Message: "© 2023 Masayuki Yamai",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
