package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// captureStdout redirects os.Stdout to a pipe and returns the captured bytes.
// It restores os.Stdout regardless of what happens in f — including panic,
// t.Fatal, and t.FailNow paths. Coderabbit MAJ-4 (PR #55): the original
// cleanup only ran on normal return, leaving subsequent tests running
// against a leaked pipe. Defer is the standard fix.
func captureStdout(t *testing.T, f func()) (out string) {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	// Restore os.Stdout before the pipe is closed so any panic message from
	// `f` (or from the deferred restore on a different goroutine) still
	// reaches the test runner instead of being captured into the pipe.
	defer func() {
		os.Stdout = orig
	}()

	done := make(chan string, 1)
	go func() {
		var sb strings.Builder
		io.Copy(&sb, r) //nolint:errcheck
		done <- sb.String()
	}()

	defer func() {
		w.Close()
		out = <-done
		r.Close()
	}()

	f()
	return out
}

// TestRunSARIF exercises the --sarif flag path through run().
func TestRunSARIF(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "clean.go")
	if err := os.WriteFile(tmp, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := run(context.Background(), []string{"--sarif", tmp})
	if code != 0 {
		t.Errorf("expected exit 0 for --sarif on clean file, got %d", code)
	}
}

// TestRunSARIFWithFindings exercises --sarif when findings are present (exit 1).
func TestRunSARIFWithFindings(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "secrets.txt")
	if err := os.WriteFile(tmp, []byte(`AKIAIOSFODNN7EXAMPLE`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Coderabbit MAJ-3 (PR #55): the previous form ignored `run(...)`'s return
	// code, so a real exit-0 (broken matcher) would pass silently. Asserting
	// the expected non-zero exit code makes regressions visible.
	code := run(context.Background(), []string{"--sarif", tmp})
	if code == 0 {
		t.Errorf("expected non-zero exit when findings are present, got 0")
	}
}

// TestRunSARIFOutputIsValidJSON verifies --sarif emits parseable SARIF JSON.
func TestRunSARIFOutputIsValidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "clean.go")
	if err := os.WriteFile(tmp, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		run(context.Background(), []string{"--sarif", tmp}) //nolint:errcheck
	})

	var sarif map[string]any
	if err := json.Unmarshal([]byte(out), &sarif); err != nil {
		t.Fatalf("--sarif output is not valid JSON: %v\noutput: %s", err, out)
	}
	if sarif["version"] != "2.1.0" {
		t.Errorf("expected SARIF version 2.1.0, got %v", sarif["version"])
	}
	if _, ok := sarif["runs"]; !ok {
		t.Error("SARIF output missing 'runs' key")
	}
}

