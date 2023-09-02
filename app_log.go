package main

import (
	"github.com/twsnmp/twsnmpfk/datastore"
)

// GetEventLogs retunrs  event logs
func (a *App) GetEventLogs() []datastore.EventLogEnt {
	ret := []datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		ret = append(ret, *l)
		return len(ret) < maxDispLog
	})
	return ret
}

// GetAlertEventLogs retunrs  event logs about polling or ai
func (a *App) GetAlertEventLogs() []datastore.EventLogEnt {
	ret := []datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
		if l.Type == "polling" || l.Type == "ai" {
			ret = append(ret, *l)
		}
		return len(ret) < 100
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
