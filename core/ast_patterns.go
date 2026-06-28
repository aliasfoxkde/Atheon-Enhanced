package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
)

// ASTFinding represents a finding from AST-based analysis.
type ASTFinding struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Severity string
}

// ASTPattern defines an AST-based pattern using Go AST traversal.
type ASTPattern struct {
	Name        string
	Description string
	Severity    string
	Func        func(fset *token.FileSet, file *ast.File) []ASTFinding
}

// builtinASTPatterns contains built-in AST patterns.
var builtinASTPatterns = []ASTPattern{
	{
		Name:        "unsafe-deserialization",
		Description: "Uses encoding.BinaryUnmarshaler or encoding.TextUnmarshaler with user input",
		Severity:    "high",
		Func:        detectUnsafeDeserialization,
	},
	{
		Name:        "sql-injection",
		Description: "Potential SQL injection - string concatenation in query",
		Severity:    "high",
		Func:        detectSQLInjection,
	},
	{
		Name:        "command-injection",
		Description: "exec.Command with string concatenation",
		Severity:    "high",
		Func:        detectCommandInjection,
	},
	{
		Name:        "path-traversal",
		Description: "os.Open or ioutil with user input in path",
		Severity:    "high",
		Func:        detectPathTraversal,
	},
	{
		Name:        "hardcoded-credentials",
		Description: "Assignment to credential variable without env var",
		Severity:    "high",
		Func:        detectHardcodedCredentials,
	},
	{
		Name:        "error-not-handled",
		Description: "Function returns error but caller doesn't check it",
		Severity:    "medium",
		Func:        detectUnhandledError,
	},
}

// ScanFileAST performs AST-based pattern scanning on a single Go file.
func ScanFileAST(path string, patterns []ASTPattern) ([]ASTFinding, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	var findings []ASTFinding
	for _, p := range patterns {
		pFindings := p.Func(fset, file)
		for i := range pFindings {
			pFindings[i].File = path
			pFindings[i].Rule = p.Name
		}
		findings = append(findings, pFindings...)
	}

	return findings, nil
}

// ScanDirAST scans all Go files in a directory with AST patterns.
func ScanDirAST(dir string, patterns []ASTPattern) ([]ASTFinding, error) {
	var allFindings []ASTFinding

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := dir + "/" + entry.Name()
		findings, err := ScanFileAST(path, patterns)
		if err != nil {
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	return allFindings, nil
}

func detectSQLInjection(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isQueryMethod(call) {
			for _, arg := range call.Args {
				if containsStringConcat(arg) {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(arg.Pos()).Line,
						Message:  "Potential SQL injection: string concatenation in query",
						Severity: "high",
					})
				}
			}
		}
		return true
	})

	return findings
}

func detectCommandInjection(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isExecCommand(call) {
			for _, arg := range call.Args {
				if containsStringConcat(arg) || containsUserInput(arg) {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(arg.Pos()).Line,
						Message:  "Potential command injection: user input in exec.Command",
						Severity: "high",
					})
				}
			}
		}
		return true
	})

	return findings
}

func detectPathTraversal(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isFileOperation(call) {
			for _, arg := range call.Args {
				if containsUserInput(arg) {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(arg.Pos()).Line,
						Message:  "Potential path traversal: user input in file operation",
						Severity: "high",
					})
				}
			}
		}
		return true
	})

	return findings
}

func detectUnsafeDeserialization(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isUnmarshalCall(call) {
			for _, arg := range call.Args {
				if containsUserInput(arg) {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(arg.Pos()).Line,
						Message:  "Potential unsafe deserialization: unmarshal with user input",
						Severity: "high",
					})
				}
			}
		}
		return true
	})

	return findings
}

func detectHardcodedCredentials(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding
	credPattern := regexp.MustCompile(`(?i)(password|secret|apikey|token|private)[a-z0-9]*`)

	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for i, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				if credPattern.MatchString(ident.Name) {
					if i < len(assign.Rhs) {
						rhs := assign.Rhs[i]
						if isStringLiteral(rhs) && !containsEnvVar(rhs) {
							findings = append(findings, ASTFinding{
								Line:     fset.Position(assign.Pos()).Line,
								Message:  fmt.Sprintf("Hardcoded credential: %s", ident.Name),
								Severity: "high",
							})
						}
					}
				}
			}
		}
		return true
	})

	return findings
}

func detectUnhandledError(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		exprStmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}

		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isErrorReturningCall(call) {
			findings = append(findings, ASTFinding{
				Line:     fset.Position(exprStmt.Pos()).Line,
				Message:  "Error return value not checked",
				Severity: "medium",
			})
		}
		return true
	})

	return findings
}

func isQueryMethod(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		return name == "Query" || name == "Exec" || name == "QueryRow" ||
			name == "Execute" || name == "rawQuery"
	}
	// Also check for direct function calls like rawQuery(...) without receiver
	if ident, ok := call.Fun.(*ast.Ident); ok {
		name := ident.Name
		return name == "Query" || name == "Exec" || name == "QueryRow" ||
			name == "Execute" || name == "rawQuery"
	}
	return false
}

func isExecCommand(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "exec" && sel.Sel.Name == "Command"
		}
	}
	return false
}

func isFileOperation(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		return name == "Open" || name == "ReadFile" || name == "WriteFile" ||
			name == "Create" || name == "Stat"
	}
	return false
}

func isUnmarshalCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		return name == "Unmarshal" || name == "Decode" || name == "NewDecoder"
	}
	return false
}

func isErrorReturningCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		name := sel.Sel.Name
		return name == "Read" || name == "Write" || name == "Close" ||
			name == "Scan" || name == "Next"
	}
	return false
}

func containsStringConcat(n ast.Node) bool {
	var found bool
	ast.Inspect(n, func(node ast.Node) bool {
		if bin, ok := node.(*ast.BinaryExpr); ok {
			if bin.Op == '+' {
				// Check if either side is or contains a string literal anywhere in subtree
				if hasStringLiteral(bin.X) || hasStringLiteral(bin.Y) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

// hasStringLiteral returns true if n is or contains a string literal anywhere in its subtree.
func hasStringLiteral(n ast.Node) bool {
	var found bool
	ast.Inspect(n, func(node ast.Node) bool {
		if isStringType(node) {
			found = true
			return false
		}
		return true
	})
	return found
}

func containsUserInput(n ast.Node) bool {
	var found bool
	ast.Inspect(n, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			name := ident.Name
			if name == "req" || name == "request" || name == "body" ||
				name == "input" || name == "params" || name == "query" ||
				name == "form" || name == "ctx" {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func isStringType(n ast.Node) bool {
	_, ok := n.(*ast.BasicLit)
	return ok
}

func isStringLiteral(n ast.Node) bool {
	if lit, ok := n.(*ast.BasicLit); ok {
		return lit.Kind == token.STRING
	}
	return false
}

func containsEnvVar(n ast.Node) bool {
	if call, ok := n.(*ast.CallExpr); ok {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name == "os" && sel.Sel.Name == "Getenv"
			}
		}
	}
	return false
}
