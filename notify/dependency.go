package notify

import (
	"net"
	"sort"
	"strings"

	"github.com/twsnmp/twsnmpfk/datastore"
)

// DependencyAnalysisResult holds root causes and their impacted nodes
type DependencyAnalysisResult struct {
	// RootCauses contains node/network IDs that are identified as primary failures
	RootCauses []string
	// ImpactedMap maps root cause node ID to a list of secondary impacted node IDs
	ImpactedMap map[string][]string
	// ImpactedBy maps impacted node ID to its primary root cause node ID
	ImpactedBy map[string]string
}

// GetNodeOrNetworkName returns the name of a Node or Network given its ID
func GetNodeOrNetworkName(id string) string {
	if strings.HasPrefix(id, "NET:") {
		if net := datastore.GetNetwork(id); net != nil && net.Name != "" {
			return net.Name
		}
	}
	if node := datastore.GetNode(id); node != nil && node.Name != "" {
		return node.Name
	}
	return id
}

// GetRootNodeID automatically detects the root (starting) node for topology traversal.
// Priority:
// 1. Node or Network matching any of the local machine's IP addresses
// 2. Node or Network with the highest line connection degree (prioritizing router/gateway/switch icons)
// 3. Any node/network connected to a line
func GetRootNodeID() string {
	localIPs := getLocalIPs()
	var matchingID string

	// 1. Check if any node matches local host IP
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if n.IP != "" {
			for _, lip := range localIPs {
				if n.IP == lip {
					matchingID = n.ID
					return false
				}
			}
		}
		return true
	})
	if matchingID != "" {
		return matchingID
	}

	// 1b. Check if any network matches local host IP
	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		if net.IP != "" {
			for _, lip := range localIPs {
				if net.IP == lip {
					matchingID = net.ID
					return false
				}
			}
		}
		return true
	})
	if matchingID != "" {
		return matchingID
	}

	// 2. Fallback: Count line connections per node/network
	degreeMap := make(map[string]int)
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 != "" && l.NodeID2 != "" {
			degreeMap[l.NodeID1]++
			degreeMap[l.NodeID2]++
		}
		return true
	})

	if len(degreeMap) == 0 {
		return ""
	}

	bestID := ""
	bestScore := -1

	for id, degree := range degreeMap {
		score := degree * 10
		if strings.HasPrefix(id, "NET:") {
			score += 50
		} else if node := datastore.GetNode(id); node != nil {
			icon := strings.ToLower(node.Icon)
			if strings.Contains(icon, "router") || strings.Contains(icon, "gateway") {
				score += 100
			} else if strings.Contains(icon, "switch") || strings.Contains(icon, "hub") {
				score += 50
			}
		}
		if score > bestScore {
			bestScore = score
			bestID = id
		}
	}

	return bestID
}

// BuildDependencyTree creates a parentMap and depthMap from lines using BFS starting from rootID.
func BuildDependencyTree(rootID string) (parentMap map[string]string, depthMap map[string]int) {
	parentMap = make(map[string]string)
	depthMap = make(map[string]int)

	adj := make(map[string][]string)
	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		if l.NodeID1 != "" && l.NodeID2 != "" && l.NodeID1 != l.NodeID2 {
			adj[l.NodeID1] = append(adj[l.NodeID1], l.NodeID2)
			adj[l.NodeID2] = append(adj[l.NodeID2], l.NodeID1)
		}
		return true
	})

	if len(adj) == 0 {
		return parentMap, depthMap
	}

	visited := make(map[string]bool)

	bfs := func(start string) {
		queue := []string{start}
		visited[start] = true
		depthMap[start] = 0

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			for _, neighbor := range adj[curr] {
				if !visited[neighbor] {
					visited[neighbor] = true
					parentMap[neighbor] = curr
					depthMap[neighbor] = depthMap[curr] + 1
					queue = append(queue, neighbor)
				}
			}
		}
	}

	if rootID != "" && len(adj[rootID]) > 0 {
		bfs(rootID)
	}

	// For any unvisited disconnected islands in the graph
	for nodeID := range adj {
		if !visited[nodeID] {
			bfs(nodeID)
		}
	}

	return parentMap, depthMap
}

// AnalyzeFailureDependencies evaluates a list of event logs and partitions failed nodes
// into root causes and secondary impacted nodes.
func AnalyzeFailureDependencies(list []*datastore.EventLogEnt) *DependencyAnalysisResult {
	res := &DependencyAnalysisResult{
		RootCauses:  []string{},
		ImpactedMap: make(map[string][]string),
		ImpactedBy:  make(map[string]string),
	}

	// 1. Collect failed node IDs from the current event list (only polling events)
	failedInList := make(map[string]bool)
	for _, l := range list {
		if l.NodeID == "" || l.Type != "polling" || l.Level == "normal" || l.Level == "repair" || l.Level == "info" {
			continue
		}
		failedInList[l.NodeID] = true
	}

	if len(failedInList) == 0 {
		return res
	}

	// 2. Also check nodes that are currently in down/error states in the datastore
	// (in case the parent failed earlier and is already down)
	currentlyFailing := make(map[string]bool)
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		switch n.State {
		case "high", "low", "warn":
			currentlyFailing[n.ID] = true
		}
		return true
	})
	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		if net.Error != "" {
			currentlyFailing[net.ID] = true
		}
		return true
	})

	rootID := GetRootNodeID()
	parentMap, depthMap := BuildDependencyTree(rootID)

	// Sort failed nodes by depth (shallowest first)
	failedNodes := make([]string, 0, len(failedInList))
	for nid := range failedInList {
		failedNodes = append(failedNodes, nid)
	}

	sort.Slice(failedNodes, func(i, j int) bool {
		return depthMap[failedNodes[i]] < depthMap[failedNodes[j]]
	})

	for _, nid := range failedNodes {
		// Trace upwards through parentMap to find if any ancestor is failing
		ancestor := parentMap[nid]
		var rootCauseID string
		for ancestor != "" {
			if failedInList[ancestor] || currentlyFailing[ancestor] {
				if primary, ok := res.ImpactedBy[ancestor]; ok {
					rootCauseID = primary
				} else {
					rootCauseID = ancestor
				}
			}
			ancestor = parentMap[ancestor]
		}

		if rootCauseID != "" {
			res.ImpactedBy[nid] = rootCauseID
			res.ImpactedMap[rootCauseID] = append(res.ImpactedMap[rootCauseID], nid)
		} else {
			res.RootCauses = append(res.RootCauses, nid)
		}
	}

	return res
}

func getLocalIPs() []string {
	ips := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				ips = append(ips, ip.String())
			} else if ip := ipnet.IP.To16(); ip != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}
