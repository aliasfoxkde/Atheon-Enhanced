# Container Security Patterns

Detects Dockerfile and container orchestration misconfigurations.

## Patterns

- `dockerfile-privileged-mode`: Detects privileged containers
- `dockerfile-exposed-socket`: Detects Docker socket mounts
- `dockerfile-running-as-root`: Detects containers running as root
- `dockerfile-cap-add-all`: Detects excessive Linux capabilities

## References

- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [CWE-250: Execution with Unnecessary Privileges](https://cwe.mitre.org/data/definitions/250.html)
- [NIST Container Security Guide](https://csrc.nist.gov/publications/detail/sp/800-190/final)
