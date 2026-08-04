package patterns

import (
	"regexp"
)

// Built-in patterns

var (
	// CreditCardPattern matches credit card numbers (Visa, MasterCard, Amex, Discover)
	CreditCardPattern = &Pattern{
		ID:          "credit-card",
		Description: "Matches potential credit card numbers",
		Regex:       regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`),
		Severity:    "HIGH",
		Message:     "Potential credit card number detected",
	}

	// AWSKeyPattern matches AWS access key IDs
	AWSKeyPattern = &Pattern{
		ID:          "aws-key",
		Description: "Matches AWS access key IDs",
		Regex:       regexp.MustCompile(`\b(AKIA|ABIA|ACCA|ASIA)[0-9A-Z]{16}\b`),
		Severity:    "CRITICAL",
		Message:     "Potential AWS access key ID detected",
	}

	// PrivateKeyPattern matches private key headers/footers
	PrivateKeyPattern = &Pattern{
		ID:          "private-key",
		Description: "Matches private key files",
		Regex:       regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`),
		Severity:    "CRITICAL",
		Message:     "Potential private key detected",
	}

	// SQLInjectionPattern matches potential SQL injection payloads
	SQLInjectionPattern = &Pattern{
		ID:          "sql-injection",
		Description: "Matches potential SQL injection patterns",
		Regex:       regexp.MustCompile(`(?i)(?:union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|drop\s+table|update\s+.*\s+set|exec\s*\(|--|\#|\/\*.*\*\/)`),
		Severity:    "HIGH",
		Message:     "Potential SQL injection pattern detected",
	}

	// HardcodedPasswordPattern matches hardcoded password assignments
	HardcodedPasswordPattern = &Pattern{
		ID:          "hardcoded-password",
		Description: "Matches hardcoded password assignments",
		Regex:       regexp.MustCompile(`(?i)(?:password|passwd|pwd)\s*[:=]\s*["']?[^\s"']{4,}["']?`),
		Severity:    "MEDIUM",
		Message:     "Potential hardcoded password detected",
	}

	// IPAddressPattern matches IPv4 addresses
	IPAddressPattern = &Pattern{
		ID:          "ip-address",
		Description: "Matches IPv4 addresses",
		Regex:       regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`),
		Severity:    "LOW",
		Message:     "IP address detected",
	}
)

// BuiltinPatterns returns all built-in patterns as a slice
func BuiltinPatterns() []*Pattern {
	return []*Pattern{
		CreditCardPattern,
		AWSKeyPattern,
		PrivateKeyPattern,
		SQLInjectionPattern,
		HardcodedPasswordPattern,
		IPAddressPattern,
	}
}
