package datastore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfk/i18n"
	"go.etcd.io/bbolt"
)

// SensorStatsEnt represents periodic message rate / count telemetry for a sensor.
type SensorStatsEnt struct {
	Time     int64   `json:"Time"` // unix nano
	Total    int64   `json:"Total"`
	Count    int64   `json:"Count"`
	PS       float64 `json:"PS"`
	Send     int64   `json:"Send"`
	LastSend int64   `json:"LastSend"`
}

// SensorMonitorEnt represents periodic system resource monitoring telemetry for a sensor host.
type SensorMonitorEnt struct {
	Time    int64   `json:"Time"` // unix nano
	CPU     float64 `json:"CPU"`
	Mem     float64 `json:"Mem"`
	Load    float64 `json:"Load"`
	Process int64   `json:"Process"`
	Recv    int64   `json:"Recv"`
	Sent    int64   `json:"Sent"`
	TxSpeed float64 `json:"TxSpeed"`
	RxSpeed float64 `json:"RxSpeed"`
}

// SensorEnt represents a remote sensor reporting to TWSNMP (syslog, twWifiScan, twBlueScan, sflow, netflow, mqtt, etc.).
type SensorEnt struct {
	ID          string             `json:"ID"`
	Host        string             `json:"Host"`
	Type        string             `json:"Type"`
	Param       string             `json:"Param"`
	Total       int64              `json:"Total"`
	Send        int64              `json:"Send"`
	State       string             `json:"State"`
	Ignore      bool               `json:"Ignore"`
	Stats       []SensorStatsEnt   `json:"Stats,omitempty"`
	Monitors    []SensorMonitorEnt `json:"Monitors,omitempty"`
	StatsLen    int                `json:"StatsLen"`
	MonitorsLen int                `json:"MonitorsLen"`
	FirstTime   int64              `json:"FirstTime"` // unix nano
	LastTime    int64              `json:"LastTime"`  // unix nano
}

var (
	sensors sync.Map // string -> *SensorEnt
)

// makeSensorID returns a deterministic hex hash ID matching twsnmpneo/twsnmpfc.
func makeSensorID(host, sensorType, param string) string {
	raw := fmt.Sprintf("%s:%s:%s", host, sensorType, param)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func loadSensors() error {
	if db == nil {
		return nil
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sensors"))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var sn SensorEnt
			if err := json.Unmarshal(v, &sn); err == nil {
				sn.StatsLen = len(sn.Stats)
				sn.MonitorsLen = len(sn.Monitors)
				sensors.Store(sn.ID, &sn)
			}
			return nil
		})
	})
}

// ListSensors returns all sensors sorted by LastTime descending.
func ListSensors() []*SensorEnt {
	var list []*SensorEnt
	sensors.Range(func(_, v any) bool {
		if sn, ok := v.(*SensorEnt); ok {
			cp := *sn
			cp.StatsLen = len(sn.Stats)
			cp.MonitorsLen = len(sn.Monitors)
			list = append(list, &cp)
		}
		return true
	})
	sort.Slice(list, func(i, j int) bool {
		return list[i].LastTime > list[j].LastTime
	})
	return list
}

// GetSensor returns a single sensor by ID.
func GetSensor(id string) (*SensorEnt, error) {
	if v, ok := sensors.Load(id); ok {
		if sn, ok := v.(*SensorEnt); ok {
			cp := *sn
			cp.StatsLen = len(sn.Stats)
			cp.MonitorsLen = len(sn.Monitors)
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("sensor not found")
}

// SaveSensor persists a sensor to memory cache and bbolt.
func SaveSensor(sensor *SensorEnt) error {
	if sensor == nil || sensor.ID == "" {
		return fmt.Errorf("invalid sensor")
	}
	sensor.StatsLen = len(sensor.Stats)
	sensor.MonitorsLen = len(sensor.Monitors)
	sensors.Store(sensor.ID, sensor)

	if db == nil {
		return nil
	}
	data, err := json.Marshal(sensor)
	if err != nil {
		return fmt.Errorf("marshal sensor: %w", err)
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sensors"))
		if b == nil {
			return fmt.Errorf("sensors bucket not found")
		}
		return b.Put([]byte(sensor.ID), data)
	})
}

// DeleteSensors removes specific sensors by IDs.
func DeleteSensors(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		sensors.Delete(id)
	}

	if db == nil {
		return nil
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sensors"))
		if b == nil {
			return nil
		}
		for _, id := range ids {
			_ = b.Delete([]byte(id))
		}
		return nil
	})
}

