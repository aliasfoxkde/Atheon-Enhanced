package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Tests for Null Dereference Detection

func TestDetectNullDereference(t *testing.T) {
	// Test nil pointer dereference detection with explicit nil assignment
	code := `package main

type User struct {
	ID int
}

func main() {
	var user *User
	user = nil
	print(user.ID)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nil_deref.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range findings {
		if f.Rule == "null-dereference" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find null-dereference pattern")
	}
}

func TestDetectNullDereference_Safe(t *testing.T) {
	// Test safe code - nil check before use
	code := `package main

func getUser(id int) *User {
	if id <= 0 {
		return nil
	}
	return &User{ID: id}
}

func main() {
	user := getUser(1)
	if user != nil {
		print(user.ID)
	}
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_nil.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "null-dereference" {
			t.Errorf("safe code should not trigger null-dereference: %s", f.Message)
		}
	}
}

func TestDetectNullDereference_Slice(t *testing.T) {
	// Test nil slice indexing
	code := `package main

func main() {
	var items []int
	items = nil
	print(items[0])
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nil_slice.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range findings {
		if f.Rule == "null-dereference" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find null-dereference pattern for nil slice")
	}
}

func TestDetectNullDereference_MethodCall(t *testing.T) {
	// Test method call on potentially nil pointer
	code := `package main

type User struct {
	Name string
}

func (u *User) GetName() string {
	return u.Name
}

func main() {
	var u *User
	u = nil
	u.GetName()
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nil_method.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range findings {
		if f.Rule == "null-dereference" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find null-dereference pattern for method call on nil")
	}
}

// Tests for Dead Assignment Detection

func TestDetectDeadAssignment(t *testing.T) {
	// Test dead assignment detection
	code := `package main

func process() {
	x := 10
	y := 20
	print(y)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "dead_assign.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range findings {
		if f.Rule == "dead-assignment" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find dead-assignment pattern")
	}
}

func TestDetectDeadAssignment_Safe(t *testing.T) {
	// Test that used variables don't trigger dead assignment
	code := `package main

func process() {
	x := 10
	y := x + 20
	print(y)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "safe_assign.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "dead-assignment" {
			t.Errorf("used variable should not trigger dead-assignment: %s", f.Message)
		}
	}
}

func TestDetectDeadAssignment_FunctionResult(t *testing.T) {
	// Test dead assignment with function result
	code := `package main

func getValue() int {
	return 42
}

func process() {
	result := getValue()
	_ = result
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "func_result.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}

func TestDetectDeadAssignment_Parameter(t *testing.T) {
	// Test that parameters are not flagged as dead assignment
	code := `package main

func process(x int) {
	print(x)
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "param.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range findings {
		if f.Rule == "dead-assignment" {
			t.Errorf("parameter should not trigger dead-assignment: %s", f.Message)
		}
	}
}

func TestIsNilLit(t *testing.T) {
	// Test isNilLit helper function
	code := `package main

func check() {
	var x *int = nil
	_ = x
}
`
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nil_lit.go")
	if err := os.WriteFile(tmpFile, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := ScanFileAST(tmpFile, builtinASTPatterns)
	if err != nil {
		t.Fatal(err)
	}
	_ = findings
}
