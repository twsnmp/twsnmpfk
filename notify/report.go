// Package notify : 通知処理
package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"runtime"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func getEventLogSummary() string {
	high := 0
	low := 0
	warn := 0
	normal := 0
	other := 0
	st := time.Now().Add(time.Duration(-24) * time.Hour).UnixNano()
	datastore.ForEachLastEventLog(func(l *datastore.EventLogEnt) bool {
		if l.Time < st {
			return false
		}
		switch l.Level {
		case "high":
			high++
		case "low":
			low++
		case "warn":
			warn++
			return true
		case "normal", "repair":
			normal++
			return true
		default:
			other++
			return true
		}
		return true
	})
	return fmt.Sprintf(i18n.Trans("High=%d,Low=%d,Warn=%d,Normal=%d,Other=%d"), high, low, warn, normal, other)
}

func getMapInfo() []reportInfoEnt {
	high := 0
	low := 0
	warn := 0
	normal := 0
	repair := 0
	unknown := 0
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		switch n.State {
		case "high":
			high++
		case "low":
			low++
		case "warn":
			warn++
		case "normal":
			normal++
		case "repair":
			repair++
		default:
			unknown++
		}
		return true
	})
	state := i18n.Trans("Unknown")
	class := "none"
	if high > 0 {
		state = i18n.Trans("High")
		class = "high"
	} else if low > 0 {
		state = i18n.Trans("Low")
		class = "low"
	} else if warn > 0 {
		class = "warn"
		state = i18n.Trans("Warnning")
	} else if normal+repair > 0 {
		class = "normal"
		state = i18n.Trans("Normal")
	}
	return []reportInfoEnt{
		{
			Name:  i18n.Trans("MAP Name"),
			Value: datastore.MapConf.MapName,
			Class: "none",
		}, {
			Name:  i18n.Trans("MAP State"),
			Value: state,
			Class: class,
		}, {
			Name:  i18n.Trans("Node count by state"),
			Value: fmt.Sprintf(i18n.Trans("High=%d,Low=%d,Warn=%d,Normal=%d,Other=%d"), high, low, warn, repair, normal, unknown),
			Class: "none",
		},
	}
}

func getResInfo() []reportInfoEnt {
	if len(backend.MonitorDataes) < 1 {
		return []reportInfoEnt{}
	}
	cpu := []float64{}
	mem := []float64{}
	myCpu := []float64{}
	myMem := []float64{}
	swap := []float64{}
	disk := []float64{}
	load := []float64{}
	gr := []float64{}
	for _, m := range backend.MonitorDataes {
		cpu = append(cpu, m.CPU)
		mem = append(mem, m.Mem)
		myCpu = append(myCpu, m.MyCPU)
		myMem = append(myMem, m.MyMem)
		swap = append(swap, m.Swap)
		disk = append(disk, m.Disk)
		load = append(load, m.Load)
		gr = append(gr, float64(m.NumGoroutine))
	}
	cpuMin, _ := stats.Min(cpu)
	cpuMean, _ := stats.Mean(cpu)
	cpuMax, _ := stats.Max(cpu)
	memMin, _ := stats.Min(mem)
	memMean, _ := stats.Mean(mem)
	memMax, _ := stats.Max(mem)
	myCpuMin, _ := stats.Min(myCpu)
	myCpuMean, _ := stats.Mean(myCpu)
	myCpuMax, _ := stats.Max(myCpu)
	myMemMin, _ := stats.Min(myMem)
	myMemMean, _ := stats.Mean(myMem)
	myMemMax, _ := stats.Max(myMem)
	swapMin, _ := stats.Min(swap)
	swapMean, _ := stats.Mean(swap)
	swapMax, _ := stats.Max(swap)
	diskMin, _ := stats.Min(disk)
	diskMean, _ := stats.Mean(disk)
	diskMax, _ := stats.Max(disk)
	loadMin, _ := stats.Min(load)
	loadMean, _ := stats.Mean(load)
	loadMax, _ := stats.Max(load)
	grMin, _ := stats.Min(gr)
	grMean, _ := stats.Mean(gr)
	grMax, _ := stats.Max(gr)
	myMemClass := "none"
	diskClass := "none"
	loadClass := "none"
	if myMemMean > 90.0 && memMean > 90.0 {
		myMemClass = "high"
	} else if myMemMean > 80.0 && memMean > 80.0 {
		myMemClass = "low"
	} else if myMemMean > 60.0 && memMean > 60.0 {
		myMemClass = "warn"
	} else {
		myMemClass = "none"
	}
	if diskMean > 95.0 {
		diskClass = "high"
	} else if diskMean > 90.0 {
		diskClass = "low"
	} else if diskMean > 80.0 {
		diskClass = "warn"
	} else {
		diskClass = "none"
	}
	if loadMean > float64(runtime.NumCPU()) {
		loadClass = "high"
	} else {
		loadClass = "none"
	}

	return []reportInfoEnt{
		{
			Name: i18n.Trans("CPU Usage"),
			Value: fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", cpuMin),
				humanize.FormatFloat("###.##", cpuMean),
				humanize.FormatFloat("###.##", cpuMax),
			),
			Class: "none",
		}, {
			Name: i18n.Trans("Memory Usage"),
			Value: fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", memMin),
				humanize.FormatFloat("###.##", memMean),
				humanize.FormatFloat("###.##", memMax),
			),
			Class: "none",
		}, {
			Name: "My " + i18n.Trans("CPU Usage"),
			Value: fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", myCpuMin),
				humanize.FormatFloat("###.##", myCpuMean),
				humanize.FormatFloat("###.##", myCpuMax),
			),
			Class: "none",
		}, {
			Name: "My " + i18n.Trans("Memory Usage"),
			Value: fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", myMemMin),
				humanize.FormatFloat("###.##", myMemMean),
				humanize.FormatFloat("###.##", myMemMax),
			),
			Class: myMemClass,
		}, {
			Name: "Swap Usage",
			Value: fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", swapMin),
				humanize.FormatFloat("###.##", swapMean),
				humanize.FormatFloat("###.##", swapMax),
			),
			Class: "none",
		}, {
			Name: i18n.Trans("Disk Usage"),
			Value: fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", diskMin),
				humanize.FormatFloat("###.##", diskMean),
				humanize.FormatFloat("###.##", diskMax),
			),
			Class: diskClass,
		}, {
			Name: i18n.Trans("System Load"),
			Value: fmt.Sprintf(i18n.Trans("Min:%s Avg:%s Max:%s"),
				humanize.FormatFloat("###.##", loadMin),
				humanize.FormatFloat("###.##", loadMean),
				humanize.FormatFloat("###.##", loadMax),
			),
			Class: loadClass,
		}, {
			Name: "Go Routine",
			Value: fmt.Sprintf(i18n.Trans("Min:%s Avg:%s Max:%s"),
				humanize.FormatFloat("###.#", grMin),
				humanize.FormatFloat("###.#", grMean),
				humanize.FormatFloat("###.#", grMax),
			),
			Class: "none",
		},
	}
}

