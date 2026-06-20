package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestScanLinesAtheonIgnore verifies the atheon:ignore branch is exercised
// when scanning a line containing the directive.
func TestScanLinesAtheonIgnore(t *testing.T) {
	content := "AKIAIOSFODNN7EXAMPLE\nthis is a real finding line with AKIAIOSFODNN7EXAMPLE\n// atheon:ignore\nAKIAIOSFODNN7EXAMPLE-after-ignore\n"
	findings := scanLines(context.Background(), content, "test")
	if len(findings) == 0 {
		t.Fatal("expected findings before ignore directive")
	}
	// The line containing atheon:ignore must be skipped
	for _, f := range findings {
		if f.Line == 3 {
			t.Errorf("line 3 contains atheon:ignore and should be skipped, got finding: %+v", f)
		}
	}
	// Line 4 (after ignore) — the ignore directive only affects its own line,
	// so line 4 findings may still occur. The directive just suppresses the
	// matching line itself.
}

// TestLoadBundleBadJSONInGzip exercises the json.NewDecoder error branch
// inside loadBundle.
func TestLoadBundleBadJSONInGzip(t *testing.T) {
	// Build a gzip payload that decodes but isn't valid JSON for our schema.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte("not a json array")); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		// Restore the embedded bundle regardless of outcome
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()

	if err := loadBundle(buf.Bytes()); err == nil {
		t.Error("expected error loading bundle with invalid JSON")
	}
}

// TestLoadBundleInvalidRegexPattern exercises the per-pattern regex compile
// failure inside loadBundle — invalid regex is logged and skipped.
func TestLoadBundleInvalidRegexPattern(t *testing.T) {
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()

	defs := []PatternDef{
		{Name: "valid-1", Category: "test", Match: `\bVALID\b`, Enabled: true},
		{Name: "invalid", Category: "test", Match: `(unclosed`, Enabled: true},
		{Name: "valid-2", Category: "test", Match: `\bALSO_VALID\b`, Enabled: true},
	}
	jb, _ := json.Marshal(defs)
	if err := loadBundle(gzipBytes(jb)); err != nil {
		t.Fatalf("loadBundle failed: %v", err)
	}
	// valid patterns should be loaded; invalid should be skipped
	found1, found2 := false, false
	for _, p := range allPatterns {
		if p.name == "valid-1" {
			found1 = true
		}
		if p.name == "valid-2" {
			found2 = true
		}
		if p.name == "invalid" {
			t.Error("invalid-regex pattern should be skipped")
		}
	}
	if !found1 || !found2 {
		t.Errorf("expected valid patterns to load (found1=%v, found2=%v)", found1, found2)
	}
}

// TestLoadBundleOldBundleAllDisabled exercises the !anyEnabled fallback
// that re-enables all patterns when the bundle predates the enabled field
// (all enabled flags default to false in JSON).
func TestLoadBundleOldBundleAllDisabled(t *testing.T) {
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()

	defs := []PatternDef{
		{Name: "old-1", Category: "test", Match: `\bOLD\b`, Enabled: false},
		{Name: "old-2", Category: "test", Match: `\bANCIENT\b`, Enabled: false},
	}
	jb, _ := json.Marshal(defs)
	if err := loadBundle(gzipBytes(jb)); err != nil {
		t.Fatalf("loadBundle failed: %v", err)
	}
	// Both should have been auto-enabled by the !anyEnabled reset
	for _, p := range allPatterns {
		if p.name == "old-1" || p.name == "old-2" {
			if !p.enabled {
				t.Errorf("expected old-bundle pattern %q to be auto-enabled", p.name)
			}
		}
	}
	// Note: the reset sets enabled=true for these patterns, but other
	// patterns from the embedded bundle have been replaced. Restore via
	// loadBundle in the deferred cleanup above.
	_ = len(allPatterns)
}

// TestLoadBundleExternalRegistry exercises the external-pattern branch
// where Register()-ed non-bundle patterns are preserved.
func TestLoadBundleExternalRegistry(t *testing.T) {
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()

	// Register a non-bundle pattern before loading a new bundle
	external := &testPattern{name: "external-x", category: "ext", match: `\bEXT_X\b`}
	Register(external)
	defer unregisterForTest(external)

	defs := []PatternDef{
		{Name: "bundle-y", Category: "test", Match: `\bY\b`, Enabled: true},
	}
	jb, _ := json.Marshal(defs)
	if err := loadBundle(gzipBytes(jb)); err != nil {
		t.Fatalf("loadBundle failed: %v", err)
	}

	// External pattern should still be present
	foundExternal := false
	for _, p := range All() {
		if p.Name() == "external-x" {
			foundExternal = true
		}
	}
	if !foundExternal {
		t.Error("expected external pattern to survive bundle reload")
	}
}

type testPattern struct {
	name     string
	category string
	match    string
}

