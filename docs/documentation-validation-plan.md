# Smart Documentation Validation System

## Overview
Intelligent pre-commit validation that ensures documentation stays current with code changes while being smart enough to avoid false positives for internal changes.

## Design Philosophy
- **Helpful, not annoying**: Guide developers, don't just block commits
- **Context-aware**: Different rules for different types of changes
- **Flexible**: Allow overrides with clear intent
- **Accurate**: Minimize false positives through smart categorization

## File Categorization System

### 1. User-Facing Changes (Require Documentation)
**CLI/Interface Changes:**
- `main.go`, `cmd/` - New commands, flags, behavior changes
- Output format changes
- Error handling changes visible to users

**Pattern Changes:**
- `community/*.yaml` - New patterns, pattern categories
- `core/pattern*.go` - Pattern behavior changes
- `bundler/` - Pattern bundle format changes

**API Changes:**
- Exported functions, interfaces, structs
- Function signatures changing
- New capabilities or features

**Configuration Changes:**
- New environment variables
- Configuration file formats
- Installation requirements

### 2. Internal Changes (Don't Require Documentation)
**Internal Implementation:**
- Refactoring that doesn't change behavior
- Performance optimizations
- Test improvements
- Internal helper functions

**Documentation Changes:**
- Updates to documentation files themselves
- Changelog updates
- README improvements

**Trivial Changes:**
- Typo fixes
- Formatting changes
- Comment improvements
- Variable renaming (internal)

## Documentation Mapping

### Code → Documentation Relationships

| Code Location | Relevant Documentation | Priority |
|---|---|---|
| `main.go`, `cmd/` | README.md, docs/user-guide.md, docs/cli-reference.md | High |
| `community/*.yaml` | docs/patterns/*.md, README (pattern section) | High |
| `core/*.go` (exported) | docs/api.md, examples/ | Medium |
| `core/pattern*.go` | docs/pattern-development.md | Medium |
| Configuration files | docs/configuration.md, INSTALL.md | Medium |
| `.github/workflows/` | docs/development/ci-cd.md | Low |
| Test files | None (unless breaking changes) | None |

### Time-based Validation Rules

**Grace Periods:**
- **Immediate (0-1 hour)**: No validation (developer may be updating docs next)
- **Short-term (1-24 hours)**: Warning only, don't block
- **Long-term (24+ hours)**: Block commit, require explicit override

**Age Calculation:**
```bash
# Get last modification time
CODE_TIME=$(git log -1 --format=%ct -- <code_file>)
DOC_TIME=$(git log -1 --format=%ct -- <doc_file>)

# Calculate age difference
AGE_DIFF=$((DOC_TIME - CODE_TIME))

# Validate based on thresholds
if [ "$AGE_DIFF" -lt "-3600" ]; then
    # Docs updated more than 1 hour before code changes
    # Require documentation update
fi
```

## Implementation Architecture

### 1. Categorization Engine
```bash
# docs/categorize.sh
# Determines if a change requires documentation
categorize_file() {
    local file=$1
    local change_type=$2

    # User-facing changes
    if is_user_facing "$file"; then
        echo "user-facing"
        return 0
    fi

    # Internal changes
    if is_internal "$file"; then
        echo "internal"
        return 0
    fi

    # Documentation changes
    if is_documentation "$file"; then
        echo "documentation"
        return 0
    fi
}
```

### 2. Documentation Validator
```bash
# docs/validate-docs.sh
# Checks if documentation is current for code changes
validate_documentation() {
    local code_file=$1
    local relevant_docs=$2

    for doc in $relevant_docs; do
        if [ -f "$doc" ]; then
            check_age "$code_file" "$doc"
        fi
    done
}
```

### 3. Smart Exemption System
```bash
# docs/exemptions.sh
# Determines when to skip documentation checks
should_skip_docs() {
    # Skip if only test files changed
    if only_test_changes; then
        return 0
    fi

    # Skip if documentation files changed
    if doc_files_changed; then
        return 0
    fi

    # Skip if changelog updated
    if changelog_updated; then
        return 0
    fi

    return 1
}
```

## Integration Points

### Pre-commit Hook Integration
```bash
# .githooks/pre-commit
echo "=== 📚 Documentation Validation ==="

# Run smart documentation check
if ! docs/validate-docs.sh; then
    echo "⚠️  Documentation may need updating"
    echo "Affected files: <list>"
    echo "Consider updating: <relevant docs>"
    echo ""
    read -p "Continue without documentation update? (y/N): " choice
    case "$choice" in
        [Yy]* )
            echo "Proceeding without documentation update"
            ;;
        * )
            echo "Please update relevant documentation or confirm this is an internal change"
            exit 1
            ;;
    esac
fi
```

### CI/CD Integration
```bash
# In GitHub Actions
- name: Documentation freshness check
  run: |
    if [ "${{ github.event_name }}" == "pull_request" ]; then
      ./docs/validate-docs.sh --strict
    fi
```

## Configuration and Customization

### Configuration File
```yaml
# docs/validation-config.yaml
documentation_rules:
  user_facing:
    files:
      - "main.go"
      - "cmd/**/*"
      - "community/*.yaml"
    required_docs:
      - "README.md"
      - "docs/user-guide.md"
    grace_period: "1h"

  internal:
    files:
      - "internal/**/*"
      - "test/**/*"
    required_docs: []
    always_skip: true

  patterns:
    files:
      - "community/*.yaml"
      - "core/pattern*.go"
    required_docs:
      - "docs/patterns/*.md"
      - "README.md"
    grace_period: "24h"
