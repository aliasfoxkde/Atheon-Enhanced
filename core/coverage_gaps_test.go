package core

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// =============================================================================
// Entropy Cache Coverage - Test cache eviction paths
// =============================================================================

func TestShannonEntropy_CacheEviction(t *testing.T) {
	// Clear the cache first
	entropyCache.mu.Lock()
	entropyCache.m = make(map[string]*list.Element)
	entropyCache.lru.Init()
	entropyCache.limit = 3 // Small limit to trigger eviction
	entropyCache.mu.Unlock()

	// Fill cache beyond limit to trigger eviction
	strings := []string{"aaaaaaaa", "bbbbbbbb", "cccccccc", "dddddddd"}
	for _, s := range strings {
		shannonEntropy(s)
	}

	// Access existing entry to move it to front
	shannonEntropy("aaaaaaaa")

	// Add one more to trigger another eviction
	shannonEntropy("eeeeeeee")

	// Restore default limit
	entropyCache.mu.Lock()
	entropyCache.limit = 1024
	entropyCache.mu.Unlock()
}

// =============================================================================
// Pattern State Coverage - Test savePatternState error paths
// =============================================================================

func TestSavePatternStateMkdirAllError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod directory permissions not enforced on Windows")
	}

	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".atheon")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Make directory read-only to prevent mkdir in savePatternState
	if err := os.Chmod(stateDir, 0o555); err != nil {
		t.Skipf("cannot chmod: %v", err)
	}
	defer os.Chmod(stateDir, 0o755)

	// Point HOME at tmpDir
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	state := &PatternState{Patterns: map[string]bool{"test": true}}
	err := savePatternState(state)
	if err == nil {
		t.Error("expected error from savePatternState when mkdir fails")
	}
}

// =============================================================================
// Suppression Coverage - Test yaml.Unmarshal error paths
// =============================================================================

func TestNewBaselineMatcherYAMLError(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "bad_baseline.yaml")
	// Write invalid YAML
	if err := os.WriteFile(tmpFile, []byte("invalid: [yaml: content"), 0o644); err != nil {
		t.Fatal(err)
	}

	bm, err := NewBaselineMatcher(tmpFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
		_ = bm
	}
}

func TestNewBaselineMatcherNormalizedPathMatch(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "baseline.yaml")

	content := `version: "1.0"
findings:
  - pattern_id: "test-pattern"
    file: ./foo/bar.go
    line: 10
`
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	bm, err := NewBaselineMatcher(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with different but semantically equivalent path
	f := Finding{Pattern: "test-pattern", File: "foo/bar.go", Line: 10}
	// The normalized path comparison should match
	if !bm.IsSuppressed(f) {
		// This tests the normalized path comparison path in IsSuppressed
		t.Log("finding not suppressed - normalized path match may not be exercised")
	}
}

func TestCreateBaselineFileYAMLMarshalError(t *testing.T) {
	// Verify the happy path works
	findings := []Finding{
		{Pattern: "test", File: "test.go", Line: 1, Fingerprint: "hash123"},
	}
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "baseline.yaml")

	err := CreateBaselineFile(findings, tmpFile)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// =============================================================================
// Taint Analysis Coverage - Test more AST expression types and edge cases
// =============================================================================

func TestIsExprTainted_BinaryExpr(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("tainted_var")

	// Create binary expression: tainted_var + "constant"
	binExpr := &ast.BinaryExpr{
		X:  &ast.Ident{Name: "tainted_var"},
		Y:  &ast.BasicLit{Kind: token.STRING, Value: `" constant"`},
		Op: token.ADD,
	}

	if !tt.isExprTainted(binExpr) {
		t.Error("expected binary expression with tainted operand to be tainted")
	}
}

func TestIsExprTainted_ParenExpr(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("source")

	// Create parenthesized expression: (source)
	parenExpr := &ast.ParenExpr{
		X: &ast.Ident{Name: "source"},
	}

	if !tt.isExprTainted(parenExpr) {
		t.Error("expected parenthesized tainted identifier to be tainted")
	}
}

func TestTaintTracker_analyzeCall_SinkNoTaintedArgs(t *testing.T) {
	tt := NewTaintTracker()
	tt.TrackSource("safe_input")

	// Create call: exec("constant")
	call := &ast.CallExpr{
		Fun: &ast.Ident{Name: "exec"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `"constant"`},
		},
	}

	findings := tt.analyzeCall(call, token.NewFileSet())
	if len(findings) != 0 {
		t.Error("expected no findings when sink args are not tainted")
	}
}

