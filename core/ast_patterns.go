package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"regexp"
	"strings"
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
	{
		Name:        "dynamic-getattr",
		Description: "Dynamic attribute access via getattr with non-literal name",
		Severity:    "low",
		Func:        detectDynamicGetattr,
	},
	{
		Name:        "reflective-getattr-sink",
		Description: "Reflective getattr with dangerous literal sink name (e.g., getattr(os, 'system'))",
		Severity:    "high",
		Func:        detectReflectiveGetattrSink,
	},
	{
		Name:        "dangerous-execution-chain",
		Description: "Dangerous execution chain - exec/eval/compile with dynamic source",
		Severity:    "critical",
		Func:        detectDangerousExecutionChain,
	},
	{
		Name:        "code-clone",
		Description: "Detected duplicate or near-duplicate code (clone detection)",
		Severity:    "medium",
		Func:        detectClones,
	},
	{
		Name:        "high-cyclomatic-complexity",
		Description: "Function has cyclomatic complexity exceeding threshold (10)",
		Severity:    "medium",
		Func:        detectHighCyclomaticComplexity,
	},
	{
		Name:        "high-cognitive-complexity",
		Description: "Function has high cognitive complexity (nested structures)",
		Severity:    "medium",
		Func:        detectHighCognitiveComplexity,
	},
	{
		Name:        "long-function",
		Description: "Function exceeds recommended line count (50 lines)",
		Severity:    "low",
		Func:        detectLongFunction,
	},
	{
		Name:        "high-halstead-difficulty",
		Description: "Function has high Halstead difficulty (confusing code)",
		Severity:    "medium",
		Func:        detectHighHalsteadDifficulty,
	},
	{
		Name:        "semantic-redundancy",
		Description: "Detected semantically redundant code (unused functions, dead code, redundant operations)",
		Severity:    "low",
		Func:        detectSemanticRedundancy,
	},
	{
		Name:        "inconsistent-boolean-naming",
		Description: "Detected mixed use of different boolean naming conventions (Yes/No vs True/False)",
		Severity:    "low",
		Func:        detectInconsistentBooleanNaming,
	},
	{
		Name:        "wrapper-chain-too-deep",
		Description: "Detected wrapper chain exceeding depth of 2",
		Severity:    "medium",
		Func:        detectDeepWrapperChain,
	},
	{
		Name:        "duplicate-public-api",
		Description: "Detected duplicate public API signatures across modules",
		Severity:    "medium",
		Func:        detectDuplicatePublicAPI,
	},
	{
		Name:        "multiple-responsibilities",
		Description: "Module or type appears to have multiple responsibilities",
		Severity:    "medium",
		Func:        detectMultipleResponsibilities,
	},
	{
		Name:        "abstraction-depth-exceeded",
		Description: "Abstraction depth exceeds recommended maximum of 3",
		Severity:    "medium",
		Func:        detectExcessiveAbstractionDepth,
	},
	{
		Name:        "poorly-named-identifier",
		Description: "Identifier name suggests poor naming convention (real_, helper_impl, util2, etc.)",
		Severity:    "low",
		Func:        detectPoorlyNamedIdentifier,
	},
	{
		Name:        "duplicate-configuration",
		Description: "Detected duplicate configuration values across the codebase",
		Severity:    "low",
		Func:        detectDuplicateConfiguration,
	},
	{
		Name:        "missing-return",
		Description: "Function may return without a value on some code paths (conditional return but no final return)",
		Severity:    "high",
		Func:        detectMissingReturn,
	},
	{
		Name:        "duplicate-condition",
		Description: "Detected duplicate condition in if/elif chain",
		Severity:    "medium",
		Func:        detectDuplicateCondition,
	},
	{
		Name:        "impossible-branch",
		Description: "Detected impossible branch - condition that can never be true given prior checks",
		Severity:    "high",
		Func:        detectImpossibleBranch,
	},
	{
		Name:        "iterator-modification",
		Description: "Iterator target modified during iteration - concurrent modification bug",
		Severity:    "high",
		Func:        detectIteratorModification,
	},
	{
		Name:        "lock-not-released",
		Description: "Lock acquired but not released on all code paths",
		Severity:    "high",
		Func:        detectLockNotReleased,
	},
	{
		Name:        "resource-leak",
		Description: "Resource opened but not closed on all code paths",
		Severity:    "high",
		Func:        detectResourceLeak,
	},
	{
		Name:        "transaction-not-ended",
		Description: "Transaction begun but not committed or rolled back on all code paths",
		Severity:    "high",
		Func:        detectTransactionBug,
	},
	{
		Name:        "null-dereference",
		Description: "Potential nil pointer dereference - object used after possible nil assignment",
		Severity:    "high",
		Func:        detectNullDereference,
	},
	{
		Name:        "dead-assignment",
		Description: "Variable assigned but never used - dead store",
		Severity:    "low",
		Func:        detectDeadAssignment,
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

// Dangerous names for getattr() sink detection (AST9)
var dangerousGetattrNames = map[string]bool{
	"exec":       true,
	"eval":       true,
	"system":     true,
	"popen":      true,
	"__import__": true,
}

// detectDynamicGetattr detects AST7: getattr() with non-literal attribute name
func detectDynamicGetattr(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is a getattr call
		if !isGetattrCall(call) {
			return true
		}

		// getattr requires at least 2 args: object and attribute name
		if len(call.Args) < 2 {
			return true
		}

		// Check if second argument is NOT a literal (meaning dynamic attribute name)
		secondArg := call.Args[1]
		if isStringLiteral(secondArg) {
			// This is AST9 (reflective getattr sink), not AST7
			return true
		}

		findings = append(findings, ASTFinding{
			Line:     fset.Position(call.Pos()).Line,
			Message:  "Dynamic attribute access via getattr() with non-literal name",
			Severity: "low",
		})
		return true
	})

	return findings
}

// detectReflectiveGetattrSink detects AST9: getattr with dangerous literal sink name
func detectReflectiveGetattrSink(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !isGetattrCall(call) {
			return true
		}

		if len(call.Args) < 2 {
			return true
		}

		secondArg := call.Args[1]
		if !isStringLiteral(secondArg) {
			return true
		}

		// Extract the literal string value
		if lit, ok := secondArg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			// Remove quotes
			val := lit.Value
			if len(val) >= 2 {
				val = val[1 : len(val)-1]
			}
			if dangerousGetattrNames[val] {
				findings = append(findings, ASTFinding{
					Line:     fset.Position(call.Pos()).Line,
					Message:  fmt.Sprintf("Reflective dangerous call via getattr() with literal sink name: %s", val),
					Severity: "high",
				})
			}
		}
		return true
	})

	return findings
}

