package sigma

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CorrelationConfig defines correlation conditions for a rule
type CorrelationConfig struct {
	Frequency int           // Required occurrence count to trigger
	Timeframe time.Duration // Sliding time window
	GroupBy   []string      // Keys to group by (e.g. "client", "user")
}

// ParseCorrelationConfig parses correlation configuration from Rule.AdditionalFields
func ParseCorrelationConfig(raw interface{}) *CorrelationConfig {
	if raw == nil {
		return nil
	}

	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}

	conf := &CorrelationConfig{
		Frequency: 1,
		Timeframe: 60 * time.Second,
		GroupBy:   []string{"client"},
	}

	// 1. Frequency
	if freqRaw, exists := m["frequency"]; exists {
		switch v := freqRaw.(type) {
		case int:
			conf.Frequency = v
		case int64:
			conf.Frequency = int(v)
		case float64:
			conf.Frequency = int(v)
		case string:
			if iv, err := strconv.Atoi(v); err == nil {
				conf.Frequency = iv
			}
		}
	}

	// 2. Timeframe
	if tfRaw, exists := m["timeframe"]; exists {
		switch v := tfRaw.(type) {
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				conf.Timeframe = d
			} else if sec, err := strconv.Atoi(strings.TrimSuffix(v, "s")); err == nil {
				conf.Timeframe = time.Duration(sec) * time.Second
			}
		case int:
			conf.Timeframe = time.Duration(v) * time.Second
		case float64:
			conf.Timeframe = time.Duration(int(v)) * time.Second
		}
	}

	// 3. GroupBy
	if gbRaw, exists := m["group_by"]; exists {
		var gbList []string
		switch v := gbRaw.(type) {
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					gbList = append(gbList, s)
				}
			}
		case []string:
			gbList = v
		case string:
			if v != "" {
				gbList = []string{v}
			}
		}
		if len(gbList) > 0 {
			conf.GroupBy = gbList
		}
	}

	if conf.Frequency <= 1 && conf.Timeframe <= 0 {
		return nil
	}
	return conf
}

// CorrelationTracker manages in-memory sliding windows for correlation rules
type CorrelationTracker struct {
	sync.Mutex
	windows       map[string][]int64 // key -> slice of event timestamps in unix nano
	lastAlert     map[string]int64   // key -> last alert timestamp in unix nano
	lastCleanTime int64
}

// NewCorrelationTracker creates a new CorrelationTracker
func NewCorrelationTracker() *CorrelationTracker {
	return &CorrelationTracker{
		windows:   make(map[string][]int64),
		lastAlert: make(map[string]int64),
	}
}

// Reset clears all tracked windows
func (ct *CorrelationTracker) Reset() {
	ct.Lock()
	defer ct.Unlock()
	ct.windows = make(map[string][]int64)
	ct.lastAlert = make(map[string]int64)
	ct.lastCleanTime = 0
}

// RecordAndCheck records an event occurrence and checks if correlation threshold is reached
func (ct *CorrelationTracker) RecordAndCheck(ruleID string, groupKey string, eventTimeNano int64, conf *CorrelationConfig) bool {
	if conf == nil || conf.Frequency <= 1 {
		return true // No correlation required; pass through
	}

	if eventTimeNano <= 0 {
		eventTimeNano = time.Now().UnixNano()
	}

	ct.Lock()
	defer ct.Unlock()

	// Periodic cleanup every 5 minutes relative to event stream
	if ct.lastCleanTime == 0 || eventTimeNano-ct.lastCleanTime > int64(5*time.Minute) {
		ct.cleanupLocked(eventTimeNano)
		ct.lastCleanTime = eventTimeNano
	}

	key := fmt.Sprintf("%s\t%s", ruleID, groupKey)
	cutoff := eventTimeNano - conf.Timeframe.Nanoseconds()

	// 1. Evict expired events from window
	events := ct.windows[key]
	validEvents := events[:0]
	for _, ts := range events {
		if ts >= cutoff {
			validEvents = append(validEvents, ts)
		}
	}

	// 2. Append current event
	validEvents = append(validEvents, eventTimeNano)
	ct.windows[key] = validEvents

	// 3. Check if threshold reached
	if len(validEvents) >= conf.Frequency {
		// Threshold met! Reset window to prevent duplicate consecutive alert floods
		ct.windows[key] = nil
		ct.lastAlert[key] = eventTimeNano
		return true
	}

	return false
}

// cleanupLocked removes inactive entries from tracker to prevent memory leak
func (ct *CorrelationTracker) cleanupLocked(nowNano int64) {
	expireNano := int64(10 * time.Minute)
	for k, events := range ct.windows {
		if len(events) == 0 {
			delete(ct.windows, k)
			continue
		}
		if nowNano-events[len(events)-1] > expireNano {
			delete(ct.windows, k)
		}
	}
	for k, ts := range ct.lastAlert {
		if nowNano-ts > expireNano {
			delete(ct.lastAlert, k)
		}
	}
}