```

## User Experience

### Clear Messaging
```
=== 📚 Documentation Validation ===

🔍 Detected user-facing changes in:
  • main.go (new CLI flag --json-output)
  • community/secrets/api-keys.yaml (new API key pattern)

📚 Relevant documentation that may need updating:
  • README.md (CLI usage section)
  • docs/user-guide.md (JSON output format)
  • docs/patterns/api-keys.md (new pattern documentation)

⏰ Documentation age:
  • README.md: Updated 2 days ago (before code changes)
  • docs/user-guide.md: Updated 1 week ago (stale)

💡 Recommendation: Update user-guide.md for new JSON output format

Continue without documentation update? (y/N):
```

### Override Options
```bash
# Explicit override for internal changes
git commit -m "refactor: improve pattern matching performance" \
  -m "[skip-docs] Internal optimization, no user-facing changes"

# Automatic documentation update
git commit -m "feat: add JSON output format" \
  -m "[doc-update] Updated user-guide.md with JSON examples"
```

## Advanced Features

### 1. Semantic Analysis
```bash
# Analyze commit message for intent
if commit_message_contains "internal|refactor|perf"; then
    skip_documentation_check
fi

if commit_message_contains "feat|add|new"; then
    require_documentation_update
fi
```

### 2. Impact Analysis
```bash
# Determine scope of changes
if affects_public_api; then
    priority="high"
elif affects_patterns; then
    priority="medium"
else
    priority="low"
fi
```

### 3. Documentation Templates
```bash
# Generate template for missing documentation
generate_doc_template() {
    local feature=$1
    local template="docs/templates/$feature.md"

    cat > "$template" << EOF
# $feature Documentation

## Overview
[Description of the feature]

## Usage
[Examples and usage patterns]

## Configuration
[Configuration options if applicable]

## Examples
[Real-world examples]
EOF

    echo "Created documentation template: $template"
}
```

## Testing and Validation

### Test Cases
1. **User-facing feature without docs**: Should block or warn
2. **Internal refactoring**: Should skip automatically
3. **Documentation update**: Should recognize and pass
4. **Mixed changes**: Should handle appropriately
5. **Override with intent**: Should respect explicit intent

### Validation Script
```bash
# docs/test-validation.sh
test_cases=(
    "user_facing_no_docs:fail"
    "internal_refactor:pass"
    "doc_update:pass"
    "mixed_changes:warn"
)

for test_case in "${test_cases[@]}"; do
    run_test_case "$test_case"
done
```

## Rollout Plan

### Phase 1: Warning Only (1 week)
- Implement validation in warning mode
- Gather feedback and tune rules
- Fix false positives

### Phase 2: Soft Block (1 week)
- Block commits but easy override
- Ensure team adapts to workflow
- Continue tuning

### Phase 3: Full Enforcement (ongoing)
- Standard enforcement with clear override paths
- Regular review of exemption rules
- Continuous improvement

## Metrics and Success Criteria

### Success Metrics
- **False positive rate**: < 10% (changes flagged that don't need docs)
- **Documentation coverage**: > 80% of features documented within grace period
- **Developer satisfaction**: Minimal friction, helpful guidance

### Monitoring
```bash
# Track validation statistics
docs_stats --monthly
# Output:
# Documentation validations: 150
# Warnings issued: 30 (20%)
# Blocks issued: 10 (7%)
# Overrides used: 5 (3%)
# False positives: 2 (1%)
```

## Maintenance and Evolution

### Regular Updates
- Update categorization rules as project evolves
- Adjust grace periods based on team feedback
- Expand documentation mappings as new features added

### Feedback Loop
```bash
# Report false positive
git config docs.validation.skip "reason: false positive"
# or
gh issue create --title "Documentation Validation False Positive" \
  --body "File X doesn't require documentation update because..."
```

---

## Implementation Checklist

- [ ] Create categorization database
- [ ] Implement file categorization script
- [ ] Build documentation validator
- [ ] Add smart exemption logic
- [ ] Integrate with pre-commit hook
- [ ] Create configuration system
- [ ] Implement override mechanisms
- [ ] Add comprehensive messaging
- [ ] Test with various change types
- [ ] Roll out in warning mode
- [ ] Gather feedback and tune
- [ ] Enable full enforcement
- [ ] Monitor and improve continuously

---

**Status**: Planning Phase
**Next Steps**: Create categorization scripts and validation logic
**Maintainer**: Micheal Kinney (aliasfoxkde)