// detectDangerousExecutionChain detects AST8: exec/eval/compile with dangerous source
func detectDangerousExecutionChain(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if this is exec, eval, or compile
		name := getCallName(call)
		if name != "exec" && name != "eval" && name != "compile" {
			return true
		}

		// Check if arguments contain dangerous sources
		for _, arg := range call.Args {
			if containsDangerousSource(arg) {
				findings = append(findings, ASTFinding{
					Line:     fset.Position(call.Pos()).Line,
					Message:  fmt.Sprintf("Dangerous execution chain: %s() with dynamic source", name),
					Severity: "critical",
				})
				break
			}
		}
		return true
	})

	return findings
}

// containsDangerousSource checks if an AST node contains a dangerous source
// that could be used in an execution chain (e.g., __import__, subprocess, etc.)
func containsDangerousSource(n ast.Node) bool {
	var found bool
	ast.Inspect(n, func(node ast.Node) bool {
		if found {
			return false
		}

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		name := getCallName(call)
		if name == "" {
			return true
		}

		// Check for dangerous sources
		if name == "__import__" || name == "importlib.__import__" {
			found = true
			return false
		}

		// Check for remote fetch patterns (simplified)
		if name == "requests.get" || name == "requests.post" ||
			name == "urllib.request.urlopen" || name == "httpx.get" ||
			name == "urllib.request.urlretrieve" {
			found = true
			return false
		}

		// Check for encoding functions often used in obfuscation
		if name == "base64.b64decode" || name == "base64.decode" ||
			name == "codecs.decode" || name == "marshal.loads" {
			found = true
			return false
		}

		return true
	})
	return found
}

// getCallName returns the full name of a call expression (e.g., "os.environ.get")
func getCallName(call *ast.CallExpr) string {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name
	}
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name + "." + sel.Sel.Name
		}
	}
	return ""
}

// isGetattrCall returns true if this call is getattr
func isGetattrCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "getattr"
	}
	return false
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

// CyclomaticComplexityConfig holds thresholds for complexity detection.
var CyclomaticComplexityConfig = struct {
	Warning int
	High    int
}{
	Warning: 10,
	High:    20,
}

// detectHighCyclomaticComplexity detects functions with cyclomatic complexity > threshold.
// Cyclomatic complexity counts decision points: if, for, switch, case, &&, ||, ?.
func detectHighCyclomaticComplexity(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		complexity := calculateCyclomaticComplexity(funcDecl.Body)
		line := fset.Position(funcDecl.Pos()).Line

		if complexity > CyclomaticComplexityConfig.High {
			findings = append(findings, ASTFinding{
				Line:     line,
				Message:  fmt.Sprintf("Very high cyclomatic complexity (%d) in function %s - consider refactoring", complexity, funcDecl.Name.Name),
				Severity: "high",
			})
		} else if complexity > CyclomaticComplexityConfig.Warning {
			findings = append(findings, ASTFinding{
				Line:     line,
				Message:  fmt.Sprintf("High cyclomatic complexity (%d) in function %s", complexity, funcDecl.Name.Name),
				Severity: "medium",
			})
		}
	}

	return findings
}

// calculateCyclomaticComplexity counts decision points in a statement tree.
func calculateCyclomaticComplexity(stmt ast.Stmt) int {
	var complexity int

	ast.Inspect(stmt, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.SwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		case *ast.BinaryExpr:
			if n.Op == token.LAND || n.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	// Base complexity is 1
	if complexity == 0 {
		return 1
	}
	return complexity
}

// CognitiveComplexityConfig holds thresholds for cognitive complexity.
var CognitiveComplexityConfig = struct {
	Warning int
	High    int
}{
	Warning: 15,
	High:    30,
}

// detectHighCognitiveComplexity detects functions with high cognitive complexity.
// Cognitive complexity penalizes nested structures more than cyclomatic.
func detectHighCognitiveComplexity(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		complexity := calculateCognitiveComplexity(funcDecl.Body, 0)
		line := fset.Position(funcDecl.Pos()).Line

		if complexity > CognitiveComplexityConfig.High {
			findings = append(findings, ASTFinding{
				Line:     line,
				Message:  fmt.Sprintf("Very high cognitive complexity (%d) in function %s - deeply nested logic", complexity, funcDecl.Name.Name),
				Severity: "high",
			})
		} else if complexity > CognitiveComplexityConfig.Warning {
			findings = append(findings, ASTFinding{
				Line:     line,
				Message:  fmt.Sprintf("High cognitive complexity (%d) in function %s", complexity, funcDecl.Name.Name),
				Severity: "medium",
			})
		}
	}

	return findings
}

// calculateCognitiveComplexity calculates cognitive complexity recursively.
// Levels: +1 for each nesting level, +1 for each break in structure.
func calculateCognitiveComplexity(stmt ast.Stmt, depth int) int {
	var complexity int

	switch s := stmt.(type) {
	case *ast.IfStmt:
		complexity += depth + 1 // Structured - costs nesting level + 1
		if s.Else != nil {
			complexity += calculateCognitiveComplexity(s.Else, depth+1)
		}
		complexity += calculateCognitiveComplexity(s.Body, depth+1)
	case *ast.ForStmt:
		complexity += depth + 1
		complexity += calculateCognitiveComplexity(s.Body, depth+1)
	case *ast.RangeStmt:
		complexity += depth + 1
		complexity += calculateCognitiveComplexity(s.Body, depth+1)
	case *ast.SwitchStmt:
		complexity += depth + 1
		for _, clause := range s.Body.List {
			if ic, ok := clause.(*ast.CaseClause); ok {
				complexity += calculateCognitiveComplexity(ic, depth+1)
			}
		}
	case *ast.CaseClause:
		complexity += depth
		for _, stmt := range s.Body {
			complexity += calculateCognitiveComplexity(stmt, depth+1)
		}
	case *ast.ReturnStmt:
		complexity += depth
	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			if isRecursiveCall(call) {
				complexity += depth + 1 // Recursion breaks structure
			}
		}
	}

	return complexity
}

