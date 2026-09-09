package datastore

import (
	"bytes"
	"compress/flate"
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func getTmpDBFile() (string, error) {
	f, err := os.CreateTemp("", "twsnmpfk_test")
	if err != nil {
		return "", err
	}
	return f.Name(), err
}

func TestDataStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	td, err := os.MkdirTemp("", "twsnmpfk_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	wg := &sync.WaitGroup{}
	Init(ctx, td, wg)
	MapConf.MapName = "Test123"
	if err := SaveMapConf(); err != nil {
		t.Fatal(err)
	}
	defer cancel()
	backdb, err := getTmpDBFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backdb)
	if err := BackupDB(backdb); err != nil {
		t.Fatal(err)
	}
	CloseDB()
	MapConf.MapName = ""
	err = openDB(backdb)
	if err != nil {
		t.Fatal(err)
	}
	if MapConf.MapName != "Test123" {
		t.Errorf("Backup MapName = '%s'", MapConf.MapName)
	}
	CloseDB()
}

func TestCompressDecompress(t *testing.T) {
	orig := []byte(`{"Time":123456789,"Type":"syslog","Log":"test message test message test message test message test message"}`)
	comp := compressLog(orig)
	decomp := deCompressLog(comp)
	if !bytes.Equal(orig, decomp) {
		t.Fatalf("Decompress failed: got %s, want %s", decomp, orig)
	}

	// Test uncompressed
	uncomp := []byte(`{"Time":123456789,"Type":"syslog","Log":"short"}`)
	result := deCompressLog(uncomp)
	if !bytes.Equal(uncomp, result) {
		t.Fatalf("deCompressLog on uncompressed data modified it: got %s, want %s", result, uncomp)
	}

	// Test backward compatibility with old compressLog format (f.Flush() + f.Close())
	var b bytes.Buffer
	f, _ := flate.NewWriter(&b, flate.DefaultCompression)
	f.Write(orig)
	f.Flush()
	f.Close()
	oldComp := b.Bytes()
	decompOld := deCompressLog(oldComp)
	if !bytes.Equal(orig, decompOld) {
		t.Fatalf("Decompress old format failed: got %s, want %s", decompOld, orig)
	}
}

func TestSaveAndRetrieveSyslog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	td, err := os.MkdirTemp("", "twsnmpfk_test_syslog")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	wg := &sync.WaitGroup{}
	Init(ctx, td, wg)
	defer cancel()
	defer CloseDB()

	now := time.Now().UnixNano()
	logs := []*LogEnt{
		{
			Time: now,
			Type: "syslog",
			Log:  `{"severity":6,"facility":1,"hostname":"host1","tag":"test","content":"short"}`,
		},
		{
			Time: now + 1,
			Type: "syslog",
			Log:  `{"severity":3,"facility":1,"hostname":"host2","tag":"test","content":"this is a long log message that will exceed one hundred bytes in length when serialized to json LogEnt structure"}`,
		},
	}
	SaveLogBuffer(logs)

	count := 0
	err = ForEachSyslog(now-1000, now+1000000, func(s *SyslogEnt) bool {
		count++
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("Expected 2 logs, got %d", count)
	}
}



