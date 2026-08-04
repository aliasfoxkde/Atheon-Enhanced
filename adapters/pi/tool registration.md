# Pi CLI Integration

## Overview

The Atheon PI adapter registers an `atheon` tool with the Pi coding agent, allowing Pi to scan code for secrets, PII, and code quality issues directly within the Pi workflow.

## Architecture

```
Pi CLI → Tool Registry → Atheon Adapter → Core Runtime → Findings
```

## Tool Registration

The `atheon` tool is registered with Pi and provides:

- `scan <path>` - Scan a directory or file for issues
- `list categories` - List available pattern categories
- `list patterns` - List all patterns
- `version` - Show Atheon version

## Usage

Within Pi, users can invoke:

```
@atheon scan <path>
```

Example responses:

### Scan Request
```json
{
  "command": "scan",
  "paths": ["./src"]
}
```

### Scan Response
```json
{
  "findings": [
    {
      "pattern": "aws-access-key",
      "file": "./src/auth.go",
      "line": 42,
      "column": 5,
      "severity": "high",
      "category": "secrets",
      "message": "Detected potential API key"
    }
  ],
  "summary": {
    "total": 1,
    "critical": 0,
    "high": 1,
    "medium": 0,
    "low": 0
  },
  "risk_score": {
    "score": 30,
    "level": "low"
  }
}
```

## Building

```bash
go build -o atheon-adapter ./adapters/pi
```

## Testing

```bash
go test ./adapters/pi/... -v
```
