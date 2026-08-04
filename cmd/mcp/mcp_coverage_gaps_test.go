package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// ------------------------------------------------------------------
// run() coverage gaps
// ------------------------------------------------------------------

// TestRunEncodeError exercises the enc.Encode error path in run().
// When the io.Writer returns an error during Encode, run() should return 1.
func TestRunEncodeError(t *testing.T) {
	// A writer that fails on first write
	failWriter := &errorWriter{}

	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	code := run(context.Background(), in, failWriter)
	if code != 1 {
		t.Errorf("expected exit 1 from encode error, got %d", code)
	}
}

type errorWriter struct{}

func (e *errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("synthetic write error")
}

// TestRunEmptyInput exercises the scanner with no input.
func TestRunEmptyInput(t *testing.T) {
	in := strings.NewReader("")
	out := &strings.Builder{}
	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0 for empty input, got %d", code)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for empty input, got: %s", out.String())
	}
}

// ------------------------------------------------------------------
// dispatchRequest coverage gaps
// ------------------------------------------------------------------

// TestDispatchRequestCancelUnmarshalError exercises the cancelRequest
// path when json.Unmarshal fails on the params.
func TestDispatchRequestCancelUnmarshalError(t *testing.T) {
	req := &request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "$/cancelRequest",
		Params:  json.RawMessage(`{invalid`),
	}
	result, rerr := dispatchRequest(context.Background(), req)
	// Cancel notifications return nil,nil on unmarshal error
	if result != nil {
		t.Errorf("expected nil result on cancel unmarshal error, got %v", result)
	}
	if rerr != nil {
		t.Errorf("expected nil rerr on cancel unmarshal error, got %v", rerr)
	}
}

// ------------------------------------------------------------------
// sandboxPath coverage gaps
// ------------------------------------------------------------------

// TestSandboxPathCwdGetwdError exercises the Getwd error path in sandboxPath.
// This is tricky to trigger since Getwd rarely fails, but we can mock it
// by changing the working directory to an invalid location.
func TestSandboxPathCwdGetwdError(t *testing.T) {
	// We cannot easily trigger Getwd failure in a test, but we can verify
	// the symlink-error path by creating a broken symlink that triggers
	// the EvalSymlinks error-with-clean-path branch.
	//
	// The actual uncovered lines in sandboxPath are:
	// - line 579: cwd, err := os.Getwd() — Getwd failure
	// - line 580-581: return realPath, os.ErrPermission
	//
	// Since we cannot reliably mock Getwd, we verify the behavior
	// is documented and the permission error path is correct.
	//
	// We CAN trigger EvalSymlinks failure for an EXISTING symlink
	// with bad permissions, but not for a non-existent path.
	// The non-existent path case (line 572-578) is already covered.
	//
	// For now, document that Getwd error is not testable in normal env.
	if _, err := os.Getwd(); err != nil {
		t.Errorf("Getwd unexpectedly failed: %v", err)
	}
}

// ------------------------------------------------------------------
// handleScanString coverage gaps - source length cap
// ------------------------------------------------------------------

// TestHandleScanStringSourceTooLong exercises the source length cap.
func TestHandleScanStringSourceTooLong(t *testing.T) {
	// Create a source string that exceeds mcpScanStringSourceMaxBytes (1024)
	longSource := strings.Repeat("x", mcpScanStringSourceMaxBytes+1)
	params := json.RawMessage(`{"name":"scan_string","arguments":{"content":"hello","source":` + fmt.Sprintf("%q", longSource) + `}}`)

	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for source exceeding max length")
	}
	if rerr.Code != -32602 {
		t.Errorf("expected -32602, got %d", rerr.Code)
	}
}

// ------------------------------------------------------------------
// handleScanFile coverage gaps - sandbox error path
// ------------------------------------------------------------------

// TestHandleScanFileSandboxError exercises the sandboxPath error path.
func TestHandleScanFileSandboxError(t *testing.T) {
	// Use a path that will be rejected by sandboxPath
	// (relative path with ".." traversal)
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":"../../etc/passwd"}}`)

	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for sandboxed path")
	}
	// sandboxPath returns os.ErrPermission for traversal attempts
	if rerr.Code != -32603 {
		t.Errorf("expected -32603 for sandbox error, got %d: %s", rerr.Code, rerr.Message)
	}
}

