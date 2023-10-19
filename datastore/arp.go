package datastore

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type ArpEnt struct {
	IP     string `json:"IP"`
	MAC    string `json:"MAC"`
	NodeID string `json:"NodeID"`
	Vendor string `json:"Vendor"`
}

func UpdateArpEnt(ip, mac string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		log.Printf("UpdateArpEnt dur=%v", time.Since(st))
		return b.Put([]byte(ip), []byte(mac))
	})
}

func ForEachArp(f func(*ArpEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			ip := string(k)
			mac := string(v)
			nodeID := ""
			if n := FindNodeFromIP(ip); n != nil {
				nodeID = n.ID
			} else if n := FindNodeFromMAC(mac); n != nil {
				nodeID = n.ID
			}
			var e = ArpEnt{
				IP:     ip,
				MAC:    mac,
				NodeID: nodeID,
				Vendor: FindVendor(mac),
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

// ResetArpTableは、ARPテーブルとARPログをクリアする
func ResetArpTable() error {
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("arp"))
		tx.DeleteBucket([]byte("arplog"))
		tx.CreateBucketIfNotExists([]byte("arp"))
		tx.CreateBucketIfNotExists([]byte("arplog"))
		log.Printf("ResetArpTable  dur=%v", time.Since(st))
		return nil
	})
}

// DeleteArpEntは、指定のIPアドレスに関連したARPテーブルとARPログを削除する
func DeleteArpEnt(ip string) error {
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if ip == string(k) {
				c.Delete()
			}
		}
		b = tx.Bucket([]byte("arplog"))
		if b == nil {
			return nil
		}
		c = b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if bytes.HasSuffix(v, []byte{0, 0, 255, 255}) {
				v = deCompressLog(v)
			}
			var l LogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				log.Println(err)
				continue
			}
			a := strings.Split(l.Log, ",")
			if len(a) < 3 {
				continue
			}
			ip := a[1]
			if ip == a[1] {
				c.Delete()
			}
		}
		log.Printf("DeleteArpEnt i=%s dur=%v", ip, time.Since(st))
		return nil
	})
}
