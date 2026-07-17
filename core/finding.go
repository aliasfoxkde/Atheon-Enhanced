package core

// Finding represents a single pattern match produced by ScanFile, ScanDir,
// ScanString, or ScanEnv. File is the source path (or "env:KEY" when
// scanning the process environment); Line is the 1-indexed line number
// within the source (0 for env scans); Content is the trimmed matching
// line or, for env scans, the matching value.
//
// Severity is the pattern's declared severity at the time of the match —
// one of "low", "medium", "high", "critical". It's copied off the Pattern
// at match time so toggling severity later doesn't rewrite history.
//
// Column is the 1-indexed byte offset of the match's first byte within
// the line. 0 means "unknown" (the pattern is not a bundlePattern, or
// the line had a stripped trailing newline that shifted byte positions).
// Downstream consumers (SARIF, IDE integrations) translate this into
// end-user coordinates; SARIF wants 1-indexed columns so we store 1-indexed
// values directly to avoid an off-by-one at every conversion site.
type Finding struct {
	Pattern     string
	File        string
	Line        int
	Column      int
	Content     string
	Severity    string
	Category    string   // Pattern category at match time (copied like Severity)
	Description string   // Pattern description at match time
	Reference   string   // Pattern reference URL at match time
	Tags        []string // Pattern tags at match time
	Fingerprint string   // Stable deduplication key: "pattern|file|line|col"
	Confidence  string   // Pattern confidence: "high", "medium", or "low"
}

// Stats summarizes the work performed by ScanFile or ScanDir. Files is the
// number of files whose contents were scanned (binary files and skipped
// directories are excluded); Bytes is the total number of bytes scanned;
// ElapsedMs is the wall-clock duration of the scan in milliseconds.
// Errors collects any per-file read errors encountered during a ScanDir
// walk so the caller can surface them instead of silently dropping them.
type Stats struct {
	Files     int
	Bytes     int64
	ElapsedMs int64
	Errors    []error
}
