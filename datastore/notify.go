package datastore

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

type NotifyConfEnt struct {
	MailServer         string `json:"MailServer"`
	InsecureSkipVerify bool   `json:"InsecureSkipVerify"`
	User               string `json:"User"`
	Password           string `json:"Password"`
	MailTo             string `json:"MailTo"`
	MailFrom           string `json:"MailFrom"`
	Subject            string `json:"Subject"`
	Interval           int    `json:"Interval"`
	Level              string `json:"Level"`
	Report             bool   `json:"Report"`
	NotifyRepair       bool   `json:"NotifyRepair"`
	ExecCmd            string `json:"ExecCmd"`
	BeepHigh           string `json:"BeepHigh"`
	BeepLow            string `json:"BeepLow"`
}

func SaveNotifyConf() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(NotifyConf)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveNotifyConf dur=%v", time.Since(st))
		return b.Put([]byte("notifyConf"), s)
	})
}

func LoadMailTemplate(t string) string {
	f := fmt.Sprintf("mail_%s.html", t)
	if r, err := os.Open(filepath.Join(dspath, f)); err == nil {
		b, err := io.ReadAll(r)
		if err == nil {
			return string(b)
		}
	}
	return mailTemplate[t]
}
