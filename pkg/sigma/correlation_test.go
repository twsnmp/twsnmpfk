package sigma

import (
	"testing"
	"time"
)

func TestParseCorrelationConfig(t *testing.T) {
	// Nil
	if conf := ParseCorrelationConfig(nil); conf != nil {
		t.Errorf("expected nil for nil input, got %+v", conf)
	}

	// Simple valid map
	m := map[string]interface{}{
		"frequency": 5,
		"timeframe": "120s",
		"group_by":  []interface{}{"client", "user"},
	}
	conf := ParseCorrelationConfig(m)
	if conf == nil {
		t.Fatalf("expected non-nil correlation config")
	}
	if conf.Frequency != 5 {
		t.Errorf("expected frequency 5, got %d", conf.Frequency)
	}
	if conf.Timeframe != 120*time.Second {
		t.Errorf("expected timeframe 120s, got %v", conf.Timeframe)
	}
	if len(conf.GroupBy) != 2 || conf.GroupBy[0] != "client" || conf.GroupBy[1] != "user" {
		t.Errorf("expected group_by [client, user], got %v", conf.GroupBy)
	}
}

func TestCorrelationTracker_RecordAndCheck(t *testing.T) {
	tracker := NewCorrelationTracker()
	conf := &CorrelationConfig{
		Frequency: 3,
		Timeframe: 60 * time.Second,
		GroupBy:   []string{"client"},
	}

	baseTime := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC).UnixNano()

	// 1st occurrence -> no trigger
	if tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime, conf) {
		t.Errorf("expected false on 1st event")
	}

	// 2nd occurrence (after 10s) -> no trigger
	if tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime+int64(10*time.Second), conf) {
		t.Errorf("expected false on 2nd event")
	}

	// 3rd occurrence from DIFFERENT IP -> no trigger
	if tracker.RecordAndCheck("rule1", "192.168.1.2", baseTime+int64(15*time.Second), conf) {
		t.Errorf("expected false for different groupKey")
	}

	// 3rd occurrence from SAME IP (after 20s) -> TRIGGER!
	if !tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime+int64(20*time.Second), conf) {
		t.Errorf("expected true on 3rd event within timeframe")
	}

	// Immediate next occurrence after trigger should NOT trigger again (window was reset)
	if tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime+int64(21*time.Second), conf) {
		t.Errorf("expected false after trigger reset")
	}
}

func TestCorrelationTracker_WindowExpiration(t *testing.T) {
	tracker := NewCorrelationTracker()
	conf := &CorrelationConfig{
		Frequency: 2,
		Timeframe: 30 * time.Second,
		GroupBy:   []string{"client"},
	}

	baseTime := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC).UnixNano()

	// 1st event at T=0
	if tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime, conf) {
		t.Errorf("expected false on 1st event")
	}

	// 2nd event at T=35s (> 30s timeframe) -> 1st event expired -> no trigger
	if tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime+int64(35*time.Second), conf) {
		t.Errorf("expected false when 1st event expired")
	}

	// 3rd event at T=40s (within 30s of 2nd event at T=35s) -> TRIGGER!
	if !tracker.RecordAndCheck("rule1", "192.168.1.1", baseTime+int64(40*time.Second), conf) {
		t.Errorf("expected true when 2nd and 3rd event are within timeframe")
	}
}
