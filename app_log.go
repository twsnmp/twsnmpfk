package main

import (
	"github.com/twsnmp/twsnmpfk/datastore"
)

// GetLastEventLogs retunrs last event logs
func (a *App) GetLastEventLogs(count int) []datastore.EventLogEnt {
	ret := []datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		ret = append(ret, *l)
		return len(ret) < count
	})
	return ret
}
