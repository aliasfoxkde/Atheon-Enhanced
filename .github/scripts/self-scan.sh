#!/bin/bash
# Self-scanning integration script
# This script runs Atheon on itself to catch security issues and code quality problems

set -e

echo "=== Atheon Self-Scan Integration ==="

# Build Atheon first
echo "Building Atheon..."
go build -o atheon . || { echo "Build failed"; exit 1; }

# Create scan results directory
mkdir -p scan-results

# Scan with different configurations
echo "Scanning with production profile..."
./atheon . --profile config/profiles/production.json > scan-results/production-findings.json 2>&1 || true

echo "Scanning with development profile..."
./atheon . --profile config/profiles/development.json > scan-results/development-findings.json 2>&1 || true

echo "Scanning with all patterns enabled..."
./atheon . --all > scan-results/all-patterns-findings.json 2>&1 || true

# Analyze results
echo "=== Scan Analysis ==="

# Count findings by category
echo "Production profile findings:"
jq '[.[] | .pattern] | group_by(.) | map({pattern: .[0], count: length}) | sort_by(-.count)' scan-results/production-findings.json || echo "No production findings"

echo "Development profile findings:"
jq '[.[] | .pattern] | group_by(.) | map({pattern: .[0], count: length}) | sort_by(-.count)' scan-results/development-findings.json || echo "No development findings"

echo "All patterns findings:"
jq '[.[] | .pattern] | group_by(.) | map({pattern: .[0], count: length}) | sort_by(-.count)' scan-results/all-patterns-findings.json || echo "No all-patterns findings"

# Check for critical security issues
echo "=== Critical Security Check ==="
CRITICAL=$(jq '[.[] | select(.pattern | test("api-key|secret|token|password"; "i"))] | length' scan-results/all-patterns-findings.json || echo "0")

if [ "$CRITICAL" -gt 0 ]; then
    echo "⚠️  Found $CRITICAL potential security issues in codebase"
    echo "Review findings before committing to main"

    # Show some examples
    echo "Sample critical findings:"
    jq '[.[] | select(.pattern | test("api-key|secret|token|password"; "i"))] | .[:5]' scan-results/all-patterns-findings.json

    # Ask for confirmation
    read -p "Continue despite security warnings? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Commit cancelled due to security concerns"
        exit 1
    fi
else
    echo "✅ No critical security issues found"
fi

echo "=== Self-Scan Complete ==="