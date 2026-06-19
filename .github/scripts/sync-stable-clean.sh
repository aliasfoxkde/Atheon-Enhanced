#!/bin/bash
# Automated sync script for stable/clean branch
# This script keeps the stable/clean branch synchronized with upstream Atheon

set -e

echo "=== Syncing stable/clean with upstream HoraDomu/Atheon ==="

# Checkout stable/clean branch
git checkout stable/clean || git checkout -b stable/clean origin/main

# Fetch latest from upstream
echo "Fetching from upstream..."
git fetch upstream

# Merge upstream main
echo "Merging upstream/main into stable/clean..."
git merge upstream/main -m "Sync with upstream HoraDomu/Atheon main ($(date +%Y-%m-%d))"

# Push to origin
echo "Pushing updated stable/clean to origin..."
git push origin stable/clean

echo "=== Sync complete ==="
echo "stable/clean is now up to date with upstream HoraDomu/Atheon main"