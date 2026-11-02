package logreport

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MQTTRecord represents an incoming MQTT publish message for report ingestion.
type MQTTRecord struct {
	Time     int64
	ClientID string
	Remote   string
	Topic    string
	Payload  []byte
}

var (
	twAgentSDRegex      = regexp.MustCompile(`([a-zA-Z0-9_]+)="((?:\\.|[^"\\])*)"`)
	twAgentTrapVarRegex = regexp.MustCompile(`(?:^|\s)([a-zA-Z0-9_]+(?:\.[0-9]+)*)=`)
)

// ProcessTwAgentMQTT processes MQTT messages published under twsnmpneo/agent/{agent_id}/...
func (s *Session) ProcessTwAgentMQTT(mr *MQTTRecord) bool {
	if mr == nil || mr.Topic == "" || len(mr.Payload) == 0 {
		return false
	}

	agentID, subTopic := parseTwAgentTopic(mr.Topic, mr.ClientID)
	if subTopic == "" {
		return false
	}

	tNano := mr.Time
	if tNano <= 0 {
		tNano = time.Now().UnixNano()
	}

	ip := cleanHostIP(mr.Remote)
	info := s.getOrCreateAgentInfo(agentID, "", ip, "mqtt", tNano)

	switch subTopic {
	case "metrics", "resource":
		var m struct {
			Timestamp     string    `json:"timestamp"`
			AgentID       string    `json:"agent_id"`
			Host          string    `json:"host"`
			Hostname      string    `json:"hostname"`
			OS            string    `json:"os"`
			Version       string    `json:"version"`
			IP            string    `json:"ip"`
			MAC           string    `json:"mac"`
			CPUPercent    float64   `json:"cpu_percent"`
			PerCoreCPU    []float64 `json:"per_core_cpu"`
			LoadAvg1      float64   `json:"load_avg_1"`
			LoadAvg5      float64   `json:"load_avg_5"`
			LoadAvg15     float64   `json:"load_avg_15"`
			MemTotalBytes uint64    `json:"mem_total_bytes"`
			MemUsedBytes  uint64    `json:"mem_used_bytes"`
			MemPercent    float64   `json:"mem_percent"`
			SwapUsedBytes uint64    `json:"swap_used_bytes"`
			Disks         []struct {
				MountPoint  string  `json:"mount_point"`
				FsType      string  `json:"fs_type"`
				TotalBytes  uint64  `json:"total_bytes"`
				UsedBytes   uint64  `json:"used_bytes"`
				UsedPercent float64 `json:"used_percent"`
				ReadBps     float64 `json:"read_bps"`
				WriteBps    float64 `json:"write_bps"`
			} `json:"disks"`
			Interfaces []struct {
				Name        string  `json:"name"`
				BytesRecv   uint64  `json:"bytes_recv"`
				BytesSent   uint64  `json:"bytes_sent"`
				PacketsRecv uint64  `json:"packets_recv"`
				PacketsSent uint64  `json:"packets_sent"`
				ErrIn       uint64  `json:"err_in"`
				ErrOut      uint64  `json:"err_out"`
				DropIn      uint64  `json:"drop_in"`
				DropOut     uint64  `json:"drop_out"`
				RxMbps      float64 `json:"rx_mbps"`
				TxMbps      float64 `json:"tx_mbps"`
			} `json:"interfaces"`
		}
		if err := json.Unmarshal(mr.Payload, &m); err != nil {
			return false
		}
		hn := m.Hostname
		if hn == "" {
			hn = m.Host
		}
		if hn != "" && (info.Hostname == "" || info.Hostname == info.IP) {
			info.Hostname = hn
			info.Host = hn
		}
		if m.OS != "" && info.OS == "" {
			info.OS = m.OS
		}
		if m.Version != "" && info.Version == "" {
			info.Version = m.Version
		}
		if m.IP != "" && (info.IP == "" || info.IP == "127.0.0.1") {
			info.IP = m.IP
		}
		if m.MAC != "" && info.MAC == "" {
			info.MAC = m.MAC
		}
		tRec := parseTime(m.Timestamp, tNano)
		info.CPUPercent = m.CPUPercent
		if len(m.PerCoreCPU) > 0 {
			info.PerCoreCPU = m.PerCoreCPU
		}
		info.LoadAvg1 = m.LoadAvg1
		info.LoadAvg5 = m.LoadAvg5
		info.LoadAvg15 = m.LoadAvg15
		info.MemTotalBytes = m.MemTotalBytes
		info.MemUsedBytes = m.MemUsedBytes
		info.MemPercent = m.MemPercent
		info.SwapUsedBytes = m.SwapUsedBytes
		info.Status = "online"

		var maxDisk float64
		disks := make([]TwAgentDiskEnt, 0, len(m.Disks))
		for _, d := range m.Disks {
			if d.UsedPercent > maxDisk {
				maxDisk = d.UsedPercent
			}
			disks = append(disks, TwAgentDiskEnt{
				MountPoint:  d.MountPoint,
				FsType:      d.FsType,
				TotalBytes:  d.TotalBytes,
				UsedBytes:   d.UsedBytes,
				UsedPercent: d.UsedPercent,
				ReadBps:     d.ReadBps,
				WriteBps:    d.WriteBps,
			})
			// Infer OS from mount point if OS is not yet set
			if info.OS == "" {
				if strings.Contains(d.MountPoint, ":") || strings.HasPrefix(d.MountPoint, "C") {
					info.OS = "windows"
				} else if d.MountPoint == "/System/Volumes/Data" {
					info.OS = "darwin"
				}
			}
		}
		if len(disks) > 0 {
			info.Disks = disks
			info.DiskPercent = maxDisk
			info.DiskUsedPercent = maxDisk
		}

		var rxMbps, txMbps float64
		var pktRecv, pktSent uint64
		for _, iface := range m.Interfaces {
			if isLoopbackName(iface.Name) {
				continue
			}
			rxMbps += iface.RxMbps
			txMbps += iface.TxMbps
			pktRecv += iface.PacketsRecv
			pktSent += iface.PacketsSent
		}
		if len(m.Interfaces) > 0 {
			info.RxMbps = rxMbps
			info.TxMbps = txMbps
		}

		// Avoid duplicate sample when "resource" and "metrics" mirror topics arrive together
		if len(info.Samples) == 0 || (tRec-info.Samples[len(info.Samples)-1].Time) > int64(2*time.Second) {
			info.Samples = append(info.Samples, TwAgentSampleEnt{
				Time:   tRec,
				CPU:    m.CPUPercent,
				Mem:    m.MemPercent,
				Load1:  m.LoadAvg1,
				Disk:   maxDisk,
				RxMbps: info.RxMbps,
				TxMbps: info.TxMbps,
			})
			if len(info.Samples) > MaxSeriesSize {
				info.Samples = info.Samples[len(info.Samples)-MaxSeriesSize:]
			}
			info.Count++
		}
		if tRec > info.LastTime {
			info.LastTime = tRec
		}
		putEnt(s, KindTwAgentInfo, info.ID, info)
		s.Processed++

		// Feed SensorMonitor on "metrics" topic for SensorReport compatibility
		if subTopic == "metrics" && s.store != nil {
			sensorHost := info.Hostname
			if sensorHost == "" {
				sensorHost = info.IP
			}
			if sensorHost == "" {
				sensorHost = info.AgentID
			}
			s.store.CheckSensorMonitor(sensorHost, "twagentneo", info.AgentID,
				m.CPUPercent, m.MemPercent, m.LoadAvg1, info.TxMbps, info.RxMbps, int64(pktSent), int64(pktRecv), 0)
		}
		return true

	case "interface", "network":
		var rawIfaces []struct {
			Index       int      `json:"index"`
			Name        string   `json:"name"`
			MTU         int      `json:"mtu"`
			MAC         string   `json:"mac"`
			Status      string   `json:"status"`
			Addrs       []string `json:"addrs"`
			BytesRecv   uint64   `json:"bytes_recv"`
			BytesSent   uint64   `json:"bytes_sent"`
			RxMbps      float64  `json:"rx_mbps"`
			TxMbps      float64  `json:"tx_mbps"`
			RxPps       float64  `json:"rx_pps"`
			TxPps       float64  `json:"tx_pps"`
			RxErrorPps  float64  `json:"rx_error_pps"`
			TxErrorPps  float64  `json:"tx_error_pps"`
			RxDropPps   float64  `json:"rx_drop_pps"`
			TxDropPps   float64  `json:"tx_drop_pps"`
		}
		if err := json.Unmarshal(mr.Payload, &rawIfaces); err != nil {
			return false
		}
		ifaces := make([]TwAgentIfaceEnt, 0, len(rawIfaces))
		var totalRx, totalTx float64
		for _, ri := range rawIfaces {
			ifaces = append(ifaces, TwAgentIfaceEnt{
				Index:      ri.Index,
				Name:       ri.Name,
				MTU:        ri.MTU,
				MAC:        ri.MAC,
				Status:     ri.Status,
				Addrs:      ri.Addrs,
				BytesRecv:  ri.BytesRecv,
				BytesSent:  ri.BytesSent,
				RxMbps:     ri.RxMbps,
				TxMbps:     ri.TxMbps,
				RxPps:      ri.RxPps,
				TxPps:      ri.TxPps,
				RxErrorPps: ri.RxErrorPps,
				TxErrorPps: ri.TxErrorPps,
				RxDropPps:  ri.RxDropPps,
				TxDropPps:  ri.TxDropPps,
			})
			if !isLoopbackName(ri.Name) {
				totalRx += ri.RxMbps
				totalTx += ri.TxMbps
				if info.MAC == "" && ri.MAC != "" && ri.Status == "up" {
					info.MAC = ri.MAC
				}
				if (info.IP == "" || info.IP == "127.0.0.1" || info.IP == "::1") && ri.Status == "up" {
					for _, a := range ri.Addrs {
						cleanA := strings.Split(a, "/")[0]
						if parsed := net.ParseIP(cleanA); parsed != nil && parsed.To4() != nil && !parsed.IsLoopback() {
							info.IP = cleanA
							s.enrichAgentNode(info)
							break
						}
					}
				}
			}
		}
		info.Interfaces = ifaces
		info.RxMbps = totalRx
		info.TxMbps = totalTx
		if tNano > info.LastTime {
			info.LastTime = tNano
		}
		putEnt(s, KindTwAgentInfo, info.ID, info)
		s.Processed++
		return true

	case "status", "heartbeat":
		var st struct {
			AgentID   string `json:"agent_id"`
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal(mr.Payload, &st); err != nil {
			return false
		}
		if st.Status != "" {
			info.Status = strings.ToLower(st.Status)
		}
		tRec := parseTime(st.Timestamp, tNano)
		if tRec > info.LastTime {
			info.LastTime = tRec
		}
		putEnt(s, KindTwAgentInfo, info.ID, info)
		s.Processed++

		if s.store != nil && info.Status == "online" {
			sensorHost := info.Hostname
			if sensorHost == "" {
				sensorHost = info.IP
			}
			if sensorHost == "" {
				sensorHost = info.AgentID
			}
			s.store.UpdateSensor(sensorHost, "twagentneo", info.AgentID, 1)
		}
		return true

	case "tcp":
		var conns []struct {
			RemoteIP string   `json:"remote_ip"`
			Process  string   `json:"process"`
			Services []string `json:"services"`
			Count    int      `json:"count"`
		}
		if err := json.Unmarshal(mr.Payload, &conns); err != nil {
			return false
		}
		hostLabel := info.Hostname
		if hostLabel == "" {
			hostLabel = info.AgentID
		}
		for _, c := range conns {
			if c.RemoteIP == "" {
				continue
			}
			proc := c.Process
			if proc == "" {
				proc = "unknown"
			}
			id := makeID(fmt.Sprintf("%s:%s:%s", info.AgentID, proc, c.RemoteIP))
			ent := getEnt[TwAgentTCPEnt](s, KindTwAgentTCP, id)
			if ent == nil {
				ent = &TwAgentTCPEnt{
					ID:        id,
					AgentID:   info.AgentID,
					Host:      hostLabel,
					IP:        info.IP,
					NodeID:    info.NodeID,
					NodeName:  info.NodeName,
					Process:   proc,
					RemoteIP:  c.RemoteIP,
					RemoteLoc: s.store.GetLoc(c.RemoteIP),
					Services:  c.Services,
					ConnCount: c.Count,
					Count:     1,
					FirstTime: tNano,
					LastTime:  tNano,
				}
			} else {
				ent.Host = hostLabel
				if info.IP != "" {
					ent.IP = info.IP
				}
				if info.NodeID != "" {
					ent.NodeID = info.NodeID
					ent.NodeName = info.NodeName
				}
				if ent.RemoteLoc == "" {
					ent.RemoteLoc = s.store.GetLoc(c.RemoteIP)
				}
				ent.Services = mergeStringSlice(ent.Services, c.Services)
				ent.ConnCount = c.Count
				ent.Count++
				if tNano > ent.LastTime {
					ent.LastTime = tNano
				}
			}
			ent.Penalty = calcTwAgentTCPPenalty(ent)
			putEnt(s, KindTwAgentTCP, id, ent)
			s.Processed++
		}
		return true

	case "fim":
		var fimEv struct {
			Timestamp string `json:"timestamp"`
			AgentID   string `json:"agent_id"`
			Hostname  string `json:"hostname"`
			Op        string `json:"op"`
			Path      string `json:"path"`
			Hash      string `json:"hash"`
			RuleTitle string `json:"rule_title"`
			SourceLog string `json:"source_log"`
		}
		if err := json.Unmarshal(mr.Payload, &fimEv); err != nil {
			return false
		}
		if fimEv.Hostname != "" && info.Hostname == "" {
			info.Hostname = fimEv.Hostname
		}
		if fimEv.AgentID != "" && (info.AgentID == "" || info.AgentID == info.IP) {
			info.AgentID = fimEv.AgentID
		}
		if fimEv.Path == "" && fimEv.SourceLog != "" {
			for _, token := range strings.Fields(fimEv.SourceLog) {
				if strings.HasPrefix(token, "path=") {
					fimEv.Path = strings.TrimPrefix(token, "path=")
				} else if strings.HasPrefix(token, "op=") && fimEv.Op == "" {
					fimEv.Op = strings.TrimPrefix(token, "op=")
				} else if strings.HasPrefix(token, "hash=") && fimEv.Hash == "" {
					fimEv.Hash = strings.TrimPrefix(token, "hash=")
				}
			}
		}
		if fimEv.Path == "" && fimEv.RuleTitle != "" {
			fimEv.Path = fimEv.RuleTitle
		}
		if fimEv.Path == "" {
			return false
		}
		if fimEv.Op == "" {
			fimEv.Op = "WRITE"
		}
		tRec := parseTime(fimEv.Timestamp, tNano)
		s.upsertTwAgentFIM(info, "mqtt", fimEv.Op, fimEv.Path, fimEv.Hash, tRec)
		return true

	case "sca":
		type aiJudgePayload struct {
			IsThreat            bool     `json:"is_threat"`
			ComplianceViolated  bool     `json:"compliance_violated"`
			ComplianceTypes     []string `json:"compliance_types"`
			Compliance          []string `json:"compliance"`
			Confidence          float64  `json:"confidence"`
			SummaryJa           string   `json:"summary_ja"`
			ActionJa            string   `json:"action_ja"`
			RecommendedActionJa string   `json:"recommended_action_ja"`
		}
		var scaEv struct {
			Timestamp   string          `json:"timestamp"`
			AgentID     string          `json:"agent_id"`
			Hostname    string          `json:"hostname"`
			RuleID      string          `json:"rule_id"`
			Title       string          `json:"title"`
			Description string          `json:"description"`
			Severity    string          `json:"severity"`
			FilePath    string          `json:"file_path"`
			Proof       string          `json:"proof"`
			Remediation string          `json:"remediation"`
			Compliance  []string        `json:"compliance"`
			AIJudge     *aiJudgePayload `json:"ai_judge"`
			AIJudgment  *aiJudgePayload `json:"ai_judgment"`
		}
		if err := json.Unmarshal(mr.Payload, &scaEv); err != nil {
			return false
		}
		if scaEv.Hostname != "" && info.Hostname == "" {
			info.Hostname = scaEv.Hostname
		}
		if scaEv.AgentID != "" && (info.AgentID == "" || info.AgentID == info.IP) {
			info.AgentID = scaEv.AgentID
		}
		tRec := parseTime(scaEv.Timestamp, tNano)
		judge := scaEv.AIJudge
		if judge == nil {
			judge = scaEv.AIJudgment
		}
		aiJudged := judge != nil
		var isThreat, compViol bool
		var conf float64
		var sumJa, actJa string
		if judge != nil {
			isThreat = judge.IsThreat
			compViol = judge.ComplianceViolated
			conf = judge.Confidence
			sumJa = judge.SummaryJa
			actJa = judge.ActionJa
			if actJa == "" {
				actJa = judge.RecommendedActionJa
			}
		} else {
			sev := strings.ToLower(scaEv.Severity)
			isThreat = sev == "critical" || sev == "high"
		}
		s.upsertTwAgentSCA(info, "mqtt", scaEv.RuleID, scaEv.Title, scaEv.Description, scaEv.Severity,
			scaEv.FilePath, scaEv.Proof, scaEv.Remediation, scaEv.Compliance,
			aiJudged, isThreat, compViol, conf, sumJa, actJa, tRec)
		return true

	case "alert":
		var raw map[string]interface{}
		if err := json.Unmarshal(mr.Payload, &raw); err != nil {
			return false
		}
		// 1. Check if it is a FIM event mirrored to "alert"
		category, _ := raw["category"].(string)
		ruleTitle, _ := raw["rule_title"].(string)
		sourceLog, _ := raw["source_log"].(string)
		op, _ := raw["op"].(string)
		path, _ := raw["path"].(string)
		hash, _ := raw["hash"].(string)
		if op == "" && sourceLog != "" {
			for _, token := range strings.Fields(sourceLog) {
				if strings.HasPrefix(token, "path=") && path == "" {
					path = strings.TrimPrefix(token, "path=")
				} else if strings.HasPrefix(token, "op=") && op == "" {
					op = strings.TrimPrefix(token, "op=")
				} else if strings.HasPrefix(token, "hash=") && hash == "" {
					hash = strings.TrimPrefix(token, "hash=")
				}
			}
		}
		if op == "" && strings.HasPrefix(ruleTitle, "File ") {
			rest := strings.TrimPrefix(ruleTitle, "File ")
			if parts := strings.SplitN(rest, ":", 2); len(parts) == 2 {
				op = strings.TrimSpace(parts[0])
				if path == "" {
					path = strings.TrimSpace(parts[1])
				}
			}
		}
		if op != "" || strings.EqualFold(category, "fim") || strings.HasPrefix(ruleTitle, "FIM:") {
			if path == "" {
				path = ruleTitle
			}
			if op == "" {
				op = "WRITE"
			}
			tsStr, _ := raw["timestamp"].(string)
			tRec := parseTime(tsStr, tNano)
			s.upsertTwAgentFIM(info, "mqtt", op, path, hash, tRec)
			return true
		}
		// 2. Check if it is a ResourceAlert
		if resName, ok := raw["resource"].(string); ok && resName != "" {
			val, _ := raw["value"].(float64)
			thresh, _ := raw["threshold"].(float64)
			mount, _ := raw["mount_path"].(string)
			tsStr, _ := raw["timestamp"].(string)
			tRec := parseTime(tsStr, tNano)
			title := fmt.Sprintf("Resource %s exceeded threshold: %.1f%% (threshold: %.1f%%)", resName, val, thresh)
			if mount != "" {
				title = fmt.Sprintf("Resource %s (%s) exceeded threshold: %.1f%% (threshold: %.1f%%)", resName, mount, val, thresh)
			}
			s.upsertTwAgentAlert(info, "mqtt", "resource-"+resName, title, "warn", "Resource", title,
				false, false, false, nil, 0, "", "", tRec)
			return true
		}
		// 3. SecurityAlert (Sigma + optional AIJudge)
		type aiJudgePayload struct {
			IsThreat            bool     `json:"is_threat"`
			ComplianceViolated  bool     `json:"compliance_violated"`
			ComplianceTypes     []string `json:"compliance_types"`
			Compliance          []string `json:"compliance"`
			Confidence          float64  `json:"confidence"`
			SummaryJa           string   `json:"summary_ja"`
			ActionJa            string   `json:"action_ja"`
			RecommendedActionJa string   `json:"recommended_action_ja"`
		}
		var secAlert struct {
			Timestamp   string                 `json:"timestamp"`
			RuleID      string                 `json:"rule_id"`
			RuleTitle   string                 `json:"rule_title"`
			Level       string                 `json:"level"`
			Category    string                 `json:"category"`
			SourceLog   string                 `json:"source_log"`
			MatchedData map[string]interface{} `json:"matched_data"`
			AIJudge     *aiJudgePayload        `json:"ai_judge"`
			AIJudgment  *aiJudgePayload        `json:"ai_judgment"`
		}
		if err := json.Unmarshal(mr.Payload, &secAlert); err != nil || (secAlert.RuleTitle == "" && secAlert.RuleID == "") {
			return false
		}
		tRec := parseTime(secAlert.Timestamp, tNano)
		judge := secAlert.AIJudge
		if judge == nil {
			judge = secAlert.AIJudgment
		}
		aiJudged := judge != nil
		var isThreat, compViol bool
		var compTypes []string
		var conf float64
		var sumJa, actJa string
		if judge != nil {
			isThreat = judge.IsThreat
			compViol = judge.ComplianceViolated
			compTypes = judge.ComplianceTypes
			if len(compTypes) == 0 {
				compTypes = judge.Compliance
			}
			conf = judge.Confidence
			sumJa = judge.SummaryJa
			actJa = judge.ActionJa
			if actJa == "" {
				actJa = judge.RecommendedActionJa
			}
		} else {
			l := strings.ToLower(secAlert.Level)
			isThreat = l == "critical" || l == "high"
		}
		s.upsertTwAgentAlert(info, "mqtt", secAlert.RuleID, secAlert.RuleTitle, secAlert.Level, secAlert.Category, secAlert.SourceLog,
			aiJudged, isThreat, compViol, compTypes, conf, sumJa, actJa, tRec)

		if strings.EqualFold(secAlert.Category, "Configuration Audit") || strings.EqualFold(secAlert.Category, "SCA") || strings.HasPrefix(secAlert.RuleID, "CIS-") {
			var desc, proof, remediation string
			var compliance []string
			if secAlert.MatchedData != nil {
				desc, _ = secAlert.MatchedData["description"].(string)
				proof, _ = secAlert.MatchedData["proof"].(string)
				remediation, _ = secAlert.MatchedData["remediation"].(string)
				if comps, ok := secAlert.MatchedData["compliance"].([]interface{}); ok {
					for _, c := range comps {
						if cStr, ok := c.(string); ok {
							compliance = append(compliance, cStr)
						}
					}
				}
			}
			s.upsertTwAgentSCA(info, "mqtt", secAlert.RuleID, secAlert.RuleTitle, desc, secAlert.Level,
				secAlert.SourceLog, proof, remediation, compliance, aiJudged, isThreat, compViol, conf, sumJa, actJa, tRec)
		}
		return true
	}

	return false
}

// processTwAgentSyslog processes RFC 5424 syslog messages emitted by twagentneo.
func (s *Session) processTwAgentSyslog(r Record) bool {
	cleanTag := strings.ToLower(CleanTag(r.Tag))
	isTwAgentTag := strings.HasPrefix(cleanTag, "twagentneo")
	hasSD := strings.Contains(r.Content, "twsnmpneo@17862")
	if !isTwAgentTag && !hasSD {
		return false
	}

	sd := parseTwAgentSD(r.Content)
	msgBody := stripTwAgentSD(r.Content)

	agentID := sd["agent_id"]
	hostname := r.Host
	ip := cleanHostIP(r.Remote)
	if ip == "" && net.ParseIP(r.Host) != nil {
		ip = r.Host
	}

	info := s.getOrCreateAgentInfo(agentID, hostname, ip, "syslog", r.Time)

	// 1. FIM syslog
	if cleanTag == "twagentneo-fim" || sd["op"] != "" {
		op := sd["op"]
		path := sd["path"]
		if path == "" {
			path = msgBody
		}
		s.upsertTwAgentFIM(info, "syslog", op, path, "", r.Time)
		return true
	}

	// 2. Resource syslog
	if cleanTag == "twagentneo-resource" || sd["resource"] != "" {
		resName := sd["resource"]
		title := msgBody
		if title == "" {
			title = fmt.Sprintf("Resource %s exceeded (%s%% > %s%%)", resName, sd["val"], sd["thresh"])
		}
		s.upsertTwAgentAlert(info, "syslog", "resource-"+resName, title, "warn", "Resource", msgBody,
			false, false, false, nil, 0, "", "", r.Time)
		return true
	}

	// 3. Security alert syslog
	ruleTitle := sd["rule"]
	if ruleTitle == "" {
		ruleTitle = msgBody
	}
	threatStr := strings.ToUpper(sd["threat"])
	compStr := strings.ToUpper(sd["comp"])
	aiJudged := threatStr != "" || compStr != ""
	isThreat := threatStr == "YES"
	compViol := compStr == "YES"

	cat := "Security"
	sourceLog := msgBody
	if idx := strings.Index(msgBody, ": "); idx > 0 && idx < 32 {
		cat = strings.TrimSpace(msgBody[:idx])
		sourceLog = strings.TrimSpace(msgBody[idx+2:])
	}

	level := syslogSeverityToLevel(r.Severity)
	if !aiJudged && (level == "critical" || level == "high") {
		isThreat = true
	}

	s.upsertTwAgentAlert(info, "syslog", "", ruleTitle, level, cat, sourceLog,
		aiJudged, isThreat, compViol, nil, 0, "", "", r.Time)

	// If SCA finding
	if cleanTag == "twagentneo-sca" || sd["sca"] != "" || strings.EqualFold(cat, "Configuration Audit") || strings.EqualFold(cat, "SCA") || strings.HasPrefix(ruleTitle, "CIS-") {
		ruleID := sd["sca"]
		if ruleID == "" && strings.HasPrefix(ruleTitle, "CIS-") {
			parts := strings.SplitN(ruleTitle, ":", 2)
			ruleID = strings.TrimSpace(parts[0])
		}
		s.upsertTwAgentSCA(info, "syslog", ruleID, ruleTitle, "", level, sourceLog, "", "", nil,
			aiJudged, isThreat, compViol, 0, "", "", r.Time)
	}
	return true
}

// processTwAgentTrap processes SNMP traps matching TWAGENT-NEO-MIB (.1.3.6.1.4.1.17862.3).
func (s *Session) processTwAgentTrap(tr *TrapRecord) bool {
	if tr == nil {
		return false
	}
	ttLower := strings.ToLower(tr.TrapType)
	entLower := strings.ToLower(tr.Enterprise)
	if !strings.Contains(ttLower, "twagentneo") &&
		!strings.Contains(ttLower, "17862.3") &&
		!strings.Contains(entLower, "twagentneo") &&
		!strings.Contains(tr.Variables, "twagentNeo") {
		return false
	}

	vars := parseTwAgentTrapVars(tr.Variables)
	hostname := vars["twagentNeoHostname"]
	version := vars["twagentNeoVersion"]
	osName := vars["twagentNeoOS"]
	sevStr := vars["twagentNeoTrapSeverity"]
	cat := vars["twagentNeoTrapCategory"]
	ruleTitle := vars["twagentNeoTrapRuleTitle"]
	aiJudge := vars["twagentNeoTrapAIJudge"]
	aiSummary := vars["twagentNeoTrapAISummary"]
	aiAction := vars["twagentNeoTrapAIAction"]
	sourceLog := vars["twagentNeoTrapSourceLog"]
	if sourceLog == "" {
		sourceLog = tr.Variables
	}

	ip := cleanHostIP(tr.FromAddress)
	agentID := hostname
	if agentID == "" {
		agentID = ip
	}

	tNano := tr.Time
	if tNano <= 0 {
		tNano = time.Now().UnixNano()
	}

	info := s.getOrCreateAgentInfo(agentID, hostname, ip, "trap", tNano)
	if version != "" {
		info.Version = version
	}
	if osName != "" {
		info.OS = osName
	}
	info.Status = "online"
	info.Count++
	if tNano > info.LastTime {
		info.LastTime = tNano
	}
	putEnt(s, KindTwAgentInfo, info.ID, info)

	// Heartbeat trap
	if strings.Contains(ttLower, "heartbeat") || strings.HasSuffix(ttLower, "17862.3.2.0.4") || strings.EqualFold(cat, "Heartbeat") {
		s.Processed++
		return true
	}

	// FIM trap
	if strings.Contains(ttLower, "fim") || strings.HasSuffix(ttLower, "17862.3.2.0.3") || strings.EqualFold(cat, "FIM") {
		op := "MODIFY"
		path := ruleTitle
		// Format from twagentneo: "File <OP>: <PATH>"
		if strings.HasPrefix(ruleTitle, "File ") {
			rest := strings.TrimPrefix(ruleTitle, "File ")
			if parts := strings.SplitN(rest, ":", 2); len(parts) == 2 {
				op = strings.TrimSpace(parts[0])
				path = strings.TrimSpace(parts[1])
			}
		}
		s.upsertTwAgentFIM(info, "trap", op, path, "", tNano)
		return true
	}

	// Resource or Security Alert trap
	level := trapSeverityToLevel(sevStr)
	if cat == "" {
		if strings.Contains(ttLower, "resourcealert") || strings.HasSuffix(ttLower, "17862.3.2.0.1") {
			cat = "Resource"
		} else {
			cat = "Security"
		}
	}
	if ruleTitle == "" {
		ruleTitle = tr.TrapType
	}

	aiJudged := aiJudge != ""
	isThreat := strings.Contains(aiJudge, "Threat:YES")
	compViol := strings.Contains(aiJudge, "CompViolation:YES")
	var conf float64
	for _, part := range strings.Split(aiJudge, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Conf:") {
			conf, _ = strconv.ParseFloat(strings.TrimPrefix(part, "Conf:"), 64)
		}
	}
	if !aiJudged && (level == "critical" || level == "high") {
		isThreat = true
	}

	s.upsertTwAgentAlert(info, "trap", "", ruleTitle, level, cat, sourceLog,
		aiJudged, isThreat, compViol, nil, conf, aiSummary, aiAction, tNano)

	// If SCA trap
	if strings.EqualFold(cat, "Configuration Audit") || strings.EqualFold(cat, "SCA") || strings.Contains(ttLower, "twagentneosca") || strings.HasPrefix(ruleTitle, "CIS-") {
		ruleID := ""
		if strings.HasPrefix(ruleTitle, "CIS-") {
			parts := strings.SplitN(ruleTitle, ":", 2)
			ruleID = strings.TrimSpace(parts[0])
		}
		s.upsertTwAgentSCA(info, "trap", ruleID, ruleTitle, "", level, sourceLog, "", "", nil,
			aiJudged, isThreat, compViol, conf, aiSummary, aiAction, tNano)
	}
	return true
}