func getAIInfo() []string {
	ret := []string{"Score,Node,Polling,Count"}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode != datastore.LogModeAI {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) < 1 {
			return true
		}
		ret = append(ret,
			fmt.Sprintf("%.2f,%s,%s,%d",
				air.ScoreData[len(air.ScoreData)-1][1],
				n.Name,
				p.Name,
				len(air.ScoreData),
			))
		return true
	})
	return ret
}

type aiResultEnt struct {
	LastScore   float64
	NodeName    string
	PollingName string
	Count       int
	LastTime    int64
}

func getAIList() []aiResultEnt {
	ret := []aiResultEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.LogMode != datastore.LogModeAI {
			return true
		}
		n := datastore.GetNode(p.NodeID)
		if n == nil {
			return true
		}
		air, err := datastore.GetAIReesult(p.ID)
		if err != nil || len(air.ScoreData) < 1 {
			return true
		}
		ret = append(ret, aiResultEnt{
			LastScore:   air.ScoreData[len(air.ScoreData)-1][1],
			NodeName:    n.Name,
			PollingName: p.Name,
			Count:       len(air.ScoreData),
			LastTime:    air.LastTime,
		})
		return true
	})
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LastScore > ret[j].LastScore
	})
	return ret
}

type reportInfoEnt struct {
	Name  string
	Class string
	Value string
}

func sendReport() {
	info := []reportInfoEnt{}
	info = append(info, getMapInfo()...)
	info = append(info, getResInfo()...)
	info = append(info, reportInfoEnt{
		Name:  i18n.Trans("DB Size"),
		Value: humanize.Bytes(uint64(datastore.GetDBSize())),
		Class: "none",
	})
	logSum := getEventLogSummary()
	info = append(info, reportInfoEnt{
		Name:  i18n.Trans("Log count by level"),
		Value: logSum,
		Class: "none",
	})
	title := fmt.Sprintf(i18n.Trans("%s(report) at %s"), datastore.NotifyConf.Subject, time.Now().Format("2006/01/02 15:04:05"))
	f := template.FuncMap{
		"levelName":     levelName,
		"formatLogTime": formatLogTime,
		"formatScore":   formatScore,
		"scoreClass":    scoreClass,
		"aiScoreClass":  aiScoreClass,
		"formatAITime":  formatAITime,
		"formatCount":   formatCount,
	}
	t, err := template.New("report").Funcs(f).Parse(datastore.LoadMailTemplate("report"))
	if err != nil {
		log.Printf("send report mail err=%v", err)
		return
	}
	body := new(bytes.Buffer)
	if err = t.Execute(body, map[string]interface{}{
		"Title":  title,
		"Info":   info,
		"AIList": getAIList(),
	}); err != nil {
		log.Printf("send report mail err=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "low",
			Event: fmt.Sprintf(i18n.Trans("Failed to send report mail err=%v"), err),
		})
		return
	}
	if err := SendMail(title, body.String()); err != nil {
		log.Printf("send report mail err=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "low",
			Event: fmt.Sprintf(i18n.Trans("Failed to send report mail err=%v"), err),
		})
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: i18n.Trans("Send report mail"),
		})
	}
}
