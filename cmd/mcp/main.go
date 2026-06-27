package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/aliasfoxkde/Atheon/core"
	"github.com/aliasfoxkde/Atheon/internal/errors"
)

// version is the server version, set at build time via:
//
//	-ldflags "-X main.version=1.2.3"
//
// Defaults to "dev" so `go run ./cmd/mcp` is usable without a build script.
var version = "dev"

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// Data is an optional structured payload per JSON-RPC 2.0 spec.
	// Used to convey category tags for MCP clients that need to route
	// errors programmatically (e.g., "rate_limit", "concurrent_limit").
	Data any `json:"data,omitempty"`
}

// rateLimitCode is the JSON-RPC error code returned when a request is
// denied by the rate limiter. JSON-RPC reserves -32000..-32099 for
// implementation-defined server errors; -32600 is "Invalid Request",
// which is the wrong code for a throttling response.
const rateLimitCode = -32000

// rateLimiter implements a simple token bucket rate limiter.
// Uses stdlib only to avoid external dependencies.
type rateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTime time.Time
}

// newRateLimiter creates a rate limiter allowing maxTokens per second, up to burst.
func newRateLimiter(tokensPerSecond, burst float64) *rateLimiter {
	return &rateLimiter{
		tokens:   burst,
		max:      burst,
		rate:     tokensPerSecond,
		lastTime: time.Now(),
	}
}

// Allow checks if a request is permitted under the rate limit.
// Returns true if allowed, false if rate limited.
func (rl *rateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.lastTime = now

	// Add tokens based on elapsed time
	rl.tokens += elapsed * rl.rate
	if rl.tokens > rl.max {
		rl.tokens = rl.max
	}

	if rl.tokens < 1 {
		return false
	}
	rl.tokens--
	return true
}

// activeRequests tracks in-flight request IDs and their cancel functions.
// Used to implement $/cancelRequest: when a cancel notification arrives,
// the cancel function is called, aborting the in-flight handler.
var activeRequests sync.Map

// cancelRequestCode is the JSON-RPC error code returned when a request
// was successfully canceled per the MCP spec.
const cancelRequestCode = -32802

// normalizeID converts a JSON-RPC request ID to a string for use as a
// sync.Map key. JSON-RPC allows IDs to be strings, numbers, or null.
// sync.Map requires interface{} keys, so we stringify for consistency
// and to avoid json.Number vs int mismatches (e.g. "1" != 1).
func normalizeID(id any) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", id)
	}
}

// envInt parses an int env var with a default.
func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// envBytes parses a size in bytes from an env var. Supports bare integers
// only (K/M/G suffix parsing can be added if needed).
func envBytes(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// envDuration parses a duration string (e.g., "30s") from an env var.
func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

// mcpRateLimiter is the global rate limiter for MCP requests.
// Configurable via ATHEON_MCP_RATE_LIMIT and ATHEON_MCP_RATE_BURST.
var mcpRateLimiter *rateLimiter

// mcpConcurrentCap is the maximum number of concurrent request handlers.
// Configurable via ATHEON_MCP_CONCURRENT_CAP.
var mcpConcurrentCap int

// mcpInflight tracks the number of active request handlers using an
// atomic Int so dispatchRequest can check and increment/decrement
// without holding a mutex.
var mcpInflight atomic.Int64

// mcpMaxRequestBytes caps the size of a single JSON-RPC request line.
// Configurable via ATHEON_MCP_MAX_REQUEST_BYTES (bytes, default 64MiB).
var mcpMaxRequestBytes int

// mcpRequestTimeout bounds the time a single request handler can run.
// Configurable via ATHEON_MCP_REQUEST_TIMEOUT (e.g., "30s").
var mcpRequestTimeout time.Duration

// mcpScanStringMaxBytes caps the scan_string tool's content argument.
// Configurable via ATHEON_MCP_SCAN_STRING_MAX_BYTES (bytes, default 32MiB).
var mcpScanStringMaxBytes int

func init() {
	rateLimit := envInt("ATHEON_MCP_RATE_LIMIT", 10)
	rateBurst := envInt("ATHEON_MCP_RATE_BURST", 20)
	mcpRateLimiter = newRateLimiter(float64(rateLimit), float64(rateBurst))

	mcpConcurrentCap = envInt("ATHEON_MCP_CONCURRENT_CAP", 50)
	mcpMaxRequestBytes = envBytes("ATHEON_MCP_MAX_REQUEST_BYTES", 64<<20)
	mcpRequestTimeout = envDuration("ATHEON_MCP_REQUEST_TIMEOUT", 30*time.Second)
	mcpScanStringMaxBytes = envBytes("ATHEON_MCP_SCAN_STRING_MAX_BYTES", 32<<20)
}

// JSON-RPC method names handled by the MCP server. Extracted as
// constants so goconst can verify they're not duplicated and so
// readers can see the protocol surface in one place.
const (
	methodInitialize = "initialize"
	methodToolsList  = "tools/list"
	methodToolsCall  = "tools/call"
)

func main() {
	configureLogging()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Stdin, os.Stdout)
	cancel() // explicit: os.Exit skips deferred cancel
	os.Exit(code)
}