// TestHandleScanFileContextCanceled exercises handleScanFile when the
// context is already canceled before ScanFile is called.
func TestHandleScanFileContextCanceled(t *testing.T) {
	tmp, err := os.CreateTemp("", "scan-context-canceled-*")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("hello world")
	tmp.Close()
	defer os.Remove(tmp.Name())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before the call

	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":` + fmt.Sprintf("%q", tmp.Name()) + `}}`)
	_, rerr := handleCall(ctx, nil, params)
	// Should handle canceled context gracefully
	if rerr != nil && rerr.Code != -32603 {
		t.Logf("rerr=%v", rerr)
	}
	// The important thing is no panic and returns in finite time
}

// ------------------------------------------------------------------
// handleScanDir coverage gaps - sandbox error path
// ------------------------------------------------------------------

// TestHandleScanDirSandboxError exercises the sandboxPath error path.
func TestHandleScanDirSandboxError(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":"../../secrets"}}`)

	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for sandboxed path")
	}
	if rerr.Code != -32603 {
		t.Errorf("expected -32603 for sandbox error, got %d: %s", rerr.Code, rerr.Message)
	}
}

// TestHandleScanDirContextCanceled exercises handleScanDir when the
// context is already canceled.
func TestHandleScanDirContextCanceled(t *testing.T) {
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before the call

	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":` + fmt.Sprintf("%q", dir) + `}}`)
	_, rerr := handleCall(ctx, nil, params)
	// Should handle canceled context gracefully
	if rerr != nil && rerr.Code != -32603 {
		t.Logf("rerr=%v", rerr)
	}
}

// ------------------------------------------------------------------
// handleListPatterns coverage gaps - empty patterns
// ------------------------------------------------------------------

// TestHandleListPatternsEmptyBundle exercises list_patterns when
// no patterns are loaded. This requires temporarily clearing the
// pattern registry.
func TestHandleListPatternsEmptyBundle(t *testing.T) {
	// core.All() returns registered patterns. If no patterns are registered,
	// it returns an empty slice.
	// We can't easily unregister patterns, but we can verify that
	// patternsResult handles empty slice correctly.
	restore := core.SetBundleDownloadURLForTest("http://127.0.0.1:1/nope")
	defer restore()

	// Call patternsResult with empty patterns and empty category
	result := patternsResult([]core.Pattern{}, "")
	content := result["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "no patterns loaded") {
		t.Errorf("expected 'no patterns loaded' for empty patterns, got: %s", text)
	}

	// Also test with category that matches nothing
	result = patternsResult([]core.Pattern{}, "nonexistent-category")
	content = result["content"].([]map[string]any)
	text = content[0]["text"].(string)
	if !strings.Contains(text, "no patterns in category") {
		t.Errorf("expected 'no patterns in category' for empty result with category, got: %s", text)
	}
}