// TestRunFileWithFindings exercises the --file exit-1 branch (findings found).
func TestRunFileWithFindings(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "secrets.txt")
	// A string that looks like an AWS key to trigger a finding
	if err := os.WriteFile(tmp, []byte(`aws_key = "AKIAIOSFODNN7EXAMPLE"`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Coderabbit MAJ-3 (PR #55): assert the exit code instead of ignoring it
	// so a regression that exits 0 (broken matcher) fails the test.
	code := run(context.Background(), []string{"--file", tmp})
	if code == 0 {
		t.Errorf("expected non-zero exit when findings are present, got 0")
	}
}

// TestRunDefaultPathWithFindings exercises the default-branch file scan exit-1 path.
func TestRunDefaultPathWithFindings(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "secrets.txt")
	if err := os.WriteFile(tmp, []byte(`AKIAIOSFODNN7EXAMPLE`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Coderabbit MAJ-3 (PR #55): assert the exit code instead of ignoring it
	// so a regression that exits 0 (broken matcher) fails the test.
	code := run(context.Background(), []string{tmp})
	if code == 0 {
		t.Errorf("expected non-zero exit when findings are present, got 0")
	}
}

// TestRunSARIFDir exercises --sarif on a directory path.
func TestRunSARIFDir(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "clean.go")
	if err := os.WriteFile(f, []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := run(context.Background(), []string{"--sarif", dir})
	if code != 0 {
		t.Errorf("expected exit 0 for --sarif on clean dir, got %d", code)
	}
}

// TestPrintFindingsLineZero exercises the f.Line == 0 branch in printFindings
// where loc stays as f.File (no ":N" suffix appended).
func TestPrintFindingsLineZero(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "test", File: "somefile.txt", Line: 0, Content: "secret content here!"},
	}
	out := captureStdout(t, func() {
		printFindings(findings, nil, false, false)
	})
	if !strings.Contains(out, "somefile.txt") {
		t.Errorf("expected file name in output, got: %s", out)
	}
	// Must NOT contain ":0" — the zero-line branch skips the colon-number suffix
	if strings.Contains(out, ":0") {
		t.Errorf("output should not contain ':0' for zero line number, got: %s", out)
	}
}

// TestPrintFindingsStatsZeroFiles exercises the stats.Files == 0 branch
// (stats block is suppressed when no files were scanned).
func TestPrintFindingsStatsZeroFiles(t *testing.T) {
	findings := []core.Finding{}
	stats := &core.Stats{Files: 0, Bytes: 0, ElapsedMs: 0}
	out := captureStdout(t, func() {
		printFindings(findings, stats, false, false)
	})
	// Stats line ("scanned N file(s)...") must not appear when Files == 0
	if strings.Contains(out, "scanned") {
		t.Errorf("expected no stats line for Files=0, got: %s", out)
	}
}

// TestBuildSARIFRulesEmpty exercises buildSARIFRules with no findings.
// Post-PR-#96 the rules universe is bundle-wide, so an empty-findings
// scan still produces the full enabled-pattern set (~270 rules). The
// pre-#96 contract (return nil for empty findings) was a finding-derived
// subset and is intentionally broken here — see TestSARIFRuleUniversePresent
// for the new invariant.
func TestBuildSARIFRulesEmpty(t *testing.T) {
	rules := buildSARIFRules(nil)
	if rules == nil {
		t.Fatal("expected full rule universe even with no findings, got nil")
	}
	if len(rules) < 250 {
		t.Errorf("expected full universe (>=250 rules), got %d", len(rules))
	}
}

// TestBuildSARIFRulesSingleFinding verifies that the SARIF rules universe
// is bundle-wide, not finding-derived. A single finding against a custom
// pattern still produces the full enabled-pattern set — the new pattern's
// severity flows through to security-severity fields when it's loaded.
// The "single finding" angle is exercised by TestSARIFRuleUniversePresent
// (every enabled pattern is in the universe) and TestBuildSARIFResultsLocation
// (the finding itself is rendered as a SARIF result).
func TestBuildSARIFRulesSingleFinding(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "aws-access-key", File: "x.go", Line: 1, Severity: "high"},
	}
	rules := buildSARIFRules(findings)
	if len(rules) < 250 {
		t.Fatalf("expected full universe regardless of findings, got %d", len(rules))
	}
	// Spot-check: every rule carries the new schema fields.
	for _, r := range rules {
		if r["kind"] != "rule" {
			t.Errorf("rule kind: got %v", r["kind"])
		}
		props, ok := r["properties"].(map[string]any)
		if !ok {
			t.Errorf("rule missing properties: %v", r)
			continue
		}
		if _, ok := props["tags"].([]string); !ok {
			t.Errorf("rule %v missing properties.tags", r["id"])
		}
		if props["precision"] != "high" {
			t.Errorf("rule %v missing properties.precision 'high'", r["id"])
		}
	}
}

// TestBuildSARIFRulesDeduplicates is now subsumed by the
// TestSARIFRuleUniversePresent invariant — duplicate pattern names are
// impossible when iterating core.All(). Kept as a no-op doc test so
// future maintainers don't think the dedup branch was deleted by
// accident.
func TestBuildSARIFRulesDeduplicates(t *testing.T) {
	rules := buildSARIFRules([]core.Finding{
		{Pattern: "aws-access-key", File: "a.go", Line: 1},
		{Pattern: "aws-access-key", File: "b.go", Line: 2},
	})
	seen := map[string]int{}
	for _, r := range rules {
		if id, ok := r["id"].(string); ok {
			seen[id]++
		}
	}
	for id, count := range seen {
		if count > 1 {
			t.Errorf("duplicate rule id %q appears %d times", id, count)
		}
	}
}

