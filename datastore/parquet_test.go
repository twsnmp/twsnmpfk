package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.etcd.io/bbolt"
)

func TestParquetLogDataStore_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parquet_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewParquetLogDataStore()
	if err := store.Open(tmpDir); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer store.Close()

	now := time.Now().UnixNano()
	logs := []*LogEnt{
		{Time: now - 3000, Type: "syslog", Log: `{"hostname":"host1","severity":3,"facility":1,"content":"msg1"}`},
		{Time: now - 2000, Type: "syslog", Log: `{"hostname":"host2","severity":4,"facility":1,"content":"msg2"}`},
		{Time: now - 1000, Type: "syslog", Log: `{"hostname":"host1","severity":5,"facility":1,"content":"msg3"}`},
	}

	if err := store.SaveLogs("syslog", logs); err != nil {
		t.Fatalf("SaveLogs failed: %v", err)
	}
	if err := store.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Test ForEachLog (chronological)
	var forwardLogs []*LogEnt
	store.ForEachLog("syslog", now-4000, now, func(l *LogEnt) bool {
		forwardLogs = append(forwardLogs, l)
		return true
	})
	if len(forwardLogs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(forwardLogs))
	}
	if forwardLogs[0].Time != now-3000 || forwardLogs[2].Time != now-1000 {
		t.Errorf("unexpected log order in ForEachLog: %+v", forwardLogs)
	}

	// Test ForEachLastLog (reverse chronological)
	var reverseLogs []*LogEnt
	store.ForEachLastLog("syslog", func(l *LogEnt) bool {
		reverseLogs = append(reverseLogs, l)
		return true
	})
	if len(reverseLogs) != 3 {
		t.Fatalf("expected 3 reverse logs, got %d", len(reverseLogs))
	}
	if reverseLogs[0].Time != now-1000 || reverseLogs[2].Time != now-3000 {
		t.Errorf("unexpected reverse log order in ForEachLastLog: %+v", reverseLogs)
	}
}

func TestParquetLogDataStore_PollingLogs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parquet_polling_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewParquetLogDataStore()
	if err := store.Open(tmpDir); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer store.Close()

	now := time.Now().UnixNano()
	pLogs := []*PollingLogEnt{
		{Time: now - 3000, PollingID: "poll-1", State: "normal", Result: map[string]interface{}{"val": 1.0}},
		{Time: now - 2000, PollingID: "poll-2", State: "warn", Result: map[string]interface{}{"val": 2.0}},
		{Time: now - 1000, PollingID: "poll-1", State: "high", Result: map[string]interface{}{"val": 3.0}},
	}

	if err := store.SavePollingLogs(pLogs); err != nil {
		t.Fatalf("SavePollingLogs failed: %v", err)
	}
	if err := store.Flush(); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Test GetAllPollingLog for poll-1
	allPoll1 := store.GetAllPollingLog("poll-1")
	if len(allPoll1) != 2 {
		t.Fatalf("expected 2 logs for poll-1, got %d", len(allPoll1))
	}
	if allPoll1[0].Time != now-3000 || allPoll1[1].Time != now-1000 {
		t.Errorf("unexpected GetAllPollingLog times: %+v", allPoll1)
	}

	// Test ForEachLastPollingLog for poll-1
	var lastPoll1 []*PollingLogEnt
	store.ForEachLastPollingLog("poll-1", func(l *PollingLogEnt) bool {
		lastPoll1 = append(lastPoll1, l)
		return true
	})
	if len(lastPoll1) != 2 {
		t.Fatalf("expected 2 last logs for poll-1, got %d", len(lastPoll1))
	}
	if lastPoll1[0].Time != now-1000 || lastPoll1[1].Time != now-3000 {
		t.Errorf("unexpected ForEachLastPollingLog times: %+v", lastPoll1)
	}
}

func TestParquetLogDataStore_Cleanup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parquet_cleanup_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	store := NewParquetLogDataStore()
	if err := store.Open(tmpDir); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer store.Close()

	// Create an old date folder (10 days ago) and a today folder
	oldDate := time.Now().AddDate(0, 0, -10).UnixNano()
	oldLogs := []*LogEnt{
		{Time: oldDate, Type: "syslog", Log: `{"hostname":"old","severity":6,"facility":1,"content":"old msg"}`},
	}
	todayLogs := []*LogEnt{
		{Time: time.Now().UnixNano(), Type: "syslog", Log: `{"hostname":"today","severity":6,"facility":1,"content":"today msg"}`},
	}

	_ = store.SaveLogs("syslog", oldLogs)
	_ = store.SaveLogs("syslog", todayLogs)
	_ = store.Flush()

	oldDateStr := time.Unix(0, oldDate).Format("2006-01-02")
	oldDir := filepath.Join(tmpDir, "type=syslog", "date="+oldDateStr)
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Fatalf("expected old directory %s to exist before cleanup", oldDir)
	}

	// Run cleanup with retention 7 days
	if err := store.Cleanup(7); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Errorf("expected old directory %s to be removed by cleanup", oldDir)
	}

	todayDateStr := time.Now().Format("2006-01-02")
	todayDir := filepath.Join(tmpDir, "type=syslog", "date="+todayDateStr)
	if _, err := os.Stat(todayDir); os.IsNotExist(err) {
		t.Errorf("expected today directory %s to remain after cleanup", todayDir)
	}
}

