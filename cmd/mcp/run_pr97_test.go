package main

// Tests for PR #97 (Wave 8): MCP server hardening — panic recovery,
// rate-limit-all-methods, size caps, per-request timeout, error log,
// and notification / JSONRPC validation.
//
// Conventions:
//   - All tests feed NDJSON into run(ctx, in, out) via strings.Reader.
//   - Tests use the global mcpRateLimiter; where burst-size matters
//     we swap it out for a deterministic one and restore via t.Cleanup.
//   - Tests are package-internal so they can reach toolHandlers,
//     dispatchRequest, and the size-cap constants directly.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestMCPPanicRecovered installs a panicking tool handler, sends a
// tools/call through the run() loop, and asserts that:
//
//	(a) the panic is converted to an -32603 (internal error) response, and
//	(b) the server continues serving — the next initialize returns 0.
//
// (b) is the actual security property: pre-PR-#97 a single panicking
// handler tore down the whole process and forced every connected
// client to reconnect. The defer recover() in dispatchRequest is what
// keeps one bad request from killing every session.
func TestMCPPanicRecovered(t *testing.T) {
	const panicTool = "__pr97_panic_tool__"

	prevHandler, existed := toolHandlers[panicTool]
	toolHandlers[panicTool] = func(_ context.Context, _ json.RawMessage) (any, *rpcError) {
		panic("synthetic panic for TestMCPPanicRecovered")
	}
	t.Cleanup(func() {
		delete(toolHandlers, panicTool)
		if existed {
			toolHandlers[panicTool] = prevHandler
		}
	})

	// First request: panicking tools/call.
	first := `{"jsonrpc":"2.0","id":99,"method":"tools/call","params":{"name":"` + panicTool + `","arguments":{}}}` + "\n"
	// Second request: a normal initialize to confirm the server still
	// serves after the panic.
	second := `{"jsonrpc":"2.0","id":100,"method":"initialize"}` + "\n"

	out := &strings.Builder{}
	code := run(context.Background(), strings.NewReader(first+second), out)
	if code != 0 {
		t.Errorf("expected run() to survive panic (exit 0), got %d", code)
	}
	outStr := out.String()
	if !strings.Contains(outStr, `"code":-32603`) {
		t.Errorf("expected -32603 internal error in panic response, got: %s", outStr)
	}
	if !strings.Contains(outStr, `"protocolVersion"`) {
		t.Errorf("expected server to recover and serve initialize after panic, got: %s", outStr)
	}
}

// TestMCPDispatchRecoverIsolated exercises dispatchRequest directly
// with a panicking inline function. Complements TestMCPPanicRecovered
// (which goes through the run() loop) and proves the defer recover()
// sits inside dispatchRequest, not just around the JSON-decoder level.
func TestMCPDispatchRecoverIsolated(t *testing.T) {
	// We can't substitute a method handler without touching toolHandlers,
	// so we trigger a panic via the toolHandlers map the same way
	// TestMCPPanicRecovered does — the dispatch-level assertion is that
	// result is nil and rerr has the expected code.
	const panicTool = "__pr97_panic_dispatch__"
	toolHandlers[panicTool] = func(_ context.Context, _ json.RawMessage) (any, *rpcError) {
		panic("dispatch-level panic")
	}
	t.Cleanup(func() { delete(toolHandlers, panicTool) })

	req := &request{JSONRPC: "2.0", ID: 1, Method: "tools/call"}
	params := json.RawMessage(fmt.Sprintf(`{"name":%q,"arguments":{}}`, panicTool))

	result, rerr := dispatchRequest(context.Background(), req)
	// For tools/call the params must be a real envelope — go through
	// handleCall instead, since dispatchRequest forwards to it.
	_ = result
	_ = rerr
	_ = params

	// Now go through the proper dispatchRequest → handleCall path
	// by passing a params envelope that references our panic handler.
	// dispatchRequest for tools/call delegates to handleCall(ctx, req.ID, req.Params).
	req.Params = params
	result, rerr = dispatchRequest(context.Background(), req)
	if result != nil {
		t.Errorf("expected nil result on panic, got %v", result)
	}
	if rerr == nil {
		t.Fatal("expected rpcError on panic, got nil")
	}
	if rerr.Code != -32603 {
		t.Errorf("expected panic to surface as -32603, got %d (%s)", rerr.Code, rerr.Message)
	}
}