// configureLogging mirrors cmd/atheon's setup so MCP server logs are
// configurable via the same env vars (ATHEON_LOG_FORMAT, ATHEON_LOG_LEVEL).
// Without this, slog's default text handler is used and downstream
// aggregators have to parse key=value pairs from a non-deterministic
// format.
func configureLogging() {
	var level slog.Level
	switch strings.ToLower(os.Getenv("ATHEON_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.EqualFold(os.Getenv("ATHEON_LOG_FORMAT"), "json") {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}

// run executes the JSON-RPC loop reading from r and writing to w, returning
// the exit code. Separated from main() so tests can call it without os.Exit
// terminating the test process.
//
// The context is forwarded into the core scan helpers so a SIGTERM
// received mid-scan aborts cleanly.
func run(ctx context.Context, r io.Reader, w io.Writer) int {
	// bufio.Scanner with a 64 MiB max replaces the prior 1 MiB
	// Scanner cap. Scanner reads one line at a time and returns
	// bufio.ErrTooLong if a single line exceeds the configured max,
	// giving us a hard memory ceiling on a single malformed/oversized
	// request without per-line allocations.
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), mcpMaxRequestBytes)
	enc := json.NewEncoder(w)

	for sc.Scan() {
		var req request
		if err := json.Unmarshal(sc.Bytes(), &req); err != nil {
			// Malformed JSON has no ID we can echo back. Log and
			// keep reading; a flood of bad requests is rate-limited
			// below anyway.
			slog.Warn("malformed JSON-RPC request; skipping", "err", err)
			continue
		}

		// Rate limit at the top of the loop — BEFORE any per-method
		// dispatch — so initialize / tools/list floods count against
		// the same token bucket as tools/call. Pre-PR-#97 only
		// tools/call was throttled, so an attacker could pin the
		// server by spamming `initialize`.
		if !mcpRateLimiter.Allow() {
			slog.Warn("rate limit exceeded for MCP request", "method", req.Method)
			// Only send a response when the client asked for one
			// (notifications have nil ID and expect no reply).
			if req.ID != nil {
				_ = enc.Encode(response{
					JSONRPC: "2.0",
					ID:      req.ID,
					Error:   &rpcError{Code: rateLimitCode, Message: "rate limit exceeded", Data: "rate_limit"},
				})
			}
			continue
		}

		// Per-request structured log so MCP traffic is observable in ELK /
		// Loki / Datadog. Gated at Debug so the default Info level stays
		// quiet for the common case. Use fmt.Sprintf for ID since the
		// field is `any` and JSON-encoding a nil ID emits "null" which is
		// technically correct but harder to grep than "<notif>".
		idStr := "<notif>"
		if req.ID != nil {
			idStr = fmt.Sprintf("%v", req.ID)
		}
		slog.Debug("mcp request", "method", req.Method, "id", idStr)
		if req.Method == "initialized" {
			continue
		}

		// Per-request timeout: derive a child ctx with a 30s deadline
		// so a stuck handler can't wedge the server. cancel() runs
		// via defer to release the timer promptly.
		reqCtx, cancel := context.WithTimeout(ctx, mcpRequestTimeout)
		result, rerr := dispatchRequest(reqCtx, &req)
		cancel()

		// Notifications (JSON-RPC requests with no ID) expect no
		// reply — emitting one would confuse well-behaved clients
		// into treating the reply as a response to a later request.
		if req.ID == nil {
			continue
		}

		resp := response{JSONRPC: "2.0", ID: req.ID}
		if rerr != nil {
			resp.Error = rerr
		} else {
			resp.Result = result
		}
		if err := enc.Encode(resp); err != nil {
			// Encode failure usually means the client disconnected
			// mid-write. Log and exit cleanly so we don't loop
			// forever trying to reply to a vanished peer.
			slog.Error("mcp encode error", "err", err)
			return 1
		}
	}
	return 0
}

// dispatchRequest runs a single JSON-RPC request and returns its
// result or error. Extracted from run() so a defer can wrap it in
// recover() — that way a panic in any tool handler is converted to
// an -32603 response instead of killing the entire MCP server (which
// would terminate every active agent session on the host).
//
// Pre-PR-#97 the switch lived inline in run(); a panic in handleScanDir
// (e.g. a future nil-deref in a pattern that exposes a config knob)
// would tear down the whole process and force every connected client
// to reconnect. With recover() in place the bad request gets an error
// response and the server keeps serving.
func dispatchRequest(ctx context.Context, req *request) (result any, rerr *rpcError) {
	// Track whether we incremented the inflight counter so the defer
	// knows whether to decrement. This avoids a double-decrement when
	// the cap-check early-return path also runs the defer.
	didIncrement := false
	//nolint:revive // intentional MCP server resilience: one tool panic must not kill the daemon
	defer func() {
		if r := recover(); r != nil {
			slog.Error("mcp handler panic", "method", req.Method, "panic", fmt.Sprintf("%v", r))
			rerr = &rpcError{Code: -32603, Message: "internal error"}
			result = nil
		}
		if didIncrement {
			mcpInflight.Add(-1)
		}
	}()

	// Concurrent request cap: check before doing any work. This prevents
	// a burst of scan_dir requests (each of which can run for seconds)
	// from creating unbounded goroutines. We increment first so the
	// counter is pessimistic — a handler that returns early still counts
	// against the cap for the duration of its work.
	if mcpInflight.Add(1) > int64(mcpConcurrentCap) {
		mcpInflight.Add(-1) // undo the increment; didIncrement stays false
		return nil, &rpcError{Code: -32001, Message: fmt.Sprintf("concurrent request limit reached (%d)", mcpConcurrentCap), Data: "concurrent_limit"}
	}
	didIncrement = true

	// JSON-RPC 2.0 requires the jsonrpc field. Anything else is a
	// protocol-level malformed request — return -32600 (Invalid
	// Request) so a misbehaving client sees a clear error rather
	// than the method-not-found fallback.
	if req.JSONRPC != "2.0" {
		return nil, &rpcError{Code: -32600, Message: "jsonrpc field must be \"2.0\""}
	}

	switch req.Method {
	case methodInitialize:
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "atheon", "version": version},
		}, nil
	case methodToolsList:
		return map[string]any{"tools": toolList()}, nil
	case methodToolsCall:
		return handleCall(ctx, req.ID, req.Params)
	case "$/cancelRequest":
		// JSON-RPC cancel notification: mark the request ID as canceled.
		// handleCall checks activeRequests before invoking the tool;
		// if the ID is present it returns -32802 immediately.
		var cancel struct {
			ID any `json:"id"`
		}
		if err := json.Unmarshal(req.Params, &cancel); err != nil {
			return nil, nil // notifications return no response
		}
		if key := normalizeID(cancel.ID); key != "" {
			activeRequests.Store(key, struct{}{})
		}
		return nil, nil
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

// toolList returns the MCP tool registry. The schema helper wraps a Go
// property bag into the JSON Schema shape MCP expects.
func toolList() []map[string]any {
	schema := func(required []string, props map[string]any) map[string]any {
		return map[string]any{"type": "object", "properties": props, "required": required}
	}
	str := map[string]any{"type": "string"}
	cats := map[string]any{"type": "array", "items": str, "description": "categories to scan (omit for all)"}

	return []map[string]any{
		{
			"name":        "scan_string",
			"description": "Scan a string for pattern matches",
			"inputSchema": schema([]string{"content"}, map[string]any{
				"content":    map[string]any{"type": "string"},
				"source":     str,
				"categories": cats,
			}),
		},
		{
			"name":        "scan_file",
			"description": "Scan a file for pattern matches",
			"inputSchema": schema([]string{"path"}, map[string]any{
				"path":       map[string]any{"type": "string"},
				"categories": cats,
			}),
		},
		{
			"name":        "scan_dir",
			"description": "Scan a directory for pattern matches",
			"inputSchema": schema([]string{"path"}, map[string]any{
				"path":       map[string]any{"type": "string"},
				"categories": cats,
			}),
		},
		{
			"name":        "scan_env",
			"description": "Scan process environment variables for pattern matches",
			"inputSchema": schema([]string{}, map[string]any{
				"categories": cats,
			}),
		},
		{
			"name":        "list_patterns",
			"description": "List all loaded patterns (name, category, enabled)",
			"inputSchema": schema([]string{}, map[string]any{
				"category": map[string]any{
					"type":        "string",
					"description": "filter to a single category (omit for all)",
				},
			}),
		},
		{
			"name":        "list_categories",
			"description": "List all pattern categories available in the bundle",
			"inputSchema": schema([]string{}, map[string]any{}),
		},
		{
			"name":        "update_bundle",
			"description": "Download the latest pattern bundle from the configured URL",
			"inputSchema": schema([]string{}, map[string]any{
				"force": map[string]any{
					"type":        "boolean",
					"description": "bypass the 24-hour freshness cache and force a re-download",
				},
			}),
		},
	}
}

// toolHandler is the signature every per-tool dispatcher implements.
// Extracted so handleCall stays under the lint funlen limit and each
// tool's parse-and-execute logic is independently testable.
type toolHandler func(ctx context.Context, args json.RawMessage) (any, *rpcError)

// toolHandlers maps each tool name to its handler. Lookups via map keep
// handleCall flat instead of an ever-growing switch.
var toolHandlers = map[string]toolHandler{
	"scan_string":     handleScanString,
	"scan_file":       handleScanFile,
	"scan_dir":        handleScanDir,
	"scan_env":        handleScanEnv,
	"list_patterns":   handleListPatterns,
	"list_categories": handleListCategories,
	"update_bundle":   handleUpdateBundle,
}

// handleCall parses the JSON-RPC params envelope, looks up the tool
// handler, and dispatches. Rate-limit is applied at the top of run()
// so initialize/tools/list floods count against the same token bucket
// as tools/call (PR #97).
//
// Cancellation: if the request ID appears in the activeRequests map
// (set by a prior $/cancelRequest notification), handleCall returns
// -32802 immediately without invoking the tool. This prevents a
// canceled request from wasting CPU on a long-running scan.
func handleCall(ctx context.Context, id any, params json.RawMessage) (any, *rpcError) {
	// Check cancel map before doing any work.
	if key := normalizeID(id); key != "" {
		if _, canceled := activeRequests.LoadAndDelete(key); canceled {
			return nil, &rpcError{Code: cancelRequestCode, Message: "request was canceled"}
		}
	}

	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "invalid params"}
	}
	h, ok := toolHandlers[p.Name]
	if !ok {
		return nil, &rpcError{Code: -32601, Message: "unknown tool: " + p.Name}
	}
	return h(ctx, p.Arguments)
}

