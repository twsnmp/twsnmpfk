package logreport

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

// NodeSummary holds node information needed by reports.
type NodeSummary struct {
	Name string
	ID   string
	IP   string
	MAC  string
}

// ReportStore abstracts the underlying persistence and system lookup helpers.
type ReportStore interface {
	GetLogReportData(ctx context.Context, kind, id string) ([]byte, error)
	ListLogReportData(ctx context.Context, kind string) (map[string][]byte, error)
	SaveLogReportData(ctx context.Context, kind string, items map[string][]byte) error
	DeleteLogReportData(ctx context.Context, kind string, ids []string) error
	ResetLogReportData(ctx context.Context, kind string) error
	FindVendor(mac string) string
	GetLoc(ip string) string
	LookupNode(ip string) NodeSummary
	ListNodes(ctx context.Context) ([]NodeSummary, error)
	UpdateSensor(host, sensorType, param string, count int64)
	CheckSensorMonitor(host, sensorType, param string, cpu, mem, load, txSpeed, rxSpeed float64, sent, recv, proc int64)
	GetRetentionConfig(ctx context.Context) (days int, limit int)
}

// DefaultDatastoreAdapter implements ReportStore using twsnmpfk/datastore package functions.
type DefaultDatastoreAdapter struct{}

// NewDefaultStore creates a ReportStore backed by twsnmpfk/datastore.
func NewDefaultStore() ReportStore {
	return &DefaultDatastoreAdapter{}
}

func (d *DefaultDatastoreAdapter) GetLogReportData(_ context.Context, kind, id string) ([]byte, error) {
	return datastore.GetLogReportData(kind, id)
}

func (d *DefaultDatastoreAdapter) ListLogReportData(_ context.Context, kind string) (map[string][]byte, error) {
	return datastore.ListLogReportData(kind)
}

func (d *DefaultDatastoreAdapter) SaveLogReportData(_ context.Context, kind string, items map[string][]byte) error {
	return datastore.SaveLogReportData(kind, items)
}

func (d *DefaultDatastoreAdapter) DeleteLogReportData(_ context.Context, kind string, ids []string) error {
	return datastore.DeleteLogReportData(kind, ids)
}

func (d *DefaultDatastoreAdapter) ResetLogReportData(_ context.Context, kind string) error {
	return datastore.ResetLogReportData(kind)
}

func (d *DefaultDatastoreAdapter) FindVendor(mac string) string {
	return datastore.FindVendor(mac)
}

func (d *DefaultDatastoreAdapter) GetLoc(ip string) string {
	return datastore.GetLoc(ip)
}

type ipToNameCacheEnt struct {
	Name      string
	NodeID    string
	TimeLimit int64
}

var ipToNameCache sync.Map

func (d *DefaultDatastoreAdapter) LookupNode(ip string) NodeSummary {
	var nSum NodeSummary
	nSum.Name = ip
	nSum.IP = ip
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if n.IP == ip {
			nSum.Name = n.Name
			nSum.ID = n.ID
			return false
		}
		return true
	})
	if nSum.ID != "" {
		return nSum
	}
	now := time.Now().Unix()
	if v, ok := ipToNameCache.Load(ip); ok {
		if c, ok := v.(*ipToNameCacheEnt); ok && c.TimeLimit > now {
			nSum.Name = c.Name
			return nSum
		}
	}
	if net.ParseIP(ip) != nil {
		r := &net.Resolver{}
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		if names, err := r.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
			name := strings.TrimSuffix(names[0], ".")
			if name != "" {
				nSum.Name = name
			}
		}
	}
	ipToNameCache.Store(ip, &ipToNameCacheEnt{
		Name:      nSum.Name,
		TimeLimit: now + 24*3600,
	})
	return nSum
}

func (d *DefaultDatastoreAdapter) ListNodes(_ context.Context) ([]NodeSummary, error) {
	var list []NodeSummary
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		list = append(list, NodeSummary{Name: n.Name, ID: n.ID, IP: n.IP, MAC: n.MAC})
		return true
	})
	return list, nil
}

func (d *DefaultDatastoreAdapter) UpdateSensor(host, sensorType, param string, count int64) {
	datastore.UpdateSensor(host, sensorType, param, count)
}

func (d *DefaultDatastoreAdapter) CheckSensorMonitor(host, sensorType, param string, cpu, mem, load, txSpeed, rxSpeed float64, sent, recv, proc int64) {
	datastore.CheckSensorMonitor(host, sensorType, param, cpu, mem, load, txSpeed, rxSpeed, sent, recv, proc)
}

func (d *DefaultDatastoreAdapter) GetRetentionConfig(_ context.Context) (int, int) {
	days := datastore.MapConf.LogDays
	if days <= 0 {
		days = 30
	}
	return days, 10000
}
