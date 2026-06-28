# devops

Patterns in this category:

| Pattern | Description | Severity |
|---------|-------------|----------|
| `action-with-SSH-key` | Pattern: action-with-SSH-key — Detects actions using SSH keys for git operations. SSH keys in CI/CD can be exfiltrated if the runner is compromised. | high |
| `ci-bypass` | Detected potential CI/security bypass: ci bypass. Bypass mechanisms | medium |
| `ci-cd` | Pattern: ci cd. Pattern in category . This is a medium-severity | medium |
| `ci-missing-required` | Pattern: CI missing required checks. Pattern in category devops. Detects CI configurations with missing or disabled required status checks. | medium |
| `ci-skip-bypass` | Detected potential CI/security bypass: ci skip bypass. Bypass mechanisms | medium |
| `concurrent-workflow-cancellation` | Pattern: concurrent-workflow-cancellation — Detects workflows with concurrency set but missing cancel-in-progress: true. Without this, multiple runs of the same workflow may execute simultaneously wasting resources. | low |
| `deprecated-action-version` | Pattern: deprecated-action-version — Detects use of deprecated actions/checkout version syntax (v1 or v2). Use actions/checkout@v4 or later for latest features and security fixes. | medium |
| `dockerfile` | Pattern: dockerfile. Pattern in category . This is a medium-severity | medium |
| `dockerignore-missing` | Docker build command detected. Ensure --pull is used and a .dockerignore file excludes secrets from the build context. | low |
| `env-file-committed` | Pattern: env file committed. Pattern in category devops. Detects .env files or environment variables checked into source control. | high |
| `env-secrets-in-plaintext` | Pattern: env-secrets-in-plaintext — Detects environment variables containing secrets being printed or logged. Secret values should never appear in workflow logs. | high |
| `git-hook` | Pattern: git hook. Pattern in category . This is a medium-severity | medium |
| `github-actions-secret` | Detected potential secret or credential: github actions hardcoded | medium |
| `github-token-wildcard` | Pattern: github-token-wildcard — Detects GITHUB_TOKEN with overly broad wildcard (*) permissions. Follows principle of least privilege; specify exact permissions needed. | high |
| `github-workflow` | Pattern: github workflow. Pattern in category . This is a medium-severity | medium |
| `jenkins-api-token` | Detected potential API key or token: jenkins api token. Exposing | medium |
| `kubernetes` | Pattern: kubernetes ci usage. Pattern in category . This is | medium |
| `missing-timeout` | Pattern: missing-timeout — Detects workflows using workflow_dispatch or workflow_call without timeout-minutes set. Workflows should have a timeout to prevent runaway jobs. | medium |
| `no-continue-on-error` | Pattern: no-continue-on-error — Detects explicit continue-on-error: false which is the default behavior. This is redundant noise. Consider removing or using proper error handling. | low |
| `ubuntu-latest-unsigned` | Pattern: ubuntu-latest-unsigned — Detects use of ubuntu-latest runner without additional verification. ubuntu-latest is a shared runner that could have been modified. Consider using specific version or self-hosted runners. | medium |
| `unpinned-action` | Pattern: unpinned-action — Detects GitHub Actions used without SHA pin, using @latest, @master, or @main. This is a critical security risk as the action could be modified by an attacker to execute arbitrary code. | critical |
| `workflow-dispatch-secret` | Pattern: workflow-dispatch-secret — Detects secrets passed to workflow_dispatch inputs. Secrets should not be passed as workflow inputs as they may be exposed in logs. | high |

---
*Auto-generated from pattern YAML files*