func isLoopbackIP(ip string) bool {
	ip = strings.TrimSpace(ip)
	return ip == "" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
}

func (s *Session) getOrCreateAgentInfo(agentID, hostname, ip, proto string, tNano int64) *TwAgentInfoEnt {
	cleanAgentID := strings.TrimSpace(agentID)
	cleanHostname := strings.TrimSpace(hostname)
	cleanIP := strings.TrimSpace(ip)
	// If cleanIP is actually a hostname (not valid IP), move it to cleanHostname if empty
	if cleanIP != "" && net.ParseIP(cleanIP) == nil {
		if cleanHostname == "" {
			cleanHostname = cleanIP
		}
		cleanIP = ""
	}

	key := strings.ToLower(cleanAgentID)
	if key == "" {
		key = strings.ToLower(cleanHostname)
	}
	if key == "" {
		key = cleanIP
	}
	if key == "" {
		key = "unknown"
	}

	id := makeID("twagent:" + key)
	info := getEnt[TwAgentInfoEnt](s, KindTwAgentInfo, id)

	if info == nil {
		for _, existing := range allEnts[TwAgentInfoEnt](s, KindTwAgentInfo) {
			// Match by AgentID
			if cleanAgentID != "" && existing.AgentID != "" && strings.EqualFold(existing.AgentID, cleanAgentID) {
				info = existing
				break
			}
			// Match by IP
			if cleanIP != "" && existing.IP != "" && !isLoopbackIP(cleanIP) && !isLoopbackIP(existing.IP) && existing.IP == cleanIP {
				info = existing
				break
			}
			// Match by Hostname
			if cleanHostname != "" && existing.Hostname != "" && strings.EqualFold(existing.Hostname, cleanHostname) {
				info = existing
				break
			}
			// Match if existing.IP was previously set to hostname
			if cleanHostname != "" && existing.IP != "" && strings.EqualFold(existing.IP, cleanHostname) {
				info = existing
				break
			}
			// Match if node matching
			if s.store != nil && cleanIP != "" {
				if n := s.lookupNode(cleanIP); n.ID != "" && (strings.EqualFold(n.Name, existing.Hostname) || strings.EqualFold(n.Name, existing.Host)) {
					info = existing
					break
				}
			}
		}
	}

	if info == nil {
		info = &TwAgentInfoEnt{
			ID:        id,
			AgentID:   cleanAgentID,
			Host:      cleanHostname,
			Hostname:  cleanHostname,
			IP:        cleanIP,
			Status:    "online",
			Protocols: []string{proto},
			FirstTime: tNano,
			LastTime:  tNano,
		}
	} else {
		if cleanAgentID != "" && (info.AgentID == "" || info.AgentID == info.IP) {
			info.AgentID = cleanAgentID
		}
		if cleanHostname != "" && (info.Hostname == "" || info.Hostname == info.IP) {
			info.Hostname = cleanHostname
			info.Host = cleanHostname
		}
		if cleanIP != "" && !isLoopbackIP(cleanIP) && (info.IP == "" || isLoopbackIP(info.IP) || net.ParseIP(info.IP) == nil) {
			info.IP = cleanIP
		}
		info.Protocols = mergeStringSlice(info.Protocols, []string{proto})
		if tNano > info.LastTime {
			info.LastTime = tNano
		}
	}

	if info.Host == "" && info.Hostname != "" {
		info.Host = info.Hostname
	}
	if info.Hostname == "" && info.Host != "" {
		info.Hostname = info.Host
	}

	// Deduplicate any duplicate agent entities in store (e.g. from prior IP/hostname confusion)
	for _, other := range allEnts[TwAgentInfoEnt](s, KindTwAgentInfo) {
		if other.ID == info.ID {
			continue
		}
		isDup := false
		if info.IP != "" && other.IP == info.IP && !isLoopbackIP(info.IP) {
			isDup = true
		} else if info.Hostname != "" && (strings.EqualFold(other.Hostname, info.Hostname) || strings.EqualFold(other.IP, info.Hostname)) {
			isDup = true
		} else if info.AgentID != "" && strings.EqualFold(other.AgentID, info.AgentID) {
			isDup = true
		}

		if isDup {
			if (info.Hostname == "" || info.Hostname == info.IP) && other.Hostname != "" {
				info.Hostname = other.Hostname
				info.Host = other.Hostname
			}
			if net.ParseIP(info.IP) == nil && net.ParseIP(other.IP) != nil {
				info.IP = other.IP
			}
			if info.AgentID == "" && other.AgentID != "" {
				info.AgentID = other.AgentID
			}
			if info.MAC == "" && other.MAC != "" {
				info.MAC = other.MAC
			}
			if other.FirstTime > 0 && (info.FirstTime == 0 || other.FirstTime < info.FirstTime) {
				info.FirstTime = other.FirstTime
			}
			if other.LastTime > info.LastTime {
				info.LastTime = other.LastTime
			}
			info.Protocols = mergeStringSlice(info.Protocols, other.Protocols)
			// Remove duplicate from session cache and persistent store
			if s.cache[KindTwAgentInfo] != nil {
				delete(s.cache[KindTwAgentInfo], other.ID)
			}
			if s.dirty[KindTwAgentInfo] != nil {
				delete(s.dirty[KindTwAgentInfo], other.ID)
			}
			if s.store != nil {
				_ = s.store.DeleteLogReportData(s.ctx, KindTwAgentInfo, []string{other.ID})
			}
		}
	}

	s.enrichAgentNode(info)
	putEnt(s, KindTwAgentInfo, info.ID, info)
	return info
}

