// Package notify : 通知処理
package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func sendReport() {
	sendReportHTML()
}

func getLastEventLog() ([]string, []string, []*datastore.EventLogEnt) {
	sum := []string{}
	slogs := []string{}
	logs := []*datastore.EventLogEnt{}
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
		ts := time.Unix(0, l.Time).Local().Format(time.RFC3339Nano)
		slogs = append(slogs, fmt.Sprintf("%s,%s,%s,%s,%s", l.Level, ts, l.Type, l.NodeName, l.Event))
		logs = append(logs, l)
		return true
	})
	sum = append(sum,
		fmt.Sprintf(i18n.Trans("High=%d,Low=%d,Warn=%d,Normal=%d,Other=%d"), high, low, warn, normal, other))
	return sum, slogs, logs
}

func getMapInfo(htmlMode bool) []string {
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
	if htmlMode {
		return []string{
			datastore.MapConf.MapName,
			state,
			fmt.Sprintf(i18n.Trans("High=%d,Low=%d,Warn=%d,Normal=%d,Other=%d"), high, low, warn, repair, normal, unknown),
			class,
		}
	}
	return []string{
		fmt.Sprintf(i18n.Trans("MAP=%s"), datastore.MapConf.MapName),
		fmt.Sprintf(i18n.Trans("MAP State=%s"), state),
		fmt.Sprintf(i18n.Trans("High=%d,Low=%d,Warn=%d,Normal=%d,Other=%d"), high, low, warn, repair, normal, unknown),
	}
}

func getResInfo(htmlMode bool) []string {
	if len(backend.MonitorDataes) < 1 {
		return []string{}
	}
	cpu := []float64{}
	mem := []float64{}
	disk := []float64{}
	load := []float64{}
	for _, m := range backend.MonitorDataes {
		cpu = append(cpu, m.CPU)
		mem = append(mem, m.Mem)
		disk = append(disk, m.Disk)
		load = append(load, m.Load)
	}
	cpuMin, _ := stats.Min(cpu)
	cpuMean, _ := stats.Mean(cpu)
	cpuMax, _ := stats.Max(cpu)
	memMin, _ := stats.Min(mem)
	memMean, _ := stats.Mean(mem)
	memMax, _ := stats.Max(mem)
	diskMin, _ := stats.Min(disk)
	diskMean, _ := stats.Mean(disk)
	diskMax, _ := stats.Max(disk)
	loadMin, _ := stats.Min(load)
	loadMean, _ := stats.Mean(load)
	loadMax, _ := stats.Max(load)
	if htmlMode {
		return []string{
			fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", cpuMin),
				humanize.FormatFloat("###.##", cpuMean),
				humanize.FormatFloat("###.##", cpuMax),
			),
			fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", memMin),
				humanize.FormatFloat("###.##", memMean),
				humanize.FormatFloat("###.##", memMax),
			),
			fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", diskMin),
				humanize.FormatFloat("###.##", diskMean),
				humanize.FormatFloat("###.##", diskMax),
			),
			fmt.Sprintf(i18n.Trans("Min:%s%% Avg:%s%% Max:%s%%"),
				humanize.FormatFloat("###.##", loadMin),
				humanize.FormatFloat("###.##", loadMean),
				humanize.FormatFloat("###.##", loadMax),
			),
		}
	}
	return []string{
		fmt.Sprintf("CPU=%s/%s/%s %%",
			humanize.FormatFloat("###.##", cpuMin),
			humanize.FormatFloat("###.##", cpuMean),
			humanize.FormatFloat("###.##", cpuMax),
		),
		fmt.Sprintf("Mem=%s/%s/%s %%",
			humanize.FormatFloat("###.##", memMin),
			humanize.FormatFloat("###.##", memMean),
			humanize.FormatFloat("###.##", memMax),
		),
		fmt.Sprintf("Disk=%s/%s/%s %%",
			humanize.FormatFloat("###.##", diskMin),
			humanize.FormatFloat("###.##", diskMean),
			humanize.FormatFloat("###.##", diskMax),
		),
		fmt.Sprintf("Load=%s/%s/%s",
			humanize.FormatFloat("###.##", loadMin),
			humanize.FormatFloat("###.##", loadMean),
			humanize.FormatFloat("###.##", loadMax),
		),
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
	return ret
}

type reportInfoEnt struct {
	Name  string
	Class string
	Value string
}

// HTML版レポートの送信
func sendReportHTML() {
	info := []reportInfoEnt{}
	a := getMapInfo(true)
	if len(a) > 3 {
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("MAP Name"),
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("MAP State"),
			Value: a[1],
			Class: a[3],
		})
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("Node count by state"),
			Value: a[2],
			Class: "none",
		})
	}
	a = getResInfo(true)
	if len(a) > 3 {
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("CPU Usage"),
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("Memory Usage"),
			Value: a[1],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("Disk Usage"),
			Value: a[2],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("System Load"),
			Value: a[3],
			Class: "none",
		})
	}
	logSum, _, logs := getLastEventLog()
	if len(logSum) > 0 {
		info = append(info, reportInfoEnt{
			Name:  i18n.Trans("Log count by level"),
			Value: logSum[0],
			Class: "none",
		})
	}
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
		"Logs":   logs,
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
	if err := sendMail(title, body.String()); err != nil {
		log.Printf("send report mail err=%v", err)
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: i18n.Trans("Send report mail"),
		})
	}
}
