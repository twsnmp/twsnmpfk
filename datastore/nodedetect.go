package datastore

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/gosnmp/gosnmp"
)

// SensorPollingDef represents sensor polling configuration (CPU, memory, temp, fan, power).
type SensorPollingDef struct {
	Name   string `json:"Name"`
	Type   string `json:"Type"`
	Mode   string `json:"Mode"`
	Params string `json:"Params"`
	Script string `json:"Script"`
	Level  string `json:"Level"`
	Descr  string `json:"Descr"`
}

// NodeDetectRule represents a device/OS detection rule.
type NodeDetectRule struct {
	ID               string             `json:"ID"`
	Name             string             `json:"Name"`
	Category         string             `json:"Category"`
	SubType          string             `json:"SubType,omitempty"`
	Icon             string             `json:"Icon"`
	SysObjectIDs     []string           `json:"SysObjectIDs,omitempty"`
	SysDescrPattern  string             `json:"SysDescrPattern,omitempty"`
	HTTPTitlePattern string             `json:"HTTPTitlePattern,omitempty"`
	VendorPattern    string             `json:"VendorPattern,omitempty"`
	SensorPollings   []SensorPollingDef `json:"SensorPollings,omitempty"`

	sysDescrReg  *regexp.Regexp
	httpTitleReg *regexp.Regexp
	vendorReg    *regexp.Regexp
}

// DetectInput holds signatures collected from a node for classification.
type DetectInput struct {
	IP          string
	Name        string
	HostName    string
	SysObjectID string
	SysDescr    string
	HTTPTitle   string
	HTTPServer  string
	HTTPBody    string
	Vendor      string
	IsGateway   bool
}

// DetectResult contains the classified node type, icon, and recommended pollings.
type DetectResult struct {
	RuleID         string             `json:"RuleID"`
	Name           string             `json:"Name"`
	Category       string             `json:"Category"`
	SubType        string             `json:"SubType"`
	Icon           string             `json:"Icon"`
	Confidence     int                `json:"Confidence"`
	SensorPollings []SensorPollingDef `json:"SensorPollings"`
}

var (
	nodeDetectRules []*NodeDetectRule
	nodeDetectLock  sync.RWMutex
)

// LoadNodeDetectRules loads detection rules from embedded FS and optional user override.
func LoadNodeDetectRules() {
	nodeDetectLock.Lock()
	defer nodeDetectLock.Unlock()

	nodeDetectRules = []*NodeDetectRule{}

	// 1. Load embedded conf/nodedetect.json
	if r, err := conf.Open("conf/nodedetect.json"); err == nil {
		if b, err := io.ReadAll(r); err == nil && len(b) > 0 {
			var rules []*NodeDetectRule
			if err := json.Unmarshal(b, &rules); err == nil {
				compileRules(rules)
				nodeDetectRules = append(nodeDetectRules, rules...)
			} else {
				log.Printf("LoadNodeDetectRules json err=%v", err)
			}
		}
		_ = r.Close()
	} else {
		log.Printf("LoadNodeDetectRules open embedded err=%v", err)
	}

	// 2. Load user override dspath/nodedetect.json if present (prepended for higher priority)
	userFile := filepath.Join(dspath, "nodedetect.json")
	if b, err := os.ReadFile(userFile); err == nil && len(b) > 0 {
		var userRules []*NodeDetectRule
		if err := json.Unmarshal(b, &userRules); err == nil {
			compileRules(userRules)
			nodeDetectRules = append(userRules, nodeDetectRules...)
		}
	}
}

func compileRules(rules []*NodeDetectRule) {
	for _, rule := range rules {
		if rule.SysDescrPattern != "" {
			rule.sysDescrReg, _ = regexp.Compile(rule.SysDescrPattern)
		}
		if rule.HTTPTitlePattern != "" {
			rule.httpTitleReg, _ = regexp.Compile(rule.HTTPTitlePattern)
		}
		if rule.VendorPattern != "" {
			rule.vendorReg, _ = regexp.Compile(rule.VendorPattern)
		}
	}
}

func normalizeOID(oid string) string {
	oid = strings.TrimSpace(oid)
	if oid == "" {
		return ""
	}
	if !strings.HasPrefix(oid, ".") {
		oid = "." + oid
	}
	return strings.TrimRight(oid, ".")
}

