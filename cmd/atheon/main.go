package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/aliasfoxkde/Atheon/core"
)

// version is injected at build time via ldflags
var version = "dev"

func main() {
	configureLogging()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:])
	cancel() // explicit: os.Exit skips deferred cancel
	os.Exit(code)
}

// configureLogging sets the default slog handler based on env vars so
// operators can pipe logs to ELK / Loki / Datadog without code changes.
// ATHEON_LOG_FORMAT=json   — JSON-structured records (one per line)
// ATHEON_LOG_FORMAT=text   — human-readable text (default)
// ATHEON_LOG_LEVEL=debug   — surface slog.Debug calls; default is info
// Called once from main() so all downstream package logs inherit the
// configured handler via slog.Default().
func configureLogging() {
	var level slog.Level
	switch strings.ToLower(os.Getenv("ATHEON_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.EqualFold(os.Getenv("ATHEON_LOG_FORMAT"), "json") {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}

// run executes the CLI with the given args and returns the exit code.
// This is separated from main() so tests can call it without os.Exit
// terminating the test process.
//
// The context flows through every Scan*/DownloadBundle call so callers
// (typically signal.NotifyContext from main) can cancel in-flight work.
func run(ctx context.Context, args []string) int {
	// Strip --json / --sarif first so that `atheon --json --version`
	// (a common CI-wrapper invocation) prints the version rather than
	// falling through to the default branch and erroring with
	// "path not found: --version".
	jsonOutput := len(args) > 0 && args[0] == "--json"
	sarifOutput := len(args) > 0 && args[0] == "--sarif"
	if jsonOutput || sarifOutput {
		args = args[1:]
	}

	// Handle --version flag (checked AFTER the json/sarif strip so that
	// the flag order is forgiving).
	if len(args) > 0 && args[0] == "--version" {
		fmt.Printf("atheon %s\n", version)
		return 0
	}

	cats, args, enableAll := parseCategories(args)
	if enableAll {
		core.EnableAllPatterns()
	}
	core.SetActiveCategories(cats)

	if len(args) == 0 {
		printHelp()
		return 0
	}

	switch args[0] {
	case "update":
		fmt.Println("downloading patterns bundle...")
		if err := core.DownloadBundle(ctx, false); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("patterns updated.")
		return 0

	case "enable":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: enable requires a pattern name")
			return 1
		}
		if !core.EnablePattern(args[1]) {
			fmt.Fprintf(os.Stderr, "error: pattern '%s' not found\n", args[1])
			return 1
		}
		fmt.Printf("enabled pattern: %s\n", args[1])
		return 0

	case "disable":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: disable requires a pattern name")
			return 1
		}
		if !core.DisablePattern(args[1]) {
			fmt.Fprintf(os.Stderr, "error: pattern '%s' not found\n", args[1])
			return 1
		}
		fmt.Printf("disabled pattern: %s\n", args[1])
		return 0

	case "list":
		return cmdList(args[1:])

	case "--help", "help", "-h":
		printHelp()
		return 0

	case "--env":
		findings := core.ScanEnv(ctx)
		printFindings(findings, nil, jsonOutput, sarifOutput)
		if len(findings) > 0 {
			return 1
		}
		return 0

	case "-", "--stdin":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: reading stdin:", err)
			return 1
		}
		findings := core.ScanString(ctx, string(data), "stdin")
		printFindings(findings, nil, jsonOutput, sarifOutput)
		if len(findings) > 0 {
			return 1
		}
		return 0

	case "--file":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "error: --file requires a path")
			return 1
		}
		findings, stats, err := core.ScanFile(ctx, args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		printFindings(findings, stats, jsonOutput, sarifOutput)
		if len(findings) > 0 || scanErrorsPresent(stats) {
			return 1
		}
		return 0

	default:
		path := args[0]
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: path not found:", path)
			return 1
		}
		var findings []core.Finding
		var stats *core.Stats
		if info.IsDir() {
			findings, stats, err = core.ScanDir(ctx, path, scanOpts(args[1:]))
		} else {
			findings, stats, err = core.ScanFile(ctx, path)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		printFindings(findings, stats, jsonOutput, sarifOutput)
		if len(findings) > 0 || scanErrorsPresent(stats) {
			return 1
		}
		return 0
	}
}

