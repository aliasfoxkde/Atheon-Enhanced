# security-hardening

Patterns in this category:

| Pattern | Description | Severity |
|---------|-------------|----------|
| `authentication-bypass` | Detected potential CI/security bypass: authentication bypass. Bypass | high |
| `authentication` | Detected hardcoded value: hardcoded auth credential. Hardcoded values | high |
| `cors` | Pattern: cors. Pattern in category . This is a high-severity | high |
| `csrf` | Detected missing CSRF protection: csrf. Forms should include CSRF | high |
| `debug-mode-production` | Detected debug or logging statement: debug mode production. Debug | high |
| `docker-env-secret` | Detected potential secret or credential: docker env secret. Exposing | high |
| `github-actions-secret` | Detected potential secret or credential: github actions secret exposure. | high |
| `hardcoded-password` | Detected hardcoded credential: hardcoded password. Credentials should | high |
| `hardcoded-secrets` | Detected potential secret or credential: hardcoded secrets. Exposing | critical |
| `http-without-tls` | HTTP URL detected. Ensure the endpoint uses HTTPS for encrypted transport. | medium |
| `injection` | Detected potential SQL injection risk: injection. User input should | high |
| `input-validation` | Pattern: input validation. Pattern in category . This is a high-severity | high |
| `insecure-dependencies` | Pattern: insecure dependencies. Security hardening issue that weakens | high |
| `insecure-oauth` | Pattern: insecure OAuth. Pattern in category security-hardening. Detects OAuth implementations without PKCE or using implicit flow. | high |
| `path-traversal` | Pattern: path traversal. Security hardening issue that weakens the | critical |
| `python-command-injection` | Detected potential SQL injection risk: python command injection. | critical |
| `self-signed-cert` | Pattern: self-signed certificate. Pattern in category security-hardening. Detects usage of self-signed certificates or disabled certificate verification. | medium |
| `session-fixation` | Pattern: session fixation. Security hardening issue that weakens | high |
| `session-management` | Pattern: session management. Pattern in category . This is a | high |
| `weak-cryptography` | Pattern: weak cryptography. Pattern in category . This is a | high |
| `xss` | Detected potential XSS vulnerability: xss. User input should be | high |

---
*Auto-generated from pattern YAML files*
