package main

import (
	"fmt"
	"log"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetMqttStatList() []*datastore.MqttStatEnt {
	ret := []*datastore.MqttStatEnt{}
	datastore.ForEachMqttStat(func(s *datastore.MqttStatEnt) bool {
		ret = append(ret, s)
		return true
	})
	return ret
}

func (a *App) DeleteMqttStats(ids []string) {
	datastore.DeleteMqttStats(ids)
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Delete MQTT stat(%d)"), len(ids)),
	})
}

func (a *App) DeleteAllMqttStat() bool {
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
	if err := datastore.DeleteAllMqttStat(); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all MQTT stat"),
	})
	return true
}
