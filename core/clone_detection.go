package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CloneType represents the classification of code clones.
// Based on Baker's taxonomy:
// Type 1: Identical code with at least one parameterization difference (e.g., variable names)
// Type 2: Syntactically identical except for variable names, type names, literals
// Type 3: Statements are copied with modifications (insertions, deletions, alterations)
// Type 4: Different code that computes the same functionality
type CloneType int

const (
	// CloneType1 is identical code with normalization differences
	CloneType1 CloneType = iota + 1
	// CloneType2 is syntactically similar with minor variations
	CloneType2
	// CloneType3 is copied with modifications
	CloneType3
	// CloneType4 is different code that computes the same functionality
	CloneType4
)

// Clone represents a detected code clone pair.
type Clone struct {
	FileA         string
	LineA         int
	FuncA         string
	FileB         string
	LineB         int
	FuncB         string
	Similarity    float64
	CloneType     CloneType
	TokenCount    int
	MatchedTokens int
}

// CloneDetectionConfig holds configuration for clone detection.
type CloneDetectionConfig struct {
	MinSimilarity float64 // Minimum similarity ratio (0.0-1.0), default 0.75
	MinTokens     int     // Minimum token count to consider for detection
	MaxDepth      int     // Maximum directory depth to scan
}

// DefaultCloneDetectionConfig returns the default configuration.
func DefaultCloneDetectionConfig() *CloneDetectionConfig {
	return &CloneDetectionConfig{
		MinSimilarity: 0.75,
		MinTokens:     20,
		MaxDepth:      10,
	}
}

// CloneDetector detects code clones using AST-based token comparison.
type CloneDetector struct {
	config *CloneDetectionConfig
}

// NewCloneDetector creates a new clone detector with the given config.
func NewCloneDetector(config *CloneDetectionConfig) *CloneDetector {
	if config == nil {
		config = DefaultCloneDetectionConfig()
	}
	return &CloneDetector{config: config}
}

// FunctionInfo holds information about a extracted function.
type FunctionInfo struct {
	Name     string
	File     string
	Line     int
	EndLine  int
	ASTNode  *ast.FuncDecl
	Tokens   []string       // Normalized token sequence
	TokenSet map[string]int // Token frequency map for comparison
}

// ExtractFunctions extracts all function declarations from a Go source file.
func (d *CloneDetector) ExtractFunctions(path string) ([]FunctionInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	var functions []FunctionInfo
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Skip test files for cleaner results (optional)
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		tokens := d.extractTokens(funcDecl)
		if len(tokens) < d.config.MinTokens {
			continue
		}

		tokenSet := d.buildTokenSet(tokens)

		functions = append(functions, FunctionInfo{
			Name:     funcDecl.Name.Name,
			File:     path,
			Line:     fset.Position(funcDecl.Pos()).Line,
			EndLine:  fset.Position(funcDecl.End()).Line,
			ASTNode:  funcDecl,
			Tokens:   tokens,
			TokenSet: tokenSet,
		})
	}

	return functions, nil
}

// extractTokens extracts and normalizes tokens from a function AST.
// This implements the normalization step similar to PMD CPD.
func (d *CloneDetector) extractTokens(funcDecl *ast.FuncDecl) []string {
	var tokens []string

	// Add function signature tokens
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		tokens = append(tokens, "RECV")
	}
	tokens = append(tokens, "FUNC", funcDecl.Name.Name)

	// Add parameter types
	for _, field := range funcDecl.Type.Params.List {
		tokens = append(tokens, d.typeToken(field.Type)...)
	}

	// Add return types
	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			tokens = append(tokens, d.typeToken(field.Type)...)
		}
	}

	// Extract tokens from body
	if funcDecl.Body != nil {
		tokens = append(tokens, d.stmtTokens(funcDecl.Body)...)
	}

	return tokens
}

// typeToken returns a normalized token representation for a type.
func (d *CloneDetector) typeToken(expr ast.Expr) []string {
	switch t := expr.(type) {
	case *ast.Ident:
		return []string{"TYPE", t.Name}
	case *ast.SelectorExpr:
		var prefix string
		if ident, ok := t.X.(*ast.Ident); ok {
			prefix = ident.Name
		}
		return []string{"TYPE", prefix + "." + t.Sel.Name}
	case *ast.StarExpr:
		return append([]string{"PTR"}, d.typeToken(t.X)...)
	case *ast.ArrayType:
		if t.Len == nil {
			return []string{"TYPE", "SLICE"}
		}
		return []string{"TYPE", "ARRAY"}
	case *ast.MapType:
		return []string{"TYPE", "MAP"}
	case *ast.InterfaceType:
		return []string{"TYPE", "INTERFACE"}
	case *ast.FuncType:
		return []string{"TYPE", "FUNC"}
	case *ast.ChanType:
		return []string{"TYPE", "CHAN"}
	default:
		return []string{"TYPE", "OTHER"}
	}
}

