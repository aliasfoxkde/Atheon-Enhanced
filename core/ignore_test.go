package core

import (
	"os"
	"testing"
)

func TestIgnorePatternToRegexp(t *testing.T) {
	tests := []struct {
		pattern   string
		shouldErr bool
		matches   []string
		notMatch  []string
	}{
		{"foo", false, []string{"foo", "bar/foo"}, []string{"fooo", "bar/fooo"}},
		{"*.go", false, []string{"foo.go", "bar/test.go"}, []string{"foo.ts", "bar/foo.go/baz"}},
		{"**/foo", false, []string{"foo", "bar/foo", "bar/baz/foo", "/foo"}, []string{}},
		{"dist/**", false, []string{"dist/foo", "dist/a/b/c"}, []string{"dist"}},
		{"dist/", false, []string{"dist", "dist/foo"}, []string{"distfile"}},
	}

	for _, tc := range tests {
		re, err := ignorePatternToRegexp(tc.pattern)
		if tc.shouldErr {
			if err == nil {
				t.Errorf("expected error for pattern %q", tc.pattern)
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for pattern %q: %v", tc.pattern, err)
			continue
		}
		for _, m := range tc.matches {
			if !re.MatchString(m) {
				t.Errorf("pattern %q should match %q", tc.pattern, m)
			}
		}
		for _, m := range tc.notMatch {
			if re.MatchString(m) {
				t.Errorf("pattern %q should not match %q", tc.pattern, m)
			}
		}
	}
}

func TestCompileIgnoreFile(t *testing.T) {
	content := `# comment
*.log
!important.log
dist/
`
	tmpfile, err := os.CreateTemp("", "ignore-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	m, err := compileIgnoreFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("compileIgnoreFile failed: %v", err)
	}
	if len(m.rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(m.rules))
	}
}

func TestCompileIgnoreFile_Empty(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "ignore-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	// Empty file
	tmpfile.Close()

	m, err := compileIgnoreFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("compileIgnoreFile failed: %v", err)
	}
	if len(m.rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(m.rules))
	}
}

func TestIgnoreMatcher_MatchesPath(t *testing.T) {
	content := `*.log
!important.log
`
	tmpfile, err := os.CreateTemp("", "ignore-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	m, err := compileIgnoreFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("compileIgnoreFile failed: %v", err)
	}

	if !m.matchesPath("test.log") {
		t.Error("*.log should match test.log")
	}
	if m.matchesPath("important.log") {
		t.Error("important.log should NOT match (negated by !)")
	}
	if m.matchesPath("test.txt") {
		t.Error("*.log should not match test.txt")
	}
}
