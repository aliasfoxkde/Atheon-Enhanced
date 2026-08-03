package patterns

import (
	"regexp"
	"strings"
)

// Pattern represents a pattern for matching content.
type Pattern struct {
	ID          string
	Description string
	Regex       *regexp.Regexp
	Severity    string
	Message     string
}

// Matcher matches content against a set of patterns.
type Matcher struct {
	patterns []*Pattern
}

// NewMatcher creates a new Matcher with the given patterns.
func NewMatcher(patterns []*Pattern) *Matcher {
	return &Matcher{patterns: patterns}
}

// Match checks the content against all patterns and returns all matches.
func (m *Matcher) Match(content string) []*Match {
	var matches []*Match
	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		lineMatches := m.MatchLine(line, lineNum+1)
		matches = append(matches, lineMatches...)
	}
	return matches
}

// MatchLine checks a single line against all patterns.
func (m *Matcher) MatchLine(line string, lineNum int) []*Match {
	var matches []*Match
	for _, p := range m.patterns {
		indices := p.Regex.FindAllStringIndex(line, -1)
		for _, idx := range indices {
			matches = append(matches, &Match{
				PatternID: p.ID,
				Line:      lineNum,
				Column:    idx[0] + 1,
				Length:    idx[1] - idx[0],
				Message:   p.Message,
			})
		}
	}
	return matches
}

// Match represents a single pattern match.
type Match struct {
	PatternID string
	Line      int
	Column    int
	Length    int
	Message   string
}
