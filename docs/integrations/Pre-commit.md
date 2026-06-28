# Pre-Commit Hook Integration

Atheon ships Git hooks that run quality checks before every commit and push. Run the installer once after cloning:

```bash
./scripts/install-hooks.sh
```

That's it. The script wires Git to use `scripts/hooks/` and optionally installs helper tools.

---

## What the hooks do

### `pre-commit` (runs on `git commit`)

Triggered only when Go files or community pattern YAML files are staged — a clean commit with only docs changes exits immediately with no delay.

| Change type | Checks run |
|-------------|-----------|
| Go files | `gofmt` (auto-fix and re-stage), `go vet`, tests for changed packages |
| `core/` touched | Full test suite + 70% coverage gate |
| Community YAML | `go run ./bundler` to regenerate `core/patterns.bundle` and re-stage it |

**Auto-fix behavior:** `gofmt` and `goimports` run automatically and re-stage fixed files. If your editor already formats on save, you won't notice them. If it doesn't, the commit still succeeds with auto-formatted code.

### `pre-push` (runs on `git push`)

Runs the full test suite with race detection. Slower but comprehensive — catches data races that wouldn't show up in unit tests.

---

## Installation details

```bash
# What install-hooks.sh does:
git config core.hooksPath scripts/hooks    # redirect Git to repo hooks
pre-commit install                          # optional: YAML/whitespace checks
```

The hooks live at `scripts/hooks/pre-commit` and `scripts/hooks/pre-push`. They are shell scripts — no external framework required beyond Go.

**Optional tools** (installed by the script if Go is available):

| Tool | Purpose |
|------|---------|
| `goimports` | Organizes imports as part of auto-format |
| `staticcheck` | Static analysis beyond `go vet` |

---

## Skip a hook once

```bash
git commit --no-verify   # skip pre-commit
git push --no-verify     # skip pre-push
```

Use sparingly — the hooks are the first line of defense against CI failures.

---

## Troubleshooting

**`go: command not found`**

The pre-commit hook requires Go in `$PATH`. The hook checks several common paths:
```
/usr/local/go/bin
/c/Program Files/Go/bin
$HOME/go/bin
$HOME/.local/go/bin
```

If none of these work, add Go to your shell profile: `export PATH="$PATH:/usr/local/go/bin"`.

**Hook runs but tests fail on commit**

The test failure output is shown in the terminal. Fix the failing test before committing, or use `--no-verify` if the failure is a known pre-existing issue (and create a tracking issue for it).

**Bundle rebuild fails after pattern changes**

```bash
go run ./bundler    # run manually to see the error
```

Usually caused by invalid YAML syntax in a community pattern file. Check the output for the offending file path.

**Hooks stopped running**

Check that `core.hooksPath` is still set:
```bash
git config core.hooksPath
# should print: scripts/hooks
```

Re-run `./scripts/install-hooks.sh` if the value is missing or wrong.

---

## Using pre-commit framework (optional)

If you have [`pre-commit`](https://pre-commit.com) installed, the installer activates it automatically. It adds YAML syntax validation and trailing-whitespace checks on top of the Git hooks.

```bash
pip install pre-commit   # or: brew install pre-commit
./scripts/install-hooks.sh
```

There is no `.pre-commit-config.yaml` in this repo yet — the framework is wired but uses default hooks only. To add specific checks, create a `.pre-commit-config.yaml` in the repo root.

---

## Atheon as a pre-commit hook (self-scan)

You can run Atheon itself as a pre-commit check to scan staged files for secrets before they leave your machine. To add this, edit `scripts/hooks/pre-commit` and append:

```bash
# Scan staged files for secrets before they leave your machine:
STAGED_FILES=$(git diff --cached --name-only)
if [ -n "$STAGED_FILES" ]; then
    ./atheon --categories=secrets $STAGED_FILES 2>/dev/null
fi
```

> **Note**: This is proposed configuration — the self-scan snippet above is not currently in `scripts/hooks/pre-commit`. It must be added manually.

The self-scan CI job (`.github/workflows/security.yml`, `self-scan-secrets` job) already does this on every push to main and every PR. The local hook adds an earlier catch point before the code leaves your machine.