func (s *Session) enrichAgentNode(info *TwAgentInfoEnt) {
	if info == nil {
		return
	}
	if info.Host == "" && info.Hostname != "" {
		info.Host = info.Hostname
	}
	if info.Hostname == "" && info.Host != "" {
		info.Hostname = info.Host
	}
	if info.IP != "" && net.ParseIP(info.IP) != nil {
		n := s.lookupNode(info.IP)
		if n.ID != "" {
			info.NodeID = n.ID
			info.NodeName = n.Name
			if info.Hostname == "" {
				info.Hostname = n.Name
				info.Host = n.Name
			}
			return
		}
	}
	// Fallback: match node by name == Hostname or AgentID
	if s.store != nil && info.NodeID == "" && (info.Hostname != "" || info.AgentID != "") {
		if nodes, err := s.store.ListNodes(s.ctx); err == nil {
			for _, n := range nodes {
				if (info.Hostname != "" && strings.EqualFold(n.Name, info.Hostname)) ||
					(info.AgentID != "" && strings.EqualFold(n.Name, info.AgentID)) {
					info.NodeID = n.ID
					info.NodeName = n.Name
					if info.IP == "" && n.IP != "" {
						info.IP = n.IP
					}
					if info.MAC == "" && n.MAC != "" {
						info.MAC = n.MAC
					}
					break
				}
			}
		}
	}
}

