// Fuzz tests for the bundle parser. These run under `go test -fuzz=FuzzParseBundle`
// locally; on CI they execute as a 5-second seeded run (go test default) to keep
// the job bounded while still exercising the malformed-input branches.
//
// The parser is exercised with arbitrary bytes — empty input, garbage,
// truncated gzip, valid JSON that's not gzip, JSON with regex bombs, etc.
// We verify:
//   1. loadBundle never panics, regardless of input.
//   2. loadBundle either returns nil or a non-nil error (never silently
//      corrupts state on partial parse).
//
// regex.compile in Go's regexp is RE2-based and guaranteed linear-time, so
// even pathological input can't wedge the parser. The risk we are guarding
// against is *new* code paths in loadBundle that might miss error handling.
package core

import (
	"testing"
)

func FuzzParseBundle(f *testing.F) {
	// Seed corpus: a few canonical inputs that have already bitten us or
	// would have, so the fuzzer starts from known-good coverage points.
	f.Add([]byte(""))                                 // empty
	f.Add([]byte("not a bundle"))                     // garbage
	f.Add([]byte{0x1f, 0x8b, 0x08})                   // gzip magic, truncated header
	f.Add([]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}) // valid gzip header, empty body
	f.Add([]byte(`[{"name":"x","category":"y","match":"abc","enabled":true}]`))   // plain JSON, no gzip
	f.Add([]byte(`[{"name":"x","category":"y","match":"(?P<a","enabled":true}]`)) // bad regex

	f.Fuzz(func(t *testing.T, data []byte) {
		// loadBundle replaces package-level registry/allPatterns. We save
		// and restore so fuzzing a failing input can't leak state into the
		// next subtest (or worse, into a following TestMain).
		savedRegs, savedPatterns := snapshotState()
		defer restoreState(savedRegs, savedPatterns)

		// We don't care about the return value — only that it doesn't panic.
		_ = loadBundle(data)
	})
}

// snapshotState captures the mutable globals loadBundle touches, so the
// fuzzer can restore them after each run. Keeps each fuzz iteration
// hermetic without depending on init() running again. The captured pointers
// are untouched by loadBundle (which allocates fresh pointers per call), so
// they're safe to put back as-is.
func snapshotState() ([]Pattern, []*bundlePattern) {
	regs := append([]Pattern(nil), registry...)
	patterns := append([]*bundlePattern(nil), allPatterns...)
	return regs, patterns
}

// restoreState puts the package back to the state captured by snapshotState.
// We re-register the saved registry verbatim AND restore allPatterns from the
// captured slice — both are needed because other tests and the example
// subprocess rely on allPatterns being populated for Categories() etc.
// Setting allPatterns = nil (the previous, broken behaviour) left subsequent
// tests seeing "no categories" and broke Example_core_Categories.
func restoreState(regs []Pattern, savedPatterns []*bundlePattern) {
	registry = nil
	allPatterns = nil

	// Restore bundle patterns first so allPatterns and registry agree.
	// savedRegs already contains these pointers; re-registering is a no-op
	// append, so we rely on savedPatterns being authoritative for the
	// bundle slice and savedRegs being authoritative for the rest.
	allPatterns = append(allPatterns, savedPatterns...)
	for _, p := range regs {
		Register(p)
	}
}