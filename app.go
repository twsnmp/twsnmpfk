package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx      context.Context
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
	a.ctx = ctx
}

// GetVersin returns version
func (a *App) GetVersion() string {
	return fmt.Sprintf("%s(%s)", version, commit)
}

// GetSettings returns settings
func (a *App) GetSettings() Settings {
	return a.settings
}