// isRecursiveCall checks if a call expression is a recursive call.
func isRecursiveCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name != "" // Simplified - would need func name context
	}
	return false
}

// LongFunctionConfig holds thresholds for function length.
var LongFunctionConfig = struct {
	Warning int
	High    int
}{
	Warning: 50,
	High:    100,
}

// detectLongFunction detects functions exceeding recommended line counts.
func detectLongFunction(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		startLine := fset.Position(funcDecl.Pos()).Line
		endLine := fset.Position(funcDecl.End()).Line
		lines := endLine - startLine

		if lines > LongFunctionConfig.High {
			findings = append(findings, ASTFinding{
				Line:     startLine,
				Message:  fmt.Sprintf("Very long function %s (%d lines) - consider splitting", funcDecl.Name.Name, lines),
				Severity: "medium",
			})
		} else if lines > LongFunctionConfig.Warning {
			findings = append(findings, ASTFinding{
				Line:     startLine,
				Message:  fmt.Sprintf("Long function %s (%d lines)", funcDecl.Name.Name, lines),
				Severity: "low",
			})
		}
	}

	return findings
}

// HalsteadConfig holds thresholds for Halstead metrics.
var HalsteadConfig = struct {
	DifficultyWarning float64
	VolumeWarning     float64
}{
	DifficultyWarning: 10.0,
	VolumeWarning:     300.0,
}

// detectHighHalsteadDifficulty detects functions with confusing Halstead metrics.
func detectHighHalsteadDifficulty(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		metrics := calculateHalsteadMetrics(funcDecl.Body)
		line := fset.Position(funcDecl.Pos()).Line

		if metrics.Difficulty > HalsteadConfig.DifficultyWarning {
			findings = append(findings, ASTFinding{
				Line:     line,
				Message:  fmt.Sprintf("High Halstead difficulty (%.1f) in function %s - consider simplifying", metrics.Difficulty, funcDecl.Name.Name),
				Severity: "medium",
			})
		}
	}

	return findings
}

// HalsteadMetrics holds computed Halstead software metrics.
type HalsteadMetrics struct {
	Volume      float64
	Difficulty  float64
	Effort      float64
	Operators   int
	Operands    int
	UniqueOps   int
	UniqueOprds int
}

// calculateHalsteadMetrics computes Halstead metrics for a statement.
func calculateHalsteadMetrics(stmt ast.Stmt) HalsteadMetrics {
	var m HalsteadMetrics
	operators := make(map[string]int)
	operands := make(map[string]int)

	countHalstead(stmt, operators, operands)

	m.Operators = 0
	for _, c := range operators {
		m.Operators += c
	}
	m.Operands = 0
	for _, c := range operands {
		m.Operands += c
	}
	m.UniqueOps = len(operators)
	m.UniqueOprds = len(operands)

	// Volume = N * log2(n) where N = n1 + n2, n = n1 + n2
	N := float64(m.Operators + m.Operands)
	n := float64(m.UniqueOps + m.UniqueOprds)
	if n > 0 && N > 0 {
		m.Volume = N * (math.Log2(n))
	}

	// Difficulty = (n1/2) * (N2/n2) where n1=unique ops, N2=total operands, n2=unique operands
	if m.UniqueOps > 0 && m.UniqueOprds > 0 {
		m.Difficulty = (float64(m.UniqueOps) / 2.0) * (float64(m.Operands) / float64(m.UniqueOprds))
	}

	// Effort = Volume * Difficulty
	m.Effort = m.Volume * m.Difficulty

	return m
}

// countHalstead counts operators and operands in a statement tree.
func countHalstead(stmt ast.Stmt, operators, operands map[string]int) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.BinaryExpr:
			operators[e.Op.String()]++
			// Recurse into children
		case *ast.UnaryExpr:
			operators[e.Op.String()]++
		case *ast.Ident:
			operands[e.Name]++
		case *ast.BasicLit:
			operands[e.Value]++
		case *ast.CallExpr:
			operators["call"]++
		case *ast.AssignStmt:
			for _, lhs := range e.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					operands[ident.Name]++
				}
			}
		}
		return true
	})
}

// SemanticRedundancyConfig holds thresholds for semantic redundancy.
var SemanticRedundancyConfig = struct {
	UnusedVars   bool
	DeadCode     bool
	RedundantOps bool
}{
	UnusedVars:   true,
	DeadCode:     true,
	RedundantOps: true,
}

// detectSemanticRedundancy detects semantically redundant code patterns.
func detectSemanticRedundancy(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	// Collect all function declarations first
	funcDecls := make(map[string]*ast.FuncDecl)
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && f.Name != nil {
			funcDecls[f.Name.Name] = f
		}
	}

	// Find unused functions (only in non-test files)
	for name, decl := range funcDecls {
		if !isFunctionUsed(file, name) {
			findings = append(findings, ASTFinding{
				Line:     fset.Position(decl.Pos()).Line,
				Message:  fmt.Sprintf("Unused function: %s is defined but never called", name),
				Severity: "low",
			})
		}
	}

	// Find redundant operations within functions
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}
		line := fset.Position(funcDecl.Pos()).Line
		findings = append(findings, detectRedundantOperations(funcDecl.Body, fset, line)...)
	}

	return findings
}

// isFunctionUsed checks if a function name is referenced anywhere in the file.
func isFunctionUsed(file *ast.File, funcName string) bool {
	used := false
	ast.Inspect(file, func(n ast.Node) bool {
		if used {
			return false
		}
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == funcName {
				used = true
				return false
			}
		}
		return true
	})
	return used
}

// detectRedundantOperations finds redundant operations like x = x, x+0, etc.
func detectRedundantOperations(stmt ast.Stmt, fset *token.FileSet, funcLine int) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(stmt, func(n ast.Node) bool {
		switch a := n.(type) {
		case *ast.AssignStmt:
			if len(a.Lhs) == 1 && len(a.Rhs) == 1 {
				if isRedundantAssignment(a.Lhs[0], a.Rhs[0]) {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(a.Pos()).Line,
						Message:  "Redundant assignment: variable assigned to itself",
						Severity: "low",
					})
				}
			}
		case *ast.BinaryExpr:
			if isRedundantBinaryOp(a) {
				findings = append(findings, ASTFinding{
					Line:     fset.Position(a.Pos()).Line,
					Message:  fmt.Sprintf("Redundant operation: %s", formatBinaryExpr(a)),
					Severity: "low",
				})
			}
		}
		return true
	})

	return findings
}

