package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// AuditLayer represents a layer in the layered audit system.
// Each layer focuses on a specific aspect of code quality.
type AuditLayer int

const (
	// LayerFormatting is Layer 1: Formatting - whitespace, line length, naming conventions
	LayerFormatting AuditLayer = iota + 1
	// LayerSyntax is Layer 2: Syntax - basic AST syntax errors
	LayerSyntax
	// LayerTypeChecking is Layer 3: TypeChecking - Go type system
	LayerTypeChecking
	// LayerSecurity is Layer 4: Security - injection vulnerabilities
	LayerSecurity
	// LayerComplexity is Layer 5: Complexity - Cyclomatic, Cognitive, Halstead
	LayerComplexity
	// LayerCodeSmells is Layer 6: CodeSmells - Long functions, dead code
	LayerCodeSmells
	// LayerArchitecture is Layer 7: Architecture - wrapper chains, abstraction depth
	LayerArchitecture
	// LayerRedundancy is Layer 8: Redundancy - clone detection
	LayerRedundancy
	// LayerSpecConformance is Layer 9: SpecConformance - pattern validation
	LayerSpecConformance
)

// LayerInfo describes an audit layer.
type LayerInfo struct {
	Layer       AuditLayer
	Name        string
	Description string
	Order       int
}

// AuditLayerDefinitions defines all available audit layers.
var AuditLayerDefinitions = []LayerInfo{
	{LayerFormatting, "Formatting", "Whitespace, line length, naming conventions", 1},
	{LayerSyntax, "Syntax", "AST syntax errors and structural issues", 2},
	{LayerTypeChecking, "Type Checking", "Go type system validation", 3},
	{LayerSecurity, "Security", "Injection vulnerabilities, credentials, unsafe operations", 4},
	{LayerComplexity, "Complexity", "Cyclomatic, Cognitive, Halstead metrics", 5},
	{LayerCodeSmells, "Code Smells", "Long functions, dead code, redundancy", 6},
	{LayerArchitecture, "Architecture", "Engineering policies: wrapper chains, abstraction depth", 7},
	{LayerRedundancy, "Redundancy", "Clone detection, duplicate code", 8},
	{LayerSpecConformance, "Spec Conformance", "Pattern validation and specification adherence", 9},
}

// AuditFinding represents a finding from any audit layer.
type AuditFinding struct {
	Layer     AuditLayer
	LayerName string
	File      string
	Line      int
	Rule      string
	Message   string
	Severity  string
}

// AuditConfig configures which layers to run.
type AuditConfig struct {
	Layers          []AuditLayer
	MinSeverity     string
	IncludePatterns []string
	ExcludePatterns []string
}

// DefaultAuditConfig returns a config with all layers enabled.
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Layers: []AuditLayer{
			LayerFormatting,
			LayerSyntax,
			LayerTypeChecking,
			LayerSecurity,
			LayerComplexity,
			LayerCodeSmells,
			LayerArchitecture,
			LayerRedundancy,
			LayerSpecConformance,
		},
	}
}

