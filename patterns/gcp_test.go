package patterns

import "testing"

func TestGCPPatternMatchesServiceAccountKeyFields(t *testing.T) {
	pattern := &gcpPattern{
		privateKeyID: gcpPrivateKeyIDPattern,
		clientEmail:  gcpClientEmailPattern,
	}

	matches := []string{
		`  "private_key_id": "0123456789abcdef0123456789abcdef01234567",`,
		`  "client_email": "atheon@example-project.iam.gserviceaccount.com",`,
	}

	for _, line := range matches {
		if !pattern.Matches(line) {
			t.Fatalf("expected %q to match", line)
		}
	}
}

func TestGCPPatternIgnoresUnrelatedFields(t *testing.T) {
	pattern := &gcpPattern{
		privateKeyID: gcpPrivateKeyIDPattern,
		clientEmail:  gcpClientEmailPattern,
	}

	nonMatches := []string{
		`  "private_key_id": "not-a-service-account-key",`,
		`  "client_email": "user@example.com",`,
	}

	for _, line := range nonMatches {
		if pattern.Matches(line) {
			t.Fatalf("expected %q not to match", line)
		}
	}
}