// invalidParams is a small helper so every tool handler returns the
// same JSON-RPC error shape on argument parse failure.
func invalidParams(err error) *rpcError {
	return &rpcError{Code: -32602, Message: "invalid params: " + err.Error(), Data: "invalid_params"}
}

// safeError maps a Go error to a user-facing JSON-RPC error message that
// contains no raw filesystem paths, syscall strings, or internal details.
// This prevents the MCP server from leaking host structure to AI agents
// that parse error messages (a HIGH-severity finding from the Wave 9 audit).
//
// Mapping rules:
// Uses the shared internal/errors.SafeError. The per-comment above describes
// the mapping: os.IsNotExist → "file not found", etc.

// sandboxPath evaluates any symlinks in path and verifies the result
// stays within the process's current working directory. This prevents
// an MCP client from passing "../../../etc/passwd" or a relative symlink
// that resolves outside the cwd (e.g. cwd/subdir -> /etc).
//
// Absolute paths (e.g. /tmp/file.txt) are passed through directly —
// a user explicitly requesting /tmp is intentional filesystem access.
// Relative paths must resolve under the cwd after symlink evaluation.
func sandboxPath(path string) (string, error) {
	// Absolute paths are allowed — the user explicitly named them.
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Relative path: detect traversal before calling EvalSymlinks (which
	// fails on non-existent paths). filepath.Clean collapses ".." without
	// requiring filesystem access.
	clean := filepath.Clean(path)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", os.ErrPermission
	}

	// Resolve symlinks. Catches cases like "cmd/../../etc/passwd".
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		// Broken or non-existent path — reject it rather than passing a
		// raw path that may have traversal components still in it.
		return "", os.ErrNotExist
	}
	cwd, err := os.Getwd()
	if err != nil {
		return realPath, os.ErrPermission
	}
	cwdReal, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return realPath, os.ErrPermission
	}
	// Block traversal: e.g. cwd/subdir -> /etc via symlink.
	if !strings.HasPrefix(realPath, cwdReal) {
		return "", os.ErrPermission
	}
	return realPath, nil
}

