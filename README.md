# Atheon

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Patterns](https://img.shields.io/badge/patterns-6-blueviolet)

> **Status:** Feature complete. The engine is done. What grows from here are patterns and bug fixes — nothing more, nothing less.

---

**One tool. All patterns. Any input.**

Atheon is a pattern matching engine. You tell it what to look for. You point it at anything. It finds every match and tells you exactly where.

---

## Why this matters

Data ends up where it shouldn't. A hardcoded credential in a config file. A production secret in a log. A sensitive string committed into a repository by accident and now permanently in git history. These mistakes happen constantly — across every team, every stack, every domain.

The problem isn't that people are careless. The problem is there's no systematic way to catch what you can't see.

Atheon is that system. A pattern matching engine you define, run anywhere, and trust completely — because you wrote the rules.

---

## What pattern matching means

A pattern is a rule: "if a line looks like this, flag it." That rule can be a regex, a keyword check, a structural test — anything that returns true or false. Every pattern has a name. Every match tells you the file, the line, and what was found.

The engine itself is deliberately minimal. It doesn't know what a secret is, what compliance means, or what matters to your organization. You do. So you define it, and the engine enforces it — over files, directories, environment variables, or any stream of text piped through it.

Pattern matching is useful in any domain where text contains something that shouldn't be there, or something that must be there. Security. Compliance. Legal. Operations. Healthcare. Finance. If you can describe the rule, Atheon can run it.

---

## The scenario that makes this real

A developer wraps up a sprint and pushes a configuration file. Inside it, buried in a comment from a debugging session three weeks ago, is a production API key. The commit goes through. The pipeline passes. The key is now in git history, in the build artifact, and eventually in a production image. Someone rotates it two months later after a billing alert.

Atheon, wired into a pre-push hook:

```
$ atheon ./

[api-key] config/app.yaml:47  →  # debug key: sk-prod-a8f3c...
```

Exit code `1`. The push never happens. The key never leaves the machine.

That's it. That's the product.

---

## Install

Download the binary for your platform from [Releases](https://github.com/HoraDomu/Atheon/releases/latest). No install, no runtime, no dependencies. Drop it in your PATH and run it.

**Or build from source:**

```
go build -o atheon .
```

Cross-compile for any platform:

```
GOOS=windows GOARCH=amd64 go build -o atheon.exe .
GOOS=linux   GOARCH=amd64 go build -o atheon-linux .
GOOS=darwin  GOARCH=arm64 go build -o atheon-macos .
```

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

Drop the file in `patterns/`, rebuild. It appears in `atheon list` automatically.

The same two methods work for anything — credentials, PII, internal token formats, compliance markers, prohibited strings. If you can describe the rule, this is all the code it takes.

---

## Contributing

The engine is done. What grows from here are patterns — and that's where you come in.

Every pattern in Atheon is one file and two methods. That's it. If you've ever spotted a credential format that should be caught, a token structure that's leaking in the wild, a compliance rule your team has to enforce — you already have everything you need to add it. The barrier is as low as it gets. The impact is permanent.

This is the right project to contribute to if you want something that ships, stays, and runs in the real world. No framework to learn. No architecture to understand. One file, two methods, done.

---

### The complete workflow — from idea to merged

**Step 1 — Define what you're detecting**

Answer two questions before writing anything:
- What does it look like? A fixed prefix, a structural shape, a known format.
- Why does it matter? A leaked credential, a compliance violation, a prohibited string.

If you can describe the rule in one sentence, you're ready. If you can't, the pattern isn't ready yet.

**Step 2 — Check it doesn't already exist**

```
atheon list
```

If it's already there, you're done. If not, continue.

**Step 3 — Create the pattern file**

Create a new `.go` file in `patterns/`. Name it after what you're detecting.

```
patterns/my-pattern.go
```

Every pattern follows the same structure — no exceptions:

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

- `Name()` — what shows up in findings and `atheon list`. Use lowercase with hyphens. Be specific: `stripe-live-key`, not `stripe`.
- `Matches(line string) bool` — return `true` if this line should be flagged. One line at a time.

The match logic doesn't have to be a regex. It can be anything that returns a bool:

```go
func (p *myPattern) Matches(line string) bool {
    return strings.Contains(line, "INTERNAL_ONLY") && strings.Contains(line, "=")
}
```

If you can describe the rule, you can write the function.

**Step 4 — Build**

```
go build -o atheon .
```

**Step 5 — Confirm it loaded**

```
atheon list
```

Your pattern's `Name()` should appear. If it doesn't, check that `init()` is present and the file is in `patterns/`.

**Step 6 — Test it**

Create a file with lines that should match and lines that shouldn't:

```
# test.txt
this should match: MY_TOKEN=abc123xyz
this should not:   MY_TOKEN=
this should not:   # MY_TOKEN=abc123xyz
this should match: export MY_TOKEN=xyz987abc
```

Run it:

```
atheon --file test.txt
```

Every expected match appears. No unexpected matches. If anything is off, adjust the logic and test again. A pattern that cries wolf is worse than no pattern at all.

**Step 7 — Submit**

Three ways:

**Open a GitHub issue** at [github.com/HoraDomu/Atheon/issues](https://github.com/HoraDomu/Atheon/issues) — include what it detects, why it matters, the logic you used, and example lines that match and don't.

**Open a pull request** — fork the repo, add your file to `patterns/`, confirm `atheon list` shows it, confirm your test file passes, and open the PR. Description should include what it detects, why it matters, and your test cases.

**Email directly** — [dommcpro@gmail.com](mailto:dommcpro@gmail.com) — same information, attach or paste the `.go` file.

---

### What happens after

A maintainer reviews for correctness, false positive rate, name clarity, and whether it overlaps with an existing pattern. If changes are needed, you'll hear back. If it's good, it gets merged and ships in the next release.

---

### The whole thing, at a glance

```
1. Describe the rule in one sentence
2. atheon list           →  confirm it doesn't exist
3. patterns/my-pattern.go  →  Name() + Matches()
4. go build -o atheon .
5. atheon list           →  confirm it loaded
6. atheon --file test.txt  →  correct matches, no false positives
7. Issue / PR / email
8. Review → merge → ships
```

Same process. Every time. Pattern six or pattern ten thousand.

---

## Releases

Every 3 patterns merged = a new version release. No exceptions, no skipping. The pattern count drives the version. When the badge hits the next multiple of 3, a release goes out.

To build binaries for all platforms at once:

```
GOOS=windows GOARCH=amd64 go build -o dist/atheon-windows-amd64.exe .
GOOS=linux   GOARCH=amd64 go build -o dist/atheon-linux-amd64 .
GOOS=darwin  GOARCH=amd64 go build -o dist/atheon-macos-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o dist/atheon-macos-arm64 .
```

Put the outputs in a `dist/` folder, attach them to the GitHub release, and update the patterns badge at the top of this file.

---

## License

MIT with Additional Terms — Copyright © 2026 Dominick Yanez

You are free to fork this repository, clone the source, study it, modify it for personal or internal use, and contribute patterns or bug fixes back. That's encouraged.

What you may not do:
- Ship this software, or any derivative of it, as your own standalone product, tool, or service under a different name or brand
- Remove or obscure the author's name or copyright notice from any copy, fork, or derivative work

Violating these terms is copyright infringement. Legal action will follow.

For permissions beyond this scope: [dommcpro@gmail.com](mailto:dommcpro@gmail.com)

See the full [LICENSE](./LICENSE) file for complete terms.
