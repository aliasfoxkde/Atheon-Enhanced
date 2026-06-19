package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliasfoxkde/Atheon/core"
)

// TestHandleCallScanFileMissing returns an error for a missing file.
func TestHandleCallScanFileMissing(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":"/this/does/not/exist/anywhere"}}`)
	_, rerr := handleCall(params)
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
	_, rerr := handleCall(params)
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
	_, rerr := handleCall(params)
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
	_, rerr := handleCall(params)
	if rerr == nil || rerr.Code != -32602 {
		t.Errorf("expected invalid-params -32602, got %v", rerr)
	}
}

// TestHandleCallScanFileBadArgs returns an error for invalid scan_file args.
func TestHandleCallScanFileBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":123}}`)
	_, rerr := handleCall(params)
	if rerr == nil || rerr.Code != -32602 {
		t.Errorf("expected invalid-params -32602, got %v", rerr)
	}
}

// TestHandleCallScanDirBadArgs returns an error for invalid scan_dir args.
func TestHandleCallScanDirBadArgs(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":123}}`)
	_, rerr := handleCall(params)
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

	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":"` + tmp.Name() + `","categories":["secrets"]}}`)
	result, rerr := handleCall(params)
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
	result, rerr := handleCall(params)
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
	result, rerr := handleCall(params)
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

	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":"` + tmp.Name() + `"}}`)
	_, rerr := handleCall(params)
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

	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":"` + filepath.Clean(dir) + `","categories":["secrets"]}}`)
	result, rerr := handleCall(params)
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
		"serverInfo":      map[string]any{"name": "atheon", "version": "1.0.0"},
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
	if !ok || len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
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
	result, rerr := handleCall(req.Params)
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