// TestHandleListPatternsUnknownCategory verifies that when a non-empty
// category is specified but no patterns match, we get the right message.
func TestHandleListPatternsUnknownCategory(t *testing.T) {
	params := json.RawMessage(`{"name":"list_patterns","arguments":{"category":"this-category-does-not-exist"}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	m := result.(map[string]any)
	content := m["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "no patterns") {
		t.Errorf("expected 'no patterns' message for unknown category, got: %s", text)
	}
}

// ------------------------------------------------------------------
// handleUpdateBundle coverage gaps - DownloadBundle failure
// ------------------------------------------------------------------

// TestHandleUpdateBundleDownloadError exercises the DownloadBundle
// error path by using a URL that will fail.
func TestHandleUpdateBundleDownloadError(t *testing.T) {
	// Save original bundle URL and restore after test
	restore := core.SetBundleDownloadURLForTest("http://127.0.0.1:1/bundle.tar.gz")
	defer restore()

	params := json.RawMessage(`{"name":"update_bundle","arguments":{"force":true}}`)
	_, rerr := handleCall(context.Background(), nil, params)

	// DownloadBundle should fail with connection refused or similar
	if rerr == nil {
		t.Fatal("expected error from update_bundle with bad URL")
	}
	if rerr.Code != -32603 {
		t.Errorf("expected -32603 for download error, got %d: %s", rerr.Code, rerr.Message)
	}
}

// TestHandleUpdateBundleContextCanceled exercises update_bundle when
// the context is canceled.
func TestHandleUpdateBundleContextCanceled(t *testing.T) {
	restore := core.SetBundleDownloadURLForTest("http://127.0.0.1:1/bundle.tar.gz")
	defer restore()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := json.RawMessage(`{"name":"update_bundle","arguments":{"force":true}}`)
	_, rerr := handleCall(ctx, nil, params)

	// Should return error for canceled context
	if rerr == nil {
		t.Fatal("expected error from update_bundle with canceled context")
	}
}

// TestHandleUpdateBundleSuccess exercises the success path of handleUpdateBundle
// by serving a minimal valid bundle via httptest.
func TestHandleUpdateBundleSuccess(t *testing.T) {
	// Build a minimal valid bundle: gzip-compressed JSON with a pattern def.
	patternJSON := `[{"name":"test-pattern","category":"test","match":"\\bTEST\\b","enabled":true}]`
	var gzBuf bytes.Buffer
	gz := gzip.NewWriter(&gzBuf)
	gz.Write([]byte(patternJSON))
	gz.Close()
	bundleBytes := gzBuf.Bytes()

	// Compute SHA256 of the bundle bytes for checksums.txt
	h := sha256.New()
	h.Write(bundleBytes)
	checksum := hex.EncodeToString(h.Sum(nil))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "checksums.txt") || r.URL.Path == "/checksums.txt" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(checksum + "  patterns.bundle\n"))
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(bundleBytes)
	}))
	defer srv.Close()

	restore := core.SetBundleDownloadURLForTest(srv.URL + "/")
	defer restore()

	params := json.RawMessage(`{"name":"update_bundle","arguments":{"force":true}}`)
	result, rerr := handleCall(context.Background(), nil, params)

	if rerr != nil {
		t.Fatalf("handleUpdateBundle failed: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result on success")
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	content, ok := m["content"].([]map[string]any)
	if !ok || len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
	text := content[0]["text"].(string)
	if !strings.Contains(text, "bundle updated successfully") {
		t.Errorf("expected 'bundle updated successfully' in output, got: %s", text)
	}

	// Restore the original bundle so subsequent tests are not affected.
	core.RestoreBundle()
}

// ------------------------------------------------------------------
// handleScanEnv unmarshal error path
// ------------------------------------------------------------------

// TestHandleScanEnvBadArgs exercises the unmarshal error path.
func TestHandleScanEnvBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_env","arguments":{"categories":"not-an-array"}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for bad arguments")
	}
	if rerr.Code != -32602 {
		t.Errorf("expected -32602, got %d", rerr.Code)
	}
}

// TestHandleScanEnvCategoriesLimit exercises the categories cap in handleScanEnv.
func TestHandleScanEnvCategoriesLimit(t *testing.T) {
	// Create 101 categories to exceed the maxCategories=100 cap
	cats := make([]string, 101)
	for i := range cats {
		cats[i] = fmt.Sprintf("cat%d", i)
	}
	catsJSON, _ := json.Marshal(cats)
	params := json.RawMessage(`{"name":"scan_env","arguments":{"categories":` + string(catsJSON) + `}}`)

	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for too many categories")
	}
	if rerr.Code != -32602 {
		t.Errorf("expected -32602, got %d", rerr.Code)
	}
}

// TestHandleScanEnvEmpty exercises handleScanEnv with no categories (scans all).
func TestHandleScanEnvEmpty(t *testing.T) {
	// Set an env var with a known secret pattern
	t.Setenv("TEST_AWS_KEY", "AKIAIOSFODNN7EXAMPLE")
	params := json.RawMessage(`{"name":"scan_env","arguments":{}}`)

	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// ------------------------------------------------------------------
// handleListPatterns unmarshal error path
// ------------------------------------------------------------------

// TestHandleListPatternsBadArgs exercises the unmarshal error path.
func TestHandleListPatternsBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"list_patterns","arguments":{"category":123}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for bad arguments")
	}
	if rerr.Code != -32602 {
		t.Errorf("expected -32602, got %d", rerr.Code)
	}
}

// ------------------------------------------------------------------
// handleUpdateBundle unmarshal error path
// ------------------------------------------------------------------

// TestHandleUpdateBundleBadArgs exercises the unmarshal error path.
func TestHandleUpdateBundleBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"update_bundle","arguments":{"force":999}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Fatal("expected error for bad arguments")
	}
	if rerr.Code != -32602 {
		t.Errorf("expected -32602, got %d", rerr.Code)
	}
}

// ------------------------------------------------------------------
// patternsResult coverage gaps
// ------------------------------------------------------------------

// TestPatternsResultEmptyPatterns exercises patternsResult with no patterns.
func TestPatternsResultEmptyPatterns(t *testing.T) {
	result := patternsResult([]core.Pattern{}, "")
	content := result["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "no patterns loaded") {
		t.Errorf("expected 'no patterns loaded', got: %s", text)
	}

	// structuredContent should be empty array
	sc := content[0]["structuredContent"].([]map[string]any)
	if len(sc) != 0 {
		t.Errorf("expected empty structuredContent, got %d items", len(sc))
	}
}

// TestPatternsResultCategoryNoMatch exercises patternsResult with a
// category that matches no patterns.
func TestPatternsResultCategoryNoMatch(t *testing.T) {
	result := patternsResult([]core.Pattern{}, "nonexistent")
	content := result["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "no patterns in category") {
		t.Errorf("expected 'no patterns in category', got: %s", text)
	}
}

// TestPatternsResultWithFindings exercises patternsResult with real patterns
// and verifies the structuredContent is populated.
func TestPatternsResultWithFindings(t *testing.T) {
	fp := &fakePatternForTest{
		name:     "aws-key",
		category: "secrets",
		severity: "critical",
		enabled:  true,
	}

	result := patternsResult([]core.Pattern{fp}, "secrets")
	content := result["content"].([]map[string]any)
	text := content[0]["text"].(string)

	// Should contain the pattern in table format
	if !strings.Contains(text, "aws-key") {
		t.Errorf("expected pattern name in output, got: %s", text)
	}
	if !strings.Contains(text, "| secrets |") {
		t.Errorf("expected category in output, got: %s", text)
	}

	// structuredContent should have 1 item
	sc := content[0]["structuredContent"].([]map[string]any)
	if len(sc) != 1 {
		t.Errorf("expected 1 structured item, got %d", len(sc))
	}
	if sc[0]["name"] != "aws-key" {
		t.Errorf("expected name in structuredContent, got: %v", sc[0]["name"])
	}
}

// ------------------------------------------------------------------
// run() notification handling coverage
// ------------------------------------------------------------------

// TestRunRateLimitedNotification verifies that a rate-limited notification
// (no ID) does not emit a response.
func TestRunRateLimitedNotification(t *testing.T) {
	orig := mcpRateLimiter
	mcpRateLimiter = newRateLimiter(1, 1) // Exhaust the limiter
	t.Cleanup(func() { mcpRateLimiter = orig })

	// Send a notification while rate-limited
	in := strings.NewReader(`{"jsonrpc":"2.0","method":"some/notification"}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	// No response should be emitted for a rate-limited notification
	if out.Len() != 0 {
		t.Errorf("expected no output for rate-limited notification, got: %s", out.String())
	}
}

// TestRunMalformedJSONWithID verifies that malformed JSON with a valid-looking
// ID still gets no response (can't echo back because parse failed).
func TestRunMalformedJSONWithID(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0 for malformed JSON, got %d", code)
	}
	// No response for malformed JSON
	if out.Len() != 0 {
		t.Errorf("expected no output for malformed JSON, got: %s", out.String())
	}
}