// isRedundantAssignment checks for x = x patterns.
func isRedundantAssignment(lhs, rhs ast.Expr) bool {
	lhsIdent, ok1 := lhs.(*ast.Ident)
	rhsIdent, ok2 := rhs.(*ast.Ident)
	if ok1 && ok2 {
		return lhsIdent.Name == rhsIdent.Name
	}
	return false
}

// isRedundantBinaryOp checks for x+0, x*1, x||true, x&&false, etc.
func isRedundantBinaryOp(expr *ast.BinaryExpr) bool {
	// x + 0, x - 0, x * 1, x / 1
	if expr.Op == token.ADD || expr.Op == token.SUB {
		if isZero(expr.Y) {
			return true
		}
	}
	if expr.Op == token.MUL {
		if isOne(expr.Y) {
			return true
		}
	}
	if expr.Op == token.QUO {
		if isOne(expr.Y) {
			return true
		}
	}
	// x || true, x && false
	if expr.Op == token.LOR {
		if isTrue(expr.Y) {
			return true
		}
	}
	if expr.Op == token.LAND {
		if isFalse(expr.Y) {
			return true
		}
	}
	return false
}

func isZero(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok {
		return lit.Kind == token.INT && lit.Value == "0"
	}
	return false
}

func isOne(expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok {
		return lit.Kind == token.INT && lit.Value == "1"
	}
	return false
}

func isTrue(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "true"
	}
	return false
}

func isFalse(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "false"
	}
	return false
}

func formatBinaryExpr(expr *ast.BinaryExpr) string {
	var op string
	switch expr.Op {
	case token.ADD:
		op = "+"
	case token.SUB:
		op = "-"
	case token.MUL:
		op = "*"
	case token.QUO:
		op = "/"
	case token.LOR:
		op = "||"
	case token.LAND:
		op = "&&"
	}
	return "x " + op + " [identity value]"
}

// Bad naming patterns for poorly-named identifiers
var badNamingPatterns = []string{
	"real_",   // real_impl, real_helper, etc.
	"helper_", // helper_impl, helper_util
	"_impl",   // foo_impl
	"impl_",   // impl_foo
	"util2",   // util2, utils2
	"util3",   // utils3
	"temp_",   // temp_var, temp_helper
	"tmp_",    // tmp_file
	"misc_",   // misc_helper
	"do_",     // do_something - verbs at start often indicate poor design
}

// detectInconsistentBooleanNaming detects mixed use of Yes/No and True/False style names.
func detectInconsistentBooleanNaming(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding
	var boolVars []string

	// Collect all boolean variable/function names
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name != nil {
				name := d.Name.Name
				if isBooleanIdentifier(name) {
					boolVars = append(boolVars, name)
				}
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, ident := range vs.Names {
						if isBooleanIdentifier(ident.Name) {
							boolVars = append(boolVars, ident.Name)
						}
					}
				}
			}
		}
	}

	// Categorize boolean names
	var yesNoStyle, trueFalseStyle []string
	for _, name := range boolVars {
		if strings.HasPrefix(name, "is") || strings.HasPrefix(name, "are") ||
			strings.HasPrefix(name, "has") || strings.HasPrefix(name, "was") ||
			strings.HasPrefix(name, "were") || strings.HasPrefix(name, "does") ||
			strings.HasPrefix(name, "can") || strings.HasPrefix(name, "could") ||
			strings.HasPrefix(name, "will") || strings.HasPrefix(name, "would") ||
			strings.HasPrefix(name, "should") {
			yesNoStyle = append(yesNoStyle, name)
		} else if strings.HasPrefix(name, "enable") || strings.HasPrefix(name, "disable") ||
			strings.HasPrefix(name, "set") || strings.HasPrefix(name, "flag") ||
			strings.HasPrefix(name, "active") || strings.HasPrefix(name, "enabled") ||
			strings.HasPrefix(name, "valid") || strings.HasPrefix(name, "found") ||
			strings.HasPrefix(name, "ok") || strings.HasPrefix(name, "success") {
			trueFalseStyle = append(trueFalseStyle, name)
		}
	}

	// Report mixing of styles
	if len(yesNoStyle) > 0 && len(trueFalseStyle) > 0 {
		findings = append(findings, ASTFinding{
			Line:     fset.Position(file.Package).Line,
			Message:  fmt.Sprintf("Inconsistent boolean naming: mixing Yes/No style (%v) with True/False style (%v). Prefer consistent naming.", yesNoStyle[:min(3, len(yesNoStyle))], trueFalseStyle[:min(3, len(trueFalseStyle))]),
			Severity: "low",
		})
	}

	return findings
}

func isBooleanIdentifier(name string) bool {
	// Simple check for common boolean prefixes/suffixes
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, "is") || strings.HasPrefix(lower, "are") ||
		strings.HasPrefix(lower, "has") || strings.HasPrefix(lower, "was") ||
		strings.HasPrefix(lower, "were") || strings.HasPrefix(lower, "does") ||
		strings.HasPrefix(lower, "can") || strings.HasPrefix(lower, "could") ||
		strings.HasPrefix(lower, "will") || strings.HasPrefix(lower, "would") ||
		strings.HasPrefix(lower, "should") || strings.HasPrefix(lower, "enable") ||
		strings.HasPrefix(lower, "disable") || strings.HasSuffix(lower, "_flag") ||
		strings.HasSuffix(lower, "_enabled") || strings.HasSuffix(lower, "_active") ||
		strings.HasSuffix(lower, "_valid") || strings.HasSuffix(lower, "_found") ||
		strings.HasSuffix(lower, "_ok") || strings.HasSuffix(lower, "_success")
}

// MaxWrapperChainDepth is the maximum allowed depth for wrapper chains.
var MaxWrapperChainDepth = 2

// detectDeepWrapperChain detects wrapper chains that exceed the configured depth.
func detectDeepWrapperChain(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		// Count wrapper calls in function body
		wrapperDepth := countWrapperChain(funcDecl.Body)
		if wrapperDepth > MaxWrapperChainDepth {
			findings = append(findings, ASTFinding{
				Line:     fset.Position(funcDecl.Pos()).Line,
				Message:  fmt.Sprintf("Deep wrapper chain (%d) in function %s - exceeds max depth of %d", wrapperDepth, funcDecl.Name.Name, MaxWrapperChainDepth),
				Severity: "medium",
			})
		}
	}

	return findings
}