func parseCategories(args []string) (cats, rest []string, enableAll bool) {
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--categories="):
			val := strings.TrimPrefix(a, "--categories=")
			for _, c := range strings.Split(val, ",") {
				if c = strings.TrimSpace(c); c != "" {
					cats = append(cats, c)
				}
			}
		case a == "--all":
			enableAll = true
		default:
			rest = append(rest, a)
		}
	}
	return
}

// scanOpts extracts directory-scan flags from the post-path argument
// tail and translates them into a core.ScanOpts. CLI defaults keep the
// historical behaviour (follow symlinks, package-level maxFileSize) so
// existing scripts don't change semantics silently; the new
// --no-follow-symlinks flag opts in to the safer default that the MCP
// server uses unconditionally.
func scanOpts(rest []string) core.ScanOpts {
	var opts core.ScanOpts
	for _, a := range rest {
		if a == "--no-follow-symlinks" {
			opts.NoFollowSymlinks = true
		}
	}
	return opts
}

func printFindings(findings []core.Finding, stats *core.Stats, jsonOutput, sarifOutput bool) {
	if jsonOutput {
		printJSONFindings(findings)
		return
	}
	if sarifOutput {
		printSARIFFindings(findings)
		return
	}
	if len(findings) == 0 {
		fmt.Println("no findings.")
	} else {
		for _, f := range findings {
			loc := f.File
			if f.Line > 0 {
				loc = fmt.Sprintf("%s:%d", f.File, f.Line)
			}
			fmt.Printf("%s  %s\n", f.Pattern, loc)
			if f.Content != "" {
				fmt.Println(" ", redact(f.Content))
			}
		}
		fmt.Printf("\n%d finding(s)\n", len(findings))
	}
	if stats != nil && stats.Files > 0 {
		fmt.Printf("scanned %d file(s)  %s  %dms\n",
			stats.Files, formatBytes(stats.Bytes), stats.ElapsedMs)
	}
	// Surface per-file read errors so silent data loss (permission denied,
	// unreadable files) is visible — a scan that "succeeds" with half the
	// tree skipped should not return exit 0 without warning. JSON/SARIF
	// paths stay clean (errors don't pollute the structured stream).
	if !jsonOutput && !sarifOutput && stats != nil && len(stats.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d file(s) could not be read:\n", len(stats.Errors))
		for _, e := range stats.Errors {
			fmt.Fprintf(os.Stderr, "  %v\n", e)
		}
	}
}

// scanErrorsPresent reports whether a scan silently dropped files. The
// caller can use this to bump the process exit code so CI consumers see
// the partial failure even when they don't read stderr (e.g. when only
// stdout is captured by a CI artifact).
func scanErrorsPresent(stats *core.Stats) bool {
	return stats != nil && len(stats.Errors) > 0
}

func printJSONFindings(findings []core.Finding) {
	items := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		items = append(items, map[string]any{"pattern": f.Pattern, "file": f.File, "line": f.Line, "match": redact(f.Content)})
	}
	if err := json.NewEncoder(os.Stdout).Encode(items); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}

