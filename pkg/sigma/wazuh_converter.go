package sigma

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// WazuhRuleXML represents a <rule> element in Wazuh XML
type WazuhRuleXML struct {
	ID           string           `xml:"id,attr"`
	Level        int              `xml:"level,attr"`
	Frequency    int              `xml:"frequency,attr"`
	Timeframe    int              `xml:"timeframe,attr"`
	NoAlert      string           `xml:"noalert,attr"`
	Ignore       string           `xml:"ignore,attr"`
	IfSID        string           `xml:"if_sid"`
	IfMatchedSID string           `xml:"if_matched_sid"`
	IfGroup      string           `xml:"if_group"`
	DecodedAs    string           `xml:"decoded_as"`
	ProgramName  string           `xml:"program_name"`
	Matches      []string         `xml:"match"`
	Regexes      []string         `xml:"regex"`
	Fields       []WazuhRuleField `xml:"field"`
	Description  string           `xml:"description"`
	SameSourceIP *struct{}        `xml:"same_source_ip"`
	SameUser     *struct{}        `xml:"same_user"`
	SameLocation *struct{}        `xml:"same_location"`
	MitreIDs     []string         `xml:"mitre>id"`
	Groups       string           `xml:"group"`
	ParentGroup  string           `xml:"-"`
}

// WazuhRuleField represents a <field name="..."> element
type WazuhRuleField struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// WazuhGroupXML represents a <group name="..."> element
type WazuhGroupXML struct {
	Name  string         `xml:"name,attr"`
	Rules []WazuhRuleXML `xml:"rule"`
}

// WazuhRulesRoot wraps multiple groups or rules for XML parsing
type WazuhRulesRoot struct {
	Groups []WazuhGroupXML `xml:"group"`
	Rules  []WazuhRuleXML  `xml:"rule"`
}

// SigmaCorrelation represents correlation metadata stored in AdditionalFields
type SigmaCorrelation struct {
	Frequency int      `yaml:"frequency,omitempty"`
	Timeframe string   `yaml:"timeframe,omitempty"`
	GroupBy   []string `yaml:"group_by,omitempty"`
}

// SigmaRuleOutput represents the output structure for a converted Sigma rule
type SigmaRuleOutput struct {
	Title          string                 `yaml:"title"`
	ID             string                 `yaml:"id"`
	Status         string                 `yaml:"status"`
	Description    string                 `yaml:"description"`
	Author         string                 `yaml:"author"`
	Date           string                 `yaml:"date"`
	Tags           []string               `yaml:"tags,omitempty"`
	Logsource      map[string]string      `yaml:"logsource"`
	Detection      map[string]interface{} `yaml:"detection"`
	Falsepositives []string               `yaml:"falsepositives,omitempty"`
	Level          string                 `yaml:"level"`
	Correlation    *SigmaCorrelation      `yaml:"correlation,omitempty"`
}

// ParseWazuhRulesXML parses a Wazuh rules XML document
func ParseWazuhRulesXML(data []byte) ([]WazuhRuleXML, error) {
	// Wrap in a synthetic root element because Wazuh XML often contains multiple top-level elements or comments
	wrapped := fmt.Sprintf("<root>\n%s\n</root>", string(data))
	var root WazuhRulesRoot
	decoder := xml.NewDecoder(bytes.NewReader([]byte(wrapped)))
	decoder.CharsetReader = nil
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to parse Wazuh XML: %w", err)
	}

	var rules []WazuhRuleXML
	for _, g := range root.Groups {
		for _, r := range g.Rules {
			r.ParentGroup = g.Name
			rules = append(rules, r)
		}
	}
	for _, r := range root.Rules {
		rules = append(rules, r)
	}
	return rules, nil
}

// WazuhConvertOptions controls the conversion process
type WazuhConvertOptions struct {
	MinLevel       int
	SkipFrequency  bool
	DefaultProduct string
	DefaultService string
}

// ResolvedWazuhRule contains merged conditions from parent rules
type ResolvedWazuhRule struct {
	Rule        WazuhRuleXML
	AllMatches  []string
	AllRegexes  []string
	AllFields   []WazuhRuleField
	DecodedAs   string
	ProgramName string
	Service     string
}