func (s *Session) upsertTwAgentFIM(info *TwAgentInfoEnt, proto, op, path, hash string, tNano int64) {
	if op == "" {
		op = "MODIFY"
	}
	op = strings.ToUpper(op)
	proto = strings.ToLower(strings.TrimSpace(proto))
	if proto == "" {
		proto = "mqtt"
	}
	hostLabel := info.Hostname
	if hostLabel == "" {
		hostLabel = info.AgentID
	}
	if hostLabel == "" {
		hostLabel = info.IP
	}

	id := makeID(fmt.Sprintf("%s:%s:%s", hostLabel, op, path))
	ent := getEnt[TwAgentFIMEnt](s, KindTwAgentFIM, id)
	if ent == nil {
		for _, existing := range allEnts[TwAgentFIMEnt](s, KindTwAgentFIM) {
			if existing.Op == op && existing.Path == path {
				if (info.AgentID != "" && existing.AgentID == info.AgentID) ||
					(hostLabel != "" && strings.EqualFold(existing.Host, hostLabel)) ||
					(info.Hostname != "" && strings.EqualFold(existing.Host, info.Hostname)) ||
					(info.IP != "" && existing.IP == info.IP) {
					ent = existing
					id = existing.ID
					break
				}
			}
		}
	}

	isRecent := ent != nil && (tNano-ent.LastTime) < int64(3*time.Second) && (tNano-ent.LastTime) > -int64(3*time.Second)

	if ent == nil {
		ent = &TwAgentFIMEnt{
			ID:        id,
			AgentID:   info.AgentID,
			Host:      hostLabel,
			IP:        info.IP,
			NodeID:    info.NodeID,
			NodeName:  info.NodeName,
			Protocol:  proto,
			Protocols: []string{proto},
			Op:        op,
			Path:      path,
			Hash:      hash,
			Count:     1,
			FirstTime: tNano,
			LastTime:  tNano,
		}
		info.FIMCount++
	} else {
		// Check if this exact protocol was already received recently (e.g. MQTT dual "fim" and "alert" publish)
		alreadyHasProto := false
		for _, p := range ent.Protocols {
			if strings.EqualFold(p, proto) {
				alreadyHasProto = true
				break
			}
		}

		if isRecent && alreadyHasProto && (hash == "" || ent.Hash == hash) {
			// Suppress duplicate within same protocol
			return
		}

		ent.Protocols = mergeStringSlice(ent.Protocols, []string{proto})
		ent.Protocol = strings.Join(ent.Protocols, ", ")

		if hash != "" {
			ent.Hash = hash
		}
		if info.IP != "" {
			ent.IP = info.IP
		}
		if info.NodeID != "" {
			ent.NodeID = info.NodeID
			ent.NodeName = info.NodeName
		}

		// If this is a concurrent multi-protocol delivery (e.g. MQTT + Syslog + Trap within 3s),
		// do not inflate the modification event count, but ensure protocol and timestamp are merged.
		if !isRecent {
			ent.Count++
			info.FIMCount++
		}
		if tNano > ent.LastTime {
			ent.LastTime = tNano
		}
	}
	putEnt(s, KindTwAgentInfo, info.ID, info)
	putEnt(s, KindTwAgentFIM, id, ent)
	s.Processed++
}

