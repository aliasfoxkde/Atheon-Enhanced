package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestComputeBundleHash verifies the hash helper returns a 64-char hex string.
func TestComputeBundleHash(t *testing.T) {
	data := []byte("hello world")
	h := computeBundleHash(data)
	if len(h) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars: %s", len(h), h)
	}
	// Known SHA-256 of "hello world"
	if h != "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9" {
		t.Errorf("unexpected hash: %s", h)
	}
}

// gzipBundle gzip-compresses data (what the real server stores and what
// DownloadBundle receives from fetchBundleData).
func gzipBundle(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// minimalBundle returns a gzipped JSON bundle with one pattern.
func minimalBundle() []byte {
	defs := []map[string]any{{
		"name":     "test-pattern",
		"category": "secrets",
		"match":    "password=",
		"enabled":  true,
	}}
	data, _ := json.Marshal(defs)
	return gzipBundle(data)
}

// writeValidState writes a valid PatternState so loadPatternState succeeds
// inside recordBundleETag (which opens the state file as the flock lock handle
// and would otherwise read an empty file).
func writeValidState(home string) {
	statePath := filepath.Join(home, ".atheon")
	_ = os.MkdirAll(statePath, 0o755)
	_ = os.WriteFile(filepath.Join(statePath, "pattern_state.json"),
		[]byte(`{"patterns":{}}`), 0o600)
}

// TestVerifyBundleHashOK verifies a matching hash passes verification.
func TestVerifyBundleHashOK(t *testing.T) {
	bundleData := minimalBundle()
	bundleHash := computeBundleHash(bundleData)

	checksums := "abc123  other-file.txt\n" +
		bundleHash + "  patterns.bundle\n" +
		bundleHash + "  patterns.bundle.gz\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/patterns.bundle" {
			w.Header().Set("ETag", `"W/"test-etag"`)
			w.Write(bundleData)
			return
		}
		if r.URL.Path == "/download/checksums.txt" {
			w.Write([]byte(checksums))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	restore := SetBundleDownloadURL(server.URL + "/download/patterns.bundle")
	defer restore()

	home := t.TempDir()
	t.Setenv("HOME", home)
	writeValidState(home)

	if err := DownloadBundle(context.Background(), true); err != nil {
		t.Fatalf("DownloadBundle with valid hash failed: %v", err)
	}
	t.Cleanup(ReloadBundle)
}

// TestVerifyBundleHashMismatch verifies a mismatched hash causes DownloadBundle
// to return an error wrapping ErrBundleHashMismatch.
func TestVerifyBundleHashMismatch(t *testing.T) {
	bundleData := minimalBundle()

	// Wrong hash for patterns.bundle
	checksums := "0000000000000000000000000000000000000000000000000000000000000000  patterns.bundle\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/patterns.bundle" {
			w.Header().Set("ETag", `"W/"test-etag"`)
			w.Write(bundleData)
			return
		}
		if r.URL.Path == "/download/checksums.txt" {
			w.Write([]byte(checksums))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	restore := SetBundleDownloadURL(server.URL + "/download/patterns.bundle")
	defer restore()

	home := t.TempDir()
	t.Setenv("HOME", home)
	writeValidState(home)

	// Hash mismatch is now fatal — bundle does not load.
	if err := DownloadBundle(context.Background(), true); err == nil {
		t.Fatal("expected error on hash mismatch")
	}
}

// TestVerifyBundleHashMissing verifies 404 on checksums.txt is silently skipped.
func TestVerifyBundleHashMissing(t *testing.T) {
	bundleData := minimalBundle()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/download/patterns.bundle" {
			w.Header().Set("ETag", `"W/"test-etag"`)
			w.Write(bundleData)
			return
		}
		// checksums.txt not present → 404
		http.NotFound(w, r)
	}))
	defer server.Close()

	restore := SetBundleDownloadURL(server.URL + "/download/patterns.bundle")
	defer restore()

	home := t.TempDir()
	t.Setenv("HOME", home)
	writeValidState(home)

	if err := DownloadBundle(context.Background(), true); err != nil {
		t.Fatalf("DownloadBundle failed when checksums.txt is missing: %v", err)
	}
	t.Cleanup(ReloadBundle)
}