// DeleteAllSensors purges all sensors from memory and disk.
func DeleteAllSensors() error {
	sensors.Range(func(k, _ any) bool {
		sensors.Delete(k)
		return true
	})

	if db == nil {
		return nil
	}
	return db.Update(func(tx *bbolt.Tx) error {
		_ = tx.DeleteBucket([]byte("sensors"))
		_, err := tx.CreateBucketIfNotExists([]byte("sensors"))
		return err
	})
}

// ToggleSensorIgnore toggles the Ignore state of a sensor.
func ToggleSensorIgnore(id string) error {
	v, ok := sensors.Load(id)
	if !ok {
		return fmt.Errorf("sensor not found")
	}
	sn, ok := v.(*SensorEnt)
	if !ok {
		return fmt.Errorf("invalid sensor entity")
	}

	sn.Ignore = !sn.Ignore
	if sn.Ignore {
		sn.State = "off"
	} else {
		sn.State = "normal"
	}
	sensors.Store(sn.ID, sn)

	if db == nil {
		return nil
	}
	data, err := json.Marshal(sn)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("sensors"))
		if b == nil {
			return fmt.Errorf("sensors bucket not found")
		}
		return b.Put([]byte(sn.ID), data)
	})
}

// UpdateSensor updates or registers a sensor receiving data.
func UpdateSensor(host, sensorType, param string, count int64) {
	if host == "" {
		return
	}
	if count <= 0 {
		count = 1
	}
	id := makeSensorID(host, sensorType, param)
	now := time.Now().UnixNano()

	if v, ok := sensors.Load(id); ok {
		sn := v.(*SensorEnt)
		sn.Total += count
		sn.Send++
		sn.LastTime = now
		if !sn.Ignore && sn.State != "normal" && sn.State != "off" {
			sn.State = "normal"
		}
		return
	}

	// Also check fallback without param if param is empty
	if param == "" {
		idOld := fmt.Sprintf("%s:%s:", host, sensorType)
		if v, ok := sensors.Load(idOld); ok {
			sn := v.(*SensorEnt)
			sn.Total += count
			sn.Send++
			sn.LastTime = now
			return
		}
	}

	newSn := &SensorEnt{
		ID:        id,
		Host:      host,
		Type:      sensorType,
		Param:     param,
		Total:     count,
		Send:      1,
		State:     "normal",
		FirstTime: now,
		LastTime:  now,
	}
	sensors.Store(id, newSn)

	// Async persist to bbolt
	if db != nil {
		go func(ent *SensorEnt) {
			data, err := json.Marshal(ent)
			if err == nil {
				_ = db.Update(func(tx *bbolt.Tx) error {
					b := tx.Bucket([]byte("sensors"))
					if b != nil {
						return b.Put([]byte(ent.ID), data)
					}
					return nil
				})
			}
		}(newSn)
	}
}

// CheckSensorStats appends a stats telemetry sample to a sensor.
func CheckSensorStats(host, sensorType, param string, count, send, total int64, ps float64) {
	if host == "" {
		return
	}
	if send < 1 {
		send = 1
	}
	id := makeSensorID(host, sensorType, param)
	now := time.Now().UnixNano()

	statItem := SensorStatsEnt{
		Time:     now,
		Total:    total,
		Count:    count,
		Send:     send,
		LastSend: send,
		PS:       ps,
	}

	if v, ok := sensors.Load(id); ok {
		sn := v.(*SensorEnt)
		sn.Total += count
		sn.Send += send
		sn.Stats = append(sn.Stats, statItem)
		if len(sn.Stats) > 2880 {
			sn.Stats = sn.Stats[len(sn.Stats)-2880:]
		}
		sn.StatsLen = len(sn.Stats)
		sn.LastTime = now
		return
	}

	newSn := &SensorEnt{
		ID:          id,
		Host:        host,
		Type:        sensorType,
		Param:       param,
		Total:       count,
		Send:        send,
		State:       "normal",
		Stats:       []SensorStatsEnt{statItem},
		StatsLen:    1,
		FirstTime:   now,
		LastTime:    now,
	}
	sensors.Store(id, newSn)
}