// stmtTokens extracts tokens from a statement, normalizing variable names.
func (d *CloneDetector) stmtTokens(stmt ast.Stmt) []string {
	var tokens []string

	// Token markers for statement types (helps identify clone type)
	switch s := stmt.(type) {
	case *ast.IfStmt:
		tokens = append(tokens, "STMT_IF")
		if s.Init != nil {
			tokens = append(tokens, d.stmtTokens(s.Init)...)
		}
		tokens = append(tokens, d.exprTokens(s.Cond)...)
		tokens = append(tokens, d.stmtTokens(s.Body)...)
		if s.Else != nil {
			tokens = append(tokens, "STMT_ELSE")
			tokens = append(tokens, d.stmtTokens(s.Else)...)
		}
	case *ast.ForStmt:
		tokens = append(tokens, "STMT_FOR")
		if s.Init != nil {
			tokens = append(tokens, d.stmtTokens(s.Init)...)
		}
		if s.Cond != nil {
			tokens = append(tokens, d.exprTokens(s.Cond)...)
		}
		if s.Post != nil {
			tokens = append(tokens, d.stmtTokens(s.Post)...)
		}
		tokens = append(tokens, d.stmtTokens(s.Body)...)
	case *ast.RangeStmt:
		tokens = append(tokens, "STMT_RANGE")
		tokens = append(tokens, d.exprTokens(s.X)...)
		tokens = append(tokens, d.stmtTokens(s.Body)...)
	case *ast.SwitchStmt:
		tokens = append(tokens, "STMT_SWITCH")
		if s.Init != nil {
			tokens = append(tokens, d.stmtTokens(s.Init)...)
		}
		if s.Tag != nil {
			tokens = append(tokens, d.exprTokens(s.Tag)...)
		}
		for _, clause := range s.Body.List {
			if ic, ok := clause.(*ast.CaseClause); ok {
				tokens = append(tokens, "CASE")
				for _, expr := range ic.List {
					tokens = append(tokens, d.exprTokens(expr)...)
				}
				for _, stmt := range ic.Body {
					tokens = append(tokens, d.stmtTokens(stmt)...)
				}
			}
		}
	case *ast.ReturnStmt:
		tokens = append(tokens, "STMT_RETURN")
		for _, result := range s.Results {
			tokens = append(tokens, d.exprTokens(result)...)
		}
	case *ast.DeferStmt:
		tokens = append(tokens, "STMT_DEFER")
		tokens = append(tokens, d.exprTokens(s.Call)...)
	case *ast.GoStmt:
		tokens = append(tokens, "STMT_GO")
		tokens = append(tokens, d.exprTokens(s.Call)...)
	case *ast.AssignStmt:
		tokens = append(tokens, "STMT_ASSIGN", fmt.Sprintf("NASSIGN_%d", len(s.Lhs)))
		for _, lhs := range s.Lhs {
			tokens = append(tokens, d.exprTokens(lhs)...)
		}
		for _, rhs := range s.Rhs {
			tokens = append(tokens, d.exprTokens(rhs)...)
		}
	case *ast.ExprStmt:
		tokens = append(tokens, d.exprTokens(s.X)...)
	case *ast.IncDecStmt:
		tokens = append(tokens, "STMT_INCDEC", s.Tok.String())
	case *ast.SendStmt:
		tokens = append(tokens, "STMT_SEND")
		tokens = append(tokens, d.exprTokens(s.Chan)...)
		tokens = append(tokens, d.exprTokens(s.Value)...)
	default:
		tokens = append(tokens, "STMT_OTHER")
	}

	return tokens
}

