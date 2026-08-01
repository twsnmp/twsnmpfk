package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

type LLMResp struct {
	Results string `json:"Results"`
	Error   string `json:"Error"`
}

type LLMMIBSearchResp struct {
	ObjectName string `json:"ObjectName"`
	OID        string `json:"OID"`
	Error      string `json:"Error"`
}

func (a *App) LLMMIBSearch(prompt string) *LLMMIBSearchResp {
	r := new(LLMMIBSearchResp)
	ctx := a.ctx
	llm, err := datastore.GetLLM(ctx)
	if err != nil {
		log.Printf("LLMMIBSearch err=%v", err)
		r.Error = err.Error()
		return r
	}
	system := `You are an expert on SNMP MIBs. Fulfill the requests entered by the user. Please provide the object name and OID of the SNMP MIB.
Please be sure to answer only the object name and OID in the following format. No need for extra explanation.
Object name,OID`
	if i18n.GetLang() == "ja" {
		system = `あなたはSNMPのMIBに関する専門家です。ユーザーの入力した要望を満たす。SNMPのMIBのオブジェクト名とOIDを答えてください。
必ずオブジェクト名とOIDのみを以下の形式で回答してください。余計な解説は不要です。
オブジェクト名,OID
`
	}
	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}
	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		log.Printf("LLMMIBSearch err=%v", err)
		r.Error = err.Error()
		return r
	}
	if len(resp.Choices) < 1 {
		r.Error = "no response from LLM"
		return r
	}
	res := strings.TrimSpace(resp.Choices[0].Content)
	res = strings.TrimPrefix(res, "```")
	res = strings.TrimSuffix(res, "```")
	res = strings.TrimSpace(res)
	if a := strings.SplitN(res, ",", 2); len(a) == 2 {
		r.ObjectName = strings.TrimSpace(a[0])
		r.OID = strings.TrimSpace(a[1])
	} else {
		r.Error = resp.Choices[0].Content
	}
	return r
}

func (a *App) LLMAskMIB(prompt string) *LLMResp {
	system := `You are an expert on SNMP MIBs.
Please explain the SNMP acquisition results entered by the user.`
	if i18n.GetLang() == "ja" {
		system = `あなたはSNMPのMIBに関する専門家です。
ユーザーの入力したSNMPの取得結果について解説してください。`
	}
	return a.llmAsk(prompt, system)
}

func (a *App) LLMAskLog(prompt string) *LLMResp {
	system := `You are an expert in log analysis.
Please explain the log input by the user.`
	if i18n.GetLang() == "ja" {
		system = `あなたはログ分析に関する専門家です。
ユーザーの入力したログについて解説してください。`
	}
	return a.llmAsk(prompt, system)
}

func formatPollingResultForLLM(res map[string]interface{}) string {
	if len(res) == 0 {
		return "{}"
	}
	formatted := make([]string, 0, len(res))
	for k, v := range res {
		vf, isNum := toFloat64(v)
		if isNum {
			switch k {
			case "rtt", "response_time":
				ms := vf / 1000000.0
				formatted = append(formatted, fmt.Sprintf("%s: %.2f ms (%v ns)", k, ms, v))
			case "speed":
				formatted = append(formatted, fmt.Sprintf("%s: %.2f Mbps", k, vf))
			case "days":
				formatted = append(formatted, fmt.Sprintf("%s: %.0f days", k, vf))
			default:
				formatted = append(formatted, fmt.Sprintf("%s: %v", k, v))
			}
		} else {
			formatted = append(formatted, fmt.Sprintf("%s: %v", k, v))
		}
	}
	return "{" + strings.Join(formatted, ", ") + "}"
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int64:
		return float64(val), true
	case int:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}

