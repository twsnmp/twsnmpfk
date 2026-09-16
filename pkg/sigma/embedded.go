package sigma

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed all:rules all:config
var sigmaFS embed.FS

// GetAvailableSigmaPacks returns list of available embedded rule packs
func GetAvailableSigmaPacks() []string {
	ret := []string{}
	entries, err := sigmaFS.ReadDir("rules/packs")
	if err != nil {
		return ret
	}
	for _, entry := range entries {
		if entry.IsDir() {
			ret = append(ret, entry.Name())
		}
	}
	return ret
}

// SigmaPackInfo contains metadata about an embedded Sigma rule pack
type SigmaPackInfo struct {
	Name        string   `json:"name"`
	RuleCount   int      `json:"rule_count"`
	Description string   `json:"description"`
	Rules       []string `json:"rules,omitempty"`
}

// GetSigmaPackDescription returns a brief human-readable description of the pack
func GetSigmaPackDescription(pack string) string {
	switch pack {
	case "windows-essential":
		return "Essential security monitoring for Windows environments (failed logons, privilege escalation, defender tampering)"
	case "windows-ad":
		return "Active Directory / Domain Controller threats (Kerberoasting, AS-REP roasting, DCSync, trust changes)"
	case "windows-client":
		return "Windows endpoint & client threats (suspicious RDP, UAC bypass, USB media, LSASS dump)"
	case "linux-auth":
		return "Linux authentication and privilege escalation (SSH brute-force, sudo abuse, PAM failures)"
	case "linux-system":
		return "Linux persistence, rootkits, suspicious cron jobs, and system tampering"
	case "network-threats":
		return "Network devices, firewalls, and VPN threats (Cisco, Fortinet, PAN-OS, brute-force)"
	case "web-attacks":
		return "Web application attacks (SQL injection, XSS, directory traversal, web shells)"
	case "wazuh-compliance":
		return "Wazuh compliance and hardening verification rules (PCI-DSS, NIST, GDPR, CIS)"
	case "wazuh-linux":
		return "Wazuh converted Linux security and daemon rules"
	case "wazuh-network":
		return "Wazuh converted network device and firewall rules"
	case "wazuh-web":
		return "Wazuh converted web application and server access rules"
	default:
		return "Embedded Sigma rule pack: " + pack
	}
}

// GetSigmaPackInfo returns pack information including rule count and optional rule list
func GetSigmaPackInfo(pack string, includeRules bool) (*SigmaPackInfo, error) {
	packDir := path.Join("rules", "packs", pack)
	info := &SigmaPackInfo{
		Name:        pack,
		Description: GetSigmaPackDescription(pack),
	}
	err := fs.WalkDir(sigmaFS, packDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == ".yaml" || ext == ".yml" {
			info.RuleCount++
			if includeRules {
				info.Rules = append(info.Rules, filepath.Base(filePath))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if info.RuleCount == 0 {
		return nil, fmt.Errorf("pack %s not found or contains no rules", pack)
	}
	return info, nil
}

// GetAllSigmaPacksInfo returns metadata for all available packs
func GetAllSigmaPacksInfo() []*SigmaPackInfo {
	packs := GetAvailableSigmaPacks()
	var ret []*SigmaPackInfo
	for _, p := range packs {
		if info, err := GetSigmaPackInfo(p, false); err == nil {
			ret = append(ret, info)
		}
	}
	return ret
}

// ForEachSigmaConfig iterates over embedded configs and optional custom directory configs
func ForEachSigmaConfig(customDir string, callBack func(c string, d []byte)) {
	entries, err := sigmaFS.ReadDir("config")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".yaml" || ext == ".yml" {
				c := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
				p := path.Join("config", entry.Name())
				if d, err := sigmaFS.ReadFile(p); err == nil {
					callBack(c, d)
				}
			}
		}
	}
	if customDir == "" {
		return
	}
	_ = filepath.WalkDir(customDir, func(p string, info fs.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		c := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		if d, err := os.ReadFile(p); err == nil {
			callBack(c, d)
		}
		return nil
	})
}

// ForEachSigmaRulesWithSource iterates over specified packs and rule paths, passing raw YAML and source label
func ForEachSigmaRulesWithSource(rulePath string, packs []string, callBack func(c []byte, path, source string)) {
	// 1. Load packs from embed
	for _, pack := range packs {
		packDir := path.Join("rules", "packs", pack)
		source := "pack:" + pack
		_ = fs.WalkDir(sigmaFS, packDir, func(filePath string, info fs.DirEntry, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(filePath))
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}
			if c, err := sigmaFS.ReadFile(filePath); err == nil {
				callBack(c, filePath, source)
			}
			return nil
		})
	}

	// 2. Load custom external rule files/dirs if specified
	if rulePath != "" {
		for _, p := range getRulePaths(rulePath) {
			c, err := os.ReadFile(p)
			if err != nil {
				log.Printf("invalid rule file %s err=%v", p, err)
				continue
			}
			callBack(c, p, "file:"+p)
		}
	}
}

// getRulePaths recursively resolves rule file paths
func getRulePaths(root string) []string {
	var ret []string
	info, err := os.Stat(root)
	if err != nil {
		return ret
	}
	if !info.IsDir() {
		ext := strings.ToLower(filepath.Ext(root))
		if ext == ".yaml" || ext == ".yml" {
			return []string{root}
		}
		return ret
	}
	_ = filepath.WalkDir(root, func(p string, info fs.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext == ".yaml" || ext == ".yml" {
			ret = append(ret, p)
		}
		return nil
	})
	return ret
}
