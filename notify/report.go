// Package notify : 通知処理
package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func sendReport() {
	if datastore.NotifyConf.HTMLMail {
		sendReportHTML()
	} else {
		sendReportPlain()
	}
}

func sendReportPlain() {
	body := []string{}
	body = append(body, "【現在のマップ情報】")
	body = append(body, getMapInfo(false)...)
	body = append(body, "")
	body = append(body, "")
	body = append(body, "【システムリソース情報】(Min/Mean/Max)")
	body = append(body, getResInfo(false)...)
	body = append(body, "")
	logSum, _, _ := getLastEventLog()
	body = append(body, "【最新24時間のログ集計】")
	body = append(body, logSum...)
	body = append(body, "")
	body = append(body, "【AI分析情報】")
	body = append(body, getAIInfo()...)
	body = append(body, "")

	subject := fmt.Sprintf("%s(定期レポート) at %s", datastore.NotifyConf.Subject, time.Now().Format(time.RFC3339))
	if err := sendMail(subject, strings.Join(body, "\r\n")); err != nil {
		log.Printf("send report mail err=%v", err)
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
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
	datastore.ForEachLastEventLog(0, func(l *datastore.EventLogEnt) bool {
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
		fmt.Sprintf("重度=%d,軽度=%d,注意=%d,正常=%d,その他=%d", high, low, warn, normal, other))
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
	state := "不明"
	class := "none"
	if high > 0 {
		state = "重度"
		class = "high"
	} else if low > 0 {
		state = "軽度"
		class = "low"
	} else if warn > 0 {
		class = "warn"
		state = "注意"
	} else if normal+repair > 0 {
		class = "normal"
		state = "正常"
	}
	if htmlMode {
		return []string{
			datastore.MapConf.MapName,
			state,
			fmt.Sprintf("重度=%d,軽度=%d,注意=%d,復帰=%d,正常=%d,不明=%d", high, low, warn, repair, normal, unknown),
			class,
		}
	}
	return []string{
		fmt.Sprintf("マップ名=%s", datastore.MapConf.MapName),
		fmt.Sprintf("マップ状態=%s", state),
		fmt.Sprintf("重度=%d,軽度=%d,注意=%d,復帰=%d,正常=%d,不明=%d", high, low, warn, repair, normal, unknown),
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
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", cpuMin),
				humanize.FormatFloat("###.##", cpuMean),
				humanize.FormatFloat("###.##", cpuMax),
			),
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", memMin),
				humanize.FormatFloat("###.##", memMean),
				humanize.FormatFloat("###.##", memMax),
			),
			fmt.Sprintf("最小:%s%% 平均:%s%% 最大:%s%%",
				humanize.FormatFloat("###.##", diskMin),
				humanize.FormatFloat("###.##", diskMean),
				humanize.FormatFloat("###.##", diskMax),
			),
			fmt.Sprintf("最小:%s 平均:%s 最大:%s",
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
			Name:  "マップ名",
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "マップの状態",
			Value: a[1],
			Class: a[3],
		})
		info = append(info, reportInfoEnt{
			Name:  "状態別のノード数",
			Value: a[2],
			Class: "none",
		})
	}
	a = getResInfo(true)
	if len(a) > 3 {
		info = append(info, reportInfoEnt{
			Name:  "CPU使用率",
			Value: a[0],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "メモリ使用率",
			Value: a[1],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "ディスク使用率",
			Value: a[2],
			Class: "none",
		})
		info = append(info, reportInfoEnt{
			Name:  "システム負荷",
			Value: a[3],
			Class: "none",
		})
	}
	logSum, _, logs := getLastEventLog()
	if len(logSum) > 0 {
		info = append(info, reportInfoEnt{
			Name:  "状態別のログ数",
			Value: logSum[0],
			Class: "none",
		})
	}
	title := fmt.Sprintf("%s(定期レポート) at %s", datastore.NotifyConf.Subject, time.Now().Format("2006/01/02 15:04:05"))
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
		"URL":    datastore.NotifyConf.URL,
		"Info":   info,
		"Logs":   logs,
		"AIList": getAIList(),
	}); err != nil {
		log.Printf("send report mail err=%v", err)
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "low",
			Event: fmt.Sprintf("定期レポートメール送信失敗 err=%v", err),
		})
		return
	}
	if err := sendMail(title, body.String()); err != nil {
		log.Printf("send report mail err=%v", err)
	} else {
		datastore.AddEventLog(&datastore.EventLogEnt{
			Type:  "system",
			Level: "info",
			Event: "定期レポートメール送信",
		})
	}
}
