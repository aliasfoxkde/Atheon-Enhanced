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
func TestBuildSARIFRulesEmpty(t *testing.T) {
	rules := buildSARIFRules(nil)
	if rules != nil {
		t.Errorf("expected nil rules for empty findings, got %v", rules)
	}
}

// TestBuildSARIFRulesSingleFinding verifies that a single finding produces
// exactly one rule with the expected id, name, kind, and properties fields.
func TestBuildSARIFRulesSingleFinding(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "my-pattern", File: "x.go", Line: 1},
	}
	rules := buildSARIFRules(findings)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	r := rules[0]
	if r["id"] != "my-pattern" {
		t.Errorf("expected rule id 'my-pattern', got %v", r["id"])
	}
	if r["name"] != "my-pattern" {
		t.Errorf("expected rule name 'my-pattern', got %v", r["name"])
	}
	if r["kind"] != "rule" {
		t.Errorf("expected rule kind 'rule', got %v", r["kind"])
	}
	props, ok := r["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties map, got %T", r["properties"])
	}
	if props["security-severity"] != "High" {
		t.Errorf("expected security-severity 'High', got %v", props["security-severity"])
	}
}

// TestBuildSARIFRulesDeduplicates verifies that duplicate pattern names
// produce only one rule (deduplication logic in buildSARIFRules).
func TestBuildSARIFRulesDeduplicates(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "dup-pat", File: "a.go", Line: 1},
		{Pattern: "dup-pat", File: "b.go", Line: 2},
		{Pattern: "other-pat", File: "c.go", Line: 3},
	}
	rules := buildSARIFRules(findings)
	if len(rules) != 2 {
		t.Errorf("expected 2 deduplicated rules, got %d", len(rules))
	}
	ids := make(map[string]bool)
	for _, r := range rules {
		if id, ok := r["id"].(string); ok {
			ids[id] = true
		}
	}
	if !ids["dup-pat"] {
		t.Error("expected dup-pat in rules")
	}
	if !ids["other-pat"] {
		t.Error("expected other-pat in rules")
	}
}

// TestBuildSARIFResultsLocation verifies the location structure produced by
// buildSARIFResults for a single finding.
func TestBuildSARIFResultsLocation(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "loc-pattern", File: "src/main.go", Line: 42, Content: "secret-data"},
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
		t.Errorf("expected level 'error', got %v", res["level"])
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
		printSARIFFindings(findings)
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