func handleScanString(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var args struct {
		Content    string   `json:"content"`
		Source     string   `json:"source"`
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, invalidParams(err)
	}
	// PR #97: cap the content size before decoding/holding the buffer.
	// Without this, an agent could pass an entire 1 GiB repo dump and
	// the per-pattern scanLines buffer would balloon proportionally.
	// Reject at the tool layer so the caller sees a clear error
	// rather than a silently truncated scan.
	if len(args.Content) > mcpScanStringMaxBytes {
		return nil, &rpcError{
			Code:    -32602,
			Message: fmt.Sprintf("content exceeds %d byte limit (got %d)", mcpScanStringMaxBytes, len(args.Content)),
		}
	}
	if args.Source == "" {
		args.Source = "stdin"
	}
	core.SetActiveCategories(args.Categories)
	return textResult(core.ScanString(ctx, args.Content, args.Source)), nil
}

func handleScanFile(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var args struct {
		Path       string   `json:"path"`
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, invalidParams(err)
	}
	// CRITICAL: canonicalize and sandbox the path before passing to ScanFile.
	// This prevents ../../etc/passwd and symlink-escape attacks.
	if _, err := sandboxPath(args.Path); err != nil {
		return nil, &rpcError{Code: -32603, Message: errors.SafeError(err)}
	}
	core.SetActiveCategories(args.Categories)
	findings, _, err := core.ScanFile(ctx, args.Path)
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: errors.SafeError(err)}
	}
	return textResult(findings), nil
}

