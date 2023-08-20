package main

import (
	"github.com/twsnmp/twsnmpfk/datastore"
)

// GetEventLogs retunrs  event logs
func (a *App) GetEventLogs(count int) []datastore.EventLogEnt {
	if count < 1 {
		count = 100 * 10000
	}
	ret := []datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		ret = append(ret, *l)
		return len(ret) < count
	})
	return ret
}
