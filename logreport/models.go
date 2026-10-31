// Package logreport builds the "report" datasets (Wi-Fi APs, Bluetooth devices,
// packet-capture summaries and Windows event summaries) from protocol records
// emitted by twWifiScan, twBlueScan, twpcap, twwinlog, twagentneo, etc.
//
// Records are processed in real time when received by protocol servers (matching
// TWSNMP NEO / TWSNMP FC behavior) through the Engine / Reporter pipeline.
package logreport

import "strings"

// Polling types / sources handled by this package.
const (
	SourceWifiScan = "twwifiscan"
	SourceBlueScan = "twbluescan"
	SourcePcap     = "twpcap"
	SourceWinLog   = "twwinlog"
	SourceTwAgent  = "twagentneo"
)

// Report kinds (storage buckets / API names).
const (
	KindWifiAP       = "wifiAP"
	KindBlueDevice   = "blueDevice"
	KindEnvMonitor   = "envMonitor"
	KindPowerMonitor = "powerMonitor"
	KindMotionSensor = "motionSensor"
	KindEtherType    = "etherType"
	KindDNSQ         = "dnsq"
	KindRADIUSFlow   = "radiusFlow"
	KindTLSFlow      = "tlsFlow"
	KindWinEventID   = "winEventID"
	KindWinLogon     = "winLogon"
	KindWinAccount   = "winAccount"
	KindWinKerberos  = "winKerberos"
	KindWinPrivilege = "winPrivilege"
	KindWinProcess   = "winProcess"
	KindWinTask      = "winTask"
	KindFlow         = "flow"
	KindServer       = "server"
	KindFumble       = "fumble"
	KindSyslogStats  = "syslogStats"
	KindTrapStats    = "trapStats"
	KindTwAgentInfo  = "twAgentInfo"
	KindTwAgentAlert = "twAgentAlert"
	KindTwAgentFIM   = "twAgentFIM"
	KindTwAgentSCA   = "twAgentSCA"
	KindTwAgentTCP   = "twAgentTCP"
)

// SyslogTag returns the syslog tag emitted by the source.
func SyslogTag(source string) string {
	switch strings.ToLower(source) {
	case SourceWifiScan:
		return "twWifiScan"
	case SourceBlueScan:
		return "twBlueScan"
	case SourcePcap:
		return "twpcap"
	case SourceWinLog:
		return "twwinlog"
	case SourceTwAgent:
		return "twagentneo"
	}
	return ""
}

// IsSource reports whether s is a known source / polling type.
func IsSource(s string) bool { return SyslogTag(s) != "" }

// KindsOf returns the report kinds fed by the source.
func KindsOf(source string) []string {
	switch strings.ToLower(source) {
	case SourceWifiScan:
		return []string{KindWifiAP}
	case SourceBlueScan:
		return []string{KindBlueDevice, KindEnvMonitor, KindPowerMonitor, KindMotionSensor}
	case SourcePcap:
		return []string{KindEtherType, KindDNSQ, KindRADIUSFlow, KindTLSFlow}
	case SourceWinLog:
		return []string{KindWinEventID, KindWinLogon, KindWinAccount, KindWinKerberos, KindWinPrivilege, KindWinProcess, KindWinTask}
	case SourceTwAgent:
		return []string{KindTwAgentInfo, KindTwAgentAlert, KindTwAgentFIM, KindTwAgentSCA, KindTwAgentTCP}
	}
	return nil
}

// AllKinds lists every report kind.
func AllKinds() []string {
	var ret []string
	for _, s := range []string{SourceWifiScan, SourceBlueScan, SourcePcap, SourceWinLog, SourceTwAgent} {
		ret = append(ret, KindsOf(s)...)
	}
	ret = append(ret, KindFlow, KindServer, KindFumble, KindSyslogStats, KindTrapStats)
	return ret
}

// NormalizeKind returns the canonical report kind for k, handling aliases and case-insensitivity.
func NormalizeKind(k string) string {
	if strings.EqualFold(k, "dnsQ") || strings.EqualFold(k, "dnsq") {
		return KindDNSQ
	}
	for _, v := range AllKinds() {
		if strings.EqualFold(v, k) {
			return v
		}
	}
	return k
}

