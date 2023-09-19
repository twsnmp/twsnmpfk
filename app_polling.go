package main

import (
	"fmt"
	"log"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/polling"
)

// GetPollings retunrs polling list
func (a *App) GetPollings(node string) []datastore.PollingEnt {
	ret := []datastore.PollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if node == "" || node == p.NodeID {
			ret = append(ret, *p)
		}
		return true
	})
	return ret
}

// GetPolling retunrs polling
func (a *App) GetPolling(id string) datastore.PollingEnt {
	p := datastore.GetPolling(id)
	if p != nil {
		return *p
	}
	return datastore.PollingEnt{}
}

// UpdatePolling add otr update polling
func (a *App) UpdatePolling(up datastore.PollingEnt) bool {
	if up.ID == "" {
		if err := datastore.AddPolling(&up); err != nil {
			log.Printf("Add Polling err=%v", err)
			return false
		}
		return true
	}
	p := datastore.GetPolling(up.ID)
	if p == nil {
		log.Printf("polling not found id=%+v", up)
		return false
	}
	p.Name = up.Name
	p.Type = up.Type
	p.Mode = up.Mode
	p.Params = up.Params
	p.Filter = up.Filter
	p.Extractor = up.Extractor
	p.Script = up.Script
	p.Level = up.Level
	p.PollInt = up.PollInt
	p.Timeout = up.Timeout
	p.Retry = up.Retry
	p.LogMode = up.LogMode
	datastore.UpdatePolling(p, true)
	return true
}

// CheckPolling check node polling
func (a *App) CheckPolling(node string) bool {
	if node == "all" {
		polling.CheckAllPoll()
	} else {
		polling.PollNowNode(node)
	}
	return true
}

// DeletePollings delete polling
func (a *App) DeletePollings(ids []string) {
	datastore.DeletePollings(ids)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf("ポーリングを削除しました %d件", len(ids)),
	})
}

// GetPollingLog returns polling log
func (a *App) GetPollingLogs(id string) []datastore.PollingLogEnt {
	ret := []datastore.PollingLogEnt{}
	polling := datastore.GetPolling(id)
	if polling == nil {
		return ret
	}
	datastore.ForEachLastPollingLog(id, func(l *datastore.PollingLogEnt) bool {
		ret = append(ret, *l)
		return len(ret) <= maxDispLog
	})
	return ret
}

// GetPollingTemplates returns polling templates
func (a *App) GetPollingTemplates() []*datastore.PollingTemplateEnt {
	return datastore.PollingTemplateList
}

func (a *App) GetPollingTemplate(id int) datastore.PollingTemplateEnt {
	pt := datastore.GetPollingTemplate(id)
	if pt == nil {
		return datastore.PollingTemplateEnt{
			Name: "New",
			Type: "ping",
		}
	}
	return *pt
}

// GetAutoPollingsは、ポーリングのテンプレートから自動でポーリングを作成してリストを返します。
func (a *App) GetAutoPollings(node string, id int) []*datastore.PollingEnt {
	n := datastore.GetNode(node)
	if n == nil {
		log.Printf("node not found id=%s", node)
		return nil
	}
	pt := datastore.GetPollingTemplate(id)
	if pt == nil {
		return nil
	}
	if pt.AutoParam != "" {
		return polling.GetAutoPollings(n, pt)
	}
	p := new(datastore.PollingEnt)
	p.Name = pt.Name
	p.NodeID = n.ID
	p.Type = pt.Type
	p.Params = pt.Params
	p.Mode = pt.Mode
	p.Script = pt.Script
	p.Extractor = pt.Extractor
	p.Filter = pt.Filter
	p.Level = pt.Level
	p.PollInt = datastore.MapConf.PollInt
	p.Timeout = datastore.MapConf.Timeout
	p.Retry = datastore.MapConf.Retry
	p.LogMode = 0
	p.NextTime = 0
	p.State = "unknown"
	return []*datastore.PollingEnt{p}
}
