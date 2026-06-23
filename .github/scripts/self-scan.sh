#!/usr/bin/env bash
# Atheon self-scan — run this locally before pushing or from CI.
#
# Usage:
#   .github/scripts/self-scan.sh            # full scan
#   .github/scripts/self-scan.sh --secrets  # secrets only
#   .github/scripts/self-scan.sh --quality  # code-quality only
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
cd "${REPO_ROOT}"

# ── helpers ──────────────────────────────────────────────────────────────────

info()    { echo "  [info] $*"; }
success() { echo "  [✓] $*"; }
warning() { echo "  [⚠] $*"; }
fail()    { echo "  [✗] $*" >&2; }

# Filter out test/testdata/community files from JSON findings array.
# Keeps only findings that are in production source.
filter_prod() {
  jq '[.[] | select((
    (.file | test("_test\\.go$"))   or
    (.file | test("/testdata/"))     or
    (.file | test("^community/"))    or
    (.file | test("^\\.github/wiki/"))
  ) | not)]'
}

# ── build ─────────────────────────────────────────────────────────────────────

echo ""
echo "=== Atheon Self-Scan ==="
echo ""

if [ ! -f ./atheon ]; then
  info "Building atheon binary..."
  go build -o atheon ./cmd/atheon
fi

# Validate pattern bundle.
COUNT=$(./atheon list | tail -1 | grep -oE '[0-9]+' | head -1)
info "Pattern bundle: ${COUNT} patterns loaded"
if [ "${COUNT}" -lt 250 ]; then
  fail "Expected ≥250 patterns, got ${COUNT} — bundle may be corrupt."
  exit 1
fi

# ── scans ─────────────────────────────────────────────────────────────────────

RUN_SECRETS=true
RUN_QUALITY=true

for arg in "$@"; do
  case "${arg}" in
    --secrets) RUN_QUALITY=false ;;
    --quality) RUN_SECRETS=false ;;
  esac
done

EXIT_CODE=0

# ── 1. Secrets scan (blocking) ────────────────────────────────────────────────

if [ "${RUN_SECRETS}" = "true" ]; then
  echo ""
  echo "── Secrets & PII scan (production source only) ──"
  ./atheon --json --categories=secrets,pii . | filter_prod > /tmp/atheon-secrets.json
  SEC_COUNT=$(jq 'length' /tmp/atheon-secrets.json)

  if [ "${SEC_COUNT}" -gt 0 ]; then
    fail "Found ${SEC_COUNT} secrets/PII finding(s) in production source:"
    jq -r '.[] | "  [\(.pattern)] \(.file):\(.line)  \(.match)"' /tmp/atheon-secrets.json
    EXIT_CODE=1
  else
    success "No secrets or PII in production source."
  fi
fi

# ── 2. Code-quality scan (informational) ──────────────────────────────────────

if [ "${RUN_QUALITY}" = "true" ]; then
  echo ""
  echo "── Code-quality scan (production source, informational) ──"
  ./atheon --json --categories=code-quality,devops,ai-detection . \
    | filter_prod > /tmp/atheon-quality.json
  QC_COUNT=$(jq 'length' /tmp/atheon-quality.json)

  if [ "${QC_COUNT}" -gt 0 ]; then
    warning "${QC_COUNT} code-quality finding(s) in production source (non-blocking):"
    jq -r '.[] | "  [\(.pattern)] \(.file):\(.line)"' /tmp/atheon-quality.json | head -20
    if [ "${QC_COUNT}" -gt 20 ]; then
      info "  ... and $((QC_COUNT - 20)) more (see /tmp/atheon-quality.json)"
    fi
  else
    success "No code-quality findings in production source."
  fi
fi

# ── summary ───────────────────────────────────────────────────────────────────

echo ""
echo "=== Self-Scan Complete ==="

if [ "${EXIT_CODE}" -ne 0 ]; then
  fail "Secrets scan FAILED — fix findings before pushing."
fi

exit "${EXIT_CODE}"
