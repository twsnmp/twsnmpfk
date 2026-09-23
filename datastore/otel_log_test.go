package datastore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.etcd.io/bbolt"
)

func TestForEachLastOTelLog_Parquet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otel_parquet_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewParquetLogDataStore()
	if err := store.Open(tmpDir); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer store.Close()

	origStore := logStore
	origFormat := MapConf.LogFormat
	logStore = store
	MapConf.LogFormat = "parquet"
	defer func() {
		logStore = origStore
		MapConf.LogFormat = origFormat
	}()

	now := time.Now().UnixNano()
	logs := []*LogEnt{
		// 1 hour ago: normal syslog
		{Time: now - 3600*1e9, Type: "syslog", Log: `{"hostname":"host1","severity":5,"facility":16,"tag":"sshd","content":"normal log 1"}`},
		// 50 mins ago: otel log
		{Time: now - 3000*1e9, Type: "syslog", Log: `{"hostname":"host1","severity":6,"facility":16,"tag":"otel","content":"otel log 1"}`},
		// 30 mins ago: normal syslog
		{Time: now - 1800*1e9, Type: "syslog", Log: `{"hostname":"host2","severity":5,"facility":16,"tag":"nginx","content":"normal log 2"}`},
		// 10 mins ago: otel log
		{Time: now - 600*1e9, Type: "syslog", Log: `{"hostname":"host2","severity":3,"facility":16,"tag":"otel","content":"otel log 2"}`},
	}

	if err := store.SaveLogs("syslog", logs); err != nil {
		t.Fatalf("SaveLogs failed: %v", err)
	}
	if err := store.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// 1. Should find both OTel logs in reverse chronological order
	var found []*SyslogEnt
	err = ForEachLastOTelLog(now-7200*1e9, 100, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 otel logs, got %d", len(found))
	}
	if found[0].Time != now-600*1e9 || found[1].Time != now-3000*1e9 {
		t.Errorf("unexpected order: found[0]=%d, found[1]=%d", found[0].Time, found[1].Time)
	}
	if found[0].Tag != "otel" || found[1].Tag != "otel" {
		t.Errorf("expected tag otel, got %s and %s", found[0].Tag, found[1].Tag)
	}

	// 2. Start time filter: only logs newer than 20 mins ago (should only find the 10-min ago log)
	found = nil
	err = ForEachLastOTelLog(now-1200*1e9, 100, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 otel log with st filter, got %d", len(found))
	}
	if found[0].Time != now-600*1e9 {
		t.Errorf("expected log from 10 mins ago, got %d", found[0].Time)
	}

	// 3. No OTel logs in window (e.g. st is 5 mins ago)
	found = nil
	err = ForEachLastOTelLog(now-300*1e9, 100, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("expected 0 otel logs, got %d", len(found))
	}

	// 4. maxScan limit: scan at most 1 log (the newest log is normal log at -1800s, so 0 otel found)
	found = nil
	err = ForEachLastOTelLog(now-7200*1e9, 1, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	// The newest log is otel log at -600s, so scanning 1 record yields that 1 record
	if len(found) != 1 {
		t.Fatalf("expected 1 otel log with maxScan=1, got %d", len(found))
	}
}

func TestForEachLastOTelLog_Bbolt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "otel_bbolt_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	testDB, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	origDB := db
	origFormat := MapConf.LogFormat
	origStore := logStore
	db = testDB
	MapConf.LogFormat = "bbolt"
	logStore = nil
	defer func() {
		db = origDB
		MapConf.LogFormat = origFormat
		logStore = origStore
	}()

	_ = testDB.Update(func(tx *bbolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("syslog"))
		return nil
	})

	now := time.Now().UnixNano()
	logs := []*LogEnt{
		{Time: now - 3600*1e9, Type: "syslog", Log: `{"hostname":"host1","severity":5,"facility":16,"tag":"sshd","content":"normal log 1"}`},
		{Time: now - 3000*1e9, Type: "syslog", Log: `{"hostname":"host1","severity":6,"facility":16,"tag":"otel","content":"otel log 1"}`},
		{Time: now - 1800*1e9, Type: "syslog", Log: `{"hostname":"host2","severity":5,"facility":16,"tag":"nginx","content":"normal log 2"}`},
		{Time: now - 600*1e9, Type: "syslog", Log: `{"hostname":"host2","severity":3,"facility":16,"tag":"otel","content":"otel log 2"}`},
	}

	_ = testDB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("syslog"))
		for _, l := range logs {
			k := fmt.Sprintf("%016x", l.Time)
			s, _ := json.Marshal(l)
			s = compressLog(s)
			_ = b.Put([]byte(k), s)
		}
		return nil
	})

	// 1. Should find both OTel logs in reverse chronological order
	var found []*SyslogEnt
	err = ForEachLastOTelLog(now-7200*1e9, 100, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 otel logs, got %d", len(found))
	}
	if found[0].Time != now-600*1e9 || found[1].Time != now-3000*1e9 {
		t.Errorf("unexpected order: found[0]=%d, found[1]=%d", found[0].Time, found[1].Time)
	}
	if found[0].Tag != "otel" || found[1].Tag != "otel" {
		t.Errorf("expected tag otel, got %s and %s", found[0].Tag, found[1].Tag)
	}

	// 2. Start time filter: only logs newer than 20 mins ago
	found = nil
	err = ForEachLastOTelLog(now-1200*1e9, 100, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 otel log with st filter, got %d", len(found))
	}
	if found[0].Time != now-600*1e9 {
		t.Errorf("expected log from 10 mins ago, got %d", found[0].Time)
	}

	// 3. No OTel logs in window
	found = nil
	err = ForEachLastOTelLog(now-300*1e9, 100, func(l *SyslogEnt) bool {
		found = append(found, l)
		return true
	})
	if err != nil {
		t.Fatalf("ForEachLastOTelLog failed: %v", err)
	}
	if len(found) != 0 {
		t.Fatalf("expected 0 otel logs, got %d", len(found))
	}
}
