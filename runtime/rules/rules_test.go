package rules

import (
	"testing"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.rules == nil {
		t.Error("Registry rules map is nil")
	}
}

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	rule := &Rule{
		ID:          "TEST001",
		Description: "Test rule",
		Severity:    "low",
		Check:       func(file string, node interface{}) *diagnostics.Issue { return nil },
	}

	r.Register(rule)

	if r.Get("TEST001") == nil {
		t.Error("Registered rule not found")
	}
	if r.Get("TEST001").ID != "TEST001" {
		t.Error("Retrieved rule has wrong ID")
	}
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()

	if r.Get("NONEXISTENT") != nil {
		t.Error("Get returned non-nil for nonexistent rule")
	}

	rule := &Rule{
		ID:          "TEST002",
		Description: "Test rule",
		Severity:    "info",
		Check:       func(file string, node interface{}) *diagnostics.Issue { return nil },
	}
	r.Register(rule)

	retrieved := r.Get("TEST002")
	if retrieved == nil {
		t.Fatal("Get returned nil for existing rule")
	}
	if retrieved.Description != "Test rule" {
		t.Error("Retrieved rule has wrong description")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()

	list := r.List()
	if len(list) != 0 {
		t.Error("List should be empty for new registry")
	}

	r.Register(&Rule{ID: "RULE1", Description: "Rule 1", Severity: "low", Check: func(file string, node interface{}) *diagnostics.Issue { return nil }})
	r.Register(&Rule{ID: "RULE2", Description: "Rule 2", Severity: "high", Check: func(file string, node interface{}) *diagnostics.Issue { return nil }})

	list = r.List()
	if len(list) != 2 {
		t.Errorf("List length = %d, want 2", len(list))
	}
}

func TestRegistryFilter(t *testing.T) {
	r := NewRegistry()

	r.Register(&Rule{ID: "RULE1", Description: "Rule 1", Severity: "critical", Check: func(file string, node interface{}) *diagnostics.Issue { return nil }})
	r.Register(&Rule{ID: "RULE2", Description: "Rule 2", Severity: "high", Check: func(file string, node interface{}) *diagnostics.Issue { return nil }})
	r.Register(&Rule{ID: "RULE3", Description: "Rule 3", Severity: "critical", Check: func(file string, node interface{}) *diagnostics.Issue { return nil }})

	critical := r.Filter("critical")
	if len(critical) != 2 {
		t.Errorf("Filter(critical) length = %d, want 2", len(critical))
	}

	high := r.Filter("high")
	if len(high) != 1 {
		t.Errorf("Filter(high) length = %d, want 1", len(high))
	}

	low := r.Filter("low")
	if len(low) != 0 {
		t.Errorf("Filter(low) length = %d, want 0", len(low))
	}
}

func TestBuiltinRules(t *testing.T) {
	rules := BuiltinRules()
	if len(rules) != 5 {
		t.Errorf("BuiltinRules count = %d, want 5", len(rules))
	}

	expectedIDs := []string{"G401", "G402", "G403", "G404", "G405"}
	for _, id := range expectedIDs {
		found := false
		for _, rule := range rules {
			if rule.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected rule %s not found in BuiltinRules", id)
		}
	}
}

func TestRegisterBuiltinRules(t *testing.T) {
	r := NewRegistry()
	RegisterBuiltinRules(r)

	list := r.List()
	if len(list) != 5 {
		t.Errorf("After RegisterBuiltinRules, list length = %d, want 5", len(list))
	}

	// Check that we can retrieve each built-in rule
	for _, rule := range BuiltinRules() {
		retrieved := r.Get(rule.ID)
		if retrieved == nil {
			t.Errorf("Built-in rule %s not registered", rule.ID)
		}
	}
}

func TestRuleG401CreditCard(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantIssue bool
	}{
		{"Visa card", "4111111111111111", true},
		{"MasterCard", "5500000000000004", true},
		{"Amex", "340000000000009", true},
		{"Discover", "6011000000000004", true},
		{"Clean content", "This is just some normal text without any credit card numbers", false},
		{"Partial number", "4111", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := RuleG401.Check("test.go", tt.content)
			if tt.wantIssue && issue == nil {
				t.Error("Expected issue, got nil")
			}
			if !tt.wantIssue && issue != nil {
				t.Errorf("Expected no issue, got %+v", issue)
			}
		})
	}
}

