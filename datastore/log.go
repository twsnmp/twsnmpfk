package datastore

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

const (
	LogModeNone = iota
	LogModeAlways
	LogModeOnChange
	LogModeAI
)

type EventLogEnt struct {
	Time      int64  `json:"Time"`
	Type      string `json:"Type"`
	Level     string `json:"Level"`
	NodeName  string `json:"NodeName"`
	NodeID    string `json:"NodeID"`
	Event     string `json:"Event"`
	LastLevel string `json:"LastLevel"`
}

type LogEnt struct {
	Time int64 // UnixNano()
	Type string
	Log  string
}

type LogFilterEnt struct {
	StartTime string
	EndTime   string
	Filter    string
	LogType   string
}

func AddEventLog(e *EventLogEnt) {
	e.Time = time.Now().UnixNano()
	if e.NodeID != "" && e.NodeName == "" {
		// Node IDのみの場合は、名前をここで解決する
		if n := GetNode(e.NodeID); n != nil {
			e.NodeName = n.Name
		}
	}
	eventLogCh <- e
}

func ForEachEventLog(st, et int64, f func(*EventLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	sk := fmt.Sprintf("%016x", st)
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Seek([]byte(sk)); k != nil; k, v = c.Next() {
			var e EventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				continue
			}
			if e.Time < st {
				continue
			}
			if e.Time > et {
				break
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func ForEachLastEventLog(f func(*EventLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			var e EventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				continue
			}
			if !f(&e) {
				break
			}
		}
		return nil
	})
}

func deleteOldLog(tx *bbolt.Tx, bucket string, days int) bool {
	s := time.Now()
	done := true
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		log.Printf("bucket %s not found", bucket)
		// bucketがないのは、エラーにしないでスキップする
		return done
	}
	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		if st < string(k) {
			break
		}
		if delCount > 1000 {
			done = false
			break
		}
		_ = c.Delete()
		delCount++
	}
	if delCount > 0 {
		log.Printf("delete old logs bucket=%s count=%d done=%v dur=%s", bucket, delCount, done, time.Since(s))
	}
	return done
}

func deleteOldPollingLog(tx *bbolt.Tx, bucket string, days int) bool {
	s := time.Now()
	done := true
	delCount := 0
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		log.Printf("bucket %s not found", bucket)
		// bucketがないのは、エラーにしないでスキップする
		return done
	}
	b.ForEachBucket(func(k []byte) error {
		b2 := b.Bucket(k)
		if b2 == nil {
			return nil
		}
		c := b2.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if st < string(k) {
				break
			}
			if delCount > 1000 {
				done = false
				break
			}
			_ = c.Delete()
			delCount++
		}
		return nil
	})
	if delCount > 0 {
		log.Printf("delete old logs bucket=%s count=%d done=%v dur=%s", bucket, delCount, done, time.Since(s))
	}
	return done
}

func deleteOldLogs() {
	s := time.Now()
	if MapConf.LogDays < 1 {
		log.Println("mapConf.LogDays < 1 ")
		return
	}
	tx, err := db.Begin(true)
	if err != nil {
		log.Printf("deleteOldLog err=%v", err)
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "arplog"}
	doneMap := make(map[string]bool)
	doneCount := 0
	lt := time.Now().Unix() + 50
	for doneCount < len(buckets) && lt > time.Now().Unix() {
		for _, b := range buckets {
			if _, ok := doneMap[b]; !ok {
				if b == "pollingLogs" {
					if done := deleteOldPollingLog(tx, b, MapConf.LogDays); done {
						doneMap[b] = true
						doneCount++
					}
				} else {
					if done := deleteOldLog(tx, b, MapConf.LogDays); done {
						doneMap[b] = true
						doneCount++
					}
				}
			}
			tx.Commit()
			tx, err = db.Begin(true)
			if err != nil {
				log.Printf("deleteOldLog err=%v", err)
				return
			}
		}
	}
	tx.Commit()
	log.Printf("deleteOldLogs dur=%s", time.Since(s))
}

func DeleteAllLogs(b string) error {
	return db.Batch(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket([]byte(b)); err != nil {
			return err
		}
		tx.CreateBucketIfNotExists([]byte(b))
		return nil
	})
}

