package patterns

import (
	"atheon/core"
	"regexp"
)

func init() {
	core.Register(&phonePattern{re: regexp.MustCompile(`(?:^|[^0-9A-Za-z-])(?:\+?\d{1,3}[-.\s]?)?(?:\(?\d{3}\)?[-.\s]?)\d{3}[-.\s]?\d{4}(?:$|[^0-9A-Za-z-])`)})
}

type phonePattern struct{ re *regexp.Regexp }

func (p *phonePattern) Name() string             { return "phone-number" }
func (p *phonePattern) Matches(line string) bool { return p.re.MatchString(line) }
