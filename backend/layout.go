package backend

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/twsnmp/twsnmpfk/datastore"
	"github.com/twsnmp/twsnmpfk/i18n"
)

type layoutSnapshot struct {
	nodes    map[string][2]int
	networks map[string][2]int
}

var (
	lastLayoutSnap *layoutSnapshot
	layoutMutex    sync.Mutex
)

// HasUndoLayout checks if an undo snapshot is available
func HasUndoLayout() bool {
	layoutMutex.Lock()
	defer layoutMutex.Unlock()
	return lastLayoutSnap != nil
}

// UndoLayout restores the positions of nodes and networks to the state before the last auto layout
func UndoLayout() (int, error) {
	layoutMutex.Lock()
	defer layoutMutex.Unlock()

	if lastLayoutSnap == nil {
		return 0, fmt.Errorf("no undo history available")
	}

	snap := lastLayoutSnap
	lastLayoutSnap = nil

	count := 0
	var updatedNodes []*datastore.NodeEnt

	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		if pos, ok := snap.nodes[n.ID]; ok {
			if n.X != pos[0] || n.Y != pos[1] {
				n.X = pos[0]
				n.Y = pos[1]
				updatedNodes = append(updatedNodes, n)
				count++
			}
		}
		return true
	})

	if len(updatedNodes) > 0 {
		_ = datastore.SaveNodes(updatedNodes)
	}

	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		if pos, ok := snap.networks[net.ID]; ok {
			if net.X != pos[0] || net.Y != pos[1] {
				net.X = pos[0]
				net.Y = pos[1]
				_ = datastore.UpdateNetwork(net)
				count++
			}
		}
		return true
	})

	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Undo auto layout: restored %d items"), count),
	})

	return count, nil
}

// saveSnapshot saves current positions of all nodes and networks for undo
func saveSnapshot() {
	snap := &layoutSnapshot{
		nodes:    make(map[string][2]int),
		networks: make(map[string][2]int),
	}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		snap.nodes[n.ID] = [2]int{n.X, n.Y}
		return true
	})
	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		snap.networks[net.ID] = [2]int{net.X, net.Y}
		return true
	})
	lastLayoutSnap = snap
}

// OptimizeLayout optimizes the positions of nodes and networks based on the specified mode
func OptimizeLayout(mode int) (int, error) {
	layoutMutex.Lock()
	defer layoutMutex.Unlock()

	if mode <= datastore.AutoLayoutNone {
		return 0, nil
	}

	saveSnapshot()

	switch mode {
	case datastore.AutoLayoutHierarchical:
		return layoutHierarchical()
	case datastore.AutoLayoutCluster:
		return layoutCluster()
	case datastore.AutoLayoutCategorized:
		return layoutCategorized()
	default:
		return 0, fmt.Errorf("unknown layout mode: %d", mode)
	}
}

type nodeLineConn struct {
	targetID string
	portID   string
}

// Helper: build line connection graph
func getTopologyGraph() (map[string][]nodeLineConn, []*datastore.LineEnt) {
	conns := make(map[string][]nodeLineConn)
	var lines []*datastore.LineEnt

	datastore.ForEachLines(func(l *datastore.LineEnt) bool {
		lines = append(lines, l)
		conns[l.NodeID1] = append(conns[l.NodeID1], nodeLineConn{
			targetID: l.NodeID2,
			portID:   l.PollingID1,
		})
		conns[l.NodeID2] = append(conns[l.NodeID2], nodeLineConn{
			targetID: l.NodeID1,
			portID:   l.PollingID2,
		})
		return true
	})
	return conns, lines
}

func isGatewayOrRouter(n *datastore.NodeEnt) bool {
	icon := strings.ToLower(n.Icon)
	name := strings.ToLower(n.Name)
	if strings.Contains(icon, "router") || strings.Contains(name, "router") ||
		strings.Contains(name, "gateway") || strings.Contains(name, "vyos") ||
		strings.Contains(name, "rtx") || strings.Contains(name, "gw") {
		return true
	}
	parts := strings.Split(n.IP, ".")
	if len(parts) == 4 && (parts[3] == "1" || parts[3] == "254") {
		return true
	}
	return false
}

