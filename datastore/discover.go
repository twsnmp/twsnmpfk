package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

const (
	AutoLineNone        = 0
	AutoLineStrict      = 1
	AutoLineSpeculative = 2
)

// SnmpConfEnt : 自動発見用SNMP設定
type SnmpConfEnt struct {
	SnmpMode     string `json:"SnmpMode"`
	Community    string `json:"Community"`
	SnmpUser     string `json:"SnmpUser"`
	SnmpPassword string `json:"SnmpPassword"`
}

type DiscoverConfEnt struct {
	StartIP      string        `json:"StartIP"`
	EndIP        string        `json:"EndIP"`
	Timeout      int           `json:"Timeout"`
	Retry        int           `json:"Retry"`
	X            int           `json:"X"`
	Y            int           `json:"Y"`
	AddPolling   bool          `json:"AddPolling"`
	PortScan     bool          `json:"PortScan"`
	ReCheck      bool          `json:"ReCheck"`
	AddNetwork   bool          `json:"AddNetwork"`
	AutoDetect   bool          `json:"AutoDetect"`
	AutoDetectAI bool          `json:"AutoDetectAI"`
	AutoLine     int           `json:"AutoLine"`
	SnmpConfigs  []SnmpConfEnt `json:"SnmpConfigs"`
}

// GetDiscoverSnmpConfigs returns the candidate list of SNMP configurations for discovery.
// Base MapConf is the first candidate, followed by any additional SnmpConfigs in order.
func GetDiscoverSnmpConfigs() []SnmpConfEnt {
	var list []SnmpConfEnt
	if MapConf.SnmpMode != "" && MapConf.SnmpMode != "none" {
		list = append(list, SnmpConfEnt{
			SnmpMode:     MapConf.SnmpMode,
			Community:    MapConf.Community,
			SnmpUser:     MapConf.SnmpUser,
			SnmpPassword: MapConf.SnmpPassword,
		})
	}
	for _, sc := range DiscoverConf.SnmpConfigs {
		if sc.SnmpMode == "" {
			continue
		}
		if len(list) > 0 &&
			sc.SnmpMode == list[0].SnmpMode &&
			sc.Community == list[0].Community &&
			sc.SnmpUser == list[0].SnmpUser &&
			sc.SnmpPassword == list[0].SnmpPassword {
			continue
		}
		list = append(list, sc)
	}
	if len(list) == 0 {
		list = append(list, SnmpConfEnt{
			SnmpMode:  "v2c",
			Community: "public",
		})
	}
	return list
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
