package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCommunity(t *testing.T, files map[string]string) (communityDir string) {
	t.Helper()
	tmp := t.TempDir()
	communityDir = filepath.Join(tmp, "community")
	if err := os.MkdirAll(communityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, content := range files {
		path := filepath.Join(communityDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return communityDir
}

func readBundle(t *testing.T, path string) []patternDef {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip open: %v", err)
	}
	defer r.Close()
	var defs []patternDef
	if err := json.NewDecoder(r).Decode(&defs); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	return defs
}

func TestBundleBasic(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/api-key.yaml": "name: test-api-key\nmatch: '\\bTEST_[A-Z0-9]{32}\\b'\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	n, err := bundle(community, out)
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 pattern, got %d", n)
	}

	defs := readBundle(t, out)
	if len(defs) != 1 {
		t.Fatalf("expected 1 def in bundle, got %d", len(defs))
	}
	d := defs[0]
	if d.Name != "test-api-key" {
		t.Errorf("name = %q, want test-api-key", d.Name)
	}
	if d.Category != "secrets" {
		t.Errorf("category = %q, want secrets", d.Category)
	}
	if !d.Enabled {
		t.Error("enabled should default to true")
	}
}

func TestBundleEnabledFalse(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/disabled.yaml": "name: disabled-key\nmatch: '\\bDIS_[A-Z0-9]{32}\\b'\nenabled: false\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	n, err := bundle(community, out)
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 pattern, got %d", n)
	}

	defs := readBundle(t, out)
	if defs[0].Enabled {
		t.Error("enabled should be false")
	}
}

func TestBundleMultipleCategories(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/key.yaml": "name: secret-key\nmatch: 'sk_[a-z0-9]{32}'\n",
		"tokens/jwt.yaml":  "name: jwt-token\nmatch: 'eyJ[A-Za-z0-9_-]+\\.eyJ'\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	n, err := bundle(community, out)
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 patterns, got %d", n)
	}
}

func TestBundleInvalidYAML(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/bad.yaml": "name: bad\nmatch: [unclosed\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	_, err := bundle(community, out)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestBundleMissingMatch(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/incomplete.yaml": "name: incomplete-key\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	_, err := bundle(community, out)
	if err == nil {
		t.Error("expected error for missing match field, got nil")
	}
}

func TestBundleMissingName(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/noname.yaml": "match: 'something'\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	_, err := bundle(community, out)
	if err == nil {
		t.Error("expected error for missing name field, got nil")
	}
}

func TestBundleEmptyDirectory(t *testing.T) {
	community := setupCommunity(t, map[string]string{})
	out := filepath.Join(t.TempDir(), "out.bundle")

	n, err := bundle(community, out)
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 patterns, got %d", n)
	}
}

func TestBundleBadOutputPath(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/key.yaml": "name: k\nmatch: 'x'\n",
	})
	_, err := bundle(community, "/nonexistent/dir/out.bundle")
	if err == nil {
		t.Error("expected error for bad output path, got nil")
	}
}

func TestBundleWhitespaceName(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/whitespace.yaml": "name: 'key with space'\nmatch: 'x'\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	_, err := bundle(community, out)
	if err == nil {
		t.Error("expected error for whitespace in pattern name, got nil")
	} else if !strings.Contains(err.Error(), "must not contain whitespace") {
		t.Errorf("expected 'must not contain whitespace' in error, got: %v", err)
	}
}

func TestBundleDuplicatePatternName(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/key1.yaml": "name: duplicate-key\nmatch: 'test1'\n",
		"secrets/key2.yaml": "name: duplicate-key\nmatch: 'test2'\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	_, err := bundle(community, out)
	if err == nil {
		t.Error("expected error for duplicate pattern name, got nil")
	} else if !strings.Contains(err.Error(), "duplicate pattern name") {
		t.Errorf("expected 'duplicate pattern name' in error, got: %v", err)
	}
}

func TestBundleInvalidRegex(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/bad.yaml": "name: bad-regex\nmatch: '[invalid'\n",
	})
	out := filepath.Join(t.TempDir(), "out.bundle")

	_, err := bundle(community, out)
	if err == nil {
		t.Error("expected error for invalid regex, got nil")
	} else if !strings.Contains(err.Error(), "invalid regex") {
		t.Errorf("expected 'invalid regex' in error, got: %v", err)
	}
}

// TestBundleToWriterGzipFailure exercises the writeGzipped error path in bundleToWriter.
type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errWriteFail }

var errWriteFail = errors.New("write failed")

func TestBundleToWriterGzipFailure(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/key.yaml": "name: test-key\nmatch: 'AKIAIOSFODNN7EXAMPLE'\n",
	})

	_, err := bundleToWriter(community, failWriter{})
	if err == nil {
		t.Error("expected error from failing writer, got nil")
	} else if !errors.Is(err, errWriteFail) {
		t.Errorf("expected errWriteFail sentinel, got: %v", err)
	}
}

// TestWalkPatternsWhitespaceName exercises the whitespace-in-name validation.
func TestWalkPatternsWhitespaceName(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/bad.yaml": "name: 'bad name'\nmatch: 'foo'\n",
	})

	_, err := walkPatterns(community)
	if err == nil {
		t.Error("expected error for pattern name with whitespace, got nil")
	}
}

// TestWalkPatternsMissingName exercises the missing-name-or-match validation.
func TestWalkPatternsMissingFields(t *testing.T) {
	community := setupCommunity(t, map[string]string{
		"secrets/empty.yaml": "name: ''\nmatch: 'foo'\n",
	})

	_, err := walkPatterns(community)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}