func isServer(n *datastore.NodeEnt) bool {
	icon := strings.ToLower(n.Icon)
	name := strings.ToLower(n.Name)
	descr := strings.ToLower(n.Descr)
	if strings.Contains(icon, "server") || strings.Contains(icon, "hdd") ||
		strings.Contains(icon, "database") || strings.Contains(icon, "nas") ||
		strings.Contains(icon, "cloud") {
		return true
	}
	if strings.Contains(name, "server") || strings.Contains(name, "nas") ||
		strings.Contains(descr, "protocol:http") || strings.Contains(descr, "protocol:dns") ||
		strings.Contains(descr, "protocol:ldap") || strings.Contains(descr, "protocol:imap") ||
		strings.Contains(descr, "protocol:smtp") || strings.Contains(descr, "protocol:ssh") {
		return true
	}
	return false
}

func isClient(n *datastore.NodeEnt) bool {
	icon := strings.ToLower(n.Icon)
	return strings.Contains(icon, "windows") || strings.Contains(icon, "mac") ||
		strings.Contains(icon, "linux") || strings.Contains(icon, "desktop") ||
		strings.Contains(icon, "laptop") || strings.Contains(icon, "pc")
}

// --------------------------------------------------------------------------------
// 1. 階層型レイアウト (Hierarchical / Tree)
// --------------------------------------------------------------------------------
func layoutHierarchical() (int, error) {
	nodeList := []*datastore.NodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		nodeList = append(nodeList, n)
		return true
	})

	netList := []*datastore.NetworkEnt{}
	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		netList = append(netList, net)
		return true
	})

	conns, _ := getTopologyGraph()

	placedNodes := make(map[string]bool)
	updatedNodes := make(map[string]*datastore.NodeEnt)

	// Step 1: Position Gateways / Routers at Layer 0 (Y = 100)
	var gateways []*datastore.NodeEnt
	for _, n := range nodeList {
		if isGatewayOrRouter(n) {
			gateways = append(gateways, n)
		}
	}

	gwX := 150
	for _, gw := range gateways {
		gw.X = gwX
		gw.Y = 100
		placedNodes[gw.ID] = true
		updatedNodes[gw.ID] = gw
		gwX += 180
	}

	// Step 2: Position Networks (Switches) at Layer 1 (Y = 280)
	netX := 120
	for _, net := range netList {
		net.X = netX
		net.Y = 280
		_ = datastore.UpdateNetwork(net)

		// Map ports to connected nodes
		portConnectedNodes := make(map[string][]*datastore.NodeEnt)

		netIDKey := "NET:" + net.ID
		for _, conn := range conns[netIDKey] {
			if !strings.HasPrefix(conn.targetID, "NET:") && conn.portID != "" {
				node := datastore.GetNode(conn.targetID)
				if node != nil && !placedNodes[node.ID] {
					portConnectedNodes[conn.portID] = append(portConnectedNodes[conn.portID], node)
				}
			}
		}

		// Position connected nodes directly below their respective switch ports
		// Port X position on map = net.X + port.X * 45 + 30
		for _, port := range net.Ports {
			nodesForPort := portConnectedNodes[port.ID]
			if len(nodesForPort) == 0 {
				continue
			}
			portMapX := net.X + port.X*45 + 30
			baseY := net.Y + net.H + 80

			for idx, n := range nodesForPort {
				// If multiple nodes connect to same port, stack them vertically with slight X offset
				n.X = portMapX + (idx%2)*20
				n.Y = baseY + idx*90
				placedNodes[n.ID] = true
				updatedNodes[n.ID] = n
			}
		}

		netX += net.W + 80
		if netX > 2500 {
			netX = 120
		}
	}

	// Step 3: Position remaining connected nodes (e.g. server or client connected to router or other node)
	var remainingConnected []*datastore.NodeEnt
	var unconnected []*datastore.NodeEnt

	for _, n := range nodeList {
		if placedNodes[n.ID] {
			continue
		}
		if len(conns[n.ID]) > 0 {
			remainingConnected = append(remainingConnected, n)
		} else {
			unconnected = append(unconnected, n)
		}
	}

	// Position remaining connected nodes
	remX := 120
	remY := 700
	for _, n := range remainingConnected {
		n.X = remX
		n.Y = remY
		placedNodes[n.ID] = true
		updatedNodes[n.ID] = n
		remX += 130
		if remX > 2200 {
			remX = 120
			remY += 100
		}
	}

	// Step 4: Position unconnected (isolated) nodes neatly in an inventory grid at the bottom
	uncX := 120
	uncY := remY + 160
	for _, n := range unconnected {
		n.X = uncX
		n.Y = uncY
		placedNodes[n.ID] = true
		updatedNodes[n.ID] = n
		uncX += 120
		if uncX > 2200 {
			uncX = 120
			uncY += 100
		}
	}

	// Collision resolution: ensure nodes at the same Y level have at least 60px distance
	resolveHorizontalOverlap(updatedNodes)

	var saveList []*datastore.NodeEnt
	for _, n := range updatedNodes {
		saveList = append(saveList, n)
	}
	if len(saveList) > 0 {
		_ = datastore.SaveNodes(saveList)
	}

	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Optimized map layout: Hierarchical (%d nodes, %d networks)"), len(saveList), len(netList)),
	})

	return len(saveList) + len(netList), nil
}

