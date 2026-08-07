package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aliasfoxkde/Atheon/core"
)

// ReportFormat defines the output format for reports
type ReportFormat string

const (
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatXLSX ReportFormat = "xlsx"
	ReportFormatSummary ReportFormat = "summary"
)

// ReportOptions contains options for report generation
type ReportOptions struct {
	Format         ReportFormat
	Output         string // output file path, empty = stdout
	Stats          *core.Stats
	Recursive      bool   // recursive directory scan
	FilterPatterns []string // only include these patterns
	FilterExts     []string // only include these extensions
	IncludeNetwork bool   // include network/UNC path info
}

// GenerateReport generates a report in the specified format
func GenerateReport(findings []core.Finding, stats *core.Stats, opts ReportOptions) error {
	// Apply filters
	filteredFindings := filterFindings(findings, opts)

	var data []byte
	var err error

	switch opts.Format {
	case ReportFormatCSV:
		data, err = generateCSV(filteredFindings, stats, opts.IncludeNetwork)
	case ReportFormatJSON:
		data, err = generateJSON(filteredFindings, stats, opts)
	case ReportFormatXLSX:
		data, err = generateXLSX(filteredFindings, stats)
	case ReportFormatSummary:
		data, err = generateSummary(filteredFindings, stats, opts)
	default:
		return fmt.Errorf("unknown report format: %s", opts.Format)
	}

	if err != nil {
		return err
	}

	if opts.Output == "" {
		fmt.Print(string(data))
	} else {
		if err := os.WriteFile(opts.Output, data, 0644); err != nil {
			return fmt.Errorf("writing report: %w", err)
		}
		fmt.Printf("Report written to: %s\n", opts.Output)
	}

	return nil
}

