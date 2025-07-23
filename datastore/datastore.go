// Package datastore : データ保存
package datastore

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	gomibdb "github.com/twsnmp/go-mibdb"
	"github.com/twsnmp/twsnmpfk/i18n"
	"go.etcd.io/bbolt"
)

//go:embed conf
var conf embed.FS

var (
	db          *bbolt.DB
	dspath      string
	prevDBStats bbolt.Stats
	dbOpenTime  time.Time
	// Conf Data on Memory
	MapConf      MapConfEnt
	BackImage    BackImageEnt
	NotifyConf   NotifyConfEnt
	DiscoverConf DiscoverConfEnt
	AIConf       AIConfEnt
	LocConf      LocConfEnt
	AutoCharCode bool
	// Restart snmptrapd
	RestartSnmpTrapd bool
	// Map Data on Memory
	nodes    sync.Map
	items    sync.Map
	lines    sync.Map
	networks sync.Map
	pollings sync.Map
	// MAP Changed check
	stateChangedNodes sync.Map
	lastLogAdded      time.Time
	lastNodeChanged   time.Time
	//
	MIBDB        *gomibdb.MIBDB
	eventLogCh   chan *EventLogEnt
	pollingLogCh chan *PollingLogEnt

	protMap    map[int]string
	serviceMap map[string]string
	geoip      *geoip2.Reader
	geoipMap   sync.Map
	ouiMap     map[string]string

	logSize      int64
	compLogSize  int64
	mailTemplate map[string]string
	Dark         bool
	// for TwLogEye gRPC
	CACert     string
	ClientCert string
	ClientKey  string
	// from Command Line
	PingMode    string
	SyslogPort  int
	TrapPort    int
	NetFlowPort int
	SFlowPort   int
	TCPPort     int
	SSHdPort    int
	// OpenTelemetry
	OTelHTTPPort int
	OTelgRPCPort int
	OTelCert     string
	OTelKey      string
	OTelCA       string
)

// Define errors
var (
	ErrNoPayload     = fmt.Errorf("no payload")
	ErrInvalidNode   = fmt.Errorf("invalid node")
	ErrInvalidParams = fmt.Errorf("invald params")
	ErrDBNotOpen     = fmt.Errorf("db not open")
	ErrInvalidID     = fmt.Errorf("invalid id")
)

func Init(ctx context.Context, path string, wg *sync.WaitGroup) error {
	dspath = path
	eventLogCh = make(chan *EventLogEnt, 100)
	pollingLogCh = make(chan *PollingLogEnt, 1000)
	protMap = map[int]string{
		1:   "icmp",
		2:   "igmp",
		6:   "tcp",
		8:   "egp",
		17:  "udp",
		112: "vrrp",
	}
	serviceMap = make(map[string]string)
	ouiMap = make(map[string]string)
	if err := loadDataFromFS(); err != nil {
		return err
	}
	wg.Add(1)
	go eventLogger(ctx, wg)
	wg.Add(1)
	go oldLogChecker(ctx, wg)
	return nil
}

func loadDataFromFS() error {
	if dspath == "" {
		return fmt.Errorf("no data base path")
	}
	// BBoltをオープン
	if err := openDB(filepath.Join(dspath, "twsnmpfk.db")); err != nil {
		return err
	}
	// MIBDB
	loadMIBDB()
	// サービスの定義ファイル、ユーザー指定があれば利用、なければ内蔵
	if r, err := os.Open(filepath.Join(dspath, "services.txt")); err == nil {
		loadServiceMap(r)
	} else {
		if r, err := conf.Open("conf/services.txt"); err == nil {
			loadServiceMap(r)
		} else {
			return err
		}
	}
	// OUIの定義
	if r, err := os.Open(filepath.Join(dspath, "mac-vendors-export.csv")); err == nil {
		loadOUIMap(r)
	} else {
		if r, err := conf.Open("conf/mac-vendors-export.csv"); err == nil {
			loadOUIMap(r)
		} else {
			return err
		}
	}
	p := filepath.Join(dspath, "geoip.mmdb")
	if _, err := os.Stat(p); err == nil {
		openGeoIP(p)
	}
	lang := i18n.GetLang()
	if lang != "ja" {
		lang = "en"
	}
	if r, err := conf.Open("conf/polling_" + lang + ".json"); err == nil {
		if b, err := io.ReadAll(r); err == nil && len(b) > 0 {
			if err := loadPollingTemplate(b); err != nil {
				log.Printf("load polling template err=%v", err)
			}
		}
		r.Close()
	} else {
		log.Printf("open polling template err=%v", err)
	}
	if r, err := os.Open(filepath.Join(dspath, "polling.json")); err == nil {
		if b, err := io.ReadAll(r); err == nil && len(b) > 0 {
			if err := loadPollingTemplate(b); err != nil {
				log.Printf("load polling template err=%v", err)
			}
		}
		r.Close()
	}
	mailTemplate = make(map[string]string)
	loadMailTemplateToMap("test", lang)
	loadMailTemplateToMap("notify", lang)
	loadMailTemplateToMap("report", lang)
	log.Println("loadArpTable")
	loadArpTable()
	return nil
}