// RunAuditLayers runs all configured audit layers on a file.
func RunAuditLayers(path string, config *AuditConfig) ([]AuditFinding, error) {
	if config == nil {
		config = DefaultAuditConfig()
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	var findings []AuditFinding

	for _, layer := range config.Layers {
		layerFindings := runAuditLayer(path, layer, fset, file)
		findings = append(findings, layerFindings...)
	}

	return findings, nil
}

// runAuditLayer runs a single audit layer.
func runAuditLayer(path string, layer AuditLayer, fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	switch layer {
	case LayerFormatting:
		findings = auditFormatting(fset, file)
	case LayerSyntax:
		findings = auditSyntax(fset, file)
	case LayerTypeChecking:
		findings = auditTypeChecking(fset, file)
	case LayerSecurity:
		findings = auditSecurity(fset, file)
	case LayerComplexity:
		findings = auditComplexity(fset, file)
	case LayerCodeSmells:
		findings = auditCodeSmells(fset, file)
	case LayerArchitecture:
		findings = auditArchitecture(fset, file)
	case LayerRedundancy:
		findings = auditRedundancy(fset, file)
	case LayerSpecConformance:
		findings = auditSpecConformance(fset, file)
	}

	// Set layer info on all findings
	layerName := getLayerName(layer)
	for i := range findings {
		findings[i].Layer = layer
		findings[i].LayerName = layerName
		findings[i].File = path
	}

	return findings
}

func getLayerName(layer AuditLayer) string {
	for _, def := range AuditLayerDefinitions {
		if def.Layer == layer {
			return def.Name
		}
	}
	return "Unknown"
}

// Layer 1: Formatting Audit
func auditFormatting(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	// Line length check (max 120 chars)
	for _, comment := range file.Comments {
		line := fset.Position(comment.Pos()).Line
		text := comment.Text()
		if len(text) > 120 {
			findings = append(findings, AuditFinding{
				Line:     line,
				Rule:     "line-too-long",
				Message:  fmt.Sprintf("Comment line %d exceeds 120 characters", len(text)),
				Severity: "low",
			})
		}
	}

	// Naming convention checks
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name != nil {
			name := funcDecl.Name.Name
			// Check for snake_case (should be PascalCase for exported)
			if ast.IsExported(name) && strings.Contains(name, "_") {
				findings = append(findings, AuditFinding{
					Line:     fset.Position(funcDecl.Pos()).Line,
					Rule:     "snake-case-exported",
					Message:  fmt.Sprintf("Exported function %s uses snake_case; use PascalCase", name),
					Severity: "low",
				})
			}
			// Check for single-letter names (except receivers)
			if len(name) == 1 && name != "_" {
				findings = append(findings, AuditFinding{
					Line:     fset.Position(funcDecl.Pos()).Line,
					Rule:     "single-letter-name",
					Message:  fmt.Sprintf("Function %s has a single-letter name", name),
					Severity: "low",
				})
			}
		}
	}

	return findings
}

// Layer 2: Syntax Audit
func auditSyntax(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	// Check for empty interfaces (should use interface{})
	for _, decl := range file.Decls {
		typeSpec, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range typeSpec.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				if iface, ok := ts.Type.(*ast.InterfaceType); ok {
					if len(iface.Methods.List) == 0 {
						findings = append(findings, AuditFinding{
							Line:     fset.Position(ts.Pos()).Line,
							Rule:     "empty-interface",
							Message:  fmt.Sprintf("Type %s has empty interface; use interface{} directly", ts.Name.Name),
							Severity: "medium",
						})
					}
				}
			}
		}
	}

	// Check for unreachable code (after return/continue/break without label)
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		var prevEnd token.Pos
		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.ReturnStmt:
				prevEnd = stmt.End()
			case *ast.BranchStmt:
				if stmt.Tok == token.CONTINUE || stmt.Tok == token.BREAK {
					prevEnd = stmt.End()
				}
			default:
				if prevEnd.IsValid() && stmt.Pos() > prevEnd {
					findings = append(findings, AuditFinding{
						Line:     fset.Position(stmt.Pos()).Line,
						Rule:     "unreachable-code",
						Message:  "Unreachable code detected after control flow statement",
						Severity: "medium",
					})
					prevEnd = token.NoPos
				}
			}
			return true
		})
	}

	return findings
}

// Layer 3: Type Checking Audit
func auditTypeChecking(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	// Check for unsafe.Pointer conversions
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Pointer" && strings.Contains(fset.Position(call.Pos()).Filename, "unsafe") {
					findings = append(findings, AuditFinding{
						Line:     fset.Position(call.Pos()).Line,
						Rule:     "unsafe-pointer",
						Message:  "Use of unsafe.Pointer detected - verify memory safety",
						Severity: "medium",
					})
				}
			}
		}
		return true
	})

	return findings
}

// Layer 4: Security Audit (delegates to existing patterns)
func auditSecurity(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	patterns := []struct {
		name string
		fn   func(*token.FileSet, *ast.File) []ASTFinding
	}{
		{"sql-injection", detectSQLInjection},
		{"command-injection", detectCommandInjection},
		{"path-traversal", detectPathTraversal},
		{"unsafe-deserialization", detectUnsafeDeserialization},
		{"hardcoded-credentials", detectHardcodedCredentials},
	}

	for _, p := range patterns {
		pFindings := p.fn(fset, file)
		for _, f := range pFindings {
			findings = append(findings, AuditFinding{
				Line:     f.Line,
				Rule:     p.name,
				Message:  f.Message,
				Severity: f.Severity,
			})
		}
	}

	return findings
}

