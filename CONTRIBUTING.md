# Contributing

Atheon grows through patterns. Every pattern is one YAML file — no Go required, no engine changes, no recompile needed.

## Adding a pattern

**1. Check it doesn't already exist**

```sh
atheon list
```

**2. Pick the right category**

```sh
atheon list categories
```

Category is just the folder name under `community/`. If yours doesn't fit an existing one, create a new folder — the bundler picks it up automatically.

**3. Create the YAML file**

Drop a `.yaml` file into the appropriate `community/<category>/` folder:

```yaml
name: my-service-api-key
match: '\bmsvc_[A-Za-z0-9]{32}\b'
```

- `name` — lowercase hyphenated, specific: `stripe-live-key` not `stripe`
- `match` — a valid RE2 regex. Use single quotes so backslashes don't need escaping.

**4. Rebuild the bundle**

```sh
go run ./bundler
```

This reads all YAML files in `community/` and writes `core/patterns.bundle`. Commit both the YAML file and the updated bundle.

**5. Confirm it loaded**

```sh
atheon list
```

Your pattern name should appear.

**6. Test it**

Open `core/bundle_test.go` and add a case for your pattern under the `cases` map:

```go
"my-service-api-key": {
    matches:    []string{"token=msvc_" + strings.Repeat("a", 32)},
    nonMatches: []string{"token=msvc_short", "token=other_" + strings.Repeat("a", 32)},
},
```

Then run:

```sh
go test ./...
```

The suite enforces that every registered pattern has a test case — it will fail without one.

**7. Verify manually**

```sh
atheon --file <path-to-sample>
```

Every expected match should appear. No unexpected matches.

**8. Submit**

Open a pull request. Include what the pattern detects, why it matters, and the test cases you used. Maintainers review for correctness, false positive rate, name clarity, and overlap with existing patterns.

---

## Adding a new category

Create the folder and drop a YAML file in it:

```
community/
  my-new-category/
    first-pattern.yaml
```

The bundler derives the category name from the folder. No other changes needed.

---

## Engine changes

Bug fixes and quality-of-life improvements to `core/` are welcome. Open an issue first to describe what you're changing and why — the engine is intentionally minimal and changes are reviewed carefully.
