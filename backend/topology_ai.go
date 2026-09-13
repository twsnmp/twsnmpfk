package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/twsnmp/twsnmpfk/datastore"
)

type AIInferLink struct {
	FromNode  string `json:"from_node"`
	ToNetwork string `json:"to_network"`
	PortIndex string `json:"port_index"`
	Reason    string `json:"reason"`
}

type AIInferResponse struct {
	Links []AIInferLink `json:"links"`
}

// InferRemainingLinesWithAI infers connection lines for nodes that have no lines connected
func InferRemainingLinesWithAI() ([]NeighborLineEnt, error) {
	// 1. Collect connected node IDs
	connectedNodes := make(map[string]bool)
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		connectedNodes[l.NodeID1] = true
		connectedNodes[l.NodeID2] = true
		return true
	})

	// 2. Collect unattached nodes
	var unattachedNodes []*datastore.NodeEnt
	datastore.ForEachNodes(func(nd *datastore.NodeEnt) bool {
		if !connectedNodes[nd.ID] && nd.IP != "" {
			unattachedNodes = append(unattachedNodes, nd)
		}
		return true
	})

	if len(unattachedNodes) == 0 {
		return nil, nil
	}

	// 3. Collect networks
	var networks []*datastore.NetworkEnt
	datastore.ForEachNetworks(func(netEnt *datastore.NetworkEnt) bool {
		networks = append(networks, netEnt)
		return true
	})

	if len(networks) == 0 {
		return nil, nil
	}

	// 4. Try LLM first if available
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	llm, err := datastore.GetLLM(ctx)
	if err == nil && llm != nil {
		lines, err := inferWithLLM(ctx, llm, unattachedNodes, networks)
		if err == nil && len(lines) > 0 {
			return lines, nil
		}
		log.Printf("LLM topology inference fallback to heuristic: %v", err)
	}

	// 5. Heuristic fallback: Subnet / IP segment matching
	return inferWithSubnetHeuristics(unattachedNodes, networks), nil
}

func inferWithLLM(ctx context.Context, llm llms.Model, nodes []*datastore.NodeEnt, networks []*datastore.NetworkEnt) ([]NeighborLineEnt, error) {
	type NodeSummary struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		IP       string `json:"ip"`
		MAC      string `json:"mac"`
		Vendor   string `json:"vendor"`
		Icon     string `json:"icon"`
		SysDescr string `json:"sysDescr"`
	}
	type NetworkSummary struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		IP       string   `json:"ip"`
		PortList []string `json:"ports"`
	}

	var nodeSummaries []NodeSummary
	for _, n := range nodes {
		nodeSummaries = append(nodeSummaries, NodeSummary{
			ID:       n.ID,
			Name:     n.Name,
			IP:       n.IP,
			MAC:      n.MAC,
			Vendor:   n.Vendor,
			Icon:     n.Icon,
			SysDescr: n.Descr,
		})
	}

	var netSummaries []NetworkSummary
	for _, n := range networks {
		var ports []string
		for _, p := range n.Ports {
			ports = append(ports, fmt.Sprintf("index:%s(name:%s)", p.Index, p.Name))
		}
		netSummaries = append(netSummaries, NetworkSummary{
			ID:       n.ID,
			Name:     n.Name,
			IP:       n.IP,
			PortList: ports,
		})
	}

	nodesJSON, _ := json.MarshalIndent(nodeSummaries, "", "  ")
	netsJSON, _ := json.MarshalIndent(netSummaries, "", "  ")

	systemPrompt := `You are an expert network engineer and topology engine.
You are given a list of unattached nodes and available network switches.
Infer the most probable switch connections for each unattached node based on IP subnets, device roles, vendors, and naming conventions.
Output strictly valid JSON with the following schema, with no markdown code fences:
{
  "links": [
    {
      "from_node": "node_id",
      "to_network": "network_id",
      "port_index": "port_index_or_empty",
      "reason": "explanation"
    }
  ]
}`

	userPrompt := fmt.Sprintf("Unattached Nodes:\n%s\n\nSwitches:\n%s\n\nInfer topology connections:", string(nodesJSON), string(netsJSON))

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}

	resp, err := llm.GenerateContent(ctx, content, llms.WithTemperature(0.1))
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices from LLM")
	}

	rawText := strings.TrimSpace(resp.Choices[0].Content)
	rawText = cleanMarkdownJSON(rawText)

	var aiResp AIInferResponse
	if err := json.Unmarshal([]byte(rawText), &aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	var resultLines []NeighborLineEnt
	for _, link := range aiResp.Links {
		netEnt := datastore.GetNetwork(link.ToNetwork)
		nodeEnt := datastore.GetNode(link.FromNode)
		if netEnt == nil || nodeEnt == nil {
			continue
		}

		pID := ""
		for _, p := range netEnt.Ports {
			if p.Index == link.PortIndex || p.Name == link.PortIndex {
				pID = p.ID
				break
			}
		}
		if pID == "" && len(netEnt.Ports) > 0 {
			pID = netEnt.Ports[0].ID
		}

		nodePid := findBestPollingIDForNode(nodeEnt, link.PortIndex)
		reason := "AI: " + link.Reason
		if link.Reason == "" {
			reason = "AI-Inference"
		}

		resultLines = append(resultLines, NeighborLineEnt{
			LineEnt: datastore.LineEnt{
				NodeID1:    fmt.Sprintf("NET:%s", netEnt.ID),
				PollingID1: pID,
				NodeID2:    nodeEnt.ID,
				PollingID2: nodePid,
				Width:      2,
				Info:       reason,
			},
			Confidence: "speculative",
			Reason:     reason,
		})
	}

	return resultLines, nil
}

