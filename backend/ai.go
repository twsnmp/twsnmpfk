package backend

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
	"gonum.org/v1/gonum/mat"

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

type aiDataFrame struct {
	Time []int64
	Data map[string][]float64
}

func (df aiDataFrame) Len() int {
	return len(df.Time)
}

func (df aiDataFrame) ColumnNames() []string {
	cols := []string{}
	for k := range df.Data {
		cols = append(cols, k)
	}
	return cols
}

func (df aiDataFrame) ToCSV(w io.Writer) error {
	cols := []string{}
	row := "time"
	for k := range df.Data {
		cols = append(cols, k)
		row += "," + k
	}
	if _, err := w.Write([]byte(row + "\n")); err != nil {
		return err
	}
	for i, t := range df.Time {
		row = time.Unix(t, 0).Format(time.RFC3339)
		for _, col := range cols {
			row += ","
			if v, ok := df.Data[col]; ok && len(v) > i {
				row += fmt.Sprintf("%f", v[i])
			}
		}
		if _, err := w.Write([]byte(row + "\n")); err != nil {
			return err
		}
	}
	return nil
}

type AIReq struct {
	PollingID string
	Df        aiDataFrame
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
		nextAIReqTimeMap.Delete(id)
	}
	return err
}

var nextAIReqTimeMap sync.Map

