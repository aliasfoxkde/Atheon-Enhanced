package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildTestBundleBytes returns a gzip-encoded JSON bundle containing the
// given pattern definitions. Used by DownloadBundle tests.
func buildTestBundleBytes(t *testing.T, defs []PatternDef) []byte {
	t.Helper()
	jb, err := json.Marshal(defs)
	if err != nil {
		t.Fatal(err)
	}
	return gzipBytes(jb)
}

// checksumsHandler returns an http.HandlerFunc that serves checksums.txt
// for a bundle of the given SHA-256 digest. The digest is computed from the
// bundle bytes already gzip-compressed and ready to serve.
func checksumsHandler(bundleBytes []byte) http.HandlerFunc {
	h := sha256.New()
	h.Write(bundleBytes)
	hash := h.Sum(nil)
	checksumLine := strings.TrimSpace(hex.EncodeToString(hash)) + "  patterns.bundle\n"
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") || r.URL.Path == "/checksums.txt" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(checksumLine)) //nolint:errcheck
			return
		}
	}
}

// TestDownloadBundleMockOK exercises the happy path with a mock server.
func TestDownloadBundleMockOK(t *testing.T) {
	// Build a small but valid bundle
	defs := []PatternDef{
		{Name: "dl-bundle-1", Category: "test", Match: `\bDL1\b`, Enabled: true},
		{Name: "dl-bundle-2", Category: "test", Match: `\bDL2\b`, Enabled: true},
	}
	body := buildTestBundleBytes(t, defs)
	checksums := checksumsHandler(body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checksums(w, r)
		if r.URL.Path != "checksums.txt" && r.URL.Path != "/checksums.txt" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	// Restore the embedded bundle after this test
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()

	if err := DownloadBundle(context.Background(), false); err != nil {
		t.Fatalf("DownloadBundle failed: %v", err)
	}

	// Verify new patterns are loaded
	found1, found2 := false, false
	for _, p := range allPatterns {
		if p.name == "dl-bundle-1" {
			found1 = true
		}
		if p.name == "dl-bundle-2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("expected downloaded patterns to be loaded (found1=%v, found2=%v)", found1, found2)
	}

	// Verify the bundle file was persisted to ~/.atheon/patterns.bundle
	home, _ := os.UserHomeDir()
	bundlePath := filepath.Join(home, ".atheon", "patterns.bundle")
	if _, err := os.Stat(bundlePath); err != nil {
		t.Errorf("expected bundle file at %s: %v", bundlePath, err)
	}
}

// TestDownloadBundleMockServerError exercises the non-200 branch.
func TestDownloadBundleMockServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") || r.URL.Path == "/checksums.txt" {
			// Serve a dummy checksums.txt so hash verification is not the failure.
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("deadbeef  patterns.bundle\n")) //nolint:errcheck
			return
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Fatal("expected error from server returning 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention 500, got: %v", err)
	}
}

// TestDownloadBundleMockBadGzip exercises the gzip-decode error branch
// inside DownloadBundle.
func TestDownloadBundleMockBadGzip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") || r.URL.Path == "/checksums.txt" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("deadbeef  patterns.bundle\n")) //nolint:errcheck
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not gzip data"))
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Fatal("expected error from server returning non-gzip data")
	}
}

// TestDownloadBundleMockBadJSON exercises the JSON-decode error branch.
func TestDownloadBundleMockBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") || r.URL.Path == "/checksums.txt" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("deadbeef  patterns.bundle\n")) //nolint:errcheck
			return
		}
		w.WriteHeader(http.StatusOK)
		// Gzip-compress invalid JSON
		_, _ = w.Write(gzipBytes([]byte("not json")))
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Fatal("expected error from server returning bad JSON")
	}
}

