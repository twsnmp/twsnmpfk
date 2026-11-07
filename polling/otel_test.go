package polling

import (
	"testing"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestParseOTelParams(t *testing.T) {
	// Key-value pairs
	params1 := "service=frontend, name=http_requests_total, host=web01, scope=http"
	pMap1 := parseOTelParams(params1)
	if pMap1["service"] != "frontend" || pMap1["name"] != "http_requests_total" || pMap1["host"] != "web01" || pMap1["scope"] != "http" {
		t.Errorf("parseOTelParams failed for KV: %+v", pMap1)
	}

	// Positional format
	params2 := "server1,auth-service,grpc,user_login"
	pMap2 := parseOTelParams(params2)
	if pMap2["host"] != "server1" || pMap2["service"] != "auth-service" || pMap2["scope"] != "grpc" || pMap2["name"] != "user_login" {
		t.Errorf("parseOTelParams failed for positional: %+v", pMap2)
	}
}

func TestIsOTelSpanError(t *testing.T) {
	span1 := datastore.OTelTraceSpanEnt{
		Attributes: []string{"http.status_code=500", "http.method=GET"},
	}
	if !isOTelSpanError(span1) {
		t.Errorf("expected span1 to be error (500)")
	}

	span2 := datastore.OTelTraceSpanEnt{
		Attributes: []string{"http.status_code=200", "http.method=GET"},
	}
	if isOTelSpanError(span2) {
		t.Errorf("expected span2 to be normal (200)")
	}

	span3 := datastore.OTelTraceSpanEnt{
		Attributes: []string{"status_code=STATUS_CODE_ERROR"},
	}
	if !isOTelSpanError(span3) {
		t.Errorf("expected span3 to be error (STATUS_CODE_ERROR)")
	}

	span4 := datastore.OTelTraceSpanEnt{
		Attributes: []string{"error=true"},
	}
	if !isOTelSpanError(span4) {
		t.Errorf("expected span4 to be error (error=true)")
	}

	span5 := datastore.OTelTraceSpanEnt{
		Attributes: []string{"rpc.grpc.status_code=2"},
	}
	if !isOTelSpanError(span5) {
		t.Errorf("expected span5 to be error (grpc error code 2)")
	}
}

func TestDoPollingOTelMetric(t *testing.T) {
	// Add mock metric to datastore
	m := &datastore.OTelMetricEnt{
		Host:    "web-01",
		Service: "web-app",
		Scope:   "metrics",
		Name:    "jvm_memory_used",
		Type:    "Gauge",
		First:   time.Now().UnixNano(),
		Last:    time.Now().UnixNano(),
		Count:   10,
		DataPoints: []*datastore.OTelMetricDataPointEnt{
			{
				Time:  time.Now().UnixNano(),
				Gauge: 85.5,
				Count: 1,
			},
		},
	}
	datastore.AddOTelMetric(m)

	pe := &datastore.PollingEnt{
		ID:      "test-otel-metric",
		Type:    "otel",
		Mode:    "metric",
		Params:  "service=web-app,name=jvm_memory_used",
		Script:  "gauge < 80",
		Level:   "warn",
		PollInt: 60,
		Result:  make(map[string]any),
	}

	doPollingOTel(pe)
	if pe.State != "warn" {
		t.Errorf("expected state warn for gauge=85.5 with script 'gauge < 80', got %s", pe.State)
	}

	// Now with script that passes
	pe.Script = "gauge < 90"
	doPollingOTel(pe)
	// In TWSNMP FK, recovering from warn transitions to 'repair' (acknowledged state)
	if pe.State != "repair" {
		t.Errorf("expected state repair for gauge=85.5 with script 'gauge < 90', got %s", pe.State)
	}
}

func TestDoPollingOTelHeartbeat(t *testing.T) {
	// Add fresh metric to datastore
	m := &datastore.OTelMetricEnt{
		Host:    "web-01",
		Service: "order-service",
		Scope:   "metrics",
		Name:    "active_orders",
		Last:    time.Now().UnixNano(),
		Count:   5,
	}
	datastore.AddOTelMetric(m)

	pe := &datastore.PollingEnt{
		ID:      "test-otel-hb",
		Type:    "otel",
		Mode:    "heartbeat",
		State:   "unknown",
		Params:  "service=order-service",
		Filter:  "10", // 10 seconds threshold
		Level:   "high",
		PollInt: 60,
		Result:  make(map[string]any),
	}

	// Should be normal because Last is fresh
	doPollingOTel(pe)
	if pe.State != "normal" {
		t.Errorf("expected state normal for fresh heartbeat, got %s", pe.State)
	}

	// Now simulate stale metric (older than 10 seconds)
	m.Last = time.Now().Add(-20 * time.Second).UnixNano()
	doPollingOTel(pe)
	if pe.State != "high" {
		t.Errorf("expected state high for stale heartbeat, got %s", pe.State)
	}
}
