# Atheon Pattern Categories - Comprehensive Documentation

## Overview

Atheon Enhanced now supports **355+ patterns** across **24 categories** (327 regex patterns + 28 AST-based Go security patterns), providing comprehensive coverage of modern development challenges. This documentation describes each category, its patterns, and use cases.

## Pattern Categories

### 🔍 1. Accessibility Patterns (10 patterns)
**Purpose**: Ensure WCAG compliance and accessibility for users with disabilities

**Patterns Include**:
- `aria-labels`: Detect missing ARIA labels on interactive elements
- `keyboard-navigation`: Identify keyboard interaction barriers
- `focus-management`: Detect improper focus management
- `color-contrast`: Find insufficient color contrast ratios
- `semantic-html`: Identify non-semantic HTML usage
- `form-labels`: Detect missing form labels and associations
- `heading-hierarchy`: Identify improper heading structure
- `missing-alt-text`: Find images without alt attributes
- `table-headers`: Detect missing table header associations

**Use Cases**: Web application accessibility auditing, WCAG compliance checking, screen reader optimization

---

### ⚡ 2. Performance Patterns (9 patterns)
**Purpose**: Identify performance bottlenecks and optimization opportunities

**Patterns Include**:
- `n-plus-one-query`: Detect database queries inside loops
- `memory-leak-closure`: Find resource leaks and cleanup issues
- `synchronous-api`: Identify blocking operations in async contexts
- `missing-caching`: Detect missing cache strategies
- `inefficient-data-structures`: Find inefficient data structure usage
- `missing-index`: Identify missing database indexes
- `inefficient-regex`: Detect complex regex performance issues
- `blocking-main-thread`: Find thread blocking operations
- `missing-lazy-loading`: Identify missing lazy loading patterns

**Use Cases**: Application performance profiling, database optimization, code review for performance issues

---

### 🌐 3. Web Development Patterns (7 patterns)
**Purpose**: Modern web framework best practices and common issues

**Patterns Include**:
- `nextjs`: Next.js framework-specific anti-patterns
- `typescript-any`: TypeScript type safety violations
- `bundler-optimization`: Bundle size and code splitting issues
- `react-optimization`: React performance patterns
- `state-management`: State management anti-patterns
- `error-boundaries`: Missing error boundary implementation
- `form-validation`: Form validation patterns

**Use Cases**: React/Next.js development, TypeScript projects, modern web application code review

---

### 🔌 4. API Integration Patterns (8 patterns)
**Purpose**: REST/GraphQL API best practices and common issues

**Patterns Include**:
- `rest`: Poor REST API implementation patterns
- `rate-limiting`: Missing rate limit handling
- `error-handling`: API error handling issues
- `graphql`: GraphQL optimization and caching
- `authentication`: API authentication patterns
- `pagination`: Pagination implementation issues
- `versioning`: API versioning patterns
- `timeout`: Request timeout configuration

**Use Cases**: API client development, server-side API implementation, API security auditing

---

### 🔒 5. Security Hardening Patterns (8 patterns)
**Purpose**: Security vulnerabilities beyond basic secret detection

**Patterns Include**:
- `injection`: SQL/XSS/code injection vulnerabilities
- `cors`: CORS misconfigurations
- `authentication`: Authentication vulnerabilities
- `weak-cryptography`: Outdated crypto algorithms
- `session-management`: Session security issues
- `input-validation`: Missing input validation
- `csrf`: Cross-site request forgery patterns
- `xss`: Cross-site scripting vulnerabilities

**Use Cases**: Security auditing, penetration testing support, secure code review

---

### ☁️ 6. Cloud-Native Patterns (6 patterns)
**Purpose**: Modern cloud infrastructure patterns and best practices

**Patterns Include**:
- `docker`: Containerization anti-patterns
- `kubernetes`: Kubernetes deployment patterns
- `terraform`: Infrastructure as code patterns
- `serverless`: Serverless function patterns
- `health-checks`: Health check implementations
- `multi-region`: Multi-region deployment patterns

**Use Cases**: Docker/Kubernetes development, Terraform IaC, serverless architecture review

