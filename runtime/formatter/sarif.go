package formatter

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// SARIFFormatter implements SARIF 2.1.0 output for GitHub Security tab integration.
type SARIFFormatter struct {
	version string
}

// NewSARIFFormatter creates a new SARIFFormatter.
func NewSARIFFormatter() *SARIFFormatter {
	return &SARIFFormatter{
		version: "dev",
	}
}

// SetVersion sets the scanner version for SARIF output.
func (f *SARIFFormatter) SetVersion(version string) {
	f.version = version
}

// Format outputs diagnostics in SARIF 2.1.0 format.
func (f *SARIFFormatter) Format(d *diagnostics.Diagnostics) ([]byte, error) {
	if d == nil {
		d = diagnostics.NewDiagnostics()
	}

	sarif := map[string]any{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/csd03/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]any{
			{
				"tool": map[string]any{
					"driver": map[string]any{
						"name":           "Atheon",
						"version":        f.version,
						"informationUri": "https://github.com/aliasfoxkde/Atheon-Enhanced",
						"rules":          buildRules(d.Issues),
					},
				},
				"originalUriBaseIds": map[string]any{
					"SRCROOT": map[string]any{
						"uri": "file:///",
					},
				},
				"results": buildResults(d.Issues),
				"properties": map[string]any{
					"summary": map[string]any{
						"total_findings": d.Summary.TotalFindings,
						"critical":       d.Summary.Critical,
						"high":           d.Summary.High,
						"medium":         d.Summary.Medium,
						"low":            d.Summary.Low,
						"info":           d.Summary.Info,
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SARIF: %w", err)
	}

	return append(data, '\n'), nil
}

// ContentType returns the MIME type for SARIF output.
func (f *SARIFFormatter) ContentType() string {
	return "application/sarif+json"
}

// FileExtension returns the file extension for SARIF output.
func (f *SARIFFormatter) FileExtension() string {
	return ".sarif"
}

// sarifSeverityScore maps severity string to CVSS-like 0.0-10.0 score.
func sarifSeverityScore(sev string) string {
	switch sev {
	case "critical":
		return "9.5"
	case "high":
		return "7.5"
	case "medium":
		return "5.0"
	case "low":
		return "2.5"
	default:
		return "5.0"
	}
}

// sarifLevel maps severity string to SARIF level enum.
func sarifLevel(sev string) string {
	switch sev {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "note"
	default:
		return "none"
	}
}

// buildRules creates the SARIF rules array from issues.
func buildRules(issues []diagnostics.Issue) []map[string]any {
	// Collect unique rules from issues
	ruleMap := make(map[string]map[string]any)
	for _, issue := range issues {
		if _, exists := ruleMap[issue.RuleID]; !exists {
			ruleMap[issue.RuleID] = map[string]any{
				"id":   issue.RuleID,
				"name": issue.RuleID,
				"shortDescription": map[string]any{
					"text": fmt.Sprintf("Rule %s detected an issue.", issue.RuleID),
				},
				"fullDescription": map[string]any{
					"text": issue.Message,
				},
				"kind": "rule",
				"defaultConfiguration": map[string]any{
					"level": sarifLevel(issue.Severity),
				},
				"properties": map[string]any{
					"security-severity": sarifSeverityScore(issue.Severity),
				},
			}
		}
	}

	// Convert to slice and sort for deterministic output
	rules := make([]map[string]any, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i]["id"].(string) < rules[j]["id"].(string)
	})

	return rules
}

// buildResults creates the SARIF results array from issues.
func buildResults(issues []diagnostics.Issue) []map[string]any {
	results := make([]map[string]any, 0, len(issues))
	for _, issue := range issues {
		region := map[string]any{
			"startLine": issue.Line,
		}
		if issue.Column > 0 {
			region["startColumn"] = issue.Column
		}

		result := map[string]any{
			"ruleId": issue.RuleID,
			"level":  sarifLevel(issue.Severity),
			"message": map[string]any{
				"text": issue.Message,
			},
			"locations": []map[string]any{
				{
					"physicalLocation": map[string]any{
						"artifactLocation": map[string]any{
							"uri":       issue.File,
							"uriBaseId": "%SRCROOT%",
						},
						"region": region,
					},
				},
			},
			"partialFingerprints": map[string]any{
				"atheonLoc": fmt.Sprintf("%s|%s|%d|%d", issue.RuleID, issue.File, issue.Line, issue.Column),
			},
		}

		// Add fix information if available
		if issue.Fix != "" {
			result["properties"] = map[string]any{
				"fix": map[string]any{
					"description": issue.Fix,
				},
			}
		}

		results = append(results, result)
	}
	return results
}