// TestDownloadBundleMockMkdirError exercises the MkdirAll error branch
// by setting HOME to a path that can't be created.
// Skipped on Windows: os.UserHomeDir uses USERPROFILE, not HOME env var.
func TestDownloadBundleMockMkdirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.UserHomeDir on Windows uses USERPROFILE, not HOME env var")
	}
	// Build a valid bundle so the early checks pass
	defs := []PatternDef{{Name: "x", Category: "t", Match: `x`, Enabled: true}}
	body := buildTestBundleBytes(t, defs)
	checksums := checksumsHandler(body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checksums(w, r)
		if r.URL.Path != "checksums.txt" && r.URL.Path != "/checksums.txt" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	// Point HOME at a path under a non-directory so MkdirAll fails
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", filepath.Join(blocker, "subdir"))

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Fatal("expected error when HOME points through a file")
	}
}

// TestDownloadBundleMockChangesReported exercises the added/removed
// summary branches by serving a bundle with different patterns.
func TestDownloadBundleMockChangesReported(t *testing.T) {
	// Ensure some patterns are loaded
	defer func() {
		_ = loadBundle(embeddedBundle)
		SetActiveCategories(nil)
	}()
	if len(allPatterns) == 0 {
		t.Skip("no baseline patterns")
	}

	// Build a bundle with one new pattern and one removed pattern
	defs := []PatternDef{
		{Name: "newly-added-pattern", Category: "test", Match: `\bNEW\b`, Enabled: true},
	}
	body := buildTestBundleBytes(t, defs)
	checksums := checksumsHandler(body)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checksums(w, r)
		if r.URL.Path != "checksums.txt" && r.URL.Path != "/checksums.txt" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		}
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	if err := DownloadBundle(context.Background(), false); err != nil {
		t.Fatalf("DownloadBundle failed: %v", err)
	}
}

// TestDownloadBundleNetworkError exercises the client.Get error branch
// by pointing the URL at an unroutable address.
func TestDownloadBundleNetworkError(t *testing.T) {
	// Use a closed httptest server URL — port won't accept connections
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // immediately close so the URL is invalid

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Error("expected network error from closed server URL")
	}
}

// TestDownloadBundleMockMkdirErrorUserProfile exercises the MkdirAll error
// branch in ensureAtheonDir on Windows by setting USERPROFILE to a file path.
func TestDownloadBundleMockMkdirErrorUserProfile(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("USERPROFILE-based test is Windows-only; HOME-based test covers other platforms")
	}
	defs := []PatternDef{{Name: "mkdir-err-win", Category: "t", Match: `x`, Enabled: true}}
	body := buildTestBundleBytes(t, defs)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// USERPROFILE is what os.UserHomeDir() uses on Windows.
	// Setting it to a file path makes os.MkdirAll(blocker/.atheon) fail.
	t.Setenv("USERPROFILE", blocker)

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Fatal("expected error when USERPROFILE points to a file")
	}
}

// TestFetchBundleDataInvalidURL exercises the http.NewRequestWithContext
// error branch by setting the URL to one with an invalid character.
func TestFetchBundleDataInvalidURL(t *testing.T) {
	// A URL containing a null byte (\x00) is rejected by http.NewRequestWithContext.
	restore := SetBundleDownloadURL("http://\x00invalid-url")
	defer restore()

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Error("expected error from http.NewRequestWithContext with invalid URL")
	}
}

// TestDownloadBundleReadAllError exercises the io.ReadAll error branch by
// using a server that returns chunked transfer encoding and closes mid-stream.
func TestDownloadBundleReadAllError(t *testing.T) {
	// Hijack the connection and abort it without writing a complete body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("ResponseWriter is not a Hijacker")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		// Write a partial response then close abruptly so ReadAll fails.
		_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\nshort"))
		_ = conn.Close()
	}))
	defer srv.Close()

	restore := SetBundleDownloadURL(srv.URL + "/")
	defer restore()

	err := DownloadBundle(context.Background(), false)
	if err == nil {
		t.Error("expected error from server with truncated body")
	}
}
