package backend

import (
	"log"
	"math"
	"time"

	go_iforest "github.com/codegaudi/go-iforest"
	"github.com/montanaflynn/stats"
	"github.com/twsnmp/golof/lof"
	"github.com/twsnmp/tfidf"
	"github.com/twsnmp/tfidf/seg"
	"github.com/twsnmp/twsnmpfk/datastore"
)

type SyslogAnomalyItem struct {
	Time     int64   `json:"Time"`
	Host     string  `json:"Host"`
	Tag      string  `json:"Tag"`
	Severity int     `json:"Severity"`
	Facility int     `json:"Facility"`
	Message  string  `json:"Message"`
	Score    float64 `json:"Score"`
}

type SyslogAnomalyResponse struct {
	DurationMs int64                `json:"DurationMs"`
	Logs       []*SyslogAnomalyItem `json:"Logs"`
}

// CalculateSyslogAnomaly calculates anomaly score for each syslog entry using specified algorithm and vector mode
func CalculateSyslogAnomaly(logs []*datastore.SyslogEnt, algo, vmode string) *SyslogAnomalyResponse {
	st := time.Now()
	res := &SyslogAnomalyResponse{
		Logs: make([]*SyslogAnomalyItem, 0, len(logs)),
	}
	if len(logs) == 0 {
		return res
	}

	for _, l := range logs {
		res.Logs = append(res.Logs, &SyslogAnomalyItem{
			Time:     l.Time,
			Host:     l.Host,
			Tag:      l.Tag,
			Severity: l.Severity,
			Facility: l.Facility,
			Message:  l.Message,
			Score:    50.0,
		})
	}

	if len(logs) < 2 {
		res.DurationMs = time.Since(st).Milliseconds()
		return res
	}

	vectors := extractSyslogVectors(logs, vmode)
	if len(vectors) != len(logs) || len(vectors[0]) == 0 {
		res.DurationMs = time.Since(st).Milliseconds()
		return res
	}

	rawScores := computeAnomalyScores(vectors, algo)
	if len(rawScores) == len(logs) {
		scaledScores := scaleDeviationScores(rawScores)
		for i, s := range scaledScores {
			res.Logs[i].Score = math.Round(s*100) / 100
		}
	}

	res.DurationMs = time.Since(st).Milliseconds()
	return res
}

func extractSyslogVectors(logs []*datastore.SyslogEnt, vmode string) [][]float64 {
	n := len(logs)
	switch vmode {
	case "tfidf":
		lines := make([]string, n)
		for i, l := range logs {
			lines[i] = l.Message
		}
		f := tfidf.NewTokenizer(seg.NewLogTokenizer(true))
		f.AddDocs(lines...)
		dim := 100
		if len(lines) < dim {
			dim = max(5, len(lines))
		}
		vectors := f.GetTFIDF(dim, lines...)
		// If TF-IDF returns empty or invalid dimension, fallback
		if len(vectors) == n && len(vectors[0]) > 0 {
			return vectors
		}
		return extractNumSyslogVectors(logs, true)

	case "security":
		var allKeys []string
		allKeys = append(allKeys, sqlKeys...)
		allKeys = append(allKeys, oscmdKeys...)
		allKeys = append(allKeys, dirTraversalKeys...)
		vectors := make([][]float64, n)
		for i, l := range logs {
			vectors[i] = getKeywordsVector(l.Message, allKeys)
		}
		return vectors

	case "time":
		vectors := make([][]float64, n)
		for i, l := range logs {
			ts := time.Unix(0, l.Time).Local()
			vectors[i] = []float64{
				float64(ts.Day()),
				float64(ts.Weekday()),
				float64(ts.Hour()),
				float64(ts.Minute()),
				float64(l.Severity),
				float64(l.Facility),
			}
		}
		return vectors

	case "alltime":
		vectors := make([][]float64, n)
		for i, l := range logs {
			ts := time.Unix(0, l.Time).Local()
			vectors[i] = []float64{
				strToFloat(l.Host),
				strToFloat(l.Tag),
				strToFloat(l.Message),
				float64(len(l.Message)),
				float64(l.Severity),
				float64(l.Facility),
				float64(ts.Weekday()),
				float64(ts.Hour()),
				float64(ts.Minute()),
			}
		}
		return vectors

	default: // "num"
		return extractNumSyslogVectors(logs, false)
	}
}

