package datastore

import (
	"encoding/json"
	"fmt"
	"sync"

	"go.etcd.io/bbolt"
)

type CertMonitorEnt struct {
	ID           string `json:"ID"`
	State        string `json:"State"`
	Target       string `json:"Target"`
	Port         uint16 `json:"Port"`
	Subject      string `json:"Subject"`
	Issuer       string `json:"Issuer"`
	SerialNumber string `json:"SerialNumber"`
	Verify       bool   `json:"Verify"`
	NotAfter     int64  `json:"NotAfter"`
	NotBefore    int64  `json:"NotBefore"`
	Error        string `json:"Error"`
	FirstTime    int64  `json:"FirstTime"`
	LastTime     int64  `json:"LastTime"`
}

var certMonitors sync.Map

func GetCertMonitor(id string) *CertMonitorEnt {
	if id != "" {
		if v, ok := certMonitors.Load(id); !ok {
			return v.(*CertMonitorEnt)
		}
	}
	return &CertMonitorEnt{}
}

func DeleteCertMonitor(id string) {
	if _, ok := certMonitors.Load(id); !ok {
		return
	}
	certMonitors.Delete(id)
	if db == nil {
		return
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("certMonitor"))
		if b == nil {
			return fmt.Errorf("cert monitor bucket not found")
		}
		return b.Delete([]byte(id))
	})
}

func SaveCertMonitor(c *CertMonitorEnt) error {
	if c.ID == "" {
		c.ID = makeKey()
	}
	certMonitors.Store(c.ID, c)
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("certMonitor"))
		if b == nil {
			return fmt.Errorf("cert monitor bucket not found")
		}
		s, err := json.Marshal(c)
		if err != nil {
			return err
		}
		return b.Put([]byte(c.ID), s)
	})
}

func ForEachCertMonitors(f func(*CertMonitorEnt) bool) {
	certMonitors.Range(func(k, v interface{}) bool {
		if c, ok := v.(*CertMonitorEnt); ok {
			return f(c)
		}
		return true
	})
}

func loadCertMonitor() error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("certMonitor"))
		if b == nil {
			return fmt.Errorf("cert monitor bucket not found")
		}
		return b.ForEach(func(k, v []byte) error {
			var e CertMonitorEnt
			if err := json.Unmarshal(v, &e); err == nil {
				certMonitors.Store(e.ID, &e)
			}
			return nil
		})
	})
}
