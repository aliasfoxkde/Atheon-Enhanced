package patterns

import (
	"regexp"
	"testing"
)

func TestNewMatcher(t *testing.T) {
	patterns := []*Pattern{
		{
			ID:          "test",
			Description: "Test pattern",
			Regex:       regexp.MustCompile(`test`),
			Severity:    "LOW",
			Message:     "Test message",
		},
	}
	matcher := NewMatcher(patterns)
	if matcher == nil {
		t.Fatal("NewMatcher returned nil")
	}
	if len(matcher.patterns) != 1 {
		t.Errorf("expected 1 pattern, got %d", len(matcher.patterns))
	}
}

func TestMatchLine(t *testing.T) {
	pattern := &Pattern{
		ID:          "test",
		Description: "Test pattern",
		Regex:       regexp.MustCompile(`password`),
		Severity:    "MEDIUM",
		Message:     "Password detected",
	}
	matcher := NewMatcher([]*Pattern{pattern})

	tests := []struct {
		name     string
		line     string
		lineNum  int
		expected int
	}{
		{"match", "password=secret", 1, 1},
		{"no match", "username=admin", 2, 0},
		{"multiple matches", "password=pwd1 password=pwd2", 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, tt.lineNum)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatchLine_CreditCard(t *testing.T) {
	matcher := NewMatcher([]*Pattern{CreditCardPattern})

	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"Visa", "Card: 4111111111111111", 1},
		{"MasterCard", "Card: 5500000000000004", 1},
		{"Amex", "Card: 340000000000009", 1},
		{"No match", "Card: 1234567890", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, 1)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatchLine_AWSKey(t *testing.T) {
	matcher := NewMatcher([]*Pattern{AWSKeyPattern})

	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"AKIA key", "AWS_KEY=AKIAIOSFODNN7EXAMPLE", 1},
		{"No match - invalid format", "AWS_KEY=NOTAVALIDKEY123", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, 1)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatchLine_PrivateKey(t *testing.T) {
	matcher := NewMatcher([]*Pattern{PrivateKeyPattern})

	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"RSA key", "-----BEGIN RSA PRIVATE KEY-----", 1},
		{"EC key", "-----BEGIN EC PRIVATE KEY-----", 1},
		{"OpenSSH key", "-----BEGIN OPENSSH PRIVATE KEY-----", 1},
		{"No match", "-----BEGIN PUBLIC KEY-----", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, 1)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatchLine_SQLInjection(t *testing.T) {
	matcher := NewMatcher([]*Pattern{SQLInjectionPattern})

	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"UNION SELECT", "SELECT * FROM users WHERE id=1 UNION SELECT * FROM passwords", 1},
		{"INSERT INTO", "INSERT INTO users VALUES ('admin', 'password')", 1},
		{"DROP TABLE", "DROP TABLE users", 1},
		{"SQL comment only", "some code -- comment here", 1},
		{"No match", "SELECT username", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, 1)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatchLine_HardcodedPassword(t *testing.T) {
	matcher := NewMatcher([]*Pattern{HardcodedPasswordPattern})

	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"password=", "password=secret123", 1},
		{"password:", "password: mypassword", 1},
		{"pwd=", "pwd=pass", 1},
		{"passwd:", "passwd: secret", 1},
		{"No match - too short", "password=abc", 0},
		{"No match - no assignment", "username=admin", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, 1)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatchLine_IPAddress(t *testing.T) {
	matcher := NewMatcher([]*Pattern{IPAddressPattern})

	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"valid IP", "Server: 192.168.1.1", 1},
		{"multiple IPs", "From: 10.0.0.1 to 10.0.0.255", 2},
		{"No match", "IP: 999.999.999.999", 0},
		{"No match - partial", "IP: 192.168.1", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matcher.MatchLine(tt.line, 1)
			if len(matches) != tt.expected {
				t.Errorf("expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestMatch_MultipleLines(t *testing.T) {
	pattern := &Pattern{
		ID:          "test",
		Description: "Test pattern",
		Regex:       regexp.MustCompile(`error`),
		Severity:    "HIGH",
		Message:     "Error detected",
	}
	matcher := NewMatcher([]*Pattern{pattern})

	content := "line one\nline two error\nline three\nanother error here"
	matches := matcher.Match(content)

	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}

	if matches[0].Line != 2 {
		t.Errorf("expected line 2, got %d", matches[0].Line)
	}
	if matches[1].Line != 4 {
		t.Errorf("expected line 4, got %d", matches[1].Line)
	}
}

func TestMatch_MatchDetails(t *testing.T) {
	pattern := &Pattern{
		ID:          "secret",
		Description: "Secret pattern",
		Regex:       regexp.MustCompile(`secret`),
		Severity:    "HIGH",
		Message:     "Secret found!",
	}
	matcher := NewMatcher([]*Pattern{pattern})

	matches := matcher.MatchLine("my secret key", 5)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}

	m := matches[0]
	if m.PatternID != "secret" {
		t.Errorf("expected PatternID 'secret', got '%s'", m.PatternID)
	}
	if m.Line != 5 {
		t.Errorf("expected Line 5, got %d", m.Line)
	}
	if m.Column != 4 {
		t.Errorf("expected Column 4, got %d", m.Column)
	}
	if m.Length != 6 {
		t.Errorf("expected Length 6, got %d", m.Length)
	}
	if m.Message != "Secret found!" {
		t.Errorf("expected Message 'Secret found!', got '%s'", m.Message)
	}
}

func TestBuiltinPatterns(t *testing.T) {
	patterns := BuiltinPatterns()
	if len(patterns) != 6 {
		t.Errorf("expected 6 built-in patterns, got %d", len(patterns))
	}

	// Verify each pattern has required fields
	for _, p := range patterns {
		if p.ID == "" {
			t.Error("pattern ID should not be empty")
		}
		if p.Regex == nil {
			t.Error("pattern Regex should not be nil")
		}
		if p.Severity == "" {
			t.Error("pattern Severity should not be empty")
		}
	}
}
