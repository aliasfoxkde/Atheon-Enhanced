// Package core provides pattern matching and scanning capabilities.
package core

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// TaintTracker detects data flow from sources (user input) to sinks (sensitive operations).
// This enables detection of vulnerabilities like command injection, SQL injection, etc.
type TaintTracker struct {
	// sources identifies data origin points (user input, file reads, network input)
	sources map[string]bool
	// sinks identifies dangerous destination operations (exec, eval, file writes, network)
	sinks map[string]bool
	// tainted tracks variable names that hold untrusted data
	tainted map[string]bool
}

// Severity weights for risk scoring
const (
	SeverityCritical = 40
	SeverityHigh     = 20
	SeverityMedium   = 10
	SeverityLow      = 5
)

// NewTaintTracker creates a new taint tracker with default source and sink patterns.
func NewTaintTracker() *TaintTracker {
	return &TaintTracker{
		sources: map[string]bool{
			"os.Getenv":      true,
			"os.environ":     true,
			"input":          true,
			"sys.stdin":      true,
			"request.form":   true,
			"request.args":   true,
			"request.json":   true,
			"open":           true,
			"pathlib.Path":   true,
			"read_file":      true,
			"requests.get":   true,
			"requests.post":  true,
			"httpx.get":      true,
			"urllib.request": true,
			"fetch":          true,
			"aiohttp":        true,
		},
		sinks: map[string]bool{
			"exec":           true,
			"eval":           true,
			"exec.Command":   true,
			"os.system":      true,
			"subprocess":     true,
			"os.open":        true,
			"open":           true,
			"write":          true,
			"requests.post":  true,
			"httpx.post":     true,
			"urllib.request": true,
			"send":           true,
			"compile":        true,
		},
		tainted: make(map[string]bool),
	}
}

// Source represents a taint source (data origin).
type Source struct {
	Name    string // Function or method name
	Line    int    // Source location
	Content string // Code snippet
}

// Sink represents a taint sink (dangerous operation).
type Sink struct {
	Name    string // Function or method name
	Line    int    // Sink location
	Content string // Code snippet
}

// TaintFinding represents a finding where taint flows from source to sink.
type TaintFinding struct {
	Source      Source
	Sink        Sink
	Description string
	Severity    string
	RiskScore   int
}

// TrackSource marks a variable or expression as tainted (originating from untrusted input).
func (t *TaintTracker) TrackSource(name string) {
	t.tainted[name] = true
}

// IsTainted checks if a variable or expression is marked as tainted.
func (t *TaintTracker) IsTainted(name string) bool {
	return t.tainted[name]
}

// ClearTaint removes the taint mark from a variable.
func (t *TaintTracker) ClearTaint(name string) {
	delete(t.tainted, name)
}

// ScanFileAST performs AST-based taint analysis on a Go source file.
func (t *TaintTracker) ScanFileAST(filename string) ([]TaintFinding, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var findings []TaintFinding

	// Walk the AST to find source → sink flows
	ast.Inspect(node, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.CallExpr:
			findings = append(findings, t.analyzeCall(expr, fset)...)
		case *ast.AssignStmt:
			t.analyzeAssignment(expr, fset)
		}
		return true
	})

	return findings, nil
}

// analyzeCall examines a function call for source or sink patterns.
func (t *TaintTracker) analyzeCall(call *ast.CallExpr, fset *token.FileSet) []TaintFinding {
	var findings []TaintFinding

	// Get the function name
	fnName := t.getFuncName(call.Fun)
	if fnName == "" {
		return findings
	}

	pos := fset.Position(call.Pos())

	// Check if this is a sink
	if t.sinks[fnName] {
		// Check if any argument is tainted
		for _, arg := range call.Args {
			if t.isExprTainted(arg) {
				findings = append(findings, TaintFinding{
					Source: Source{
						Name:    "tainted_argument",
						Line:    pos.Line,
						Content: t.getExprText(arg),
					},
					Sink: Sink{
						Name:    fnName,
						Line:    pos.Line,
						Content: t.getExprText(call),
					},
					Description: "Tainted data flows to sensitive operation",
					Severity:    "high",
					RiskScore:   SeverityHigh,
				})
				break
			}
		}
	}

	return findings
}

// analyzeAssignment tracks variable assignments and propagates taint.
func (t *TaintTracker) analyzeAssignment(stmt *ast.AssignStmt, fset *token.FileSet) {
	for i, expr := range stmt.Rhs {
		rhsTainted := t.isExprTainted(expr)
		if rhsTainted {
			// Mark left-hand side variables as tainted
			for _, lhs := range stmt.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					t.TrackSource(ident.Name)
				}
			}
		}
		_ = i // silence unused variable warning
	}
}

// getFuncName extracts the function name from various expression types.
func (t *TaintTracker) getFuncName(expr ast.Expr) string {
	switch fn := expr.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if ident, ok := fn.X.(*ast.Ident); ok {
			return ident.Name + "." + fn.Sel.Name
		}
	}
	return ""
}

// isExprTainted checks if an expression contains tainted data.
func (t *TaintTracker) isExprTainted(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return t.IsTainted(e.Name)
	case *ast.CallExpr:
		fnName := t.getFuncName(e.Fun)
		return t.sources[fnName]
	case *ast.BinaryExpr:
		return t.isExprTainted(e.X) || t.isExprTainted(e.Y)
	case *ast.ParenExpr:
		return t.isExprTainted(e.X)
	}
	return false
}

// getExprText returns a text representation of an expression.
func (t *TaintTracker) getExprText(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.CallExpr:
		return t.getFuncName(e.Fun) + "(...)"
	case *ast.BinaryExpr:
		return t.getExprText(e.X) + " " + e.Op.String() + " " + t.getExprText(e.Y)
	case *ast.BasicLit:
		return e.Value
	}
	return "..."
}

// SeverityLevel returns the severity string for a risk score.
func SeverityLevel(score int) string {
	switch {
	case score >= 30:
		return "critical"
	case score >= 20:
		return "high"
	case score >= 10:
		return "medium"
	default:
		return "low"
	}
}

// CalculateRiskScore computes a 0-100 risk score for a finding.
func CalculateRiskScore(severityWeight int, confidence float64) int {
	score := int(float64(severityWeight) * confidence)
	if score > 100 {
		return 100
	}
	return score
}

// ScanForTaintPatterns scans content for patterns that indicate taint flow issues.
func ScanForTaintPatterns(content string) []Finding {
	var findings []Finding

	// Simple line-based scan for taint indicators
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "os.getenv") || strings.Contains(line, "request.") {
			if strings.Contains(line, "exec(") || strings.Contains(line, "system(") {
				findings = append(findings, Finding{
					Pattern:  "taint-command-injection",
					File:     "scan",
					Line:     i + 1,
					Content:  strings.TrimSpace(line),
					Severity: "high",
					Category: "security",
				})
			}
		}
	}

	return findings
}