// ResolveRuleHierarchy resolves if_sid / if_matched_sid parent dependencies
func ResolveRuleHierarchy(rules []WazuhRuleXML, defaultService string) []ResolvedWazuhRule {
	ruleMap := make(map[string]WazuhRuleXML)
	for _, r := range rules {
		ruleMap[r.ID] = r
	}

	var resolved []ResolvedWazuhRule
	for _, r := range rules {
		res := ResolvedWazuhRule{
			Rule: r,
		}

		// Traverse ancestors to gather conditions
		currID := r.IfSID
		if currID == "" {
			currID = r.IfMatchedSID
		}
		// In case of comma-separated if_sid (e.g. "5700, 5701"), pick the first one for condition resolution
		if strings.Contains(currID, ",") {
			currID = strings.TrimSpace(strings.Split(currID, ",")[0])
		}

		visited := make(map[string]bool)
		for currID != "" && !visited[currID] {
			visited[currID] = true
			parent, exists := ruleMap[currID]
			if !exists {
				break
			}
			res.AllMatches = append(res.AllMatches, parent.Matches...)
			res.AllRegexes = append(res.AllRegexes, parent.Regexes...)
			res.AllFields = append(res.AllFields, parent.Fields...)
			if res.DecodedAs == "" {
				res.DecodedAs = parent.DecodedAs
			}
			if res.ProgramName == "" {
				res.ProgramName = parent.ProgramName
			}
			currID = parent.IfSID
			if currID == "" {
				currID = parent.IfMatchedSID
			}
			if strings.Contains(currID, ",") {
				currID = strings.TrimSpace(strings.Split(currID, ",")[0])
			}
		}

		// Add current rule's own conditions
		res.AllMatches = append(res.AllMatches, r.Matches...)
		res.AllRegexes = append(res.AllRegexes, r.Regexes...)
		res.AllFields = append(res.AllFields, r.Fields...)
		if r.DecodedAs != "" {
			res.DecodedAs = r.DecodedAs
		}
		if r.ProgramName != "" {
			res.ProgramName = r.ProgramName
		}

		// Determine service
		svc := defaultService
		if svc == "" {
			if res.DecodedAs != "" {
				svc = res.DecodedAs
			} else if res.ProgramName != "" {
				svc = strings.Trim(res.ProgramName, "^$")
			} else {
				// infer from group
				grp := r.ParentGroup + "," + r.Groups
				for _, g := range strings.Split(grp, ",") {
					g = strings.TrimSpace(g)
					if g != "" && g != "syslog" && g != "ossec" {
						svc = g
						break
					}
				}
			}
		}
		res.Service = svc
		resolved = append(resolved, res)
	}

	return resolved
}

// ConvertWazuhRuleToSigma converts a resolved Wazuh rule into a Sigma rule structure
func ConvertWazuhRuleToSigma(res ResolvedWazuhRule, opts WazuhConvertOptions) (*SigmaRuleOutput, error) {
	r := res.Rule

	// Filter rules that shouldn't alert
	if r.NoAlert == "1" || r.Level < opts.MinLevel {
		return nil, nil
	}
	if opts.SkipFrequency && r.Frequency > 0 {
		return nil, nil
	}

	product := opts.DefaultProduct
	if product == "" {
		product = "linux"
	}
	service := res.Service
	if service == "" {
		service = opts.DefaultService
	}
	if service == "" {
		service = "syslog"
	}

	// Map level
	sigmaLevel := "medium"
	switch {
	case r.Level <= 3:
		sigmaLevel = "low"
	case r.Level <= 7:
		sigmaLevel = "medium"
	case r.Level <= 11:
		sigmaLevel = "high"
	default:
		sigmaLevel = "critical"
	}

	// Build tags
	var tags []string
	for _, m := range r.MitreIDs {
		m = strings.TrimSpace(m)
		if m != "" {
			tags = append(tags, "attack."+strings.ToLower(m))
		}
	}
	tags = append(tags, "wazuh")
	allGroups := r.ParentGroup + "," + r.Groups
	for _, g := range strings.Split(allGroups, ",") {
		g = strings.TrimSpace(g)
		if g != "" && !strings.HasPrefix(g, "pci_dss") && !strings.HasPrefix(g, "gdpr") && !strings.HasPrefix(g, "nist") && !strings.HasPrefix(g, "tsc") {
			tags = append(tags, "group."+g)
		}
	}

	// Clean description
	desc := strings.TrimSpace(r.Description)
	desc = strings.ReplaceAll(desc, "\n", " ")
	for strings.Contains(desc, "  ") {
		desc = strings.ReplaceAll(desc, "  ", " ")
	}

	title := desc
	if title == "" {
		title = fmt.Sprintf("Wazuh Rule %s", r.ID)
	}

	// Build detection selection
	selection := make(map[string]interface{})

	// 1. Matches: pipe separated in Wazuh means OR within that match
	var containsList []string
	for _, m := range res.AllMatches {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		parts := strings.Split(m, "|")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				containsList = append(containsList, p)
			}
		}
	}
	if len(containsList) == 1 {
		selection["content|contains"] = containsList[0]
	} else if len(containsList) > 1 {
		selection["content|contains"] = containsList
	}

	// 2. Regexes:
	var reList []string
	for _, re := range res.AllRegexes {
		re = strings.TrimSpace(re)
		if re != "" {
			reList = append(reList, re)
		}
	}
	if len(reList) == 1 {
		selection["content|re"] = reList[0]
	} else if len(reList) > 1 {
		selection["content|re"] = reList
	}

	// 3. Fields:
	for _, f := range res.AllFields {
		fName := strings.TrimSpace(f.Name)
		fVal := strings.TrimSpace(f.Value)
		if fName != "" && fVal != "" {
			// normalize common fields: srcip -> client
			if fName == "srcip" {
				fName = "client"
			}
			selection[fName] = fVal
		}
	}

	// If no match/regex/field was found, we cannot build a valid sigma detection
	if len(selection) == 0 {
		return nil, nil
	}

	detection := map[string]interface{}{
		"selection": selection,
		"condition": "selection",
	}

	out := &SigmaRuleOutput{
		Title:       title,
		ID:          "wazuh-" + r.ID,
		Status:      "test",
		Description: desc,
		Author:      "Wazuh Ruleset / twsla converter",
		Date:        "2026-09-10",
		Tags:        tags,
		Logsource: map[string]string{
			"product": product,
			"service": service,
		},
		Detection: detection,
		Level:     sigmaLevel,
	}

	// Correlation metadata if frequency/timeframe is set
	if r.Frequency > 0 && r.Timeframe > 0 {
		groupBy := []string{}
		if r.SameSourceIP != nil {
			groupBy = append(groupBy, "client")
		}
		if r.SameUser != nil {
			groupBy = append(groupBy, "user")
		}
		if len(groupBy) == 0 {
			// default to client IP if not explicitly specified
			groupBy = append(groupBy, "client")
		}
		out.Correlation = &SigmaCorrelation{
			Frequency: r.Frequency,
			Timeframe: strconv.Itoa(r.Timeframe) + "s",
			GroupBy:   groupBy,
		}
	}

	return out, nil
}

