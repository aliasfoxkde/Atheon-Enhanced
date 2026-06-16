package patterns

import (
	"atheon/core"
	"regexp"
)

var (
	gcpPrivateKeyIDPattern = regexp.MustCompile(`"private_key_id"\s*:\s*"[0-9a-f]{40}"`)
	gcpClientEmailPattern  = regexp.MustCompile(`"client_email"\s*:\s*"[^"]+@[^"]+\.iam\.gserviceaccount\.com"`)
)

func init() {
	core.Register(&gcpPattern{
		privateKeyID: gcpPrivateKeyIDPattern,
		clientEmail:  gcpClientEmailPattern,
	})
}

type gcpPattern struct {
	privateKeyID *regexp.Regexp
	clientEmail  *regexp.Regexp
}

func (p *gcpPattern) Name() string { return "gcp-service-account-key" }
func (p *gcpPattern) Matches(line string) bool {
	return p.privateKeyID.MatchString(line) || p.clientEmail.MatchString(line)
}
