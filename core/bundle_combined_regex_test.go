package core

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
)

// atomicBoolTrue returns an atomic.Bool set to true.
func atomicBoolTrue() atomic.Bool {
	var b atomic.Bool
	b.Store(true)
	return b
}

// captureHandler is a slog.Handler that records every emitted record
// into a buffer. We use it instead of swapping slog.Default globally so
// parallel tests aren't affected and the captured output is per-test.
type captureHandler struct {
	mu  *bytes.Buffer
	h   slog.Handler
	out *[]slog.Record
}

func newCaptureHandler(buf *bytes.Buffer, records *[]slog.Record) *captureHandler {
	return &captureHandler{
		mu:  buf,
		h:   slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		out: records,
	}
}

func (c *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (c *captureHandler) Handle(ctx context.Context, r slog.Record) error {
	c.mu.WriteString(r.Message + "\n")
	*c.out = append(*c.out, r)
	return nil
}
func (c *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return c }
func (c *captureHandler) WithGroup(_ string) slog.Handler      { return c }

// TestCombinedRegexCompileErrorLogged asserts that when a category's
// combined regex fails to compile, slog.Warn is emitted (so operators
// notice broken patterns) and the category still loads via per-pattern
// matching (so a single bad regex doesn't take the whole category
// down — matches still appear, just slower).
//
// The historical behaviour (pre-PR-#95) was a silent drop: the
// category's patterns would never match because their combined
// pre-filter never built, but no error surfaced to the user. This
// regression test pins the new behaviour in place.
func TestCombinedRegexCompileErrorLogged(t *testing.T) {
	// Snapshot global state so the test is hermetic. PR #95 changed
	// rebuildActiveScanners' behaviour to fall back to per-pattern
	// matching, which mutates the package-level activeScanners slice;
	// without a snapshot, parallel tests would see partial state.
	restore := snapshotBundleState()
	defer restore()

	// Capture slog output via a one-shot handler so we don't have to
	// mutate slog.Default globally (which races with parallel tests).
	var buf bytes.Buffer
	var records []slog.Record
	prevSlog := slog.Default()
	slog.SetDefault(slog.New(newCaptureHandler(&buf, &records)))

	// Inject a category whose only pattern has a regex that RE2 cannot
	// compile (`[a-` is an unterminated character class). RE2 accepts
	// many things other engines reject (nested quantifiers are legal
	// since it can't backtrack), so we have to reach for a genuinely
	// unparseable pattern. We hold the write lock because both the
	// assignment to allPatterns AND the rebuildActiveScanners call
	// below require it (see rebuildActiveScanners' comment in
	// core/bundle.go).
	patternMu.Lock()
	allPatterns = []*bundlePattern{{
		name:     "wave8-bad-pattern",
		match:    "[a-",
		category: "wave8-test-cat",
		enabled:  atomicBoolTrue(),
		severity: "low",
	}}
	rebuildActiveScanners()
	patternMu.Unlock()

	slog.SetDefault(prevSlog)

	if buf.Len() == 0 {
		t.Fatal("expected slog.Warn output, got none")
	}
	// Look for the specific warn message — loose match on substring
	// so test isn't brittle to minor wording changes.
	if !strings.Contains(buf.String(), "combined regex compile failed") {
		t.Fatalf("expected warn about combined regex compile failure, got: %s", buf.String())
	}

	// Sanity-check: at least one record was Warn-level (the call site
	// uses slog.Warn, not slog.Info). Ensures the log isn't being
	// silently downgraded to debug.
	var sawWarn bool
	for _, r := range records {
		if r.Level >= slog.LevelWarn {
			sawWarn = true
			break
		}
	}
	if !sawWarn {
		t.Fatal("expected a Warn-level record")
	}
}

// snapshotBundleState snapshots the global bundle-pattern state and
// returns a restore function. Tests that mutate the globals for
// fault-injection (this one, and a few others) should always defer
// the restore so subsequent tests don't see polluted state.
//
// The restore re-runs initializeWith against the embedded bundle —
// that's how every other test in this package resets state (see
// init_paths_test.go). Calling rebuildActiveScanners directly would
// leave the bundle's parse-once initialisation inconsistent with the
// restored allPatterns slice.
func snapshotBundleState() func() {
	prevSlog := slog.Default()
	return func() {
		slog.SetDefault(prevSlog)
		initializeWith(embeddedBundle)
	}
}
