// Package notify : 通知処理
package notify

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/twsnmp/twsnmpfk/backend"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

var (
	lastExecLevel int
)

func Start(ctx context.Context, wg *sync.WaitGroup) error {
	lastExecLevel = -1
	wg.Add(1)
	go notifyBackend(ctx, wg)
	return nil
}

func notifyBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start notify")
	lastSendReport := time.Now().Add(time.Hour * time.Duration(-24))
	lastLog := time.Now().Add(time.Hour * time.Duration(-1)).UnixNano()
	lastLog = checkNotify(lastLog)
	timer := time.NewTicker(time.Second * 60)
	i := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("stop notify")
			timer.Stop()
			return
		case <-timer.C:
			i++
			if i >= datastore.NotifyConf.Interval {
				i = 0
				lastLog = checkNotify(lastLog)
			}
			checkExecCmd()
			if datastore.NotifyConf.Report &&
				lastSendReport.Day() != time.Now().Day() &&
				len(backend.MonitorDataes) > 1 {
				lastSendReport = time.Now()
				sendReport()
			}
		}
	}
}

func getLevelNum(l string) int {
	switch l {
	case "high":
		return 0
	case "low":
		return 1
	case "warn":
		return 2
	case "none":
		return -1
	}
	return 3
}

func checkNotify(last int64) int64 {
	list := []*datastore.EventLogEnt{}
	lastLogTime := int64(0)
	skip := 0
	datastore.ForEachLastEventLog(last, func(l *datastore.EventLogEnt) bool {
		if lastLogTime < l.Time {
			lastLogTime = l.Time
		}
		list = append(list, l)
		return true
	})
	log.Printf("check notify last=%v next=%v len=%d skip=%d", time.Unix(0, last), time.Unix(0, lastLogTime), len(list), skip)
	if len(list) > 0 {
		sendNotifyMail(list)
	}
	if lastLogTime > 0 {
		return lastLogTime
	}
	return time.Now().UnixNano()
}

type notifyData struct {
	failureSubject string
	failureBody    string
	repairSubject  string
	repairBody     string
}

// getNotifyData : 通知メールの本文と件名を作成する
func getNotifyData(list []*datastore.EventLogEnt, nl int) notifyData {
	fNodeMap := make(map[string]int)
	rNodeMap := make(map[string]int)
	failure := []*datastore.EventLogEnt{}
	repair := []*datastore.EventLogEnt{}
	ti := time.Now().Add(time.Duration(-datastore.NotifyConf.Interval) * time.Minute).UnixNano()
	for _, l := range list {
		if ti > l.Time {
			continue
		}
		if datastore.NotifyConf.NotifyRepair && l.Level == "repair" {
			// 復帰前の状態を確認する
			np := getLevelNum(l.LastLevel)
			if np > nl {
				continue
			}
			rNodeMap[l.NodeName] = np
			repair = append(repair, l)
			continue
		}
		np := getLevelNum(l.Level)
		if np > nl {
			continue
		}
		fNodeMap[l.NodeName] = np
		failure = append(failure, l)
	}
	f, r := "", ""
	fs, rs := "", ""
	if len(failure) > 0 {
		f = eventLogListToString(false, failure)
		fs = datastore.NotifyConf.Subject + i18n.Trans("(Failure)")
		fs += ":" + getNodes(fNodeMap)
	}
	if len(repair) > 0 {
		r = eventLogListToString(true, repair)
		rs = datastore.NotifyConf.Subject + i18n.Trans("(Repair)")
		rs += ":" + getNodes(rNodeMap)
	}
	return notifyData{
		failureSubject: fs,
		failureBody:    f,
		repairSubject:  rs,
		repairBody:     r,
	}
}

func getNodes(m map[string]int) string {
	nodes := []string{}
	l := 0
	for n := range m {
		nodes = append(nodes, n)
		l += len(n)
		if l > 1000 {
			break
		}
	}
	return strings.Join(nodes, ",")
}

// eventLogListToString : イベントログを通知メールの本文に変換する
func eventLogListToString(repair bool, list []*datastore.EventLogEnt) string {
	title := datastore.NotifyConf.Subject + i18n.Trans("(Failure)")
	if repair {
		title = datastore.NotifyConf.Subject + i18n.Trans("(Repair)")
	}
	f := template.FuncMap{
		"levelName":     levelName,
		"formatLogTime": formatLogTime,
	}
	t, err := template.New("notify").Funcs(f).Parse(datastore.LoadMailTemplate("notify"))
	if err != nil {
		return fmt.Sprintf("make mail err=%v", err)
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, map[string]interface{}{
		"Title": title,
		"Logs":  list,
	}); err != nil {
		return fmt.Sprintf("make mail err=%v", err)
	}
	return buffer.String()
}

func levelName(s string) string {
	switch s {
	case "high":
		return i18n.Trans("High")
	case "low":
		return i18n.Trans("Low")
	case "warn":
		return i18n.Trans("Warnning")
	case "normal", "up":
		return i18n.Trans("Normal")
	case "repair":
		return i18n.Trans("Repair")
	}
	return i18n.Trans("Unknown")
}

func formatLogTime(t int64) string {
	return time.Unix(0, t).Local().Format("2006/01/02 15:04:05")
}
func formatAITime(t int64) string {
	return time.Unix(t, 0).Local().Format("2006/01/02 15:04:05")
}

func scoreClass(s float64) string {
	if s >= 50 {
		return "info"
	} else if s > 42 {
		return "warn"
	} else if s > 33 {
		return "low"
	}
	return "high"
}

func aiScoreClass(s float64) string {
	if s > 100.0 {
		s = 1.0
	} else {
		s = 100.0 - s
	}
	return scoreClass(s)
}

func formatScore(s float64) string {
	return fmt.Sprintf("%.2f", s)
}

func formatCount(i interface{}) string {
	c := int64(0)
	switch v := i.(type) {
	case int64:
		c = v
	case int:
		c = int64(v)
	case float32:
		c = int64(v)
	case float64:
		c = int64(v)
	}
	return humanize.Comma(c)
}
