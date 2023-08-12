package main

import (
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/notify"
)

// GetMapConf returns map config
func (a *App) GetMapConf() datastore.MapConfEnt {
	return datastore.MapConf
}

// UpdateMapConf save map config
func (a *App) UpdateMapConf(m datastore.MapConfEnt) bool {
	datastore.MapConf = m
	return datastore.SaveMapConf() == nil
}

// GetMapName returns map name
func (a *App) GetMapName() string {
	return datastore.MapConf.MapName
}

// GetNotifyConf returns notify config
func (a *App) GetNotifyConf() datastore.NotifyConfEnt {
	return datastore.NotifyConf
}

// UpdateNotifyConf save notify config
func (a *App) UpdateNotifyConf(n datastore.NotifyConfEnt) bool {
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

// UpdateAIConf save AI config
func (a *App) UpdateAIConf(ai datastore.AIConfEnt) bool {
	datastore.AIConf = ai
	return datastore.SaveAIConf() == nil
}