func checkLastAIResultTime(id string) bool {
	if v, ok := nextAIReqTimeMap.Load(id); ok {
		if lt, ok := v.(int64); ok {
			return lt < time.Now().Unix()-60*60
		}
	}
	last, err := datastore.GetAIReesult(id)
	if err != nil {
		return true
	}
	nextAIReqTimeMap.Store(id, last.LastTime)
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
	if req.Df.Len() < 10 {
		return
	}
	nextAIReqTimeMap.Store(pe.ID, time.Now().Unix())
	st := time.Now()
	calcAIScore(req, pe.AIMode)
	log.Printf("calc ai score id=%s name=%s len=%d dur=%v", pe.ID, pe.Name, req.Df.Len(), time.Since(st))
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
	keys = append(keys, "state")
	req.Df = aiDataFrame{
		Time: []int64{},
		Data: make(map[string][]float64),
	}
	req.Df.Data["hour"] = []float64{}
	req.Df.Data["weekday"] = []float64{}
	for _, k := range keys {
		req.Df.Data[k] = []float64{}
	}
	logs := datastore.GetAllPollingLog(req.PollingID)
	if len(logs) < 1 {
		return fmt.Errorf("no logs")
	}
	st := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	ent := make(map[string]float64)
	maxVals := make(map[string]float64)
	for _, k := range keys {
		ent[k] = 0.0
		maxVals[k] = 0.0
	}
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
			req.Df.Time = append(req.Df.Time, ts.Unix())
			req.Df.Data["hour"] = append(req.Df.Data["hour"], float64(ts.Hour())/23)
			req.Df.Data["weekday"] = append(req.Df.Data["weekday"], float64(ts.Weekday())/6)
			for _, k := range keys {
				avg := ent[k] / count
				req.Df.Data[k] = append(req.Df.Data[k], avg)
				if maxVals[k] < avg {
					maxVals[k] = avg
				}
			}
			for _, k := range keys {
				ent[k] = 0.0
			}
			st = ct
			count = 0.0
		}
		count += 1.0
		for _, k := range keys {
			if k == "state" {
				ent["state"] += getStateNum(l.State)
				continue
			}
			if v, ok := l.Result[k]; ok {
				if fv, ok := v.(float64); ok {
					ent[k] += fv
				}
			}
		}
	}
	for _, k := range keys {
		for j := range req.Df.Data[k] {
			if maxVals[k] > 0.0 {
				req.Df.Data[k][j] /= maxVals[k]
			} else {
				req.Df.Data[k][j] = 0.0
			}
		}
	}
	if p.VectorCols != "" {
		colMap := make(map[string]bool)
		for _, c := range strings.Split(p.VectorCols, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				colMap[c] = true
			}
		}
		for k := range req.Df.Data {
			if _, ok := colMap[k]; !ok {
				delete(req.Df.Data, k)
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

func calcAIScore(req *AIReq, aiMode string) {
	var res *datastore.AIResult
	switch aiMode {
	case "hotelling":
		res = calcHotelling(req)
	case "knn":
		res = calcKNN(req)
	default:
		res = calcIForest(req)
	}
	if res == nil || len(res.ScoreData) < 1 {
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

func getSampleData(req *AIReq) [][]float64 {
	cols := req.Df.ColumnNames()
	data := make([][]float64, req.Df.Len())
	for i := range data {
		data[i] = make([]float64, len(cols))
	}
	for i, col := range cols {
		if v, ok := req.Df.Data[col]; ok {
			for j, d := range v {
				data[j][i] = d
			}
		}
	}
	return data
}

func makeDeviationScore(req *AIReq, r []float64) *datastore.AIResult {
	res := &datastore.AIResult{}
	if len(r) == 0 {
		return res
	}
	max, err := stats.Max(r)
	if err != nil {
		return res
	}
	min, err := stats.Min(r)
	if err != nil {
		return res
	}
	diff := max - min
	if diff == 0 {
		for i := range r {
			res.ScoreData = append(res.ScoreData, []float64{float64(req.Df.Time[i]), 50.0})
		}
		res.PollingID = req.PollingID
		res.LastTime = req.Df.Time[len(req.Df.Time)-1]
		return res
	}
	for i := range r {
		r[i] /= diff
		r[i] *= 100.0
	}
	mean, err := stats.Mean(r)
	if err != nil {
		return res
	}
	sd, err := stats.StandardDeviation(r)
	if err != nil || sd == 0 {
		for i := range r {
			res.ScoreData = append(res.ScoreData, []float64{float64(req.Df.Time[i]), 50.0})
		}
	} else {
		for i := range r {
			score := ((10 * (r[i] - mean) / sd) + 50)
			res.ScoreData = append(res.ScoreData, []float64{float64(req.Df.Time[i]), score})
		}
	}
	res.PollingID = req.PollingID
	res.LastTime = req.Df.Time[len(req.Df.Time)-1]
	return res
}

func calcIForest(req *AIReq) *datastore.AIResult {
	sub := 256
	if req.Df.Len() < sub {
		sub = req.Df.Len() / 2
		log.Printf("IForest subSample=%d", sub)
	}
	data := getSampleData(req)
	iforest, err := go_iforest.NewIForest(data, 1000, sub)
	if err != nil {
		log.Printf("NewIForest err=%v", err)
		return &datastore.AIResult{}
	}
	r := make([]float64, len(data))
	for i, v := range data {
		r[i] = iforest.CalculateAnomalyScore(v)
	}
	return makeDeviationScore(req, r)
}

func calcHotelling(req *AIReq) *datastore.AIResult {
	data := getSampleData(req)
	n := len(data)
	if n < 10 {
		return &datastore.AIResult{}
	}
	d := len(data[0])

	// Calculate mean vector
	mean := make([]float64, d)
	for _, row := range data {
		for j, val := range row {
			mean[j] += val
		}
	}
	for j := range mean {
		mean[j] /= float64(n)
	}

	// Calculate covariance matrix
	covData := make([]float64, d*d)
	for _, row := range data {
		for r := 0; r < d; r++ {
			for c := 0; c < d; c++ {
				covData[r*d+c] += (row[r] - mean[r]) * (row[c] - mean[c])
			}
		}
	}
	for i := range covData {
		covData[i] /= float64(n)
	}

	// Ridge Regularization to ensure invertibility
	epsilon := 1e-6
	for r := 0; r < d; r++ {
		covData[r*d+r] += epsilon
	}

	cov := mat.NewDense(d, d, covData)
	var inv mat.Dense
	if err := inv.Inverse(cov); err != nil {
		log.Printf("Hotelling Inverse err=%v", err)
		return &datastore.AIResult{}
	}

	r := make([]float64, n)
	for i, row := range data {
		diff := make([]float64, d)
		for j := range row {
			diff[j] = row[j] - mean[j]
		}
		diffVec := mat.NewVecDense(d, diff)
		var tmp mat.VecDense
		tmp.MulVec(&inv, diffVec)
		r[i] = mat.Dot(diffVec, &tmp)
	}

	return makeDeviationScore(req, r)
}

func calcKNN(req *AIReq) *datastore.AIResult {
	data := getSampleData(req)
	n := len(data)
	if n < 10 {
		return &datastore.AIResult{}
	}
	d := len(data[0])

	k := 5
	if k >= n {
		k = n - 1
	}
	if k < 1 {
		k = 1
	}

	r := make([]float64, n)
	for i := 0; i < n; i++ {
		dists := make([]float64, 0, n-1)
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			var sum float64
			for col := 0; col < d; col++ {
				diff := data[i][col] - data[j][col]
				sum += diff * diff
			}
			dists = append(dists, math.Sqrt(sum))
		}
		sort.Float64s(dists)

		var sumDist float64
		for idx := 0; idx < k; idx++ {
			sumDist += dists[idx]
		}
		r[i] = sumDist / float64(k)
	}

	return makeDeviationScore(req, r)
}