func TestMigrateBboltToParquet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrate_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "twsnmpfk.db")
	dspath = tmpDir

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := openDB(dbPath); err != nil {
		t.Fatalf("openDB failed: %v", err)
	}
	defer CloseDB()

	// Populate some logs in bbolt
	now := time.Now().UnixNano()
	_ = db.Batch(func(tx *bbolt.Tx) error {
		sb := tx.Bucket([]byte("syslog"))
		for i := 0; i < 10; i++ {
			l := LogEnt{
				Time: now + int64(i),
				Type: "syslog",
				Log:  fmt.Sprintf(`{"hostname":"host-%d","severity":5,"facility":1,"tag":"test","content":"syslog-%d"}`, i, i),
			}
			b, _ := json.Marshal(l)
			_ = sb.Put([]byte(fmt.Sprintf("%016x", l.Time)), b)
		}

		tb := tx.Bucket([]byte("trap"))
		for i := 0; i < 5; i++ {
			l := LogEnt{
				Time: now + int64(i),
				Type: "trap",
				Log:  fmt.Sprintf(`{"FromAddress":"192.168.1.%d","Variables":"snmpTrapOID.0=1.3.6.1.4.1.1"}`, i),
			}
			b, _ := json.Marshal(l)
			_ = tb.Put([]byte(fmt.Sprintf("%016x", l.Time)), b)
		}

		pb := tx.Bucket([]byte("pollingLogs"))
		psub, _ := pb.CreateBucketIfNotExists([]byte("poll-test-1"))
		for i := 0; i < 8; i++ {
			pe := PollingLogEnt{
				Time:      now + int64(i),
				PollingID: "poll-test-1",
				State:     "normal",
				Result:    map[string]interface{}{"idx": i},
			}
			b, _ := json.Marshal(pe)
			_ = psub.Put([]byte(fmt.Sprintf("%016x", pe.Time)), b)
		}
		return nil
	})

	statsBefore := GetBboltLogStats()
	if statsBefore.TotalCount != 23 {
		t.Fatalf("expected 23 records before migration, got %d", statsBefore.TotalCount)
	}

	// Run migration
	var progressRecords []int
	err = MigrateBboltToParquet(func(bucket string, current, total int) {
		progressRecords = append(progressRecords, current)
	})
	if err != nil {
		t.Fatalf("MigrateBboltToParquet failed: %v", err)
	}

	if MapConf.LogFormat != "parquet" {
		t.Errorf("expected MapConf.LogFormat to be 'parquet', got '%s'", MapConf.LogFormat)
	}

	// Verify records in parquet
	var migratedSyslogs []*SyslogEnt
	_ = ForEachSyslog(now-1000, now+1000, func(s *SyslogEnt) bool {
		migratedSyslogs = append(migratedSyslogs, s)
		return true
	})
	if len(migratedSyslogs) != 10 {
		t.Errorf("expected 10 migrated syslogs, got %d", len(migratedSyslogs))
	}

	migratedPolls := GetAllPollingLog("poll-test-1")
	if len(migratedPolls) != 8 {
		t.Errorf("expected 8 migrated polling logs, got %d", len(migratedPolls))
	}

	// Verify bbolt buckets are emptied
	statsAfter := GetBboltLogStats()
	if statsAfter.TotalCount != 0 {
		t.Errorf("expected 0 records in bbolt after migration, got %d", statsAfter.TotalCount)
	}
	_ = &wg
	_ = ctx
}

func TestBboltLogDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twsnmpfk-detect-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "twsnmpfk.db")
	dspath = tmpDir

	if err := openDB(dbPath); err != nil {
		t.Fatalf("openDB failed: %v", err)
	}

	// For a fresh map, it should default to parquet
	if MapConf.LogFormat != "parquet" {
		t.Fatalf("expected fresh map to be parquet, got '%s'", MapConf.LogFormat)
	}

	// Add a log to bbolt's syslog bucket
	now := time.Now().UnixNano()
	_ = db.Update(func(tx *bbolt.Tx) error {
		sb := tx.Bucket([]byte("syslog"))
		l := LogEnt{
			Time: now,
			Type: "syslog",
			Log:  `{"hostname":"test","severity":5,"facility":1,"tag":"app","content":"hello"}`,
		}
		b, _ := json.Marshal(l)
		return sb.Put([]byte(fmt.Sprintf("%016x", l.Time)), b)
	})

	// Close and reopen DB
	CloseDB()

	if err := openDB(dbPath); err != nil {
		t.Fatalf("openDB reopen failed: %v", err)
	}
	defer CloseDB()

	// Since bbolt has logs, it MUST be judged as bbolt!
	if !HasBboltLogs() {
		t.Errorf("expected HasBboltLogs to be true")
	}
	if MapConf.LogFormat != "bbolt" {
		t.Errorf("expected MapConf.LogFormat to be 'bbolt' when bbolt has logs, got '%s'", MapConf.LogFormat)
	}
}

