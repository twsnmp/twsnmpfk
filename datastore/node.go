package datastore

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type NodeEnt struct {
	ID           string `json:"ID"`
	Name         string `json:"Name"`
	Descr        string `json:"Descr"`
	Icon         string `json:"Icon"`
	Image        string `json:"Image"`
	State        string `json:"State"`
	X            int    `json:"X"`
	Y            int    `json:"Y"`
	IP           string `json:"IP"`
	MAC          string `json:"MAC"`
	Vendor       string `json:"Vendor"`
	SnmpMode     string `json:"SnmpMode"`
	Community    string `json:"Community"`
	User         string `json:"User"`
	SSHUser      string `json:"SSHUser"`
	Password     string `json:"Password"`
	GNMIPort     string `json:"GNMIPort"`
	GNMIEncoding string `json:"GNMIEncoding"`
	GNMIUser     string `json:"GNMIUser"`
	GNMIPassword string `json:"GNMIPassword"`
	PublicKey    string `json:"PublicKey"`
	URL          string `json:"URL"`
	AddrMode     string `json:"AddrMode"`
	AutoAck      bool   `json:"AutoAck"`
	Loc          string `json:"Loc"`
}

func loadMapData() error {
	if db == nil {
		return ErrDBNotOpen
	}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var n NodeEnt
			if err := json.Unmarshal(v, &n); err == nil {
				nodes.Store(n.ID, &n)
			}
			return nil
		})
		b = tx.Bucket([]byte("items"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var di DrawItemEnt
				if err := json.Unmarshal(v, &di); err == nil {
					items.Store(di.ID, &di)
				}
				return nil
			})
		}
		b = tx.Bucket([]byte("networks"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var n NetworkEnt
				if err := json.Unmarshal(v, &n); err == nil {
					networks.Store(n.ID, &n)
				}
				return nil
			})
		}
		b = tx.Bucket([]byte("lines"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var l LineEnt
				if err := json.Unmarshal(v, &l); err == nil {
					lines.Store(l.ID, &l)
				}
				return nil
			})
		}
		now := time.Now().UnixNano()
		b = tx.Bucket([]byte("pollings"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var p PollingEnt
				if err := json.Unmarshal(v, &p); err == nil {
					if p.Result == nil {
						p.Result = make(map[string]interface{})
					}
					if p.NextTime < now {
						p.NextTime = now
						now += 1000 * 1000 * 500
					}
					pollings.Store(p.ID, &p)
				}
				return nil
			})
		}
		return nil
	})
	return err
}

func AddNode(n *NodeEnt) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	for {
		n.ID = makeKey()
		if _, ok := nodes.Load(n.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		return b.Put([]byte(n.ID), s)
	})
	nodes.Store(n.ID, n)
	log.Printf("AddNode name=%s dur=%v", n.Name, time.Since(st))
	return nil
}

func DeleteNode(nodeID string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := nodes.Load(nodeID); !ok {
		return ErrInvalidID
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("memo"))
		b.Delete([]byte(nodeID))
		b = tx.Bucket([]byte("nodes"))
		return b.Delete([]byte(nodeID))
	})
	nodes.Delete(nodeID)
	delList := []string{}
	pollings.Range(func(k, v interface{}) bool {
		if v.(*PollingEnt).NodeID == nodeID {
			delList = append(delList, k.(string))
		}
		return true
	})
	DeletePollings(delList)
	ForEachArp(func(e *ArpEnt) bool {
		if e.NodeID == nodeID {
			e.NodeID = ""
		}
		return true
	})
	log.Printf("DeleteNode dur=%v", time.Since(st))
	return nil
}

func GetNode(nodeID string) *NodeEnt {
	if db == nil {
		return nil
	}
	if n, ok := nodes.Load(nodeID); ok {
		return n.(*NodeEnt)
	}
	return nil
}

func FindNodeFromIP(ip string) *NodeEnt {
	var ret *NodeEnt
	// IPv4
	ForEachNodes(func(n *NodeEnt) bool {
		if n.IP == ip {
			ret = n
			return false
		}
		return true
	})
	return ret
}

func FindNodeFromName(name string) *NodeEnt {
	var ret *NodeEnt
	ForEachNodes(func(n *NodeEnt) bool {
		if n.Name == name {
			ret = n
			return false
		}
		return true
	})
	return ret
}

func FindNodeFromMAC(mac string) *NodeEnt {
	var ret *NodeEnt
	if mac == "" {
		return ret
	}
	nodes.Range(func(_, p interface{}) bool {
		if strings.HasPrefix(p.(*NodeEnt).MAC, mac) {
			ret = p.(*NodeEnt)
			return false
		}
		return true
	})
	return ret
}

func ForEachNodes(f func(*NodeEnt) bool) {
	nodes.Range(func(_, p interface{}) bool {
		return f(p.(*NodeEnt))
	})
}

func saveAllNodes() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		nodes.Range(func(_, p interface{}) bool {
			pn := p.(*NodeEnt)
			s, err := json.Marshal(pn)
			if err == nil {
				b.Put([]byte(pn.ID), s)
			}
			return true
		})
		b = tx.Bucket([]byte("items"))
		items.Range(func(_, p interface{}) bool {
			di := p.(*DrawItemEnt)
			s, err := json.Marshal(di)
			if err == nil {
				b.Put([]byte(di.ID), s)
			}
			return true
		})
		b = tx.Bucket([]byte("networks"))
		networks.Range(func(_, p interface{}) bool {
			n := p.(*NetworkEnt)
			s, err := json.Marshal(n)
			if err == nil {
				b.Put([]byte(n.ID), s)
			}
			return true
		})
		return nil
	})
	log.Printf("saveAllNodes dur=%v", time.Since(st))
	return nil
}

// SetNodeStateChanged :
func SetNodeStateChanged(id string) {
	lastNodeChanged = time.Now()
	stateChangedNodes.Store(id, true)
}

func DeleteNodeStateChanged(id string) {
	stateChangedNodes.Delete(id)
}

func ForEachStateChangedNodes(f func(string) bool) {
	stateChangedNodes.Range(func(id, _ interface{}) bool {
		return f(id.(string))
	})
}

// SaveNodeMemo:ノードに関するメモを保存する
func SaveNodeMemo(nodeID, memo string) error {
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("memo"))
		if b == nil {
			return fmt.Errorf("memo bucket not found")
		}
		return b.Put([]byte(nodeID), []byte(memo))
	})
}

// GetNodeMemo:ノードに関するメモを取得する
func GetNodeMemo(nodeID string) string {
	memo := ""
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("memo"))
		if b == nil {
			return fmt.Errorf("memo bucket not found")
		}
		if v := b.Get([]byte(nodeID)); v != nil {
			memo = strings.Clone(string(v))
		}
		return nil
	})
	return memo
}
