package patterns

import (
	"atheon/core"
	"regexp"
)

func init() {
	core.Register(&creditCardPattern{re: regexp.MustCompile(`\b(?:4[0-9]{3}[- ]?(?:[0-9]{4}[- ]?){2}[0-9]{4}|5[1-5][0-9]{2}[- ]?(?:[0-9]{4}[- ]?){2}[0-9]{4}|3[47][0-9]{2}[- ]?[0-9]{6}[- ]?[0-9]{5}|6(?:011|5[0-9]{2})[- ]?(?:[0-9]{4}[- ]?){2}[0-9]{4})\b`)})
}

type creditCardPattern struct{ re *regexp.Regexp }

func (p *creditCardPattern) Name() string             { return "credit-card" }
func (p *creditCardPattern) Matches(line string) bool { return p.re.MatchString(line) }
