package main

import (
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
					sb.WriteString(fmt.Sprintf("  - %s | State: %s | Result: %v\n", tStr, l.State, l.Result))
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
Please analyze the provided node's information, polling statuses, polling logs, and related logs (Syslog, SNMP Trap, ARP).
Diagnose the overall status, highlight any anomalies or risks, and provide recommended troubleshooting or optimization actions.`
	if i18n.GetLang() == "ja" {
		system = `あなたはネットワークおよびシステム管理の専門家です。
提示されたノードの基本情報、ポーリング状態・ログ、および関連ログ（Syslog, SNMP Trap, ARP）を分析し、
1. ノードの現在の状態・健全性の要約
2. 検出された問題点・異常・リスクの評価
3. 推奨される対策・具体的なアクション
を日本語で分かりやすく診断・回答してください。`
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
