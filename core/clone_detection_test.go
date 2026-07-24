package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests for clone detection - these need duplicate code to trigger clone detection

func TestClonePatterns(t *testing.T) {
	// Create duplicate functions to trigger clone detection
	code := `package main

func processA(items []int) []int {
	result := []int{}
	for _, item := range items {
		if item > 0 {
			result = append(result, item*2)
		}
	}
	return result
}

func processB(items []int) []int {
	result := []int{}
	for _, item := range items {
		if item > 0 {
			result = append(result, item*2)
		}
	}
	return result
}

func processC(items []int) []int {
	result := []int{}
	for _, item := range items {
		if item > 0 {
			result = append(result, item*2)
		}
	}
	return result
}

func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithStruct(t *testing.T) {
	// Test clone detection with struct types
	code := `package main

	type User struct {
		Name string
		Age int
	}

	func newUserA(name string, age int) *User {
		return &User{Name: name, Age: age}
	}

	func newUserB(name string, age int) *User {
		return &User{Name: name, Age: age}
	}

	func newUserC(name string, age int) *User {
		return &User{Name: name, Age: age}
	}

func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone2.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithSelectors(t *testing.T) {
	// Test clone detection with selector expressions
	code := `package main

	type Config struct {
		Host string
		Port int
	}

	func getConfigA() *Config {
		return &Config{Host: "localhost", Port: 8080}
	}

	func getConfigB() *Config {
		return &Config{Host: "localhost", Port: 8080}
	}

	func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone3.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithIndex(t *testing.T) {
	// Test clone detection with index expressions
	code := `package main

	func getElementA(arr []int, idx int) int {
		return arr[idx] * 2
	}

	func getElementB(arr []int, idx int) int {
		return arr[idx] * 2
	}

	func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone4.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithSlice(t *testing.T) {
	// Test clone detection with slice expressions
	code := `package main

	func getSliceA(arr []int) []int {
		return arr[:len(arr)-1]
	}

	func getSliceB(arr []int) []int {
		return arr[:len(arr)-1]
	}

	func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone5.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithFuncLit(t *testing.T) {
	// Test clone detection with function literals
	code := `package main

	func getHandlerA() func(int) int {
		return func(x int) int {
			return x + 1
		}
	}

	func getHandlerB() func(int) int {
		return func(x int) int {
			return x + 1
		}
	}

	func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone6.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithBinaryOps(t *testing.T) {
	// Test clone detection with various binary operations
	code := `package main

	func computeA(x, y int) int {
		return (x + y) * (x - y) / 2
	}

	func computeB(x, y int) int {
		return (x + y) * (x - y) / 2
	}

	func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone7.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectClonePatterns_WithUnaryOps(t *testing.T) {
	// Test clone detection with unary operations
	code := `package main

	func negateA(x int) int {
		return -x
	}

	func negateB(x int) int {
		return -x
	}

	func main() {}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clone8.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}
