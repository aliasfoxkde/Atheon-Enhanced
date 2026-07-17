# Atheon Architecture

## Overview

Atheon is designed as a minimal, efficient pattern matching engine with a pluggable pattern system. The architecture prioritizes performance, simplicity, and extensibility.

## Core Components

### Pattern Engine (core/)

**bundle.go**
- Pattern loading and registration
- Category-based filtering
- Active scanner management
- Bundle compilation and decompression

**pattern.go**
- Pattern interface definition
- Pattern registry
- Pattern discovery and sorting

**runner.go**
- File system scanning
- Pattern matching engine
- Gitignore compliance
- Environment scanning

**finding.go**
- Finding structure definition
- Statistics tracking

### Pattern Bundler (bundler/)

**main.go**
- YAML pattern loading
- Bundle compilation
- JSON generation
- Gzip compression

### CLI Interface (main.go)

- Command-line parsing
- User interface
- Pattern management commands
- Output formatting

### MCP Server (cmd/mcp/)

**main.go**
- Model Context Protocol implementation
- Tool registration
- Request/response handling

## Data Flow

### Pattern Loading
1. Load embedded bundle or local bundle
2. Parse JSON and decompress gzip
3. Register all patterns
4. Build category scanners
5. Enable active patterns

### Scanning Process
1. Parse ignore files (gitignore, atheonignore)
2. Walk directory tree
3. Read file contents
4. Apply active patterns
5. Generate findings

### Pattern Matching
- Category-based pre-filtering
- Combined regex optimization
- Line-by-line scanning
- Context-aware matching

## Pattern Format

Patterns are defined as YAML files in the community/ directory:

```yaml
name: pattern-name
category: category-name
match: 'regex-pattern'
# enabled: false (optional)
```

## Category System

Patterns are organized into categories:
- **secrets**: Credentials and sensitive data
- **pii**: Personal identifiable information
- **code-quality**: Code smell detection
- **healthcare**: Medical identifiers
- **finance**: Financial patterns

## Performance Optimizations

1. **Regex Combination**: Patterns within categories are combined into single regex
2. **Early Filtering**: Gitignore and binary file filtering
3. **Concurrent Scanning**: Parallel file processing
4. **Lazy Loading**: Patterns loaded on-demand
5. **Caching**: Compiled regex patterns cached
6. **Entropy Caching**: Shannon entropy values cached with LRU eviction (1024 entry max)

## Advanced Features

### Entropy Calculation
- High-entropy string detection for cryptographic keys and secrets
- Configurable entropy threshold (default: 4.5 bits per byte)
- Caching layer reduces repeated calculations by ~100x

### Risk Scoring
- Each finding scored 0-100 based on severity and confidence
- Category-specific multipliers applied
- Baseline comparison for project-wide risk assessment

### Baseline Comparison
- Track finding counts over time
- Compare current scan against stored baseline
- Report new, resolved, and persistent findings

### Pattern Categories (36 total)
- **secrets**: AWS/GCP/Azure credentials, API keys, tokens
- **pii**: SSN, credit cards, phone numbers, emails
- **code-quality**: Code smells, anti-patterns
- **healthcare**: Clinical trial IDs, medical record numbers
- **finance**: ABA routing, credit card patterns
- **ai-detection**: AI-generated content markers, prompt injection
- **devops**: CI bypass, secrets in logs
- **kubernetes**: Security contexts, network policies, pod specs
- **terraform**: AWS/Azure/GCP misconfigurations
- **cloud-native**: ECS, EKS, Lambda security patterns
- **frameworks**: Django, Express, Flask, Laravel, Rails, Spring, Angular, Vue, React, Node.js, Go, Rust
- **supply-chain**: Typosquatting, malicious packages
- **container**: Dockerfile security, Docker Compose
- **graphql**: Introspection, query complexity
- **cloudformation**: S3, IAM, EC2 misconfigs
- **arm**: Azure resource manager templates

## Extension Points

### Adding Patterns
1. Create YAML file in community/<category>/
2. Run bundler to compile bundle
3. Tests automatically discover new patterns

### Adding Categories
1. Create new directory in community/
2. Add patterns to category
3. Category automatically discovered

### Custom Scanners
1. Implement Pattern interface
2. Register with Register()
3. Available in All() results

## Build Process

1. **Pattern Compilation**: YAML → JSON → Gzip
2. **Binary Embedding**: Bundle embedded in binary
3. **Multi-platform Builds**: Cross-compilation for all platforms (amd64, arm64)
4. **Package Generation**: Homebrew, Scoop, Docker Hub, GitHub Releases
5. **CI/CD Pipeline**: GoReleaser with CodeQL, golangci-lint, govulncheck

## Testing Strategy

- Unit tests for each component
- Integration tests for scanning
- Pattern validation tests (TestPatternDetection, TestCategoryCoverage)
- False positive guard tests
- Severity propagation tests
- Platform compatibility tests
- Performance benchmarks (entropy caching, bundle loading)

## Dependencies

**Go Standard Library:**
- regexp: Pattern matching
- compress/gzip: Bundle compression
- encoding/json: Pattern serialization
- path/filepath: Cross-platform path manipulation
- os: File system operations

**External Dependencies:**
- github.com/sabhiram/go-gitignore: Gitignore compliance
- github.com/goccy/go-yaml: YAML parsing (migrated from deprecated gopkg.in/yaml.v3)

## Security Considerations

- Pattern validation prevents regex attacks
- Binary file detection prevents resource exhaustion
- Gitignore compliance prevents scanning sensitive files
- No network access during scanning
- Sandboxed pattern evaluation

## Performance Characteristics

- **Memory**: < 50MB for typical usage
- **Speed**: < 100ms for 1000-line files
- **Scalability**: Linear scaling for file size
- **Concurrency**: Configurable worker pool (default: 256)
- **Patterns**: 400+ enabled patterns across 36 categories
- **Coverage**: 81%+ test coverage