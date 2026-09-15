package backend

import (
	"math"
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestDetectors(t *testing.T) {
	// Normal cluster around (1.0, 1.0)
	normalVectors := [][]float64{
		{1.0, 1.0},
		{1.1, 0.9},
		{0.9, 1.1},
		{1.05, 0.95},
		{0.95, 1.05},
		{1.0, 1.02},
		{0.98, 1.01},
		{1.02, 0.99},
		{1.01, 1.0},
		{0.99, 1.0},
	}
	anomalyVector := []float64{10.0, 10.0}

	t.Run("KNNDetector", func(t *testing.T) {
		d := NewKNNDetector(3)
		if err := d.Fit(normalVectors); err != nil {
			t.Fatalf("KNN Fit err: %v", err)
		}
		normScore := d.Score(normalVectors[0])
		anomScore := d.Score(anomalyVector)
		if anomScore <= normScore {
			t.Errorf("KNN anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
		}

		batchScores := d.ScoreBatch(normalVectors)
		if len(batchScores) != len(normalVectors) {
			t.Errorf("KNN ScoreBatch length expected %d, got %d", len(normalVectors), len(batchScores))
		}
	})

	t.Run("MahalanobisDetector", func(t *testing.T) {
		d := NewMahalanobisDetector()
		if err := d.Fit(normalVectors); err != nil {
			t.Fatalf("Mahalanobis Fit err: %v", err)
		}
		normScore := d.Score(normalVectors[0])
		anomScore := d.Score(anomalyVector)
		if anomScore <= normScore {
			t.Errorf("Mahalanobis anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
		}
	})

	t.Run("StatDetector", func(t *testing.T) {
		d := NewStatDetector()
		if err := d.Fit(normalVectors); err != nil {
			t.Fatalf("Stat Fit err: %v", err)
		}
		normScore := d.Score(normalVectors[0])
		anomScore := d.Score(anomalyVector)
		if anomScore <= normScore {
			t.Errorf("Stat anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
		}
	})

	t.Run("AutoencoderDetector", func(t *testing.T) {
		d := NewAutoencoderDetector()
		d.epochs = 10
		if err := d.Fit(normalVectors); err != nil {
			t.Fatalf("Autoencoder Fit err: %v", err)
		}
		normScore := d.Score(normalVectors[0])
		anomScore := d.Score(anomalyVector)
		if math.IsNaN(anomScore) || math.IsNaN(normScore) {
			t.Errorf("Autoencoder score is NaN: norm=%f, anom=%f", normScore, anomScore)
		}
		if anomScore <= normScore {
			t.Errorf("Autoencoder anomaly score (%f) should be higher than normal score (%f)", anomScore, normScore)
		}
	})

	t.Run("LSTMDetector", func(t *testing.T) {
		d := NewLSTMDetector()
		if err := d.Fit(normalVectors); err != nil {
			t.Fatalf("LSTM Fit err: %v", err)
		}
		_ = d.Score(normalVectors[0])
		nextScore := d.Score(normalVectors[1])
		anomScore := d.Score(anomalyVector)
		if math.IsNaN(anomScore) || math.IsNaN(nextScore) {
			t.Errorf("LSTM score is NaN: next=%f, anom=%f", nextScore, anomScore)
		}
		if anomScore <= nextScore {
			t.Errorf("LSTM anomaly score (%f) should be higher than normal transition score (%f)", anomScore, nextScore)
		}
	})
}

func TestCalculateSyslogAnomaly(t *testing.T) {
	now := time.Now().UnixNano()
	logs := []*datastore.SyslogEnt{
		{Time: now, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user1"},
		{Time: now + 1000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user2"},
		{Time: now + 2000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user3"},
		{Time: now + 3000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user4"},
		{Time: now + 4000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user5"},
		{Time: now + 5000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user6"},
		{Time: now + 6000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user7"},
		{Time: now + 7000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user8"},
		{Time: now + 8000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user9"},
		{Time: now + 9000, Host: "192.168.1.1", Tag: "sshd", Severity: 6, Facility: 1, Message: "Accepted publickey for user10"},
		{Time: now + 10000, Host: "192.168.1.99", Tag: "kernel", Severity: 1, Facility: 0, Message: "ALERT: rm -rf /bin ; cat /etc/passwd and select * from user -- ' OR '1'='1"},
	}

	algos := []string{"iforest", "zscore", "lof", "knn", "mahalanobis"}
	vmodes := []string{"tfidf", "security", "alltime", "time", "num"}

	for _, algo := range algos {
		for _, vmode := range vmodes {
			res := CalculateSyslogAnomaly(logs, algo, vmode)
			if res == nil {
				t.Fatalf("CalculateSyslogAnomaly returned nil for algo=%s, vmode=%s", algo, vmode)
			}
			if len(res.Logs) != len(logs) {
				t.Fatalf("expected %d logs, got %d for algo=%s, vmode=%s", len(logs), len(res.Logs), algo, vmode)
			}
		}
	}
}