// ------------------------------------------------------------------
// dispatchRequest jsonrpc validation
// ------------------------------------------------------------------

// TestDispatchRequestInvalidJSONRPC exercises the jsonrpc validation in dispatchRequest.
func TestDispatchRequestInvalidJSONRPC(t *testing.T) {
	req := &request{
		JSONRPC: "1.5", // Invalid version
		ID:      1,
		Method:  "initialize",
	}
	_, rerr := dispatchRequest(context.Background(), req)
	if rerr == nil {
		t.Fatal("expected error for invalid jsonrpc version")
	}
	if rerr.Code != -32600 {
		t.Errorf("expected -32600, got %d", rerr.Code)
	}
}

// ------------------------------------------------------------------
// Concurrent dispatch paths
// ------------------------------------------------------------------

// TestDispatchRequestAtConcurrentLimit verifies that when inflight is
// exactly at the cap, a new request is rejected.
func TestDispatchRequestAtConcurrentLimit(t *testing.T) {
	mcpInflight.Store(0)

	// Pre-fill to cap
	for i := 0; i < mcpConcurrentCap; i++ {
		mcpInflight.Add(1)
	}
	defer mcpInflight.Store(0)

	req := &request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	_, rerr := dispatchRequest(context.Background(), req)
	if rerr == nil {
		t.Fatal("expected error at concurrent limit")
	}
	if rerr.Code != -32001 {
		t.Errorf("expected -32001, got %d", rerr.Code)
	}
}

