# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest `*-enhanced` release | ✅ |
| Older releases | ❌ — please upgrade |

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Use one of these private channels:

1. **GitHub Private Advisory (preferred):** Go to [Security → Advisories](https://github.com/aliasfoxkde/Atheon-Enhanced/security/advisories) → "Report a vulnerability". This creates a private draft visible only to maintainers.

2. **Email:** Contact the maintainer directly via the email listed on their [GitHub profile](https://github.com/aliasfoxkde).

### What to include

- Description of the vulnerability and its potential impact
- Steps to reproduce or a minimal proof-of-concept
- Affected versions (if known)
- Any suggested mitigations

### Response timeline

- **Acknowledgement:** within 72 hours
- **Initial assessment:** within 7 days
- **Fix / advisory publication:** coordinated with the reporter

We follow responsible disclosure — we ask that you give us reasonable time to release a fix before public disclosure.

## Scope

This project is a **pattern-matching CLI and library**. The most relevant security concerns are:

- **False negatives** in security patterns (secrets/PII missed by the scanner) — please report these as vulnerabilities, not bugs
- **Regex denial-of-service (ReDoS)** in bundled patterns
- **Malicious pattern injection** via community YAML files

Out of scope: theoretical issues without a realistic attack path, issues in third-party dependencies (report those upstream).