// exprTokens extracts tokens from an expression, normalizing names.
func (d *CloneDetector) exprTokens(expr ast.Expr) []string {
	var tokens []string

	switch e := expr.(type) {
	case *ast.Ident:
		// Normalize all variable names to VAR
		tokens = append(tokens, "VAR")
	case *ast.BasicLit:
		// Normalize literals to their type
		switch e.Kind {
		case token.INT:
			tokens = append(tokens, "LIT_INT")
		case token.FLOAT:
			tokens = append(tokens, "LIT_FLOAT")
		case token.CHAR:
			tokens = append(tokens, "LIT_CHAR")
		case token.STRING:
			tokens = append(tokens, "LIT_STRING")
		default:
			tokens = append(tokens, "LIT_OTHER")
		}
	case *ast.BinaryExpr:
		tokens = append(tokens, "EXPR_BIN", e.Op.String())
		tokens = append(tokens, d.exprTokens(e.X)...)
		tokens = append(tokens, d.exprTokens(e.Y)...)
	case *ast.UnaryExpr:
		tokens = append(tokens, "EXPR_UN", e.Op.String())
		tokens = append(tokens, d.exprTokens(e.X)...)
	case *ast.CallExpr:
		tokens = append(tokens, "CALL")
		tokens = append(tokens, d.exprTokens(e.Fun)...)
		for _, arg := range e.Args {
			tokens = append(tokens, d.exprTokens(arg)...)
		}
	case *ast.SelectorExpr:
		tokens = append(tokens, "SEL")
		tokens = append(tokens, d.exprTokens(e.X)...)
		tokens = append(tokens, "VAR") // Selector name normalized
	case *ast.IndexExpr:
		tokens = append(tokens, "EXPR_INDEX")
		tokens = append(tokens, d.exprTokens(e.X)...)
		tokens = append(tokens, d.exprTokens(e.Index)...)
	case *ast.SliceExpr:
		tokens = append(tokens, "EXPR_SLICE")
		if e.X != nil {
			tokens = append(tokens, d.exprTokens(e.X)...)
		}
	case *ast.CompositeLit:
		tokens = append(tokens, "EXPR_COMPOSITE")
		for _, elt := range e.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				tokens = append(tokens, d.exprTokens(kv.Key)...)
				tokens = append(tokens, d.exprTokens(kv.Value)...)
			} else {
				tokens = append(tokens, d.exprTokens(elt)...)
			}
		}
	case *ast.FuncLit:
		tokens = append(tokens, "FUNC_LIT")
		if e.Body != nil {
			tokens = append(tokens, d.stmtTokens(e.Body)...)
		}
	case *ast.ParenExpr:
		tokens = append(tokens, "EXPR_PAREN")
		tokens = append(tokens, d.exprTokens(e.X)...)
	case *ast.TypeAssertExpr:
		tokens = append(tokens, "EXPR_TYPEASSERT")
		tokens = append(tokens, d.exprTokens(e.X)...)
	case *ast.StarExpr:
		tokens = append(tokens, "EXPR_DEREF")
		tokens = append(tokens, d.exprTokens(e.X)...)
	default:
		tokens = append(tokens, "EXPR_OTHER")
	}

	return tokens
}

// buildTokenSet creates a frequency map of tokens for comparison.
func (d *CloneDetector) buildTokenSet(tokens []string) map[string]int {
	set := make(map[string]int)
	for _, t := range tokens {
		set[t]++
	}
	return set
}