func handleScanDir(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var args struct {
		Path       string   `json:"path"`
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, invalidParams(err)
	}
	// CRITICAL: canonicalize and sandbox the path before passing to ScanDir.
	// This prevents ../../secrets and symlink-escape attacks even when
	// NoFollowSymlinks is true (WalkDir doesn't follow, but the path
	// itself could escape if we pass an absolute path outside cwd).
	if _, err := sandboxPath(args.Path); err != nil {
		return nil, &rpcError{Code: -32603, Message: errors.SafeError(err)}
	}
	core.SetActiveCategories(args.Categories)
	// MCP defaults to the safe symlink policy. Agents invoking scan_dir
	// are typically operating on untrusted trees (third-party repos,
	// generated code, scratch dirs), and a symlink escape would let a
	// crafted repo leak /etc/passwd or ~/.aws/credentials into the
	// findings without the operator ever noticing. The CLI keeps the
	// historical follow-symlinks behaviour behind an opt-in flag.
	findings, _, err := core.ScanDir(ctx, args.Path, core.ScanOpts{NoFollowSymlinks: true})
	if err != nil {
		return nil, &rpcError{Code: -32603, Message: errors.SafeError(err)}
	}
	return textResult(findings), nil
}

func handleScanEnv(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var args struct {
		Categories []string `json:"categories"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, invalidParams(err)
	}
	// PR #97: cap the categories slice. SetActiveCategories rebuilds
	// the per-category active filters; with no cap, a 1 MB slice of
	// garbage category names would trigger a slow string-comparison
	// sweep through every bundle pattern. 100 is generous — the
	// bundle currently has 19 categories.
	const maxCategories = 100
	if len(args.Categories) > maxCategories {
		return nil, &rpcError{
			Code:    -32602,
			Message: fmt.Sprintf("categories exceeds %d entry limit (got %d)", maxCategories, len(args.Categories)),
		}
	}
	core.SetActiveCategories(args.Categories)
	return textResult(core.ScanEnv(ctx)), nil
}

func handleListPatterns(_ context.Context, raw json.RawMessage) (any, *rpcError) {
	var args struct {
		Category string `json:"category"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, invalidParams(err)
	}
	return patternsResult(core.All(), args.Category), nil
}

