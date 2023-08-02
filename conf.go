package main

import (
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/notify"
)

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
