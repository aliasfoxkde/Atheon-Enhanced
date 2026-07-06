# Entropy Filtering Tasks

## Task 12: Add Entropy-Based Filtering

### Code Changes
- [ ] Create `core/entropy.go` with `shannonEntropy()` function
- [ ] Add `MinEntropy` constant (3.0)
- [ ] Add `IsHighEntropy()` filter function
- [ ] Create `core/entropy_test.go` with test coverage
- [ ] Modify `core/runner.go` to filter low-entropy matches
- [ ] Update `core/bundle.go` to support `minEntropy` pattern field

### Pattern Changes
- [ ] Add `minEntropy: 3.5` to high-value secret patterns

### Testing
- [ ] Verify low-entropy strings are filtered
- [ ] Verify high-entropy secrets are detected
- [ ] Run full test suite

---

## Task 13: Add Confidence Metadata

### Pattern Changes
- [ ] Add `confidence` field to pattern YAML schema
- [ ] Update `core/bundle.go` PatternDef to include confidence
- [ ] Display confidence in CLI output
- [ ] Add CLI flag `--confidence=high,medium,low` for filtering

### Confidence Levels
- `high`: Tested, low FP rate (AWS keys, GitHub tokens)
- `medium`: Generally reliable, some FP (generic API keys)
- `low`: Higher FP rate, needs tuning (AI detection patterns)

### Testing
- [ ] Test confidence filtering in CLI
- [ ] Verify confidence display in JSON output
