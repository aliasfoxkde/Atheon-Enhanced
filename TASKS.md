# Entropy Filtering Tasks

## Task 12: Add Entropy-Based Filtering

### Code Changes
- [x] Create RESEARCH.md
- [x] Create PLAN.md
- [x] Create TASKS.md
- [ ] Create `core/entropy.go` with `shannonEntropy()` function
- [ ] Add `MinEntropy` constant (3.0)
- [ ] Add `IsHighEntropy()` filter function
- [ ] Create `core/entropy_test.go` with test coverage
- [ ] Modify `core/runner.go` to filter low-entropy matches
- [ ] Run full test suite

### Testing
- [ ] Verify low-entropy strings are filtered
- [ ] Verify high-entropy secrets are detected

---

## Task 13: Add Confidence Metadata

### Pattern Changes
- [ ] Add `confidence` field to pattern YAML schema
- [ ] Update `core/bundle.go` PatternDef to include confidence
- [ ] Display confidence in CLI output
- [ ] Add CLI flag for confidence filtering
