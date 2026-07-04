# Code Quality & Feature Enhancement Roadmap

**Date**: 2026-07-03
**Based on**: Comprehensive SAST tool research and pattern analysis

---

## Executive Summary

Atheon-Enhanced has **384 patterns across 26 categories** with solid foundations in AST-based analysis. Research reveals opportunities to enhance precision, reduce false positives, and add verification capabilities comparable to industry leaders like TruffleHog and GitLeaks.

---

## Part 1: Pattern Quality Issues (Immediate)

### High-False-Positive Patterns to Fix/Disable

| Pattern | Issue | Recommendation |
|---------|-------|----------------|
| `ai-detection/ai-buzzwords.yaml` | Matches legitimate marketing text | Disable or refine with word boundaries |
| `ai-detection/ai-overuse.yaml` | Common English phrases | Disable |
| `ai-detection/ai-template.yaml` | Code review phrases | Disable or make stricter |
| `code-quality/todo-comment.yaml` | Overly broad `TODO` | Combine with existing TODO/FIXME patterns |
| `code-quality/console-log.yaml` | Matches legitimate logging | Add path exclusions or disable |
| `container/dockerfile-running-as-root.yaml` | Broken regex (`\\` vs `\`) | Fix escaping |

### Disabled Patterns to Re-evaluate

| Pattern | Current Status | Recommendation |
|---------|---------------|----------------|
| `secrets/certificate-block.yaml` | `enabled: false` | Enable - high value |
| `pii/ip-address.yaml` | `enabled: false` | Enable - commonly needed |
| `pii/ipv6-address.yaml` | `enabled: false` | Enable - commonly needed |

### Consolidation Opportunities

1. **TODO/FIXME family** (5+ patterns → 2 patterns)
   - Strict: `(?i)\b(TODO|FIXME|XXX|HACK|DEPRECATED):\s*\S`
   - Lenient: `(?i)\b(TODO|FIXME|XXX|HACK)\b`

2. **GitHub Actions Secret** (3 duplicates → 1)
   - Consolidate into `secrets/github-actions-secret.yaml`

---

## Part 2: Detection Enhancements (Near Term)

### 1. Entropy-Based Filtering

**Current**: Pure regex matching
**Add**: Shannon entropy calculation per match

```go
// Pseudocode
entropy := calculateShannonEntropy(matchedString)
if entropy < 3.0 {
    return // Low entropy - likely false positive
}
```

**Impact**: Significantly reduces false positives on high-entropy lookalikes

### 2. Automatic Decoding Pipeline

**Add pre-processing** before pattern matching:
- Base64 decoding (>=16 chars)
- Hex decoding (>=32 chars)
- Percent decoding

**Benefit**: Detects secrets hidden in encoded strings

### 3. Composite/Multi-Pattern Rules

Similar to GitLeaks `[[rules.required]]`:
```yaml
name: aws-credentials-pair
match: '\bAKIA[0-9A-Z]{16}\b'
required:
  - pattern: 'aws_secret_access_key|aws_secret'
    withinLines: 5
```

**Use cases**:
- AWS access key + secret in same file
- Database connection strings
- Multi-part OAuth tokens

### 4. Context-Aware Anchoring

**Add file-type sensitivity**:
- `.env`, `. secrets`, `config.yaml` → High confidence
- `README.md`, `test/**` → Lower confidence
- Add path-based confidence multipliers

### 5. Confidence Metadata

**Add per-pattern confidence**:
```yaml
name: aws-access-key
confidence: high  # Tested, low FP rate
---
name: generic-api-key
confidence: low   # Needs tuning
```

Display in output for risk-based triage.

---

## Part 3: Verification System (Medium Term)

### TruffleHog-Style API Verification

| Provider | Verification Method |
|----------|-------------------|
| AWS | STS `GetCallerIdentity` |
| GitHub | `POST /applications/{id}/tokens` |
| Stripe | `GET /v1/customers` |
| Slack | `auth.test` |

**Architecture**:
```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│  Scanner    │───▶│  Verifier    │───▶│  Results    │
│  (regex)   │    │  (API calls) │    │ (verified) │
└─────────────┘    └──────────────┘    └─────────────┘
```

### Webhook-Based Verification

```yaml
name: custom-api-key
match: 'custom_key_[A-Z0-9]+'
verify:
  webhook: https://example.com/verify
  method: POST
  body: '{"key": "$MATCH"}'
```

---

## Part 4: Benchmark & Metrics (Medium Term)

### Add Precision/Recall Metrics

| Metric | Formula | Target |
|--------|---------|--------|
| Precision | TP / (TP + FP) | > 90% |
| Recall | TP / (TP + FN) | > 80% |
| F1 Score | 2 × P × R / (P + R) | > 85% |

### Benchmark Dataset Structure

```
benchmark/
├── secrets/
│   ├── aws-keys/
│   ├── github-tokens/
│   └── stripe-keys/
├── clean/
│   ├── open-source-repos/
│   └── synthetic-clean/
└── expected-results.json
```

---

## Part 5: Missing Categories (Long Term)

| Category | Priority | Patterns to Add |
|----------|----------|-----------------|
| **Serialization** | High | pickle, YAML unsafe load, Java serializable |
| **Authentication** | High | Generic auth bypass patterns |
| **Cryptocurrency** | Medium | Wallet addresses, exchange API keys |
| **ML/AI** | Medium | Model files, training data exposure |
| **CI/CD (GitLab, Azure)** | Medium | GitLab CI, Azure Pipelines |
| **SBOM** | Low | Known vulnerable components |

---

## Part 6: Performance Enhancements

### Current Performance Bottlenecks

| Issue | Location | Recommendation |
|-------|----------|----------------|
| fmt.Sprintf in hot loop | `runner.go:419,495` | Use strings.Builder |
| Double JSON parse | `bundle.go:169-200` | Single-pass with Go 1.24+ |
| No ignore matcher cache | `runner.go:137,199` | Per-root caching |

### Optimization Priorities

1. **Fingerprint generation** - Use `strings.Builder` instead of `fmt.Sprintf`
2. **Ignore matcher caching** - Cache compiled patterns per root directory
3. **HTTP client reuse** - Reuse clients for connection pooling

---

## Part 7: CI/CD Improvements

### Already Fixed ✅
- Dev-testing exit code (coverage.out missing → exit 1)

### Remaining CI Issues

| Issue | Workflow | Priority |
|-------|----------|----------|
| golangci-lint version mismatch | `release.yml` vs `ci.yml` | Medium |
| 10-file cap in pattern review | `community-pattern-review.yml` | Low |
| Polling timeout ~20min | `auto-merge.yml` | Medium |

---

## Implementation Priority Matrix

| Priority | Item | Effort | Impact |
|----------|------|--------|--------|
| **P0** | Fix broken patterns | Low | High |
| **P0** | Disable high-FP AI patterns | Low | High |
| **P1** | Add entropy filtering | Medium | High |
| **P1** | Add confidence metadata | Medium | Medium |
| **P1** | Consolidate TODO/FIXME patterns | Low | Medium |
| **P2** | Automatic decoding pipeline | High | Medium |
| **P2** | Composite rules | High | Medium |
| **P2** | Context-aware anchoring | High | Medium |
| **P3** | API verification system | Very High | High |
| **P3** | Precision/recall benchmark | Medium | High |

---

## Research Sources

- [TruffleHog](https://github.com/trufflesecurity/trufflehog) - Secret verification leader
- [GitLeaks](https://github.com/gitleaks/gitleaks) - Composite rules pioneer
- [Semgrep](https://semgrep.dev) - AST-based analysis
- [Nuclei](https://github.com/projectdiscovery/nuclei) - Template-based scanning
- [Bandit](https://github.com/PyCQA/bandit) - Python AST analysis

---

## Conclusion

Atheon-Enhanced has a strong foundation with 384 patterns and AST capabilities. The highest-impact improvements are:

1. **Immediate**: Fix broken patterns, disable high-false-positive AI detection patterns
2. **Near-term**: Add entropy filtering and confidence metadata
3. **Medium-term**: Composite rules and verification system
4. **Long-term**: Full benchmark dataset with precision/recall metrics

The tool is well-positioned to compete with TruffleHog and GitLeaks if entropy filtering and verification are added.