func (a *App) LLMDiagnoseNode(nodeID string) *LLMResp {
	n := datastore.GetNode(nodeID)
	if n == nil {
		return &LLMResp{Error: "node not found"}
	}

	var sb strings.Builder

	// 1. ノード情報（パスワードなどの認証情報は除外）
	sb.WriteString("# Node Information\n")
	sb.WriteString(fmt.Sprintf("- Name: %s\n", n.Name))
	sb.WriteString(fmt.Sprintf("- State: %s\n", n.State))
	sb.WriteString(fmt.Sprintf("- IP Address: %s\n", n.IP))
	sb.WriteString(fmt.Sprintf("- MAC Address: %s\n", n.MAC))
	sb.WriteString(fmt.Sprintf("- Vendor: %s\n", n.Vendor))
	if n.Descr != "" {
		sb.WriteString(fmt.Sprintf("- Description: %s\n", n.Descr))
	}
	if n.Loc != "" {
		sb.WriteString(fmt.Sprintf("- Location: %s\n", n.Loc))
	}
	if n.URL != "" {
		sb.WriteString(fmt.Sprintf("- URL: %s\n", n.URL))
	}
	sb.WriteString(fmt.Sprintf("- SNMP Mode: %s\n", n.SnmpMode))
	if n.SnmpPort > 0 {
		sb.WriteString(fmt.Sprintf("- SNMP Port: %d\n", n.SnmpPort))
	}
	if n.AddrMode != "" {
		sb.WriteString(fmt.Sprintf("- Addr Mode: %s\n", n.AddrMode))
	}

	// 2. ポーリング情報および直近のポーリングログ
	sb.WriteString("\n# Polling Information & Recent Logs\n")
	pollings := []datastore.PollingEnt{}
	datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
		if p.NodeID == nodeID {
			pollings = append(pollings, *p)
		}
		return true
	})

	if len(pollings) == 0 {
		sb.WriteString("No pollings configured for this node.\n")
	} else {
		for _, p := range pollings {
			sb.WriteString(fmt.Sprintf("## Polling: %s (Type: %s, Mode: %s, State: %s)\n", p.Name, p.Type, p.Mode, p.State))
			sb.WriteString(fmt.Sprintf("- Poll Interval: %d sec, Level: %s\n", p.PollInt, p.Level))
			if p.Params != "" {
				sb.WriteString(fmt.Sprintf("- Params: %s\n", p.Params))
			}
			if p.Filter != "" {
				sb.WriteString(fmt.Sprintf("- Filter: %s\n", p.Filter))
			}
			sb.WriteString("- Recent Polling Logs:\n")
			logs := datastore.GetAllPollingLog(p.ID)
			if len(logs) == 0 {
				sb.WriteString("  - No polling log available\n")
			} else {
				count := 0
				for i := len(logs) - 1; i >= 0 && count < 5; i-- {
					l := logs[i]
					tStr := time.Unix(0, l.Time).Format(time.RFC3339)
					sb.WriteString(fmt.Sprintf("  - %s | State: %s | Result: %s\n", tStr, l.State, formatPollingResultForLLM(l.Result)))
					count++
				}
			}
		}
	}

	// 3. ノードに関連した syslog, snmptrap, arp ログ
	sb.WriteString("\n# Related Logs\n")

	// 3a. Syslog (過去24時間、最大30件)
	sb.WriteString("## Recent Syslog\n")
	st24 := time.Now().Add(-24 * time.Hour).UnixNano()
	etNow := time.Now().UnixNano()
	syslogCount := 0
	datastore.ForEachSyslog(st24, etNow, func(l *datastore.SyslogEnt) bool {
		if (n.IP != "" && strings.Contains(l.Host, n.IP)) || (n.Name != "" && strings.Contains(l.Host, n.Name)) {
			tStr := time.Unix(0, l.Time).Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("- %s [Facility: %d, Severity: %d, Tag: %s] Host: %s, Message: %s\n",
				tStr, l.Facility, l.Severity, l.Tag, l.Host, l.Message))
			syslogCount++
			if syslogCount >= 30 {
				return false
			}
		}
		return true
	})
	if syslogCount == 0 {
		sb.WriteString("No recent Syslog entries found.\n")
	}

	// 3b. SNMP Trap (過去24時間、最大30件)
	sb.WriteString("\n## Recent SNMP Traps\n")
	trapCount := 0
	datastore.ForEachTraps(st24, etNow, func(l *datastore.TrapEnt) bool {
		if n.IP != "" && strings.Contains(l.FromAddress, n.IP) {
			tStr := time.Unix(0, l.Time).Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("- %s [From: %s, Type: %s] Variables: %v\n",
				tStr, l.FromAddress, l.TrapType, l.Variables))
			trapCount++
			if trapCount >= 30 {
				return false
			}
		}
		return true
	})
	if trapCount == 0 {
		sb.WriteString("No recent SNMP Traps found.\n")
	}

	// 3c. ARP Logs (最大30件)
	sb.WriteString("\n## Recent ARP Logs\n")
	arpCount := 0
	datastore.ForEachLastArpLogs(func(l *datastore.ArpLogEnt) bool {
		if (n.IP != "" && l.IP == n.IP) || (n.MAC != "" && (l.NewMAC == n.MAC || l.OldMAC == n.MAC)) {
			tStr := time.Unix(0, l.Time).Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("- %s [State: %s] IP: %s, NewMAC: %s, OldMAC: %s\n",
				tStr, l.State, l.IP, l.NewMAC, l.OldMAC))
			arpCount++
			if arpCount >= 30 {
				return false
			}
		}
		return true
	})
	if arpCount == 0 {
		sb.WriteString("No recent ARP logs found.\n")
	}

	system := `You are an expert in network and system management.
Please analyze the provided node's information, polling statuses, polling logs (where 'rtt' and 'response_time' are converted to ms), and related logs.
Diagnose the overall status, highlight any anomalies or risks, and provide recommended troubleshooting or optimization actions.`
	if i18n.GetLang() == "ja" {
		system = `あなたはネットワークおよびシステム管理の専門家です。
提示されたノードの基本情報、ポーリング状態・ログ（※rtt/response_timeはミリ秒[ms]単位を付記）、および関連ログ（Syslog, SNMP Trap, ARP）を分析し、
1. ノードの現在の状態・健全性の要約
2. 検出された問題点・異常・リスクの評価
3. 推奨される対策・具体的なアクション
を日本語で分かりやすく診断・回答してください。`
	}

	return a.llmAsk(sb.String(), system)
}

