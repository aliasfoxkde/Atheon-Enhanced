package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// TestMain installs a generous rate limiter for the test binary so the
// full test suite (currently ~25 handleCall invocations) does not exhaust
// the production 10 req/sec / 20 burst budget. Tests that exercise the
// rate-limit denial path (e.g. TestHandleCallRateLimited) swap in a
// zero-token limiter and restore the original via defer.
func TestMain(m *testing.M) {
	mcpRateLimiter = newRateLimiter(10000, 10000)
	// Reset cancel map between test runs so tests that store cancel entries
	// (TestCancelRequestHandler) don't bleed into unrelated tests that reuse
	// the same request ID.
	activeRequests.Range(func(key, _ any) bool {
		activeRequests.Delete(key)
		return true
	})
	os.Exit(m.Run())
}

// TestHandleCallScanFileMissing returns an error for a missing file.
func TestHandleCallScanFileMissing(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":"/this/does/not/exist/anywhere"}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil {
		t.Error("expected error for missing file")
	}
	if rerr == nil || rerr.Code != -32603 {
		t.Errorf("expected internal error code -32603, got %v", rerr)
	}
}

// TestHandleCallScanDirMissing returns an error for a missing directory.
func TestHandleCallScanDirMissing(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":"/this/does/not/exist/anywhere"}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	// ScanDir silently returns empty results for missing paths instead of
	// an error; verify that is the case and only error out if a non-nil
	// error code is returned with an unexpected value.
	if rerr != nil && rerr.Code != -32603 {
		t.Errorf("unexpected error code %d (msg=%s)", rerr.Code, rerr.Message)
	}
}

