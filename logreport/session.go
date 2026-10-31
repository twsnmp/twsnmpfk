package logreport

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MaxSeriesSize is the maximum number of time-series samples kept per entity.
const MaxSeriesSize = 12 * 24 * 7

// Record is the part of a stored syslog message used by the reports.
type Record struct {
	Time     int64
	Host     string
	Tag      string
	Content  string
	Severity int
	Facility int
	Remote   string
}

// ParseRecord converts a stored syslog JSON document into a Record.
// src and t are used when the document does not carry a host / time.
func ParseRecord(logJSON, src string, t int64) (Record, bool) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(logJSON), &m); err != nil {
		return Record{}, false
	}
	str := func(keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k].(string); ok && v != "" {
				return v
			}
		}
		return ""
	}
	rem := str("remote", "client")
	if h, _, err := net.SplitHostPort(rem); err == nil {
		rem = h
	}
	r := Record{
		Time:     t,
		Host:     str("hostname", "host"),
		Tag:      str("tag", "app_name"),
		Content:  str("content", "message"),
		Severity: 6,
		Facility: 1,
		Remote:   rem,
	}
	if sd := str("structured_data"); sd != "" && sd != "-" && !strings.Contains(r.Content, sd) {
		if r.Content != "" {
			r.Content = sd + " " + r.Content
		} else {
			r.Content = sd
		}
	}
	if r.Host == "" {
		r.Host = src
	}
	if v, ok := m["severity"].(float64); ok {
		r.Severity = int(v)
	} else if v, ok := m["severity"].(int); ok {
		r.Severity = v
	}
	if v, ok := m["facility"].(float64); ok {
		r.Facility = int(v)
	} else if v, ok := m["facility"].(int); ok {
		r.Facility = v
	}
	return r, (r.Tag != "" || r.Host != "") && r.Content != ""
}

type nodeInfo struct{ Name, ID string }

// Session accumulates report updates for one polling run or ingestion batch.
// Entities are loaded on demand, modified in memory and written by Commit.
type Session struct {
	ctx    context.Context
	store  ReportStore
	cache  map[string]map[string]any
	dirty  map[string]map[string]struct{}
	loaded map[string]bool
	nodes  map[string]nodeInfo
	// Processed is the number of records that updated a report.
	Processed int
}

// NewSession creates a Session writing to store.
func NewSession(ctx context.Context, store ReportStore) *Session {
	return &Session{
		ctx:    ctx,
		store:  store,
		cache:  map[string]map[string]any{},
		dirty:  map[string]map[string]struct{}{},
		loaded: map[string]bool{},
	}
}

func (s *Session) kindCache(kind string) map[string]any {
	c := s.cache[kind]
	if c == nil {
		c = map[string]any{}
		s.cache[kind] = c
	}
	return c
}

// getEnt returns the entity (cached or loaded from the store) or nil.
func getEnt[T any](s *Session, kind, id string) *T {
	c := s.kindCache(kind)
	if v, ok := c[id]; ok {
		e, _ := v.(*T)
		return e
	}
	b, err := s.store.GetLogReportData(s.ctx, kind, id)
	if err != nil || b == nil {
		return nil
	}
	e := new(T)
	if json.Unmarshal(b, e) != nil {
		return nil
	}
	c[id] = e
	return e
}

// putEnt stores e in the session and marks it for writing.
func putEnt[T any](s *Session, kind, id string, e *T) {
	s.kindCache(kind)[id] = e
	d := s.dirty[kind]
	if d == nil {
		d = map[string]struct{}{}
		s.dirty[kind] = d
	}
	d[id] = struct{}{}
}

// allEnts loads the whole kind into the session and returns every entity.
func allEnts[T any](s *Session, kind string) map[string]*T {
	c := s.kindCache(kind)
	if !s.loaded[kind] {
		if m, err := s.store.ListLogReportData(s.ctx, kind); err == nil {
			for id, b := range m {
				if _, ok := c[id]; ok {
					continue
				}
				e := new(T)
				if json.Unmarshal(b, e) == nil {
					c[id] = e
				}
			}
		}
		s.loaded[kind] = true
	}
	ret := make(map[string]*T, len(c))
	for id, v := range c {
		if e, ok := v.(*T); ok {
			ret[id] = e
		}
	}
	return ret
}

func (s *Session) lookupNode(ip string) nodeInfo {
	if s.nodes == nil {
		s.nodes = map[string]nodeInfo{}
		if nodes, err := s.store.ListNodes(s.ctx); err == nil {
			for _, n := range nodes {
				if n.Name != "" {
					s.nodes[n.Name] = nodeInfo{Name: n.Name, ID: n.ID}
				}
			}
		}
	}
	res := s.store.LookupNode(ip)
	return nodeInfo{Name: res.Name, ID: res.ID}
}

