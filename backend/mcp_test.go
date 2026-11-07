package backend

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/twsnmp/twsnmpfk/datastore"
)

func TestGetOTelMetricsMCP(t *testing.T) {
	datastore.AddOTelMetric(&datastore.OTelMetricEnt{
		Host:    "app-server-1",
		Service: "payment-api",
		Scope:   "http",
		Name:    "requests_total",
		Type:    "Counter",
		Count:   42,
		Last:    time.Now().UnixNano(),
	})

	res, _, err := getOTelMetrics(context.Background(), nil, getOTelMetricsParams{
		ServiceFilter: "payment-api",
	})
	if err != nil {
		t.Fatalf("getOTelMetrics failed: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatalf("expected content, got empty")
	}

	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok || tc.Text == "" {
		t.Fatalf("expected TextContent, got %#v", res.Content[0])
	}

	var list []mcpOTelMetricEnt
	if err := json.Unmarshal([]byte(tc.Text), &list); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(list) != 1 || list[0].Service != "payment-api" {
		t.Errorf("unexpected metrics list: %+v", list)
	}
}

func TestGetSensorListMCP(t *testing.T) {
	res, _, err := getSensorList(context.Background(), nil, getSensorListParams{})
	if err != nil {
		t.Fatalf("getSensorList failed: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatalf("expected content in response")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok || tc.Text == "" {
		t.Fatalf("expected TextContent, got %#v", res.Content[0])
	}
}

func TestGetReportSummaryMCP(t *testing.T) {
	// Empty report type should return validation message
	res, _, err := getReportSummary(context.Background(), nil, getReportSummaryParams{})
	if err != nil {
		t.Fatalf("getReportSummary failed: %v", err)
	}
	if len(res.Content) == 0 {
		t.Fatalf("expected content")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok || tc.Text == "" {
		t.Fatalf("expected TextContent, got %#v", res.Content[0])
	}
}
