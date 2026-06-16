package patterns

import (
    "atheon/core"
    "regexp"
)

func init() { core.Register(&myPattern{re: regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)}) }
type myPattern struct{ re *regexp.Regexp }
func (p *myPattern) Name() string             { return "Social Security Number" }
func (p *myPattern) Matches(line string) bool { return p.re.MatchString(line) }
