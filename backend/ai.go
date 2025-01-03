package backend

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/montanaflynn/stats"

	go_iforest "github.com/codegaudi/go-iforest"
	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

func aiBackend(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("start ai")
	timer := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			log.Println("stop ai")
			return
		case <-timer.C:
			checkAI()
		}
	}
}

type AIReq struct {
	PollingID string
	TimeStamp []int64
	Data      [][]float64
}

func checkAI() {
	st := time.Now().Unix()
	datastore.ForEachPollings(func(pe *datastore.PollingEnt) bool {
		if pe.LogMode == datastore.LogModeAI {
			doAI(pe)
		}
		return time.Now().Unix()-st < 50
	})
}

func DeleteAIResult(id string) error {
	err := datastore.DeleteAIResult(id)
	if err == nil {
		delete(nextAIReqTimeMap, id)
	}
	return err
}

var nextAIReqTimeMap = make(map[string]int64)

func checkLastAIResultTime(id string) bool {
	if lt, ok := nextAIReqTimeMap[id]; ok {
		return lt < time.Now().Unix()-60*60
	}
	last, err := datastore.GetAIReesult(id)
	if err != nil {
		return true
	}
	nextAIReqTimeMap[id] = last.LastTime
	return last.LastTime < time.Now().Unix()-60*60
}

func doAI(pe *datastore.PollingEnt) {
	if !checkLastAIResultTime(pe.ID) {
		return
	}
	req := &AIReq{
		PollingID: pe.ID,
	}
	err := MakeAIData(req)
	if err != nil {
		log.Printf("make ai data id=%s name=%s err=%v", pe.ID, pe.Name, err)
		return
	}
	if len(req.Data) < 10 {
		return
	}
	nextAIReqTimeMap[pe.ID] = time.Now().Unix() + 60*60
	st := time.Now()
	calcAIScore(req)
	log.Printf("calc ai score id=%s name=%s len=%d dur=%v", pe.ID, pe.Name, len(req.Data), time.Since(st))
}

func getAIDataKeys(p *datastore.PollingEnt) []string {
	keys := []string{}
	if p.Type == "syslog" && p.Mode == "pri" {
		for i := 0; i < 256; i++ {
			keys = append(keys, fmt.Sprintf("pri_%d", i))
		}
		return keys
	}
	for k, v := range p.Result {
		// lastTimeは、測定データに含めない
		if k == "lastTime" {
			continue
		}
		if _, ok := v.(float64); !ok {
			continue
		}
		keys = append(keys, k)
	}
	return keys
}

func MakeAIData(req *AIReq) error {
	p := datastore.GetPolling(req.PollingID)
	if p == nil {
		return fmt.Errorf("no polling")
	}
	keys := getAIDataKeys(p)
	if len(keys) < 1 {
		return fmt.Errorf("no keys")
	}
	logs := datastore.GetAllPollingLog(req.PollingID)
	if len(logs) < 1 {
		return fmt.Errorf("no logs")
	}
	entLen := len(keys) + 3
	st := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	ent := make([]float64, entLen)
	maxVals := make([]float64, entLen)
	var count float64
	for _, l := range logs {
		ct := 3600 * (time.Unix(0, l.Time).Unix() / 3600)
		if st != ct {
			if count == 0.0 {
				// Dataがない場合はスキップする
				st = ct
				continue
			}
			ts := time.Unix(ct, 0)
			ent[0] = float64(ts.Hour())
			ent[1] = float64(ts.Weekday())
			for i := 0; i < len(ent); i++ {
				if i >= 3 {
					ent[i] /= count
				}
				if maxVals[i] < ent[i] {
					maxVals[i] = ent[i]
				}
			}
			req.TimeStamp = append(req.TimeStamp, ts.Unix())
			req.Data = append(req.Data, ent)
			ent = make([]float64, entLen)
			st = ct
			count = 0.0
		}
		count += 1.0
		ent[3] += getStateNum(l.State)
		for i, k := range keys {
			if v, ok := l.Result[k]; ok {
				if fv, ok := v.(float64); ok {
					ent[i+3] += fv
				}
			}
		}
	}
	for i := range req.Data {
		for j := range req.Data[i] {
			if maxVals[j] > 0.0 {
				req.Data[i][j] /= maxVals[j]
			} else {
				req.Data[i][j] = 0.0
			}
		}
	}
	return nil
}

func getStateNum(s string) float64 {
	if s == "repair" || s == "normal" {
		return 1.0
	}
	if s == "unknown" {
		return 0.5
	}
	return 0.0
}

func calcAIScore(req *AIReq) {
	var res = calcIForest(req)
	if len(res.ScoreData) < 1 {
		return
	}
	if err := datastore.SaveAIResult(res); err != nil {
		log.Printf("save ai result err=%v", err)
		return
	}
	pe := datastore.GetPolling(req.PollingID)
	if pe == nil {
		return
	}
	n := datastore.GetNode(pe.NodeID)
	if n == nil {
		return
	}
	if len(res.ScoreData) > 0 {
		ls := res.ScoreData[len(res.ScoreData)-1][1]
		level := ""
		if datastore.AIConf.HighThreshold > 0 && ls > datastore.AIConf.HighThreshold {
			level = "high"
		} else if datastore.AIConf.LowThreshold > 0 && ls > datastore.AIConf.LowThreshold {
			level = "low"
		} else if datastore.AIConf.WarnThreshold > 0 && ls > datastore.AIConf.WarnThreshold {
			level = "warn"
		}
		if level != "" {
			datastore.AddEventLog(&datastore.EventLogEnt{
				Type:     "ai",
				Level:    level,
				NodeID:   pe.NodeID,
				NodeName: n.Name,
				Event:    fmt.Sprintf(i18n.Trans("AI report:%s(%s):%f"), pe.Name, pe.Type, ls),
			})
		}
	}
}

func calcIForest(req *AIReq) *datastore.AIResult {
	res := datastore.AIResult{}
	sub := 256
	if len(req.Data) < sub {
		sub = len(req.Data) / 2
		log.Printf("IForest subSample=%d", sub)
	}
	iforest, err := go_iforest.NewIForest(req.Data, 1000, sub)
	if err != nil {
		log.Printf("NewIForest err=%v", err)
		return &res
	}
	r := make([]float64, len(req.Data))
	for i, v := range req.Data {
		r[i] = iforest.CalculateAnomalyScore(v)
	}
	max, err := stats.Max(r)
	if err != nil {
		return &res
	}
	min, err := stats.Min(r)
	if err != nil {
		return &res
	}
	diff := max - min
	if diff == 0 {
		return &res
	}
	for i := range r {
		r[i] /= diff
		r[i] *= 100.0
	}
	mean, err := stats.Mean(r)
	if err != nil {
		return &res
	}
	sd, err := stats.StandardDeviation(r)
	if err != nil {
		return &res
	}
	for i := range r {
		score := ((10 * (float64(r[i]) - mean) / sd) + 50)
		res.ScoreData = append(res.ScoreData, []float64{float64(req.TimeStamp[i]), score})
	}
	res.PollingID = req.PollingID
	res.LastTime = req.TimeStamp[len(req.TimeStamp)-1]
	return &res
}
