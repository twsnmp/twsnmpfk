package main

import (
	"log"
	"strings"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
	wails "github.com/wailsapp/wails/v2/pkg/runtime"
)

type OTelMetricEnt struct {
	Host    string `json:"Host"`
	Service string `json:"Service"`
	Scope   string `json:"Scope"`
	Name    string `json:"Name"`
	Type    string `json:"Type"`
	Count   int    `json:"Count"`
	First   int64  `json:"First"`
	Last    int64  `json:"Last"`
}

// GetOTelMetrics retunrs  OpenTelmentry metric summary list
func (a *App) GetOTelMetrics() []*OTelMetricEnt {
	ret := []*OTelMetricEnt{}
	datastore.ForEachOTelMetric(func(m *datastore.OTelMetricEnt) bool {
		ret = append(ret, &OTelMetricEnt{
			Host:    m.Host,
			Service: m.Service,
			Scope:   m.Scope,
			Name:    m.Name,
			Type:    m.Type,
			Count:   m.Count,
			First:   m.First,
			Last:    m.Last,
		})
		return true
	})
	return ret
}

func (a *App) GetOTelMetric(m OTelMetricEnt) *datastore.OTelMetricEnt {
	return datastore.FindOTelMetric(m.Host, m.Service, m.Scope, m.Name)
}

func (a *App) DeleteOTelMetric(m OTelMetricEnt) {
	pm := datastore.FindOTelMetric(m.Host, m.Service, m.Scope, m.Name)
	if pm != nil {
		datastore.DeleteOTelMetric(pm)
	}
}

type OTelTraceEnt struct {
	Bucket   string  `json:"Bucket"`
	TraceID  string  `json:"TraceID"`
	Hosts    string  `json:"Hosts"`
	Services string  `json:"Services"`
	Scopes   string  `json:"Scopes"`
	Start    int64   `json:"Start"`
	End      int64   `json:"End"`
	Dur      float64 `json:"Dur"`
	NumSpan  int     `json:"NumSpan"`
}

func (a *App) GetOTelTraceBucketList() []string {
	return datastore.GetOTelTraceBucketList()
}

// GetOTelTraces retunrs  OpenTelmentry trace summary list
func (a *App) GetOTelTraces(bks []string) []*OTelTraceEnt {
	ret := []*OTelTraceEnt{}
	for _, b := range bks {
		datastore.ForEachOTelTrace(b, func(t *datastore.OTelTraceEnt) bool {
			hosts := []string{}
			services := []string{}
			scopes := []string{}
			hostMap := make(map[string]bool)
			serviceMap := make(map[string]bool)
			scopeMap := make(map[string]bool)
			for _, span := range t.Spans {
				if _, ok := hostMap[span.Host]; !ok {
					hostMap[span.Host] = true
					hosts = append(hosts, span.Host)
				}
				if _, ok := serviceMap[span.Service]; !ok {
					serviceMap[span.Service] = true
					services = append(services, span.Service)
				}
				if _, ok := scopeMap[span.Scope]; !ok {
					scopeMap[span.Scope] = true
					scopes = append(scopes, span.Scope)
				}
			}
			ret = append(ret, &OTelTraceEnt{
				Bucket:   b,
				TraceID:  t.TraceID,
				Hosts:    strings.Join(hosts, " "),
				Services: strings.Join(services, " "),
				Scopes:   strings.Join(scopes, " "),
				Start:    t.Start,
				End:      t.End,
				Dur:      t.Dur,
				NumSpan:  len(t.Spans),
			})
			return len(ret) < 100000
		})
	}
	return ret
}

// GetOTelTrace retunrs OpenTelemtry Trace
func (a *App) GetOTelTrace(bucket, traceid string) *datastore.OTelTraceEnt {
	return datastore.GetOTelTrace(bucket, traceid)
}

// DeleteAllOTelDataは、OpenTelemetryのデータを全て削除します。
func (a *App) DeleteAllOTelData() bool {
	result, err := wails.MessageDialog(a.ctx, wails.MessageDialogOptions{
		Type:          wails.QuestionDialog,
		Title:         i18n.Trans("Confirm delete"),
		Message:       i18n.Trans("Do you want to delete?"),
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil || result == "No" {
		return false
	}
	if err := datastore.DeleteAllOTelData(); err != nil {
		log.Println(err)
		return false
	}
	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "user",
		Level: "info",
		Event: i18n.Trans("Delete all OpenTelemetry data"),
	})
	return true
}

// GetLastOTelLogs retunrs last otel syslogs
func (a *App) GetLastOTelLogs() []*datastore.SyslogEnt {
	ret := []*datastore.SyslogEnt{}
	datastore.ForEachLastSyslog(func(l *datastore.SyslogEnt) bool {
		if l.Tag != "otel" {
			return true
		}
		ret = append(ret, l)
		return len(ret) < maxDispLog
	})
	return ret
}
