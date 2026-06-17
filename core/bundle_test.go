package core_test

import (
	"atheon/core"
	"strings"
	"testing"
)

type patternCase struct {
	matches    []string
	nonMatches []string
}

func TestRegisteredPatterns(t *testing.T) {
	cases := map[string]patternCase{
		"Social Security Number": {
			matches:    []string{"ssn=123-45-6789"},
			nonMatches: []string{"ssn=123-456-789", "invoice=123-45-678"},
		},
		"aws-access-key": {
			matches:    []string{"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"},
			nonMatches: []string{"AWS_ACCESS_KEY_ID=AKIA123", "AKIAiosfodnn7example"},
		},
		"credit-card": {
			matches:    []string{"card=4242 4242 4242 4242", "amex=3782-822463-10005"},
			nonMatches: []string{"card=1234 5678 9012 3456", "order=4242"},
		},
		"gcp-api-key": {
			matches:    []string{"api_key=AIza" + strings.Repeat("a", 35)},
			nonMatches: []string{"api_key=AIza-short", "api_key=AIza" + strings.Repeat("!", 35)},
		},
		"gcp-oauth-client-id": {
			matches:    []string{"client_id=1234567890-abcdefghijklmnopqrstuvwxyz.apps.googleusercontent.com"},
			nonMatches: []string{"client_id=project.apps.googleusercontent.com", "client_id=1234567890-.apps.googleusercontent.com"},
		},
		"gcp-oauth-client-secret": {
			matches:    []string{"client_secret=GOCSPX-" + strings.Repeat("a", 28)},
			nonMatches: []string{"client_secret=GOCSPX-short", "client_secret=GOOGLE-" + strings.Repeat("a", 28)},
		},
		"gcp-service-account-email": {
			matches:    []string{"svc=my-service@project.iam.gserviceaccount.com"},
			nonMatches: []string{"svc=my-service@example.com", "svc=@project.iam.gserviceaccount.com"},
		},
		"gcp-service-account-key": {
			matches: []string{
				`"private_key_id": "` + strings.Repeat("a", 40) + `"`,
				`"client_email": "svc@project.iam.gserviceaccount.com"`,
			},
			nonMatches: []string{`"private_key_id": "short"`, `"client_email": "svc@example.com"`},
		},
		"github-pat": {
			matches:    []string{"token=ghp_" + strings.Repeat("a", 36)},
			nonMatches: []string{"token=ghp_short", "token=github_pat_" + strings.Repeat("a", 36)},
		},
		"openai-api-key": {
			matches:    []string{"OPENAI_API_KEY=sk-" + strings.Repeat("a", 20)},
			nonMatches: []string{"OPENAI_API_KEY=sk-short", "OPENAI_API_KEY=pk-" + strings.Repeat("a", 20)},
		},
		"phone-number": {
			matches:    []string{"phone=(555) 123-4567", "phone=+1 555 123 4567"},
			nonMatches: []string{"version=555-123", "ticket=555-123-456"},
		},
		"slack-bot-token": {
			matches:    []string{"SLACK_BOT_TOKEN=xoxb-12345678901-12345678901-" + strings.Repeat("a", 24)},
			nonMatches: []string{"SLACK_BOT_TOKEN=xoxb-short", "SLACK_BOT_TOKEN=xoxa-12345678901-12345678901-" + strings.Repeat("a", 24)},
		},
		"stripe-secret-key": {
			matches:    []string{"STRIPE_SECRET_KEY=sk_live_" + strings.Repeat("a", 24)},
			nonMatches: []string{"STRIPE_SECRET_KEY=sk_test_" + strings.Repeat("a", 24), "STRIPE_SECRET_KEY=sk_live_short"},
		},
		"twilio-account-sid": {
			matches:    []string{"TWILIO_ACCOUNT_SID=AC" + strings.Repeat("a", 32)},
			nonMatches: []string{"TWILIO_ACCOUNT_SID=ACshort", "TWILIO_ACCOUNT_SID=SK" + strings.Repeat("a", 32)},
		},
	}

	registered := map[string]core.Pattern{}
	for _, p := range core.All() {
		registered[p.Name()] = p
	}

	for name, tc := range cases {
		p, ok := registered[name]
		if !ok {
			t.Fatalf("pattern %q not registered", name)
		}
		t.Run(name, func(t *testing.T) {
			for _, line := range tc.matches {
				if !p.Matches(line) {
					t.Errorf("expected %q to match %q", name, line)
				}
			}
			for _, line := range tc.nonMatches {
				if p.Matches(line) {
					t.Errorf("expected %q not to match %q", name, line)
				}
			}
		})
	}

	for name := range cases {
		if _, ok := registered[name]; !ok {
			t.Fatalf("test case %q has no registered pattern", name)
		}
	}
}