func (a *App) LLMAnalyzeEventLog(l datastore.EventLogEnt) *LLMResp {
	var sb strings.Builder

	// (1) イベントログの内容
	sb.WriteString("# Target Event Log\n")
	sb.WriteString(fmt.Sprintf("- Time: %s\n", time.Unix(0, l.Time).Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("- Level: %s\n", l.Level))
	if l.LastLevel != "" {
		sb.WriteString(fmt.Sprintf("- Last Level: %s\n", l.LastLevel))
	}
	sb.WriteString(fmt.Sprintf("- Type: %s\n", l.Type))
	if l.NodeName != "" {
		sb.WriteString(fmt.Sprintf("- Node Name: %s\n", l.NodeName))
	}
	if l.NodeID != "" {
		sb.WriteString(fmt.Sprintf("- Node ID: %s\n", l.NodeID))
	}
	sb.WriteString(fmt.Sprintf("- Event: %s\n", l.Event))

	var system string
	lang := i18n.GetLang()

	switch l.Type {
	case "user":
		// ユーザー操作イベント
		if lang == "ja" {
			system = `あなたはシステム運用およびセキュリティ管理の専門家です。
ユーザーによって実行された操作・設定変更イベントログを分析し、
1. 実行された操作の目的・内容の要約
2. システムや設定、対象ノードへの影響評価
3. 運用・セキュリティ上の確認事項や注意点
を日本語で分かりやすく解説・回答してください。`
		} else {
			system = `You are an expert in system operations and security management.
Please analyze the user operation / configuration change event log provided and explain:
1. Summary of the user action and its intent
2. Evaluation of impact on the system, settings, or target components
3. Operational or security precautions to verify`
		}

		// ノード特定（もしノードに関連するユーザー操作の場合）
		var n *datastore.NodeEnt
		if l.NodeID != "" {
			n = datastore.GetNode(l.NodeID)
		}
		if n == nil && l.NodeName != "" {
			datastore.ForEachNodes(func(node *datastore.NodeEnt) bool {
				if node.Name == l.NodeName {
					n = node
					return false
				}
				return true
			})
		}
		if n != nil {
			sb.WriteString("\n# Associated Node Information\n")
			sb.WriteString(fmt.Sprintf("- Name: %s\n", n.Name))
			sb.WriteString(fmt.Sprintf("- State: %s\n", n.State))
			sb.WriteString(fmt.Sprintf("- IP Address: %s\n", n.IP))
			if n.Descr != "" {
				sb.WriteString(fmt.Sprintf("- Description: %s\n", n.Descr))
			}
		}

	case "system":
		// システムリソース・本体イベント
		if lang == "ja" {
			system = `あなたはシステムインフラおよびリソース管理の専門家です。
TWSNMP FK本体やシステム全般に関するイベントログ（リソース警告やシステム状態等）を分析し、
1. 発生したシステムイベント・警告の要約と原因
2. システム全体への影響とリスク評価
3. 推奨される対応策（リソース解放、設定見直し、メンテナンス等）
を日本語で分かりやすく解説・回答してください。`
		} else {
			system = `You are an expert in system infrastructure and resource management.
Please analyze the system-level event log (resource alerts, system state, etc.) and explain:
1. Summary and cause of the system event / alert
2. Risk assessment and impact on overall system operation
3. Recommended remediation steps (resource cleanup, configuration tuning, etc.)`
		}

		// 最近のシステムイベントログをいくつか追加してコンテキストを提供する
		sb.WriteString("\n# Recent System Event Logs\n")
		stSys := l.Time - 24*int64(time.Hour)
		etSys := l.Time + 10*int64(time.Minute)
		sysEventCount := 0
		datastore.ForEachEventLog(stSys, etSys, func(el *datastore.EventLogEnt) bool {
			if el.Type == "system" && el.Time != l.Time {
				tStr := time.Unix(0, el.Time).Format(time.RFC3339)
				sb.WriteString(fmt.Sprintf("- %s [Level: %s] %s\n", tStr, el.Level, el.Event))
				sysEventCount++
				if sysEventCount >= 10 {
					return false
				}
			}
			return true
		})
		if sysEventCount == 0 {
			sb.WriteString("No other recent system event logs found.\n")
		}

	case "cert", "cert_monitor":
		// サーバー証明書管理イベント
		if lang == "ja" {
			system = `あなたはTLS/SSLセキュリティおよびサーバー証明書管理の専門家です。
提示されたサーバー証明書監視（Cert Monitor）のイベントログおよび登録されているサーバー証明書の情報を総合的に分析し、
1. 発生した証明書イベント・警告の要約と原因（有効期限切れ、接続エラー、検証失敗等）
2. 影響範囲およびセキュリティ上のリスク評価
3. 推奨される証明書更新・設定修正・対策手順
を日本語で分かりやすく解説・回答してください。`
		} else {
			system = `You are an expert in TLS/SSL security and server certificate management.
Please analyze the provided server certificate monitor event log and certificate status details:
1. Summary and cause of the certificate event / alert (expiration, connection failure, validation error, etc.)
2. Security risk assessment and scope of impact
3. Recommended action plan for certificate renewal or configuration fixes`
		}

		// 登録されているサーバー証明書モニター一覧の情報
		sb.WriteString("\n# Server Certificate Monitor Details\n")
		certCount := 0
		datastore.ForEachCertMonitors(func(c *datastore.CertMonitorEnt) bool {
			matchedTag := ""
			if strings.Contains(l.Event, c.Target) {
				matchedTag = " [Target Match]"
			}
			tNotAfter := time.Unix(c.NotAfter, 0).Format(time.RFC3339)
			tNotBefore := time.Unix(c.NotBefore, 0).Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("## Certificate Monitor: %s:%d (State: %s)%s\n", c.Target, c.Port, c.State, matchedTag))
			if c.Subject != "" {
				sb.WriteString(fmt.Sprintf("- Subject: %s\n", c.Subject))
			}
			if c.Issuer != "" {
				sb.WriteString(fmt.Sprintf("- Issuer: %s\n", c.Issuer))
			}
			if c.NotBefore > 0 {
				sb.WriteString(fmt.Sprintf("- Valid From: %s\n", tNotBefore))
			}
			if c.NotAfter > 0 {
				sb.WriteString(fmt.Sprintf("- Valid Until (NotAfter): %s\n", tNotAfter))
			}
			if c.SerialNumber != "" {
				sb.WriteString(fmt.Sprintf("- Serial Number: %s\n", c.SerialNumber))
			}
			sb.WriteString(fmt.Sprintf("- TLS Verification: %v\n", c.Verify))
			if c.Error != "" {
				sb.WriteString(fmt.Sprintf("- Error: %s\n", c.Error))
			}
			certCount++
			return true
		})
		if certCount == 0 {
			sb.WriteString("No server certificate monitors configured.\n")
		}

	default:
		// 監視・ポーリング・ノード障害・ログイベント等 (ping, snmp, syslog, trap, http, tls, etc.)
		if lang == "ja" {
			system = `あなたはネットワーク運用、ログ解析、およびシステムトラブルシューティングの専門家です。
提示された対象イベントログの内容、関連ノード情報、関連ポーリング状態・ログ、およびイベント発生時刻前後の関連ログ（Syslog, SNMP Trap）を総合的に分析し、
1. イベントの概要および想定される発生原因
2. 関連コンポーネント・ポーリング・前後ログの状況分析
3. 推奨される確認事項・対策アクション
を日本語で分かりやすく診断・回答してください。`
		} else {
			system = `You are an expert in network operations, log analysis, and system troubleshooting.
Please analyze the provided event log, related node details, polling statuses, polling logs, and nearby Syslog/SNMP Trap logs.
Investigate and explain the following in detail:
1. Summary & Estimated Root Cause of the Event
2. Analysis of Related Components, Pollings, and Nearby Logs
3. Recommended Action Plan & Remediation Steps`
		}

		// ノードの特定
		var n *datastore.NodeEnt
		if l.NodeID != "" {
			n = datastore.GetNode(l.NodeID)
		}
		if n == nil && l.NodeName != "" {
			datastore.ForEachNodes(func(node *datastore.NodeEnt) bool {
				if node.Name == l.NodeName {
					n = node
					return false
				}
				return true
			})
		}

		// 関連ノードの情報（パスワードは除く）
		sb.WriteString("\n# Related Node Information\n")
		if n != nil {
			sb.WriteString(fmt.Sprintf("- Name: %s\n", n.Name))
			sb.WriteString(fmt.Sprintf("- State: %s\n", n.State))
			sb.WriteString(fmt.Sprintf("- IP Address: %s\n", n.IP))
			sb.WriteString(fmt.Sprintf("- MAC Address: %s\n", n.MAC))
			sb.WriteString(fmt.Sprintf("- Vendor: %s\n", n.Vendor))
			if n.Descr != "" {
				sb.WriteString(fmt.Sprintf("- Description: %s\n", n.Descr))
			}
			if n.Loc != "" {
				sb.WriteString(fmt.Sprintf("- Location: %s\n", n.Loc))
			}
			if n.URL != "" {
				sb.WriteString(fmt.Sprintf("- URL: %s\n", n.URL))
			}
			sb.WriteString(fmt.Sprintf("- SNMP Mode: %s\n", n.SnmpMode))
			if n.SnmpPort > 0 {
				sb.WriteString(fmt.Sprintf("- SNMP Port: %d\n", n.SnmpPort))
			}
			if n.AddrMode != "" {
				sb.WriteString(fmt.Sprintf("- Addr Mode: %s\n", n.AddrMode))
			}
		} else {
			sb.WriteString("No associated node information found.\n")
		}

		// 関連ポーリングの情報
		sb.WriteString("\n# Related Polling Information & Recent Logs\n")
		if n != nil {
			pollings := []datastore.PollingEnt{}
			datastore.ForEachPollings(func(p *datastore.PollingEnt) bool {
				if p.NodeID == n.ID {
					pollings = append(pollings, *p)
				}
				return true
			})

			if len(pollings) == 0 {
				sb.WriteString("No pollings configured for this node.\n")
			} else {
				for _, p := range pollings {
					matchedTag := ""
					if strings.EqualFold(p.Type, l.Type) {
						matchedTag = " [Target Type Match]"
					}
					sb.WriteString(fmt.Sprintf("## Polling: %s (Type: %s, Mode: %s, State: %s)%s\n", p.Name, p.Type, p.Mode, p.State, matchedTag))
					sb.WriteString(fmt.Sprintf("- Poll Interval: %d sec, Level: %s\n", p.PollInt, p.Level))
					if p.Params != "" {
						sb.WriteString(fmt.Sprintf("- Params: %s\n", p.Params))
					}
					if p.Filter != "" {
						sb.WriteString(fmt.Sprintf("- Filter: %s\n", p.Filter))
					}
					sb.WriteString("- Recent Polling Logs:\n")
					pLogs := datastore.GetAllPollingLog(p.ID)
					if len(pLogs) == 0 {
						sb.WriteString("  - No polling log available\n")
					} else {
						count := 0
						for i := len(pLogs) - 1; i >= 0 && count < 5; i-- {
							pl := pLogs[i]
							tStr := time.Unix(0, pl.Time).Format(time.RFC3339)
							sb.WriteString(fmt.Sprintf("  - %s | State: %s | Result: %s\n", tStr, pl.State, formatPollingResultForLLM(pl.Result)))
							count++
						}
					}
				}
			}
		} else {
			sb.WriteString("No associated polling information found.\n")
		}

		// 発生時刻に近い関連性のありそうな syslog, snmptrap
		sb.WriteString("\n# Related Logs Near Event Time\n")
		st := l.Time - 30*int64(time.Minute)
		et := l.Time + 10*int64(time.Minute)

		maxSyslog := 30
		maxTrap := 30
		if l.Type == "syslog" {
			maxSyslog = 50
		} else if l.Type == "trap" || l.Type == "snmp" {
			maxTrap = 50
		}

		// Syslog
		sb.WriteString("## Related Syslog\n")
		syslogCount := 0
		datastore.ForEachSyslog(st, et, func(sl *datastore.SyslogEnt) bool {
			isMatch := false
			if n != nil {
				if (n.IP != "" && strings.Contains(sl.Host, n.IP)) || (n.Name != "" && strings.Contains(sl.Host, n.Name)) {
					isMatch = true
				}
			} else if l.NodeName != "" && strings.Contains(sl.Host, l.NodeName) {
				isMatch = true
			}
			if isMatch {
				tStr := time.Unix(0, sl.Time).Format(time.RFC3339)
				sb.WriteString(fmt.Sprintf("- %s [Facility: %d, Severity: %d, Tag: %s] Host: %s, Message: %s\n",
					tStr, sl.Facility, sl.Severity, sl.Tag, sl.Host, sl.Message))
				syslogCount++
				if syslogCount >= maxSyslog {
					return false
				}
			}
			return true
		})
		if syslogCount == 0 {
			sb.WriteString("No related Syslog entries found near event time.\n")
		}

		// SNMP Trap
		sb.WriteString("\n## Related SNMP Traps\n")
		trapCount := 0
		datastore.ForEachTraps(st, et, func(tr *datastore.TrapEnt) bool {
			isMatch := false
			if n != nil {
				if n.IP != "" && strings.Contains(tr.FromAddress, n.IP) {
					isMatch = true
				}
			}
			if isMatch {
				tStr := time.Unix(0, tr.Time).Format(time.RFC3339)
				sb.WriteString(fmt.Sprintf("- %s [From: %s, Type: %s] Variables: %v\n",
					tStr, tr.FromAddress, tr.TrapType, tr.Variables))
				trapCount++
				if trapCount >= maxTrap {
					return false
				}
			}
			return true
		})
		if trapCount == 0 {
			sb.WriteString("No related SNMP Traps found near event time.\n")
		}
	}

	return a.llmAsk(sb.String(), system)
}