func loadMailTemplateToMap(t, lang string) {
	if r, err := os.Open(filepath.Join(dspath, "mail_"+t+".html")); err == nil {
		if b, err := io.ReadAll(r); err == nil && len(b) > 0 {
			log.Printf("load mail template=%s", t)
			mailTemplate[t] = string(b)
		}
		r.Close()
	}
	if r, err := conf.Open("conf/mail_" + t + "_" + lang + ".html"); err == nil {
		if b, err := io.ReadAll(r); err == nil && len(b) > 0 {
			log.Printf("load defalut mail template=%s_%s", t, lang)
			mailTemplate[t] = string(b)
		}
		r.Close()
	}
}

func openDB(path string) error {
	log.Println("start openDB")
	var err error
	db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	log.Println("db.Stats")
	prevDBStats = db.Stats()
	dbOpenTime = time.Now()
	log.Println("initDB")
	err = initDB()
	if err != nil {
		db.Close()
		return err
	}
	log.Println("loadConf")
	err = loadConf()
	if err != nil {
		db.Close()
		return err
	}
	log.Println("loadMapData")
	err = loadMapData()
	if err != nil {
		db.Close()
		return err
	}
	log.Println("loadPKI")
	if err = loadPKIConf(); err != nil {
		log.Printf("load PKI err=%v", err)
	}
	log.Println("load cert monitor")
	if err = loadCertMonitor(); err != nil {
		log.Printf("load cert monitor err=%v", err)
	}
	log.Println("end openDB")
	return nil
}

func initDB() error {
	buckets := []string{
		"config", "nodes", "items", "lines", "networks", "pollings", "logs", "pollingLogs",
		"syslog", "trap", "arplog", "arp", "ai", "grok", "images",
		"ipfix", "netflow",
		"sflow", "sflowCounter", "certs",
		"memo", "otelTrace", "certMonitor",
	}
	initConf()
	return db.Update(func(tx *bbolt.Tx) error {
		for _, b := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// CloseDB : DBをクローズする
func CloseDB() {
	closeGeoIP()
	if db == nil {
		return
	}
	if err := saveAllNodes(); err != nil {
		log.Printf("saveAllNodes err=%v", err)
	}
	if err := saveAllPollings(); err != nil {
		log.Printf("saveAllPollings err=%v", err)
	}
	saveArpTable()
	db.Close()
	db = nil
}

func GetDBSize() int64 {
	if db == nil {
		return 0
	}
	var dbSize int64
	db.View(func(tx *bbolt.Tx) error {
		dbSize = tx.Size()
		return nil
	})
	return dbSize
}

// SaveMapData saves the map data to the DB every 24 hours.
func SaveMapData() {
	if db == nil {
		return
	}
	if err := saveAllNodes(); err != nil {
		log.Printf("saveAllNodes err=%v", err)
	}
	if err := saveAllPollings(); err != nil {
		log.Printf("saveAllPollings err=%v", err)
	}
}

// bboltに保存する場合のキーを時刻から生成する。
func makeKey() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}

// GetDataStorePath returns the path to the data store.
func GetDataStorePath() string {
	return dspath
}