func handleListCategories(_ context.Context, _ json.RawMessage) (any, *rpcError) {
	return categoriesResult(core.Categories()), nil
}

func handleUpdateBundle(ctx context.Context, raw json.RawMessage) (any, *rpcError) {
	var args struct {
		Force bool `json:"force"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, invalidParams(err)
	}
	if err := core.DownloadBundle(ctx, args.Force); err != nil {
		return nil, &rpcError{Code: -32603, Message: errors.SafeError(err)}
	}
	return map[string]any{
		"content": []map[string]any{{
			"type": "text",
			"text": "bundle updated successfully",
		}},
	}, nil
}

func textResult(findings []core.Finding) map[string]any {
	var sb strings.Builder
	if len(findings) == 0 {
		sb.WriteString("no findings")
	} else {
		for _, f := range findings {
			fmt.Fprintf(&sb, "%s  %s:%d\n", f.Pattern, f.File, f.Line)
		}
		fmt.Fprintf(&sb, "\n%d finding(s)", len(findings))
	}
	// structuredContent carries machine-readable findings for clients that
	// need parsed output rather than plain text.
	structured := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		structured = append(structured, map[string]any{
			"pattern":     f.Pattern,
			"file":        f.File,
			"line":        f.Line,
			"column":      f.Column,
			"content":     f.Content,
			"severity":    f.Severity,
			"category":    f.Category,
			"fingerprint": f.Fingerprint,
		})
	}
	return map[string]any{
		"content": []map[string]any{{
			"type":               "text",
			"text":               sb.String(),
			"structuredContent":  structured,
		}},
	}
}

// patternsResult renders the pattern list as a markdown table inside the
// MCP text-content wrapper. If category is non-empty, only patterns whose
// Category() matches are returned.
func patternsResult(patterns []core.Pattern, category string) map[string]any {
	var sb strings.Builder
	fmt.Fprintf(&sb, "| Name | Category | Enabled |\n")
	fmt.Fprintf(&sb, "|------|----------|---------|\n")
	count := 0
	structured := make([]map[string]any, 0)
	for _, p := range patterns {
		if category != "" && p.Category() != category {
			continue
		}
		enabled := "no"
		if p.Enabled() {
			enabled = "yes"
		}
		fmt.Fprintf(&sb, "| %s | %s | %s |\n", p.Name(), p.Category(), enabled)
		structured = append(structured, map[string]any{
			"name":     p.Name(),
			"category": p.Category(),
			"enabled":  p.Enabled(),
		})
		count++
	}
	if count == 0 {
		if category != "" {
			fmt.Fprintf(&sb, "\n(no patterns in category %q)", category)
		} else {
			sb.WriteString("\n(no patterns loaded)")
		}
	} else {
		fmt.Fprintf(&sb, "\n%d pattern(s)", count)
	}
	return map[string]any{
		"content": []map[string]any{{
			"type":               "text",
			"text":               sb.String(),
			"structuredContent":  structured,
		}},
	}
}

// categoriesResult renders the category list as a simple comma-separated
// text block, since categories are short labels.
func categoriesResult(cats []string) map[string]any {
	var sb strings.Builder
	if len(cats) == 0 {
		sb.WriteString("(no categories loaded)")
	} else {
		fmt.Fprintf(&sb, "%s\n\n%d categories", strings.Join(cats, ", "), len(cats))
	}
	return map[string]any{
		"content": []map[string]any{{
			"type":               "text",
			"text":               sb.String(),
			"structuredContent":  cats,
		}},
	}
}
