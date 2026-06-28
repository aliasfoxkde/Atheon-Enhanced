# Community Patterns

Pattern files for Atheon-Enhanced, organized by category. Each `.yaml` file defines one pattern.

## Categories

| Category | Patterns | Description |
|----------|----------|-------------|
| [accessibility](accessibility/) | 19 | WCAG compliance, ARIA, screen reader issues |
| [ai-detection](ai-detection/) | 21 | AI-generated code markers, anti-cheating patterns |
| [api-integration](api-integration/) | 9 | API keys, webhook secrets, integration tokens |
| [cloud-native](cloud-native/) | 14 | Kubernetes, Docker, cloud deployment patterns |
| [code-quality](code-quality/) | 53 | Debug artifacts, hardcoded values, dead code |
| [compliance](compliance/) | 4 | GDPR, HIPAA, PCI compliance patterns |
| [data-visualization](data-visualization/) | 5 | Chart/graph library patterns |
| [devops](devops/) | 22 | CI/CD bypass markers, pipeline secrets |
| [finance](finance/) | 6 | Payment identifiers, financial data |
| [frameworks](frameworks/) | 8 | Django, Node.js, React, Vue framework patterns |
| [git-hygiene](git-hygiene/) | 4 | Merge conflicts, fixup commits, git hygiene |
| [git-ops](git-ops/) | 3 | GitOps workflow patterns |
| [healthcare](healthcare/) | 7 | PHI, HIPAA-relevant field patterns |
| [kubernetes](kubernetes/) | 6 | Kubernetes-specific patterns |
| [metadata](metadata/) | 4 | File metadata patterns |
| [performance](performance/) | 12 | Blocking calls, synchronous patterns |
| [pii](pii/) | 15 | Personally identifiable information |
| [pwa](pwa/) | 5 | Progressive Web App patterns |
| [secrets](secrets/) | 66 | API keys, tokens, credentials, private keys |
| [security-hardening](security-hardening/) | 21 | Insecure configs, weak crypto, unsafe calls |
| [terraform](terraform/) | 4 | Terraform-specific patterns |
| [web-development](web-development/) | 12 | Frontend anti-patterns |
| [web-security](web-security/) | 15 | XSS, SQLi, CORS, injection risks |

**Total: 335 patterns across 23 active categories**

## Pattern File Format

```yaml
name: pattern-name          # lowercase, hyphenated, unique
category: secrets           # must match directory name
match: "regex-pattern"      # RE2-compatible (no lookahead/lookbehind)
enabled: true               # false = opt-in only
description: "What it detects"
severity: high              # high, medium, low (optional)
```

> **RE2 constraint**: Go uses the RE2 engine. No lookahead (`(?=...)`), lookbehind (`(?<=...)`), or backreferences (`\1`). Use prefix/suffix context in the pattern itself instead.

## Adding a Pattern

1. Create `community/<category>/<name>.yaml`
2. Run `go run ./bundler` to rebuild `core/patterns.bundle`
3. Run `go test ./... -p 1` to verify all tests pass
4. Open a PR — CI validates the bundle automatically

See [docs/guides/PATTERN_AUTHORING.md](../docs/guides/PATTERN_AUTHORING.md) or the [CONTRIBUTING](../.github/CONTRIBUTING.md) guide.

## Category README Files

Each category folder contains a `README.md` that documents all patterns in that category with descriptions.