// WazuhDecoderXML represents a <decoder> element in Wazuh XML
type WazuhDecoderXML struct {
	Name        string `xml:"name,attr"`
	Parent      string `xml:"parent"`
	ProgramName string `xml:"program_name"`
	Prematch    string `xml:"prematch"`
	Regex       string `xml:"regex"`
	Order       string `xml:"order"`
}

// WazuhDecodersRoot wraps multiple decoders for XML parsing
type WazuhDecodersRoot struct {
	Decoders []WazuhDecoderXML `xml:"decoder"`
}

// ParseWazuhDecodersXML parses a Wazuh decoders XML document
func ParseWazuhDecodersXML(data []byte) ([]WazuhDecoderXML, error) {
	wrapped := fmt.Sprintf("<root>\n%s\n</root>", string(data))
	var root WazuhDecodersRoot
	decoder := xml.NewDecoder(bytes.NewReader([]byte(wrapped)))
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("failed to parse Wazuh Decoders XML: %w", err)
	}
	return root.Decoders, nil
}

// ConvertDecoderToNamedRegex converts a Wazuh decoder to a Go named-capture regexp string
func ConvertDecoderToNamedRegex(d WazuhDecoderXML) (string, error) {
	if d.Regex == "" || d.Order == "" {
		return "", fmt.Errorf("decoder %s has no regex or order", d.Name)
	}

	fields := strings.Split(d.Order, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
		if fields[i] == "srcip" {
			fields[i] = "client"
		}
	}

	// Replace consecutive un-named capture groups `(` with `(?P<field_name>`
	re := d.Regex
	var buf strings.Builder
	fieldIdx := 0

	for i := 0; i < len(re); i++ {
		if re[i] == '(' {
			// Check if escaped
			if i > 0 && re[i-1] == '\\' {
				buf.WriteByte('(')
				continue
			}
			// Check if already a group flag like (?: or (?P<
			if i+2 < len(re) && re[i+1] == '?' {
				buf.WriteByte('(')
				continue
			}
			if fieldIdx < len(fields) {
				buf.WriteString(fmt.Sprintf("(?P<%s>", fields[fieldIdx]))
				fieldIdx++
			} else {
				buf.WriteByte('(')
			}
		} else {
			buf.WriteByte(re[i])
		}
	}

	result := buf.String()
	// Validate regexp syntax
	if _, err := regexp.Compile(result); err != nil {
		return "", fmt.Errorf("invalid converted regex %q: %w", result, err)
	}
	return result, nil
}

// FormatSigmaYAML serializes SigmaRuleOutput to YAML bytes
func FormatSigmaYAML(rule *SigmaRuleOutput) ([]byte, error) {
	return yaml.Marshal(rule)
}
