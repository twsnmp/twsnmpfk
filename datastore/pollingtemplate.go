package datastore

import (
	"encoding/json"
)

type PollingTemplateEnt struct {
	ID        int    `json:"ID"`
	Name      string `json:"Name"`
	Level     string `json:"Level"`
	Type      string `json:"Type"`
	Mode      string `json:"Mode"`
	Params    string `json:"Params"`
	Filter    string `json:"Filter"`
	Extractor string `json:"Extractor"`
	Script    string `json:"Script"`
	Descr     string `json:"Descr"`
	AutoParam string `json:"AutoParam"`
}

var PollingTemplateList []*PollingTemplateEnt

func GetPollingTemplate(id int) *PollingTemplateEnt {
	if id > 0 && id <= len(PollingTemplateList) {
		return PollingTemplateList[id-1]
	}
	return nil
}

func loadPollingTemplate(js []byte) error {
	var list []PollingTemplateEnt
	if err := json.Unmarshal(js, &list); err != nil {
		return err
	}
	for i := range list {
		list[i].ID = len(PollingTemplateList) + 1
		PollingTemplateList = append(PollingTemplateList, &list[i])
	}
	return nil
}
