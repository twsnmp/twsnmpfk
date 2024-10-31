package backend

import (
	"context"
	"log"
	"os"
	"runtime"
	"strings"
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
	Time         int64   `json:"Time"`
	CPU          float64 `json:"CPU"`
	Mem          float64 `json:"Mem"`
	MyCPU        float64 `json:"MyCPU"`
	MyMem        float64 `json:"MyMem"`
	Swap         float64 `json:"Swap"`
	Disk         float64 `json:"Disk"`
	Load         float64 `json:"Load"`
	Bytes        float64 `json:"Bytes"`
	Net          float64 `json:"Net"`
	Conn         int     `json:"Conn"`
	Proc         int     `json:"Proc"`
	DBSize       int64   `json:"DBSize"`
	HeapAlloc    int64   `json:"HeapAlloc"`
	Sys          int64   `json:"Sys"`
	NumGoroutine int     `json:"NumGoroutine"`
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
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	m.HeapAlloc = int64(ms.HeapAlloc)
	m.Sys = int64(ms.Sys)
	v, err := mem.VirtualMemory()
	if err == nil {
		m.Mem = v.UsedPercent
	}
	s, err := mem.SwapMemory()
	if err == nil {
		m.Swap = s.UsedPercent
	}
	m.Time = time.Now().UnixNano()
	d, err := disk.Usage(dspath)
	if err == nil {
		m.Disk = d.UsedPercent
	}
	n, err := gopsnet.IOCounters(true)
	if err == nil {
		for _, nif := range n {
			if isMonitorIF(nif.Name) {
				m.Bytes += float64(nif.BytesRecv)
				m.Bytes += float64(nif.BytesSent)
			}
		}
		if len(MonitorDataes) >= 1 {
			o := MonitorDataes[len(MonitorDataes)-1]
			if m.Bytes >= o.Bytes && m.Time > o.Time {
				m.Net = float64(1000*1000*1000*8.0*(m.Bytes-o.Bytes)) / float64(m.Time-o.Time)
			} else {
				log.Println("skip net monior")
			}
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
	pid := os.Getpid()
	pr, err := process.NewProcess(int32(pid))
	if err == nil {
		if v, err := pr.CPUPercent(); err == nil {
			m.MyCPU = v
		}
		if v, err := pr.MemoryPercent(); err == nil {
			m.MyMem = float64(v)
		}
	}
	m.NumGoroutine = runtime.NumGoroutine()

	m.DBSize = datastore.GetDBSize()
	for len(MonitorDataes) > maxMonitorData {
		MonitorDataes = append(MonitorDataes[:0], MonitorDataes[1:]...)
	}
	MonitorDataes = append(MonitorDataes, m)
}

func isMonitorIF(n string) bool {
	if runtime.GOOS == "darwin" {
		if strings.HasPrefix(n, "utun") {
			return false
		}
		if strings.HasPrefix(n, "lo") {
			return false
		}
	}
	return true
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
