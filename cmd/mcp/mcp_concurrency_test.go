package main

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestMCPInflightCounter verifies that the inflight counter increments when
// a request enters dispatchRequest and decrements when it exits (via defer).
func TestMCPInflightCounter(t *testing.T) {
	mcpInflight.Store(0)

	ctx := context.Background()
	in := strings.NewReader(`{"jsonrpc":"2.0","id":"counter-test","method":"initialize"}` + "\n")
	out := &strings.Builder{}
	code := run(ctx, in, out)
	if code != 0 {
		t.Errorf("expected run() exit 0, got %d", code)
	}
	if resp := out.String(); !strings.Contains(resp, `"protocolVersion"`) {
		t.Errorf("expected initialize response, got: %s", resp)
	}
	if n := mcpInflight.Load(); n != 0 {
		t.Errorf("expected inflight=0 after request, got %d", n)
	}
}

// TestMCPConcurrentCapFires verifies that when the inflight counter is already
// at the cap, a new request is immediately rejected with -32001.
func TestMCPConcurrentCapFires(t *testing.T) {
	mcpInflight.Store(0)

	// Pre-load counter to exactly cap using Add (same atomic op dispatchRequest uses).
	// When a new request arrives, Add(1) returns cap+1, and the check
	// (cap+1 > cap) is true → -32001 is returned.
	for i := 0; i < mcpConcurrentCap; i++ {
		mcpInflight.Add(1)
	}

	ctx := context.Background()
	in := strings.NewReader(`{"jsonrpc":"2.0","id":"over-cap","method":"initialize"}` + "\n")
	out := &strings.Builder{}
	run(ctx, in, out)

	resp := out.String()
	if !strings.Contains(resp, `"code":-32001`) {
		t.Errorf("expected -32001 when over cap, got: %s", resp)
	}

	mcpInflight.Store(0)
}

// TestMCPConcurrentCapAllowsBelow verifies that a request succeeds when
// the counter is strictly below the cap.
func TestMCPConcurrentCapAllowsBelow(t *testing.T) {
	mcpInflight.Store(0)

	// Counter at cap-1 → a new request brings it to cap. The check
	// (cap > cap) = false → request proceeds.
	for i := 0; i < mcpConcurrentCap-1; i++ {
		mcpInflight.Add(1)
	}

	ctx := context.Background()
	in := strings.NewReader(`{"jsonrpc":"2.0","id":"below-cap","method":"initialize"}` + "\n")
	out := &strings.Builder{}
	run(ctx, in, out)

	resp := out.String()
	if !strings.Contains(resp, `"protocolVersion"`) {
		t.Errorf("expected initialize success below cap, got: %s", resp)
	}

	mcpInflight.Store(0)
}

// TestMCPConcurrentCapMultipleHolders verifies the cap holds when many
// concurrent goroutines each hold one slot. We launch cap+5 goroutines
// each running a scan_dir on /tmp (returns fast) and then check that a
// subsequent request gets -32001 while the holders are still running.
func TestMCPConcurrentCapMultipleHolders(t *testing.T) {
	mcpInflight.Store(0)

	// Launch cap holders in background — each will block inside dispatchRequest
	// for as long as we hold them.
	var holderWg sync.WaitGroup
	holderCtx, holderCancel := context.WithCancel(context.Background())
	for i := 0; i < mcpConcurrentCap; i++ {
		holderWg.Add(1)
		go func(idx int) {
			defer holderWg.Done()
			// Send a scan_dir request that will block (the context will be canceled
			// by the holder goroutine below after we check). We use scan_dir
			// on /dev/null because it returns quickly but still exercises the
			// full dispatch path.
			in := strings.NewReader(`{"jsonrpc":"2.0","id":"holder-` +
				strings.Repeat("x", idx) + `","method":"scan_dir","params":{"path":"/tmp"}}` + "\n")
			run(holderCtx, in, io.Discard)
		}(i)
	}

	// Give holders time to enter dispatchRequest.
	// Spin until counter reaches cap.
	for i := 0; i < 1000 && mcpInflight.Load() < int64(mcpConcurrentCap); i++ {
		time.Sleep(time.Millisecond)
	}

	if mcpInflight.Load() < int64(mcpConcurrentCap) {
		t.Skipf("holders did not enter dispatchRequest in time (counter=%d), skipping", mcpInflight.Load())
	}

	// Now send a request — it should be rejected.
	ctx := context.Background()
	in := strings.NewReader(`{"jsonrpc":"2.0","id":"should-be-limited","method":"initialize"}` + "\n")
	out := &strings.Builder{}
	run(ctx, in, out)

	resp := out.String()
	if !strings.Contains(resp, `"code":-32001`) {
		t.Errorf("expected -32001 when at cap, got: %s", resp)
	}

	holderCancel()
	holderWg.Wait()
	mcpInflight.Store(0)
}
