package datastore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"encoding/pem"

	"github.com/twsnmp/twsnmpfk/i18n"
	"go.etcd.io/bbolt"
	"golang.org/x/crypto/ssh"
)

type BackImageEnt struct {
	X      int    `json:"X"`
	Y      int    `json:"Y"`
	Width  int    `json:"Width"`
	Height int    `json:"Height"`
	Path   string `json:"Path"`
}

// MapConfEnt :  マップ設定
type MapConfEnt struct {
	MapName        string `json:"MapName"`
	PollInt        int    `json:"PollInt"`
	Timeout        int    `json:"Timeout"`
	Retry          int    `json:"Retry"`
	LogDays        int    `json:"LogDays"`
	SnmpMode       string `json:"SnmpMode"`
	Community      string `json:"Community"`
	SnmpUser       string `json:"SnmpUser"`
	SnmpPassword   string `json:"SnmpPassword"`
	EnableSyslogd  bool   `json:"EnableSyslogd"`
	EnableTrapd    bool   `json:"EnableTrapd"`
	EnableArpWatch bool   `json:"EnableArpWatch"`
	EnableNetflowd bool   `json:"EnableNetflowd"`
	EnableSshd     bool   `json:"EnableSshd"`
	EnableSFlowd   bool   `json:"EnableSFlowd"`
	EnableTcpd     bool   `json:"EnableTcpd"`
	IconSize       int    `json:"IconSize"`
	MapSize        int    `json:"MapSize"`
}

// LocConfEnt : 地図設定
type LocConfEnt struct {
	Style    string  `json:"Style"`
	Center   string  `json:"Center"`
	Zoom     float64 `json:"Zoom"`
	IconSize int     `json:"IconSize"`
}

func initConf() {
	MapConf.PollInt = 60
	MapConf.Retry = 1
	MapConf.Timeout = 1
	MapConf.LogDays = 14
	MapConf.Community = "public"
	MapConf.SnmpMode = "v2c"
	MapConf.EnableArpWatch = true
	MapConf.IconSize = 2
	DiscoverConf.AddPolling = true
	DiscoverConf.Retry = 1
	DiscoverConf.Timeout = 1
	NotifyConf.InsecureSkipVerify = true
	NotifyConf.Interval = 60
	NotifyConf.Subject = i18n.Trans("Notify from TWSNMP")
	NotifyConf.Level = "none"
	LocConf.Zoom = 2
	LocConf.Center = "139.75,35.68"
	LocConf.IconSize = 24
}

func loadConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	bSaveConf := false
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("mapConf"))
		if v == nil {
			bSaveConf = true
			return nil
		}
		if err := json.Unmarshal(v, &MapConf); err != nil {
			bSaveConf = true
			return err
		}
		v = b.Get([]byte("backImage"))
		if v != nil {
			json.Unmarshal(v, &BackImage)
		}
		v = b.Get([]byte("discoverConf"))
		if v != nil {
			json.Unmarshal(v, &DiscoverConf)
		}
		v = b.Get([]byte("notifyConf"))
		if v != nil {
			json.Unmarshal(v, &NotifyConf)
		}
		v = b.Get([]byte("aiConf"))
		if v != nil {
			json.Unmarshal(v, &AIConf)
		}
		v = b.Get([]byte("locConf"))
		if v != nil {
			json.Unmarshal(v, &LocConf)
		}
		v = b.Get([]byte("icons"))
		if v != nil {
			if err := json.Unmarshal(v, &icons); err != nil {
				log.Printf("load icons err=%v", err)
			}
		}
		v = b.Get([]byte("dark"))
		if v != nil && string(v) == "yes" {
			Dark = true
		}
		return nil
	})
	if err == nil && bSaveConf {
		if err := SaveMapConf(); err != nil {
			log.Printf("save map conf err=%v", err)
		}
		if err := SaveNotifyConf(); err != nil {
			log.Printf("save notify conf err=%v", err)
		}
		if err := SaveDiscoverConf(); err != nil {
			log.Printf("save discover conf err=%v", err)
		}
		if err := SaveAIConf(); err != nil {
			log.Printf("save ai conf err=%v", err)
		}
		if err := SaveLocConf(); err != nil {
			log.Printf("save loc conf err=%v", err)
		}
	}
	return err
}

