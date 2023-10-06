package backend

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/twsnmp/twsnmpfk/datastore"
)

const (
	maxMonitorData = 12 * 24 * 7
)

// MonitorDataEnt :
type MonitorDataEnt struct {
	Time   int64   `json:"Time"`
	CPU    float64 `json:"CPU"`
	Mem    float64 `json:"Mem"`
	Disk   float64 `json:"Disk"`
	Load   float64 `json:"Load"`
	Bytes  float64 `json:"Byte"`
	Net    float64 `json:"Net"`
	Proc   int     `json:"Proc"`
	Conn   int     `json:"Conn"`
	DBSize int64   `json:"DBSize"`
}

// MonitorDataes : モニターデータ
var MonitorDataes []*MonitorDataEnt

func updateMonData() {
	m := &MonitorDataEnt{}
	cpus, err := cpu.Percent(0, false)
	if err == nil {
		m.CPU = cpus[0]
	}
	l, err := load.Avg()
	if err == nil {
		m.Load = l.Load1
	}
	v, err := mem.VirtualMemory()
	if err == nil {
		m.Mem = v.UsedPercent
	}
	m.Time = time.Now().UnixNano()
	d, err := disk.Usage(dspath)
	if err == nil {
		m.Disk = d.UsedPercent
	}
	n, err := gopsnet.IOCounters(false)
	if err == nil {
		m.Bytes = float64(n[0].BytesRecv)
		m.Bytes += float64(n[0].BytesSent)
		if len(MonitorDataes) > 1 {
			o := MonitorDataes[len(MonitorDataes)-1]
			m.Net = float64(1000*1000*1000*8.0*(m.Bytes-o.Bytes)) / float64(m.Time-o.Time)
		}
	}
	conn, err := gopsnet.Connections("tcp")
	if err == nil {
		m.Conn = len(conn)
	}
	pids, err := process.Pids()
	if err == nil {
		m.Proc = len(pids)
	}
	m.DBSize = datastore.GetDBSize()
	for len(MonitorDataes) > maxMonitorData {
		MonitorDataes = append(MonitorDataes[:0], MonitorDataes[1:]...)
	}
	MonitorDataes = append(MonitorDataes, m)
}

// monitor :
func monitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start monitor")
	timer := time.NewTicker(time.Second * 300)
	updateMonData()
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop monitor")
			return
		case <-timer.C:
			updateMonData()
		}
	}
}
