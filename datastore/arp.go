package datastore

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

type ArpEnt struct {
	IP        string `json:"IP"`
	MAC       string `json:"MAC"`
	NodeID    string `json:"NodeID"`
	Vendor    string `json:"Vendor"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

var arpTable = sync.Map{}

func GetArpEnt(ip string) *ArpEnt {
	if v, ok := arpTable.Load(ip); ok {
		if e, ok := v.(*ArpEnt); ok {
			return e
		}
	}
	return nil
}

func UpdateArpEnt(ip, mac string) {
	if v, ok := arpTable.Load(ip); ok {
		if e, ok := v.(*ArpEnt); ok {
			e.MAC = mac
			e.LastTime = time.Now().Unix()
			if e.NodeID != "" {
				if n := GetNode(e.NodeID); n != nil && (e.MAC == mac || e.IP == ip) {
					return
				}
			}
			if n := FindNodeFromIP(ip); n != nil {
				e.NodeID = n.ID
			} else if n := FindNodeFromMAC(mac); n != nil {
				e.NodeID = n.ID
			} else {
				e.NodeID = ""
			}
		}
		return
	}
	var e ArpEnt
	if n := FindNodeFromIP(ip); n != nil {
		e.NodeID = n.ID
	} else if n := FindNodeFromMAC(mac); n != nil {
		e.NodeID = n.ID
	}
	e.IP = ip
	e.MAC = mac
	e.Vendor = FindVendor(mac)
	e.FirstTime = time.Now().Unix()
	e.LastTime = e.FirstTime
	arpTable.Store(ip, &e)
}

func ForEachArp(f func(*ArpEnt) bool) {
	arpTable.Range(func(k, v any) bool {
		if _, ok := k.(string); ok {
			if e, ok := v.(*ArpEnt); ok {
				return f(e)
			}
		}
		return true
	})
}

// ResetArpTableは、ARPテーブルとARPログをクリアする
func ResetArpTable() error {
	st := time.Now()
	arpTable = sync.Map{}
	err := db.Batch(func(tx *bbolt.Tx) error {
		tx.DeleteBucket([]byte("arp"))
		tx.DeleteBucket([]byte("arplog"))
		tx.CreateBucketIfNotExists([]byte("arp"))
		tx.CreateBucketIfNotExists([]byte("arplog"))
		return nil
	})
	log.Printf("ResetArpTable  dur=%v", time.Since(st))
	return err
}

// DeleteArpEntは、指定のIPアドレスに関連したARPテーブルとARPログを削除する
func DeleteArpEnt(ips []string) error {
	st := time.Now()
	err := db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		delMap := make(map[string]bool)
		for _, ip := range ips {
			b.Delete([]byte(ip))
			arpTable.Delete(ip)
			delMap[ip] = true
		}
		b = tx.Bucket([]byte("arplog"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
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
			if _, ok := delMap[a[1]]; ok {
				c.Delete()
			}
		}
		return nil
	})
	log.Printf("DeleteArpEnt len=%d dur=%v", len(ips), time.Since(st))
	return err
}

func loadArpTable() error {
	if db == nil {
		return ErrDBNotOpen
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			ip := string(k)
			var e ArpEnt
			if bytes.HasPrefix(v, []byte("{")) {
				if err := json.Unmarshal(v, &e); err != nil {
					return nil
				}
				if e.Vendor == "Unknown" {
					e.Vendor = FindVendor(e.MAC)
				}
				if n := GetNode(e.NodeID); n == nil {
					e.NodeID = ""
				}
			} else {
				// Old Arp Data
				mac := string(v)
				if n := FindNodeFromIP(ip); n != nil {
					e.NodeID = n.ID
				} else if n := FindNodeFromMAC(mac); n != nil {
					e.NodeID = n.ID
				}
				e.IP = ip
				e.MAC = mac
				e.Vendor = FindVendor(mac)
				e.FirstTime = time.Now().Unix()
				e.LastTime = e.FirstTime
			}
			arpTable.Store(ip, &e)
			return nil
		})
		return nil
	})
	return nil
}

func saveArpTable() error {
	if db == nil {
		return ErrDBNotOpen
	}
	st := time.Now()
	err := db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		arpTable.Range(func(k, v any) bool {
			if ip, ok := k.(string); ok {
				if e, ok := v.(*ArpEnt); ok {
					if j, err := json.Marshal(e); err == nil {
						b.Put([]byte(ip), j)
					}
				}
			}
			return true
		})
		return nil
	})
	log.Printf("save arp table dur=%v", time.Since(st))
	return err
}

func deleteOldArpTable() {
	if db == nil {
		return
	}
	st := time.Now()
	delList := []string{}
	th := time.Now().Unix() - int64(MapConf.LogDays*24*3600)
	arpTable.Range(func(k, v any) bool {
		if ip, ok := k.(string); ok {
			if e, ok := v.(*ArpEnt); ok {
				if e.LastTime < th {
					delList = append(delList, ip)
				}
			}
		}
		return true
	})
	if len(delList) < 1 {
		return
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		for _, ip := range delList {
			arpTable.Delete(ip)
			b.Delete([]byte(ip))
		}
		return nil
	})
	log.Printf("delete old arp table len=%d dur=%v", len(delList), time.Since(st))
}
