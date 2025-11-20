package datastore

import (
	"context"
	"os"
	"sync"
	"testing"
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
