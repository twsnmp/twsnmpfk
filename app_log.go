package main

import (
	"fmt"
	"log"
	"regexp"
	"time"

	wails "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	"github.com/twsnmp/twsnmpfk/logger"
)

type EventLogFilterEnt struct {
	NodeID    string `json:"NodeID"`
	Start     string `json:"Start"`
	End       string `json:"End"`
	EventType string `json:"EventType"`
	NodeName  string `json:"NodeName"`
	Event     string `json:"Event"`
	Level     int    `json:"Level"`
}

// GetEventLogs retunrs  event logs
func (a *App) GetEventLogs(filter EventLogFilterEnt) []*datastore.EventLogEnt {
	ret := []*datastore.EventLogEnt{}
	typeFilter := makeStringFilter(filter.EventType)
	nodeFilter := makeStringFilter(filter.NodeName)
	eventFilter := makeStringFilter(filter.Event)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachEventLog(st, et, func(l *datastore.EventLogEnt) bool {
		if filter.NodeID != "" && filter.NodeID != l.NodeID {
			return true
		}
		if typeFilter != nil && !typeFilter.MatchString(l.Type) {
			return true
		}
		if nodeFilter != nil && !nodeFilter.MatchString(l.NodeName) {
			return true
		}
		if eventFilter != nil && !eventFilter.MatchString(l.Event) {
			return true
		}
		if filter.Level != 0 && filter.Level > getLevelNum(l.Level) {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}

func getLevelNum(l string) int {
	switch l {
	case "high":
		return 3
	case "low":
		return 2
	case "warn":
		return 1
	}
	return 0
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

type SyslogFilterEnt struct {
	Start    string `json:"Start"`
	End      string `json:"End"`
	Host     string `json:"Host"`
	Tag      string `json:"Tag"`
	Message  string `json:"Message"`
	Severity int    `json:"Severity"`
}

// GetSyslogs retunrs syslogs
func (a *App) GetSyslogs(filter SyslogFilterEnt) []*datastore.SyslogEnt {
	ret := []*datastore.SyslogEnt{}
	hostFilter := makeStringFilter(filter.Host)
	tagFilter := makeStringFilter(filter.Tag)
	msgFilter := makeStringFilter(filter.Message)
	st := makeTimeFilter(filter.Start, 1)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachSyslog(st, et, func(l *datastore.SyslogEnt) bool {
		if filter.Severity < l.Severity {
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

type TrapFilterEnt struct {
	Start string `json:"Start"`
	End   string `json:"End"`
	From  string `json:"From"`
	Type  string `json:"Type"`
}

// GetTraps retunrs SNMP Trap log
func (a *App) GetTraps(filter TrapFilterEnt) []*datastore.TrapEnt {
	ret := []*datastore.TrapEnt{}
	fromFilter := makeStringFilter(filter.From)
	typeFilter := makeStringFilter(filter.Type)
	st := makeTimeFilter(filter.Start, 24)
	et := makeTimeFilter(filter.End, 0)
	datastore.ForEachTraps(st, et, func(l *datastore.TrapEnt) bool {
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
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm clear"),
		Message:       i18n.Trans("Do you want to clear?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	logger.ResetArpWatch = true
	if err := datastore.ResetArpTable(); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Clear all arp watch info"),
	})
	return true
}

func (a *App) DeleteArpEnt(ips []string) bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteArpEnt(ips); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Delete arp ent ip=%v"), ips),
	})
	return true
}

type ArpLogEnt struct {
	Time      int64  `json:"Time"`
	State     string `json:"State"`
	IP        string `json:"IP"`
	Node      string `json:"Node"`
	NewMAC    string `json:"NewMAC"`
	NewVendor string `json:"NewVendor"`
	OldMAC    string `json:"OldMAC"`
	OldVendor string `json:"OldVendor"`
}

// GetArpLogsは、最新のARP Logを返します。
func (a *App) GetArpLogs() []*ArpLogEnt {
	ret := []*ArpLogEnt{}
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		node := ""
		if n := datastore.FindNodeFromIP(l.IP); n != nil {
			node = n.Name
		}
		ret = append(ret, &ArpLogEnt{
			Time:      l.Time,
			State:     l.State,
			IP:        l.IP,
			NewMAC:    l.NewMAC,
			OldMAC:    l.OldMAC,
			NewVendor: datastore.FindVendor(l.NewMAC),
			OldVendor: datastore.FindVendor(l.OldMAC),
			Node:      node,
		})
		return len(ret) < maxDispLog
	})
	return ret
}

// DeleteAllEventLogsは、Event logを全て削除します。
func (a *App) DeleteAllEventLogs() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("logs"); err != nil {
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all event logs"),
	})
	return true
}

// DeleteAllSyslogは、Syslogを全て削除します。
func (a *App) DeleteAllSyslog() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("syslog"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all syslog"),
	})
	return true
}

// DeleteAllTrapsは、TRAP logを全て削除します。
func (a *App) DeleteAllTraps() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("trap"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all TRAP logs"),
	})
	return true
}

// DeleteAllPollingLogsは、ポーリングログを全て削除します。
func (a *App) DeleteAllPollingLogs() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllLogs("pollingLogs"); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all polling logs"),
	})
	return true
}

func makeTimeFilter(dt string, oh int) int64 {
	if dt == "" {
		return time.Now().Add(-time.Hour * time.Duration(oh)).UnixNano()
	}
	zone, _ := time.Now().Zone()
	var t time.Time
	var err error
	if t, err = time.Parse("2006-01-02T15:04 MST", dt+" "+zone); err != nil {
		log.Println(err)
		t = time.Now().Add(-time.Hour * time.Duration(oh))
	}
	return t.UnixNano()
}

func makeStringFilter(f string) *regexp.Regexp {
	if f == "" {
		return nil
	}
	r, err := regexp.Compile(f)
	if err != nil {
		return nil
	}
	return r
}
