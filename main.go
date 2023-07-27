package main

import (
	"embed"
	"flag"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var version = "vx.x.x"
var commit = ""

var dataStorePath = ""
var backupPath = ""
var kiosk = false
var lock = false

func init() {
	flag.StringVar(&dataStorePath, "datastore", "./datastore", "Path to Data Store directory")
	flag.StringVar(&backupPath, "backup", "", "Backup path")
	flag.BoolVar(&kiosk, "kiosk", false, "Kisok mode(Frameless and Full screen)")
	flag.BoolVar(&lock, "lock", false, "Lock mad edit")
	flag.VisitAll(func(f *flag.Flag) {
		log.Printf("args %s=%s", f.Name, f.Value)
	})
	flag.Parse()
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:      "TWSNMP FK",
		Width:      1024,
		Height:     768,
		Fullscreen: kiosk,
		Frameless:  kiosk,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
