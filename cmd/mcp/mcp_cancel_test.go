package main

import (
	"context"
	"strings"
	"testing"
)

// TestCancelRequestHandler verifies that a $/cancelRequest notification
// with ID X causes a subsequent tools/call with the same ID X to return
// -32802 immediately, without invoking the tool handler.
//
// Because the run() loop is sequential, the cancel must be sent BEFORE
// the tools/call for this to work: the cancel stores X in activeRequests,
// then when tools/call arrives handleCall finds X and returns -32802
// without calling the tool.
func TestCancelRequestHandler(t *testing.T) {
	const callID = "cancel-handler-test-001"

	// Send cancel first, then the tools/call with the same ID.
	// The cancel marks callID as canceled; handleCall finds it and returns -32802.
	in := strings.NewReader(
		// Cancel for callID FIRST (so activeRequests has it when tools/call arrives)
		`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":"` + callID + `"}}` + "\n" +
			// tools/call with the same ID — should be rejected immediately
			`{"jsonrpc":"2.0","id":"` + callID + `","method":"tools/call","params":{"name":"scan_string","arguments":{"content":"hello","source":"x"}}}` + "\n")

	out := &strings.Builder{}
	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected run() to exit 0, got %d", code)
	}

	resp := out.String()
	// The tools/call response must contain the cancel error.
	if !strings.Contains(resp, `"code":-32802`) {
		t.Errorf("expected -32802 cancel error in response, got: %s", resp)
	}
}

// TestCancelRequestNoID verifies that a cancel notification with no matching
// pending request is silently ignored (no response emitted).
func TestCancelRequestNoID(t *testing.T) {
	in := strings.NewReader(
		// A regular initialize
		`{"jsonrpc":"2.0","id":2,"method":"initialize"}` + "\n" +
			// A cancel for an ID that was never pending
			`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":999999}}` + "\n")

	out := &strings.Builder{}
	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected run() to exit 0, got %d", code)
	}

	resp := out.String()
	// Should have exactly one response (the initialize reply).
	lines := strings.Split(strings.TrimSpace(resp), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 response line, got %d: %s", len(lines), resp)
	}
	if !strings.Contains(resp, `"protocolVersion"`) {
		t.Errorf("expected initialize response, got: %s", resp)
	}
	// Must not contain -32802 — the canceled ID was never pending.
	if strings.Contains(resp, `"code":-32802`) {
		t.Errorf("unexpected cancel error for never-pending ID, got: %s", resp)
	}
}

// TestCancelRequestUnknownMethod verifies that the cancel handler is
// invisible to the method routing — a $/cancelRequest with a valid ID
// does not produce a "method not found" error.
func TestCancelRequestUnknownMethod(t *testing.T) {
	// Deliberately send cancel as a request (with ID) rather than notification.
	// The server should route it to the cancel handler and return nil (no response).
	// We verify no -32601 appears.
	in := strings.NewReader(
		`{"jsonrpc":"2.0","id":"cancel-unknown-method","method":"$/cancelRequest","params":{"id":"cancel-unknown-method"}}` + "\n")

	out := &strings.Builder{}
	code := run(context.Background(), in, out)
	if code != 0 {
		t.Errorf("expected run() to exit 0, got %d", code)
	}

	resp := out.String()
	if strings.Contains(resp, `"code":-32601`) {
		t.Errorf("cancelRequest should not return method-not-found, got: %s", resp)
	}
}