func (a *App) llmAsk(prompt, system string) *LLMResp {
	r := new(LLMResp)
	ctx := a.ctx
	llm, err := datastore.GetLLM(ctx)
	if err != nil {
		log.Printf("llmAsk err=%v", err)
		r.Error = err.Error()
		return r
	}
	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}
	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		log.Printf("llmAsk err=%v", err)
		r.Error = err.Error()
		return r
	}
	if len(resp.Choices) < 1 {
		r.Error = "no response from LLM"
		return r

	}
	r.Results = resp.Choices[0].Content
	return r
}

type LLMAssistPollingResp struct {
	TemplateID int    `json:"TemplateID"`
	Name       string `json:"Name"`
	Type       string `json:"Type"`
	Mode       string `json:"Mode"`
	Params     string `json:"Params"`
	Filter     string `json:"Filter"`
	Extractor  string `json:"Extractor"`
	Script     string `json:"Script"`
	Level      string `json:"Level"`
	PollInt    int    `json:"PollInt"`
	Timeout    int    `json:"Timeout"`
	Retry      int    `json:"Retry"`
	Advice     string `json:"Advice"`
	Error      string `json:"Error"`
}

func (a *App) LLMAssistPolling(nodeID string, prompt string) *LLMAssistPollingResp {
	r := &LLMAssistPollingResp{
		Level:   "high",
		PollInt: 60,
		Timeout: 3,
		Retry:   1,
	}
	ctx := a.ctx
	llm, err := datastore.GetLLM(ctx)
	if err != nil {
		log.Printf("LLMAssistPolling err=%v", err)
		r.Error = err.Error()
		return r
	}

	templates := a.GetPollingTemplates()
	var tmplSB strings.Builder
	for _, t := range templates {
		tmplSB.WriteString(fmt.Sprintf("- ID: %d, Name: %s, Type: %s, Mode: %s, Descr: %s, Params: %s, Filter: %s, Extractor: %s, Script: %s\n",
			t.ID, t.Name, t.Type, t.Mode, t.Descr, t.Params, t.Filter, t.Extractor, t.Script))
	}

	var nodeInfoSB strings.Builder
	if nodeID != "" {
		if n := datastore.GetNode(nodeID); n != nil {
			nodeInfoSB.WriteString(fmt.Sprintf("Name: %s, IP: %s, Vendor: %s, Descr: %s, SnmpMode: %s\n", n.Name, n.IP, n.Vendor, n.Descr, n.SnmpMode))
		}
	}

	system := `You are an expert network monitoring engineer specialized in TWSNMP FK.
Select the best polling template and native parameters based on the user's intent and target node context.

### CRITICAL TWSNMP FK SCRIPT & EVALUATION RULES:
1. **Script Condition Logic (CRITICAL)**:
   - In TWSNMP FK, the 'Script' expression MUST evaluate to **true when NORMAL / OK**, and **false when ALARM / FAULT**.
   - Example (Ping RTT < 100ms): 'rtt < 100 * 1000 * 1000' (evaluates to true if normal). DO NOT write 'rtt >= 100'!
   - Example (HTTP Status 200): 'code == 200' (evaluates to true if status is 200).
   - Example (CPU < 90%): 'cpu < 90.0' (evaluates to true if CPU is normal).

2. **Variable Names & RTT Units (CRITICAL)**:
   - DO NOT prefix variables with 'stats.'. Use direct top-level variables: 'rtt', 'code', 'status', 'count', 'exitCode', etc.
   - 'rtt' is in **NANOSECONDS (ns)**! (1 ms = 1,000,000 ns).
   - 100ms is written as '100 * 1000 * 1000' or '100000000' ns (or 'rtt / 1000000 < 100').

3. **Native Polling Types**:
   - 'ping': Mode: "" or "rtt". Script: 'rtt < 100 * 1000 * 1000' (or 'rtt / 1000000 < 100').
   - 'snmp': Mode: "get", "stats", etc. Params: OID or MIB name. Script: e.g. 'hrProcessorLoad < 90.0'.
   - 'http': Mode: "" or "status" or "https". Script: 'code == 200' or 'rtt < 2000 * 1000 * 1000'.
   - 'tls': Mode: "expire". Script: e.g. '30' (days remaining).
   - 'syslog' / 'trap': Filter: Regex pattern. Script: 'count == 0' (0 errors is normal). Params: "" (empty).
   - 'mqtt': Mode: "subscribe". Params: MQTT Broker URL (e.g. "tcp://localhost:1883"). Filter: Topic path. Script: Evaluation script on payload data.
   - 'netflow': Mode: "traffic" or "count". Filter: Address/Protocol filter (e.g. "SrcAddr = 'x.x.x.x' and DstAddr = 'y.y.y.y'"). Params: "" (empty).
   - 'command': Use ONLY when native ping, snmp, http, tls, syslog, mqtt, netflow cannot fulfill the goal.

4. **Log Filter Generalization Rule (CRITICAL)**:
   - When generating Filter regex for syslog or trap, DO NOT hardcode exact dynamic values from a single sample log (such as ephemeral source port numbers, process PIDs, timestamps, or random session IDs).
   - Replace variable parts with generic regex patterns (e.g. 'port \d+', '\[\d+\]', or omit them) so future logs with different ports/PIDs can be correctly matched.

5. **Log Polling Params Rule (CRITICAL)**:
   - For syslog, trap, netflow, and sflow log-based pollings, the target node is determined by NodeID binding and log criteria are defined in Filter.
   - Do NOT set source IP addresses or hostnames into Params. Leave Params as an empty string "".

Return ONLY a raw JSON object with the following structure:
{
  "TemplateID": <int, matching template ID or 0>,
  "Name": "<string, polling name>",
  "Type": "<string, native type>",
  "Mode": "<string, mode>",
  "Params": "<string, params>",
  "Filter": "<string, filter>",
  "Extractor": "<string, extractor>",
  "Script": "<string, script evaluating to true for NORMAL condition, e.g. rtt < 100 * 1000 * 1000>",
  "Level": "<string, fault level when Script becomes false: info, low, warn, high>",
  "PollInt": <int, interval in sec>,
  "Timeout": <int, timeout in sec>,
  "Retry": <int, retry count>,
  "Advice": "<string, detailed explanation in English on why this native polling configuration was selected, mentioning nanoseconds units and true=normal logic>"
}`

	if i18n.GetLang() == "ja" {
		system = `あなたはTWSNMP FKに特化したネットワーク監視専門家です。
ユーザーの入力した監視目的と対象ノードの情報に基づき、TWSNMP FKの組み込み仕様に完全に合致した正しいポーリング設定とパラメータ、アドバイスを生成してください。

### 【絶対順守】TWSNMP FK のスクリプト判定・変数仕様ルール:

1. **スクリプトの評価論理 (最重要)**:
   - TWSNMP FK の 'Script' は「**正常（正常稼働）である条件式 (true = 正常, false = 障害発動)**」を記述します！
   - ○ 正しい例 (Ping応答時間が100ms未満で正常): 'rtt < 100 * 1000 * 1000' （100ms未満なら true で正常、100ms以上だと false になり障害検知）
   - × 誤った例: 'rtt >= 100' （障害条件を書くと常時アラートになります）
   - ○ 正しい例 (HTTPステータスが200で正常): 'code == 200'
   - ○ 正しい例 (エラーログ件数が0で正常): 'count == 0'

2. **変数名と RTT の単位 (最重要)**:
   - 変数名に 'stats.' のような接頭辞を付けず、'rtt', 'code', 'status', 'count', 'exitCode' 等の直下変数を直接参照してください。
   - 'rtt' の単位は **ナノ秒 (ns)** です！（1 ms = 1,000,000 ns）。
   - 100ms は '100 * 1000 * 1000' または '100000000' (または 'rtt / 1000000 < 100') と記述します。

3. **ポーリング種別と設定例**:
   - 'ping': Mode: "" または "rtt"。Script: 'rtt < 100 * 1000 * 1000' (または 'rtt / 1000000 < 100')。
   - 'snmp': Mode: "get", "stats" 等。Params: OID/MIB名。Script: 'hrProcessorLoad < 90.0' 等。
   - 'http': Mode: "" や "status", "https"。Script: 'code == 200' や 'rtt < 2000 * 1000 * 1000'。
   - 'tls': Mode: "expire"。Script: '30' (残り30日以上で正常)。
   - 'syslog' / 'trap': Filter: ログ正規表現。Script: 'count == 0' (0件で正常)。Params: "" (空文字)。
   - 'mqtt': Mode: "subscribe"。Params: ブローカーURL (例: "tcp://localhost:1883")。Filter: Topic。Script: ペイロードデータの判定式。
   - 'netflow': Mode: "traffic" や "count"。Filter: 送受信フィルタ条件 (例: "SrcAddr = 'x.x.x.x' and DstAddr = 'y.y.y.y'")。Params: "" (空文字)。
   - 'command': 組み込みの ping, snmp, http, tls, syslog, mqtt, netflow 等で実現不可能な場合のみ使用。

4. **ログフィルタ正規表現の汎化・抽象化ルール (最重要)**:
   - SyslogやSNMP Trapのログから 'Filter' (正規表現) を作成する際、単一ログのサンプルに含まれる「送信元ポート番号 (例: port 56332)」、「プロセスPID (例: [1234])」、「タイムスタンプ」、「セッションID」等の動的・可変な数値をそのまま固定値(リテラル)としてハードコードしないでください！
   - 送信元ポート番号やPID等は汎用パターン (例: 'port \\d+' や '\\d+') に置き換えるか、またはポート番号部分自体を省略して、次回以降に異なるポート番号等で発生する同一イベントログも確実に検知できるように汎用的な正規表現を生成してください。

5. **ログ監視 (syslog, trap, netflow, sflow等) の Params 設定ルール (最重要)**:
   - syslog, trap, netflow, sflow 等のログ監視ポーリングでは、対象ノードは紐付け先ノード (NodeID) で識別され、ログ絞り込みは Filter で行うため、Params (パラメータ) に送信元IPアドレスやホスト名を指定せず、空文字 "" に設定してください！

【重要ルール】
PingやHTTP等の計測で外部OSコマンド('command')を無理に使用せず、TWSNMP FKネイティブのポーリング種別('ping', 'http', 'snmp', 'tls', 'syslog', 'mqtt', 'netflow')を優先選択してください。

必ず余計な説明を行わず、以下の構造のJSONオブジェクトのみを出力してください（Markdownのコードブロック枠線も不要）。

{
  "TemplateID": テンプレートID(該当するものがあれば数値、なければ0),
  "Name": "ポーリング名",
  "Type": "ポーリング種別(ping, snmp, http, tls, command, script, syslog, trap等)",
  "Mode": "モード",
  "Params": "パラメータ",
  "Filter": "フィルター条件式",
  "Extractor": "抽出式",
  "Script": "正常時にtrueとなるJavaScript判定式 (例: rtt < 100 * 1000 * 1000)",
  "Level": "Scriptがfalse(障害)となった時の障害レベル(info, low, warn, high)",
  "PollInt": ポーリング間隔(秒単位の数値),
  "Timeout": タイムアウト(秒単位の数値),
  "Retry": リトライ回数(数値),
  "Advice": "なぜこの設定が最適か、および単位(ナノ秒)と正常時trueの判定式に関するわかりやすい解説(日本語)"
}`
	}

	var userContentSB strings.Builder
	userContentSB.WriteString("# Target Node Information\n")
	if nodeInfoSB.Len() > 0 {
		userContentSB.WriteString(nodeInfoSB.String())
	} else {
		userContentSB.WriteString("No specific node selected.\n")
	}
	userContentSB.WriteString("\n# Available Polling Templates\n")
	userContentSB.WriteString(tmplSB.String())
	userContentSB.WriteString("\n# User Monitoring Purpose / Goal\n")
	if strings.TrimSpace(prompt) != "" {
		userContentSB.WriteString(prompt)
	} else {
		userContentSB.WriteString("Please suggest the top 1 most recommended polling setting for this node.")
	}

	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, userContentSB.String()),
	}

	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		log.Printf("LLMAssistPolling err=%v", err)
		r.Error = err.Error()
		return r
	}
	if len(resp.Choices) < 1 {
		r.Error = "no response from LLM"
		return r
	}

	raw := strings.TrimSpace(resp.Choices[0].Content)
	if idx := strings.Index(raw, "{"); idx != -1 {
		if lastIdx := strings.LastIndex(raw, "}"); lastIdx != -1 && lastIdx > idx {
			raw = raw[idx : lastIdx+1]
		}
	}

	if err := json.Unmarshal([]byte(raw), r); err != nil {
		log.Printf("LLMAssistPolling json parse err=%v, raw=%s", err, resp.Choices[0].Content)
		r.Advice = resp.Choices[0].Content
	}
	return r
}