// Layer 5: Complexity Audit (delegates to existing patterns)
func auditComplexity(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	patterns := []struct {
		name string
		fn   func(*token.FileSet, *ast.File) []ASTFinding
	}{
		{"high-cyclomatic-complexity", detectHighCyclomaticComplexity},
		{"high-cognitive-complexity", detectHighCognitiveComplexity},
		{"high-halstead-difficulty", detectHighHalsteadDifficulty},
		{"long-function", detectLongFunction},
	}

	for _, p := range patterns {
		pFindings := p.fn(fset, file)
		for _, f := range pFindings {
			findings = append(findings, AuditFinding{
				Line:     f.Line,
				Rule:     f.Rule,
				Message:  f.Message,
				Severity: f.Severity,
			})
		}
	}

	return findings
}

// Layer 6: Code Smells Audit
func auditCodeSmells(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	patterns := []struct {
		name string
		fn   func(*token.FileSet, *ast.File) []ASTFinding
	}{
		{"semantic-redundancy", detectSemanticRedundancy},
		{"inconsistent-boolean-naming", detectInconsistentBooleanNaming},
		{"poorly-named-identifier", detectPoorlyNamedIdentifier},
	}

	for _, p := range patterns {
		pFindings := p.fn(fset, file)
		for _, f := range pFindings {
			findings = append(findings, AuditFinding{
				Line:     f.Line,
				Rule:     f.Rule,
				Message:  f.Message,
				Severity: f.Severity,
			})
		}
	}

	return findings
}

// Layer 7: Architecture Audit
func auditArchitecture(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	patterns := []struct {
		name string
		fn   func(*token.FileSet, *ast.File) []ASTFinding
	}{
		{"wrapper-chain-too-deep", detectDeepWrapperChain},
		{"duplicate-public-api", detectDuplicatePublicAPI},
		{"multiple-responsibilities", detectMultipleResponsibilities},
		{"abstraction-depth-exceeded", detectExcessiveAbstractionDepth},
		{"duplicate-configuration", detectDuplicateConfiguration},
	}

	for _, p := range patterns {
		pFindings := p.fn(fset, file)
		for _, f := range pFindings {
			findings = append(findings, AuditFinding{
				Line:     f.Line,
				Rule:     f.Rule,
				Message:  f.Message,
				Severity: f.Severity,
			})
		}
	}

	return findings
}

// Layer 8: Redundancy Audit
func auditRedundancy(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	cloneFindings := detectClones(fset, file)
	for _, f := range cloneFindings {
		findings = append(findings, AuditFinding{
			Line:     f.Line,
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: f.Severity,
		})
	}

	return findings
}

// Layer 9: Spec Conformance Audit
func auditSpecConformance(fset *token.FileSet, file *ast.File) []AuditFinding {
	var findings []AuditFinding

	// Check for TODO comments that should be addressed
	todoPattern := regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX)\b`)
	for _, comment := range file.Comments {
		if todoPattern.MatchString(comment.Text()) {
			findings = append(findings, AuditFinding{
				Line:     fset.Position(comment.Pos()).Line,
				Rule:     "unresolved-todo",
				Message:  fmt.Sprintf("Unresolved TODO/FIXME comment: %s", strings.TrimSpace(comment.Text())),
				Severity: "low",
			})
		}
	}

	return findings
}

// PrintAuditReport prints a formatted audit report.
func PrintAuditReport(findings []AuditFinding) string {
	if len(findings) == 0 {
		return "No issues found in audit."
	}

	var sb strings.Builder

	// Group by layer
	byLayer := make(map[AuditLayer][]AuditFinding)
	for _, f := range findings {
		byLayer[f.Layer] = append(byLayer[f.Layer], f)
	}

	sb.WriteString("\n=== AUDIT REPORT ===\n\n")

	for _, def := range AuditLayerDefinitions {
		if layerFindings, ok := byLayer[def.Layer]; ok {
			sb.WriteString(fmt.Sprintf("--- %s ---\n", def.Name))
			for _, f := range layerFindings {
				sb.WriteString(fmt.Sprintf("  [%s:%d] %s: %s\n", f.File, f.Line, f.Rule, f.Message))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// GetAuditSummary returns a summary of findings by severity.
func GetAuditSummary(findings []AuditFinding) map[string]int {
	summary := make(map[string]int)
	for _, f := range findings {
		summary[f.Severity]++
	}
	return summary
}