func eventLogger(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start eventlog")
	timer := time.NewTicker(time.Second * 10)
	eventLogList := []*EventLogEnt{}
	pollingLogList := []*PollingLogEnt{}
	for {
		select {
		case <-ctx.Done():
			if len(eventLogList) > 0 {
				saveLogList(eventLogList)
			}
			if len(pollingLogList) > 0 {
				savePollingLogList(pollingLogList)
			}
			timer.Stop()
			log.Println("stop eventlog")
			return
		case e := <-eventLogCh:
			eventLogList = append(eventLogList, e)
		case e := <-pollingLogCh:
			pollingLogList = append(pollingLogList, e)
		case <-timer.C:
			if len(eventLogList) > 0 {
				saveLogList(eventLogList)
				eventLogList = []*EventLogEnt{}
			}
			if len(pollingLogList) > 0 {
				savePollingLogList(pollingLogList)
				pollingLogList = []*PollingLogEnt{}
			}
		}
	}
}

func oldLogChecker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start old log checker")
	timer := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop old log checker")
			return
		case <-timer.C:
			deleteOldLogs()
		}
	}
}

func saveLogList(list []*EventLogEnt) {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		for i, e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			err = b.Put([]byte(fmt.Sprintf("%016x", e.Time+int64(i))), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
	log.Printf("save event log count=%d,dur=%v", len(list), time.Since(st))
}

func savePollingLogList(list []*PollingLogEnt) {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		for i, e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			bs, err := b.CreateBucketIfNotExists([]byte(e.PollingID))
			if err != nil {
				return err
			}
			err = bs.Put([]byte(fmt.Sprintf("%016x", e.Time+int64(i))), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
	log.Printf("save polling log count=%d,dur=%v", len(list), time.Since(st))
}

func SaveLogBuffer(logBuffer []*LogEnt) {
	if db == nil {
		return
	}
	st := time.Now()
	db.Batch(func(tx *bbolt.Tx) error {
		if time.Since(st) > time.Duration(time.Second) {
			log.Printf("SaveLogBuffer batch over 1sec dur=%v", time.Since(st))
		}
		syslog := tx.Bucket([]byte("syslog"))
		trap := tx.Bucket([]byte("trap"))
		arp := tx.Bucket([]byte("arplog"))
		sc := 0
		nfc := 0
		tc := 0
		ac := 0
		oc := 0
		for i, l := range logBuffer {
			k := fmt.Sprintf("%016x", l.Time+int64(i))
			s, err := json.Marshal(l)
			if err != nil {
				return err
			}
			logSize += int64(len(s))
			if len(s) > 100 {
				s = compressLog(s)
			}
			compLogSize += int64(len(s))
			switch l.Type {
			case "syslog":
				sc++
				syslog.Put([]byte(k), []byte(s))
			case "trap":
				tc++
				trap.Put([]byte(k), []byte(s))
			case "arplog":
				ac++
				arp.Put([]byte(k), []byte(s))
			default:
				oc++
			}
		}
		log.Printf("syslog=%d,netflow=%d,trap=%d,arplog=%d,other=%d,dur=%v", sc, nfc, tc, ac, oc, time.Since(st))
		return nil
	})
}

func compressLog(s []byte) []byte {
	var b bytes.Buffer
	f, _ := flate.NewWriter(&b, flate.DefaultCompression)
	if _, err := f.Write(s); err != nil {
		return s
	}
	if err := f.Flush(); err != nil {
		return s
	}
	if err := f.Close(); err != nil {
		return s
	}
	return b.Bytes()
}

func deCompressLog(s []byte) []byte {
	r := flate.NewReader(bytes.NewBuffer(s))
	d, err := io.ReadAll(r)
	if err != nil {
		return s
	}
	return d
}

// for syslog
type SyslogEnt struct {
	Time     int64  `json:"Time"`
	Level    string `json:"Level"`
	Host     string `json:"Host"`
	Type     string `json:"Type"`
	Tag      string `json:"Tag"`
	Message  string `json:"Message"`
	Severity int    `json:"Severity"`
	Facility int    `json:"Facility"`
}

var severityNames = []string{
	"emerg",
	"alert",
	"crit",
	"err",
	"warning",
	"notice",
	"info",
	"debug",
}

var facilityNames = []string{
	"kern",
	"user",
	"mail",
	"daemon",
	"auth",
	"syslog",
	"lpr",
	"news",
	"uucp",
	"cron",
	"authpriv",
	"ftp",
	"ntp",
	"logaudit",
	"logalert",
	"clock",
	"local0",
	"local1",
	"local2",
	"local3",
	"local4",
	"local5",
	"local6",
	"local7",
}

func getSyslogType(sv, fac int) string {
	r := ""
	if sv >= 0 && sv < len(severityNames) {
		r += severityNames[sv]
	} else {
		r += "unknown"
	}
	r += ":"
	if fac >= 0 && fac < len(facilityNames) {
		r += facilityNames[fac]
	} else {
		r += "unknown"
	}
	return r
}

func getLevelFromSeverity(sv int) string {
	if sv < 3 {
		return "high"
	}
	if sv < 4 {
		return "low"
	}
	if sv == 4 {
		return "warn"
	}
	if sv == 7 {
		return "debug"
	}
	return "info"
}

// ForEachLastSyslog  get syslogs
func ForEachLastSyslog(f func(*SyslogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("syslog"))
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
			var sl = make(map[string]interface{})
			if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
				continue
			}
			var ok bool
			re := new(SyslogEnt)
			var sv float64
			if sv, ok = sl["severity"].(float64); !ok {
				continue
			}
			var fac float64
			if fac, ok = sl["facility"].(float64); !ok {
				continue
			}
			if re.Host, ok = sl["hostname"].(string); !ok {
				continue
			}
			if re.Tag, ok = sl["tag"].(string); !ok {
				if re.Tag, ok = sl["app_name"].(string); !ok {
					continue
				}
				re.Message = ""
				for i, k := range []string{"proc_id", "msg_id", "message", "structured_data"} {
					if m, ok := sl[k].(string); ok && m != "" {
						if i > 0 {
							re.Message += " "
						}
						re.Message += m
					}
				}
			} else {
				if re.Message, ok = sl["content"].(string); !ok {
					continue
				}
			}
			re.Time = l.Time
			re.Level = getLevelFromSeverity(int(sv))
			re.Type = getSyslogType(int(sv), int(fac))
			re.Facility = int(fac)
			re.Severity = int(sv)
			if !f(re) {
				break
			}
		}
		return nil
	})
}