// countWrapperChain counts the depth of nested wrapper calls.
func countWrapperChain(stmt ast.Stmt) int {
	maxDepth := 0

	ast.Inspect(stmt, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			depth := countSingleChain(call)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return true
	})

	return maxDepth
}

// countSingleChain counts the chain depth of a single call expression.
func countSingleChain(call *ast.CallExpr) int {
	depth := 0
	current := call

	for current != nil {
		if sel, ok := current.Fun.(*ast.SelectorExpr); ok {
			// Check if the selector is a wrapper pattern (getter/setter/delegator)
			name := sel.Sel.Name
			if isWrapperPattern(name) {
				depth++
				if recv, ok := sel.X.(*ast.CallExpr); ok {
					current = recv
				} else {
					current = nil
				}
			} else {
				current = nil
			}
		} else {
			current = nil
		}
	}

	return depth
}

// isWrapperPattern returns true if the function name suggests it's a wrapper.
func isWrapperPattern(name string) bool {
	wrappers := []string{"get", "set", "add", "remove", "update", "create", "delete",
		"fetch", "retrieve", "load", "save", "handle", "process", "wrap", "unwrap",
		"delegate", "forward", "proxy", "adapt", "transform", "convert"}
	for _, w := range wrappers {
		if name == w {
			return true
		}
	}
	return false
}

// PublicAPISignature represents a public API signature.
type PublicAPISignature struct {
	Name       string
	Parameters []string
	ReturnType string
	File       string
}

// detectDuplicatePublicAPI detects duplicate public API signatures.
func detectDuplicatePublicAPI(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding
	var publicAPIs []PublicAPISignature

	// Collect all public APIs
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil {
			continue
		}

		// Only exported functions (starts with uppercase)
		if !ast.IsExported(funcDecl.Name.Name) {
			continue
		}

		sig := PublicAPISignature{
			Name:       funcDecl.Name.Name,
			Parameters: extractParameterTypes(funcDecl),
			ReturnType: extractReturnType(funcDecl),
			File:       fset.Position(funcDecl.Pos()).Filename,
		}
		publicAPIs = append(publicAPIs, sig)
	}

	// Check for duplicates
	for i := 0; i < len(publicAPIs); i++ {
		for j := i + 1; j < len(publicAPIs); j++ {
			if publicAPIs[i].Name == publicAPIs[j].Name &&
				publicAPIs[i].ReturnType == publicAPIs[j].ReturnType &&
				slicesEqual(publicAPIs[i].Parameters, publicAPIs[j].Parameters) {
				findings = append(findings, ASTFinding{
					Line:     fset.Position(file.Package).Line,
					Message:  fmt.Sprintf("Duplicate public API: %s (also defined in %s)", publicAPIs[i].Name, publicAPIs[j].File),
					Severity: "medium",
				})
			}
		}
	}

	return findings
}

func extractParameterTypes(funcDecl *ast.FuncDecl) []string {
	var types []string
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			types = append(types, typeToString(param.Type))
		}
	}
	return types
}

func extractReturnType(funcDecl *ast.FuncDecl) string {
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		return typeToString(funcDecl.Type.Results.List[0].Type)
	}
	return "void"
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SliceExpr:
		return "[]" + typeToString(t.X)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func"
	default:
		return "unknown"
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// MaxAbstractionDepth is the maximum allowed abstraction depth.
var MaxAbstractionDepth = 3

// detectExcessiveAbstractionDepth detects interfaces nested too deeply.
func detectExcessiveAbstractionDepth(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	// Count interface nesting depth
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if iface, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						depth := countInterfaceDepth(iface)
						if depth > MaxAbstractionDepth {
							findings = append(findings, ASTFinding{
								Line:     fset.Position(genDecl.Pos()).Line,
								Message:  fmt.Sprintf("Excessive interface abstraction depth (%d) - max recommended is %d", depth, MaxAbstractionDepth),
								Severity: "medium",
							})
						}
					}
				}
			}
		}
	}

	return findings
}

// countInterfaceDepth counts the nesting depth of embedded interfaces.
func countInterfaceDepth(iface *ast.InterfaceType) int {
	maxDepth := 0
	for _, field := range iface.Methods.List {
		if embed, ok := field.Type.(*ast.InterfaceType); ok {
			depth := 1 + countInterfaceDepth(embed)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}
	return maxDepth
}

// detectMultipleResponsibilities detects types that may have multiple responsibilities.
func detectMultipleResponsibilities(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		typeSpec, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range typeSpec.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				methods := countMethodsForType(file, ts.Name.Name)
				if methods > 20 {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(ts.Pos()).Line,
						Message:  fmt.Sprintf("Type %s has %d methods - may have multiple responsibilities. Consider splitting.", ts.Name.Name, methods),
						Severity: "medium",
					})
				}
			}
		}
	}

	return findings
}

// countMethodsForType counts the number of methods associated with a type.
func countMethodsForType(file *ast.File, typeName string) int {
	count := 0
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
					if ident.Name == typeName {
						count++
					}
				}
				// Also check pointer receiver
				if star, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
					if ident, ok := star.X.(*ast.Ident); ok {
						if ident.Name == typeName {
							count++
						}
					}
				}
			}
		}
	}
	return count
}

// detectPoorlyNamedIdentifier detects poorly named identifiers.
func detectPoorlyNamedIdentifier(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			for _, pattern := range badNamingPatterns {
				if strings.Contains(strings.ToLower(ident.Name), pattern) {
					findings = append(findings, ASTFinding{
						Line:     fset.Position(ident.Pos()).Line,
						Message:  fmt.Sprintf("Poorly named identifier: %s matches pattern '%s'", ident.Name, pattern),
						Severity: "low",
					})
					break
				}
			}
		}
		return true
	})

	return findings
}

// ConfigValue represents a configuration value for duplication detection.
type ConfigValue struct {
	Key   string
	Value string
}