// DetectNode identifies device category, OS, icon, and recommended pollings.
func DetectNode(input *DetectInput) *DetectResult {
	nodeDetectLock.RLock()
	rules := nodeDetectRules
	nodeDetectLock.RUnlock()

	if len(rules) == 0 {
		LoadNodeDetectRules()
		nodeDetectLock.RLock()
		rules = nodeDetectRules
		nodeDetectLock.RUnlock()
	}

	var bestRule *NodeDetectRule
	highestScore := 0

	normInputOID := normalizeOID(input.SysObjectID)
	sysObjName := ""
	if normInputOID != "" && MIBDB != nil {
		sysObjName = MIBDB.OIDToName(normInputOID)
	}

	isGateway := input.IsGateway
	if !isGateway && input.IP != "" {
		if gw := GetDefaultGateway(); gw != "" && gw == input.IP {
			isGateway = true
		}
	}

	for _, rule := range rules {
		score := 0

		// 1. sysObjectID check
		if normInputOID != "" && len(rule.SysObjectIDs) > 0 {
			for _, target := range rule.SysObjectIDs {
				targetOID := ""
				if MIBDB != nil {
					targetOID = MIBDB.NameToOID(target)
				}
				if targetOID == ".0.0" || targetOID == "" {
					targetOID = target
				}
				normTarget := normalizeOID(targetOID)
				if normTarget != "" && (normInputOID == normTarget || strings.HasPrefix(normInputOID, normTarget+".")) {
					score += 100 + len(normTarget)
					break
				}
				// Also check if symbolic OIDToName matches target
				if sysObjName != "" {
					if sysObjName == target || strings.HasPrefix(sysObjName, target+".") {
						score += 100 + len(target)
						break
					}
				}
			}
		}

		// 2. sysDescr / HostName / Name check
		if rule.sysDescrReg != nil {
			if input.SysDescr != "" && rule.sysDescrReg.MatchString(input.SysDescr) {
				score += 50
			} else if input.HostName != "" && rule.sysDescrReg.MatchString(input.HostName) {
				score += 40
			} else if input.Name != "" && rule.sysDescrReg.MatchString(input.Name) {
				score += 40
			}
		}

		// 3. HTTP check
		if rule.httpTitleReg != nil {
			matchedHTTP := false
			if input.HTTPTitle != "" && rule.httpTitleReg.MatchString(input.HTTPTitle) {
				matchedHTTP = true
			} else if input.HTTPServer != "" && rule.httpTitleReg.MatchString(input.HTTPServer) {
				matchedHTTP = true
			} else if input.HTTPBody != "" && rule.httpTitleReg.MatchString(input.HTTPBody) {
				matchedHTTP = true
			}
			if matchedHTTP {
				score += 30
			}
		}

		// 4. MAC Vendor check
		if input.Vendor != "" && rule.vendorReg != nil {
			if rule.vendorReg.MatchString(input.Vendor) {
				score += 30
				// Bonus when physical hardware vendor matches alongside SNMP/agent OID
				if normInputOID != "" && len(rule.SysObjectIDs) > 0 {
					score += 30
				}
			}
		}

		// 5. Gateway check (bonus for router rules that already matched vendor/HTTP/SNMP)
		if isGateway && rule.Category == "router" && score > 0 {
			score += 40
		}

		if score > highestScore {
			highestScore = score
			bestRule = rule
		}
	}

	if bestRule != nil && highestScore >= 20 {
		return &DetectResult{
			RuleID:         bestRule.ID,
			Name:           bestRule.Name,
			Category:       bestRule.Category,
			SubType:        bestRule.SubType,
			Icon:           bestRule.Icon,
			Confidence:     highestScore,
			SensorPollings: bestRule.SensorPollings,
		}
	}

	// Fallback to Gateway Router if it is the host's default gateway
	if isGateway {
		return &DetectResult{
			RuleID:         "default_gateway_router",
			Name:           "Gateway Router",
			Category:       "router",
			Icon:           "router",
			Confidence:     80,
			SensorPollings: nil,
		}
	}

	// Fallback to Unknown
	defaultIcon := "desktop"
	if normInputOID != "" {
		defaultIcon = "hdd"
	}
	return &DetectResult{
		RuleID:         "unknown",
		Name:           "Unknown Device",
		Category:       "unknown",
		Icon:           defaultIcon,
		Confidence:     0,
		SensorPollings: nil,
	}
}

// CheckSensorSupport checks whether the target SNMP agent responds successfully to a sensor OID/symbol.
func CheckSensorSupport(agent *gosnmp.GoSNMP, symbolWithIndex string) bool {
	if agent == nil || symbolWithIndex == "" {
		return false
	}
	needClose := false
	if agent.Conn == nil {
		if err := agent.Connect(); err != nil {
			return false
		}
		needClose = true
	}
	if needClose {
		defer agent.Conn.Close()
	}
	oid := symbolWithIndex
	if MIBDB != nil {
		resolved := MIBDB.NameToOID(symbolWithIndex)
		if resolved != "" && resolved != ".0.0" {
			oid = resolved
		} else {
			parts := strings.SplitN(symbolWithIndex, ".", 2)
			if len(parts) == 2 {
				baseOID := MIBDB.NameToOID(parts[0])
				if baseOID != "" && baseOID != ".0.0" {
					oid = baseOID + "." + parts[1]
				}
			}
		}
	}
	res, err := agent.Get([]string{oid})
	if err != nil || len(res.Variables) == 0 {
		return false
	}
	v := res.Variables[0]
	if v.Type == gosnmp.NoSuchObject || v.Type == gosnmp.NoSuchInstance || v.Type == gosnmp.Null {
		return false
	}
	return true
}