// filterFindings applies pattern and extension filters to findings
func filterFindings(findings []core.Finding, opts ReportOptions) []core.Finding {
	if len(opts.FilterPatterns) == 0 && len(opts.FilterExts) == 0 {
		return findings
	}

	filtered := make([]core.Finding, 0)
	for _, f := range findings {
		// Pattern filter
		if len(opts.FilterPatterns) > 0 {
			found := false
			for _, p := range opts.FilterPatterns {
				if strings.Contains(strings.ToLower(f.Pattern), strings.ToLower(p)) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Extension filter
		if len(opts.FilterExts) > 0 {
			ext := strings.ToLower(filepath.Ext(f.File))
			found := false
			for _, e := range opts.FilterExts {
				if ext == "."+strings.TrimPrefix(e, ".") {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, f)
	}
	return filtered
}

// generateCSV creates a CSV report with extension analysis
func generateCSV(findings []core.Finding, stats *core.Stats, includeNetwork bool) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Dynamic header based on includeNetwork
	header := []string{"Pattern", "Severity", "Category", "File", "Extension", "Line", "Column", "Match", "Description"}
	if includeNetwork {
		header = append(header, "PathType", "Directory")
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Sort findings
	sortedFindings := sortFindings(findings)

	// Data rows
	for _, f := range sortedFindings {
		ext := strings.ToLower(filepath.Ext(f.File))
		row := []string{
			f.Pattern,
			f.Severity,
			f.Category,
			f.File,
			ext,
			strconv.Itoa(f.Line),
			strconv.Itoa(f.Column),
			redact(f.Content),
			f.Description,
		}
		if includeNetwork {
			pathType := getPathType(f.File)
			dir := filepath.Dir(f.File)
			row = append(row, pathType, dir)
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(buf.String()), nil
}

// generateJSON creates a detailed JSON report with full analysis
func generateJSON(findings []core.Finding, stats *core.Stats, opts ReportOptions) ([]byte, error) {
	sortedFindings := sortFindings(findings)

	// Build report structure
	report := map[string]any{
		"generated":     time.Now().UTC().Format(time.RFC3339),
		"version":       version,
		"total":         len(findings),
		"recursive":     opts.Recursive,
		"filter_exts":   opts.FilterExts,
		"filter_pats":   opts.FilterPatterns,
		"findings":       sortedFindings,
	}

	if stats != nil {
		report["stats"] = map[string]any{
			"files":      stats.Files,
			"bytes":      stats.Bytes,
			"elapsed_ms": stats.ElapsedMs,
			"errors":     len(stats.Errors),
		}
	}

	// Add comprehensive summary
	sevCounts, catCounts, extCounts, fileCounts, patCounts, pathTypes := summarizeFindings(findings)

	report["summary"] = map[string]any{
		"by_severity": sevCounts,
		"by_category": catCounts,
		"by_extension": extCounts,
		"by_pattern":  patCounts,
		"by_file":      fileCounts,
		"by_path_type": pathTypes,
	}

	return json.MarshalIndent(report, "", "  ")
}

// generateSummary creates a comprehensive text summary report
func generateSummary(findings []core.Finding, stats *core.Stats, opts ReportOptions) ([]byte, error) {
	var buf strings.Builder

	buf.WriteString("=== Atheon Security & Pattern Scan Report ===\n")
	buf.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("Version: %s\n", version))
	buf.WriteString(fmt.Sprintf("Recursive: %v\n", opts.Recursive))
	if len(opts.FilterPatterns) > 0 {
		buf.WriteString(fmt.Sprintf("Pattern Filters: %s\n", strings.Join(opts.FilterPatterns, ", ")))
	}
	if len(opts.FilterExts) > 0 {
		buf.WriteString(fmt.Sprintf("Extension Filters: %s\n", strings.Join(opts.FilterExts, ", ")))
	}
	buf.WriteString("\n")

	// Summary stats
	buf.WriteString(fmt.Sprintf("Total Findings: %d\n", len(findings)))
	if stats != nil {
		buf.WriteString(fmt.Sprintf("Files Scanned: %d\n", stats.Files))
		buf.WriteString(fmt.Sprintf("Bytes Scanned: %s\n", formatBytes(stats.Bytes)))
		buf.WriteString(fmt.Sprintf("Scan Time: %dms\n", stats.ElapsedMs))
	}
	buf.WriteString("\n")

	// Comprehensive analysis
	sevCounts, catCounts, extCounts, fileCounts, patCounts, pathTypes := summarizeFindings(findings)

	buf.WriteString("--- By Severity ---\n")
	for _, sev := range []string{"critical", "high", "medium", "low"} {
		count := sevCounts[sev]
		if count > 0 {
			buf.WriteString(fmt.Sprintf("  %s: %d\n", sev, count))
		}
	}
	buf.WriteString("\n")

	buf.WriteString("--- By Category ---\n")
	sortedCats := sortMapByValue(catCounts)
	for _, cat := range sortedCats {
		buf.WriteString(fmt.Sprintf("  %s: %d\n", cat.key, cat.value))
	}
	buf.WriteString("\n")

	buf.WriteString("--- By Extension ---\n")
	sortedExts := sortMapByValue(extCounts)
	for _, ext := range sortedExts {
		if ext.key == "" {
			ext.key = "(no extension)"
		}
		buf.WriteString(fmt.Sprintf("  %s: %d\n", ext.key, ext.value))
	}
	buf.WriteString("\n")

	buf.WriteString("--- By Path Type ---\n")
	for _, pt := range sortMapByValue(pathTypes) {
		buf.WriteString(fmt.Sprintf("  %s: %d\n", pt.key, pt.value))
	}
	buf.WriteString("\n")

	buf.WriteString("--- Top Files (Most Issues) ---\n")
	for _, fc := range takeTop(fileCounts, 10) {
		buf.WriteString(fmt.Sprintf("  %s: %d\n", filepath.Base(fc.key), fc.value))
	}
	buf.WriteString("\n")

	buf.WriteString("--- Top Patterns ---\n")
	for _, pc := range takeTop(patCounts, 15) {
		buf.WriteString(fmt.Sprintf("  %s: %d\n", pc.key, pc.value))
	}

	return []byte(buf.String()), nil
}

// generateXLSX creates an Excel XLSX report
func generateXLSX(findings []core.Finding, stats *core.Stats) ([]byte, error) {
	// Return informative error - CSV is the primary export format
	return nil, fmt.Errorf("Excel format requires the excelize library. Use --report-format=csv or --report-format=json instead")
}

// sortFindings sorts findings by severity, then category, then file
func sortFindings(findings []core.Finding) []core.Finding {
	sorted := make([]core.Finding, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		sevOrder := map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3}
		si := sevOrder[strings.ToLower(sorted[i].Severity)]
		sj := sevOrder[strings.ToLower(sorted[j].Severity)]
		if si != sj {
			return si < sj
		}
		if sorted[i].Category != sorted[j].Category {
			return sorted[i].Category < sorted[j].Category
		}
		return sorted[i].File < sorted[j].File
	})
	return sorted
}

// summarizeFindings creates comprehensive statistics from findings
func summarizeFindings(findings []core.Finding) (sev, cat, ext, file, pat, path map[string]int) {
	sev = map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0}
	cat = map[string]int{}
	ext = map[string]int{}
	file = map[string]int{}
	pat = map[string]int{}
	path = map[string]int{}

	for _, f := range findings {
		// Severity
		sevKey := strings.ToLower(f.Severity)
		sev[sevKey]++

		// Category
		cat[f.Category]++

		// Extension
		extKey := strings.ToLower(filepath.Ext(f.File))
		if extKey == "" {
			extKey = "(no ext)"
		}
		ext[extKey]++

		// File
		file[f.File]++

		// Pattern
		pat[f.Pattern]++

		// Path type
		pathType := getPathType(f.File)
		path[pathType]++
	}
	return
}

// getPathType determines the path type (local, network, UNC, etc.)
func getPathType(path string) string {
	if strings.HasPrefix(path, "\\\\") || strings.HasPrefix(path, "//") {
		return "UNC/Network"
	}
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "/home") && !strings.HasPrefix(path, "/Users") {
		return "Unix Absolute"
	}
	if strings.HasPrefix(path, "C:") || strings.HasPrefix(path, "D:") || strings.HasPrefix(path, "E:") {
		return "Windows Absolute"
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return "URL"
	}
	return "Relative"
}

// mapEntry is a key-value pair for sorted output
type mapEntry struct {
	key   string
	value int
}

// sortMapByValue sorts a map by int value, returns top N
func sortMapByValue(m map[string]int) []mapEntry {
	list := make([]mapEntry, 0, len(m))
	for k, v := range m {
		list = append(list, mapEntry{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].value > list[j].value
	})
	return list
}

// takeTop returns the top N items from a map
func takeTop(m map[string]int, n int) []mapEntry {
	sorted := sortMapByValue(m)
	if len(sorted) < n {
		return sorted
	}
	return sorted[:n]
}

// parseReportFormat parses the report format from a string
func parseReportFormat(s string) (ReportFormat, error) {
	switch strings.ToLower(s) {
	case "csv":
		return ReportFormatCSV, nil
	case "json":
		return ReportFormatJSON, nil
	case "xlsx", "excel":
		return ReportFormatXLSX, nil
	case "summary", "text", "txt":
		return ReportFormatSummary, nil
	default:
		return "", fmt.Errorf("unknown format: %s (use csv, json, xlsx, or summary)", s)
	}
}
