package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestRunInitialize exercises the run() loop with an initialize request.
func TestRunInitialize(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2024-11-05"`) {
		t.Errorf("expected protocolVersion in output, got: %s", out.String())
	}
}

// TestRunToolsList exercises the tools/list request.
func TestRunToolsList(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"scan_string"`) {
		t.Errorf("expected scan_string tool in output, got: %s", out.String())
	}
	if !strings.Contains(out.String(), `"scan_file"`) {
		t.Errorf("expected scan_file tool in output")
	}
	if !strings.Contains(out.String(), `"scan_dir"`) {
		t.Errorf("expected scan_dir tool in output")
	}
}

// TestRunToolsCall exercises a tools/call request.
func TestRunToolsCall(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"scan_string","arguments":{"content":"AKIAIOSFODNN7EXAMPLE","source":"x"}}}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"result"`) {
		t.Errorf("expected result in output, got: %s", out.String())
	}
}

// TestRunUnknownMethod exercises the unknown-method error branch.
func TestRunUnknownMethod(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":4,"method":"frobnicate"}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"code":-32601`) {
		t.Errorf("expected method-not-found error code, got: %s", out.String())
	}
}

// TestRunInvalidJSONSkipped exercises the JSON-skip branch.
func TestRunInvalidJSONSkipped(t *testing.T) {
	in := strings.NewReader("{not-json\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	// No response should be emitted for invalid JSON
	if out.Len() != 0 {
		t.Errorf("expected no output for invalid JSON, got: %s", out.String())
	}
}

// TestRunInitializedNotification exercises the 'initialized' notification
// which is silently skipped (no response).
func TestRunInitializedNotification(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","method":"initialized"}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for 'initialized' notification, got: %s", out.String())
	}
}

// TestRunMultipleRequests exercises several requests in sequence.
func TestRunMultipleRequests(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"frobnicate"}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"scan_string","arguments":{"content":"AKIAIOSFODNN7EXAMPLE","source":"x"}}}`,
	}
	in := strings.NewReader(strings.Join(requests, "\n") + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	// Count responses — should be 4 (one per request)
	dec := json.NewDecoder(strings.NewReader(out.String()))
	count := 0
	for {
		var resp response
		if err := dec.Decode(&resp); err != nil {
			break
		}
		count++
	}
	if count != 4 {
		t.Errorf("expected 4 responses, got %d\n%s", count, out.String())
	}
}

// TestRunListPatterns exercises the list_patterns tool handler.
func TestRunListPatterns(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_patterns","arguments":{}}}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"result"`) {
		t.Errorf("expected result in output, got: %s", out.String())
	}
}

// TestRunListPatternsWithCategory exercises list_patterns with a category filter.
func TestRunListPatternsWithCategory(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_patterns","arguments":{"category":"secrets"}}}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"result"`) {
		t.Errorf("expected result in output, got: %s", out.String())
	}
}

// TestRunListCategories exercises the list_categories tool handler.
func TestRunListCategories(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_categories","arguments":{}}}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"result"`) {
		t.Errorf("expected result in output, got: %s", out.String())
	}
}

// TestRunUpdateBundle exercises the update_bundle tool handler.
// Note: This test uses a non-existent URL so it will fail, but it exercises
// the error path of the handler.
func TestRunUpdateBundle(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"update_bundle","arguments":{"force":false}}}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	// Exit code 0 even on error (handler returns error as JSON-RPC error)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	// Should contain either result or error
	if !strings.Contains(out.String(), `"result"`) && !strings.Contains(out.String(), `"error"`) {
		t.Errorf("expected result or error in output, got: %s", out.String())
	}
}