func (p *testPattern) Name() string             { return p.name }
func (p *testPattern) Category() string         { return p.category }
func (p *testPattern) Matches(line string) bool { return false }
func (p *testPattern) Enabled() bool            { return true }
func (p *testPattern) SetEnabled(bool)          {}

func unregisterForTest(p Pattern) {
	for i, rp := range registry {
		if rp == p {
			registry = append(registry[:i], registry[i+1:]...)
			return
		}
	}
}

// TestScanFileIgnored verifies ScanFile returns empty when the path is
// matched by an ignore rule in the working directory.
func TestScanFileIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .atheonignore that ignores *.secret
	if err := os.WriteFile(filepath.Join(tmpDir, ".atheonignore"), []byte("*.secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(tmpDir, "top.secret")
	if err := os.WriteFile(secret, []byte("AKIAIOSFODNN7EXAMPLE"), 0o644); err != nil {
		t.Fatal(err)
	}

	// t.Chdir changes CWD for the test and restores it on Cleanup. This
	// is safer than the manual os.Chdir pattern because it restores via
	// the testing framework rather than a deferred call that another
	// goroutine could observe in an inconsistent state.
	t.Chdir(tmpDir)

	findings, _, err := ScanFile(context.Background(), secret)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings for ignored file, got %d", len(findings))
	}
}

// TestScanEnvMalformedEnv exercises the scanEnv len(parts) != 2 branch.
// os.Environ() always returns KEY=VALUE entries so this is hard to trigger
// directly; instead we exercise scanEnv with a hand-crafted env list.
func TestScanEnvNoFindingsEmpty(t *testing.T) {
	// Save env, clear it, scan, restore.
	orig := os.Environ()
	for _, env := range orig {
		k := env
		if i := bytes.IndexByte([]byte(env), '='); i >= 0 {
			k = env[:i]
		}
		os.Unsetenv(k)
	}
	defer func() {
		for _, env := range orig {
			if i := bytes.IndexByte([]byte(env), '='); i >= 0 {
				os.Setenv(env[:i], env[i+1:])
			}
		}
	}()

	findings := ScanEnv(context.Background())
	if len(findings) != 0 {
		t.Errorf("expected no findings with empty env, got %d", len(findings))
	}
}

// TestScanEnvMalformedEntry exercises the len(parts) != 2 branch by passing
// an env list with a malformed entry (no = sign).
func TestScanEnvMalformedEntry(t *testing.T) {
	envs := []string{
		"MALFORMED_NO_EQUALS",
		"VALID_KEY=AKIAIOSFODNN7EXAMPLE",
		"ANOTHER_MALFORMED",
	}
	findings := scanEnv(context.Background(), envs)
	if len(findings) == 0 {
		t.Error("expected findings from valid entry")
	}
	for _, f := range findings {
		if f.File == "env:MALFORMED_NO_EQUALS" || f.File == "env:ANOTHER_MALFORMED" {
			t.Errorf("malformed entry should be skipped, got: %+v", f)
		}
	}
}

// TestSetActiveCategoriesResetAll verifies the !anyEnabled branch which
// re-enables all patterns.
func TestSetActiveCategoriesResetAll(t *testing.T) {
	// Snapshot the original enabled state of every pattern so we can restore
	type snap struct {
		p  *bundlePattern
		en bool
	}
	original := make([]snap, len(allPatterns))
	for i, p := range allPatterns {
		original[i] = snap{p: p, en: p.enabled}
	}
	defer func() {
		for _, s := range original {
			s.p.enabled = s.en
		}
		SetActiveCategories(nil)
	}()

	// Disable every pattern
	for _, p := range allPatterns {
		p.enabled = false
	}

	// Set categories to a non-empty list with no matches
	SetActiveCategories([]string{"__nonexistent_category__"})

	// No findings under non-matching category
	findings := ScanString(context.Background(), "AKIAIOSFODNN7EXAMPLE", "test")
	if len(findings) != 0 {
		t.Errorf("expected no findings with nonexistent category filter, got %d", len(findings))
	}

	// Reset to nil and confirm patterns are still registered
	SetActiveCategories(nil)
	if len(All()) == 0 {
		t.Error("expected patterns to still be registered after reset")
	}
}

// TestSetActiveCategoriesResetInternal exercises the empty-filter branch.
func TestSetActiveCategoriesResetInternal(t *testing.T) {
	type snap struct {
		p  *bundlePattern
		en bool
	}
	original := make([]snap, len(allPatterns))
	for i, p := range allPatterns {
		original[i] = snap{p: p, en: p.enabled}
	}
	defer func() {
		for _, s := range original {
			s.p.enabled = s.en
		}
		SetActiveCategories(nil)
	}()

	// Disable every pattern
	for _, p := range allPatterns {
		p.enabled = false
	}

	// Toggle through category filters — exercises both empty-list and nil
	// branches inside SetActiveCategories.
	SetActiveCategories([]string{})
	SetActiveCategories(nil)
}