func (s *Session) upsertTwAgentAlert(
	info *TwAgentInfoEnt,
	proto, ruleID, ruleTitle, level, category, sourceLog string,
	aiJudged, isThreat, compViol bool,
	compTypes []string,
	confidence float64,
	summaryJa, actionJa string,
	tNano int64,
) {
	proto = strings.ToLower(strings.TrimSpace(proto))
	if proto == "" {
		proto = "mqtt"
	}
	hostLabel := info.Hostname
	if hostLabel == "" {
		hostLabel = info.AgentID
	}
	if hostLabel == "" {
		hostLabel = info.IP
	}
	if level == "" {
		level = "info"
	} else {
		level = strings.ToLower(level)
	}
	if category == "" {
		category = "Security"
	}

	id := makeID(fmt.Sprintf("%s:%s:%s", hostLabel, category, ruleTitle))
	ent := getEnt[TwAgentAlertEnt](s, KindTwAgentAlert, id)
	if ent == nil {
		ent = &TwAgentAlertEnt{
			ID:                 id,
			AgentID:            info.AgentID,
			Host:               hostLabel,
			IP:                 info.IP,
			NodeID:             info.NodeID,
			NodeName:           info.NodeName,
			Protocol:           proto,
			Protocols:          []string{proto},
			RuleID:             ruleID,
			RuleTitle:          ruleTitle,
			Level:              level,
			Category:           category,
			SourceLog:          sourceLog,
			AIJudged:           aiJudged,
			IsThreat:           isThreat,
			ComplianceViolated: compViol,
			ComplianceTypes:    compTypes,
			Confidence:         confidence,
			SummaryJa:          summaryJa,
			ActionJa:           actionJa,
			Count:              1,
			FirstTime:          tNano,
			LastTime:           tNano,
		}
	} else {
		ent.Protocols = mergeStringSlice(ent.Protocols, []string{proto})
		ent.Protocol = strings.Join(ent.Protocols, ", ")
		if ruleID != "" {
			ent.RuleID = ruleID
		}
		ent.Level = level
		if sourceLog != "" {
			ent.SourceLog = sourceLog
		}
		if aiJudged {
			ent.AIJudged = true
			ent.IsThreat = isThreat
			ent.ComplianceViolated = compViol
		}
		if len(compTypes) > 0 {
			ent.ComplianceTypes = mergeStringSlice(ent.ComplianceTypes, compTypes)
		}
		if confidence > 0 {
			ent.Confidence = confidence
		}
		if summaryJa != "" {
			ent.SummaryJa = summaryJa
		}
		if actionJa != "" {
			ent.ActionJa = actionJa
		}
		if info.IP != "" {
			ent.IP = info.IP
		}
		if info.NodeID != "" {
			ent.NodeID = info.NodeID
			ent.NodeName = info.NodeName
		}
		ent.Count++
		if tNano > ent.LastTime {
			ent.LastTime = tNano
		}
	}
	info.AlertCount++
	if isThreat {
		info.ThreatCount++
	}
	putEnt(s, KindTwAgentInfo, info.ID, info)
	putEnt(s, KindTwAgentAlert, id, ent)
	s.Processed++
}

