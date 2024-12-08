package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type DiscoverConfEnt struct {
	StartIP    string `json:"StartIP"`
	EndIP      string `json:"EndIP"`
	Timeout    int    `json:"Timeout"`
	Retry      int    `json:"Retry"`
	X          int    `json:"X"`
	Y          int    `json:"Y"`
	AddPolling bool   `json:"AddPolling"`
	PortScan   bool   `json:"PortScan"`
	ReCheck    bool   `json:"ReCheck"`
	AddNetwork bool   `json:"AddNetwork"`
}

func SaveDiscoverConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(DiscoverConf)
	if err != nil {
		return err
	}
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveDiscoverConf dur=%v", time.Since(st))
		return b.Put([]byte("discoverConf"), s)
	})
}
