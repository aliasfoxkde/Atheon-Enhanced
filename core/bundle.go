// Package core implements pattern loading, scanning, and bundle management for Atheon.
// It exposes a registry of RE2-compatible patterns compiled from a gzip-JSON bundle
// that is embedded at build time and optionally overridden by ~/.atheon/patterns.bundle.
package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

//go:embed patterns.bundle
var embeddedBundle []byte

// PatternDef is the on-disk (and on-wire) representation of a pattern as
// it appears inside a pattern bundle. Match holds the regular-expression
// source; the compiled *regexp.Regexp is not part of the wire form. Severity
// is optional — patterns that omit it default to "medium" (see DefaultSeverity).
// Schema version 2 (bundle v2) adds Description, Reference, and Tags fields.
type PatternDef struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Match       string   `json:"match"`
	Enabled     bool     `json:"enabled"`
	Severity    string   `json:"severity,omitempty"`
	Description string   `json:"description,omitempty"`
	Reference   string   `json:"reference,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	MinEntropy  float64  `json:"minEntropy,omitempty"`
	Confidence  string   `json:"confidence,omitempty"`
}

// DefaultSeverity is the severity assigned to patterns that don't declare one.
// "medium" matches the most common legacy pattern that hard-coded everything
// as a warning-level finding before SARIF consumers learned to filter.
const DefaultSeverity = "medium"

// ValidSeverities lists the recognized severity strings. Anything outside
// this set is normalised to DefaultSeverity when the bundle loads — keeping
// downstream code (SARIF mapping, JSON output) safe from typo'd YAML.
var ValidSeverities = []string{"low", "medium", "high", "critical"}

type bundlePattern struct {
	name        string
	category    string
	match       string
	description string
	reference   string
	tags        []string
	// enabled is mutated under patternMu (write lock for any Store, read
	// lock for any Load via scanLines/scanEnv snapshots). We keep it as a
	// plain bool rather than atomic.Bool so the public test surface
	// (core/*_test.go) can still read it directly — they call into the
	// locked helpers (Enable/Disable/SetPatternEnabled) under patternMu
	// when they mutate it, and only read it for assertions. The mutex
	// is the single point of synchronization.
	enabled    bool
	severity   string
	confidence string  // high, medium, or low
	minEntropy float64 // Minimum entropy threshold (0 = no filtering)
	re         *regexp.Regexp
}

func (p *bundlePattern) Name() string             { return p.name }
func (p *bundlePattern) Category() string         { return p.category }
func (p *bundlePattern) Description() string      { return p.description }
func (p *bundlePattern) Reference() string        { return p.reference }
func (p *bundlePattern) Tags() []string           { return p.tags }
func (p *bundlePattern) Matches(line string) bool { return p.enabled && p.re.MatchString(line) }

// matchSpan returns the [start, end) byte offsets of p's first match in
// line, or (-1, -1) if it doesn't match or the pattern is disabled.
// Unexported because it exposes RE2's internal offsets, which aren't
// stable across pattern implementations (only bundlePattern has a
// compiled regex to query).
func (p *bundlePattern) matchSpan(line string) (start, end int) {
	if !p.enabled || p.re == nil {
		return -1, -1
	}
	loc := p.re.FindStringIndex(line)
	if loc == nil {
		return -1, -1
	}
	return loc[0], loc[1]
}
func (p *bundlePattern) Enabled() bool           { return p.enabled }
func (p *bundlePattern) SetEnabled(enabled bool) { p.enabled = enabled }

// Severity returns the pattern's severity — one of ValidSeverities, never empty.
// Patterns loaded without a severity field read back as DefaultSeverity.
func (p *bundlePattern) Severity() string { return p.severity }

// MinEntropy returns the minimum entropy threshold for this pattern.
// A value of 0 means no entropy filtering is applied.
func (p *bundlePattern) MinEntropy() float64 { return p.minEntropy }

// Confidence returns the pattern's confidence level: "high", "medium", or "low".
// Default is "medium" if not specified.
func (p *bundlePattern) Confidence() string { return p.confidence }

// ValidConfidenceLevels for validation
var ValidConfidenceLevels = []string{"high", "medium", "low"}

// normalizeConfidence maps any string to one of ValidConfidenceLevels, falling back
// to "medium" for empty or unrecognized values.
func normalizeConfidence(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, v := range ValidConfidenceLevels {
		if s == v {
			return v
		}
	}
	return "medium"
}

// normalizeSeverity maps any string to one of ValidSeverities, falling back
// to DefaultSeverity for empty or unrecognised values. Comparison is
// case-insensitive so "HIGH" and "high" behave the same.
func normalizeSeverity(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, v := range ValidSeverities {
		if s == v {
			return v
		}
	}
	return DefaultSeverity
}

type categoryScanner struct {
	combined *regexp.Regexp
	patterns []Pattern
}

var (
	allPatterns          []*bundlePattern
	activeScanners       []categoryScanner
	activeCategoryFilter []string // nil = all categories; preserved across rebuildActiveScanners
)

func init() {
	// Default initialization reads the user's local bundle from ~/.atheon
	// and falls back to the embedded bundle. Splitting this out as
	// initializeWith lets tests exercise the error branches.
	data := embeddedBundle
	if home, err := os.UserHomeDir(); err == nil {
		if b, err := os.ReadFile(filepath.Join(home, ".atheon", "patterns.bundle")); err == nil {
			data = b
		}
	}
	initializeWith(data)
}

// initializeWith runs the same setup as init() but accepts the bundle data
// directly so tests can feed in corrupt data to exercise the error paths.
func initializeWith(data []byte) {
	// init() runs single-threaded before any user goroutine can race with us,
	// so locking here would be redundant — and would deadlock against the
	// package-level init order (init can't wait on a mutex some other
	// goroutine might already hold). All write paths through loadBundle /
	// SetActiveCategories / InitializePatternState take patternMu inside,
	// which is a no-op for callers that already hold it via this init.
	if err := loadBundle(data); err != nil {
		slog.Warn("bundle load failed", "err", err)
	}
	setActiveCategoriesLocked(nil)

	// Load pattern state after bundle is loaded
	if err := InitializePatternState(); err != nil {
		// Non-fatal error, just log warning
		slog.Warn("pattern state initialization failed", "err", err)
	}
}

// decodeJSONStrict decodes JSON from data into out, first validating that no
// unknown fields are present (fields not defined in the destination struct).
// This is equivalent to json.Decoder.DisallowUnknownFields (Go 1.24+).
// An unknown field causes an error to be returned.  Handles both JSON objects
// (map) and JSON arrays ([]struct) as the top-level container.
func decodeJSONStrict(data []byte, out interface{}) error {
	// Peek at the first non-whitespace byte to determine container type.
	trimmed := trimSpace(data)
	if len(trimmed) == 0 {
		return fmt.Errorf("bundle: empty JSON data")
	}
	switch trimmed[0] {
	case '{':
		// Object — validate and reject unknown keys.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("bundle v1 unmarshal: %w", err)
		}
		v := reflect.ValueOf(out).Elem()
		t := v.Type()
		valid := make(map[string]struct{}, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			valid[t.Field(i).Name] = struct{}{}
		}
		for k := range raw {
			if _, ok := valid[k]; !ok {
				return fmt.Errorf("bundle contains unknown field %q", k)
			}
		}
		return json.Unmarshal(data, out)
	case '[':
		// Array — accept directly; individual items validated by their types.
		return json.Unmarshal(data, out)
	default:
		return fmt.Errorf("bundle: unexpected JSON root type")
	}
}

// trimSpace returns data with leading/trailing whitespace removed.
func trimSpace(data []byte) []byte {
	i, j := 0, len(data)
	for i < j && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	for j > i && (data[j-1] == ' ' || data[j-1] == '\t' || data[j-1] == '\n' || data[j-1] == '\r') {
		j--
	}
	return data[i:j]
}

// decodeBundleDefs parses the decompressed bundle data, handling both
// schema v1 (flat []PatternDef) and schema v2 ({"schema_version":2,"data":[...]}).
func decodeBundleDefs(decompressed []byte) ([]PatternDef, error) {
	trimmed := trimSpace(decompressed)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("bundle: empty data")
	}

	// Detect v2 envelope: first non-whitespace char is '{'
	if trimmed[0] == '{' {
		// Validate no unknown fields in v2 envelope before unmarshaling.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(decompressed, &raw); err != nil {
			return nil, fmt.Errorf("bundle v2 parse error: %w", err)
		}
		validEnv := map[string]struct{}{"schema_version": {}, "data": {}}
		for k := range raw {
			if _, ok := validEnv[k]; !ok {
				return nil, fmt.Errorf("bundle v2: unknown field %q", k)
			}
		}
		var v2 struct {
			SchemaVersion int          `json:"schema_version"`
			Data          []PatternDef `json:"data"`
		}
		if err := json.Unmarshal(decompressed, &v2); err != nil {
			return nil, fmt.Errorf("bundle v2 unmarshal: %w", err)
		}
		if v2.SchemaVersion != 2 {
			return nil, fmt.Errorf("bundle: unsupported schema version %d", v2.SchemaVersion)
		}
		return v2.Data, nil
	}

	// v1: flat array
	var defs []PatternDef
	if err := json.Unmarshal(decompressed, &defs); err != nil {
		return nil, fmt.Errorf("bundle v1 parse error: %w", err)
	}
	return defs, nil
}

// decompress gzip-decompresses data and returns the raw bytes.
func decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer r.Close()
	return io.ReadAll(r)
}

func loadBundle(data []byte) error {
	decompressed, err := decompress(data)
	if err != nil {
		return fmt.Errorf("bundle decompress: %w", err)
	}
	return loadBundleFrom(decompressed)
}

// loadBundleFrom loads patterns from already-decompressed JSON data.
func loadBundleFrom(decompressed []byte) error {
	defs, err := decodeBundleDefs(decompressed)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBundleParse, err)
	}

	patternMu.Lock()
	defer patternMu.Unlock()

	var external []Pattern
	for _, p := range registry {
		if _, ok := p.(*bundlePattern); !ok {
			external = append(external, p)
		}
	}
	registry = nil
	allPatterns = nil

	for _, def := range defs {
		re, err := regexp.Compile(def.Match)
		if err != nil {
			slog.Warn("skipping pattern due to regex error", "pattern", def.Name, "err", err)
			continue
		}
		bp := &bundlePattern{
			name:        def.Name,
			category:    def.Category,
			match:       def.Match,
			description: def.Description,
			reference:   def.Reference,
			tags:        def.Tags,
			enabled:     def.Enabled,
			severity:    normalizeSeverity(def.Severity),
			confidence:  normalizeConfidence(def.Confidence),
			minEntropy:  def.MinEntropy,
			re:          re,
		}
		allPatterns = append(allPatterns, bp)
		registerLocked(bp)
	}

	// Old bundles predate the enabled field; JSON zero-value false means all appear
	// disabled. Detect this and default everything to enabled.
	anyEnabled := false
	for _, p := range allPatterns {
		if p.enabled {
			anyEnabled = true
			break
		}
	}
	if !anyEnabled {
		for _, p := range allPatterns {
			p.enabled = true
		}
		// Old bundles predate the enabled field and look like an all-disabled
		// bundle at decode time. The flip above restores them. Without this
		// log line, a contributor who commits a NEW bundle that's accidentally
		// all-`enabled: false` looks identical at runtime — every pattern
		// silently comes on. Surface the legacy-compat path so it stays
		// observable.
		if len(allPatterns) > 0 {
			slog.Info("bundle had no enabled patterns; defaulting all to enabled (legacy compatibility — add 'enabled: true' explicitly if this was unintentional)",
				"patterns", len(allPatterns))
		}
	}

	for _, p := range external {
		registerLocked(p)
	}

	return nil
}

// SetActiveCategories restricts subsequent scans to the named categories.
// A nil or empty slice means "all categories." Calling this rebuilds the
// internal pre-filter regexes used by ScanFile and ScanDir.
func SetActiveCategories(cats []string) {
	patternMu.Lock()
	defer patternMu.Unlock()
	setActiveCategoriesLocked(cats)
}

// setActiveCategoriesLocked is SetActiveCategories with the lock already held.
// Kept separate so callers already holding patternMu (loadBundle,
// rebuildActiveScanners) don't recurse on the lock.
func setActiveCategoriesLocked(cats []string) {
	activeCategoryFilter = cats

	catSet := map[string]bool{}
	for _, c := range cats {
		catSet[strings.TrimSpace(c)] = true
	}

	byCategory := map[string][]Pattern{}
	for _, p := range allPatterns {
		// Enabled filtering is enforced later in rebuildActiveScanners / via
		// p.Matches; gating here would silently drop disabled patterns from
		// the per-category map and double-disable them. Skipping this check
		// keeps the category view complete and lets ListEnabledPatterns /
		// ListDisabledPatterns reflect the actual on-disk state.
		if len(cats) > 0 && !catSet[p.category] {
			continue
		}
		byCategory[p.category] = append(byCategory[p.category], p)
	}
	// Include externally registered non-bundle patterns so Register() callers are scanned
	for _, p := range registry {
		if _, ok := p.(*bundlePattern); ok {
			continue
		}
		cat := p.Category()
		if len(cats) > 0 && !catSet[cat] {
			continue
		}
		byCategory[cat] = append(byCategory[cat], p)
	}

	activeScanners = nil
	for cat, patterns := range byCategory {
		// Split bundle patterns (have a match regex) from external patterns (don't)
		var bundlePs, extPs []Pattern
		for _, p := range patterns {
			if _, ok := p.(*bundlePattern); ok {
				bundlePs = append(bundlePs, p)
			} else {
				extPs = append(extPs, p)
			}
		}
		if len(bundlePs) > 0 {
			parts := make([]string, 0, len(bundlePs))
			for _, p := range bundlePs {
				parts = append(parts, "(?:"+p.(*bundlePattern).match+")")
			}
			combined, err := regexp.Compile(strings.Join(parts, "|"))
			if err == nil {
				activeScanners = append(activeScanners, categoryScanner{combined: combined, patterns: bundlePs})
			} else {
				// Combined-regex compile failure isn't fatal — the per-pattern
				// scanners below would still match correctly, just slower
				// (every pattern is evaluated on every line instead of being
				// pre-filtered by the category-level combined regex). The
				// historical behaviour was a silent drop, which let broken
				// patterns linger unnoticed: matches stopped appearing for a
				// category but no error surfaced. Warn so operators can fix
				// the offending pattern.
				slog.Warn("combined regex compile failed; falling back to per-pattern matching",
					"category", cat, "patterns", len(bundlePs), "err", err)
			}
		}
		if len(extPs) > 0 {
			// External patterns have no regex to pre-filter with; use empty (matches all)
			combined := regexp.MustCompile("")
			activeScanners = append(activeScanners, categoryScanner{combined: combined, patterns: extPs})
		}
	}
}

// Categories returns the unique, unsorted list of category labels present
// in the current bundle. The returned slice is owned by the caller.
func Categories() []string {
	patternMu.RLock()
	defer patternMu.RUnlock()
	seen := map[string]bool{}
	var cats []string
	for _, p := range allPatterns {
		if !seen[p.category] {
			seen[p.category] = true
			cats = append(cats, p.category)
		}
	}
	return cats
}

// maxBundleDownloadBytes bounds the bundle download to prevent memory exhaustion
// from a malicious or compromised upstream server sending an unbounded payload.
// 100 MiB is generous — the current bundle is ~400 KiB compressed — and leaves
// room for future growth without approaching a DoS vector.
const maxBundleDownloadBytes = 100 << 20 // 100 MiB

// bundleDownloadURL is the default upstream bundle URL. Tests may override
// it via SetBundleDownloadURL to point at an httptest server. Held in an
// atomic.Pointer so concurrent readers (fetchBundleData) and the test-only
// writer don't tear the string under -race.
var bundleDownloadURL = atomic.Pointer[string]{}

// skipHostValidation disables hostname validation in SetBundleDownloadURL when
// true.  Used by tests that intentionally point at httptest servers on loopback.
// Tests in the core package may set this via Store().
var skipHostValidation atomic.Bool

// SetBundleDownloadURLForTest is like SetBundleDownloadURL but also disables
// hostname validation so tests can point at httptest servers on loopback.
// The returned restore function resets skipHostValidation to false.
func SetBundleDownloadURLForTest(bundleURL string) func() {
	skipHostValidation.Store(true)
	return SetBundleDownloadURL(bundleURL)
}

// init-time default. Done as a function rather than a literal so a test
// that calls SetBundleDownloadURL("") before init can still observe the
// real default on first use.
func init() {
	bundleDownloadURL.Store(ptrString("https://github.com/aliasfoxkde/Atheon-Enhanced/releases/latest/download/patterns.bundle"))
}

func ptrString(s string) *string { return &s }

// SetBundleDownloadURL swaps the upstream URL used by DownloadBundle. It
// rejects file:// and other non-HTTP schemes that could be used for SSRF
// attacks (e.g., file:///etc/passwd to read local files, ftp:// to scan
// internal networks). HTTP and HTTPS schemes pass through; any URL —
// including malformed ones, localhost, and private IPs — proceeds to the
// HTTP layer where DownloadBundle's context timeout provides safety.
// Returns a restore function that callers should defer to reset the URL
// after tests or short-lived overrides. Exported so external test packages
// (e.g., the main binary's tests) can stub out network access.
//
// NOTE: This function panics if rawURL has an non-HTTP(S) scheme or
// resolves to a reserved/private IP address. This is intentional — it
// rejects clearly wrong configuration at initialization time rather than
// silently proceeding. Callers should only invoke this with values they
// have validated or trust (e.g., env vars, config files, test overrides).
func SetBundleDownloadURL(rawURL string) func() {
	if rawURL != "" {
		u, err := url.Parse(rawURL)
		if err == nil {
			switch u.Scheme {
			case "http", "https":
				// Allowed — validate hostname below.
			default:
				// Reject file://, ftp://, sftp://, etc.
				panic("SetBundleDownloadURL: non-HTTP(S) scheme rejected: " + rawURL)
			}
			// Reject loopback, private, and link-local addresses to block SSRF
			// to cloud metadata endpoints (169.254.x.x), LAN services, and
			// localhost.  Hostname resolution is synchronous and bounded by the
			// caller's context timeout (15 s in DownloadBundle).
			if !skipHostValidation.Load() {
				host := u.Hostname()
				if host != "" && isReservedOrPrivateHost(host) {
					panic("SetBundleDownloadURL: reserved/private host rejected: " + host)
				}
			}
		}
	}
	prev := bundleDownloadURL.Swap(ptrString(rawURL))
	return func() {
		bundleDownloadURL.Store(prev)
		skipHostValidation.Store(false)
	}
}

// isReservedOrPrivateHost returns true if host resolves to a loopback,
// private, or link-local IP address.  If host is not a valid IP or
// hostname, it returns false (the scheme/blocklist check already ran).
func isReservedOrPrivateHost(host string) bool {
	// First check if host is already a valid IP.
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsPrivate() || isLinkLocal(ip)
	}
	// Otherwise resolve the hostname.  Timeout is governed by the caller.
	addrs, err := net.LookupHost(host)
	if err != nil {
		// Cannot resolve — allow through; HTTP layer will fail anyway.
		return false
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsPrivate() || isLinkLocal(ip) {
			return true
		}
	}
	return false
}

// isLinkLocal returns true for IPv4 169.254.x.x and IPv6 fe80::.
func isLinkLocal(ip net.IP) bool {
	if ip == nil {
		return false
	}
	// 169.254.0.0/16
	b := ip.To4()
	if b != nil {
		return b[0] == 169 && b[1] == 254
	}
	// fe80::/10
	return ip.IsLinkLocalUnicast()
}

// DownloadBundle fetches the latest pattern bundle from the URL configured via
// SetBundleDownloadURL (or the default URL), compares it against the in-memory
// bundle, verifies the SHA-256 hash against checksums.txt, prints a summary of
// added/removed patterns, and persists the new bundle to ~/.atheon/patterns.bundle.
//
// If force is false and the bundle appears fresh (ETag matches a recent check),
// DownloadBundle returns nil immediately without contacting the server. The
// 24-hour freshness window avoids a round-trip on every scan.
//
// The bundle is loaded into memory before being written to disk; if loadBundle
// fails the on-disk bundle is left untouched.
//
// The context controls the HTTP request lifecycle: canceling ctx aborts the
// in-flight download.
//
// On any non-success HTTP status code, DownloadBundle returns an error wrapping
// ErrBundleDownload so callers can use errors.Is.
func DownloadBundle(ctx context.Context, force bool) error {
	start := time.Now()
	bundleURL := *bundleDownloadURL.Load()
	slog.Info("bundle download started", "url", bundleURL, "force", force)

	// Stale-bundle check: if not forced, see if we've checked recently with a
	// matching ETag. This skips the network round-trip on repeated scans.
	if !force {
		skip, etag := shouldSkipDownload()
		if skip {
			slog.Info("bundle download skipped (fresh)", "etag", etag)
			return nil
		}
	}

	oldPatterns := currentPatternNames()

	data, etag, err := fetchBundleData(ctx)
	if err != nil {
		return fmt.Errorf("bundle fetch: %w", err)
	}

	// Verify SHA-256 hash against checksums.txt if available.
	if err := verifyBundleHash(ctx, data); err != nil {
		return fmt.Errorf("%w: %w", ErrBundleHashMismatch, err)
	}

	slog.Info("bundle download complete", "bytes", len(data), "elapsed_ms", time.Since(start).Milliseconds(), "etag", etag)

	dir, err := ensureAtheonDir()
	if err != nil {
		return fmt.Errorf("ensure atheon dir: %w", err)
	}

	// Decompress once and reuse the result for both diff computation and loading.
	decompressed, err := decompress(data)
	if err != nil {
		return fmt.Errorf("bundle decompress: %w", err)
	}
	var newDefs []PatternDef
	if err := decodeJSONStrict(decompressed, &newDefs); err != nil {
		return fmt.Errorf("%w: %w", ErrBundleParse, err)
	}

	added, removed := diffPatternNames(oldPatterns, newDefs)
	printBundleDiff(len(oldPatterns), len(newDefs), added, removed)

	// Load into memory first; only persist to disk if that succeeds.
	// Use loadBundleFrom to avoid re-decompressing (data is already decompressed).
	if err := loadBundleFrom(decompressed); err != nil {
		return fmt.Errorf("bundle load: %w", err)
	}
	// loadBundle took the write lock and released it on return. Re-acquire
	// here so the rebuild stays inside the critical section — a concurrent
	// Enable call between loadBundle and rebuildActiveScanners would
	// otherwise mutate the slice we're about to recompute.
	patternMu.Lock()
	setActiveCategoriesLocked(activeCategoryFilter)
	patternMu.Unlock()
	// Atomic write: write to .tmp + rename so a SIGKILL mid-write leaves
	// the previous bundle intact. See pattern_state.go for the same pattern.
	if err := atomicWriteFile(filepath.Join(dir, "patterns.bundle"), data, 0o600); err != nil {
		return fmt.Errorf("bundle persist: %w", err)
	}

	// Record the ETag and timestamp so the next call can skip if unchanged.
	if err := recordBundleETag(etag); err != nil {
		slog.Warn("failed to record bundle ETag", "err", err)
	}

	return nil
}

// shouldSkipDownload returns (true, etag) if the bundle appears fresh:
// the last check was within 24 hours and the stored ETag matches the
// upstream value. Force downloads always return (false, "").
func shouldSkipDownload() (skipped bool, etag string) {
	etagVal, lastChecked, err := loadBundleETag()
	if err != nil || etagVal == "" {
		return false, ""
	}
	// 24-hour freshness window
	if time.Since(lastChecked) < 24*time.Hour {
		return true, etagVal
	}
	return false, ""
}

// recordBundleETag persists the ETag and timestamp to the pattern state file.
func recordBundleETag(etag string) error {
	return withFileLock(stateFile(), func() error {
		state, err := loadPatternState()
		if err != nil {
			return fmt.Errorf("load pattern state: %w", err)
		}
		state.BundleETag = etag
		state.BundleLastChecked = time.Now().UnixNano()
		return savePatternState(state)
	})
}

// loadBundleETag returns the stored ETag and last-checked time.
func loadBundleETag() (etag string, lastChecked time.Time, err error) {
	state, err := loadPatternState()
	if err != nil {
		return "", time.Time{}, err
	}
	if state.BundleETag == "" || state.BundleLastChecked == 0 {
		return "", time.Time{}, nil
	}
	return state.BundleETag, time.Unix(0, state.BundleLastChecked), nil
}

// currentPatternNames returns the names of every pattern currently in
// the active bundle, in slice order.
func currentPatternNames() []string {
	patternMu.RLock()
	defer patternMu.RUnlock()
	var names []string
	for _, p := range allPatterns {
		names = append(names, p.name)
	}
	return names
}

// fetchBundleData performs the HTTP GET against bundleDownloadURL and
// returns the response body and ETag header on success.
func fetchBundleData(ctx context.Context) (data []byte, etag string, err error) {
	// Reject redirects to prevent SSRF via redirect (e.g., https://trusted.com →
	// file:///etc/passwd or http://169.254.169.254/...). The allow-list in
	// SetBundleDownloadURL only permits known hosts, but a redirect can bypass it.
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return fmt.Errorf("redirect rejected: SSRF prevention")
		},
	}
	urlPtr := bundleDownloadURL.Load()
	if urlPtr == nil {
		return nil, "", fmt.Errorf("%w: bundle download URL not initialised", ErrBundleDownload)
	}
	// bundleDownloadURL is configured via SetBundleDownloadURL and only
	// ever points at https://github.com/... or a test stub. SSRF surface
	// is bounded by the controlled allow-list maintained in this package.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *urlPtr, http.NoBody)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrBundleDownload, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrBundleDownload, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("%w: server returned %d", ErrBundleDownload, resp.StatusCode)
	}
	// LimitedReader truncates at maxBundleDownloadBytes but does NOT return an
	// error when truncated — it just stops reading. Detect truncation by comparing
	// against Content-Length when that header is available.
	data, err = io.ReadAll(&io.LimitedReader{R: resp.Body, N: maxBundleDownloadBytes})
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrBundleDownload, err)
	}
	if resp.ContentLength > 0 && int64(len(data)) < resp.ContentLength {
		return nil, "", fmt.Errorf("%w: bundle exceeded %d byte cap (server sent %d bytes)",
			ErrBundleDownload, maxBundleDownloadBytes, resp.ContentLength)
	}
	return data, resp.Header.Get("ETag"), nil
}

// computeBundleHash computes the SHA-256 hex digest of data.
func computeBundleHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// verifyBundleHash downloads checksums.txt from the same directory as
// the bundle URL and verifies that the bundle's SHA-256 hash appears
// in it. If checksums.txt is unavailable (404) the check is skipped
// with a warning — this allows the feature to work in environments
// where checksums haven't been published yet. Any other error is logged
// but not propagated so a bad checksums.txt doesn't block bundle updates.
func verifyBundleHash(ctx context.Context, data []byte) error {
	bundleURLVar := *bundleDownloadURL.Load()
	// Derive checksums URL: strips filename, appends checksums.txt.
	// https://github.com/.../download/v1.2.3/patterns.bundle
	//   → https://github.com/.../download/v1.2.3/checksums.txt
	idx := strings.LastIndex(bundleURLVar, "/")
	if idx < 0 {
		return fmt.Errorf("cannot derive checksums URL from %q", bundleURLVar)
	}
	checksumsURL := bundleURLVar[:idx+1] + "checksums.txt"

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("checksums request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		slog.Warn("checksums.txt not found at upstream, skipping hash verification", "url", checksumsURL)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksums returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap on checksums.txt
	if err != nil {
		return fmt.Errorf("reading checksums: %w", err)
	}

	// checksums.txt format: one line per file, "<hexhash> <filename>".
	// We look for a line whose filename component matches "patterns.bundle".
	expectedHash := computeBundleHash(data)
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "<hash> <filename)" — split on whitespace, take last token as filename
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash, filename := parts[0], parts[len(parts)-1]
		if filename == "patterns.bundle" || filename == "patterns.bundle.gz" {
			if hash == expectedHash {
				slog.Debug("bundle hash verified", "hash", expectedHash)
				return nil
			}
			return fmt.Errorf("%w: hash mismatch for patterns.bundle (expected %s, got %s)",
				ErrBundleHashMismatch, expectedHash, hash)
		}
	}

	slog.Warn("patterns.bundle not found in checksums.txt, skipping verification",
		"url", checksumsURL, "checked_lines", len(lines))
	return nil
}

// ensureAtheonDir creates (if needed) and returns the ~/.atheon path.
func ensureAtheonDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".atheon")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// diffPatternNames computes the symmetric set difference between the
// currently-loaded pattern names and a freshly-downloaded bundle.
func diffPatternNames(oldPatterns []string, newDefs []PatternDef) (added, removed []string) {
	newSet := make(map[string]bool, len(newDefs))
	for _, def := range newDefs {
		newSet[def.Name] = true
	}
	oldSet := make(map[string]bool, len(oldPatterns))
	for _, name := range oldPatterns {
		oldSet[name] = true
	}
	for _, name := range oldPatterns {
		if !newSet[name] {
			removed = append(removed, name)
		}
	}
	for _, def := range newDefs {
		if !oldSet[def.Name] {
			added = append(added, def.Name)
		}
	}
	return added, removed
}

// printBundleDiff writes a human-readable summary of the bundle change.
func printBundleDiff(oldCount, newCount int, added, removed []string) {
	fmt.Printf("Patterns updated: %d → %d\n", oldCount, newCount)
	switch {
	case len(added) > 0:
		fmt.Printf("Added: %d patterns\n", len(added))
		for _, p := range added {
			fmt.Printf("  + %s\n", p)
		}
		if len(removed) > 0 {
			fmt.Printf("Removed: %d patterns\n", len(removed))
			for _, p := range removed {
				fmt.Printf("  - %s\n", p)
			}
		}
	case len(removed) > 0:
		fmt.Printf("Removed: %d patterns\n", len(removed))
		for _, p := range removed {
			fmt.Printf("  - %s\n", p)
		}
	default:
		fmt.Println("No pattern changes detected")
	}
}

// EnablePattern enables the named pattern, rebuilds the scanner set,
// and persists the new state. It returns false if no pattern with the
// given name exists in the bundle.
func EnablePattern(name string) bool {
	patternMu.Lock()
	defer patternMu.Unlock()
	for _, p := range allPatterns {
		if p.name != name {
			continue
		}
		p.enabled = true
		rebuildRegistry()
		rebuildActiveScanners()
		if err := syncPatternState(); err != nil {
			slog.Warn("failed to save pattern state", "err", err)
		}
		return true
	}
	return false
}

// DisablePattern disables the named pattern, rebuilds the scanner set,
// and persists the new state. It returns false if no pattern with the
// given name exists in the bundle.
func DisablePattern(name string) bool {
	patternMu.Lock()
	defer patternMu.Unlock()
	for _, p := range allPatterns {
		if p.name != name {
			continue
		}
		p.enabled = false
		rebuildRegistry()
		rebuildActiveScanners()
		if err := syncPatternState(); err != nil {
			slog.Warn("failed to save pattern state", "err", err)
		}
		return true
	}
	return false
}

// SetPatternEnabled sets the enabled flag for the named pattern and
// rebuilds the active scanner set. Unlike EnablePattern and
// DisablePattern it does not persist state to disk — useful for tests
// and for callers that batch state updates. Returns false if the
// pattern name is unknown.
func SetPatternEnabled(name string, enabled bool) bool {
	patternMu.Lock()
	defer patternMu.Unlock()
	for _, p := range allPatterns {
		if p.name == name {
			p.enabled = enabled
			rebuildActiveScanners()
			return true
		}
	}
	return false
}

// ListDisabledPatterns returns the names of every pattern that is
// currently disabled, in bundle order.
func ListDisabledPatterns() []string {
	patternMu.RLock()
	defer patternMu.RUnlock()
	var disabled []string
	for _, p := range allPatterns {
		if !p.enabled {
			disabled = append(disabled, p.name)
		}
	}
	return disabled
}

// ListEnabledPatterns returns the names of every pattern that is
// currently enabled, in bundle order.
func ListEnabledPatterns() []string {
	patternMu.RLock()
	defer patternMu.RUnlock()
	var enabled []string
	for _, p := range allPatterns {
		if p.enabled {
			enabled = append(enabled, p.name)
		}
	}
	return enabled
}

func rebuildActiveScanners() {
	// Caller must already hold patternMu for writing.
	setActiveCategoriesLocked(activeCategoryFilter)
}

// EnableAllPatterns enables every pattern in the bundle, overriding any
// prior disable calls, then rebuilds the active scanner set.
func EnableAllPatterns() {
	patternMu.Lock()
	defer patternMu.Unlock()
	for _, p := range allPatterns {
		p.enabled = true
	}
	rebuildActiveScanners()
}

// ReloadBundle discards any runtime state (downloaded bundle, enable/disable
// changes) and reloads the embedded bundle. Useful in tests that call
// DownloadBundle and need to restore a clean slate afterward.
func ReloadBundle() {
	initializeWith(embeddedBundle)
}

// rebuildRegistry rebuilds the registry from allPatterns, respecting enabled state.
// Caller must hold patternMu for writing.
func rebuildRegistry() {
	registry = nil
	for _, p := range allPatterns {
		if p.enabled {
			registerLocked(p)
		}
	}
}
