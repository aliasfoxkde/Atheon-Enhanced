# Contributing Patterns

How to add a new pattern to the Atheon community library.

## Pattern File Format

Each pattern is a single YAML file in `community/<category>/`. The filename becomes the pattern's file identifier — use lowercase with hyphens.

```yaml
name: my-pattern-name        # lowercase, hyphenated, unique across all categories
category: secrets            # must match the directory name
match: "regex-pattern"       # RE2-compatible regex
enabled: true                # false = opt-in only (pattern loads but is disabled by default)
description: "What this detects and why it matters"
```

### Minimal valid pattern

```yaml
name: github-personal-access-token
match: '\bghp_[A-Za-z0-9]{36}\b'
```

### Full pattern

```yaml
name: github-personal-access-token
category: secrets
match: '\bghp_[A-Za-z0-9]{36}\b'
enabled: true
description: "GitHub personal access token (classic). Full repo/user scope. Rotate immediately if found in code."
```

## RE2 Constraint

Go uses the **RE2 regex engine**. These features are NOT supported:

| Unsupported | Why | Alternative |
|-------------|-----|-------------|
| `(?=...)` lookahead | RE2 limitation | Encode context into the match itself |
| `(?<=...)` lookbehind | RE2 limitation | Use prefix group: `(PREFIX)(VALUE)` |
| `\1` backreferences | RE2 limitation | Not possible — simplify the pattern |
| `(?i)` inline flag mid-pattern | Partially supported | Use `(?i)` at the start only |

**Test your regex with Go's regexp package:**

Save the snippet below as `regex_check.go` and run `go run regex_check.go` — Go's `run` command does not support a `-e` flag, so a here-doc style is the canonical way to iterate quickly.

```go
package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`\bghp_[A-Za-z0-9]{36}\b`)
	fmt.Println(re.MatchString("ghp_abc123def456ghi789jkl012mno345pqr6"))
}
```

Or test directly with Atheon:

```bash
echo 'token = "ghp_abc123def456ghi789jkl012mno345pqr6"' | ./atheon --string -
```

## Categories

Place your pattern in the most specific matching category:

| Category | Use for |
|----------|---------|
| `secrets` | API keys, tokens, credentials, private keys |
| `pii` | Personal identifiable information (SSN, credit cards, emails in code) |
| `code-quality` | Debug artifacts, hardcoded values, dead code markers |
| `devops` | CI/CD bypass flags, pipeline secrets, Docker/K8s misconfigs |
| `security-hardening` | Insecure function calls, weak crypto, unsafe configs |
| `web-security` | XSS vectors, SQLi patterns, CORS misconfigs |
| `ai-detection` | AI-generated code markers, LLM prompt injection |
| `compliance` | GDPR, HIPAA, PCI field patterns |

If none fit, propose a new category in your PR description.

## False Positive Risk

Before opening a PR, verify your pattern against common false positives:

1. **Variable name collisions** — does the regex match common variable names like `token`, `key`, `id`? If so, tighten the prefix.
2. **Test fixture matches** — does it match your own test fixtures or documentation examples? Add `enabled: false` for patterns that commonly fire in test code.
3. **Generality** — a pattern like `[0-9]{16}` matches credit cards AND UUIDs AND phone numbers. Add structural context (Luhn-prefixes, separator patterns) to narrow the match.

**Quick false positive test:**

```bash
# Run against a diverse open-source repo
git clone https://github.com/golang/go /tmp/golang-go --depth=1
./atheon /tmp/golang-go --categories=secrets 2>/dev/null | grep "my-pattern-name" | wc -l
```

If you see more than a handful of hits in a standard codebase, the pattern needs narrowing.

## Severity Guidelines

Use the `description` field to communicate risk level. Categories imply severity:
- `secrets`: always high impact — rotation required if found
- `pii`: medium–high impact — depends on context (code vs. real data)
- `code-quality`: low–medium impact — technical debt
- `devops`: medium impact — pipeline security
- `security-hardening` / `web-security`: varies — document the specific risk

## Step-by-Step: Adding a Pattern

```bash
# 1. Create the pattern file
cat > community/secrets/my-new-token.yaml << 'EOF'
name: my-new-token
category: secrets
match: '\bmnt_[A-Za-z0-9]{32}\b'
enabled: true
description: "MyService API token — provides full API access."
EOF

# 2. Rebuild the pattern bundle
go run ./bundler

# 3. Verify the pattern loads
./atheon list | grep my-new-token

# 4. Test a positive match
echo 'api_key = "mnt_abc123def456ghi789jkl012mno34567"' | ./atheon --string -

# 5. Run the full test suite
go test ./... -p 1 -timeout 15m

# 6. Commit and open a PR
git checkout -b feat/pattern-my-new-token
git add community/secrets/my-new-token.yaml core/patterns.bundle
git commit -m "feat(patterns): add my-new-token pattern"
```

## PR Checklist

- [ ] Pattern file is in the correct `community/<category>/` directory
- [ ] `name` is unique (run `grep -r "^name: my-pattern-name" community/` to verify)
- [ ] Regex compiles under RE2 (tested with `go run` or `./atheon --string -`)
- [ ] `core/patterns.bundle` is regenerated (`go run ./bundler`)
- [ ] `go test ./... -p 1` passes
- [ ] Description explains what is detected and why it's risky
- [ ] No more than ~10 false positives on a representative open-source repo

## Common Mistakes

**Forgot to run bundler:** The pattern file exists but `./atheon list` doesn't show it. Run `go run ./bundler` to regenerate `core/patterns.bundle`.

**Lookahead in regex:** `(?=...)` compiles with PCRE tools but panics at runtime in Go. Use `re2online.com` or Go's `regexp` package to validate.

**Pattern too broad:** `\b[A-Z0-9]{32}\b` will match MD5 hashes, UUIDs, and thousands of other things. Add a token-specific prefix like `sk_`, `ghp_`, `AKIA`, etc.

**Name collision:** Two patterns with the same `name` field — the second silently overwrites the first. Keep names unique and descriptive.
