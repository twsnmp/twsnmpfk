package main

import (
	"embed"
	"flag"
	"fmt"
	"log"

	"github.com/twsnmp/twsnmpfk/clog"
	"github.com/twsnmp/twsnmpfk/datastore"
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
var lock = ""
var trapPort = 162
var syslogPort = 514
var sshdPort = 2022
var netflowPort = 2055
var sFlowPort = 6343
var tcpdPort = 8086
var maxDispLog = 10000
var lang = ""
var clientCert = ""
var clientKey = ""
var caCert = ""

func init() {
	flag.StringVar(&dataStorePath, "datastore", "", "Path to data store directory")
	flag.BoolVar(&kiosk, "kiosk", false, "Kisok mode(frameless and full screen)")
	flag.StringVar(&lock, "lock", "", "Disable edit map and lock page(map or loc)")
	flag.IntVar(&trapPort, "trapPort", 162, "SNMP TRAP port")
	flag.IntVar(&syslogPort, "syslogPort", 514, "Syslog port")
	flag.IntVar(&sshdPort, "sshdPort", 2022, "SSH server port")
	flag.IntVar(&netflowPort, "netflowPort", 2055, "Netflow port")
	flag.IntVar(&sFlowPort, "sFlowPort", 6343, "sFlow port")
	flag.IntVar(&tcpdPort, "tcpdPort", 8086, "tcp server port")
	flag.IntVar(&maxDispLog, "maxDispLog", 10000, "Max log size to diplay")
	flag.StringVar(&pingMode, "ping", "", "ping mode icmp or udp")
	flag.StringVar(&lang, "lang", "", "Language(en|jp)")
	flag.StringVar(&clientCert, "clientCert", "", "Client cert path")
	flag.StringVar(&clientKey, "clientKey", "", "Client key path")
	flag.StringVar(&caCert, "caCert", "", "CA Cert path")
	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(args) == 3 && args[0] == "compact" {
		log.Println("start compact DB")
		if err := datastore.CompactDB(args[1], args[2]); err != nil {
			log.Fatalf("compact DB err=%v", err)
		}
		log.Println("end compact DB")
		return
	}

	flag.VisitAll(func(f *flag.Flag) {
		log.Printf("args %s=%s", f.Name, f.Value)
	})

	if lang != "" {
		i18n.SetLang(lang)
	}
	datastore.ClientCert = clientCert
	datastore.ClientKey = clientKey
	datastore.CACert = caCert

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
		Logger: clog.New(),
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
