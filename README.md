# Atheon

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)

**Scan your code, configs, and environment for leaked secrets before they become a problem.**

Atheon is a command-line tool that reads files, directories, environment variables, or piped input and flags any lines that match known secret patterns — API keys, tokens, credentials. It ships with patterns for the most common providers out of the box, and you can add your own in minutes.

---

## Download

Grab the binary for your platform from [Releases](https://github.com/HoraDomu/Atheon/releases/latest) — no install, no runtime required.

---

## What Atheon detects (built-in)

| Pattern | Example match |
|---|---|
| `aws-access-key` | `AKIA...` / `ASIA...` |
| `github-pat` | `ghp_...` |
| `openai-api-key` | `sk-...` |
| `slack-token` | `xox[bprs]-...` |
| `stripe-key` | `sk_live_...` / `pk_live_...` |
| `twilio-token` | Twilio account SIDs and auth tokens |

Run `atheon list` to see every loaded pattern.

---

## Example scenario

You're about to push a feature branch. You want to make sure no credentials slipped into the diff.

```
$ atheon ./src

[aws-access-key] config/deploy.yaml:14  →  AWS_KEY=AKIAIOSFODNN7EXAMPLE
[openai-api-key] .env.local:3           →  OPENAI_API_KEY=sk-proj-abc123...
```

Two findings, two lines, no guesswork. Fix them before the push. Exit code `1` means something was found — wire that into your CI pipeline and the check runs automatically on every commit.

---

## Usage

```
atheon <path>          scan a directory
atheon --file <path>   scan a single file
atheon --env           scan environment variables
atheon list            list loaded patterns
```

Pipe support:

```
cat file.txt | atheon -
```

Exit code `0` = clean. Exit code `1` = findings. CI-friendly by default.

---

## Adding a pattern

One file. Two methods.

```go
package patterns

import (
    "atheon/core"
    "regexp"
)

func init() {
    core.Register(&myPattern{re: regexp.MustCompile(`your-regex-here`)})
}

type myPattern struct{ re *regexp.Regexp }

func (p *myPattern) Name() string             { return "my-pattern-name" }
func (p *myPattern) Matches(line string) bool { return p.re.MatchString(line) }
```

Drop the file in `patterns/`, rebuild. It appears in `atheon list` automatically. If you have an internal token format, a company-specific credential, or a compliance rule — this is all you need.

---

## Build

```
go build -o atheon .
```

Cross-compile:

```
GOOS=windows GOARCH=amd64 go build -o atheon.exe .
GOOS=linux   GOARCH=amd64 go build -o atheon-linux .
GOOS=darwin  GOARCH=arm64 go build -o atheon-macos .
```

---

## License

MIT — Copyright © 2026 Dominick Yanez