var imageListCache = []string{}

func SaveImage(path string, img []byte) error {
	st := time.Now()
	imageListCache = []string{}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveImage dur=%v", time.Since(st))
		return b.Put([]byte(path), img)
	})
}

func DelteImage(path string) error {
	st := time.Now()
	imageListCache = []string{}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("DeleteImage dur=%v", time.Since(st))
		return b.Delete([]byte(path))
	})
}

func GetImageList() []string {
	if db == nil || len(imageListCache) > 0 {
		return imageListCache
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket iamges is nil")
		}
		return b.ForEach(func(k, v []byte) error {
			imageListCache = append(imageListCache, string(k))
			return nil
		})
	})
	return imageListCache
}

func GetImage(path string) ([]byte, error) {
	var r []byte
	if db == nil {
		return r, ErrDBNotOpen
	}
	return r, db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("images"))
		if b == nil {
			return fmt.Errorf("bucket iamges is nil")
		}
		r = b.Get([]byte(path))
		return nil
	})
}

func SaveMapConf() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(MapConf)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveMapConf dur=%v", time.Since(st))
		return b.Put([]byte("mapConf"), s)
	})
}

func SaveBackImage() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(BackImage)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveBackImage dur=%v", time.Since(st))
		return b.Put([]byte("backImage"), s)
	})
}

type IconEnt struct {
	Name string `json:"Name"`
	Icon string `json:"Icon"`
	Code int64  `json:"Code"`
}

var icons []*IconEnt

func GetIcons() []*IconEnt {
	return icons
}

func AddOrUpdateIcon(i *IconEnt) error {
	for _, e := range icons {
		if e.Icon == i.Icon {
			e.Name = i.Name
			e.Code = i.Code
			return saveIcons()
		}
	}
	icons = append(icons, i)
	return saveIcons()
}

func DeleteIcon(icon string) error {
	tmp := icons
	icons = []*IconEnt{}
	for _, i := range tmp {
		if i.Icon != icon {
			icons = append(icons, i)
		}
	}
	return saveIcons()
}

func saveIcons() error {
	st := time.Now()
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(icons)
	if err != nil {
		return err
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("saveIcons dur=%v", time.Since(st))
		return b.Put([]byte("icons"), s)
	})
}

func SetDark(dark bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	Dark = dark
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		if dark {
			return b.Put([]byte("dark"), []byte("yes"))
		} else {
			b.Delete([]byte("dark"))
		}
		return nil
	})
}

func SaveLocConf() error {
	if db == nil {
		return ErrDBNotOpen
	}
	s, err := json.Marshal(LocConf)
	if err != nil {
		return err
	}
	st := time.Now()
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		log.Printf("SaveLocConf dur=%v", time.Since(st))
		return b.Put([]byte("locConf"), s)
	})
}

func GetSshdPublicKeys() string {
	r := ""
	if db == nil {
		return r
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		r = string(b.Get([]byte("sshdPublicKeys")))
		return nil
	})
	return r
}

func SaveSshdPublicKeys(pk string) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("sshdPublicKeys"), []byte(pk))
	})
}

func GetPrivateKey(pm bool) []byte {
	var kb []byte
	if db == nil {
		return kb
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		kb = b.Get([]byte("sshdPrivateKey"))
		return nil
	})
	if len(kb) < 1 {
		kb = GenSSHPrivateKey()
	}
	if !pm {
		return kb
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: kb,
	}
	return pem.EncodeToMemory(block)
}

func GenSSHPrivateKey() []byte {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Printf("genSSHPrivateKey err=%v", err)
		return []byte{}
	}
	kb := x509.MarshalPKCS1PrivateKey(key)
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("bucket config is nil")
		}
		return b.Put([]byte("sshdPrivateKey"), kb)
	})
	return kb
}

func GetSSHPublicKey() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	comment := fmt.Sprintf("twsnmp@%s", host)
	kb := GetPrivateKey(false)

	priv, err := x509.ParsePKCS1PrivateKey(kb)
	if err != nil {
		return "", fmt.Errorf("key not found")
	}
	rsaKey := priv.PublicKey
	pubkey, _ := ssh.NewPublicKey(&rsaKey)
	return fmt.Sprintf("%s %s", strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pubkey))), comment), nil
}