func (s *Session) upsertTwAgentSCA(
	info *TwAgentInfoEnt,
	proto, ruleID, title, description, severity, filePath, proof, remediation string,
	compliance []string,
	aiJudged, isThreat, compViol bool,
	confidence float64,
	summaryJa, actionJa string,
	tNano int64,
) {
	proto = strings.ToLower(strings.TrimSpace(proto))
	if proto == "" {
		proto = "mqtt"
	}
	hostLabel := info.Hostname
	if hostLabel == "" {
		hostLabel = info.AgentID
	}
	if hostLabel == "" {
		hostLabel = info.IP
	}
	if severity == "" {
		severity = "info"
	} else {
		severity = strings.ToLower(severity)
	}

	keyPart := ruleID
	if keyPart == "" {
		keyPart = title
	}
	id := makeID(fmt.Sprintf("%s:%s:%s", hostLabel, keyPart, filePath))
	ent := getEnt[TwAgentSCAEnt](s, KindTwAgentSCA, id)
	if ent == nil {
		for _, existing := range allEnts[TwAgentSCAEnt](s, KindTwAgentSCA) {
			if existing.FilePath == filePath && (existing.RuleID == ruleID || existing.Title == title) {
				if (info.AgentID != "" && existing.AgentID == info.AgentID) ||
					(hostLabel != "" && strings.EqualFold(existing.Host, hostLabel)) ||
					(info.Hostname != "" && strings.EqualFold(existing.Host, info.Hostname)) ||
					(info.IP != "" && existing.IP == info.IP) {
					ent = existing
					id = existing.ID
					break
				}
			}
		}
	}

	isRecent := ent != nil && (tNano-ent.LastTime) < int64(3*time.Second) && (tNano-ent.LastTime) > -int64(3*time.Second)

	if ent == nil {
		ent = &TwAgentSCAEnt{
			ID:                 id,
			AgentID:            info.AgentID,
			Host:               hostLabel,
			IP:                 info.IP,
			NodeID:             info.NodeID,
			NodeName:           info.NodeName,
			Protocol:           proto,
			Protocols:          []string{proto},
			RuleID:             ruleID,
			Title:              title,
			Description:        description,
			Severity:           severity,
			FilePath:           filePath,
			Proof:              proof,
			Remediation:        remediation,
			Compliance:         compliance,
			AIJudged:           aiJudged,
			IsThreat:           isThreat,
			ComplianceViolated: compViol,
			Confidence:         confidence,
			SummaryJa:          summaryJa,
			ActionJa:           actionJa,
			Count:              1,
			FirstTime:          tNano,
			LastTime:           tNano,
		}
		info.SCACount++
	} else {
		ent.Protocols = mergeStringSlice(ent.Protocols, []string{proto})
		ent.Protocol = strings.Join(ent.Protocols, ", ")
		if ruleID != "" {
			ent.RuleID = ruleID
		}
		if title != "" {
			ent.Title = title
		}
		if description != "" {
			ent.Description = description
		}
		if severity != "" {
			ent.Severity = severity
		}
		if proof != "" {
			ent.Proof = proof
		}
		if remediation != "" {
			ent.Remediation = remediation
		}
		if len(compliance) > 0 {
			ent.Compliance = mergeStringSlice(ent.Compliance, compliance)
		}
		if aiJudged {
			ent.AIJudged = true
			ent.IsThreat = isThreat
			ent.ComplianceViolated = compViol
		}
		if confidence > 0 {
			ent.Confidence = confidence
		}
		if summaryJa != "" {
			ent.SummaryJa = summaryJa
		}
		if actionJa != "" {
			ent.ActionJa = actionJa
		}
		if info.IP != "" {
			ent.IP = info.IP
		}
		if info.NodeID != "" {
			ent.NodeID = info.NodeID
			ent.NodeName = info.NodeName
		}
		if !isRecent {
			ent.Count++
			info.SCACount++
		}
		if tNano > ent.LastTime {
			ent.LastTime = tNano
		}
	}
	putEnt(s, KindTwAgentInfo, info.ID, info)
	putEnt(s, KindTwAgentSCA, id, ent)
	s.Processed++
}

