package main

import (
	"log"
	"regexp"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/logger"
)

// GetEventLogs retunrs  event logs
func (a *App) GetEventLogs(id string) []*datastore.EventLogEnt {
	ret := []*datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(func(l *datastore.EventLogEnt) bool {
		if id == "" || id == l.NodeID {
			ret = append(ret, l)
		}
		return len(ret) < maxDispLog
	})
	return ret
}

// GetMapEventLogs retunrs  event logs
func (a *App) GetMapEventLogs() []*datastore.EventLogEnt {
	ret := []*datastore.EventLogEnt{}
	datastore.ForEachLastEventLog(func(l *datastore.EventLogEnt) bool {
		if l.Type == "user" || l.Type == "oprate" || l.Type == "arpwatch" {
			return true
		}
		ret = append(ret, l)
		return len(ret) < 100
	})
	return ret
}

// GetSyslogs retunrs syslogs
func (a *App) GetSyslogs(severity int, host, tag, msg string) []*datastore.SyslogEnt {
	ret := []*datastore.SyslogEnt{}
	var hostFilter *regexp.Regexp
	var tagFilter *regexp.Regexp
	var msgFilter *regexp.Regexp
	var err error
	if host != "" {
		if hostFilter, err = regexp.Compile(host); err != nil {
			log.Println(err)
			return ret
		}
	}
	if tag != "" {
		if tagFilter, err = regexp.Compile(tag); err != nil {
			log.Println(err)
			return ret
		}
	}
	if msg != "" {
		if msgFilter, err = regexp.Compile(msg); err != nil {
			log.Println(err)
			return ret
		}
	}
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		if severity < l.Severity {
			return true
		}
		if hostFilter != nil && !hostFilter.MatchString(l.Host) {
			return true
		}
		if tagFilter != nil && !tagFilter.MatchString(l.Tag) {
			return true
		}
		if msgFilter != nil && !msgFilter.MatchString(l.Message) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

// GetTraps retunrs SNMP Trap log
func (a *App) GetTraps(from, trapType string) []*datastore.TrapEnt {
	ret := []*datastore.TrapEnt{}
	var fromFilter *regexp.Regexp
	var typeFilter *regexp.Regexp
	var err error
	if from != "" {
		if fromFilter, err = regexp.Compile(from); err != nil {
			log.Println(err)
			return ret
		}
	}
	if trapType != "" {
		if typeFilter, err = regexp.Compile(trapType); err != nil {
			log.Println(err)
			return ret
		}
	}
	datastore.ForEachLastTraps(func(l *datastore.TrapEnt) bool {
		if fromFilter != nil && !fromFilter.MatchString(l.FromAddress) {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.TrapType) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

// GetArpTableは、ARP Tableを返します。
func (a *App) GetArpTable() []*datastore.ArpEnt {
	ret := []*datastore.ArpEnt{}
	datastore.ForEachArp(func(l *datastore.ArpEnt) bool {
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

// ResetArpTableは、ARP Tableをクリアします。
func (a *App) ResetArpTable() bool {
	logger.ResetArpWatch = true
	return datastore.ResetArpTable() == nil
}

// GetArpLogsは、最新のARP Logを返します。
func (a *App) GetArpLogs() []*datastore.ArpLogEnt {
	ret := []*datastore.ArpLogEnt{}
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}