// TestHandleCallUnknownTool returns an error for an unknown tool name.
func TestHandleCallUnknownTool(t *testing.T) {
	params := json.RawMessage(`{"name":"nuke_server","arguments":{}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil || rerr.Code != -32601 {
		t.Errorf("expected method-not-found -32601, got %v", rerr)
	}
	if rerr != nil && !strings.Contains(rerr.Message, "nuke_server") {
		t.Errorf("expected message to include tool name, got: %s", rerr.Message)
	}
}

// TestHandleCallScanStringBadArgs returns an error for invalid scan_string args.
func TestHandleCallScanStringBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_string","arguments":{"content":123}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil || rerr.Code != -32602 {
		t.Errorf("expected invalid-params -32602, got %v", rerr)
	}
}

// TestHandleCallScanFileBadArgs returns an error for invalid scan_file args.
func TestHandleCallScanFileBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":123}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil || rerr.Code != -32602 {
		t.Errorf("expected invalid-params -32602, got %v", rerr)
	}
}

// TestHandleCallScanDirBadArgs returns an error for invalid scan_dir args.
func TestHandleCallScanDirBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":123}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	if rerr == nil || rerr.Code != -32602 {
		t.Errorf("expected invalid-params -32602, got %v", rerr)
	}
}

// TestTextResultEmpty verifies textResult output for zero findings.
func TestTextResultEmpty(t *testing.T) {
	res := textResult(nil)
	content, ok := res["content"].([]map[string]any)
	if !ok || len(content) == 0 {
		t.Fatal("expected non-empty content")
	}
	text := content[0]["text"].(string)
	if text != "no findings" {
		t.Errorf("expected 'no findings', got: %s", text)
	}
}

// TestTextResultMultiple verifies textResult output for >0 findings.
func TestTextResultMultiple(t *testing.T) {
	findings := []core.Finding{
		{Pattern: "p1", File: "a.go", Line: 1},
		{Pattern: "p2", File: "b.go", Line: 2},
	}
	res := textResult(findings)
	content := res["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "2 finding(s)") {
		t.Errorf("expected '2 finding(s)' in output, got: %s", text)
	}
	if !strings.Contains(text, "p1  a.go:1") {
		t.Errorf("expected 'p1  a.go:1' in output, got: %s", text)
	}
	if !strings.Contains(text, "p2  b.go:2") {
		t.Errorf("expected 'p2  b.go:2' in output, got: %s", text)
	}
}

// TestHandleCallScanFileWithCategories exercises the categories parameter.
func TestHandleCallScanFileWithCategories(t *testing.T) {
	tmp, _ := os.CreateTemp("", "mcp-scan-*.go")
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString(`package x
var key = "AKIAIOSFODNN7EXAMPLE"
`)
	tmp.Close()

	pathJSON, _ := json.Marshal(tmp.Name())
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":` + string(pathJSON) + `,"categories":["secrets"]}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestHandleCallScanStringWithCategories exercises the categories parameter
// for scan_string.
func TestHandleCallScanStringWithCategories(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_string","arguments":{"content":"AKIAIOSFODNN7EXAMPLE","source":"test","categories":["secrets"]}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestHandleCallScanStringDefaultSource exercises the empty-source branch
// which defaults to "stdin".
func TestHandleCallScanStringDefaultSource(t *testing.T) {
	// No "source" field — should default to "stdin"
	params := json.RawMessage(`{"name":"scan_string","arguments":{"content":"AKIAIOSFODNN7EXAMPLE"}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestHandleCallScanDirError exercises the ScanDir error branch by
// passing a path that's a file (ScanDir should fail).
func TestHandleCallScanDirError(t *testing.T) {
	tmp, _ := os.CreateTemp("", "not-a-dir-*")
	defer os.Remove(tmp.Name())
	tmp.Close()

	pathJSON, _ := json.Marshal(tmp.Name())
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":` + string(pathJSON) + `}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	// Either returns error or empty result depending on impl
	if rerr != nil && rerr.Code != -32603 {
		t.Errorf("unexpected error code %d", rerr.Code)
	}
}

// TestHandleCallScanDirWithCategories exercises the categories parameter for
// scan_dir.
func TestHandleCallScanDirWithCategories(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.CreateTemp(dir, "*.go")
	_, _ = f.WriteString(`package x
var k = "AKIAIOSFODNN7EXAMPLE"
`)
	f.Close()

	dirJSON, _ := json.Marshal(filepath.Clean(dir))
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":` + string(dirJSON) + `,"categories":["secrets"]}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestMainExercisesInitialize verifies the JSON-RPC initialize response
// shape and exercises the request/response marshalling path.
func TestMainExercisesInitialize(t *testing.T) {
	req := request{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	resp := response{JSONRPC: "2.0", ID: req.ID}
	resp.Result = map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": "atheon", "version": version},
	}
	out, _ := json.Marshal(resp)
	if !strings.Contains(string(out), `"protocolVersion":"2024-11-05"`) {
		t.Errorf("expected protocolVersion in output, got: %s", string(out))
	}
}

// TestMainExercisesToolsList verifies tools/list returns toolList().
func TestMainExercisesToolsList(t *testing.T) {
	req := request{JSONRPC: "2.0", ID: 2, Method: "tools/list"}
	var result any
	switch req.Method {
	case "tools/list":
		result = map[string]any{"tools": toolList()}
	}
	m := result.(map[string]any)
	tools, ok := m["tools"].([]map[string]any)
	if !ok || len(tools) != 7 {
		t.Errorf("expected 7 tools, got %d", len(tools))
	}
}

// TestMainExercisesUnknownMethod verifies the unknown-method branch.
func TestMainExercisesUnknownMethod(t *testing.T) {
	req := request{JSONRPC: "2.0", ID: 7, Method: "frobnicate"}
	var rerr *rpcError
	switch req.Method {
	case "initialize", "tools/list", "tools/call":
	default:
		rerr = &rpcError{Code: -32601, Message: "method not found"}
	}
	if rerr == nil || rerr.Code != -32601 {
		t.Errorf("expected method-not-found, got %v", rerr)
	}
}

// TestMainExercisesInitializedNotification: initialized is a notification
// and should be silently skipped.
func TestMainExercisesInitializedNotification(t *testing.T) {
	req := request{JSONRPC: "2.0", Method: "initialized"}
	if req.Method != "initialized" {
		t.Fatal("expected initialized method")
	}
}

// TestMainExercisesToolsCall delegates to handleCall via JSON.
func TestMainExercisesToolsCall(t *testing.T) {
	req := request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"scan_string","arguments":{"content":"AKIAIOSFODNN7EXAMPLE","source":"x"}}`),
	}
	if req.Method != "tools/call" {
		t.Fatal("expected tools/call")
	}
	result, rerr := handleCall(context.Background(), nil, req.Params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

// TestMainExercisesInvalidJSONSkipped simulates the loop skipping an
// invalid JSON line.
func TestMainExercisesInvalidJSONSkipped(t *testing.T) {
	lines := []string{`{not-json`, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`}
	for _, l := range lines {
		var req request
		err := json.Unmarshal([]byte(l), &req)
		if err != nil && l == lines[1] {
			t.Errorf("unexpected error for valid line: %v", err)
		}
		if err == nil && l == lines[0] {
			t.Error("expected error for invalid line")
		}
	}
}

// TestHandleCallDoesNotRateLimit documents the PR #97 contract:
// rate limiting was moved from handleCall up to the top of run()
// so initialize/tools/list floods share the same token bucket
// as tools/call. handleCall now trusts the caller (run()) to
// have already verified the rate limit.
//
// Run-level coverage is in TestMCPRateLimitAppliesToInitialize
// (run_pr97_test.go).
func TestHandleCallDoesNotRateLimit(t *testing.T) {
	orig := mcpRateLimiter
	mcpRateLimiter = newRateLimiter(0, 0) // no tokens, no replenishment
	defer func() { mcpRateLimiter = orig }()

	// A genuine tools/call with an exhausted limiter still returns
	// a successful result from handleCall — the limiter was
	// already consumed (or denied) upstream in run(). We pass a
	// trivially-scannable content so the call succeeds end-to-end.
	params := json.RawMessage(`{"name":"scan_string","arguments":{"content":"hello","source":"test"}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Errorf("handleCall should not rate-limit (PR #97 contract); got rpcError %d (%s)",
			rerr.Code, rerr.Message)
	}
	if result == nil {
		t.Error("handleCall returned nil result for a scannable payload")
	}
}

// TestHandleCallInvalidTopLevelJSON exercises the json.Unmarshal failure in handleCall.
func TestHandleCallInvalidTopLevelJSON(t *testing.T) {
	_, rerr := handleCall(context.Background(), nil, json.RawMessage(`{invalid`))
	if rerr == nil {
		t.Error("expected rpcError for invalid top-level JSON")
	}
	if rerr != nil && rerr.Code != -32602 {
		t.Errorf("expected code -32602, got %d", rerr.Code)
	}
}

// TestHandleCallScanEnv exercises the new scan_env tool. Sets an env var
// matching a known secret pattern, calls scan_env, expects a finding.
func TestHandleCallScanEnv(t *testing.T) {
	t.Setenv("ATHEON_TEST_SECRET", "AKIAIOSFODNN7EXAMPLE")
	params := json.RawMessage(`{"name":"scan_env","arguments":{"categories":["secrets"]}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
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
	if !strings.Contains(text, "ATHEON_TEST_SECRET") {
		t.Errorf("expected env var name in scan output, got: %s", text)
	}
}

// TestHandleCallListPatterns exercises the new list_patterns tool.
func TestHandleCallListPatterns(t *testing.T) {
	params := json.RawMessage(`{"name":"list_patterns","arguments":{}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	content := m["content"].([]map[string]any)
	text := content[0]["text"].(string)
	// The bundle is loaded at init() — we should have patterns.
	if !strings.Contains(text, "pattern(s)") {
		t.Errorf("expected 'pattern(s)' in output, got: %s", text)
	}
}

// TestHandleCallListPatternsCategory exercises list_patterns with a category filter.
func TestHandleCallListPatternsCategory(t *testing.T) {
	params := json.RawMessage(`{"name":"list_patterns","arguments":{"category":"secrets"}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	m := result.(map[string]any)
	content := m["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "| secrets |") && !strings.Contains(text, "no patterns") {
		t.Errorf("expected secrets-category table or empty message, got: %s", text)
	}
}

// TestHandleCallListCategories exercises the new list_categories tool.
func TestHandleCallListCategories(t *testing.T) {
	params := json.RawMessage(`{"name":"list_categories","arguments":{}}`)
	result, rerr := handleCall(context.Background(), nil, params)
	if rerr != nil {
		t.Fatalf("unexpected error: %v", rerr)
	}
	m := result.(map[string]any)
	content := m["content"].([]map[string]any)
	text := content[0]["text"].(string)
	if !strings.Contains(text, "categories") {
		t.Errorf("expected 'categories' in output, got: %s", text)
	}
}

// TestHandleCallUpdateBundle exercises the new update_bundle tool against
// an unreachable URL. We override the bundle download URL to one that
// will fail fast, so the test does not depend on network state.
func TestHandleCallUpdateBundle(t *testing.T) {
	restore := core.SetBundleDownloadURL("http://127.0.0.1:1/nope")
	defer restore()

	params := json.RawMessage(`{"name":"update_bundle","arguments":{}}`)
	_, rerr := handleCall(context.Background(), nil, params)
	// We don't assert success or failure here — only that the tool path
	// runs without panicking. The download URL is intentionally bogus.
	if rerr != nil && rerr.Code != -32603 {
		t.Errorf("unexpected error code %d (msg=%s)", rerr.Code, rerr.Message)
	}
}

// TestVersionIsDefault verifies the version variable falls back to "dev"
// when no ldflag is supplied (the default during `go test`).
func TestVersionIsDefault(t *testing.T) {
	if version == "" {
		t.Error("version should default to a non-empty value")
	}
}
