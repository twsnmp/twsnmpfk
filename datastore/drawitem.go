package datastore

import (
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
)

type DrawItemType int

const (
	DrawItemTypeRect = iota
	DrawItemTypeEllipse
	DrawItemTypeText
	DrawItemTypeImage
	DrawItemTypePollingText
	DrawItemTypePollingGauge
	DrawItemTypePollingNewGauge
	DrawItemTypePollingBar
	DrawItemTypePollingLine
)

type DrawItemEnt struct {
	ID        string       `json:"ID"`
	Type      DrawItemType `json:"Type"`
	X         int          `json:"X"`
	Y         int          `json:"Y"`
	W         int          `json:"W"`
	H         int          `json:"H"`
	Color     string       `json:"Color"`
	Path      string       `json:"Path"`
	Text      string       `json:"Text"`
	Size      int          `json:"Size"`
	PollingID string       `json:"PollingID"`
	VarName   string       `json:"VarName"`
	Format    string       `json:"Format"`
	Value     float64      `json:"Value"`
	Scale     float64      `json:"Scale"`
	Cond      int          `json:"Cond"`
	Values    []float64    `json:"Values"`
}

func AddDrawItem(di *DrawItemEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		di.ID = makeKey()
		if _, ok := items.Load(di.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(di)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Put([]byte(di.ID), s)
	})
	items.Store(di.ID, di)
	log.Printf("AddItem  dur=%v", time.Since(st))
	return nil
}

func DeleteDrawItem(id string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := items.Load(id); !ok {
		return ErrInvalidID
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("items"))
		return b.Delete([]byte(id))
	})
	items.Delete(id)
	log.Printf("DeleteDrawItem dur=%v", time.Since(st))
	return nil
}

func GetDrawItem(id string) *DrawItemEnt {
	if db == nil {
		return nil
	}
	if di, ok := items.Load(id); ok {
		return di.(*DrawItemEnt)
	}
	return nil
}

func ForEachItems(f func(*DrawItemEnt) bool) {
	items.Range(func(_, p interface{}) bool {
		return f(p.(*DrawItemEnt))
	})
}
