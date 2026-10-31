package logreport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/twsnmp/twsnmpfk/datastore"
)

// MQTTReporter is an optional extension of Reporter that receives client metadata with MQTT messages.
type MQTTReporter interface {
	ProcessMQTTMsg(clientID, remote, topic string, payload []byte)
}

// Reporter is the interface implemented by the report engine for real-time
// ingestion from various protocol receivers (Syslog, NetFlow, sFlow, SNMP Trap, MQTT, OTel, etc.).
type Reporter interface {
	// ProcessSyslog ingests a raw syslog message map.
	ProcessSyslog(sl map[string]interface{})
	// ProcessNetFlow ingests a NetFlow/IPFIX record or map.
	ProcessNetFlow(record any)
	// ProcessSFlow ingests an sFlow record or map.
	ProcessSFlow(record any)
	// ProcessTrap ingests an SNMP trap message or map.
	ProcessTrap(record any)
	// ProcessMQTT is a hook for MQTT sensor/topic report ingestion.
	ProcessMQTT(topic string, payload []byte)
	// ProcessOTel is a hook for OpenTelemetry report ingestion.
	ProcessOTel(record any)

	// Start starts background worker routines (queue processing, periodic cleanup).
	Start(ctx context.Context) error
	// Stop flushes remaining records and stops background workers.
	Stop()
}

// Engine processes incoming events asynchronously and updates the log report datastore.
type Engine struct {
	store         ReportStore
	syslogQueue   chan Record
	flowQueue     chan *FlowRecord
	trapQueue     chan *TrapRecord
	mqttQueue     chan *MQTTRecord
	stopCh        chan struct{}
	flushCh       chan chan struct{}
	wg            sync.WaitGroup
	cleanInterval time.Duration
	flushInterval time.Duration
	maxBatchSize  int
}

// Option configures the Engine.
type Option func(*Engine)

// WithCleanInterval overrides the default cleanup interval (1 hour).
func WithCleanInterval(d time.Duration) Option {
	return func(e *Engine) {
		e.cleanInterval = d
	}
}

// WithFlushInterval overrides the default batch flush interval (200ms).
func WithFlushInterval(d time.Duration) Option {
	return func(e *Engine) {
		e.flushInterval = d
	}
}