// inferWithSubnetHeuristics maps nodes to switches on the same IPv4 subnet
func inferWithSubnetHeuristics(nodes []*datastore.NodeEnt, networks []*datastore.NetworkEnt) []NeighborLineEnt {
	var resultLines []NeighborLineEnt

	type netSubnet struct {
		netEnt *datastore.NetworkEnt
		ip     net.IP
	}
	var switchList []netSubnet
	for _, n := range networks {
		if ip := net.ParseIP(n.IP).To4(); ip != nil {
			switchList = append(switchList, netSubnet{netEnt: n, ip: ip})
		}
	}

	if len(switchList) == 0 {
		return nil
	}

	for _, nd := range nodes {
		nodeIP := net.ParseIP(nd.IP).To4()
		if nodeIP == nil {
			continue
		}

		// Find closest switch (matching /24 subnet first)
		var bestNet *datastore.NetworkEnt
		for _, s := range switchList {
			if s.ip[0] == nodeIP[0] && s.ip[1] == nodeIP[1] && s.ip[2] == nodeIP[2] {
				bestNet = s.netEnt
				break
			}
		}
		if bestNet == nil {
			bestNet = switchList[0].netEnt
		}

		pID := ""
		if len(bestNet.Ports) > 0 {
			pID = bestNet.Ports[0].ID
		}
		nodePid := findBestPollingIDForNode(nd, "")

		reason := fmt.Sprintf("Subnet Heuristic (%s)", bestNet.Name)
		resultLines = append(resultLines, NeighborLineEnt{
			LineEnt: datastore.LineEnt{
				NodeID1:    fmt.Sprintf("NET:%s", bestNet.ID),
				PollingID1: pID,
				NodeID2:    nd.ID,
				PollingID2: nodePid,
				Width:      2,
				Info:       reason,
			},
			Confidence: "speculative",
			Reason:     reason,
		})
	}

	return resultLines
}

func cleanMarkdownJSON(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		var body []string
		for _, line := range lines {
			if !strings.HasPrefix(line, "```") {
				body = append(body, line)
			}
		}
		text = strings.Join(body, "\n")
	}
	return strings.TrimSpace(text)
}