// detectDuplicateConfiguration detects duplicate configuration values.
func detectDuplicateConfiguration(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding
	var configs []ConfigValue

	// Look for config patterns (struct tags with json/yaml, const declarations)
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for i, name := range vs.Names {
						if i < len(vs.Values) {
							if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
								configs = append(configs, ConfigValue{
									Key:   name.Name,
									Value: lit.Value,
								})
							}
						}
					}
				}
			}
		}
	}

	// Check for duplicate values with different keys
	for i := 0; i < len(configs); i++ {
		for j := i + 1; j < len(configs); j++ {
			if configs[i].Value == configs[j].Value &&
				configs[i].Key != configs[j].Key &&
				!strings.Contains(configs[i].Key, "test") &&
				!strings.Contains(configs[j].Key, "test") {
				findings = append(findings, ASTFinding{
					Line:     fset.Position(file.Package).Line,
					Message:  fmt.Sprintf("Duplicate configuration value: %s and %s have identical value '%s'", configs[i].Key, configs[j].Key, configs[i].Value),
					Severity: "low",
				})
			}
		}
	}

	return findings
}

// detectMissingReturn checks for functions that have conditional returns
// but no final return. E.g., if x > 5: return x (no return after if).
func detectMissingReturn(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		// Skip if return type indicates no value needed (void)
		if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
			continue
		}

		// Check if function body has conditional returns but no unconditional return at end
		hasConditionalReturn := hasConditionalReturnPath(funcDecl.Body)
		hasUnconditionalReturn := hasUnconditionalReturnAtEnd(funcDecl.Body)

		if hasConditionalReturn && !hasUnconditionalReturn {
			findings = append(findings, ASTFinding{
				Line:     fset.Position(funcDecl.Pos()).Line,
				Message:  fmt.Sprintf("Function %s may return without a value on some code paths", funcDecl.Name.Name),
				Severity: "high",
			})
		}
	}

	return findings
}

// hasConditionalReturnPath checks if the function has a return inside a conditional (if/for/etc).
func hasConditionalReturnPath(stmt ast.Stmt) bool {
	hasReturn := false

	ast.Inspect(stmt, func(n ast.Node) bool {
		if hasReturn {
			return false
		}

		switch s := n.(type) {
		case *ast.IfStmt:
			// Check if there's a return in the if body
			if hasReturnInBody(s.Body) {
				hasReturn = true
				return false
			}
			// Check else body too
			if s.Else != nil {
				if retStmt, ok := s.Else.(*ast.ReturnStmt); ok && retStmt.Results != nil {
					hasReturn = true
					return false
				}
				if hasReturnInBody(s.Else) {
					hasReturn = true
					return false
				}
			}
		case *ast.ForStmt:
			if hasReturnInBody(s.Body) {
				hasReturn = true
				return false
			}
		case *ast.RangeStmt:
			if hasReturnInBody(s.Body) {
				hasReturn = true
				return false
			}
		case *ast.SwitchStmt:
			for _, clause := range s.Body.List {
				if hasReturnInBody(clause) {
					hasReturn = true
					return false
				}
			}
		}
		return true
	})

	return hasReturn
}

// hasReturnInBody checks if a statement list contains a return statement.
func hasReturnInBody(n ast.Node) bool {
	hasReturn := false
	ast.Inspect(n, func(node ast.Node) bool {
		if hasReturn {
			return false
		}
		if ret, ok := node.(*ast.ReturnStmt); ok && ret.Results != nil {
			hasReturn = true
			return false
		}
		return true
	})
	return hasReturn
}

// hasUnconditionalReturnAtEnd checks if the function's last statement is a return.
func hasUnconditionalReturnAtEnd(stmt ast.Stmt) bool {
	block, ok := stmt.(*ast.BlockStmt)
	if !ok {
		return false
	}

	if len(block.List) == 0 {
		return false
	}

	last := block.List[len(block.List)-1]
	_, isUnconditionalReturn := last.(*ast.ReturnStmt)
	return isUnconditionalReturn
}

// detectDuplicateCondition checks for duplicate conditions in if/elif chains.
func detectDuplicateCondition(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		checkDuplicateConditionsInStmt(funcDecl.Body, fset, &findings)
	}

	return findings
}

// checkDuplicateConditionsInStmt walks a statement tree looking for if chains.
func checkDuplicateConditionsInStmt(stmt ast.Stmt, fset *token.FileSet, findings *[]ASTFinding) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Collect all conditions in the if-elif chain
		conditions := []ast.Expr{ifStmt.Cond}
		elseNode := ifStmt.Else

		for elseNode != nil {
			switch elseClause := elseNode.(type) {
			case *ast.IfStmt:
				conditions = append(conditions, elseClause.Cond)
				elseNode = elseClause.Else
			default:
				elseNode = nil
			}
		}

		// Check for duplicate conditions
		seen := make(map[string]bool)
		for i, cond := range conditions {
			condStr := exprToString(cond)
			if condStr == "" {
				continue
			}
			if seen[condStr] {
				line := fset.Position(cond.Pos()).Line
				*findings = append(*findings, ASTFinding{
					Line:     line,
					Message:  fmt.Sprintf("Duplicate condition in if/elif chain: condition appears more than once"),
					Severity: "medium",
				})
				break
			}
			seen[condStr] = true
			_ = i // silence unused variable
		}

		return true
	})
}

// exprToString converts an expression to a canonical string for comparison.
func exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.BasicLit:
		return e.Value
	case *ast.BinaryExpr:
		return exprToString(e.X) + " " + e.Op.String() + " " + exprToString(e.Y)
	case *ast.UnaryExpr:
		return e.Op.String() + exprToString(e.X)
	case *ast.ParenExpr:
		return "(" + exprToString(e.X) + ")"
	case *ast.SelectorExpr:
		var xName string
		if ident, ok := e.X.(*ast.Ident); ok {
			xName = ident.Name
		}
		return xName + "." + e.Sel.Name
	case *ast.IndexExpr:
		var xName string
		if ident, ok := e.X.(*ast.Ident); ok {
			xName = ident.Name
		}
		return xName + "[" + exprToString(e.Index) + "]"
	case *ast.CallExpr:
		return callExprToString(e)
	default:
		return ""
	}
}

