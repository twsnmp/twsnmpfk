package main

import (
	"github.com/twsnmp/twsnmpfk/datastore"
)

// GetEventLogs retunrs  event logs
func (a *App) GetEventLogs(count int) []datastore.EventLogEnt {
	if count < 1 {
		count = maxDispLog
	}
	ret := []datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		ret = append(ret, *l)
		return len(ret) < count
	})
	return ret
}

// GetSyslogs retunrs syslogs
func (a *App) GetSyslogs() []datastore.SyslogEnt {
	ret := []datastore.SyslogEnt{}
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		ret = append(ret, *l)
		return len(ret) < maxDispLog
	})
	return ret
}

// GetTraps retunrs syslogs
func (a *App) GetTraps() []datastore.TrapEnt {
	ret := []datastore.TrapEnt{}
	datastore.ForEachLastTraps(func(l *datastore.TrapEnt) bool {
		ret = append(ret, *l)
		return len(ret) < maxDispLog
	})
	return ret
}
