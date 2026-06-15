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
## Atheons Mission
Atheon isn't trying to be the next big GitLeaks or anything of that nature, really. It's not competing to be this super giant; it's trying to be a platform.

Imagine this: a CLI tool that a team of devs can use, right? And for some reason, in their code, they're working with sensitive data. They add their own pattern to Atheon, then push it to the codebase. Now Atheon's new release and their local version have this pattern registered.

They find the sensitive info they were looking for, and now it's not committed.

But imagine that now Atheon has this pattern, and another group of devs decides to use it because they're in a similar situation. The pattern is already registered. Now they run it on their codebase and boom—the pattern is found, day is saved.

That's the idea. That's Atheon: a community-driven CLI tool that lets me, you, and Billy Bob add our own patterns that others can use as well.

Refer to the contributing guide to add your own pattern, but do keep in mind that the creator, HoraDomu, made the architecture sound for simple, easy additions.

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

## Contributing

 See [CONTRIBUTING.md](CONTRIBUTING.md).
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
