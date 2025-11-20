package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/twsnmp/twsnmpfk/clog"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"gopkg.in/ini.v1"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var version = "vx.x.x"
var commit = ""

var dataStorePath = ""
var kiosk = false
var lock = ""
var maxDispLog = 10000
var lang = ""

func init() {
	flag.StringVar(&dataStorePath, "datastore", "", "Path to data store directory")
	flag.BoolVar(&kiosk, "kiosk", false, "Kisok mode(frameless and full screen)")
	flag.StringVar(&lock, "lock", "", "Disable edit map and lock page(map or loc)")
	flag.IntVar(&datastore.TrapPort, "trapPort", 162, "SNMP TRAP port")
	flag.IntVar(&datastore.SyslogPort, "syslogPort", 514, "Syslog port")
	flag.IntVar(&datastore.SSHdPort, "sshdPort", 2022, "SSH server port")
	flag.IntVar(&datastore.NetFlowPort, "netflowPort", 2055, "Netflow port")
	flag.IntVar(&datastore.SFlowPort, "sFlowPort", 6343, "sFlow port")
	flag.IntVar(&datastore.TCPPort, "tcpdPort", 8086, "tcp server port")
	flag.IntVar(&datastore.OTelgRPCPort, "otelGRPCPort", 4317, "OpenTelemetry server gRPC port")
	flag.IntVar(&datastore.OTelHTTPPort, "otelHTTPPort", 4318, "OpenTelemetry server HTTP port")
	flag.IntVar(&datastore.NotifyOAuth2RedirectPort, "notifyOAuth2Port", 8180, "OAuth2 redirect port")
	flag.IntVar(&maxDispLog, "maxDispLog", 10000, "Max log size to diplay")
	flag.StringVar(&datastore.PingMode, "ping", "", "ping mode icmp or udp")
	flag.StringVar(&lang, "lang", "", "Language(en|jp)")
	flag.StringVar(&datastore.ClientCert, "clientCert", "", "Client cert path")
	flag.StringVar(&datastore.ClientKey, "clientKey", "", "Client key path")
	flag.StringVar(&datastore.CACert, "caCert", "", "CA Cert path")
	flag.StringVar(&datastore.OTelCert, "otelCert", "", "OpenTelemetry server cert path")
	flag.StringVar(&datastore.OTelKey, "otelKey", "", "OpenTelemetry server key path")
	flag.StringVar(&datastore.OTelCA, "otelCA", "", "OpenTelementry CA cert path")
	flag.StringVar(&datastore.MCPCert, "mcpCert", "", "MCP server cert path")
	flag.StringVar(&datastore.MCPKey, "mcpKey", "", "MCP server key path")
	flag.IntVar(&datastore.MqttTCPPort, "mqttTCPPort", 1883, "MQTT server TCP port")
	flag.IntVar(&datastore.MqttWSPort, "mqttWSPort", 1884, "MQTT server WebSock port")
	flag.StringVar(&datastore.MqttCert, "mqttCert", "", "MQTT server cert path")
	flag.StringVar(&datastore.MqttKey, "mqttKey", "", "MQTT server key path")
	flag.StringVar(&datastore.MqttFrom, "mqttFrom", "", "MQTT client IP")
	flag.StringVar(&datastore.MqttUsers, "mqttUsers", "", "MQTT user and password")
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
	loadIni()

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
		Logger: clog.New(),
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func loadIni() {
	cfg, err := ini.Load(filepath.Join(dataStorePath, ".twsnmpfk.ini"))
	if err != nil {
		log.Printf("Fail to read ini file: %v", err)
		return
	}
	// main
	if v := cfg.Section("").Key("lang").MustString(""); v != "" {
		lang = v
	}
	if v := cfg.Section("").Key("lock").MustString(""); v != "" {
		lock = v
	}
	if v := cfg.Section("").Key("maxDispLog").MustInt(0); v > 0 {
		maxDispLog = v
	}
	if v := cfg.Section("").Key("kiosk").MustBool(false); v {
		kiosk = v
	}
	if v := cfg.Section("").Key("notifyOAuth2Port").MustInt(0); v > 0 {
		datastore.NotifyOAuth2RedirectPort = v
	}
	// logger
	if v := cfg.Section("logger").Key("trapPort").MustInt(0); v > 0 {
		datastore.TrapPort = v
	}
	if v := cfg.Section("logger").Key("syslogPort").MustInt(0); v > 0 {
		datastore.SyslogPort = v
	}
	if v := cfg.Section("logger").Key("sshdPort").MustInt(0); v > 0 {
		datastore.SSHdPort = v
	}
	if v := cfg.Section("logger").Key("netflowPort").MustInt(0); v > 0 {
		datastore.NetFlowPort = v
	}
	if v := cfg.Section("logger").Key("sFlowPort").MustInt(0); v > 0 {
		datastore.SFlowPort = v
	}
	if v := cfg.Section("logger").Key("tcpdPort").MustInt(0); v > 0 {
		datastore.TCPPort = v
	}
	// Open Telemetry
	if v := cfg.Section("OTel").Key("otelGRPCPort").MustInt(0); v > 0 {
		datastore.OTelgRPCPort = v
	}
	if v := cfg.Section("OTel").Key("otelHTTPPort").MustInt(0); v > 0 {
		datastore.OTelHTTPPort = v
	}
	if v := cfg.Section("OTel").Key("otelCert").MustString(""); v != "" {
		datastore.OTelCert = v
	}
	if v := cfg.Section("OTel").Key("otelKey").MustString(""); v != "" {
		datastore.OTelKey = v
	}
	if v := cfg.Section("OTel").Key("otelCA").MustString(""); v != "" {
		datastore.OTelCA = v
	}
	// TLS | gRPC Client
	if v := cfg.Section("client").Key("clientCert").MustString(""); v != "" {
		datastore.ClientCert = v
	}
	if v := cfg.Section("client").Key("clientKey").MustString(""); v != "" {
		datastore.ClientKey = v
	}
	if v := cfg.Section("client").Key("caCert").MustString(""); v != "" {
		datastore.CACert = v
	}
	// MCP
	if v := cfg.Section("MCP").Key("mcpCert").MustString(""); v != "" {
		datastore.MCPCert = v
	}
	if v := cfg.Section("MCP").Key("mcpKey").MustString(""); v != "" {
		datastore.MCPKey = v
	}
	// MQTT
	if v := cfg.Section("MQTT").Key("mqttTCPPort").MustInt(0); v > 0 {
		datastore.MqttTCPPort = v
	}
	if v := cfg.Section("MQTT").Key("mqttWSPort").MustInt(0); v > 0 {
		datastore.MqttWSPort = v
	}
	if v := cfg.Section("MQTT").Key("mqttCert").MustString(""); v != "" {
		datastore.MqttCert = v
	}
	if v := cfg.Section("MQTT").Key("MqttKey").MustString(""); v != "" {
		datastore.MqttKey = v
	}
	if v := cfg.Section("MQTT").Key("mqttFrom").MustString(""); v != "" {
		datastore.MqttFrom = v
	}
	if v := cfg.Section("MQTT").Key("MqttUsers").MustString(""); v != "" {
		datastore.MqttUsers = v
	}
}