// callExprToString converts a call expression to string.
func callExprToString(call *ast.CallExpr) string {
	var fnName string
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		fnName = fn.Name
	case *ast.SelectorExpr:
		if ident, ok := fn.X.(*ast.Ident); ok {
			fnName = ident.Name + "." + fn.Sel.Name
		} else {
			fnName = fn.Sel.Name
		}
	}
	var args []string
	for _, arg := range call.Args {
		args = append(args, exprToString(arg))
	}
	return fnName + "(" + strings.Join(args, ",") + ")"
}

// detectImpossibleBranch detects impossible branches like if x is None: ... elif x is None:.
func detectImpossibleBranch(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		checkImpossibleBranches(funcDecl.Body, fset, &findings)
	}

	return findings
}

// checkImpossibleBranches walks the AST looking for impossible branch patterns.
func checkImpossibleBranches(stmt ast.Stmt, fset *token.FileSet, findings *[]ASTFinding) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Track type constraints through if-elif chain
		// After if x == nil, else means x is not nil
		constraints := make(map[string]string) // var name -> constraint

		// Process else if chain - check each elif for contradictions
		elseNode := ifStmt.Else
		for elseNode != nil {
			switch elseClause := elseNode.(type) {
			case *ast.IfStmt:
				// Before checking, establish that parent if was false
				// (we're in else, so parent condition was not true)
				updateConstraintsFromFalseBranch(ifStmt.Cond, constraints)

				// Now check if this elif contradicts what we know
				if checkBranchContradiction(elseClause.Cond, constraints) {
					line := fset.Position(elseClause.Cond.Pos()).Line
					*findings = append(*findings, ASTFinding{
						Line:     line,
						Message:  "Impossible branch: condition contradicts earlier type check",
						Severity: "high",
					})
					return false
				}
				// Now add this elif's condition for subsequent elif checks
				extractConstraints(elseClause.Cond, constraints)
				// Move to next elif
				ifStmt = elseClause // For next iteration, this elif becomes the parent
				elseNode = elseClause.Else
			default:
				elseNode = nil
			}
		}

		return true
	})
}

// updateConstraintsFromFalseBranch updates constraints when we know the parent condition was false.
// If we had "if x == nil", and we're in else, then x is not nil.
func updateConstraintsFromFalseBranch(cond ast.Expr, constraints map[string]string) {
	if bin, ok := cond.(*ast.BinaryExpr); ok {
		if bin.Op == token.EQL || bin.Op == token.NEQ {
			if isNoneOrNil(bin.Y) {
				if ident, ok := bin.X.(*ast.Ident); ok {
					if bin.Op == token.EQL {
						// After if x == nil was false, x is not nil
						constraints[ident.Name] = "not_nil"
					}
				}
			}
		}
	}
}

// extractConstraints extracts type constraints from a condition.
func extractConstraints(cond ast.Expr, constraints map[string]string) {
	// Handle "x is None" or "x == nil"
	if bin, ok := cond.(*ast.BinaryExpr); ok {
		if bin.Op == token.EQL || bin.Op == token.NEQ {
			if isNoneOrNil(bin.Y) {
				if ident, ok := bin.X.(*ast.Ident); ok {
					if bin.Op == token.EQL {
						constraints[ident.Name] = "nil"
					} else {
						delete(constraints, ident.Name)
					}
				}
			}
		}
	}
}

// isNoneOrNil checks if an expression represents None/nil.
func isNoneOrNil(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "nil" || ident.Name == "None"
	}
	return false
}

// checkBranchContradiction checks if a condition in else-if contradicts constraints.
// When we reach else-if, we know prior if conditions were false, so constraints reflect that.
func checkBranchContradiction(cond ast.Expr, constraints map[string]string) bool {
	if bin, ok := cond.(*ast.BinaryExpr); ok {
		if bin.Op == token.EQL {
			if isNoneOrNil(bin.Y) {
				if ident, ok := bin.X.(*ast.Ident); ok {
					// If we already know x is not_nil, then "x == nil" is impossible
					if existing, exists := constraints[ident.Name]; exists {
						return existing == "not_nil"
					}
				}
			}
		}
	}
	return false
}

// detectIteratorModification detects modification of iterator target during iteration.
func detectIteratorModification(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		findIteratorModifications(funcDecl, fset, &findings)
	}

	return findings
}

// findIteratorModifications checks for concurrent modification of iterators.
func findIteratorModifications(funcDecl *ast.FuncDecl, fset *token.FileSet, findings *[]ASTFinding) {
	// For each for loop, collect the iterator variable and check for modifications
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		forStmt, ok := n.(*ast.RangeStmt)
		if !ok {
			return true
		}

		// Get the iterator variable
		var iterVar string
		switch ident := forStmt.X.(type) {
		case *ast.Ident:
			iterVar = ident.Name
		}

		if iterVar == "" {
			return true
		}

		// Check if the iterator variable is modified inside the loop body
		if modifiesVariable(forStmt.Body, iterVar) {
			line := fset.Position(forStmt.Pos()).Line
			*findings = append(*findings, ASTFinding{
				Line:     line,
				Message:  fmt.Sprintf("Iterator modification: variable %s is modified during iteration", iterVar),
				Severity: "high",
			})
		}

		return true
	})
}

// modifiesVariable checks if any statement modifies the given variable.
func modifiesVariable(stmt ast.Stmt, varName string) bool {
	modified := false

	ast.Inspect(stmt, func(n ast.Node) bool {
		if modified {
			return false
		}

		switch a := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range a.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					if ident.Name == varName {
						modified = true
						return false
					}
				}
				// Check for index expressions like list[i]
				if _, ok := lhs.(*ast.IndexExpr); ok {
					// Could be modifying the variable if it's a range over the same thing
				}
			}
		case *ast.IncDecStmt:
			if ident, ok := a.X.(*ast.Ident); ok {
				if ident.Name == varName {
					modified = true
					return false
				}
			}
		case *ast.SendStmt:
			if ident, ok := a.Chan.(*ast.Ident); ok {
				if ident.Name == varName {
					modified = true
					return false
				}
			}
		}

		return true
	})

	return modified
}

// CFG-based pattern wrappers that convert CFGFinding to ASTFinding

// detectLockNotReleased wraps DetectLockNotReleased for ASTPattern interface.
func detectLockNotReleased(fset *token.FileSet, file *ast.File) []ASTFinding {
	cfgFindings := DetectLockNotReleased(fset, file)
	var findings []ASTFinding
	for _, f := range cfgFindings {
		findings = append(findings, ASTFinding{
			File:     f.File,
			Line:     f.Line,
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: f.Severity,
		})
	}
	return findings
}

