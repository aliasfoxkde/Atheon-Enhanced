# Contributing

Atheon grows through patterns. Every pattern is one file with two methods, so keep contributions small, focused, and easy to review.

## Pattern workflow

1. **Define what you are detecting**
   - What does it look like: a fixed prefix, structural shape, or known format?
   - Why does it matter: leaked credential, compliance violation, or prohibited string?

2. **Check it does not already exist**
   ```sh
   atheon list
   ```

3. **Create the pattern file**
   Add a new `.go` file in `patterns/`, named after what it detects.
   ```go
   package patterns

   import (
       "atheon/core"
       "regexp"
   )

   func init() { core.Register(&myPattern{re: regexp.MustCompile(`your-regex-here`)}) }
   type myPattern struct{ re *regexp.Regexp }
   func (p *myPattern) Name() string             { return "my-pattern-name" }
   func (p *myPattern) Matches(line string) bool { return p.re.MatchString(line) }
   ```
   Use lowercase hyphenated names, and be specific: `stripe-live-key`, not `stripe`.

4. **Build and confirm it loaded**
   ```sh
   go build -o atheon . || go build .
   atheon list
   ```

5. **Test the pattern**
   Create sample lines that should match and should not match, then run:
   ```sh
   atheon "scan file path" 
   ```
   Every expected match should appear, with no unexpected matches.

6. **Submit the contribution**
   Open a pull request with what the pattern detects, why it matters, and the test cases you used.

Maintainers review for correctness, false positive rate, name clarity, and overlap with existing patterns.
Also PLEASE PLEASE PLEASE dont vibe code I want this working, if you're reading this I know you're smart you can do this! Only use it for maybe writing like explaing what was fixed.
