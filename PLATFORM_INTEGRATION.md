# Platform Integration: Atheon-Enhanced

**Component:** Atheon-Enhanced pattern scanner  
**Template version:** 1.0  
**Created:** 2026-08-30  
**Status:** Experimental security/quality scanning component; promotion requires changed-line evidence and runtime qualification

## Role

Atheon-Enhanced provides deterministic pattern scanning for secrets, sensitive data, AI-generated code indicators, and quality-policy findings. It is the platform's Aegis-compatible scanning surface for early feedback and policy enforcement, and exposes a local MCP server for agent/tool integration.

## Ownership boundary

Atheon owns pattern definitions, bundle loading, scan execution, finding classification, CLI/MCP protocol behavior, and release artifacts. Aegis and policy decide how findings are interpreted and whether a gate blocks a change; GitForge owns pipeline execution and artifact retention; Control Center owns operator presentation. Atheon must not receive or persist provider credentials.

## Canonical repository path

```text
/nas/Temp/repos/Atheon-Enhanced
```

## Startup and health

Atheon is primarily a CLI and stdio MCP server, not a long-running HTTP service.

```bash
cd /nas/Temp/repos/Atheon-Enhanced
go build -o bin/atheon ./cmd/atheon
go build -o bin/atheon-mcp ./cmd/mcp
bin/atheon --version
```

For MCP readiness, use the stdio protocol rather than an HTTP health URL:

```bash
printf '%s\\n' '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | bin/atheon-mcp
```

The MCP process must keep protocol responses on stdout and diagnostics on stderr.

## API surface

### Inbound interfaces

| Interface | Transport | Purpose |
| --- | --- | --- |
| `atheon <path>` | CLI | Scan files/directories and emit human, JSON, or SARIF findings |
| `atheon list/enable/disable` | CLI | Inspect and manage pattern state |
| `atheon-mcp` | stdio JSON-RPC | Expose scanner tools to MCP clients |
| `core.ScanString`, `core.ScanFile`, `core.ScanDir`, `core.ScanEnv` | Go library | Programmatic scanning with context cancellation |

### Outbound dependencies

| Component | Boundary | Purpose |
| --- | --- | --- |
| GitForge | CI job/artifact boundary | Run bounded scans and retain machine-readable findings |
| Aegis/policy | Scanner result contract | Classify findings and enforce fail-closed gates |
| Control Center | Service/receipt contract | Present scan status and provenance after validation |

## Depends on

- Go 1.21 or newer, according to `go.mod`
- Versioned pattern bundle and configuration profiles
- Bounded filesystem access to the scan target
- Optional MCP-compatible host for stdio operation

## Used by

- GitForge CI quality/security jobs
- Aegis pattern and policy validation
- MCP-capable coding agents and local development hooks

## Required environment variables

No provider credentials are required for deterministic local scanning. Scan paths, configuration, and output format should be passed explicitly by the caller. Never place secrets in pattern fixtures or diagnostic output.

## Test and quality commands

Run through GitForge or another bounded remote executor for heavy suites:

```bash
cd /nas/Temp/repos/Atheon-Enhanced
go test ./... -p 1 -timeout 15m
go vet ./...
test -z "$(gofmt -l .)"
go build -o /nas/Temp/work/atheon-build/atheon ./cmd/atheon
go build -o /nas/Temp/work/atheon-build/atheon-mcp ./cmd/mcp
```

The repository Makefile provides `make test`, `make lint`, `make build`, and `make ci`; do not run watch-mode or unbounded concurrent tests on shared infrastructure.

## Current gaps

- [ ] Establish current clean-branch test, vet, formatting, and coverage evidence; existing claims are not a release gate until reproduced.
- [ ] Add Atheon/Aegis changed-line scan receipts to the platform verification manifest.
- [ ] Verify MCP `initialize` and tool execution end to end, not only process startup or `tools/list`.
- [ ] Reconcile the experimental-fork boundary with upstream Atheon and document which patterns are production-approved.
- [ ] Ensure release artifacts include both `atheon` and `atheon-mcp` with checksums and provenance.
- [ ] Resolve the existing dirty working tree before implementation changes; this contract branch is documentation-only.

## Validation boundary

This contract does not claim Atheon-Enhanced is production-ready or that advertised coverage/performance figures are current. Findings must be preserved with scanner version, pattern bundle hash, target commit, configuration, and explicit `SAFE`/`FINDINGS`/`BLOCKED` classification.
