<!--
Thanks for contributing to Atheon-Enhanced! Please fill out this template so
reviewers can quickly understand your change. Delete sections that don't apply.

Title format: type(scope): short description
  type   = feat | fix | docs | test | refactor | build | ci | chore
  scope  = core | cli | mcp | patterns | ci | docs | security | deps

Branch prefix: feature/ | fix/ | docs/ | test/ | refactor/
  (never commit directly to main)
-->

## Summary

<!-- One or two sentences on what this PR does and why. -->

-

## Root cause (fix PRs only)

<!-- What was actually wrong? Link the issue. -->

## Changes

<!-- Bullet list of the meaningful changes (not file-by-file noise). -->

-

## Validation

<!-- Commands you ran locally that prove it works. -->

- [ ] `go vet ./...` — OK
- [ ] `go test ./... -p 1` — OK (note `-p 1` is mandatory, see ADR 0006)
- [ ] `go build ./...` — OK
- [ ] `golangci-lint run .` — 0 issues
- [ ] `./scripts/pattern-count.sh --total` — source count matches bundle (if touching patterns)

## Quality gates (reviewer-facing)

<!-- Don't fill this in yourself — it lists what CI/Coderabbit/etc will run. -->

- [ ] All required status checks pass (see branch protection)
- [ ] `CODEOWNERS` review (only @aliasfoxkde at present)
- [ ] No new patterns trigger self-scan-secrets
- [ ] `CHANGELOG.md` updated under `[Unreleased]` for user-visible changes
- [ ] Pattern count docs updated if pattern count changed (`scripts/pattern-count.sh`)
- [ ] ADR added/updated under `docs/architecture/decisions/` for any
      architectural decision

## Breaking changes

<!-- If this breaks existing behaviour, call it out explicitly and link the
     CHANGELOG entry. Otherwise delete this section. -->

None.

## References

<!-- Link related issues, ADRs, or upstream PRs. -->

- Fixes #
- Relates to #