func (a *App) LLMGeneratePollingScript(pollingType, mode, purpose string) *LLMResp {
	system := `You are an expert in writing JavaScript evaluation scripts and regex extractors for TWSNMP FK.
CRITICAL RULES:
1. In TWSNMP FK, Script MUST evaluate to true when NORMAL / OK, and false when ALARM / FAULT.
2. 'rtt' unit is NANOSECONDS (ns) (1ms = 1,000,000ns). Example: rtt < 100 * 1000 * 1000 for RTT < 100ms.
3. DO NOT use 'stats.' prefix on variables; use direct variable names (rtt, code, count, status).`
	if i18n.GetLang() == "ja" {
		system = `あなたはTWSNMP FK用判定スクリプト記述の専門家です。
【重要ルール】:
1. Scriptは「正常である条件式 (true = 正常, false = 障害検知)」を記述してください。
2. rttの単位はナノ秒(ns)です (1ms = 1,000,000ns)。例: 100ms未満なら rtt < 100 * 1000 * 1000
3. 変数名に stats. 接頭辞を付けず、rtt, code, count, status 等を直接参照してください。`
	}
	prompt := fmt.Sprintf("Polling Type: %s, Mode: %s\nPurpose: %s", pollingType, mode, purpose)
	return a.llmAsk(prompt, system)
}