// IsKind reports whether k is a known report kind.
func IsKind(k string) bool {
	norm := NormalizeKind(k)
	for _, v := range AllKinds() {
		if v == norm {
			return true
		}
	}
	return false
}

// ScoreInfo is embedded by entities that carry a penalty and a deviation score.
type ScoreInfo struct {
	Penalty    int     `json:"Penalty"`
	Score      float64 `json:"Score"`
	ValidScore bool    `json:"ValidScore"`
}

func (x *ScoreInfo) scoreInfo() *ScoreInfo { return x }

// RSSIEnt is one RSSI sample.
type RSSIEnt struct {
	Value int   `json:"Value"`
	Time  int64 `json:"Time"`
}

// WifiAPEnt is a Wi-Fi access point seen by twWifiScan.
type WifiAPEnt struct {
	ID        string    `json:"ID"`
	Host      string    `json:"Host"`
	BSSID     string    `json:"BSSID"`
	SSID      string    `json:"SSID"`
	Channel   string    `json:"Channel"`
	Info      string    `json:"Info"`
	Vendor    string    `json:"Vendor"`
	Count     int       `json:"Count"`
	Change    int       `json:"Change"`
	RSSI      []RSSIEnt `json:"RSSI"`
	FirstTime int64     `json:"FirstTime"`
	LastTime  int64     `json:"LastTime"`
}

// BlueDeviceEnt is a Bluetooth device seen by twBlueScan.
type BlueDeviceEnt struct {
	ID          string    `json:"ID"`
	Host        string    `json:"Host"`
	Address     string    `json:"Address"`
	AddressType string    `json:"AddressType"`
	Name        string    `json:"Name"`
	Vendor      string    `json:"Vendor"`
	Info        string    `json:"Info"`
	Count       int       `json:"Count"`
	RSSI        []RSSIEnt `json:"RSSI"`
	FirstTime   int64     `json:"FirstTime"`
	LastTime    int64     `json:"LastTime"`
}

// EnvDataEnt is one environment sensor sample.
type EnvDataEnt struct {
	Time               int64   `json:"Time"`
	RSSI               int     `json:"RSSI"`
	Temp               float64 `json:"Temp"`
	Humidity           float64 `json:"Humidity"`
	Illuminance        float64 `json:"Illuminance"`
	BarometricPressure float64 `json:"BarometricPressure"`
	Sound              float64 `json:"Sound"`
	ETVOC              float64 `json:"ETVOC"`
	ECo2               float64 `json:"ECo2"`
	Battery            int     `json:"Battery"`
}

// EnvMonitorEnt is a Bluetooth environment sensor (OMRON / SwitchBot / Inkbird).
type EnvMonitorEnt struct {
	ID        string       `json:"ID"`
	Host      string       `json:"Host"`
	Address   string       `json:"Address"`
	Name      string       `json:"Name"`
	Count     int          `json:"Count"`
	EnvData   []EnvDataEnt `json:"EnvData"`
	FirstTime int64        `json:"FirstTime"`
	LastTime  int64        `json:"LastTime"`
}

// PowerMonitorDataEnt is one smart plug sample.
type PowerMonitorDataEnt struct {
	Time   int64   `json:"Time"`
	Load   float64 `json:"Load"`
	Switch bool    `json:"Switch"`
	Over   bool    `json:"Over"`
	RSSI   int     `json:"RSSI"`
}

// PowerMonitorEnt is a SwitchBot plug mini.
type PowerMonitorEnt struct {
	ID        string                `json:"ID"`
	Host      string                `json:"Host"`
	Address   string                `json:"Address"`
	Name      string                `json:"Name"`
	Count     int                   `json:"Count"`
	Data      []PowerMonitorDataEnt `json:"Data"`
	FirstTime int64                 `json:"FirstTime"`
	LastTime  int64                 `json:"LastTime"`
}

