# Backend Integration Guide

Integrating Atheon-Enhanced with backend services (Rust, Python, Go, etc.)

## Overview

Atheon-Enhanced can be integrated with backend services in several ways:

1. **MCP Server** (Recommended) - JSON-RPC over stdio
2. **CLI Subprocess** - Spawn and communicate via stdout/stderr
3. **Library Import** - Future Go library binding

---

## Option 1: MCP Server (Recommended)

The MCP server provides a clean JSON-RPC interface over stdio.

### Installation

```bash
# Download MCP server
wget -q https://github.com/aliasfoxkde/Atheon-Enhanced/releases/latest/download/atheon-mcp
chmod +x atheon-mcp
sudo mv atheon-mcp /usr/local/bin/
```

### Rust Integration

```rust
use std::process::{Command, Stdio};
use std::io::{BufRead, BufReader, Write};

// Call Atheon-Enhanced MCP server
fn scan_via_mcp(content: &str) -> Result<Vec<Finding>, Box<dyn Error>> {
    let mut child = Command::new("atheon-mcp")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::null())
        .spawn()?;

    // Send JSON-RPC request
    let request = serde_json::json!({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "scan_string",
            "arguments": {
                "content": content,
                "categories": ["secrets", "pii"]
            }
        }
    });

    let stdin = child.stdin.as_mut().unwrap();
    stdin.write_all(format!("{}\n", request).as_bytes())?;
    drop(stdin);

    // Read response
    let reader = BufReader::new(child.stdout.unwrap());
    for line in reader.lines() {
        let response: Response = serde_json::from_str(&line?)?;
        return Ok(parse_findings(response.result));
    }

    Ok(vec![])
}
```

### Python Integration

```python
import subprocess
import json

def scan_via_mcp(content: str, categories: list[str] = None) -> list[dict]:
    if categories is None:
        categories = ["secrets", "pii"]

    proc = subprocess.Popen(
        ["atheon-mcp"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        text=True
    )

    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "scan_string",
            "arguments": {
                "content": content,
                "categories": categories
            }
        }
    }

    stdout, _ = proc.communicate(input=json.dumps(request) + "\n")
    response = json.loads(stdout)

    return parse_findings(response["result"])
```

### MCP Tools Available

| Tool | Description |
|------|-------------|
| `scan_string` | Scan a string for patterns |
| `scan_file` | Scan a file for patterns |
| `scan_dir` | Scan a directory for patterns |
| `scan_env` | Scan environment variables |
| `list_patterns` | List all patterns |
| `list_categories` | List all categories |
| `update_bundle` | Update pattern bundle |

---

## Option 2: CLI Subprocess

Spawn the CLI as a subprocess and parse output.

### Rust Integration

```rust
use std::process::Command;
use serde::Deserialize;

#[derive(Deserialize)]
struct AtheonOutput {
    findings: Vec<Finding>,
}

#[derive(Deserialize)]
struct Finding {
    pattern: String,
    file: String,
    line: u32,
    content: String,
}

fn scan_file(path: &str) -> Result<Vec<Finding>, Box<dyn Error>> {
    let output = Command::new("atheon")
        .args(["--json", "--categories=secrets,pii", path])
        .output()?;

    if !output.status.success() {
        return Ok(vec![]); // No findings or error
    }

    let result: AtheonOutput = serde_json::from_slice(&output.stdout)?;
    Ok(result.findings)
}
```

### Python Integration

```python
import subprocess
import json

def scan_file(path: str, categories: list[str] = None) -> list[dict]:
    if categories is None:
        categories = ["secrets", "pii"]

    cmd = ["atheon", "--json", f"--categories={','.join(categories)}", path]
    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.returncode != 0:
        return []

    data = json.loads(result.stdout)
    return data.get("findings", [])
```

### CLI Output Formats

**Text (default):**
```
secret-api-key  config.py:42
secret-password  user.rb:15
pii-email        users.csv:23
```

**JSON:**
```json
{
  "findings": [
    {
      "pattern": "secret-api-key",
      "file": "config.py",
      "line": 42,
      "column": 5,
      "content": "api_key = 'sk-1234567890abcdef'",
      "category": "secrets",
      "severity": "high"
    }
  ],
  "stats": {
    "files": 1,
    "bytes": 1024,
    "elapsedMs": 15
  }
}
```

**SARIF (for GitHub Security tab):**
```json
{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [{
    "results": [{
      "ruleId": "secret-api-key",
      "level": "error",
      "message": { "text": "Potential API key detected" },
      "locations": [{
        "physicalLocation": {
          "artifactLocation": { "uri": "config.py" },
          "region": { "startLine": 42 }
        }
      }]
    }]
  }]
}
```

---

## Option 3: Library (Future)

Future Go library binding via CGO or pure Go API.

```go
// Future API (not yet implemented)
import "github.com/aliasfoxkde/Atheon"

func scan() {
    engine := atheon.New()
    findings := engine.ScanString("api_key = 'sk-xxx'")
    for _, f := range findings {
        fmt.Printf("%s: %s\n", f.Pattern, f.Content)
    }
}
```

---

## Configuration for Backend Use

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ATHEON_LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `ATHEON_LOG_FORMAT` | `text` | Log format: text, json |
| `ATHEON_MCP_RATE_LIMIT` | `10` | MCP requests per second |
| `ATHEON_MCP_RATE_BURST` | `20` | MCP burst limit |
| `ATHEON_MCP_CONCURRENT_CAP` | `50` | Max concurrent MCP requests |
| `ATHEON_MCP_REQUEST_TIMEOUT` | `30s` | MCP request timeout |

### Rate Limiting

For high-throughput backends, configure rate limiting:

```bash
export ATHEON_MCP_RATE_LIMIT=100
export ATHEON_MCP_RATE_BURST=200
export ATHEON_MCP_CONCURRENT_CAP=100
```

---

## Performance Considerations

### Latency

| Method | Latency | Best For |
|--------|---------|----------|
| MCP (stdio) | ~5-10ms | Real-time scanning |
| CLI (subprocess) | ~50-100ms | Batch scanning |
| Library (future) | ~1-5ms | High-frequency scanning |

### Throughput

- MCP: ~100 req/sec per instance
- CLI: ~10 files/sec per instance
- Consider connection pooling for high throughput

### Memory

- MCP server: ~10MB baseline
- Per-request: ~1KB per scan
- Large file scanning: streaming mode recommended

---

## Error Handling

```rust
match scan_result {
    Ok(findings) => {
        for f in findings {
            println!("{}:{} - {}", f.file, f.line, f.pattern);
        }
    }
    Err(e) => {
        eprintln!("Scan error: {}", e);
        // Fallback or retry logic
    }
}
```

---

## Example: backend-fixed Integration

The `backend-fixed` project has a stub implementation at:

```
~/repos/backend-fixed/crates/backend-mcp/src/atheon.rs
```

To complete the integration, replace the stub with one of the methods above.

---

## See Also

- [MCP Integration](../integrations/Mcp.md)
- [Pattern Categories](../architecture/PATTERN_CATEGORIES.md)
- [Atheon-Enhanced Repository](https://github.com/aliasfoxkde/Atheon-Enhanced)