// TestBuildSARIFResultsLocation verifies the location structure produced by
// buildSARIFResults for a single finding.
func TestBuildSARIFResultsLocation(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "loc-pattern", File: "src/main.go", Line: 42, Content: "secret-data", Severity: "high"},
	}
	results := buildSARIFResults(findings)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	res := results[0]
	if res["ruleId"] != "loc-pattern" {
		t.Errorf("expected ruleId 'loc-pattern', got %v", res["ruleId"])
	}
	if res["level"] != "error" {
		t.Errorf("expected level 'error' for high severity, got %v", res["level"])
	}
	locs, ok := res["locations"].([]map[string]any)
	if !ok || len(locs) == 0 {
		t.Fatalf("expected non-empty locations slice, got %v", res["locations"])
	}
	physLoc, ok := locs[0]["physicalLocation"].(map[string]any)
	if !ok {
		t.Fatalf("expected physicalLocation map, got %T", locs[0]["physicalLocation"])
	}
	artifactLoc, ok := physLoc["artifactLocation"].(map[string]any)
	if !ok {
		t.Fatalf("expected artifactLocation map, got %T", physLoc["artifactLocation"])
	}
	if artifactLoc["uri"] != "src/main.go" {
		t.Errorf("expected uri 'src/main.go', got %v", artifactLoc["uri"])
	}
	region, ok := physLoc["region"].(map[string]any)
	if !ok {
		t.Fatalf("expected region map, got %T", physLoc["region"])
	}
	if region["startLine"] != 42 {
		t.Errorf("expected startLine 42, got %v", region["startLine"])
	}
}

// TestBuildSARIFResultsEmptyNonNil verifies buildSARIFResults on an empty
// slice returns an empty (non-nil) slice rather than nil — the make()
// call in buildSARIFResults guarantees this.
func TestBuildSARIFResultsEmptyNonNil(t *testing.T) {
	results := buildSARIFResults([]core.Finding{})
	if results == nil {
		t.Error("expected non-nil empty slice from buildSARIFResults, got nil")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestPrintFindingsLinePositive exercises the f.Line > 0 branch in
// printFindings where the location string becomes "file:N".
func TestPrintFindingsLinePositive(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "test-pat", File: "code.go", Line: 7, Content: "secret content here!"},
	}
	out := captureStdout(t, func() {
		printFindings(findings, nil, false, false)
	})
	if !strings.Contains(out, "code.go:7") {
		t.Errorf("expected 'code.go:7' in output for Line=7, got: %s", out)
	}
}

// TestPrintFindingsStatsPositiveFiles exercises the stats.Files > 0 branch
// in printFindings (the "scanned N file(s)..." stats line is emitted).
func TestPrintFindingsStatsPositiveFiles(t *testing.T) {
	findings := []core.Finding{}
	stats := &core.Stats{Files: 3, Bytes: 4096, ElapsedMs: 12}
	out := captureStdout(t, func() {
		printFindings(findings, stats, false, false)
	})
	if !strings.Contains(out, "scanned") {
		t.Errorf("expected stats 'scanned' line for Files=3, got: %s", out)
	}
	if !strings.Contains(out, "3") {
		t.Errorf("expected file count '3' in stats line, got: %s", out)
	}
}

// TestPrintFindingsNoFindingsText verifies the "no findings." message is
// printed when the findings slice is empty and output is plain text.
func TestPrintFindingsNoFindingsText(t *testing.T) {
	out := captureStdout(t, func() {
		printFindings(nil, nil, false, false)
	})
	if !strings.Contains(out, "no findings") {
		t.Errorf("expected 'no findings' text for empty slice, got: %s", out)
	}
}

