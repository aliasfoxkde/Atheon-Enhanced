# Pattern Expansion: Issue #149

## Context
Based on real-world repo benchmarking via Atheon-GitHub-Scanner and Atheon-Benchmark projects.

## Reference Projects

### Atheon-GitHub-Scanner (`/nas/Temp/repos/Atheon-GitHub-Scanner`)
Scans real GitHub repositories to find security issues and identify pattern gaps.

### Atheon-Benchmark (`/nas/Temp/repos/Atheon-Benchmark`)
Benchmarks AI code generation using Atheon's pattern scanning as quality gates:
- Uses **185+ patterns** from Atheon bundle
- Validates code across **8 categories**
- Tests AI outputs for security and quality issues
- Reference: `/nas/Temp/repos/Atheon-Benchmark/dashboard/lib/atheon/quality-gates.ts`

## Benchmark Results Summary (Atheon-GitHub-Scanner)

From `/nas/Temp/repos/Atheon-GitHub-Scanner/pipeline_results.json`:
- Repositories scanned: 5
- Total findings: 55
- Trending patterns discovered: 12
- PRs created from findings: 2

## Validated Pattern Suggestions (from Atheon-GitHub-Scanner)

The Atheon-GitHub-Scanner benchmarked 2 pattern candidates with accuracy scores:

### 1. API Key Exposure in Configuration Files
- **Pattern ID**: `pattern_api_key_leak_001`
- **Severity**: critical
- **Confidence**: 0.92
- **Benchmark Score**: 75
- **Accuracy**: 85%
- **Pattern**: `(?:\api[_-]?key|apikey|secret|token)\s*[:=]\s*['"]?([a-zA-Z0-9_-]{16,})['"]?`
- **CWE**: CWE-798
- **OWASP**: A07:2021 - Identification and Authentication Failures
- **Source**: `popular-javascript-lib/config/database.js`
- **Examples**:
  - `config.API_KEY = "sk_live_1234567890abcdef"`
  - `const apiKey = process.env.SECRET_KEY;`

### 2. SQL Injection via String Concatenation
- **Pattern ID**: `pattern_sql_injection_002`
- **Severity**: high
- **Confidence**: 0.87
- **Benchmark Score**: 80
- **Accuracy**: 88%
- **Pattern**: `(?:SELECT|INSERT|UPDATE|DELETE)\s+.*?\s+(?:WHERE|SET)\s+.*?(?:\+|concat\()`
- **CWE**: CWE-89
- **OWASP**: A03:2021 - Injection
- **Source**: `awesome-python-project/models/user.py`
- **Examples**:
  - `"SELECT * FROM users WHERE id = " + userInput`
  - `"INSERT INTO logs VALUES ('" + msg + "')"`

## High-Priority Pattern Categories (Based on Findings)

### 1. API Keys & Secrets (High Priority)
From benchmark findings, these patterns are most frequently found:
- AWS access keys (already exists: `aws-access-key`)
- Azure client secrets (already exists: `azure-client-secret`)
- GCP API keys (already exists: `gcp-api-key`)
- Generic API key patterns (needs enhancement)

### 2. SQL Injection (High Priority)
Benchmark found this in `awesome-python-project`:
- String concatenation in SQL queries
- Dynamic SQL construction without parameterization

### 3. Infrastructure as Code (Medium Priority)
From 55 total findings, common patterns:
- Kubernetes secrets in env vars
- Terraform state files with sensitive data
- Docker config with credentials

### 4. CI/CD Specific (Medium Priority)
- GitHub Actions secrets exposure (already exists: `github-actions-secret`)
- GitLab CI variable usage
- Jenkins credentials (already exists: `jenkins-crumb`)
- CircleCI context leaks

### 5. Database Connection Strings (Medium Priority)
- Redis connection strings (already exists: `redis-connection-string`)
- MongoDB connection strings (already exists: `mongodb-connection-string`)
- PostgreSQL connection strings (already exists: `postgres-connection-string`)

## Gaps Identified by Benchmark

The benchmark identified these gaps not covered by existing patterns:
1. **SQL Injection via string concatenation** - NOT in current patterns → ADDED
2. **Generic API key patterns** - partial coverage only → ADDED
3. **Environment variable secrets in IaC** - needs new category

## Implementation (Completed)

### Patterns Added
1. `community/code-quality/sql-injection-string-concat.yaml` - SQL injection via string concat
2. `community/secrets/generic-api-key-config.yaml` - Generic API key in config files

Both patterns validated against Atheon-GitHub-Scanner benchmark findings.

### Validation via Atheon-Benchmark
The added patterns can be validated using Atheon-Benchmark's quality gates:
```typescript
// Reference: /nas/Temp/repos/Atheon-Benchmark/dashboard/lib/atheon/quality-gates.ts
// Uses 185+ patterns across 8 categories for validation
```

## Source Data References

- Atheon-GitHub-Scanner pipeline: `/nas/Temp/repos/Atheon-GitHub-Scanner/pipeline_results.json`
- Combined scan results: `/nas/Temp/repos/Atheon-GitHub-Scanner/data/combined_scan_results.json`
- Atheon-Benchmark quality gates: `/nas/Temp/repos/Atheon-Benchmark/dashboard/lib/atheon/quality-gates.ts`
- Atheon-Benchmark architecture: `/nas/Temp/repos/Atheon-Benchmark/docs/ARCHITECTURE.md`

---

**Date:** 2026-06-22
**Branch:** `pr/149-patterns-expansion`
**Sources:**
- Atheon-GitHub-Scanner (5 repos, 55 findings, 2 validated patterns)
- Atheon-Benchmark (185+ patterns, 8 categories for validation)
