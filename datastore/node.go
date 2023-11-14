package datastore

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type NodeEnt struct {
	ID        string `json:"ID"`
	Name      string `json:"Name"`
	Descr     string `json:"Descr"`
	Icon      string `json:"Icon"`
	State     string `json:"State"`
	X         int    `json:"X"`
	Y         int    `json:"Y"`
	IP        string `json:"IP"`
	MAC       string `json:"MAC"`
	Vendor    string `json:"Vendor"`
	SnmpMode  string `json:"SnmpMode"`
	Community string `json:"Community"`
	User      string `json:"User"`
	Password  string `json:"Password"`
	PublicKey string `json:"PublicKey"`
	URL       string `json:"URL"`
	AddrMode  string `json:"AddrMode"`
	AutoAck   bool   `json:"AutoAck"`
	Loc       string `json:"Loc"`
}

type DrawItemType int

const (
	DrawItemTypeRect = iota
	DrawItemTypeEllipse
	DrawItemTypeText
	DrawItemTypeImage
	DrawItemTypePollingText
	DrawItemTypePollingGauge
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

func DeleteNode(nodeID string) error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	if _, ok := nodes.Load(nodeID); !ok {
		return ErrInvalidID
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
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
	log.Printf("DeleteNode dur=%v", time.Since(st))
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

func GetNode(nodeID string) *NodeEnt {
	if db == nil {
		return nil
	}
	if n, ok := nodes.Load(nodeID); ok {
		return n.(*NodeEnt)
	}
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

func ForEachItems(f func(*DrawItemEnt) bool) {
	items.Range(func(_, p interface{}) bool {
		return f(p.(*DrawItemEnt))
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
