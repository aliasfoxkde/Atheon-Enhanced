package core

import (
	"context"
	"testing"
)

// TestScanLinesWithEntropyFiltering tests that patterns with minEntropy > 0
// properly filter out low-entropy content.
func TestScanLinesWithEntropyFiltering(t *testing.T) {
	// Load a minimal bundle with a pattern that matches any line with 20+ chars
	// and has minEntropy set to filter low-entropy content
	testBundle := `
[
  {
    "name": "test-entropy",
    "category": "test",
    "match": ".{20,}",
    "enabled": true,
    "severity": "medium",
    "minEntropy": 4.0
  }
]
`
	bundleData := gzipBundle([]byte(testBundle))

	// Save and restore global state
	origBundle := embeddedBundle
	defer func() {
		_ = loadBundle(origBundle)
		SetActiveCategories(nil)
	}()

	// Load our test bundle
	if err := loadBundle(bundleData); err != nil {
		t.Fatal(err)
	}
	SetActiveCategories([]string{"test"})

	// High entropy content (mixed chars) should match
	highEntropy := "ABCDEFGHIJKLMNOPQRSTUV" // 22 uppercase letters, high entropy ~4.4
	findings := ScanString(context.Background(), highEntropy, "test")
	if len(findings) == 0 {
		t.Error("expected high-entropy content to match")
	}

	// Low entropy content (repeated pattern) should NOT match (entropy < 4.0)
	lowEntropy := "AAAAAAAAAAAAAAAAAAAA" // 20 As - very low entropy ~0
	findings = ScanString(context.Background(), lowEntropy, "test")
	if len(findings) != 0 {
		t.Error("expected low-entropy content to be filtered out")
	}
}