type TrapEnt struct {
	Time        int64  `json:"Time"`
	FromAddress string `json:"FromAddress"`
	TrapType    string `json:"TrapType"`
	Variables   string `json:"Variables"`
}

var trapOidRegexp = regexp.MustCompile(`snmpTrapOID.0=(\S+)`)

// ForEachLastTraps  get TRAP
func ForEachLastTraps(f func(*TrapEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("trap"))
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
			var sl = make(map[string]interface{})
			if err := json.Unmarshal([]byte(l.Log), &sl); err != nil {
				continue
			}
			var ok bool
			re := new(TrapEnt)
			if fa, ok := sl["FromAddress"].(string); !ok {
				continue
			} else {
				a := strings.SplitN(fa, ":", 2)
				if len(a) == 2 {
					re.FromAddress = a[0]
					n := FindNodeFromIP(a[0])
					if n != nil {
						re.FromAddress += "(" + n.Name + ")"
					}
				} else {
					re.FromAddress = fa
				}
			}
			if re.Variables, ok = sl["Variables"].(string); !ok {
				continue
			}
			var ent string
			if ent, ok = sl["Enterprise"].(string); !ok || ent == "" {
				a := trapOidRegexp.FindStringSubmatch(re.Variables)
				if len(a) > 1 {
					re.TrapType = a[1]
				} else {
					re.TrapType = ""
				}
			} else {
				var gen float64
				if gen, ok = sl["GenericTrap"].(float64); !ok {
					continue
				}
				var spe float64
				if spe, ok = sl["SpecificTrap"].(float64); !ok {
					continue
				}
				re.TrapType = fmt.Sprintf("%s:%d:%d", ent, int(gen), int(spe))
			}
			re.Time = l.Time
			if !f(re) {
				break
			}
		}
		return nil
	})
}

type ArpLogEnt struct {
	Time   int64  `json:"Time"`
	State  string `json:"State"`
	IP     string `json:"IP"`
	OldMAC string `json:"OldMAC"`
	NewMAC string `json:"NewMAC"`
}

// ForEachLastArpLogs は最新のARP Logを返します。
func ForEachLastArpLogs(f func(*ArpLogEnt) bool) error {
	if db == nil {
		return ErrDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arplog"))
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
			st := a[0]
			ip := a[1]
			newMac := a[2]
			oldMac := ""
			if len(a) > 3 {
				oldMac = a[3]
			}
			if !f(&ArpLogEnt{
				Time:   l.Time,
				State:  st,
				IP:     ip,
				NewMAC: newMac,
				OldMAC: oldMac,
			}) {
				break
			}
		}
		return nil
	})
}