// --------------------------------------------------------------------------------
// 2. クラスタ型レイアウト (Cluster / Hub & Spoke)
// --------------------------------------------------------------------------------
func layoutCluster() (int, error) {
	nodeList := []*datastore.NodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		nodeList = append(nodeList, n)
		return true
	})

	netList := []*datastore.NetworkEnt{}
	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		netList = append(netList, net)
		return true
	})

	conns, _ := getTopologyGraph()
	placedNodes := make(map[string]bool)
	updatedNodes := make(map[string]*datastore.NodeEnt)

	// Step 1: Arrange Networks (Switches) as cluster centers
	clusterCenterX := 500
	clusterCenterY := 500
	maxClustW := 700

	for idx, net := range netList {
		net.X = clusterCenterX - net.W/2
		net.Y = clusterCenterY - net.H/2
		_ = datastore.UpdateNetwork(net)

		// Collect all nodes connected to this network
		var connectedNodes []*datastore.NodeEnt
		netIDKey := "NET:" + net.ID
		for _, conn := range conns[netIDKey] {
			if !strings.HasPrefix(conn.targetID, "NET:") {
				n := datastore.GetNode(conn.targetID)
				if n != nil && !placedNodes[n.ID] {
					connectedNodes = append(connectedNodes, n)
					placedNodes[n.ID] = true
				}
			}
		}

		// Arrange nodes in concentric rings or circle around cluster center
		nodeCount := len(connectedNodes)
		if nodeCount > 0 {
			radius := math.Max(float64(net.W/2+140), float64(nodeCount*28))
			for i, n := range connectedNodes {
				angle := float64(i) * (2.0 * math.Pi / float64(nodeCount))
				n.X = int(float64(clusterCenterX) + radius*math.Cos(angle))
				n.Y = int(float64(clusterCenterY) + radius*math.Sin(angle))
				updatedNodes[n.ID] = n
			}
		}

		// Next cluster center position
		clusterCenterX += maxClustW + int(float64(nodeCount)*15)
		if clusterCenterX > 2200 {
			clusterCenterX = 500
			clusterCenterY += 800
		}
		_ = idx
	}

	// Step 2: Handle remaining nodes with connections (sub-hubs or standard nodes)
	var remaining []*datastore.NodeEnt
	for _, n := range nodeList {
		if !placedNodes[n.ID] {
			remaining = append(remaining, n)
		}
	}

	remX := 120
	remY := clusterCenterY + 500
	for _, n := range remaining {
		n.X = remX
		n.Y = remY
		updatedNodes[n.ID] = n
		remX += 120
		if remX > 2200 {
			remX = 120
			remY += 90
		}
	}

	var saveList []*datastore.NodeEnt
	for _, n := range updatedNodes {
		saveList = append(saveList, n)
	}
	if len(saveList) > 0 {
		_ = datastore.SaveNodes(saveList)
	}

	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Optimized map layout: Cluster (%d nodes, %d networks)"), len(saveList), len(netList)),
	})

	return len(saveList) + len(netList), nil
}

