package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// YARARule represents a compiled YARA rule for scanning.
type YARARule struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Patterns    []string `json:"patterns"`
	Severity    string   `json:"severity"`
	Category    string   `json:"category"`
}

// YARAScanner provides YARA-like pattern matching capabilities.
// Note: This is a simplified implementation. For full YARA support,
// the go-yara library can be integrated.
type YARAScanner struct {
	rulesDir string
	rules    map[string]*YARARule
}

// NewYARAScanner creates a new YARA scanner.
func NewYARAScanner(rulesDir string) *YARAScanner {
	return &YARAScanner{
		rulesDir: rulesDir,
		rules:    make(map[string]*YARARule),
	}
}

// YARAFinding represents a finding from YARA-like pattern matching.
type YARAFinding struct {
	Rule        string `json:"rule"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Matched     string `json:"matched"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// DefaultYARARules returns the default built-in rules.
func DefaultYARARules() []*YARARule {
	return []*YARARule{
		{
			Name:        "malware_indicator",
			Description: "Detects potential malware indicators",
			Tags:        []string{"malware", "threat"},
			Patterns:    []string{"eval(base64_decode(", "exec(base64_decode(", "__import__(\"os\")", "os.system("},
			Severity:    "critical",
			Category:    "security",
		},
		{
			Name:        "crypto_miner",
			Description: "Detects cryptocurrency mining patterns",
			Tags:        []string{"crypto", "miner", "malware"},
			Patterns:    []string{"stratum+tcp://", ".wallet.", "cryptonight", "mining_pool"},
			Severity:    "critical",
			Category:    "security",
		},
		{
			Name:        "suspicious_script",
			Description: "Detects suspicious script patterns",
			Tags:        []string{"script", "suspicious"},
			Patterns:    []string{"curl.*|.*bash", "wget.*|.*sh", "rm -rf /"},
			Severity:    "high",
			Category:    "security",
		},
		{
			Name:        "credential_theft",
			Description: "Detects patterns associated with credential theft",
			Tags:        []string{"credentials", "theft"},
			Patterns:    []string{"keylogger", "password.*clipboard", "\\.get\\(\"password\"\\)", "keystroke"},
			Severity:    "critical",
			Category:    "secrets",
		},
		{
			Name:        "network_recon",
			Description: "Detects network reconnaissance patterns",
			Tags:        []string{"network", "recon"},
			Patterns:    []string{"nmap", "netstat", "ping sweep", "port scan"},
			Severity:    "medium",
			Category:    "security",
		},
		{
			Name:        "persistence_indicator",
			Description: "Detects system persistence mechanisms",
			Tags:        []string{"persistence", "registry"},
			Patterns:    []string{"HKLM\\\\Software", "HKCU\\\\Software", "~/.bashrc", "crontab"},
			Severity:    "high",
			Category:    "security",
		},
		{
			Name:        "data_exfiltration",
			Description: "Detects potential data exfiltration patterns",
			Tags:        []string{"exfil", "data"},
			Patterns:    []string{"Authorization: Bearer", "password=", "api_key=", ".send_keys("},
			Severity:    "high",
			Category:    "security",
		},
		{
			Name:        "reverse_shell",
			Description: "Detects reverse shell patterns",
			Tags:        []string{"shell", "backdoor"},
			Patterns:    []string{"/bin/bash -i", "/dev/tcp/", "nc -e ", "bash -i >& /dev/tcp/"},
			Severity:    "critical",
			Category:    "security",
		},
	}
}

// ScanFile scans a file using YARA-like rules.
func (ys *YARAScanner) ScanFile(filename string) ([]YARAFinding, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var findings []YARAFinding

	// Load default rules if none loaded
	rules := ys.rules
	if len(rules) == 0 {
		for _, rule := range DefaultYARARules() {
			rules[rule.Name] = rule
		}
	}

	// Match patterns against file content
	for _, rule := range rules {
		for lineNum, line := range lines {
			for _, pattern := range rule.Patterns {
				if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
					findings = append(findings, YARAFinding{
						Rule:        rule.Name,
						File:        filename,
						Line:        lineNum + 1,
						Matched:     strings.TrimSpace(line),
						Severity:    rule.Severity,
						Category:    rule.Category,
						Description: rule.Description,
					})
					break // Only report once per rule per line
				}
			}
		}
	}

	return findings, nil
}

// ScanDir scans all files in a directory using YARA-like rules.
func (ys *YARAScanner) ScanDir(dir string) ([]YARAFinding, error) {
	var allFindings []YARAFinding

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and binary files
		if info.IsDir() {
			return nil
		}

		// Skip common non-text directories
		skipDirs := []string{".git", "node_modules", "vendor", "target", ".venv"}
		for _, skip := range skipDirs {
			if strings.Contains(path, skip) {
				return nil
			}
		}

		findings, err := ys.ScanFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		allFindings = append(allFindings, findings...)
		return nil
	})

	return allFindings, err
}

// AddRule adds a custom rule to the scanner.
func (ys *YARAScanner) AddRule(rule *YARARule) {
	ys.rules[rule.Name] = rule
}

// ConvertToFinding converts a YARA finding to a core Finding.
func ConvertToFinding(yf YARAFinding) Finding {
	return Finding{
		Pattern:     "yara_" + yf.Rule,
		File:        yf.File,
		Line:        yf.Line,
		Content:     yf.Matched,
		Severity:    yf.Severity,
		Category:    yf.Category,
		Fingerprint: fmt.Sprintf("%s|%s|%d", yf.Rule, yf.File, yf.Line),
	}
}