func extractNumSyslogVectors(logs []*datastore.SyslogEnt, includeTime bool) [][]float64 {
	n := len(logs)
	vectors := make([][]float64, n)
	for i, l := range logs {
		v := []float64{
			float64(l.Severity),
			float64(l.Facility),
			float64(len(l.Message)),
			strToFloat(l.Host),
			strToFloat(l.Tag),
		}
		if includeTime {
			ts := time.Unix(0, l.Time).Local()
			v = append(v, float64(ts.Weekday()), float64(ts.Hour()))
		}
		vectors[i] = v
	}
	return vectors
}

func computeAnomalyScores(vectors [][]float64, algo string) []float64 {
	n := len(vectors)
	scores := make([]float64, n)

	switch algo {
	case "zscore", "stat":
		stat := NewStatDetector()
		if err := stat.Fit(vectors); err != nil {
			log.Printf("Stat Fit err=%v", err)
			return scores
		}
		for i, v := range vectors {
			scores[i] = stat.Score(v)
		}

	case "lof":
		samples := lof.GetSamplesFromFloat64s(vectors)
		k := 5
		if k >= n {
			k = max(1, n-1)
		}
		lofGetter := lof.NewLOF(k)
		if err := lofGetter.Train(samples); err != nil {
			log.Printf("LOF Train err=%v", err)
			return scores
		}
		for i, s := range samples {
			scores[i] = lofGetter.GetLOF(s, "fast")
		}

	case "knn":
		knn := NewKNNDetector(5)
		if err := knn.Fit(vectors); err != nil {
			log.Printf("KNN Fit err=%v", err)
			return scores
		}
		scores = knn.ScoreBatch(vectors)

	case "mahalanobis":
		md := NewMahalanobisDetector()
		if err := md.Fit(vectors); err != nil {
			log.Printf("Mahalanobis Fit err=%v", err)
			return scores
		}
		for i, v := range vectors {
			scores[i] = md.Score(v)
		}

	case "autoencoder", "ae":
		ae := NewAutoencoderDetector()
		if err := ae.Fit(vectors); err != nil {
			log.Printf("Autoencoder Fit err=%v", err)
			return scores
		}
		for i, v := range vectors {
			scores[i] = ae.Score(v)
		}

	case "lstm":
		lstm := NewLSTMDetector()
		if err := lstm.Fit(vectors); err != nil {
			log.Printf("LSTM Fit err=%v", err)
			return scores
		}
		for i, v := range vectors {
			scores[i] = lstm.Score(v)
		}

	case "sum":
		for i, v := range vectors {
			var sum float64
			for _, val := range v {
				sum += val
			}
			scores[i] = sum
		}

	default: // "iforest"
		sub := 256
		if n < sub {
			sub = max(10, n/2)
		}
		iforest, err := go_iforest.NewIForest(vectors, 1000, sub)
		if err != nil {
			log.Printf("IForest err=%v", err)
			return scores
		}
		for i, v := range vectors {
			scores[i] = iforest.CalculateAnomalyScore(v)
		}
	}

	return scores
}

func scaleDeviationScores(r []float64) []float64 {
	n := len(r)
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	maxVal, err := stats.Max(r)
	if err != nil {
		for i := range out {
			out[i] = 50.0
		}
		return out
	}
	minVal, err := stats.Min(r)
	if err != nil || maxVal == minVal {
		for i := range out {
			out[i] = 50.0
		}
		return out
	}

	norm := make([]float64, n)
	diff := maxVal - minVal
	for i := range r {
		norm[i] = ((r[i] - minVal) / diff) * 100.0
	}

	mean, err := stats.Mean(norm)
	if err != nil {
		return norm
	}
	sd, err := stats.StandardDeviation(norm)
	if err != nil || sd == 0 {
		for i := range out {
			out[i] = 50.0
		}
		return out
	}

	for i := range norm {
		score := ((10 * (norm[i] - mean) / sd) + 50)
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		out[i] = score
	}
	return out
}