// printSARIFFindings outputs findings in SARIF 2.1.0 format for GitHub
// Security tab integration. The schema URL points at the OASIS-frozen
// `csd03` tag (not `master`) so consumers can pin against a stable
// revision; `master` shifts as the spec evolves and breaks tooling that
// caches the schema.
func printSARIFFindings(findings []core.Finding) {
	sarif := map[string]any{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/csd03/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]any{
			{
				"tool": map[string]any{
					"driver": map[string]any{
						"name":           "Atheon",
						"version":        version,
						"informationUri": "https://github.com/aliasfoxkde/Atheon-Enhanced",
						"rules":          buildSARIFRules(findings),
						"supportedTaxonomies": []map[string]any{
							{"name": "CWE", "shortDescription": map[string]any{"text": "Common Weakness Enumeration"}},
						},
					},
				},
				// originalUriBaseIds lets downstream tools (GitHub
				// code-scanning, IDE plugins) resolve the `uriBaseId`
				// references each artifactLocation carries. Without
				// this, file paths in the SARIF are dangling strings
				// — GitHub shows them as relative to nothing, IDEs
				// can't navigate to them.
				"originalUriBaseIds": map[string]any{
					"SRCROOT": map[string]any{
						"uri": "file:///",
					},
				},
				"results": buildSARIFResults(findings),
			},
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(sarif); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}

// sarifSeverityScore maps an Atheon severity to the CVSS-like 0.0–10.0 score
// that GitHub code-scanning consumes via security-severity. The mapping is
// deliberately coarse — pattern authors shouldn't think in CVSS — but the
// scores land on the boundaries GitHub uses for its severity buckets.
func sarifSeverityScore(sev string) string {
	switch strings.ToLower(sev) {
	case "critical":
		return "9.5"
	case "high":
		return "7.5"
	case "medium":
		return "5.0"
	case "low":
		return "2.5"
	default:
		return "5.0"
	}
}

// sarifLevel maps Atheon severity to SARIF's level enum
// (none/note/warning/error). The historical behaviour escalated any
// unknown severity to "warning", which is wrong — an empty severity
// is the scanner saying "I don't know how loud to be", not "I'm
// fairly sure this is a problem". SARIF's "none" is the right
// mapping for that case and lets GitHub code-scanning render the
// result without colouring it.
func sarifLevel(sev string) string {
	switch strings.ToLower(sev) {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "note"
	default:
		return "none"
	}
}

// buildSARIFRules emits the full rule universe for every enabled pattern
// in the bundle — not just the rules that produced findings in this
// scan. Before PR #96 the rules array was derived from the findings
// slice, which meant rules that DIDN'T match any file were invisible to
// GitHub code-scanning (the Security tab shows "0 alerts" for a rule
// but never lists it as available). GitHub also uses the rules array
// to render the rule description and severity before any result is
// produced, so a thin universe leaves the Security tab feeling empty.
//
// Patterns iterate via core.All(); external (non-bundle) patterns are
// included too so Register() callers see their rules. Sort by name for
// deterministic SARIF output (goldens + diff reviewers rely on it).
func buildSARIFRules(findings []core.Finding) []map[string]any {
	patterns := core.All()
	// Deterministic sort: alphabetise by name. Stable across runs so the
	// SARIF diff in PR review is meaningful.
	sort.Slice(patterns, func(i, j int) bool { return patterns[i].Name() < patterns[j].Name() })
	var rules []map[string]any
	for _, p := range patterns {
		if !p.Enabled() {
			continue
		}
		rules = append(rules, map[string]any{
			"id":   p.Name(),
			"name": p.Name(),
			"shortDescription": map[string]any{
				"text": "Pattern " + p.Name() + " matched a line in the scanned tree.",
			},
			"fullDescription": map[string]any{
				"text": "Atheon pattern " + p.Name() + " from category '" + p.Category() + "' matched a line. See the rule's match regex in community/" + p.Category() + "/" + p.Name() + ".yaml.",
			},
			"helpUri": "https://github.com/aliasfoxkde/Atheon-Enhanced/wiki/patterns#" + p.Name(),
			"kind":    "rule",
			"defaultConfiguration": map[string]any{
				"level": sarifLevel(p.Severity()),
			},
			"properties": map[string]any{
				"security-severity":       sarifSeverityScore(p.Severity()),
				"security-severity-label": p.Severity(),
				"category":                p.Category(),
				"tags":                    []string{p.Category(), "security"},
				// Heuristic: bundle patterns are regex-exact → "high"
				// precision. External Register() callers may be
				// keyword-style → "medium". This is a placeholder that
				// pattern authors can override by extending the
				// Pattern interface with a Precision() method later.
				"precision": "high",
			},
		})
	}
	return rules
}

func buildSARIFResults(findings []core.Finding) []map[string]any {
	results := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		region := map[string]any{
			"startLine": f.Line,
		}
		// Column: 0 means the scanner couldn't compute a span
		// (non-bundlePattern, or trailing-newline shenanigans). We
		// emit startColumn / endColumn only when we know them — SARIF
		// allows omitting them, and a wrong 1:1 region is worse than
		// no region at all (consumers might highlight the wrong word).
		if f.Column > 0 {
			region["startColumn"] = f.Column
			region["endColumn"] = f.Column + len(f.Content)
		}
		// Snippet text: redact f.Content before putting it in the
		// SARIF artifact. GitHub stores SARIF uploads indefinitely
		// and renders snippets into the Security tab UI — a literal
		// secret in the snippet would exfiltrate via the SARIF
		// pipeline. redact() keeps the first/last 4 chars so the
		// operator can still recognise the match shape.
		region["snippet"] = map[string]any{"text": redact(f.Content)}

		results = append(results, map[string]any{
			"ruleId": f.Pattern,
			"level":  sarifLevel(f.Severity),
			"message": map[string]any{
				// The message text is the human-readable explanation
				// GitHub surfaces. Use the file path + line as the
				// primary signal; the snippet is in region.snippet
				// for context.
				"text": fmt.Sprintf("%s found in %s at line %d", f.Pattern, f.File, f.Line),
			},
			"locations": []map[string]any{
				{
					"physicalLocation": map[string]any{
						"artifactLocation": map[string]any{
							"uri":       f.File,
							"uriBaseId": "%SRCROOT%",
						},
						"region": region,
					},
				},
			},
			// partialFingerprints lets GitHub's deduplication merge
			// results across runs without exact-match fragility
			// (whitespace, line shifts). Hash of (pattern + file +
			// line + column) — content is excluded so an operator
			// editing the matched line still gets one alert per
			// logical location.
			"partialFingerprints": map[string]any{
				"atheonLoc": fmt.Sprintf("%s|%s|%d|%d", f.Pattern, f.File, f.Line, f.Column),
			},
		})
	}
	return results
}

