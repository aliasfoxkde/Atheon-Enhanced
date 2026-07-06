# Entropy Filtering Implementation Plan

## Phase 1: Core Entropy Module
1. Create `core/entropy.go` with:
   - `shannonEntropy()` - calculate Shannon entropy
   - `MinEntropy` constant (3.0)
   - `IsHighEntropy()` - filter function

## Phase 2: Integration
2. Modify `core/runner.go` `scanLines()`:
   - After regex match, calculate entropy of matched content
   - Skip match if entropy < MinEntropy

## Phase 3: Testing
3. Create `core/entropy_test.go` with test coverage

## Phase 4: Pattern Support (Future)
4. Add optional `minEntropy` field to pattern YAML

## Files to Create/Modify
- `core/entropy.go` (new)
- `core/entropy_test.go` (new)
- `core/runner.go` (modify)