// TestMCPRateLimitAppliesToInitialize verifies the rate limiter sits
// at the TOP of run(), so initialize floods count against the same
// token bucket as tools/call. Pre-PR-#97 only tools/call was
// throttled, so an attacker spamming initialize could DoS the server
// without consuming any tokens.
//
// We swap in a deterministic limiter (burst 1) so the test doesn't
// depend on test ordering or global state.
func TestMCPRateLimitAppliesToInitialize(t *testing.T) {
	orig := mcpRateLimiter
	t.Cleanup(func() { mcpRateLimiter = orig })
	mcpRateLimiter = newRateLimiter(1, 1) // 1 token/sec, burst 1

	in := strings.NewReader(strings.Repeat(
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`+"\n", 5))
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"code":-32000`) {
		t.Errorf("expected rate-limit error (-32000) for initialize flood, got: %s", out.String())
	}
}

// TestMCPLargeRequestRejected feeds a JSON request larger than
// mcpMaxRequestBytes (64 MiB) through run() and verifies that:
//   - the decoder doesn't OOM the process (io.LimitReader caps reads)
//   - the run loop doesn't panic
//   - no partial / truncated response is emitted
//
// Pre-PR-#97 the bufio.Scanner-based reader had a 1 MiB buffer and
// silently errored past it; the response behavior on truncation was
// undefined. With io.LimitReader we have a hard memory ceiling.
//
// Note on exit code: the run loop's stuck-input detector returns 1
// after consecutive parse errors with no InputOffset progress —
// broken connection is a non-zero exit so a supervisor (systemd,
// Docker) can restart the daemon cleanly. We accept either 0 or 1
// here because both mean "didn't panic, didn't emit garbage".
func TestMCPLargeRequestRejected(t *testing.T) {
	// 70 MiB of padding exceeds mcpMaxRequestBytes (64 MiB) by enough
	// margin that even after JSON-escaping the value would still be
	// truncated. We don't try to make valid JSON here — we want the
	// decoder to fail and the run loop to log + continue.
	var sb strings.Builder
	sb.WriteString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"junk":"`)
	sb.WriteString(strings.Repeat("x", 70<<20))
	sb.WriteString(`"}}`)

	out := &strings.Builder{}
	code := run(context.Background(), strings.NewReader(sb.String()), out)
	if code != 0 && code != 1 {
		t.Errorf("expected exit 0 or 1 for oversized request, got %d", code)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for oversized request, got %d bytes: %s",
			out.Len(), out.String()[:min(200, out.Len())])
	}
}

// TestMCPScanStringSizeCap exercises the 32 MiB content cap in
// handleScanString. Sends a tools/call with content just over the
// limit and asserts a -32602 (invalid params) error with a clear
// message naming the cap.
func TestMCPScanStringSizeCap(t *testing.T) {
	// 33 MiB content + JSON envelope stays well under the 64 MiB global
	// request cap, so this exercises the per-tool check rather than the
	// global one.
	content := strings.Repeat("A", 33<<20)
	payload := fmt.Sprintf(
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"scan_string","arguments":{"content":%q,"source":"x"}}}`,
		content)

	out := &strings.Builder{}
	code := run(context.Background(), strings.NewReader(payload), out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "exceeds 33554432 byte limit") {
		t.Errorf("expected scan_string content-size error, got: %s",
			out.String()[:min(500, out.Len())])
	}
}

// TestMCPScanEnvCategoriesCap exercises the 100-entry cap on the
// scan_env categories argument. A malicious caller could otherwise
// pin the server by passing a million-category slice that triggers
// a slow string-comparison sweep in SetActiveCategories.
func TestMCPScanEnvCategoriesCap(t *testing.T) {
	cats := make([]string, 101)
	for i := range cats {
		cats[i] = fmt.Sprintf("cat%d", i)
	}
	catsJSON, err := json.Marshal(cats)
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"scan_env","arguments":{"categories":` +
		string(catsJSON) + `}}}`

	out := &strings.Builder{}
	code := run(context.Background(), strings.NewReader(payload), out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "categories exceeds 100 entry limit") {
		t.Errorf("expected categories-cap error, got: %s",
			out.String()[:min(500, out.Len())])
	}
}

// TestMCPInvalidJSONRPCVersion verifies that requests with a jsonrpc
// field other than "2.0" are rejected with -32600 (Invalid Request).
// Pre-PR-#97 they fell through to the unknown-method branch and
// returned -32601, which is misleading — the version mismatch is a
// protocol-level error, not a "method not found".
func TestMCPInvalidJSONRPCVersion(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"1.0","id":1,"method":"initialize"}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), `"code":-32600`) {
		t.Errorf("expected -32600 invalid request error, got: %s", out.String())
	}
}

// TestMCPNotificationNoResponse verifies that a request with no ID
// (JSON-RPC notification) emits no response. Pre-PR-#97 the loop
// happily wrote a response with a null ID, confusing well-behaved
// clients into treating the reply as a response to a later request.
func TestMCPNotificationNoResponse(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","method":"some/notification"}` + "\n")
	out := &strings.Builder{}

	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if out.Len() != 0 {
		t.Errorf("expected no response for notification, got: %s", out.String())
	}
}

// TestMCPPerRequestTimeout verifies that a cancelled parent ctx
// doesn't wedge run(). The per-request child ctx is derived with
// a 30s deadline, so once the parent is cancelled the child is too —
// handlers that respect ctx return promptly and run() exits cleanly.
// We can't easily force a handler to hang, so the assertion is "no
// hang for 2 seconds with a pre-cancelled ctx".
func TestMCPPerRequestTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	out := &strings.Builder{}

	done := make(chan int, 1)
	go func() { done <- run(ctx, in, out) }()

	select {
	case code := <-done:
		// Exit code is implementation-defined (0 if initialize returned
		// before checking ctx, 1 if the encode branch saw the cancellation).
		// The property under test is "doesn't hang", not the exit code.
		t.Logf("run() returned %d with pre-cancelled parent ctx", code)
	case <-time.After(2 * time.Second):
		t.Fatal("run() did not return within 2s with cancelled parent ctx")
	}
}

// min is a tiny helper to keep the failure-message slice operations
// safe in the tests above without importing the built-in (Go 1.21
// stdlib has min/max but using them directly keeps the import list
// in these helpers minimal).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
