# Entropy Filtering Implementation Plan

## Phase 1: Core Entropy Module
1. Create `core/entropy.go` with:
   - `shannonEntropy(s string) float64` - calculate Shannon entropy
   - `MinEntropy float64 = 3.0` - minimum entropy threshold
   - `IsHighEntropy(s string) bool` - filter function

## Phase 2: Integration
2. Modify `core/runner.go` `scanLines()`:
   - After regex match, calculate entropy of matched content
   - Skip match if entropy < MinEntropy

3. Add `ENTROPY_THRESHOLD` constant for configuration

## Phase 3: Testing
4. Create `core/entropy_test.go`:
   - Test entropy calculation
   - Test threshold filtering
   - Test edge cases (empty string, single char, etc.)

## Phase 4: Pattern Support
5. Add optional `minEntropy` field to pattern YAML:
   ```yaml
   minEntropy: 3.5
   ```
6. Update bundle loading to support this field
7. Patterns can override default MinEntropy

## Files to Modify
- `core/entropy.go` (new)
- `core/runner.go` (modify scanLines)
- `core/bundle.go` (add minEntropy to PatternDef)
- `core/entropy_test.go` (new)
