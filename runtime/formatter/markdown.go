package formatter

import (
	"fmt"
	"strings"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// MarkdownFormatter implements Markdown table output for diagnostics.
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new MarkdownFormatter.
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Format outputs diagnostics in Markdown table format.
func (f *MarkdownFormatter) Format(d *diagnostics.Diagnostics) ([]byte, error) {
	if d == nil {
		return []byte("# Diagnostics\n\nNo findings.\n"), nil
	}

	var sb strings.Builder

	// Summary section
	sb.WriteString("# Diagnostics Report\n\n")

	if d.Metadata.Version != "" {
		fmt.Fprintf(&sb, "**Version:** %s\n", d.Metadata.Version)
	}
	fmt.Fprintf(&sb, "\n## Summary\n\n")
	fmt.Fprintf(&sb, "| Metric | Value |\n")
	fmt.Fprintf(&sb, "|--------|-------|\n")
	fmt.Fprintf(&sb, "| Total Findings | %d |\n", d.Summary.TotalFindings)
	fmt.Fprintf(&sb, "| Files Scanned | %d |\n", d.Statistics.FilesScanned)
	fmt.Fprintf(&sb, "| Files Analyzed | %d |\n", d.Statistics.FilesAnalyzed)
	fmt.Fprintf(&sb, "| Duration | %.2fms |\n", d.Statistics.Duration)

	if d.Summary.TotalFindings > 0 {
		sb.WriteString("\n## Severity Breakdown\n\n")
		fmt.Fprintf(&sb, "| Severity | Count |\n")
		fmt.Fprintf(&sb, "|----------|-------|\n")
		if d.Summary.Critical > 0 {
			fmt.Fprintf(&sb, "| Critical | %d |\n", d.Summary.Critical)
		}
		if d.Summary.High > 0 {
			fmt.Fprintf(&sb, "| High | %d |\n", d.Summary.High)
		}
		if d.Summary.Medium > 0 {
			fmt.Fprintf(&sb, "| Medium | %d |\n", d.Summary.Medium)
		}
		if d.Summary.Low > 0 {
			fmt.Fprintf(&sb, "| Low | %d |\n", d.Summary.Low)
		}
		if d.Summary.Info > 0 {
			fmt.Fprintf(&sb, "| Info | %d |\n", d.Summary.Info)
		}
	}

	// Issues table
	if len(d.Issues) > 0 {
		sb.WriteString("\n## Findings\n\n")
		sb.WriteString("| Rule ID | Severity | File | Line | Message |\n")
		sb.WriteString("|---------|----------|------|------|--------|\n")

		for _, issue := range d.Issues {
			severity := issue.Severity
			if severity == "" {
				severity = "unknown"
			}

			// Escape markdown special characters in message
			message := escapeMarkdown(issue.Message)
			if len(message) > 60 {
				message = message[:57] + "..."
			}

			location := issue.File
			if issue.Line > 0 {
				location = fmt.Sprintf("%s:%d", issue.File, issue.Line)
			}

			fmt.Fprintf(&sb, "| %s | %s | %s | %d | %s |\n",
				escapeMarkdown(issue.RuleID),
				severity,
				escapeMarkdown(location),
				issue.Line,
				message,
			)
		}
	} else {
		sb.WriteString("\n## Findings\n\n")
		sb.WriteString("No findings to display.\n")
	}

	// Timing section
	if d.Timing.Start > 0 || d.Timing.End > 0 {
		sb.WriteString("\n## Timing\n\n")
		fmt.Fprintf(&sb, "| Phase | Value |\n")
		fmt.Fprintf(&sb, "|-------|-------|\n")
		fmt.Fprintf(&sb, "| Start | %dms |\n", d.Timing.Start)
		fmt.Fprintf(&sb, "| End | %dms |\n", d.Timing.End)
	}

	return []byte(sb.String()), nil
}

// ContentType returns the MIME type for Markdown output.
func (f *MarkdownFormatter) ContentType() string {
	return "text/markdown"
}

// FileExtension returns the file extension for Markdown output.
func (f *MarkdownFormatter) FileExtension() string {
	return ".md"
}

// escapeMarkdown escapes special markdown characters.
func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