func TestRuleG402AWSKey(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantIssue bool
	}{
		{"AKIA access key", "AKIAIOSFODNN7EXAMPLE", true},
		{"AWS env var style", "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE", true},
		{"AWS secret style", "aws_secret_access_key=example-secret-key", true},
		{"Clean content", "This is normal configuration code", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := RuleG402.Check("config.py", tt.content)
			if tt.wantIssue && issue == nil {
				t.Error("Expected issue, got nil")
			}
			if !tt.wantIssue && issue != nil {
				t.Errorf("Expected no issue, got %+v", issue)
			}
		})
	}
}

func TestRuleG403PrivateKey(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantIssue bool
	}{
		{"RSA private key", "-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJB\n-----END RSA PRIVATE KEY-----", true},
		{"EC private key", "-----BEGIN EC PRIVATE KEY-----\nMHQCAQEE\n-----END EC PRIVATE KEY-----", true},
		{"DSA private key", "-----BEGIN DSA PRIVATE KEY-----\nMIIBQg\n-----END DSA PRIVATE KEY-----", true},
		{"OpenSSH key", "-----BEGIN OPENSSH PRIVATE KEY-----\n-----END OPENSSH PRIVATE KEY-----", true},
		{"Clean content", "This is just normal text content", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := RuleG403.Check("key.pem", tt.content)
			if tt.wantIssue && issue == nil {
				t.Error("Expected issue, got nil")
			}
			if !tt.wantIssue && issue != nil {
				t.Errorf("Expected no issue, got %+v", issue)
			}
		})
	}
}

func TestRuleG404Password(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantIssue bool
	}{
		{"password assignment", `password = "mysecretpass"`, true},
		{"passwd assignment", `passwd = "mysecretpass"`, true},
		{"secret assignment", `secret = "mysecretpass"`, true},
		{"api_key assignment", `api_key = "mysecretpass"`, true},
		{"apikey assignment", `apikey = "mysecretpass"`, true},
		{"Clean content", "This is normal text content", false},
		{"Too short", "password = \"abc\"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := RuleG404.Check("config.py", tt.content)
			if tt.wantIssue && issue == nil {
				t.Error("Expected issue, got nil")
			}
			if !tt.wantIssue && issue != nil {
				t.Errorf("Expected no issue, got %+v", issue)
			}
		})
	}
}

func TestRuleG405SQLInjection(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantIssue bool
	}{
		{"SELECT concatenation", "SELECT * FROM users WHERE id = '" + "' + userInput", true},
		{"INSERT concatenation", "INSERT INTO users VALUES ('" + "' + name)", true},
		{"UPDATE concatenation", "UPDATE users SET password = '" + "' + pass", true},
		{"exec with concatenation", "exec(\"SELECT * FROM \" + tableName)", true},
		{"execute with concatenation", "execute(\"SELECT * FROM \" + table)", true},
		{"query with concatenation", "query(\"SELECT\" +)", true},
		{"Clean content", "SELECT * FROM users WHERE id = ?", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := RuleG405.Check("query.go", tt.content)
			if tt.wantIssue && issue == nil {
				t.Error("Expected issue, got nil")
			}
			if !tt.wantIssue && issue != nil {
				t.Errorf("Expected no issue, got %+v", issue)
			}
		})
	}
}

func TestRuleCheckWithNonStringNode(t *testing.T) {
	// All rules should handle non-string nodes gracefully
	rules := []*Rule{RuleG401, RuleG402, RuleG403, RuleG404, RuleG405}

	for _, rule := range rules {
		issue := rule.Check("test.go", 12345)       // integer node
		if issue != nil {
			t.Errorf("Rule %s should return nil for integer node, got %+v", rule.ID, issue)
		}

		issue = rule.Check("test.go", nil)           // nil node
		if issue != nil {
			t.Errorf("Rule %s should return nil for nil node, got %+v", rule.ID, issue)
		}

		issue = rule.Check("test.go", []byte("test")) // byte slice node
		if issue != nil {
			t.Errorf("Rule %s should return nil for []byte node, got %+v", rule.ID, issue)
		}
	}
}

func TestRuleSeverityLevels(t *testing.T) {
	expected := map[string]string{
		"G401": "critical",
		"G402": "critical",
		"G403": "critical",
		"G404": "high",
		"G405": "high",
	}

	r := NewRegistry()
	RegisterBuiltinRules(r)

	for id, expectedSev := range expected {
		rule := r.Get(id)
		if rule == nil {
			t.Errorf("Rule %s not found in registry", id)
			continue
		}
		if rule.Severity != expectedSev {
			t.Errorf("Rule %s severity = %s, want %s", id, rule.Severity, expectedSev)
		}
	}
}
