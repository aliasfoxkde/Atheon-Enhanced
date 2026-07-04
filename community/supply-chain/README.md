# Supply Chain Security Patterns

Detects package management vulnerabilities, dependency confusion attacks, typosquatting, and other supply chain risks.

## Pattern Fields

This category uses the enhanced schema with:
- `cwe`: CWE identifier for standardized vulnerability categorization
- `remediation`: Suggested fix or mitigation
- `tags`: Classification tags for filtering

## Patterns

- `npm-internal-package-mismatch`: Detects npm packages referencing internal/private registries without explicit scope
- `npm-install-script`: Detects npm packages with potentially malicious install scripts
- `pypi-internal-package-mismatch`: Detects pip requirements referencing internal PyPI servers
- `pip-extra-index-url`: Detects pip installs from untrusted extra index URLs
- `typosquat-common-packages`: Detects typosquatting targets (react, vue, lodash, etc.)

## References

- [CWE-1395: Dependency/Supply Chain](https://cwe.mitre.org/data/definitions/1395.html)
- [OWASP Dependency Confusion](https://owasp.org/www-project-top-ten/2017/A8)
- [npm Typosquatting Research](https://arxiv.org/abs/2012.1429)