func cmdList(args []string) int {
	if len(args) > 0 && args[0] == "categories" {
		for _, c := range core.Categories() {
			fmt.Println(c)
		}
		return 0
	}

	var categoryFilter string
	showEnabled := false
	showDisabled := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--category="):
			categoryFilter = strings.TrimPrefix(a, "--category=")
		case a == "--enabled":
			showEnabled = true
		case a == "--disabled":
			showDisabled = true
		}
	}

	// Validate --category against known categories. Without this check,
	// a typo (e.g. `--category=secrets ` or `--category=secret`) silently
	// filters to zero matches and the user sees "0 pattern(s)" — which
	// looks like the category has no patterns, not that the name was wrong.
	if categoryFilter != "" {
		known := false
		for _, c := range core.Categories() {
			if c == categoryFilter {
				known = true
				break
			}
		}
		if !known {
			fmt.Fprintf(os.Stderr, "error: unknown category %q\n", categoryFilter)
			fmt.Fprintln(os.Stderr, "known categories:")
			for _, c := range core.Categories() {
				fmt.Fprintf(os.Stderr, "  %s\n", c)
			}
			return 1
		}
	}

	var filtered []core.Pattern
	for _, p := range core.All() {
		if categoryFilter != "" && p.Category() != categoryFilter {
			continue
		}
		if showEnabled && !p.Enabled() {
			continue
		}
		if showDisabled && p.Enabled() {
			continue
		}
		filtered = append(filtered, p)
	}

	for _, p := range filtered {
		status := "enabled"
		if !p.Enabled() {
			status = "disabled"
		}
		fmt.Printf("%s [%s] [%s]\n", p.Name(), p.Category(), status)
	}
	fmt.Printf("\n%d pattern(s)\n", len(filtered))
	return 0
}

func printHelp() {
	fmt.Print(`atheon - pattern matching engine

usage:
  atheon <path>                       scan a directory or file
  atheon <path> --no-follow-symlinks  scan a directory without following symlinks
  atheon --file <path>                scan a single file explicitly
  atheon --env                        scan environment variables
  atheon - / --stdin                  scan from stdin
  atheon --json <path>                print findings as JSON (must be first flag)
  atheon --categories=<c1,c2> <path>  scan specific categories only
  atheon --all <path>                 scan all patterns including disabled ones
  atheon list                         list all patterns with enabled/disabled status
  atheon list --enabled               list only enabled patterns
  atheon list --disabled              list only disabled patterns
  atheon list --category=<cat>        list patterns in a specific category
  atheon list categories              list available category names
  atheon enable <pattern>             enable a pattern by name
  atheon disable <pattern>            disable a pattern by name
  atheon update                       download latest patterns bundle
  atheon --version                    show version
  atheon --help                       show this message
`)
}

func redact(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

func formatBytes(b int64) string {
	if b >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	}
	if b >= 1<<10 {
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	}
	return fmt.Sprintf("%d B", b)
}