// detectResourceLeak wraps DetectResourceLeak for ASTPattern interface.
func detectResourceLeak(fset *token.FileSet, file *ast.File) []ASTFinding {
	cfgFindings := DetectResourceLeak(fset, file)
	var findings []ASTFinding
	for _, f := range cfgFindings {
		findings = append(findings, ASTFinding{
			File:     f.File,
			Line:     f.Line,
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: f.Severity,
		})
	}
	return findings
}

// detectTransactionBug wraps DetectTransactionBug for ASTPattern interface.
func detectTransactionBug(fset *token.FileSet, file *ast.File) []ASTFinding {
	cfgFindings := DetectTransactionBug(fset, file)
	var findings []ASTFinding
	for _, f := range cfgFindings {
		findings = append(findings, ASTFinding{
			File:     f.File,
			Line:     f.Line,
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: f.Severity,
		})
	}
	return findings
}

// detectNullDereference detects potential nil pointer dereferences.
// It tracks variables that may be assigned nil and checks for their subsequent use.
func detectNullDereference(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		findNullDereferences(funcDecl, fset, &findings)
	}

	return findings
}

// findNullDereferences checks a function for nil dereference patterns.
func findNullDereferences(funcDecl *ast.FuncDecl, fset *token.FileSet, findings *[]ASTFinding) {
	// Track variables that are potentially nil
	nilVars := make(map[string]bool)
	// Track where variables were last assigned nil
	lastNilAssign := make(map[string]int)

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			// Check bounds before accessing RHS
			if len(stmt.Rhs) >= len(stmt.Lhs) {
				for i, lhs := range stmt.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						if isNilLit(stmt.Rhs[i]) {
							nilVars[ident.Name] = true
							lastNilAssign[ident.Name] = fset.Position(ident.Pos()).Line
						} else if ident.Name != "_" {
							// Variable is assigned something that might not be nil
							delete(nilVars, ident.Name)
							delete(lastNilAssign, ident.Name)
						}
					}
				}
			}

		case *ast.ReturnStmt:
			for _, res := range stmt.Results {
				if ident, ok := res.(*ast.Ident); ok && ident.Name != "" {
					delete(nilVars, ident.Name)
				}
			}
		}

		// Check for selector expressions that might dereference nil
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				if nilVars[ident.Name] {
					line := fset.Position(sel.Pos()).Line
					if lastLine, exists := lastNilAssign[ident.Name]; exists && line > lastLine {
						*findings = append(*findings, ASTFinding{
							Line:     line,
							Message:  fmt.Sprintf("Potential nil dereference: %s used after possible nil assignment", ident.Name),
							Severity: "high",
						})
					}
				}
			}
		}

		// Check for index expressions on potentially nil slices
		if idx, ok := n.(*ast.IndexExpr); ok {
			if ident, ok := idx.X.(*ast.Ident); ok {
				if nilVars[ident.Name] {
					line := fset.Position(idx.Pos()).Line
					if lastLine, exists := lastNilAssign[ident.Name]; exists && line > lastLine {
						*findings = append(*findings, ASTFinding{
							Line:     line,
							Message:  fmt.Sprintf("Potential nil dereference: slice %s indexed after possible nil", ident.Name),
							Severity: "high",
						})
					}
				}
			}
		}

		// Check for method calls on potentially nil pointers
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if nilVars[ident.Name] {
						line := fset.Position(call.Pos()).Line
						if lastLine, exists := lastNilAssign[ident.Name]; exists && line > lastLine {
							*findings = append(*findings, ASTFinding{
								Line:     line,
								Message:  fmt.Sprintf("Potential nil dereference: method call on %s after possible nil", ident.Name),
								Severity: "high",
							})
						}
					}
				}
			}
		}

		return true
	})
}

// isNilLit checks if an expression is a nil literal.
func isNilLit(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "nil"
	}
	return false
}

// detectDeadAssignment detects variables that are assigned but never used.
func detectDeadAssignment(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		findDeadAssignments(funcDecl, fset, &findings)
	}

	return findings
}

// findDeadAssignments checks a function for dead assignment patterns.
func findDeadAssignments(funcDecl *ast.FuncDecl, fset *token.FileSet, findings *[]ASTFinding) {
	// Collect all used variables in the function body
	usedVars := make(map[string]bool)
	assignedVars := make(map[string]int) // Line where last assigned

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		// Track variable declarations (:= or var)
		if assign, ok := n.(*ast.AssignStmt); ok {
			if assign.Tok == token.DEFINE {
				for _, lhs := range assign.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						if ident.Name != "_" {
							assignedVars[ident.Name] = fset.Position(ident.Pos()).Line
						}
					}
				}
				// Mark RHS identifiers as used
				for _, rhs := range assign.Rhs {
					markIdentifiersUsed(rhs, usedVars)
				}
			}
		}

		// Track function call arguments - marks variables as used
		if call, ok := n.(*ast.CallExpr); ok {
			for _, arg := range call.Args {
				markIdentifiersUsed(arg, usedVars)
			}
		}

		return true
	})

	// Find variables assigned but never used in a way that contributes to output
	for varName, assignLine := range assignedVars {
		if !usedVars[varName] {
			// Check if it's not a function parameter
			isParam := false
			if funcDecl.Type.Params != nil {
				for _, param := range funcDecl.Type.Params.List {
					for _, name := range param.Names {
						if name.Name == varName {
							isParam = true
							break
						}
					}
				}
			}

			if !isParam {
				*findings = append(*findings, ASTFinding{
					Line:     assignLine,
					Message:  fmt.Sprintf("Dead assignment: variable %s is assigned but never used", varName),
					Severity: "low",
				})
			}
		}
	}
}

// isLHSOfAssign checks if an identifier is on the LHS of an assignment.
func isLHSOfAssign(ident *ast.Ident) bool {
	parent := ident
	if parent.Obj != nil && parent.Obj.Kind == ast.Var {
		return true
	}
	return false
}

// markIdentifiersUsed recursively marks all identifiers in an expression as used.
func markIdentifiersUsed(expr ast.Expr, usedVars map[string]bool) {
	ast.Inspect(expr, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name != "_" {
				usedVars[ident.Name] = true
			}
			return false
		}
		return true
	})
}