// NewEngine creates a new Engine writing report data to store.
func NewEngine(store ReportStore, opts ...Option) *Engine {
	e := &Engine{
		store:         store,
		syslogQueue:   make(chan Record, 10000),
		flowQueue:     make(chan *FlowRecord, 10000),
		trapQueue:     make(chan *TrapRecord, 10000),
		mqttQueue:     make(chan *MQTTRecord, 10000),
		stopCh:        make(chan struct{}),
		flushCh:       make(chan chan struct{}),
		cleanInterval: time.Hour,
		flushInterval: 200 * time.Millisecond,
		maxBatchSize:  500,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// ProcessSyslog parses a syslog message and enqueues it for report ingestion and statistics aggregation.
func (e *Engine) ProcessSyslog(sl map[string]interface{}) {
	if e == nil || sl == nil {
		return
	}
	tag, _ := sl["tag"].(string)
	if tag == "" {
		tag, _ = sl["app_name"].(string)
	}

	content, _ := sl["content"].(string)
	if content == "" {
		content, _ = sl["message"].(string)
	}
	if sd, ok := sl["structured_data"].(string); ok && sd != "" && sd != "-" && !strings.Contains(content, sd) {
		if content != "" {
			content = sd + " " + content
		} else {
			content = sd
		}
	}

	// If tag is empty or not directly matched, check if content starts with a known tag prefix
	src := TagToSource(tag)
	if src == "" && content != "" {
		cTrim := strings.TrimSpace(content)
		for _, prefix := range []string{"twWifiScan", "twBlueScan", "twpcap", "twwinlog", "twsdrpower", "twagentneo"} {
			if strings.HasPrefix(strings.ToLower(cTrim), strings.ToLower(prefix)) {
				tag = prefix
				break
			}
		}
		if tag == "" && strings.Contains(cTrim, "[twsnmpneo@17862") {
			tag = "twagentneo"
		}
	}

	var clientIP string
	if client, ok := sl["client"].(string); ok && client != "" {
		if h, _, err := net.SplitHostPort(client); err == nil {
			clientIP = h
		} else {
			clientIP = client
		}
	} else if r, ok := sl["remote"].(string); ok && r != "" {
		if h, _, err := net.SplitHostPort(r); err == nil {
			clientIP = h
		} else {
			clientIP = r
		}
	}

	host, _ := sl["hostname"].(string)
	if host == "" {
		host, _ = sl["host"].(string)
	}
	if host == "" {
		host = clientIP
	}
	if host == "" {
		host = "localhost"
	}

	tNano := time.Now().UnixNano()
	if ts, ok := sl["timestamp"].(time.Time); ok && !ts.IsZero() {
		tNano = ts.UnixNano()
	} else if tsUnix, ok := sl["time"].(int64); ok && tsUnix > 0 {
		tNano = tsUnix
	}

	sev := 6
	if v, ok := sl["severity"].(float64); ok {
		sev = int(v)
	} else if v, ok := sl["severity"].(int); ok {
		sev = v
	}

	fac := 1
	if v, ok := sl["facility"].(float64); ok {
		fac = int(v)
	} else if v, ok := sl["facility"].(int); ok {
		fac = v
	}

	rec := Record{
		Time:     tNano,
		Host:     host,
		Tag:      tag,
		Content:  content,
		Severity: sev,
		Facility: fac,
		Remote:   clientIP,
	}

	select {
	case e.syslogQueue <- rec:
	default:
		slog.Warn("logreport engine: syslog queue full, dropping record", "tag", tag, "host", host)
	}
}

// CleanTag extracts and cleans the syslog tag from a tag string.
func CleanTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.TrimSuffix(tag, ":")
	tag = strings.TrimSpace(tag)
	if idx := strings.Index(tag, "["); idx >= 0 {
		tag = strings.TrimSpace(tag[:idx])
	}
	return tag
}

// TagToSource maps syslog tag name to internal source identifier.
func TagToSource(tag string) string {
	cleaned := CleanTag(tag)
	switch strings.ToLower(cleaned) {
	case "twwifiscan":
		return SourceWifiScan
	case "twbluescan":
		return SourceBlueScan
	case "twpcap":
		return SourcePcap
	case "twwinlog":
		return SourceWinLog
	case "twagentneo", "twagentneo-fim", "twagentneo-sca", "twagentneo-resource":
		return SourceTwAgent
	default:
		return ""
	}
}

// ProcessNetFlow converts and enqueues a NetFlow/IPFIX record.
func (e *Engine) ProcessNetFlow(record any) {
	if e == nil || record == nil {
		return
	}
	var fr *FlowRecord
	switch r := record.(type) {
	case *datastore.NetFlowEnt:
		fr = &FlowRecord{
			Time:     r.Time,
			SrcIP:    r.SrcAddr,
			SrcPort:  r.SrcPort,
			DstIP:    r.DstAddr,
			DstPort:  r.DstPort,
			Protocol: r.Protocol,
			Packets:  int64(r.Packets),
			Bytes:    int64(r.Bytes),
			Duration: r.Dur,
			TCPFlags: string(r.TCPFlags),
		}
	case *FlowRecord:
		fr = r
	default:
		return
	}

	select {
	case e.flowQueue <- fr:
	default:
		slog.Warn("logreport engine: flow queue full, dropping record")
	}
}

// ProcessSFlow converts and enqueues an sFlow record.
func (e *Engine) ProcessSFlow(record any) {
	if e == nil || record == nil {
		return
	}
	var fr *FlowRecord
	switch r := record.(type) {
	case *datastore.SFlowEnt:
		fr = &FlowRecord{
			Time:     r.Time,
			SrcIP:    r.SrcAddr,
			SrcPort:  r.SrcPort,
			DstIP:    r.DstAddr,
			DstPort:  r.DstPort,
			Protocol: r.Protocol,
			Packets:  1,
			Bytes:    int64(r.Bytes),
			TCPFlags: string(r.TCPFlags),
		}
	case *FlowRecord:
		fr = r
	default:
		return
	}

	select {
	case e.flowQueue <- fr:
	default:
		slog.Warn("logreport engine: flow queue full, dropping record")
	}
}

// ProcessTrap converts and enqueues an SNMP trap message.
func (e *Engine) ProcessTrap(record any) {
	if e == nil || record == nil {
		return
	}
	var tr *TrapRecord
	switch r := record.(type) {
	case *TrapRecord:
		tr = r
	case map[string]interface{}:
		timeVal, _ := r["Time"].(int64)
		if timeVal == 0 {
			timeVal = time.Now().UnixNano()
		}
		from, _ := r["FromAddress"].(string)
		ttype, _ := r["TrapType"].(string)
		ent, _ := r["Enterprise"].(string)
		vars, _ := r["Variables"].(string)
		lvl, _ := r["Level"].(string)
		tr = &TrapRecord{
			Time:        timeVal,
			FromAddress: from,
			TrapType:    ttype,
			Enterprise:  ent,
			Variables:   vars,
			Level:       lvl,
		}
	default:
		if b, err := json.Marshal(record); err == nil {
			var parsed TrapRecord
			if json.Unmarshal(b, &parsed) == nil && (parsed.TrapType != "" || parsed.Variables != "" || parsed.FromAddress != "") {
				if parsed.Time == 0 {
					parsed.Time = time.Now().UnixNano()
				}
				tr = &parsed
			}
		}
		if tr == nil {
			return
		}
	}

	select {
	case e.trapQueue <- tr:
	default:
		slog.Warn("logreport engine: trap queue full, dropping record")
	}
}

// ProcessMQTT ingests an MQTT topic and payload for report aggregation.
func (e *Engine) ProcessMQTT(topic string, payload []byte) {
	e.ProcessMQTTMsg("", "", topic, payload)
}

// ProcessMQTTMsg ingests an MQTT topic and payload with optional client metadata.
func (e *Engine) ProcessMQTTMsg(clientID, remote, topic string, payload []byte) {
	if e == nil || topic == "" {
		return
	}
	cp := make([]byte, len(payload))
	copy(cp, payload)
	mr := &MQTTRecord{
		Time:     time.Now().UnixNano(),
		Topic:    topic,
		ClientID: clientID,
		Remote:   remote,
		Payload:  cp,
	}
	select {
	case e.mqttQueue <- mr:
	default:
		slog.Warn("logreport engine: mqtt queue full, dropping record", "topic", topic)
	}
}

// ProcessOTel is a hook for future OpenTelemetry metrics / trace ingestion.
func (e *Engine) ProcessOTel(record any) {
	// To be expanded in upcoming iterations
}

// Start launches the ingestion worker loop and periodic cleanup scheduler.
func (e *Engine) Start(ctx context.Context) error {
	e.wg.Add(1)
	go e.workerLoop(ctx)
	return nil
}

// Stop stops workers and waits for queue drain and flush.
func (e *Engine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

// Flush triggers an immediate synchronous commit of all queued items.
func (e *Engine) Flush() {
	done := make(chan struct{})
	select {
	case e.flushCh <- done:
		<-done
	case <-time.After(5 * time.Second):
	}
}

func (e *Engine) workerLoop(ctx context.Context) {
	defer e.wg.Done()

	flushTicker := time.NewTicker(e.flushInterval)
	defer flushTicker.Stop()

	cleanTicker := time.NewTicker(e.cleanInterval)
	defer cleanTicker.Stop()

	var session *Session
	ensureSession := func() *Session {
		if session == nil {
			session = NewSession(ctx, e.store)
		}
		return session
	}

	commitSession := func() {
		if session != nil && session.Processed > 0 {
			if err := session.Commit(); err != nil {
				slog.Error("logreport engine: failed to commit session", "error", err)
			}
			session = nil
		}
	}

	drainAll := func() {
		for {
			select {
			case rec := <-e.syslogQueue:
				s := ensureSession()
				src := TagToSource(rec.Tag)
				if src != "" {
					s.Process(src, rec)
				}
				s.ProcessSyslogStat(rec, rec.Facility)
			case fr := <-e.flowQueue:
				ensureSession().ProcessFlow(fr)
			case tr := <-e.trapQueue:
				s := ensureSession()
				s.ProcessTrapStat(tr)
				s.processTwAgentTrap(tr)
			case mr := <-e.mqttQueue:
				ensureSession().ProcessTwAgentMQTT(mr)
			default:
				commitSession()
				return
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			drainAll()
			return
		case <-e.stopCh:
			drainAll()
			return
		case done := <-e.flushCh:
			drainAll()
			close(done)
		case rec := <-e.syslogQueue:
			s := ensureSession()
			src := TagToSource(rec.Tag)
			if src != "" {
				s.Process(src, rec)
			}
			s.ProcessSyslogStat(rec, rec.Facility)
			if s.Processed >= e.maxBatchSize {
				commitSession()
			}
		case fr := <-e.flowQueue:
			s := ensureSession()
			s.ProcessFlow(fr)
			if s.Processed >= e.maxBatchSize {
				commitSession()
			}
		case tr := <-e.trapQueue:
			s := ensureSession()
			s.ProcessTrapStat(tr)
			s.processTwAgentTrap(tr)
			if s.Processed >= e.maxBatchSize {
				commitSession()
			}
		case mr := <-e.mqttQueue:
			s := ensureSession()
			s.ProcessTwAgentMQTT(mr)
			if s.Processed >= e.maxBatchSize {
				commitSession()
			}
		case <-flushTicker.C:
			commitSession()
		case <-cleanTicker.C:
			commitSession()
			if e.store != nil {
				_ = CleanupAll(ctx, e.store, 0, 0)
			}
		}
	}
}
