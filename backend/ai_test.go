package backend

import (
	"math"
	"testing"

	_ "github.com/twsnmp/twsnmpfk/datastore"
)

func TestMakeDeviationScore(t *testing.T) {
	req := &AIReq{
		PollingID: "test-poll",
		Df: aiDataFrame{
			Time: []int64{100, 200, 300},
			Data: map[string][]float64{
				"v": {10, 20, 30},
			},
		},
	}
	r := []float64{1.0, 2.0, 3.0}
	res := makeDeviationScore(req, r)
	if len(res.ScoreData) != 3 {
		t.Fatalf("expected 3 items, got %d", len(res.ScoreData))
	}
	// Verify that the scores are around deviation values (mean is 50, standard deviation 10)
	// r = {1, 2, 3} -> Normalized: min=1, max=3, diff=2. normalized r = {0, 50, 100}
	// mean of normalized r = 50. stddev = 40.8248
	// deviation score of 0 = 10*(0-50)/40.8248 + 50 = 37.75
	// deviation score of 50 = 50.0
	// deviation score of 100 = 62.25
	if math.Abs(res.ScoreData[0][1]-37.75) > 0.1 {
		t.Errorf("expected ~37.75, got %f", res.ScoreData[0][1])
	}
	if math.Abs(res.ScoreData[1][1]-50.0) > 0.1 {
		t.Errorf("expected ~50.0, got %f", res.ScoreData[1][1])
	}
	if math.Abs(res.ScoreData[2][1]-62.25) > 0.1 {
		t.Errorf("expected ~62.25, got %f", res.ScoreData[2][1])
	}

	// Constant value check
	rConst := []float64{2.0, 2.0, 2.0}
	resConst := makeDeviationScore(req, rConst)
	for _, score := range resConst.ScoreData {
		if score[1] != 50.0 {
			t.Errorf("expected constant score to be 50.0, got %f", score[1])
		}
	}
}

func TestCalcHotelling(t *testing.T) {
	// Normal 2D dataset
	times := make([]int64, 20)
	v1 := make([]float64, 20)
	v2 := make([]float64, 20)
	for i := 0; i < 20; i++ {
		times[i] = int64(1000 + i*10)
		v1[i] = float64(i)
		v2[i] = float64(i * 2)
	}
	// Introduce one outlier at index 10
	v1[10] = 100.0

	req := &AIReq{
		PollingID: "test-hotelling",
		Df: aiDataFrame{
			Time: times,
			Data: map[string][]float64{
				"v1": v1,
				"v2": v2,
			},
		},
	}

	res := calcHotelling(req)
	if len(res.ScoreData) != 20 {
		t.Fatalf("expected 20 items, got %d", len(res.ScoreData))
	}
	// Outlier at 10 should have the highest anomaly score (thus highest deviation score)
	maxScore := 0.0
	maxIdx := -1
	for i, sd := range res.ScoreData {
		if sd[1] > maxScore {
			maxScore = sd[1]
			maxIdx = i
		}
	}
	if maxIdx != 10 {
		t.Errorf("expected outlier at index 10 to have max score, but got index %d with score %f", maxIdx, maxScore)
	}

	// Test zero variance edge case (should not panic due to Ridge Regularization)
	vZero := make([]float64, 20)
	reqZero := &AIReq{
		PollingID: "test-hotelling-zero",
		Df: aiDataFrame{
			Time: times,
			Data: map[string][]float64{
				"v1": vZero,
				"v2": vZero,
			},
		},
	}
	// Should return successfully without panic (all scores 50.0)
	resZero := calcHotelling(reqZero)
	if len(resZero.ScoreData) != 20 {
		t.Errorf("expected 20 items for zero-variance, got %d", len(resZero.ScoreData))
	}
}

func TestCalcKNN(t *testing.T) {
	times := make([]int64, 15)
	v1 := make([]float64, 15)
	for i := 0; i < 15; i++ {
		times[i] = int64(1000 + i*10)
		v1[i] = 10.0 // mostly constant
	}
	v1[7] = 50.0 // outlier

	req := &AIReq{
		PollingID: "test-knn",
		Df: aiDataFrame{
			Time: times,
			Data: map[string][]float64{
				"v1": v1,
			},
		},
	}

	res := calcKNN(req)
	if len(res.ScoreData) != 15 {
		t.Fatalf("expected 15 items, got %d", len(res.ScoreData))
	}
	maxScore := 0.0
	maxIdx := -1
	for i, sd := range res.ScoreData {
		if sd[1] > maxScore {
			maxScore = sd[1]
			maxIdx = i
		}
	}
	if maxIdx != 7 {
		t.Errorf("expected outlier at index 7 to have max score, but got index %d with score %f", maxIdx, maxScore)
	}
}
