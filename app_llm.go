package main

import (
	"log"
	"strings"

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