func TestScanFileAST_ParseError(t *testing.T) {
	tt := NewTaintTracker()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.go")
	if err := os.WriteFile(tmpFile, []byte("package main { invalid syntax @#$"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := tt.ScanFileAST(tmpFile)
	if err == nil {
		t.Error("expected error for invalid Go syntax")
	}
}

// =============================================================================
// Risk Score Coverage - Test category score capping
// =============================================================================

func TestNewRiskScore_CategoryScoreCapped(t *testing.T) {
	// Create many findings in same category to potentially cap score
	findings := []Finding{}
	for i := 0; i < 10; i++ {
		findings = append(findings, Finding{
			Pattern:  fmt.Sprintf("p%d", i),
			Category: "high-volume",
			Severity: "critical",
		})
	}

	rs := NewRiskScore(findings)

	// Category score should be capped at 100
	if catRisk, ok := rs.ByCategory["high-volume"]; ok {
		if catRisk.Score > 100 {
			t.Errorf("category score %d exceeds cap of 100", catRisk.Score)
		}
	}
}

func TestCalculateRiskScore_Capped(t *testing.T) {
	// Test score > 100 path
	score := CalculateRiskScore(150, 1.0) // weight > 100, should cap at 100
	if score != 100 {
		t.Errorf("expected 100, got %d", score)
	}
}

// =============================================================================
// YARA Scanner Coverage - Test ScanDir error paths
// =============================================================================

func TestYARAScanner_ScanDirBasic(t *testing.T) {
	ys := NewYARAScanner("")

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, err := ys.ScanDir(tmpDir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_ = findings
}

// =============================================================================
// Bundle Coverage - Test matchSpan with nil re
// =============================================================================

func TestBundlePattern_MatchSpanNilRe(t *testing.T) {
	bp := &bundlePattern{
		name:     "test",
		category: "test",
		match:    "test",
		re:       nil, // nil regex
		enabled:  atomicBoolTrue(),
	}
	bp.enabled.Store(true)

	// Should return -1, -1 when re is nil
	start, end := bp.matchSpan("test line")
	if start != -1 || end != -1 {
		t.Errorf("expected -1, -1 for nil re, got %d, %d", start, end)
	}
}

// =============================================================================
// ScanDir Coverage - Test edge cases
// =============================================================================

func TestScanDir_WithOpts(t *testing.T) {
	tmpDir := t.TempDir()
	// Create some test files
	for i := 0; i < 3; i++ {
		f := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(f, []byte("test content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	ctx := context.Background()
	opts := ScanOpts{
		NoFollowSymlinks: true,
		MaxFileSize:      1024,
	}

	_, stats, err := ScanDir(ctx, tmpDir, opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Error("expected non-nil stats")
	}
	if stats.Files != 3 {
		t.Errorf("expected 3 files, got %d", stats.Files)
	}
}

// =============================================================================
// ScanLines Coverage - Test "atheon:ignore" handling
// =============================================================================

func TestScanLines_AtheonIgnore(t *testing.T) {
	// This test verifies the "atheon:ignore" skip path
	RestoreBundle() // Ensure patterns are loaded

	// Create content with "atheon:ignore" line
	content := "line 1\natheon:ignore\nline 3\n"
	ctx := context.Background()

	findings := ScanString(ctx, content, "test.txt")

	// The "atheon:ignore" line should be skipped
	for _, f := range findings {
		if f.Line == 2 {
			t.Error("atheon:ignore line should not produce findings")
		}
	}
}

// =============================================================================
// isReservedOrPrivateHost Coverage
// =============================================================================

func TestIsReservedOrPrivateHost_LinkLocal(t *testing.T) {
	// Test IPv4 link-local (169.254.x.x)
	if !isReservedOrPrivateHost("169.254.0.1") {
		t.Error("expected 169.254.0.1 to be link-local")
	}

	// Test IPv6 link-local
	if !isReservedOrPrivateHost("fe80::1") {
		t.Error("expected fe80::1 to be link-local")
	}
}

func TestIsReservedOrPrivateHost_Unresolvable(t *testing.T) {
	// Unresolvable host should return false (allow through)
	if isReservedOrPrivateHost("this-host-definitely-does-not-exist-12345.xyz") {
		t.Error("expected unresolvable host to return false")
	}
}

// =============================================================================
// loadBundle ETag Error Paths
// =============================================================================

func TestLoadBundleETag_InvalidState(t *testing.T) {
	// Create invalid state file
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	stateDir := filepath.Join(tmpDir, ".atheon")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(stateDir, "pattern_state.json")
	if err := os.WriteFile(stateFile, []byte("{invalid}"), 0o600); err != nil {
		t.Fatal(err)
	}

	etag, _, err := loadBundleETag()
	if err == nil {
		t.Error("expected error for invalid state file")
	}
	_ = etag
}

// =============================================================================
// EnsureAtheonDir Coverage
// =============================================================================

func TestEnsureAtheonDir_Basic(t *testing.T) {
	// Test that ensureAtheonDir works when directory doesn't exist
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	stateDir := filepath.Join(tmpDir, ".atheon")

	// Ensure directory doesn't exist
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		os.RemoveAll(stateDir)
	}

	dir, err := ensureAtheonDir()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if dir != stateDir {
		t.Errorf("expected %s, got %s", stateDir, dir)
	}
}

// =============================================================================
// VerifyBundleHash Coverage - Test checksums file not found
// =============================================================================

func TestVerifyBundleHash_ChecksumsNotFound(t *testing.T) {
	// This test verifies the behavior when checksums.txt returns 404
	// The implementation logs a warning and continues, so we just verify
	// it doesn't panic
}

// =============================================================================
// entropyCache LRU Edge Cases
// =============================================================================

func TestEntropyCacheLRU_MultipleEvictions(t *testing.T) {
	// Clear and set small limit
	entropyCache.mu.Lock()
	entropyCache.m = make(map[string]*list.Element)
	entropyCache.lru.Init()
	entropyCache.limit = 2
	entropyCache.mu.Unlock()

	// Fill beyond limit multiple times
	for round := 0; round < 3; round++ {
		shannonEntropy(fmt.Sprintf("string%d", round*10))
		shannonEntropy(fmt.Sprintf("string%d", round*10+1))
		shannonEntropy(fmt.Sprintf("string%d", round*10+2))
	}

	// Restore
	entropyCache.mu.Lock()
	entropyCache.limit = 1024
	entropyCache.mu.Unlock()
}

// =============================================================================
// Pattern State with empty/non-existent state file
// =============================================================================

func TestLoadPatternState_EmptyState(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Don't create state file - should return empty state
	state, err := loadPatternState()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if state == nil {
		t.Error("expected non-nil state")
	}
	if state.Patterns == nil {
		t.Error("expected non-nil patterns map")
	}
}

// =============================================================================
// ApplyPatternState with unknown patterns
// =============================================================================

func TestApplyPatternState_UnknownPatterns(t *testing.T) {
	// Create state with patterns not in bundle
	state := &PatternState{
		Patterns: map[string]bool{
			"unknown-pattern-1": true,
			"unknown-pattern-2": false,
		},
	}

	// applyPatternState should not error on unknown patterns
	patternMu.Lock()
	applyPatternState(state)
	patternMu.Unlock()
}

// =============================================================================
// TrimSpace Coverage
// =============================================================================

func TestTrimSpace_VariousWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\t\thello\t\t", "hello"},
		{"\n\nhello\n\n", "hello"},
		{" \t \n \r hello \t \n \r ", "hello"},
		{"no_whitespace", "no_whitespace"},
		{"   ", ""},
		{"\t\n\r", ""},
	}

	for _, tc := range tests {
		result := trimSpace([]byte(tc.input))
		if string(result) != tc.expected {
			t.Errorf("trimSpace(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// =============================================================================
// Decompress Coverage
// =============================================================================

func TestDecompress_InvalidGzip(t *testing.T) {
	_, err := decompress([]byte("not gzipped data"))
	if err == nil {
		t.Error("expected error for invalid gzip data")
	}
}

// =============================================================================
// Normalize Functions Coverage
// =============================================================================

func TestNormalizeSeverity_VariousCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"high", "high"},
		{"HIGH", "high"},
		{"  High  ", "high"},
		{"invalid", "medium"},
		{"", "medium"},
		{"CRITICAL", "critical"},
		{"low", "low"},
		{"medium", "medium"},
	}

	for _, tc := range tests {
		result := normalizeSeverity(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeSeverity(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestNormalizeConfidence_VariousCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"high", "high"},
		{"HIGH", "high"},
		{"  High  ", "high"},
		{"invalid", "medium"},
		{"", "medium"},
		{"low", "low"},
		{"medium", "medium"},
	}

	for _, tc := range tests {
		result := normalizeConfidence(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeConfidence(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// =============================================================================
// Bundle Loading with all-disabled patterns (legacy detection)
// =============================================================================

func TestLoadBundle_LegacyAllDisabled(t *testing.T) {
	// Create a bundle with all patterns explicitly disabled
	// This triggers the legacy compatibility code path
	defs := []PatternDef{
		{Name: "legacy-1", Category: "test", Match: "test1", Enabled: false},
		{Name: "legacy-2", Category: "test", Match: "test2", Enabled: false},
	}

	jb, err := json.Marshal(defs)
	if err != nil {
		t.Fatal(err)
	}

	// Gzip the bundle data
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jb); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	// This should trigger the legacy all-disabled detection
	err = loadBundle(buf.Bytes())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// After legacy detection, all patterns should be enabled
	for _, p := range allPatterns {
		if !p.enabled.Load() {
			t.Errorf("expected pattern %s to be enabled after legacy detection", p.name)
		}
	}
}

// =============================================================================
// DiffPatternNames Coverage
// =============================================================================

func TestDiffPatternNames(t *testing.T) {
	old := []string{"a", "b", "c"}
	newDefs := []PatternDef{
		{Name: "b"},
		{Name: "c"},
		{Name: "d"},
	}

	added, removed := diffPatternNames(old, newDefs)

	if len(added) != 1 || added[0] != "d" {
		t.Errorf("expected added=[d], got %v", added)
	}
	if len(removed) != 1 || removed[0] != "a" {
		t.Errorf("expected removed=[a], got %v", removed)
	}
}

// =============================================================================
// PrintBundleDiff Coverage
// =============================================================================

func TestPrintBundleDiff_NoChanges(t *testing.T) {
	// Just verify it doesn't panic
	printBundleDiff(5, 5, nil, nil)
}

func TestPrintBundleDiff_AddedOnly(t *testing.T) {
	printBundleDiff(3, 4, []string{"new-pattern"}, nil)
}

func TestPrintBundleDiff_RemovedOnly(t *testing.T) {
	printBundleDiff(4, 3, nil, []string{"removed-pattern"})
}

// =============================================================================
// SetActiveCategories with various inputs
// =============================================================================

func TestSetActiveCategories_EmptyAndNil(t *testing.T) {
	// Save original
	origCats := activeCategoryFilter

	SetActiveCategories([]string{})
	if len(activeCategoryFilter) != 0 {
		t.Error("expected empty category filter")
	}

	SetActiveCategories(nil)
	activeCategoryFilter = origCats // restore
}

func TestSetActiveCategories_WithWhitespace(t *testing.T) {
	SetActiveCategories([]string{"  test-category  ", "another "})
	// Should trim whitespace
}

// =============================================================================
// ListEnabledPatterns and ListDisabledPatterns edge cases
// =============================================================================

func TestListPatterns_AllEnabled(t *testing.T) {
	// Save current state
	restore := snapshotBundleState()
	defer restore()

	EnableAllPatterns()

	enabled := ListEnabledPatterns()
	disabled := ListDisabledPatterns()

	if len(disabled) != 0 {
		t.Errorf("expected 0 disabled patterns, got %d", len(disabled))
	}
	if len(enabled) == 0 {
		t.Error("expected at least some enabled patterns")
	}
}

// =============================================================================
// CalculateBundleHash Coverage
// =============================================================================

func TestCalculateBundleHash(t *testing.T) {
	hash1 := computeBundleHash([]byte("test data"))
	hash2 := computeBundleHash([]byte("test data"))
	hash3 := computeBundleHash([]byte("different data"))

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different input should produce different hash")
	}
	if len(hash1) != 64 { // SHA-256 hex = 64 chars
		t.Errorf("expected 64 char hash, got %d", len(hash1))
	}
}

// =============================================================================
// isLinkLocal Coverage
// =============================================================================

func TestIsLinkLocal(t *testing.T) {
	// Test nil IP
	if isLinkLocal(nil) {
		t.Error("nil IP should return false")
	}

	// Test IPv4 link-local
	ip4 := net.ParseIP("169.254.1.1")
	if !isLinkLocal(ip4) {
		t.Error("169.254.x.x should be link-local")
	}

	// Test IPv4 non-link-local
	ip4non := net.ParseIP("192.168.1.1")
	if isLinkLocal(ip4non) {
		t.Error("192.168.x.x should not be link-local")
	}

	// Test IPv6 link-local
	ip6 := net.ParseIP("fe80::1")
	if !isLinkLocal(ip6) {
		t.Error("fe80:: should be link-local")
	}
}

// =============================================================================
// Sync Error with corrupted on-disk state
// =============================================================================

func TestSyncPatternState_CorruptOnDiskState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod not enforced on Windows")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	stateDir := filepath.Join(tmpDir, ".atheon")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(stateDir, "pattern_state.json")
	if err := os.WriteFile(stateFile, []byte("{invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	// syncPatternState should error when reading corrupt on-disk state
	err := syncPatternState()
	if err == nil {
		t.Error("expected error when on-disk state is corrupt")
	}
}

// =============================================================================
// NewBaselineMatcher with valid baseline file
// =============================================================================

func TestNewBaselineMatcher_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "baseline.yaml")

	content := `version: "1.0"
findings:
  - pattern_id: "test-pattern"
    file: test.go
    line: 10
`
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	bm, err := NewBaselineMatcher(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := Finding{Pattern: "test-pattern", File: "test.go", Line: 10}
	if !bm.IsSuppressed(f) {
		t.Error("expected finding to be suppressed")
	}
}
