# kubernetes

Patterns in this category:

| Pattern | Description | Severity |
|---------|-------------|----------|
| `hostnetwork` | Pattern: host network. Pattern in category kubernetes. Detects containers using host network namespace. | high |
| `hostpid` | Pattern: host PID namespace. Pattern in category kubernetes. Detects containers sharing host PID namespace. | high |
| `latest-tag` | Pattern: latest image tag. Pattern in category kubernetes. Detects use of non-reproducible latest image tag in container definitions. | medium |
| `no-resource-limits` | Detects Kubernetes containers missing CPU/memory resource limits. Add resources.requests and resources.limits to container specs. | medium |
| `privileged-container` | Pattern: privileged container. Pattern in category kubernetes. Detects containers running in privileged mode with full host access. | critical |
| `secrets-in-manifest` | Pattern: secrets in K8s manifest. Pattern in category kubernetes. Detects secrets defined directly in Kubernetes manifests rather than using external secrets management. | high |

---
*Auto-generated from pattern YAML files*
