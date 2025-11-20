package discover

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/ping"
)

func TestDiscover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ping.Start(ctx, &sync.WaitGroup{})
	defer cancel()
	time.Sleep(time.Second * 1)
	td, err := os.MkdirTemp("", "twsnmpfk_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(td)
	datastore.Init(ctx, td, &sync.WaitGroup{})
	datastore.MapConf.MapName = "Test123"
	if err := datastore.SaveMapConf(); err != nil {
		t.Fatal(err)
	}
	datastore.MapConf.Community = "public"
	datastore.DiscoverConf.StartIP = "192.168.1.1"
	datastore.DiscoverConf.EndIP = "192.168.1.2"
	datastore.DiscoverConf.Retry = 1
	datastore.DiscoverConf.Timeout = 2
	err = StartDiscover()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 15)
	t.Log("Done")
}
