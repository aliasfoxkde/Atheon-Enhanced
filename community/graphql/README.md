# GraphQL Security Patterns

Detects GraphQL API security misconfigurations.

## Patterns

- `graphql-introspection-enabled`: Detects introspection enabled in production
- `graphql-query-depth-unlimited`: Detects missing query depth limiting
- `graphql-field-cost-undefined`: Detects missing field cost analysis
- `graphql-debug-mode`: Detects debug mode enabled in production

## References

- [GraphQL Security Best Practices](https://graphql.org/learn/security/)
- [CWE-200: Exposure of Sensitive Information](https://cwe.mitre.org/data/definitions/200.html)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
