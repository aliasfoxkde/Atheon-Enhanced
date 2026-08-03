package diagnostics

// Diagnostics is the canonical output format for all analysis
type Diagnostics struct {
	Summary    Summary    `json:"summary"`
	Statistics Statistics `json:"statistics"`
	Issues     []Issue    `json:"issues"`
	Timing     Timing     `json:"timing"`
	Metadata   Metadata   `json:"metadata"`
}

type Summary struct {
	TotalFindings int `json:"total_findings"`
	Critical     int `json:"critical"`
	High         int `json:"high"`
	Medium       int `json:"medium"`
	Low          int `json:"low"`
	Info         int `json:"info"`
}

type Statistics struct {
	FilesScanned    int     `json:"files_scanned"`
	FilesAnalyzed   int     `json:"files_analyzed"`
	PatternsMatched int     `json:"patterns_matched"`
	Duration        float64 `json:"duration_ms"`
}

type Issue struct {
	RuleID   string `json:"rule_id"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Fix      string `json:"fix,omitempty"`
}

type Timing struct {
	Start int64 `json:"start_ms"`
	End   int64 `json:"end_ms"`
}

type Metadata struct {
	Version    string `json:"version"`
	ConfigFile string `json:"config_file,omitempty"`
}

// NewDiagnostics creates a new Diagnostics instance with initialized collections.
func NewDiagnostics() *Diagnostics {
	return &Diagnostics{
		Issues: make([]Issue, 0),
	}
}

// AddIssue adds an issue to the diagnostics and updates the summary.
func AddIssue(d *Diagnostics, issue Issue) {
	if d == nil {
		d = NewDiagnostics()
	}

	d.Issues = append(d.Issues, issue)
	d.Summary.TotalFindings++
	d.updateSeverityCount(issue.Severity)
}

// updateSeverityCount updates the severity counts in the summary.
func (d *Diagnostics) updateSeverityCount(severity string) {
	switch severity {
	case "critical":
		d.Summary.Critical++
	case "high":
		d.Summary.High++
	case "medium":
		d.Summary.Medium++
	case "low":
		d.Summary.Low++
	case "info":
		d.Summary.Info++
	}
}

// MergeDiagnostics merges two Diagnostics instances and returns a new one.
func MergeDiagnostics(d1, d2 *Diagnostics) *Diagnostics {
	if d1 == nil && d2 == nil {
		return NewDiagnostics()
	}
	if d1 == nil {
		return d2.copy()
	}
	if d2 == nil {
		return d1.copy()
	}

	result := NewDiagnostics()

	// Merge summaries
	result.Summary = mergeSummary(d1.Summary, d2.Summary)

	// Merge statistics
	result.Statistics = mergeStatistics(d1.Statistics, d2.Statistics)

	// Merge issues
	result.Issues = make([]Issue, 0, len(d1.Issues)+len(d2.Issues))
	result.Issues = append(result.Issues, d1.Issues...)
	result.Issues = append(result.Issues, d2.Issues...)

	// Merge timing - take the earliest start and latest end times
	if d1.Timing.Start < d2.Timing.Start {
		result.Timing.Start = d1.Timing.Start
	} else {
		result.Timing.Start = d2.Timing.Start
	}
	if d1.Timing.End > d2.Timing.End {
		result.Timing.End = d1.Timing.End
	} else {
		result.Timing.End = d2.Timing.End
	}

	// Merge metadata (prefer d1's values for conflicts)
	result.Metadata = mergeMetadata(d1.Metadata, d2.Metadata)

	return result
}

// copy creates a deep copy of the Diagnostics struct.
func (d *Diagnostics) copy() *Diagnostics {
	if d == nil {
		return nil
	}

	result := &Diagnostics{
		Summary:   d.Summary,
		Statistics: d.Statistics,
		Timing:    d.Timing,
		Metadata:  d.Metadata,
	}

	// Deep copy issues
	result.Issues = make([]Issue, len(d.Issues))
	copy(result.Issues, d.Issues)

	return result
}

// mergeSummary merges two Summary structs.
func mergeSummary(s1, s2 Summary) Summary {
	return Summary{
		TotalFindings: s1.TotalFindings + s2.TotalFindings,
		Critical:      s1.Critical + s2.Critical,
		High:          s1.High + s2.High,
		Medium:        s1.Medium + s2.Medium,
		Low:           s1.Low + s2.Low,
		Info:          s1.Info + s2.Info,
	}
}

// mergeStatistics merges two Statistics structs.
func mergeStatistics(s1, s2 Statistics) Statistics {
	return Statistics{
		FilesScanned:    s1.FilesScanned + s2.FilesScanned,
		FilesAnalyzed:   s1.FilesAnalyzed + s2.FilesAnalyzed,
		PatternsMatched: s1.PatternsMatched + s2.PatternsMatched,
		Duration:        s1.Duration + s2.Duration,
	}
}

// mergeMetadata merges two Metadata structs.
func mergeMetadata(m1, m2 Metadata) Metadata {
	result := Metadata{}

	if m1.Version != "" {
		result.Version = m1.Version
	} else {
		result.Version = m2.Version
	}

	if m1.ConfigFile != "" {
		result.ConfigFile = m1.ConfigFile
	} else {
		result.ConfigFile = m2.ConfigFile
	}

	return result
}
