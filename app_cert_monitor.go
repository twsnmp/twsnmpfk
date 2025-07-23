package main

import (
	"log"

	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func (a *App) GetCertMonitorList() []datastore.CertMonitorEnt {
	ret := []datastore.CertMonitorEnt{}
	datastore.ForEachCertMonitors(func(c *datastore.CertMonitorEnt) bool {
		ret = append(ret, *c)
		return true
	})
	return ret
}

type EditCertMonitorEnt struct {
	ID     string `json:"ID"`
	Target string `json:"Target"`
	Port   int    `json:"Port"`
}

func (a *App) UpateCertMonitor(p EditCertMonitorEnt) bool {
	c := datastore.GetCertMonitor(p.ID)
	if c.ID != "" {
		if c.Target != p.Target || c.Port != uint16(p.Port) {
			c.Subject = ""
			c.Issuer = ""
			c.SerialNumber = ""
			c.State = "unknown"
			c.NotAfter = 0
			c.NotBefore = 0
			c.FirstTime = 0
			c.LastTime = 0
		}
	} else {
		c.State = "unknown"
	}
	c.Target = p.Target
	c.Port = uint16(p.Port)
	if err := datastore.SaveCertMonitor(c); err != nil {
		log.Printf("UpdateCertMonitor err=%v", err)
		return false
	}
	backend.DoCehckCertMonitor()
	return true
}

func (a *App) DeleteCertMonitor(id string) {
	datastore.DeleteCertMonitor(id)
}
