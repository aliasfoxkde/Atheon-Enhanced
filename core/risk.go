package core

import (
	"fmt"
	"math"
	"sort"
)

// Severity weights for risk calculation
const (
	RiskSeverityCritical = 40
	RiskSeverityHigh     = 20
	RiskSeverityMedium   = 10
	RiskSeverityLow      = 5
)

// RiskLevel represents the overall risk assessment.
type RiskLevel string

const (
	RiskLevelCritical RiskLevel = "critical"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelLow      RiskLevel = "low"
	RiskLevelNone     RiskLevel = "none"
)

// RiskScore represents a computed risk assessment for a set of findings.
type RiskScore struct {
	// Overall score from 0-100
	Score int `json:"score"`
	// Risk level classification
	Level RiskLevel `json:"level"`
	// Breakdown by category
	ByCategory map[string]CategoryRisk `json:"by_category"`
	// Total findings count
	FindingCount int `json:"finding_count"`
	// Highest severity finding
	HighestSeverity string `json:"highest_severity"`
}

// CategoryRisk represents risk breakdown for a single category.
type CategoryRisk struct {
	Score      int      `json:"score"`
	Count      int      `json:"count"`
	Findings   []string `json:"findings"`
	Severities []string `json:"severities"`
}

// NewRiskScore creates a new risk score from a list of findings.
func NewRiskScore(findings []Finding) *RiskScore {
	if len(findings) == 0 {
		return &RiskScore{
			Score:      0,
			Level:      RiskLevelNone,
			ByCategory: make(map[string]CategoryRisk),
		}
	}

	rs := &RiskScore{
		ByCategory:      make(map[string]CategoryRisk),
		FindingCount:    len(findings),
		HighestSeverity: "low",
	}

	// Calculate severity weights and categorize
	maxSeverity := 0
	for _, f := range findings {
		weight := severityWeight(f.Severity)
		if weight > maxSeverity {
			maxSeverity = weight
			rs.HighestSeverity = f.Severity
		}

		// Accumulate by category
		cat, ok := rs.ByCategory[f.Category]
		if !ok {
			cat = CategoryRisk{}
		}
		cat.Score += weight
		cat.Count++
		cat.Findings = append(cat.Findings, f.Pattern)
		cat.Severities = append(cat.Severities, f.Severity)
		rs.ByCategory[f.Category] = cat
	}

	// Normalize scores
	totalScore := 0
	for cat, cr := range rs.ByCategory {
		// Cap category score at 100
		if cr.Score > 100 {
			cr.Score = 100
		}
		rs.ByCategory[cat] = cr
		totalScore += cr.Score
	}

	// Calculate overall score (weighted average, capped at 100)
	rs.Score = int(math.Min(100, float64(totalScore)*float64(len(findings))/float64(len(rs.ByCategory)+1)))
	if rs.Score > 100 {
		rs.Score = 100
	}

	// Determine risk level
	rs.Level = ClassifyRisk(rs.Score)

	return rs
}

// severityWeight returns the numeric weight for a severity string.
func severityWeight(severity string) int {
	switch severity {
	case "critical":
		return RiskSeverityCritical
	case "high":
		return RiskSeverityHigh
	case "medium":
		return RiskSeverityMedium
	case "low":
		return RiskSeverityLow
	default:
		return 1
	}
}

// ClassifyRisk converts a numeric score to a risk level.
func ClassifyRisk(score int) RiskLevel {
	switch {
	case score >= 80:
		return RiskLevelCritical
	case score >= 50:
		return RiskLevelHigh
	case score >= 25:
		return RiskLevelMedium
	case score > 0:
		return RiskLevelLow
	default:
		return RiskLevelNone
	}
}

// TopRisks returns the categories sorted by risk score descending.
func (rs *RiskScore) TopRisks(limit int) []struct {
	Category string
	Risk     CategoryRisk
} {
	type kv struct {
		category string
		risk     CategoryRisk
	}
	list := make([]kv, 0, len(rs.ByCategory))
	for cat, risk := range rs.ByCategory {
		list = append(list, kv{cat, risk})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].risk.Score > list[j].risk.Score
	})

	result := make([]struct {
		Category string
		Risk     CategoryRisk
	}, 0, limit)
	for i := 0; i < limit && i < len(list); i++ {
		result = append(result, struct {
			Category string
			Risk     CategoryRisk
		}{list[i].category, list[i].risk})
	}
	return result
}

// AddFinding adds a new finding and recalculates risk.
func (rs *RiskScore) AddFinding(f Finding) {
	weight := severityWeight(f.Severity)
	rs.FindingCount++

	if weight > severityWeight(rs.HighestSeverity) {
		rs.HighestSeverity = f.Severity
	}

	cat, ok := rs.ByCategory[f.Category]
	if !ok {
		cat = CategoryRisk{}
	}
	cat.Score += weight
	cat.Count++
	cat.Findings = append(cat.Findings, f.Pattern)
	cat.Severities = append(cat.Severities, f.Severity)
	rs.ByCategory[f.Category] = cat

	// Recalculate overall
	totalScore := 0
	for _, cr := range rs.ByCategory {
		totalScore += cr.Score
	}
	rs.Score = int(math.Min(100, float64(totalScore)))
	rs.Level = ClassifyRisk(rs.Score)
}

// Summary returns a human-readable risk summary.
func (rs *RiskScore) Summary() string {
	return fmt.Sprintf("Risk: %s (score: %d/100) - %d findings across %d categories",
		rs.Level, rs.Score, rs.FindingCount, len(rs.ByCategory))
}
