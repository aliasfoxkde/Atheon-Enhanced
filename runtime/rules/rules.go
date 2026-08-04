package rules

import (
	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// Rule represents a detection rule for scanning files.
type Rule struct {
	ID          string
	Description string
	Severity    string // critical, high, medium, low, info
	Check       func(file string, node interface{}) *diagnostics.Issue
}

// Registry manages a collection of rules.
type Registry struct {
	rules map[string]*Rule
}

// NewRegistry creates a new rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]*Rule),
	}
}

// Register adds a rule to the registry.
func (r *Registry) Register(rule *Rule) {
	r.rules[rule.ID] = rule
}

// Get retrieves a rule by its ID.
func (r *Registry) Get(id string) *Rule {
	return r.rules[id]
}

// List returns all registered rules.
func (r *Registry) List() []*Rule {
	result := make([]*Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		result = append(result, rule)
	}
	return result
}

// Filter returns rules filtered by severity level.
func (r *Registry) Filter(severity string) []*Rule {
	result := make([]*Rule, 0)
	for _, rule := range r.rules {
		if rule.Severity == severity {
			result = append(result, rule)
		}
	}
	return result
}
