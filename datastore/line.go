package datastore

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type LineEnt struct {
	ID         string `json:"ID"`
	NodeID1    string `json:"NodeID1"`
	PollingID1 string `json:"PollingID1"`
	State1     string `json:"State1"`
	NodeID2    string `json:"NodeID2"`
	PollingID2 string `json:"PollingID2"`
	State2     string `json:"State2"`
	PollingID  string `json:"PollingID"`
	Width      int    `json:"Width"`
	State      string `json:"State"`
	Info       string `json:"Info"`
	Port       string `json:"Port"`
}

func AddLine(l *LineEnt) error {
	st := time.Now()
	for {
		l.ID = makeKey()
		if _, ok := lines.Load(l.ID); !ok {
			break
		}
	}
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	lines.Store(l.ID, l)
	log.Printf("AddLine dur=%v", time.Since(st))
	return nil
}

func UpdateLine(l *LineEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := lines.Load(l.ID); !ok {
		return ErrInvalidID
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Put([]byte(l.ID), s)
	})
	log.Printf("UpdateLine dur=%v", time.Since(st))
	return nil
}

func DeleteLine(lineID string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := lines.Load(lineID); !ok {
		return ErrInvalidID
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		return b.Delete([]byte(lineID))
	})
	lines.Delete(lineID)
	log.Printf("DelteLine dur=%v", time.Since(st))
	return nil
}

func GetLine(lineID string) *LineEnt {
	if db == nil {
		return nil
	}
	if n, ok := lines.Load(lineID); ok {
		return n.(*LineEnt)
	}
	return nil
}

// ForEachLines : Line毎の処理
func ForEachLines(f func(*LineEnt) bool) {
	lines.Range(func(_, v interface{}) bool {
		return f(v.(*LineEnt))
	})
}