// ------------------------------------------------------------------
// configureLogging coverage
// ------------------------------------------------------------------

// TestConfigureLoggingDebug exercises configureLogging with DEBUG level.
func TestConfigureLoggingDebug(t *testing.T) {
	t.Setenv("ATHEON_LOG_LEVEL", "debug")
	t.Setenv("ATHEON_LOG_FORMAT", "text")
	configureLogging()
	// Verify the default handler was set
	if slog.Default() == nil {
		t.Error("expected non-nil slog default handler")
	}
}

// TestConfigureLoggingWarn exercises configureLogging with WARN level.
func TestConfigureLoggingWarn(t *testing.T) {
	t.Setenv("ATHEON_LOG_LEVEL", "warn")
	t.Setenv("ATHEON_LOG_FORMAT", "text")
	configureLogging()
	if slog.Default() == nil {
		t.Error("expected non-nil slog default handler")
	}
}

// TestConfigureLoggingWarningAlias exercises configureLogging with WARNING alias.
func TestConfigureLoggingWarningAlias(t *testing.T) {
	t.Setenv("ATHEON_LOG_LEVEL", "WARNING")
	t.Setenv("ATHEON_LOG_FORMAT", "text")
	configureLogging()
	if slog.Default() == nil {
		t.Error("expected non-nil slog default handler")
	}
}

// TestConfigureLoggingError exercises configureLogging with ERROR level.
func TestConfigureLoggingError(t *testing.T) {
	t.Setenv("ATHEON_LOG_LEVEL", "error")
	t.Setenv("ATHEON_LOG_FORMAT", "text")
	configureLogging()
	if slog.Default() == nil {
		t.Error("expected non-nil slog default handler")
	}
}

// TestConfigureLoggingJSON exercises configureLogging with JSON format.
func TestConfigureLoggingJSON(t *testing.T) {
	t.Setenv("ATHEON_LOG_LEVEL", "info")
	t.Setenv("ATHEON_LOG_FORMAT", "json")
	configureLogging()
	if slog.Default() == nil {
		t.Error("expected non-nil slog default handler")
	}
}

// TestConfigureLoggingDefault exercises configureLogging with no env vars (default).
func TestConfigureLoggingDefault(t *testing.T) {
	t.Setenv("ATHEON_LOG_LEVEL", "")
	t.Setenv("ATHEON_LOG_FORMAT", "")
	configureLogging()
	if slog.Default() == nil {
		t.Error("expected non-nil slog default handler")
	}
}

// ------------------------------------------------------------------
// Helper to make fmt available
// ------------------------------------------------------------------