// --------------------------------------------------------------------------------
// 3. デバイス種別型レイアウト (Categorized / Device Type)
// --------------------------------------------------------------------------------
func layoutCategorized() (int, error) {
	nodeList := []*datastore.NodeEnt{}
	datastore.ForEachNodes(func(n *datastore.NodeEnt) bool {
		nodeList = append(nodeList, n)
		return true
	})

	netList := []*datastore.NetworkEnt{}
	datastore.ForEachNetworks(func(net *datastore.NetworkEnt) bool {
		netList = append(netList, net)
		return true
	})

	// Category buckets
	var gateways []*datastore.NodeEnt
	var servers []*datastore.NodeEnt
	var clients []*datastore.NodeEnt
	var others []*datastore.NodeEnt

	for _, n := range nodeList {
		if isGatewayOrRouter(n) {
			gateways = append(gateways, n)
		} else if isServer(n) {
			servers = append(servers, n)
		} else if isClient(n) {
			clients = append(clients, n)
		} else {
			others = append(others, n)
		}
	}

	updatedNodes := make(map[string]*datastore.NodeEnt)
	curY := 100

	// Section 1: Gateways & Routers
	if len(gateways) > 0 {
		gx := 120
		for _, n := range gateways {
			n.X = gx
			n.Y = curY
			updatedNodes[n.ID] = n
			gx += 150
			if gx > 2200 {
				gx = 120
				curY += 90
			}
		}
		curY += 120
	}

	// Section 2: Switches (NetworkEnt)
	if len(netList) > 0 {
		nx := 120
		maxH := 0
		for _, net := range netList {
			net.X = nx
			net.Y = curY
			if net.H > maxH {
				maxH = net.H
			}
			_ = datastore.UpdateNetwork(net)
			nx += net.W + 60
			if nx > 2200 {
				nx = 120
				curY += maxH + 40
				maxH = 0
			}
		}
		curY += maxH + 100
	}

	// Section 3: Servers & Infrastructure
	if len(servers) > 0 {
		sx := 120
		for _, n := range servers {
			n.X = sx
			n.Y = curY
			updatedNodes[n.ID] = n
			sx += 140
			if sx > 2200 {
				sx = 120
				curY += 90
			}
		}
		curY += 130
	}

	// Section 4: Clients & PCs
	if len(clients) > 0 {
		cx := 120
		for _, n := range clients {
			n.X = cx
			n.Y = curY
			updatedNodes[n.ID] = n
			cx += 130
			if cx > 2200 {
				cx = 120
				curY += 90
			}
		}
		curY += 130
	}

	// Section 5: Others & IoT
	if len(others) > 0 {
		ox := 120
		for _, n := range others {
			n.X = ox
			n.Y = curY
			updatedNodes[n.ID] = n
			ox += 120
			if ox > 2200 {
				ox = 120
				curY += 90
			}
		}
	}

	var saveList []*datastore.NodeEnt
	for _, n := range updatedNodes {
		saveList = append(saveList, n)
	}
	if len(saveList) > 0 {
		_ = datastore.SaveNodes(saveList)
	}

	datastore.AddEventLog(&datastore.EventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf(i18n.Trans("Optimized map layout: Categorized (%d nodes, %d networks)"), len(saveList), len(netList)),
	})

	return len(saveList) + len(netList), nil
}

// resolveHorizontalOverlap ensures nodes on similar Y levels do not collide horizontally
func resolveHorizontalOverlap(nodes map[string]*datastore.NodeEnt) {
	const minXDist = 60
	const yBand = 40

	// Group by Y band
	bands := make(map[int][]*datastore.NodeEnt)
	for _, n := range nodes {
		bandIdx := n.Y / yBand
		bands[bandIdx] = append(bands[bandIdx], n)
	}

	for _, list := range bands {
		if len(list) < 2 {
			continue
		}
		// Sort by X
		for i := 0; i < len(list)-1; i++ {
			for j := i + 1; j < len(list); j++ {
				if list[i].X > list[j].X {
					list[i], list[j] = list[j], list[i]
				}
			}
		}
		// Push overlapping nodes to the right
		for i := 0; i < len(list)-1; i++ {
			if list[i+1].X-list[i].X < minXDist {
				list[i+1].X = list[i].X + minXDist
			}
		}
	}
}