func parseTwAgentTopic(topic, clientID string) (agentID, subTopic string) {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	if len(parts) < 2 {
		return "", ""
	}
	last := strings.ToLower(parts[len(parts)-1])
	switch last {
	case "metrics", "resource", "interface", "network", "arp", "tcp", "alert", "fim", "sca", "status", "heartbeat":
		subTopic = last
		agentID = parts[len(parts)-2]
		if strings.EqualFold(agentID, "agent") && clientID != "" {
			agentID = clientID
		}
		return agentID, subTopic
	}
	return "", ""
}

func cleanHostIP(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	// Strip "(nodeName)" suffix from Trap FromAddress like "192.168.1.10(host1)"
	if idx := strings.Index(addr, "("); idx > 0 {
		addr = strings.TrimSpace(addr[:idx])
	}
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func isLoopbackName(name string) bool {
	l := strings.ToLower(strings.TrimSpace(name))
	return l == "lo" || strings.HasPrefix(l, "lo0") || strings.Contains(l, "loopback")
}

func mergeStringSlice(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	var out []string
	for _, s := range a {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	for _, s := range b {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func calcTwAgentTCPPenalty(ent *TwAgentTCPEnt) int {
	penalty := 0
	if !isSafeCountry(ent.RemoteLoc) {
		penalty += 2
	}
	for _, svc := range ent.Services {
		if strings.HasPrefix(svc, "port/") || svc == "telnet" || svc == "ftp" {
			penalty++
		}
	}
	return penalty
}

func parseTwAgentSD(s string) map[string]string {
	out := make(map[string]string)
	idx := strings.Index(s, "[twsnmpneo@17862")
	if idx < 0 {
		return out
	}
	sub := s[idx:]
	matches := twAgentSDRegex.FindAllStringSubmatch(sub, -1)
	for _, m := range matches {
		if len(m) == 3 {
			val := strings.ReplaceAll(m[2], `\"`, `"`)
			val = strings.ReplaceAll(val, `\\`, `\`)
			out[m[1]] = val
		}
	}
	return out
}

func stripTwAgentSD(s string) string {
	idx := strings.Index(s, "[twsnmpneo@17862")
	if idx < 0 {
		return strings.TrimSpace(s)
	}
	end := strings.Index(s[idx:], "]")
	if end < 0 {
		return strings.TrimSpace(s)
	}
	rest := strings.TrimSpace(s[:idx] + " " + s[idx+end+1:])
	return rest
}

func parseTwAgentTrapVars(vars string) map[string]string {
	res := make(map[string]string)
	vars = strings.TrimSpace(vars)
	if vars == "" {
		return res
	}
	matches := twAgentTrapVarRegex.FindAllStringSubmatchIndex(vars, -1)
	for i, m := range matches {
		key := vars[m[2]:m[3]]
		valStart := m[1]
		valEnd := len(vars)
		if i+1 < len(matches) {
			valEnd = matches[i+1][0]
		}
		val := strings.TrimRight(strings.TrimSpace(vars[valStart:valEnd]), ",")
		res[key] = val
		baseKey := strings.TrimSuffix(key, ".0")
		if baseKey != key {
			res[baseKey] = val
		}
	}
	return res
}

func syslogSeverityToLevel(sev int) string {
	switch {
	case sev <= 2:
		return "critical"
	case sev == 3:
		return "high"
	case sev == 4:
		return "warn"
	case sev == 5:
		return "low"
	default:
		return "info"
	}
}

func trapSeverityToLevel(sev string) string {
	s := strings.ToLower(strings.TrimSpace(sev))
	switch {
	case strings.Contains(s, "critical") || strings.HasPrefix(s, "4"):
		return "critical"
	case strings.Contains(s, "high") || strings.HasPrefix(s, "3"):
		return "high"
	case strings.Contains(s, "warn") || strings.HasPrefix(s, "2"):
		return "warn"
	default:
		return "info"
	}
}