// DetectClones scans a directory for code clones.
func (d *CloneDetector) DetectClones(dir string) ([]Clone, error) {
	var allFunctions []FunctionInfo

	// Extract functions from all Go files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		if functions, err := d.ExtractFunctions(path); err == nil {
			allFunctions = append(allFunctions, functions...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Compare all function pairs
	return d.compareFunctions(allFunctions), nil
}

// compareFunctions compares all pairs of functions to detect clones.
func (d *CloneDetector) compareFunctions(functions []FunctionInfo) []Clone {
	var clones []Clone

	// Sort for deterministic output
	sort.Slice(functions, func(i, j int) bool {
		if functions[i].File != functions[j].File {
			return functions[i].File < functions[j].File
		}
		return functions[i].Line < functions[j].Line
	})

	// Compare each pair
	for i := 0; i < len(functions); i++ {
		for j := i + 1; j < len(functions); j++ {
			if clone := d.comparePair(functions[i], functions[j]); clone != nil {
				clones = append(clones, *clone)
			}
		}
	}

	return clones
}

// comparePair compares two functions and returns a Clone if similarity exceeds threshold.
func (d *CloneDetector) comparePair(a, b FunctionInfo) *Clone {
	// Quick check: skip if files are identical (same function compared to itself)
	if a.File == b.File && a.Name == b.Name {
		return nil
	}

	// Calculate token-based similarity (Jaccard-like coefficient)
	similarity := d.tokenSimilarity(a.TokenSet, b.TokenSet)

	// Require minimum similarity
	if similarity < d.config.MinSimilarity {
		return nil
	}

	// Calculate matched token count
	matchedTokens := d.countMatchedTokens(a.Tokens, b.Tokens)

	// Determine clone type based on similarity
	cloneType := d.determineCloneType(similarity)

	return &Clone{
		FileA:         a.File,
		LineA:         a.Line,
		FuncA:         a.Name,
		FileB:         b.File,
		LineB:         b.Line,
		FuncB:         b.Name,
		Similarity:    similarity,
		CloneType:     cloneType,
		TokenCount:    len(a.Tokens),
		MatchedTokens: matchedTokens,
	}
}

// tokenSimilarity calculates similarity between two token frequency maps.
// Uses a normalized dot product approach (cosine-like).
func (d *CloneDetector) tokenSimilarity(setA, setB map[string]int) float64 {
	// Find common tokens
	var commonTokens int
	var totalTokens int

	// Count tokens in A
	for _, count := range setA {
		totalTokens += count
	}

	// Count common tokens and tokens in B
	var bTokens int
	for token, countB := range setB {
		bTokens += countB
		if countA, exists := setA[token]; exists {
			commonTokens += min(countA, countB)
		}
	}

	// Jaccard similarity on token sets
	unionTokens := len(setA) + len(setB)
	for token := range setA {
		if _, exists := setB[token]; exists {
			unionTokens-- // Remove from union count since it's common
		}
	}

	// Avoid division by zero
	if unionTokens == 0 {
		return 0.0
	}

	// Use minimum shared tokens vs total unique tokens
	sharedTokens := 0
	for token := range setA {
		if _, exists := setB[token]; exists {
			sharedTokens++
		}
	}

	// Calculate average token count similarity
	avgTokens := float64(totalTokens+bTokens) / 2.0
	if avgTokens == 0 {
		return 0.0
	}

	// Weighted similarity: token set overlap * token count similarity
	setSimilarity := float64(sharedTokens) / float64(len(setA)+len(setB)-sharedTokens)
	countSimilarity := float64(commonTokens*2) / float64(totalTokens+bTokens)

	return setSimilarity*0.6 + countSimilarity*0.4
}

// countMatchedTokens counts how many tokens match between two token sequences.
func (d *CloneDetector) countMatchedTokens(tokensA, tokensB []string) int {
	// Build a map of token positions for B
	bPositions := make(map[string][]int)
	for i, t := range tokensB {
		bPositions[t] = append(bPositions[t], i)
	}

	// Find longest common subsequence-like match
	matched := 0
	bUsed := make([]bool, len(tokensB))

	for _, t := range tokensA {
		if positions, exists := bPositions[t]; exists {
			for _, pos := range positions {
				if !bUsed[pos] {
					bUsed[pos] = true
					matched++
					break
				}
			}
		}
	}

	return matched
}

// determineCloneType classifies the clone based on similarity score.
func (d *CloneDetector) determineCloneType(similarity float64) CloneType {
	switch {
	case similarity >= 0.95:
		return CloneType1
	case similarity >= 0.85:
		return CloneType2
	case similarity >= 0.70:
		return CloneType3
	default:
		return CloneType4
	}
}

// CloneFinding represents a finding from clone detection.
type CloneFinding struct {
	FileA      string
	LineA      int
	FuncA      string
	FileB      string
	LineB      int
	FuncB      string
	Similarity float64
	CloneType  CloneType
}

// detectClones is the ASTPattern function for detecting clones.
func detectClones(fset *token.FileSet, file *ast.File) []ASTFinding {
	var findings []ASTFinding

	detector := NewCloneDetector(DefaultCloneDetectionConfig())

	// We need to work with the parsed file - extract functions
	var functions []FunctionInfo

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Name == nil {
			continue
		}

		tokens := detector.extractTokens(funcDecl)
		if len(tokens) < detector.config.MinTokens {
			continue
		}

		tokenSet := detector.buildTokenSet(tokens)

		functions = append(functions, FunctionInfo{
			Name:     funcDecl.Name.Name,
			File:     fset.Position(funcDecl.Pos()).Filename,
			Line:     fset.Position(funcDecl.Pos()).Line,
			EndLine:  fset.Position(funcDecl.End()).Line,
			ASTNode:  funcDecl,
			Tokens:   tokens,
			TokenSet: tokenSet,
		})
	}

	// Compare within the same file
	for i := 0; i < len(functions); i++ {
		for j := i + 1; j < len(functions); j++ {
			if clone := detector.comparePair(functions[i], functions[j]); clone != nil {
				findings = append(findings, ASTFinding{
					Line:     functions[i].Line,
					Rule:     "code-clone",
					Message:  fmt.Sprintf("Code clone detected: %s (line %d) is %.0f%% similar to %s (line %d) - Type %d clone", clone.FuncA, clone.LineA, clone.Similarity*100, clone.FuncB, clone.LineB, clone.CloneType),
					Severity: "medium",
				})
			}
		}
	}

	return findings
}
