package core_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliasfoxkde/Atheon/core"
)

func Example_core_ScanString() {
	content := `aws_key = "AKIAIOSFODNN7EXAMPLE"
github_token = "ghp_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
another_aws = "AKIAZZZZZZZZZZZZZZZZ"
`
	findings := core.ScanString(context.Background(), content, "config.txt")
	awsCount := 0
	for _, f := range findings {
		if f.Pattern == "aws-access-key" {
			awsCount++
		}
	}
	fmt.Println(awsCount)
	// Output: 2
}

func Example_core_ScanFile() {
	tmp := filepath.Join(os.TempDir(), "example_scan_file_apidoc.txt")
	_ = os.Remove(tmp)
	if err := os.WriteFile(tmp, []byte(`aws_key = AKIAIOSFODNN7EXAMPLE
`), 0o600); err != nil {
		fmt.Println("error")
		return
	}
	defer os.Remove(tmp)

	findings, _, _ := core.ScanFile(context.Background(), tmp)
	fmt.Println(len(findings))
	// Output: 1
}

func Example_core_ScanEnv() {
	findings := core.ScanEnv(context.Background())
	fmt.Printf("%T", findings)
	_ = strings.TrimSpace // keep import used
	// Output: []core.Finding
}

func Example_core_Categories() {
	cats := core.Categories()
	if len(cats) == 0 {
		fmt.Println("no categories")
		return
	}
	fmt.Println("at least one category")
	// Output: at least one category
}

func Example_core_EnablePattern() {
	// Examples run in the same process as tests, so any pattern-state
	// mutation leaks into subsequent tests. Restore the enabled state on
	// exit (and pick a non-existent name so even the disable call is
	// side-effect-free).
	ok := core.DisablePattern("credit-card-not-registered")
	defer core.EnablePattern("credit-card")
	fmt.Printf("%T", ok)
	// Output: bool
}