// MotionSensorDataEnt is one motion sensor sample.
type MotionSensorDataEnt struct {
	Time         int64  `json:"Time"`
	Event        string `json:"Event"`
	Moving       bool   `json:"Moving"`
	Light        bool   `json:"Light"`
	Battery      int    `json:"Battery"`
	LastMove     int64  `json:"LastMove"`
	LastMoveDiff int    `json:"LastMoveDiff"`
	RSSI         int    `json:"RSSI"`
}

// MotionSensorEnt is a SwitchBot motion sensor.
type MotionSensorEnt struct {
	ID        string                `json:"ID"`
	Host      string                `json:"Host"`
	Address   string                `json:"Address"`
	Name      string                `json:"Name"`
	Count     int                   `json:"Count"`
	Data      []MotionSensorDataEnt `json:"Data"`
	FirstTime int64                 `json:"FirstTime"`
	LastTime  int64                 `json:"LastTime"`
}

// EtherTypeEnt counts Ethernet frame types seen by twpcap.
type EtherTypeEnt struct {
	ID        string `json:"ID"`
	Host      string `json:"Host"`
	Type      string `json:"Type"`
	Name      string `json:"Name"`
	Count     int    `json:"Count"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// DNSQEnt is a DNS question/answer summary.
type DNSQEnt struct {
	ID         string `json:"ID"`
	Host       string `json:"Host"`
	Type       string `json:"Type"`
	Server     string `json:"Server"`
	Name       string `json:"Name"`
	Count      int    `json:"Count"`
	Change     int    `json:"Change"`
	LastClient string `json:"LastClient"`
	LastMAC    string `json:"LastMAC"`
	FirstTime  int64  `json:"FirstTime"`
	LastTime   int64  `json:"LastTime"`
}

// RADIUSFlowEnt is a RADIUS client/server flow.
type RADIUSFlowEnt struct {
	ID           string `json:"ID"`
	Client       string `json:"Client"`
	ClientName   string `json:"ClientName"`
	ClientNodeID string `json:"ClientNodeID"`
	Server       string `json:"Server"`
	ServerName   string `json:"ServerName"`
	ServerNodeID string `json:"ServerNodeID"`
	Accept       int    `json:"Accept"`
	Reject       int    `json:"Reject"`
	Request      int    `json:"Request"`
	Challenge    int    `json:"Challenge"`
	Count        int    `json:"Count"`
	ScoreInfo
	FirstTime  int64 `json:"FirstTime"`
	LastTime   int64 `json:"LastTime"`
	UpdateTime int64 `json:"UpdateTime"`
}

// TLSFlowEnt is a TLS client/server flow.
type TLSFlowEnt struct {
	ID           string `json:"ID"`
	Client       string `json:"Client"`
	ClientName   string `json:"ClientName"`
	ClientNodeID string `json:"ClientNodeID"`
	ClientLoc    string `json:"ClientLoc"`
	Server       string `json:"Server"`
	ServerName   string `json:"ServerName"`
	ServerNodeID string `json:"ServerNodeID"`
	ServerLoc    string `json:"ServerLoc"`
	Service      string `json:"Service"`
	Version      string `json:"Version"`
	Cipher       string `json:"Cipher"`
	Count        int    `json:"Count"`
	ScoreInfo
	FirstTime  int64 `json:"FirstTime"`
	LastTime   int64 `json:"LastTime"`
	UpdateTime int64 `json:"UpdateTime"`
}

// WinEventIDEnt summarises one Windows event ID.
type WinEventIDEnt struct {
	ID        string `json:"ID"`
	Level     string `json:"Level"`
	Provider  string `json:"Provider"`
	EventID   int    `json:"EventID"`
	Computer  string `json:"Computer"`
	Channel   string `json:"Channel"`
	Count     int    `json:"Count"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// WinLogonEnt summarises logons for a target/computer/IP.
type WinLogonEnt struct {
	ID         string         `json:"ID"`
	Target     string         `json:"Target"`
	Computer   string         `json:"Computer"`
	IP         string         `json:"IP"`
	Count      int            `json:"Count"`
	Logon      int            `json:"Logon"`
	Logoff     int            `json:"Logoff"`
	Failed     int            `json:"Failed"`
	LogonType  map[string]int `json:"LogonType"`
	FailedCode map[string]int `json:"FailedCode"`
	ScoreInfo
	FirstTime int64 `json:"FirstTime"`
	LastTime  int64 `json:"LastTime"`
}

// WinAccountEnt summarises account operations.
type WinAccountEnt struct {
	ID        string `json:"ID"`
	Subject   string `json:"Subject"`
	Target    string `json:"Target"`
	Computer  string `json:"Computer"`
	Count     int    `json:"Count"`
	Edit      int    `json:"Edit"`
	Password  int    `json:"Password"`
	Other     int    `json:"Other"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// WinKerberosEnt summarises Kerberos ticket requests.
type WinKerberosEnt struct {
	ID         string `json:"ID"`
	Target     string `json:"Target"`
	Computer   string `json:"Computer"`
	IP         string `json:"IP"`
	Service    string `json:"Service"`
	TicketType string `json:"TicketType"`
	Count      int    `json:"Count"`
	Failed     int    `json:"Failed"`
	ScoreInfo
	FirstTime int64 `json:"FirstTime"`
	LastTime  int64 `json:"LastTime"`
}

// WinPrivilegeEnt summarises special privilege use.
type WinPrivilegeEnt struct {
	ID        string `json:"ID"`
	Subject   string `json:"Subject"`
	Computer  string `json:"Computer"`
	Count     int    `json:"Count"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// WinProcessEnt summarises process start/exit events.
type WinProcessEnt struct {
	ID          string `json:"ID"`
	Process     string `json:"Process"`
	Computer    string `json:"Computer"`
	Count       int    `json:"Count"`
	Start       int    `json:"Start"`
	Exit        int    `json:"Exit"`
	LastSubject string `json:"LastSubject"`
	LastParent  string `json:"LastParent"`
	LastStatus  string `json:"LastStatus"`
	FirstTime   int64  `json:"FirstTime"`
	LastTime    int64  `json:"LastTime"`
}

// WinTaskEnt summarises scheduled task events.
type WinTaskEnt struct {
	ID        string `json:"ID"`
	TaskName  string `json:"TaskName"`
	Computer  string `json:"Computer"`
	Subject   string `json:"Subject"`
	Count     int    `json:"Count"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// FlowEnt represents a flow/session communication between client and server.
type FlowEnt struct {
	ID           string           `json:"ID"`
	Client       string           `json:"Client"`
	ClientName   string           `json:"ClientName"`
	ClientNodeID string           `json:"ClientNodeID"`
	ClientLoc    string           `json:"ClientLoc"`
	Server       string           `json:"Server"`
	ServerName   string           `json:"ServerName"`
	ServerNodeID string           `json:"ServerNodeID"`
	ServerLoc    string           `json:"ServerLoc"`
	Services     map[string]int64 `json:"Services"`
	Count        int64            `json:"Count"`
	Bytes        int64            `json:"Bytes"`
	Packets      int64            `json:"Packets"`
	Duration     float64          `json:"Duration"`
	ScoreInfo
	FirstTime  int64 `json:"FirstTime"`
	LastTime   int64 `json:"LastTime"`
	UpdateTime int64 `json:"UpdateTime"`
}

func (x *FlowEnt) scoreInfo() *ScoreInfo { return &x.ScoreInfo }

// ServerEnt represents aggregated server service statistics.
type ServerEnt struct {
	ID           string           `json:"ID"`
	Server       string           `json:"Server"`
	ServerName   string           `json:"ServerName"`
	ServerNodeID string           `json:"ServerNodeID"`
	Loc          string           `json:"Loc"`
	Services     map[string]int64 `json:"Services"`
	Count        int64            `json:"Count"`
	Bytes        int64            `json:"Bytes"`
	Packets      int64            `json:"Packets"`
	ScoreInfo
	FirstTime  int64 `json:"FirstTime"`
	LastTime   int64 `json:"LastTime"`
	UpdateTime int64 `json:"UpdateTime"`
}

func (x *ServerEnt) scoreInfo() *ScoreInfo { return &x.ScoreInfo }

// FumbleEnt represents suspicious/failed connections (small TCP packet flows, ICMP error replies).
type FumbleEnt struct {
	ID        string `json:"ID"`
	TCPCount  int64  `json:"TCPCount"`
	IcmpCount int64  `json:"IcmpCount"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// SyslogHostStat holds syslog statistics for a host.
type SyslogHostStat struct {
	Host        string `json:"Host"`
	NodeName    string `json:"NodeName"`
	Count       int64  `json:"Count"`
	ErrorCount  int64  `json:"ErrorCount"`
	WarnCount   int64  `json:"WarnCount"`
	NormalCount int64  `json:"NormalCount"`
	FirstTime   int64  `json:"FirstTime"`
	LastTime    int64  `json:"LastTime"`
}

// SyslogTagStat holds syslog statistics for a tag.
type SyslogTagStat struct {
	Tag       string `json:"Tag"`
	Count     int64  `json:"Count"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// SyslogStatsSummary represents the full aggregated syslog report.
type SyslogStatsSummary struct {
	ID          string                     `json:"ID"` // "summary"
	Total       int64                      `json:"Total"`
	ErrorCount  int64                      `json:"ErrorCount"`
	WarnCount   int64                      `json:"WarnCount"`
	NormalCount int64                      `json:"NormalCount"`
	Hosts       map[string]*SyslogHostStat `json:"Hosts"`
	Tags        map[string]*SyslogTagStat  `json:"Tags"`
	Facilities  map[int]int64              `json:"Facilities"`
	Severities  map[int]int64              `json:"Severities"`
	FirstTime   int64                      `json:"FirstTime"`
	LastTime    int64                      `json:"LastTime"`
	UpdateTime  int64                      `json:"UpdateTime"`
}

// TrapHostStat holds SNMP trap statistics for a host.
type TrapHostStat struct {
	Host        string `json:"Host"`
	NodeName    string `json:"NodeName"`
	Count       int64  `json:"Count"`
	ErrorCount  int64  `json:"ErrorCount"`
	WarnCount   int64  `json:"WarnCount"`
	NormalCount int64  `json:"NormalCount"`
	FirstTime   int64  `json:"FirstTime"`
	LastTime    int64  `json:"LastTime"`
}

// TrapTypeStat holds SNMP trap statistics for a trap type / OID.
type TrapTypeStat struct {
	TrapType  string `json:"TrapType"`
	Count     int64  `json:"Count"`
	FirstTime int64  `json:"FirstTime"`
	LastTime  int64  `json:"LastTime"`
}

// TrapStatsSummary represents the full aggregated SNMP trap report.
type TrapStatsSummary struct {
	ID          string                   `json:"ID"` // "summary"
	Total       int64                    `json:"Total"`
	ErrorCount  int64                    `json:"ErrorCount"`
	WarnCount   int64                    `json:"WarnCount"`
	NormalCount int64                    `json:"NormalCount"`
	Hosts       map[string]*TrapHostStat `json:"Hosts"`
	Types       map[string]*TrapTypeStat `json:"Types"`
	Enterprises map[string]int64         `json:"Enterprises"`
	FirstTime   int64                    `json:"FirstTime"`
	LastTime    int64                    `json:"LastTime"`
	UpdateTime  int64                    `json:"UpdateTime"`
}

// TwAgentDiskEnt holds per-mount disk usage and I/O telemetry reported by twagentneo.
type TwAgentDiskEnt struct {
	MountPoint  string  `json:"MountPoint"`
	FsType      string  `json:"FsType,omitempty"`
	TotalBytes  uint64  `json:"TotalBytes"`
	UsedBytes   uint64  `json:"UsedBytes"`
	UsedPercent float64 `json:"UsedPercent"`
	ReadBps     float64 `json:"ReadBps"`
	WriteBps    float64 `json:"WriteBps"`
}

// TwAgentIfaceEnt holds per-interface configuration and traffic rates reported by twagentneo.
type TwAgentIfaceEnt struct {
	Index      int      `json:"Index,omitempty"`
	Name       string   `json:"Name"`
	MTU        int      `json:"MTU,omitempty"`
	MAC        string   `json:"MAC,omitempty"`
	Status     string   `json:"Status,omitempty"`
	Addrs      []string `json:"Addrs,omitempty"`
	BytesRecv  uint64   `json:"BytesRecv,omitempty"`
	BytesSent  uint64   `json:"BytesSent,omitempty"`
	RxMbps     float64  `json:"RxMbps"`
	TxMbps     float64  `json:"TxMbps"`
	RxPps      float64  `json:"RxPps,omitempty"`
	TxPps      float64  `json:"TxPps,omitempty"`
	RxErrorPps float64  `json:"RxErrorPps,omitempty"`
	TxErrorPps float64  `json:"TxErrorPps,omitempty"`
	RxDropPps  float64  `json:"RxDropPps,omitempty"`
	TxDropPps  float64  `json:"TxDropPps,omitempty"`
}

// TwAgentSampleEnt is one historical resource/traffic snapshot for a twagentneo instance.
type TwAgentSampleEnt struct {
	Time   int64   `json:"Time"`
	CPU    float64 `json:"CPU"`
	Mem    float64 `json:"Mem"`
	Load1  float64 `json:"Load1"`
	Disk   float64 `json:"Disk"`
	RxMbps float64 `json:"RxMbps"`
	TxMbps float64 `json:"TxMbps"`
}

// TwAgentInfoEnt represents an inventory and telemetry summary of a twagentneo agent.
type TwAgentInfoEnt struct {
	ID              string             `json:"ID"`
	AgentID         string             `json:"AgentID"`
	Host            string             `json:"Host"`
	Hostname        string             `json:"Hostname"`
	IP              string             `json:"IP"`
	MAC             string             `json:"MAC"`
	NodeID          string             `json:"NodeID"`
	NodeName        string             `json:"NodeName"`
	OS              string             `json:"OS"`
	Version         string             `json:"Version"`
	Status          string             `json:"Status"` // "online" | "offline"
	Protocols       []string           `json:"Protocols"`
	CPUPercent      float64            `json:"CPUPercent"`
	PerCoreCPU      []float64          `json:"PerCoreCPU,omitempty"`
	LoadAvg1        float64            `json:"LoadAvg1"`
	LoadAvg5        float64            `json:"LoadAvg5"`
	LoadAvg15       float64            `json:"LoadAvg15"`
	MemTotalBytes   uint64             `json:"MemTotalBytes"`
	MemUsedBytes    uint64             `json:"MemUsedBytes"`
	MemPercent      float64            `json:"MemPercent"`
	SwapUsedBytes   uint64             `json:"SwapUsedBytes"`
	DiskUsedPercent float64            `json:"DiskUsedPercent"`
	RxMbps          float64            `json:"RxMbps"`
	TxMbps          float64            `json:"TxMbps"`
	Disks           []TwAgentDiskEnt   `json:"Disks,omitempty"`
	Interfaces      []TwAgentIfaceEnt  `json:"Interfaces,omitempty"`
	Samples         []TwAgentSampleEnt `json:"Samples,omitempty"`
	AlertCount      int                `json:"AlertCount"`
	ThreatCount     int                `json:"ThreatCount"`
	FIMCount        int                `json:"FIMCount"`
	SCACount        int                `json:"SCACount"`
	Count           int                `json:"Count"`
	FirstTime       int64              `json:"FirstTime"`
	LastTime        int64              `json:"LastTime"`
}

// TwAgentAlertEnt represents an aggregated Sigma + tensai local LLM security/resource alert from twagentneo.
type TwAgentAlertEnt struct {
	ID                 string   `json:"ID"`
	AgentID            string   `json:"AgentID"`
	Host               string   `json:"Host"`
	IP                 string   `json:"IP"`
	NodeID             string   `json:"NodeID"`
	NodeName           string   `json:"NodeName"`
	Protocol           string   `json:"Protocol"` // "mqtt" | "syslog" | "trap" or comma-separated
	Protocols          []string `json:"Protocols,omitempty"`
	RuleID             string   `json:"RuleID"`
	RuleTitle          string   `json:"RuleTitle"`
	Level              string   `json:"Level"` // "critical" | "high" | "medium" | "warn" | "low" | "info"
	Category           string   `json:"Category"`
	SourceLog          string   `json:"SourceLog"`
	AIJudged           bool     `json:"AIJudged"`
	IsThreat           bool     `json:"IsThreat"`
	ComplianceViolated bool     `json:"ComplianceViolated"`
	ComplianceTypes    []string `json:"ComplianceTypes,omitempty"`
	Confidence         float64  `json:"Confidence"`
	SummaryJa          string   `json:"SummaryJa"`
	ActionJa           string   `json:"ActionJa"`
	Count              int      `json:"Count"`
	FirstTime          int64    `json:"FirstTime"`
	LastTime           int64    `json:"LastTime"`
}

// TwAgentFIMEnt represents a File Integrity Monitoring event detected by twagentneo.
type TwAgentFIMEnt struct {
	ID        string   `json:"ID"`
	AgentID   string   `json:"AgentID"`
	Host      string   `json:"Host"`
	IP        string   `json:"IP"`
	NodeID    string   `json:"NodeID"`
	NodeName  string   `json:"NodeName"`
	Protocol  string   `json:"Protocol"` // "mqtt" | "syslog" | "trap" or comma-separated
	Protocols []string `json:"Protocols,omitempty"`
	Op        string   `json:"Op"` // "CREATE" | "WRITE" | "REMOVE" | "CHMOD" | "RENAME"
	Path      string   `json:"Path"`
	Hash      string   `json:"Hash"`
	Count     int      `json:"Count"`
	FirstTime int64    `json:"FirstTime"`
	LastTime  int64    `json:"LastTime"`
}

// TwAgentSCAEnt represents a Security Configuration Assessment (SCA) finding detected by twagentneo.
type TwAgentSCAEnt struct {
	ID                 string   `json:"ID"`
	AgentID            string   `json:"AgentID"`
	Host               string   `json:"Host"`
	IP                 string   `json:"IP"`
	NodeID             string   `json:"NodeID"`
	NodeName           string   `json:"NodeName"`
	Protocol           string   `json:"Protocol"` // "mqtt" | "syslog" | "trap" or comma-separated
	Protocols          []string `json:"Protocols,omitempty"`
	RuleID             string   `json:"RuleID"`
	Title              string   `json:"Title"`
	Description        string   `json:"Description"`
	Severity           string   `json:"Severity"` // "critical" | "high" | "medium" | "low" | "info"
	FilePath           string   `json:"FilePath"`
	Proof              string   `json:"Proof"`
	Remediation        string   `json:"Remediation,omitempty"`
	Compliance         []string `json:"Compliance,omitempty"`
	AIJudged           bool     `json:"AIJudged"`
	IsThreat           bool     `json:"IsThreat"`
	ComplianceViolated bool     `json:"ComplianceViolated"`
	Confidence         float64  `json:"Confidence"`
	SummaryJa          string   `json:"SummaryJa,omitempty"`
	ActionJa           string   `json:"ActionJa,omitempty"`
	Count              int      `json:"Count"`
	FirstTime          int64    `json:"FirstTime"`
	LastTime           int64    `json:"LastTime"`
}

// TwAgentTCPEnt represents aggregated active TCP connections per process reported by twagentneo.
type TwAgentTCPEnt struct {
	ID        string   `json:"ID"`
	AgentID   string   `json:"AgentID"`
	Host      string   `json:"Host"`
	IP        string   `json:"IP"`
	NodeID    string   `json:"NodeID"`
	NodeName  string   `json:"NodeName"`
	Process   string   `json:"Process"`
	RemoteIP  string   `json:"RemoteIP"`
	RemoteLoc string   `json:"RemoteLoc"`
	Services  []string `json:"Services"`
	ConnCount int      `json:"ConnCount"`
	Count     int      `json:"Count"`
	ScoreInfo
	FirstTime int64 `json:"FirstTime"`
	LastTime  int64 `json:"LastTime"`
}

func (x *TwAgentTCPEnt) scoreInfo() *ScoreInfo { return &x.ScoreInfo }