// TestRunSARIFWithFindingsOutputIsValidJSON verifies that --sarif still
// produces parseable SARIF JSON even when findings are present, and that
// the "results" array in the run is non-empty.
func TestRunSARIFWithFindingsOutputIsValidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "secrets.txt")
	// Write a known AWS-key pattern that is likely to trigger findings.
	if err := os.WriteFile(tmp, []byte(`AKIAIOSFODNN7EXAMPLE`), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		run(context.Background(), []string{"--sarif", tmp}) //nolint:errcheck
	})

	var sarif map[string]any
	if err := json.Unmarshal([]byte(out), &sarif); err != nil {
		t.Fatalf("--sarif output with findings is not valid JSON: %v\noutput: %s", err, out)
	}
	if sarif["version"] != "2.1.0" {
		t.Errorf("expected SARIF version 2.1.0, got %v", sarif["version"])
	}
	runs, ok := sarif["runs"].([]any)
	if !ok || len(runs) == 0 {
		t.Fatal("SARIF 'runs' missing or empty")
	}
	// Coderabbit MIN-5 (PR #55): the test description claims to verify that
	// `results` is non-empty, but the original code stopped at `runs`.
	run0, ok := runs[0].(map[string]any)
	if !ok {
		t.Fatal("SARIF runs[0] is not an object")
	}
	results, ok := run0["results"].([]any)
	if !ok || len(results) == 0 {
		t.Fatal("SARIF runs[0].results missing or empty (findings expected)")
	}
}

// TestPrintSARIFFindingsStructure verifies that printSARIFFindings produces
// SARIF JSON with the correct $schema, version, and runs[0].tool.driver.name.
func TestPrintSARIFFindingsStructure(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "aws-key", File: "creds.txt", Line: 1, Content: "AKIAIOSFODNN7EXAMPLE"},
	}
	out := captureStdout(t, func() {
		printSARIFFindings(findings, nil)
	})

	var sarif map[string]any
	if err := json.Unmarshal([]byte(out), &sarif); err != nil {
		t.Fatalf("printSARIFFindings output is not valid JSON: %v\noutput: %s", err, out)
	}
	if sarif["$schema"] == nil || sarif["$schema"] == "" {
		t.Error("SARIF output missing '$schema' key")
	}
	runs, ok := sarif["runs"].([]any)
	if !ok || len(runs) == 0 {
		t.Fatal("SARIF 'runs' missing or empty")
	}
	run0, ok := runs[0].(map[string]any)
	if !ok {
		t.Fatal("runs[0] is not a map")
	}
	tool, ok := run0["tool"].(map[string]any)
	if !ok {
		t.Fatal("runs[0].tool is not a map")
	}
	driver, ok := tool["driver"].(map[string]any)
	if !ok {
		t.Fatal("runs[0].tool.driver is not a map")
	}
	if driver["name"] != "Atheon" {
		t.Errorf("expected driver name 'Atheon', got %v", driver["name"])
	}
}

// TestSARIFSeverityMapping verifies each severity produces the expected
// SARIF level + CVSS-like score. Pins down the contract added in PR #84
// (severity wiring) and tightened in PR #96 (empty/unknown → "none"
// rather than the previous "warning" escalation).
func TestSARIFSeverityMapping(t *testing.T) {
	cases := []struct {
		sev   string
		level string
		score string
	}{
		{"critical", "error", "9.5"},
		{"high", "error", "7.5"},
		{"medium", "warning", "5.0"},
		{"low", "note", "2.5"},
		{"", "none", "5.0"}, // unknown → "none" (was "warning" pre-PR-#96)
		{"bogus", "none", "5.0"},
		{"HIGH", "error", "7.5"}, // case-insensitive
	}
	for _, tc := range cases {
		if got := sarifLevel(tc.sev); got != tc.level {
			t.Errorf("sarifLevel(%q) = %q, want %q", tc.sev, got, tc.level)
		}
		if got := sarifSeverityScore(tc.sev); got != tc.score {
			t.Errorf("sarifSeverityScore(%q) = %q, want %q", tc.sev, got, tc.score)
		}
	}
}

