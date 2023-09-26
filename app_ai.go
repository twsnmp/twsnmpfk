package main

import (
	"fmt"
	"log"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

type AIList struct {
	ID       string  `json:"ID"`
	Node     string  `json:"Node"`
	Polling  string  `json:"Polling"`
	Score    float64 `json:"Score"`
	Count    int     `json:"Count"`
	LastTime int64   `json:"LastTime"`
}

// GetAIList retunrs map AI List
func (a *App) GetAIList() []AIList {
	ret := []AIList{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode != datastore.LogModeAI {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) < 1 {
			return true
		}
		ret = append(ret, AIList{
			ID:       p.ID,
			Node:     n.Name,
			Polling:  p.Name,
			Score:    air.ScoreData[len(air.ScoreData)-1][1],
			Count:    len(air.ScoreData),
			LastTime: air.LastTime,
		})
		return true
	})
	return ret
}

// GetAIResult : retunrs AI Result
func (a *App) GetAIResult(id string) datastore.AIResult {
	r, err := datastore.GetAIReesult(id)
	if err != nil || len(r.ScoreData) < 1 {
		return datastore.AIResult{}
	}
	return *r
}

func (a *App) DeleteAIResult(id string) bool {
	if id == "all" {
		go func() {
			datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
				if err := backend.DeleteAIResult(p.ID); err != nil {
					log.Printf("delete ai result err=%v", err)
				}
				return true
			})
		}()
	} else {
		if err := backend.DeleteAIResult(id); err != nil {
			log.Printf("delete ai result err=%v", err)
			return false
		}
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Delete AI Result(%s)"), id),
	})
	return true
}
