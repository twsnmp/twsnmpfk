package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type PollingEnt struct {
	ID           string                 `json:"ID"`
	Name         string                 `json:"Name"`
	NodeID       string                 `json:"NodeID"`
	Type         string                 `json:"Type"`
	Mode         string                 `json:"Mode"`
	Params       string                 `json:"Params"`
	Filter       string                 `json:"Filter"`
	Extractor    string                 `json:"Extractor"`
	Script       string                 `json:"Script"`
	Level        string                 `json:"Level"`
	PollInt      int                    `json:"PollInt"`
	Timeout      int                    `json:"Timeout"`
	Retry        int                    `json:"Retry"`
	LogMode      int                    `json:"LogMode"`
	NextTime     int64                  `json:"NextTime"`
	LastTime     int64                  `json:"LastTime"`
	Result       map[string]interface{} `json:"Result"`
	State        string                 `json:"State"`
	FailAction   string                 `json:"FailAction"`
	RepairAction string                 `json:"RepairAction"`
}

type PollingLogEnt struct {
	Time      int64                  `json:"Time"`
	PollingID string                 `json:"PollingID"`
	State     string                 `json:"State"`
	Result    map[string]interface{} `json:"Result"`
}

// AddPolling : ポーリングを追加する
func AddPolling(p *PollingEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _, ok := pollings.Load(p.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
	p.Result = make(map[string]interface{})
	pollings.Store(p.ID, p)
	SetNodeStateChanged(p.NodeID)
	log.Printf("AddPolling dur=%v", time.Since(st))
	return nil
}

func UpdatePolling(p *PollingEnt, save bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	p.LastTime = time.Now().UnixNano()
	pollings.Store(p.ID, p)
	if !save {
		return nil
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		return b.Put([]byte(p.ID), s)
	})
}

func DeletePollings(ids []string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	for _, id := range ids {
		if e, ok := pollings.Load(id); ok {
			p := e.(*PollingEnt)
			SetNodeStateChanged(p.NodeID)
			pollings.Delete(id)
		}
	}
	// Delete lines
	lines.Range(func(_, p interface{}) bool {
		l := p.(*LineEnt)
		for _, id := range ids {
			if l.PollingID1 == id || l.PollingID2 == id {
				_ = DeleteLine(l.ID)
				return true
			}
		}
		return true
	})
	db.Batch(func(tx *bbolt.Tx) error {
		pb := tx.Bucket([]byte("pollings"))
		aib := tx.Bucket([]byte("ai"))
		if pb != nil && aib != nil {
			for _, id := range ids {
				pb.Delete([]byte(id))
				aib.Delete([]byte(id))
			}
		}
		return nil
	})
	go ClearPollingLogs(ids)
	log.Printf("DeletePollings dur=%v", time.Since(st))
	return nil
}

// GetPolling : ポーリングを取得する
func GetPolling(id string) *PollingEnt {
	if p, ok := pollings.Load(id); ok {
		return p.(*PollingEnt)
	}
	return nil
}

// AddPollingWithDupCheck : 重複しないようにポーリングを追加する
func AddPollingWithDupCheck(p *PollingEnt) error {
	found := false
	pollings.Range(func(_, i interface{}) bool {
		if pe, ok := i.(*PollingEnt); ok {
			if pe.NodeID == p.NodeID && pe.Type == p.Type && pe.Mode == p.Mode && pe.Params == p.Params {
				found = true
				return false
			}
		}
		return true
	})
	if found {
		return nil
	}
	return AddPolling(p)
}

// ForEachPollings : ポーリング毎の処理
func ForEachPollings(f func(*PollingEnt) bool) {
	pollings.Range(func(_, p interface{}) bool {
		return f(p.(*PollingEnt))
	})
}

func saveAllPollings() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		pollings.Range(func(_, p interface{}) bool {
			pe := p.(*PollingEnt)
			s, err := json.Marshal(pe)
			if err == nil {
				b.Put([]byte(pe.ID), s)
			}
			return true
		})
		return nil
	})
	log.Printf("saveAllPollings dur=%v", time.Since(st))
	return nil
}

func AddPollingLog(p *PollingEnt) error {
	if db == nil {
		return ErrDBNotOpen
	}
	pollingLogCh <- &PollingLogEnt{
		Time:      time.Now().UnixNano(),
		PollingID: p.ID,
		State:     p.State,
		Result:    p.Result,
	}
	return nil
}

func ForEachLastPollingLog(pollingID string, f func(*PollingLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return nil
		}
		bs := b.Bucket([]byte(pollingID))
		if bs == nil {
			return nil
		}
		c := bs.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var e PollingLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				log.Printf("load polling log err=%v", err)
				continue
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

// ClearPollingLogs : ポーリングログの削除をまとめて行う
func ClearPollingLogs(ids []string) error {
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return fmt.Errorf("bucket pollingLogs not found")
		}
		for _, id := range ids {
			b.DeleteBucket([]byte(id))
		}
		log.Printf("ClearPollingLogs dur=%v", time.Since(st))
		return nil
	})
}

// GetAllPollingLog :全てのポーリングログを取得する
func GetAllPollingLog(pollingID string) []*PollingLogEnt {
	ret := []*PollingLogEnt{}
	if db == nil {
		return ret
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return nil
		}
		bs := b.Bucket([]byte(pollingID))
		if bs == nil {
			return nil
		}
		c := bs.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var l PollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				log.Printf("load polling log err=%v", err)
				continue
			}
			ret = append(ret, &l)
		}
		return nil
	})
	return ret
}