// TestSARIFRuleUniversePresent asserts the post-PR-#96 invariant that
// `tool.driver.rules` contains every ENABLED pattern in the bundle,
// not just patterns that produced findings in this particular scan.
// GitHub's Security tab reads this list to render the "available
// rules" sidebar; if the array is finding-derived, every rule that
// didn't fire in the scan is invisible. With ~270 patterns in the
// bundle, asserting `len >= 250` catches both the regression (we'd see
// the count collapse to the per-scan match count, typically <10) and
// accidental bundle pruning.
func TestSARIFRuleUniversePresent(t *testing.T) {
	rules := buildSARIFRules(nil)
	if len(rules) < 250 {
		t.Fatalf("expected at least 250 rules in tool.driver.rules, got %d (rules universe regression)", len(rules))
	}

	// Spot-check: every enabled bundle pattern should appear. If
	// buildSARIFRules accidentally filters by `Enabled() == true &&
	// len(findings) > 0` the count would still pass but specific
	// patterns would be missing.
	seen := map[string]bool{}
	for _, r := range rules {
		if id, ok := r["id"].(string); ok {
			seen[id] = true
		}
	}
	for _, p := range core.All() {
		if p.Enabled() && !seen[p.Name()] {
			t.Errorf("enabled pattern %q missing from SARIF rules universe", p.Name())
		}
	}
}

// TestSARIFResultColumnsAndRedaction asserts the post-PR-#96 SARIF
// schema additions: region.startColumn / endColumn are populated,
// region.snippet.text is redacted (never raw Content), and the result
// carries partialFingerprints + uriBaseId. These are the properties
// GitHub code-scanning consumes to dedupe alerts and navigate to
// source — without them the Security tab shows floating paths and
// floods with duplicate alerts.
func TestSARIFResultColumnsAndRedaction(t *testing.T) {
	findings := []core.Finding{{
		Pattern:  "aws-access-key",
		File:     "config/aws.yml",
		Line:     3,
		Column:   17,
		Content:  "AKIAIOSFODNN7EXAMPLE",
		Severity: "high",
	}}
	results := buildSARIFResults(findings)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]

	// uriBaseId resolves via the run's originalUriBaseIds map. Without
	// it, GitHub renders the path as a bare string with no root.
	loc := r["locations"].([]map[string]any)[0]["physicalLocation"].(map[string]any)
	al := loc["artifactLocation"].(map[string]any)
	if al["uriBaseId"] != "%SRCROOT%" {
		t.Errorf("uriBaseId missing or wrong: %v", al["uriBaseId"])
	}
	if al["uri"] != "config/aws.yml" {
		t.Errorf("uri lost: %v", al["uri"])
	}

	// Region: column + snippet.
	region := loc["region"].(map[string]any)
	if region["startLine"] != 3 {
		t.Errorf("startLine: %v", region["startLine"])
	}
	if region["startColumn"] != 17 {
		t.Errorf("startColumn: %v", region["startColumn"])
	}
	if region["endColumn"] != 17+len("AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("endColumn: %v", region["endColumn"])
	}
	snippet, _ := region["snippet"].(map[string]any)
	snipText, _ := snippet["text"].(string)
	// redact() shows first 4 + **** + last 4. The example key is
	// 20 chars, so the redaction is AKIA****MPLE. Anything containing
	// the literal "AKIAIOSFODNN7EXAMPLE" would mean we forgot to
	// redact — a real-world secret-disclosure bug.
	if strings.Contains(snipText, "AKIAIOSFODNN7EXAMPLE") {
		t.Fatalf("snippet leaked raw secret: %q", snipText)
	}
	if !strings.Contains(snipText, "AKIA") || !strings.Contains(snipText, "MPLE") {
		t.Errorf("snippet should keep bookend chars for recognisability: %q", snipText)
	}

	// partialFingerprints — GitHub's dedup key.
	fp, _ := r["partialFingerprints"].(map[string]any)
	if fp == nil || fp["atheonLoc"] == nil {
		t.Errorf("partialFingerprints.atheonLoc missing")
	}
}

// TestSARIFSchemaURLFrozen pins the schema URL to the OASIS-frozen
// csd03 tag (not master) so consumers can pin against a stable
// revision. A regression here would silently break downstream
// tooling that caches the schema by URL.
func TestSARIFSchemaURLFrozen(t *testing.T) {
	out := captureStdout(t, func() {
		printSARIFFindings(nil, nil) //nolint:errcheck
	})
	const want = "csd03/Schemata/sarif-schema-2.1.0.json"
	if !strings.Contains(out, want) {
		t.Errorf("schema URL missing %q fragment; output: %s", want, out)
	}
	const bad = "master/Schemata/sarif-schema-2.1.0.json"
	if strings.Contains(out, bad) {
		t.Errorf("schema URL points at moving master branch; output: %s", out)
	}
}
