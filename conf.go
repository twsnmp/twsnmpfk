package main

import "github.com/twsnmp/twsnmpfk/datastore"

// GetMapConf returns map config
func (a *App) GetMapConf() datastore.MapConfEnt {
	return datastore.MapConf
}

// SetMapConf returns map config
func (a *App) SetMapConf(m datastore.MapConfEnt) bool {
	datastore.MapConf = m
	return datastore.SaveMapConf() == nil
}
