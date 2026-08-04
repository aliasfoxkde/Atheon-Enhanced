package rules

import (
	"regexp"
	"strings"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

var (
	// RuleG401 detects credit card numbers
	RuleG401 = &Rule{
		ID:          "G401",
		Description: "Credit card number detected",
		Severity:    "critical",
		Check:       detectCreditCard,
	}

	// RuleG402 detects AWS access keys
	RuleG402 = &Rule{
		ID:          "G402",
		Description: "AWS access key detected",
		Severity:    "critical",
		Check:       detectAWSKey,
	}

	// RuleG403 detects private keys
	RuleG403 = &Rule{
		ID:          "G403",
		Description: "Private key detected",
		Severity:    "critical",
		Check:       detectPrivateKey,
	}

	// RuleG404 detects hardcoded passwords
	RuleG404 = &Rule{
		ID:          "G404",
		Description: "Hardcoded password detected",
		Severity:    "high",
		Check:       detectHardcodedPassword,
	}

	// RuleG405 detects SQL injection risks
	RuleG405 = &Rule{
		ID:          "G405",
		Description: "SQL injection risk",
		Severity:    "high",
		Check:       detectSQLInjection,
	}
)

// BuiltinRules returns all built-in rules
func BuiltinRules() []*Rule {
	return []*Rule{
		RuleG401,
		RuleG402,
		RuleG403,
		RuleG404,
		RuleG405,
	}
}

// RegisterBuiltinRules registers all built-in rules to the given registry
func RegisterBuiltinRules(r *Registry) {
	for _, rule := range BuiltinRules() {
		r.Register(rule)
	}
}

// detectCreditCard detects credit card numbers in file content
func detectCreditCard(file string, node interface{}) *diagnostics.Issue {
	content, ok := node.(string)
	if !ok {
		return nil
	}

	// Common credit card patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\b4[0-9]{12}(?:[0-9]{3})?\b`),                                   // Visa
		regexp.MustCompile(`\b5[1-5][0-9]{14}\b`),                                           // MasterCard
		regexp.MustCompile(`\b3[47][0-9]{13}\b`),                                            // American Express
		regexp.MustCompile(`\b6(?:011|5[0-9]{2})[0-9]{12}\b`),                               // Discover
		regexp.MustCompile(`\b(?:3[0-9]{4}|4[0-9]{4}|5[0-9]{4})[0-9]{4}[0-9]{4}[0-9]{4}\b`), // Generic
	}

	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return &diagnostics.Issue{
				RuleID:   "G401",
				Message:  "Credit card number detected",
				Severity: "critical",
				File:     file,
			}
		}
	}

	return nil
}

// detectAWSKey detects AWS access keys in file content
func detectAWSKey(file string, node interface{}) *diagnostics.Issue {
	content, ok := node.(string)
	if !ok {
		return nil
	}

	// AWS access key patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),                  // Access Key ID
		regexp.MustCompile(`(?i)AWS_ACCESS_KEY_ID\s*=\s*[^ ]+`),     // Environment variable style
		regexp.MustCompile(`(?i)aws_secret_access_key\s*=\s*[^ ]+`), // Secret key style
	}

	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return &diagnostics.Issue{
				RuleID:   "G402",
				Message:  "AWS access key detected",
				Severity: "critical",
				File:     file,
			}
		}
	}

	return nil
}

// detectPrivateKey detects private keys in file content
func detectPrivateKey(file string, node interface{}) *diagnostics.Issue {
	content, ok := node.(string)
	if !ok {
		return nil
	}

	// Private key patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----`),
		regexp.MustCompile(`-----BEGIN\s+EC\s+PRIVATE\s+KEY-----`),
		regexp.MustCompile(`-----BEGIN\s+DSA\s+PRIVATE\s+KEY-----`),
		regexp.MustCompile(`-----BEGIN\s+OPENSSH\s+PRIVATE\s+KEY-----`),
		regexp.MustCompile(`-----BEGIN\s+PGP\s+PRIVATE\s+KEY-----`),
	}

	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return &diagnostics.Issue{
				RuleID:   "G403",
				Message:  "Private key detected",
				Severity: "critical",
				File:     file,
			}
		}
	}

	return nil
}

// detectHardcodedPassword detects hardcoded passwords in file content
func detectHardcodedPassword(file string, node interface{}) *diagnostics.Issue {
	content, ok := node.(string)
	if !ok {
		return nil
	}

	// Password patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)password\s*=\s*["'][^"']{4,}["']`),
		regexp.MustCompile(`(?i)passwd\s*=\s*["'][^"']{4,}["']`),
		regexp.MustCompile(`(?i)secret\s*=\s*["'][^"']{4,}["']`),
		regexp.MustCompile(`(?i)api[_-]?key\s*=\s*["'][^"']{4,}["']`),
		regexp.MustCompile(`(?i)apikey\s*=\s*["'][^"']{4,}["']`),
	}

	for _, pattern := range patterns {
		if match := pattern.FindString(content); match != "" {
			return &diagnostics.Issue{
				RuleID:   "G404",
				Message:  "Hardcoded password detected",
				Severity: "high",
				File:     file,
			}
		}
	}

	return nil
}

// detectSQLInjection detects potential SQL injection risks
func detectSQLInjection(file string, node interface{}) *diagnostics.Issue {
	content, ok := node.(string)
	if !ok {
		return nil
	}

	// SQL injection patterns - looking for string concatenation in SQL queries
	sqlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)SELECT\s+.*\+\s+.*FROM`),
		regexp.MustCompile(`(?i)INSERT\s+.*\+\s+.*INTO`),
		regexp.MustCompile(`(?i)UPDATE\s+.*\+\s+.*SET`),
		regexp.MustCompile(`(?i)DELETE\s+.*\+\s+.*FROM`),
		regexp.MustCompile(`(?i)exec\s*\(\s*.*\+`),
		regexp.MustCompile(`(?i)execute\s*\(\s*.*\+`),
		regexp.MustCompile(`(?i)query\s*\(.*\+\s*\)`),
	}

	// Dangerous function patterns that might indicate SQL injection
	dangerousFuncs := []string{
		"string concatenation",
		"format string",
		"sprintf",
		"string.Format",
	}

	for _, pattern := range sqlPatterns {
		if pattern.MatchString(content) {
			return &diagnostics.Issue{
				RuleID:   "G405",
				Message:  "SQL injection risk",
				Severity: "high",
				File:     file,
			}
		}
	}

	// Check for common dangerous patterns in SQL context
	lowerContent := strings.ToLower(content)
	dangerousIndicators := []string{"','", "' +"}
	for _, indicator := range dangerousIndicators {
		if strings.Contains(lowerContent, indicator) {
			return &diagnostics.Issue{
				RuleID:   "G405",
				Message:  "SQL injection risk",
				Severity: "high",
				File:     file,
			}
		}
	}

	_ = dangerousFuncs // suppress unused variable warning

	return nil
}