// Commit writes modified entities and refreshes the deviation scores of the
// kinds that have one.
func (s *Session) Commit() error {
	var firstErr error
	for kind, ids := range s.dirty {
		items := make(map[string][]byte, len(ids))
		for id := range ids {
			b, err := json.Marshal(s.cache[kind][id])
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			items[id] = b
		}
		if err := s.store.SaveLogReportData(s.ctx, kind, items); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for kind := range s.dirty {
		var err error
		switch kind {
		case KindWinLogon:
			err = recalcScores[WinLogonEnt](s, kind)
		case KindWinKerberos:
			err = recalcScores[WinKerberosEnt](s, kind)
		case KindRADIUSFlow:
			err = recalcScores[RADIUSFlowEnt](s, kind)
		case KindTLSFlow:
			err = recalcScores[TLSFlowEnt](s, kind)
		case KindFlow:
			err = recalcScores[FlowEnt](s, kind)
		case KindServer:
			err = recalcScores[ServerEnt](s, kind)
		case KindTwAgentTCP:
			err = recalcScores[TwAgentTCPEnt](s, kind)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	s.dirty = map[string]map[string]struct{}{}
	return firstErr
}

var scoreAlertSuppressed sync.Map

// recalcScores computes the deviation score (mean 50, sd 10) from penalties.
func recalcScores[T any, P interface {
	*T
	scoreInfo() *ScoreInfo
}](s *Session, kind string) error {
	m, err := s.store.ListLogReportData(s.ctx, kind)
	if err != nil {
		return err
	}
	ents := make(map[string]P, len(m))
	xs := make([]float64, 0, len(m))
	for id, b := range m {
		e := new(T)
		if json.Unmarshal(b, e) != nil {
			continue
		}
		p := P(e)
		si := p.scoreInfo()
		if si.Penalty > 100 {
			si.Penalty = 100
		}
		xs = append(xs, float64(100-si.Penalty))
		ents[id] = p
	}
	if len(xs) == 0 {
		return nil
	}
	mean, sd := meanSD(xs)
	out := make(map[string][]byte, len(ents))

	scoreThreshold := 35.0

	now := time.Now().UnixNano()
	for id, p := range ents {
		si := p.scoreInfo()
		if sd != 0 {
			si.Score = 10*(float64(100-si.Penalty)-mean)/sd + 50
		} else {
			si.Score = 50
		}
		si.ValidScore = true
		b, err := json.Marshal(p)
		if err != nil {
			continue
		}
		out[id] = b

		// Anomaly score alert suppression
		if si.Score < scoreThreshold {
			suppressKey := fmt.Sprintf("%s:%s", kind, id)
			lastAlert, exists := scoreAlertSuppressed.Load(suppressKey)
			if !exists || now-lastAlert.(int64) > int64(1*time.Hour) {
				scoreAlertSuppressed.Store(suppressKey, now)
			}
		}
	}
	return s.store.SaveLogReportData(s.ctx, kind, out)
}

func meanSD(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	var v float64
	for _, x := range xs {
		v += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(v / float64(len(xs)))
}

// Process applies one syslog record of the given source (polling type).
// It returns true when the record updated a report.
func (s *Session) Process(source string, r Record) bool {
	cleanTag := CleanTag(r.Tag)
	if !strings.EqualFold(cleanTag, SyslogTag(source)) && TagToSource(cleanTag) != source {
		if !(source == SourceTwAgent && strings.Contains(r.Content, "[twsnmpneo@17862")) {
			if !strings.Contains(strings.ToLower(r.Content), strings.ToLower(SyslogTag(source))) {
				return false
			}
		}
	}
	if source == SourceTwAgent {
		return s.processTwAgentSyslog(r)
	}
	m := parseKV(r.Content)
	t := m["type"]
	if t == "" {
		return false
	}
	var ok bool
	switch source {
	case SourceWifiScan:
		ok = s.processWifi(r, t, m)
	case SourceBlueScan:
		ok = s.processBlue(r, t, m)
	case SourcePcap:
		ok = s.processPcap(r, t, m)
	case SourceWinLog:
		ok = s.processWinLog(r, t, m)
	}
	if ok {
		s.Processed++
	}
	return ok
}

// parseKV parses "k1=v1,k2=v2" as emitted by the tw* tools.
func parseKV(c string) map[string]string {
	c = strings.TrimSpace(c)
	if idx := strings.Index(c, "type="); idx >= 0 {
		c = c[idx:]
	} else {
		// Strip any leading tag or colon prefix: "twWifiScan: type=..." -> "type=..."
		for _, prefix := range []string{"twWifiScan", "twBlueScan", "twpcap", "twwinlog", "twsdrpower"} {
			if strings.HasPrefix(strings.ToLower(c), strings.ToLower(prefix)) {
				c = strings.TrimSpace(c[len(prefix):])
				if strings.HasPrefix(c, "[") {
					if end := strings.Index(c, "]"); end >= 0 {
						c = strings.TrimSpace(c[end+1:])
					}
				}
				break
			}
		}
		c = strings.TrimPrefix(c, ":")
		c = strings.TrimSpace(c)
	}
	m := map[string]string{}
	for _, kv := range strings.Split(c, ",") {
		a := strings.SplitN(kv, "=", 2)
		if len(a) == 2 {
			m[strings.TrimSpace(a[0])] = strings.TrimSpace(a[1])
		}
	}
	return m
}

func atoi(s string) int {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return int(f)
	}
	return n
}

func atof(s string) float64 {
	s = strings.TrimSpace(s)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// parseTime parses an RFC3339 time as emitted by the tw* tools; def is returned on failure.
func parseTime(s string, def int64) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		if t.IsZero() || t.Year() <= 1970 {
			return def
		}
		return t.UnixNano()
	}
	return def
}

func makeID(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}