---

### 📱 7. PWA Patterns (5 patterns)
**Purpose**: Progressive Web App implementation and optimization

**Patterns Include**:
- `service-worker`: Service worker implementation patterns
- `manifest`: PWA manifest configuration
- `offline-support`: Offline functionality patterns
- `shortcuts`: App shortcuts implementation
- `caching-strategy`: Service worker caching strategies

**Use Cases**: PWA development, offline functionality implementation, service worker optimization

---

### 📊 8. Data Visualization Patterns (5 patterns)
**Purpose**: Chart and dashboard implementation best practices

**Patterns Include**:
- `chart-config`: Chart configuration issues
- `chart-types`: Appropriate chart type selection
- `chart-accessibility`: Chart accessibility compliance
- `color-schemes`: Color scheme and contrast issues
- `mobile-optimization`: Mobile chart optimization

**Use Cases**: Dashboard development, data visualization accessibility, mobile chart optimization

---

### 🤖 9. AI Detection Patterns (6 patterns)
**Purpose**: Identify AI-generated code and AI-assisted development patterns

**Patterns Include**:
- `ai-buzzwords`: AI-related buzzwords and terminology
- `ai-emoji`: Emoji usage patterns common in AI outputs
- `ai-incomplete-code`: Incomplete code patterns from AI generation
- `ai-overuse`: Excessive AI-generated code patterns
- `ai-safety-bypass`: AI safety restriction bypasses
- `ai-template`: AI template and boilerplate patterns

**Use Cases**: AI code review, detecting AI-generated content, AI safety enforcement

---

### 🏗️ 10. Code Quality Patterns (25+ patterns)
**Purpose**: Code maintenance and quality issues

**Patterns Include**:
- `console-log`: Debug logging in production code
- `debug-statement`: Debug statements left in code
- `deprecated-function`: Use of deprecated functions
- `dummy-function`: Placeholder and dummy functions
- `empty-catch-block`: Empty exception handlers
- `fixme-comment`: FIXME comments indicating issues
- `hardcoded-url`: Hardcoded URLs in code
- `placeholder-code`: Placeholder and temporary code
- `temporary-code`: Temporary development code
- `todo-comment`: TODO comments
- `todo-stub`: TODO stub implementations
- `unreachable-code`: Code that can never execute
- `auto-confirm`: Auto-confirmation commands
- `build-force`: Force build operations
- `git-clean-force`: Destructive git operations
- `git-force-push`: Force push operations
- `git-hard-reset`: Hard reset operations
- `insecure-flag`: Insecure configuration flags
- `mock-stub`: Mock and stub code
- `fake-data`: Fake data in production
- `package-manager-force`: Force package operations
- `skip-hooks`: Git hook bypassing
- `skip-tests`: Test skipping patterns
- `cruft`: Dead and unused code

**Use Cases**: Code quality auditing, technical debt identification, code review automation

---

### 🚀 11. DevOps Patterns (6 patterns)
**Purpose**: CI/CD and infrastructure automation patterns

**Patterns Include**:
- `ci-bypass`: CI/CD pipeline bypassing
- `ci-cd`: CI/CD configuration patterns
- `dockerfile`: Dockerfile best practices
- `git-hook`: Git hook configurations
- `github-workflow`: GitHub Actions workflow patterns
- `kubernetes`: Kubernetes configuration patterns

**Use Cases**: DevOps pipeline auditing, infrastructure security review, CI/CD optimization

---

### 💰 12. Finance Patterns (3 patterns)
**Purpose**: Financial services and payment processing patterns

**Patterns Include**:
- `aba-routing`: ABA routing numbers
- `iban`: International Bank Account Numbers
- `swift-bic`: SWIFT/BIC codes

**Use Cases**: Financial applications, payment processing, banking systems

---

### 🏥 13. Healthcare Patterns (7 patterns)
**Purpose**: Medical and healthcare data patterns

**Patterns Include**:
- `clinical-trial-id`: Clinical trial identifiers
- `healthcare-code`: Healthcare code systems
- `insurance-number`: Insurance numbers
- `medical-license`: Medical license numbers
- `medical-record-number`: Medical record numbers
- `patient-id`: Patient identifiers
- `prescription-number`: Prescription numbers