// CheckSensorMonitor appends a system resource monitor sample to a sensor.
func CheckSensorMonitor(host, sensorType, param string, cpu, mem, load, txSpeed, rxSpeed float64, sent, recv, proc int64) {
	if host == "" {
		return
	}
	id := makeSensorID(host, sensorType, param)
	now := time.Now().UnixNano()

	monItem := SensorMonitorEnt{
		Time:    now,
		CPU:     cpu,
		Mem:     mem,
		Load:    load,
		TxSpeed: txSpeed,
		RxSpeed: rxSpeed,
		Sent:    sent,
		Recv:    recv,
		Process: proc,
	}

	if v, ok := sensors.Load(id); ok {
		sn := v.(*SensorEnt)
		sn.Monitors = append(sn.Monitors, monItem)
		if len(sn.Monitors) > 2880 {
			sn.Monitors = sn.Monitors[len(sn.Monitors)-2880:]
		}
		sn.MonitorsLen = len(sn.Monitors)
		sn.LastTime = now
		return
	}

	newSn := &SensorEnt{
		ID:          id,
		Host:        host,
		Type:        sensorType,
		Param:       param,
		Total:       1,
		Send:        1,
		State:       "normal",
		Monitors:    []SensorMonitorEnt{monItem},
		MonitorsLen: 1,
		FirstTime:   now,
		LastTime:    now,
	}
	sensors.Store(id, newSn)
}

// EvaluateSensorStates checks sensor timeouts and updates sensor states.
func EvaluateSensorStates(ctx context.Context) {
	now := time.Now().UnixNano()
	to := 1 // default 1 hour timeout
	warnCutoff := time.Now().Add(-time.Duration(to) * time.Hour).UnixNano()
	lowCutoff := time.Now().Add(-time.Duration(to*6) * time.Hour).UnixNano()
	highCutoff := time.Now().Add(-time.Duration(to*24) * time.Hour).UnixNano()

	var dirtySensors []*SensorEnt

	sensors.Range(func(_, v any) bool {
		sn, ok := v.(*SensorEnt)
		if !ok {
			return true
		}
		oldState := sn.State

		if sn.Ignore {
			sn.State = "off"
		} else if sn.LastTime < highCutoff {
			sn.State = "high"
		} else if sn.LastTime < lowCutoff {
			sn.State = "low"
		} else if sn.LastTime < warnCutoff {
			sn.State = "warn"
		} else {
			sn.State = "normal"
		}

		if oldState != sn.State {
			eventMsg := fmt.Sprintf(i18n.Trans("Change sensor state:%s,%s,%s"), sn.Host, sn.Type, sn.Param)
			AddEventLog(&EventLogEnt{
				Type:     "sensor",
				Level:    sn.State,
				NodeName: sn.Host,
				Event:    eventMsg,
			})
			log.Printf("Sensor state changed: host=%s, type=%s, state=%s", sn.Host, sn.Type, sn.State)
		}

		// Calculate periodic stats for continuous feed protocols (syslog, netflow, sflow)
		if sn.Type == "netflow" || sn.Type == "ipfix" || sn.Type == "syslog" || strings.HasPrefix(sn.Type, "otel") {
			count := sn.Total
			send := sn.Send
			if len(sn.Stats) > 0 {
				last := sn.Stats[len(sn.Stats)-1]
				count -= last.Total
				send -= last.LastSend
			}
			if count < 0 {
				count = 0
			}
			if send < 0 {
				send = 0
			}
			sn.Stats = append(sn.Stats, SensorStatsEnt{
				Time:     now,
				Total:    sn.Total,
				Count:    count,
				Send:     send,
				LastSend: sn.Send,
				PS:       float64(count) / 300.0, // 5 min interval rate
			})
			if len(sn.Stats) > 2880 {
				sn.Stats = sn.Stats[len(sn.Stats)-2880:]
			}
			sn.StatsLen = len(sn.Stats)
		}

		dirtySensors = append(dirtySensors, sn)
		return true
	})

	// Batch persist updated sensors
	if len(dirtySensors) > 0 && db != nil {
		_ = db.Batch(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("sensors"))
			if b == nil {
				return nil
			}
			for _, sn := range dirtySensors {
				data, err := json.Marshal(sn)
				if err == nil {
					_ = b.Put([]byte(sn.ID), data)
				}
			}
			return nil
		})
	}
}
