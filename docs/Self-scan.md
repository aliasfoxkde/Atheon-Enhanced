# Atheon Self-Scan — Meta Integration

Atheon-Enhanced scans itself. Every push and pull request runs the Atheon pattern
engine against its own source code, enforcing the same quality and security bar that
users would apply to their own projects.

This serves two purposes: it keeps the codebase clean, and it acts as a living
integration test — real patterns against real code, not synthetic fixtures.

## How It Works

### CI Workflow

The `Security` workflow (`.github/workflows/security.yml`) runs two self-scans
on every push to `main` and every PR:

| Scan | Categories | Behaviour |
|------|-----------|-----------|
| **Secrets & PII** | `secrets`, `pii` | Blocking — fails the PR if production source has any finding |
| **Code Quality** | `code-quality`, `devops`, `ai-detection` | Informational — reports findings, does not block |

Both scans filter out:
- `*_test.go` — test files carry intentional fake credentials
- `testdata/` — synthetic test inputs
- `community/` — YAML pattern definitions contain example regex targets by design
- `.github/wiki/` — documentation source

This means the gate runs only against production source (`cmd/`, `core/`, `bundler/`).

### Local Script

Run the same scan locally before pushing:

```bash
# Full scan (secrets blocking, code-quality informational)
.github/scripts/self-scan.sh

# Secrets only
.github/scripts/self-scan.sh --secrets

# Code quality only
.github/scripts/self-scan.sh --quality
```

The script builds the binary automatically if `./atheon` doesn't exist.

### Pre-commit Hook

The pre-commit hook in the README scans only staged files, not the whole tree — fast
enough to run on every commit without friction:

```bash
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/sh
set -e
STAGED=$(git diff --cached --name-only --diff-filter=d)
if [ -z "$STAGED" ]; then exit 0; fi
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
for f in $STAGED; do
  mkdir -p "$TMP/$(dirname "$f")"
  git show ":$f" > "$TMP/$f"
done
atheon "$TMP"
EOF
chmod +x .git/hooks/pre-commit
```

## Pattern Count Validation

Every scan step also validates that the pattern bundle loaded correctly:

```bash
COUNT=$(./atheon list | tail -1 | grep -oE '[0-9]+' | head -1)
# Fails CI if COUNT < 250
```

This catches bundle corruption or a broken build before the scan runs.

## Feedback Loop

Running Atheon against real Go source — its own — surfaces edge cases that test
fixtures miss:

- **False positives** in idioms common to Go CLIs (e.g., `fmt.Println` in a CLI
  binary flagged by `fmt-println-prod`) are caught and inform whether patterns need
  confidence adjustments or scope exclusions.
- **False negatives** — real patterns that should fire but don't — are caught when a
  contributor accidentally commits a credential and CI doesn't catch it.
- **Pattern regression** — the count gate (≥250) detects if a build change
  accidentally drops patterns from the bundle.

Each finding in CI is a data point for improving pattern quality, scope, and
confidence levels. The more the engine is used on real-world code, the more refined
the patterns become.

## Extending to Your Project

The same approach works for any repository. Add the pre-commit hook and a CI step:

```yaml
- name: Atheon scan
  run: |
    atheon --json --categories=secrets,pii . \
      | jq '[.[] | select(.file | test("_test\\.go$|/testdata/") | not)]' \
      > findings.json
    COUNT=$(jq 'length' findings.json)
    [ "$COUNT" -eq 0 ] || (jq -r '.[] | "[\(.pattern)] \(.file):\(.line)"' findings.json && exit 1)
```

Replace the jq filter with exclusions appropriate to your repo layout.