**Use Cases**: Healthcare applications, medical systems, HIPAA compliance

---

### 📋 14. PII Patterns (3 patterns)
**Purpose**: Personal Identifiable Information detection

**Patterns Include**:
- `creditcard`: Credit card numbers
- `phone`: Phone numbers
- `ssn`: Social Security Numbers

**Use Cases**: PII auditing, data privacy compliance, GDPR/CCPA adherence

---

### 🔑 15. Secrets Patterns (30+ patterns)
**Purpose**: API keys, credentials, and sensitive data detection

**Patterns Include**:
- AWS, Azure, GCP, GitHub, GitLab, CircleCI, Jenkins, Docker, NPM, PyPI, Stripe, Slack, Twilio, OpenAI, and many more

**Use Cases**: Secret scanning, credential auditing, security compliance

---

### 🎨 16. Framework-Specific Patterns
**Purpose**: Framework-specific anti-patterns and best practices

**Subcategories**:
- **Django**: `django-debug-print` - Debug prints in Django
- **Node.js**: `nodejs-todo-development` - Development TODOs
- **React**: `react-console-log-dev` - Console logs in development

**Use Cases**: Framework-specific code review, technology stack optimization

---

### 🌐 17. Web Security Patterns (7 patterns)
**Purpose**: Web application security vulnerabilities

**Patterns Include**:
- `js-sql-template-literal`: SQL injection in template literals
- `react-dangerously-set-inner-html`: React XSS vulnerabilities
- `django-csrf-missing`: Django CSRF protection
- `express-form-without-csrf`: Express CSRF issues
- Plus additional web security patterns

**Use Cases**: Web security auditing, OWASP Top 10 compliance, application security review

---

## Pattern Usage Examples

### Basic Pattern Scanning
```bash
# Scan with specific categories
atheon --categories=accessibility,performance ./my-project

# Enable specific patterns
atheon enable n-plus-one-query memory-leak-closure

# Disable patterns temporarily
atheon disable console-log debug-statement

# List patterns by category
atheon list categories
```

### CI/CD Integration
```bash
# Security-focused scan
atheon --categories=secrets,security-hardening,web-security --json ./src > security-report.json

# Performance-focused scan
atheon --categories=performance ./myapp > performance-issues.txt

# Accessibility audit
atheon --categories=accessibility ./website > accessibility-report.html
```

### Pattern Management
```bash
# Enable all patterns in a category
atheon --categories=accessibility list | awk '{print $1}' | xargs atheon enable

# Create custom profile
echo '{"patterns": ["n-plus-one-query", "memory-leak", "missing-caching"]}' > performance-profile.json
atheon --profile=performance-profile.json ./myproject
```

## Pattern Development Guidelines

### Creating New Patterns
1. **Identify the Need**: Based on real-world security/performance issues
2. **Define the Pattern**: Create YAML file in appropriate `community/` subdirectory
3. **Test Thoroughly**: Ensure regex works with Go's RE2 engine
4. **Document Clearly**: Include examples and remediation steps
5. **Submit for Review**: Follow community contribution process

### Pattern Format
```yaml
name: pattern-name
match: 'regex_pattern_here'
```

### Categories
- Create new category directory under `community/`
- Category name automatically extracted from directory
- Patterns inherit category from their directory location

## Pattern Statistics

- **Total Patterns**: 152
- **Categories**: 19
- **Growth**: 157% increase from original 57 patterns
- **Coverage**: Modern web development, security, performance, accessibility
- **Maintenance**: Active community contributions and updates

## Contributing Patterns

See [CONTRIBUTING.md](../.github/CONTRIBUTING.md) for detailed guidelines on pattern submission, testing, and review processes.

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [WCAG Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [Web Performance](https://web.dev/)
- [React Best Practices](https://react.dev/learn)
- [Next.js Documentation](https://nextjs.org/docs)

---

**Last Updated**: 2026-06-19
**Pattern Version**: 2.0
**Maintained By**: Atheon Enhanced Community