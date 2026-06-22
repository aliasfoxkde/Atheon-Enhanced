# Contributing to Atheon

Atheon grows through patterns. Every pattern is one YAML file — no Go required, no engine changes, fast to review, and immediately useful to every user once merged.

---

## 🎯 Which Project to Contribute To?

### **Official HoraDomu/Atheon**
- **Best for**: Stable patterns, bug fixes, documentation
- **Process**: Standard PR review and testing
- **Impact**: Immediate benefit to all users
- **Repository**: [https://github.com/HoraDomu/Atheon](https://github.com/HoraDomu/Atheon)

### **Enhanced aliasfoxkde/Atheon**
- **Best for**: Experimental patterns, performance features, CI/CD improvements
- **Process**: Enhanced testing, validation across multiple Go versions
- **Impact**: Testing ground for future innovations
- **Repository**: [https://github.com/aliasfoxkde/Atheon-Enhanced](https://github.com/aliasfoxkde/Atheon-Enhanced)
- **Contributors**: [View Contributors Graph](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors)

---

## 👥 Contributors & Recognition

### **Official Project Contributors**
- [CONTRIBUTORS.md](CONTRIBUTORS.md) - Contributors to the official project
- All contributions are permanently credited

### **Enhanced Fork Contributors**
- [Live Contributors Graph](https://github.com/aliasfoxkde/Atheon-Enhanced/graphs/contributors) - Real-time contributor visualization
- [Project Pulse](https://github.com/aliasfoxkde/Atheon-Enhanced/pulse) - Recent activity and engagement
- Your name appears permanently in the contributor history

**Both projects value every contribution.** Whether you contribute to the official project or this enhanced testing fork, your work benefits the entire community.

---

## Development Setup

After cloning, install the pre-commit hook so every commit is automatically formatted and tested:

```sh
git config core.hooksPath hooks
```

The hook:
- Auto-formats Go files with `gofmt` and re-stages them
- Runs `go vet ./...`
- Runs the full test suite (`go test ./... -p 1`) with coverage
- Rebuilds `core/patterns.bundle` if any `community/**/*.yaml` files changed

---

## Adding a YAML pattern

**1. Check it doesn't already exist**

```sh
atheon list
atheon list categories
atheon list --category=secrets
```

**2. Create the YAML file**

Drop a `.yaml` file into the appropriate `community/<category>/` folder. The folder name becomes the pattern's category — if yours doesn't fit an existing one, create a new folder.

```yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
enabled: false   # optional — omit to default to true
```

Fields:

- `name` — lowercase hyphenated, specific: `stripe-live-key` not `stripe`
- `match` — a valid RE2 regex. Use single quotes so backslashes don't need escaping.
- `enabled` — optional, defaults to `true`. Set to `false` to ship the pattern disabled-by-default. Useful for high-false-positive patterns that users can opt into with `atheon enable <name>` if needed.

**3. Rebuild the bundle**

```sh
go run ./bundler
```

Commit both the YAML file and the updated `core/patterns.bundle`.

**4. Add a test case**

Open `core/bundle_test.go` and add an entry under the `cases` map:

```go
"my-service-api-key": {
    matches:    []string{"token=msvc_" + strings.Repeat("a", 32)},
    nonMatches: []string{"token=msvc_short", "token=other_" + strings.Repeat("a", 32)},
},
```

**5. Run tests and verify manually**

```sh
go test ./... -p 1
atheon --file <path-to-sample>
```

Every expected match should appear. No unexpected matches.

**6. Submit**

Open a pull request. Include what the pattern detects, why it matters, and the test cases you used. Maintainers review for correctness, false positive rate, name clarity, and overlap with existing patterns.

---

## Go contributions

Any Go code contributed to this project must be clean and idiomatic. That means:

- Standard Go naming conventions — exported names are `PascalCase`, unexported are `camelCase`, acronyms follow Go style (`url` not `URL` for unexported, `URL` not `Url` for exported)
- No unnecessary abstraction — if three lines do the job, don't wrap them in a helper
- Error handling is explicit — no swallowed errors without a clear reason
- No comments that explain what the code does — only add a comment when the **why** is non-obvious
- `go fmt` and `go vet` must pass before submitting
- If you add a dependency, justify it — the engine has very few and we'd like to keep it that way

The engine is intentionally minimal and stable. Contributions that touch `core/` without a clear bug fix or performance reason will not be merged. If you're unsure whether a change is in scope, open an issue first.

If you're unsure whether something is idiomatic, the [Effective Go](https://go.dev/doc/effective_go) guide and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) are the references we follow. Code that is hard to read will be sent back regardless of whether it works.
