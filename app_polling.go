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

// GetGroks retunrs grok list
func (a *App) GetGroks() []datastore.GrokEnt {
	ret := []datastore.GrokEnt{}
	datastore.ForEachGrokEnt(func(g *datastore.GrokEnt) bool {
		ret = append(ret, *g)
		return true
	})
	return ret
}
