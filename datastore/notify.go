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
	"golang.org/x/oauth2"
)

type NotifyConfEnt struct {
	Provider           string `json:"Provider"`
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
	WebHookNotify      string `json:"WebHookNotify"`
	WebHookReport      string `json:"WebHookReport"`
	ClientID           string `json:"ClientID"`
	ClientSecret       string `json:"ClientSecret"`
	MSTenant           string `json:"MSTenant"`
}

var (
	notifyOAuth2Token        *oauth2.Token
	NotifyOAuth2RedirectPort int
)

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

func SaveNotifyOAuth2Token(token *oauth2.Token) error {
	if db == nil {
		return ErrDBNotOpen
	}
	notifyOAuth2Token = token
	s, err := json.Marshal(notifyOAuth2Token)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("notifyOAuth2Token"), s)
	})
}

func DeleteNotifyOAuth2Token() error {
	if db == nil {
		return ErrDBNotOpen
	}
	notifyOAuth2Token = nil
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Delete([]byte("notifyOAuth2Token"))
	})
}

func loadNotifyOAuth2Token(b *bbolt.Bucket) {
	v := b.Get([]byte("notifyOAuth2Token"))
	if v != nil {
		var t oauth2.Token
		if err := json.Unmarshal(v, &t); err == nil {
			notifyOAuth2Token = &t
		}
	}
}

func GetNotifyOAuth2Token() *oauth2.Token {
	return notifyOAuth2Token
}

func HasValidNotifyOAuth2Token(n NotifyConfEnt) bool {
	if n.Provider == "" || n.Provider == "smtp" {
		return true
	}
	if n.Provider != NotifyConf.Provider ||
		n.ClientID != NotifyConf.ClientID ||
		n.ClientSecret != NotifyConf.ClientSecret ||
		n.MSTenant != NotifyConf.MSTenant ||
		notifyOAuth2Token == nil {
		return false
	}
	return true
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
