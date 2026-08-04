package formatter

import (
	"fmt"
	"strings"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// ConsoleFormatter implements HumanReadable output for diagnostics.
type ConsoleFormatter struct{}

// NewConsoleFormatter creates a new ConsoleFormatter.
func NewConsoleFormatter() *ConsoleFormatter {
	return &ConsoleFormatter{}
}

// Format outputs diagnostics in human-readable format.
func (f *ConsoleFormatter) Format(d *diagnostics.Diagnostics) ([]byte, error) {
	if d == nil {
		return []byte("no findings.\n"), nil
	}

	var sb strings.Builder

	if len(d.Issues) == 0 {
		sb.WriteString("no findings.\n")
	} else {
		for _, issue := range d.Issues {
			loc := issue.File
			if issue.Line > 0 {
				loc = fmt.Sprintf("%s:%d", issue.File, issue.Line)
				if issue.Column > 0 {
					loc = fmt.Sprintf("%s:%d:%d", issue.File, issue.Line, issue.Column)
				}
			}

			severityBadge := formatSeverity(issue.Severity)
			fmt.Fprintf(&sb, "[%s] %s: %s\n", severityBadge, issue.RuleID, loc)
			fmt.Fprintf(&sb, "  %s\n", issue.Message)

			if issue.Fix != "" {
				fmt.Fprintf(&sb, "  fix: %s\n", issue.Fix)
			}
		}
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "%d finding(s)\n", d.Summary.TotalFindings)
	}

	// Statistics section
	if d.Statistics.FilesScanned > 0 {
		fmt.Fprintf(&sb, "\nscanned %d file(s)\n", d.Statistics.FilesScanned)
		fmt.Fprintf(&sb, "duration: %.2fms\n", d.Statistics.Duration)
	}

	// Severity breakdown
	if d.Summary.TotalFindings > 0 {
		sb.WriteString("\nseverity breakdown:\n")
		if d.Summary.Critical > 0 {
			fmt.Fprintf(&sb, "  critical: %d\n", d.Summary.Critical)
		}
		if d.Summary.High > 0 {
			fmt.Fprintf(&sb, "  high: %d\n", d.Summary.High)
		}
		if d.Summary.Medium > 0 {
			fmt.Fprintf(&sb, "  medium: %d\n", d.Summary.Medium)
		}
		if d.Summary.Low > 0 {
			fmt.Fprintf(&sb, "  low: %d\n", d.Summary.Low)
		}
		if d.Summary.Info > 0 {
			fmt.Fprintf(&sb, "  info: %d\n", d.Summary.Info)
		}
	}

	return []byte(sb.String()), nil
}

// ContentType returns the MIME type for console output.
func (f *ConsoleFormatter) ContentType() string {
	return "text/plain"
}

// FileExtension returns the file extension for console output.
func (f *ConsoleFormatter) FileExtension() string {
	return ".txt"
}

// formatSeverity returns a colored severity badge.
func formatSeverity(sev string) string {
	switch sev {
	case "critical":
		return "CRITICAL"
	case "high":
		return "HIGH"
	case "medium":
		return "MEDIUM"
	case "low":
		return "LOW"
	case "info":
		return "INFO"
	default:
		return strings.ToUpper(sev)
	}
}
