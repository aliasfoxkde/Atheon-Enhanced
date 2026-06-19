# Pattern Development Guide

## Creating Your First Pattern

Patterns are simple YAML files with two required fields:

```yaml
# community/secrets/my-service.yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
```

**The folder name becomes the category**: `secrets/` → category: `secrets`

## Pattern Best Practices

### 1. Use Word Boundaries
```yaml
# ❌ BAD - matches substrings
match: 'sk-[A-Za-z0-9]{32}'

# ✅ GOOD - whole tokens only
match: '\bsk-[A-Za-z0-9]{32}\b'
```

### 2. Be Specific
```yaml
# ❌ TOO GENERIC
match: '[A-Z]{2}-[A-Z]{5}'

# ✅ SPECIFIC PREFIX
match: '\bAWS-[A-Z]{2}-[A-Z]{5}[A-Z0-9]{16}\b'
```

### 3. Consider False Positives
```yaml
# ❌ MANY FALSE POSITIVES
match: 'token[A-Z]*'

# ✅ CONTEXT AWARE
match: 'api[_-]?token["\s]*[:=]["\s]*[A-Za-z0-9]{32,}'
```

## Pattern Categories

### Existing Categories
- **secrets/** - API keys, tokens, credentials
- **pii/** - Personal information (SSN, credit cards, etc.)
- **code-quality/** - Debug statements, TODOs, technical debt
- **healthcare/** - Medical identifiers, PHI patterns
- **finance/** - IBAN, ABA routing numbers, SWIFT/BIC codes
- **ai-detection/** - AI-generated code markers
- **devops/** - CI/CD, Docker, Kubernetes patterns
- **frameworks/** - Framework-specific patterns

### Creating New Categories
```bash
# Create category directory
mkdir -p community/my-category

# Add pattern
echo 'name: my-pattern
match: "my-regex"' > community/my-category/my-pattern.yaml
```

## Testing Your Pattern

### 1. Create Pattern File
```yaml
# community/secrets/test-api-key.yaml
name: test-api-key
match: '\btest_[A-Za-z0-9]{32}\b'
```

### 2. Build and Test
```bash
# Build bundle
go run ./bundler

# Test pattern
echo "TEST_12345678901234567890123456789012" | atheon -
```

### 3. Validate Pattern
```bash
# Test pattern explicitly
atheon list --pattern test-api-key

# Test with real content
echo "some content TEST_12345678901234567890123456789012 more content" | atheon -
```

## Advanced Patterns

### Context-Aware Patterns
```yaml
# API key in assignment
name: api-key-assignment
match: 'api[_-]?key["\s]*[:=]["\s]*[A-Za-z0-9]{32,}'
```

### Multi-Line Patterns
```yaml
# SSH private key block
name: ssh-private-key
match: '(?m)-----BEGIN RSA PRIVATE KEY-----'
```

### Case-Insensitive Patterns
```yaml
# Password references (case insensitive)
name: password-reference
match: '(?i)password["\s]*[:=]'
```

## Pattern Metadata (Future)

Enhanced version will support additional fields:

```yaml
name: stripe-api-key
match: '\bsk_(test|live)_[0-9a-z]{32}\b'
severity: critical
confidence: high
description: Detects Stripe API keys
references:
  - https://stripe.com/docs/api/security
tags:
  - payment
  - api-key
```

## Contributing Patterns

### 1. Choose Category
Select the most appropriate existing category or create a new one.

### 2. Write Pattern
Create pattern file with clear, descriptive name and specific regex.

### 3. Test Locally
```bash
# Test pattern
atheon --all . | grep my-pattern-name

# Test with real examples
# Verify no false positives on typical code
```

### 4. Validate
```bash
# Check pattern syntax
atheon list --pattern my-pattern-name

# Test with various inputs
echo "test cases" | atheon -
```

### 5. Submit PR
```bash
# Create branch
git checkout -b patterns/add-my-pattern stable/clean

# Add pattern file
# Test thoroughly
# Commit changes
git commit -m "patterns: add my-pattern detection"

# Create PR
gh pr create --base main --head patterns/add-my-pattern
```

## Pattern Quality Checklist

- ✅ **Specificity**: Uses word boundaries `\b` where appropriate
- ✅ **Accuracy**: Minimal false positives on real code
- ✅ **Clarity**: Pattern name clearly describes what it detects
- ✅ **Testing**: Tested with real examples and edge cases
- ✅ **Documentation**: Includes description if detection isn't obvious
- ✅ **Performance**: Efficient regex without catastrophic backtracking

## Common Pattern Mistakes

### ❌ Overly Generic
```yaml
# Too many false positives
match: '[0-9]{32}'
```

### ❌ Missing Word Boundaries
```yaml
# Matches substrings
match: 'password'
```

### ✅ Proper Implementation
```yaml
# Specific, bounded, accurate
match: '\bpassword["\s]*[:=]["\s]*[^\s]'
```

## Testing Against Real Code

```bash
# Test on open source projects
atheon --all /path/to/other-project

# Look for false positives
# Adjust pattern if needed
# Verify detection accuracy
```

## Pattern Examples

### API Keys
```yaml
# Stripe
name: stripe-api-key
match: '\bsk_(test|live)_[0-9a-z]{32}\b'

# GitHub
name: github-token
match: '\b(ghp|gho|ghu|ghs)_[A-Za-z0-9]{36}\b'

# Slack
name: slack-token
match: '\bxox[bap]-[0-9]{12}-[0-9]{12}-[0-9]{12}[A-Za-z0-9]{32}\b'
```

### Code Quality
```yaml
# TODO comments
name: todo-comment
match: '(?i)TODO[^\n]*:'

# Debug prints
name: debug-print
match: '(?i)(fmt\.Print|log\.Print|console\.log|println).*(?i)debug'

# Hardcoded URLs
name: hardcoded-url
match: '(https?://)[^\s]+["\')]'
```

### DevOps
```yaml
# Docker secrets
name: docker-secret
match: 'ENV|ARG|secret[_-]?key["\s]*[:=]'

# Kubernetes config
name: kubernetes-config
match: 'kind:[Cc]onfig[Cc]Map'
```

## Pattern Validation

### Self-Validation
```bash
# Test pattern against codebase
atheon --all . > findings.json

# Analyze results
jq '.[] | select(.pattern == "my-pattern")' findings.json

# Check for false positives
# Refine pattern if needed
```

### Community Testing
```bash
# Test on diverse projects
# Gather feedback
# Iterate on pattern
```

## Next Steps

- Test your pattern thoroughly
- Check for false positives
- Validate with real code examples
- Submit PR with clear description
- Respond to review feedback

## Pattern Maintenance

### Updating Patterns
```bash
# Edit pattern file
# Rebuild bundle
go run ./bundler

# Test updated pattern
atheon --all . | grep my-pattern-name
```

### Deprecating Patterns
```yaml
# Add metadata field when supported
enabled: false
deprecation_reason: "Replaced by better-pattern"
```

---

**Need inspiration?** Check existing patterns in `community/` directory.