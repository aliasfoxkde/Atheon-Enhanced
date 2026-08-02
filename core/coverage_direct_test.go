package core

import (
	"go/ast"
	"go/token"
	"testing"
)

// TestCalculateCognitiveComplexity_Direct directly tests calculateCognitiveComplexity
// with all statement types to hit the uncovered SwitchStmt, CaseClause, and ReturnStmt branches.
func TestCalculateCognitiveComplexity_Direct(t *testing.T) {
	// Direct AST construction to test specific statement types

	// Test SwitchStmt branch
	switchBody := &ast.SwitchStmt{
		Tag: &ast.Ident{Name: "x"},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.CaseClause{
					List: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
					Body: []ast.Stmt{
						&ast.ReturnStmt{Results: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}}},
					},
				},
			},
		},
	}
	_ = calculateCognitiveComplexity(switchBody, 0)

	// Test ReturnStmt branch directly
	returnStmt := &ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "x"}}}
	_ = calculateCognitiveComplexity(returnStmt, 0)

	// Test CaseClause branch
	caseClause := &ast.CaseClause{
		List: []ast.Expr{&ast.Ident{Name: "x"}},
		Body: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}}}},
	}
	_ = calculateCognitiveComplexity(caseClause, 0)

	// Test with ForStmt
	forStmt := &ast.ForStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "i"}}},
			},
		},
	}
	_ = calculateCognitiveComplexity(forStmt, 0)

	// Test with RangeStmt
	rangeStmt := &ast.RangeStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "v"}}},
			},
		},
	}
	_ = calculateCognitiveComplexity(rangeStmt, 0)

	// Test ExprStmt with recursive call
	exprStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.Ident{Name: "recursive"},
		},
	}
	_ = calculateCognitiveComplexity(exprStmt, 0)
}

// TestTypeToStringBranch_Direct tests typeToString with uncovered branches.
func TestTypeToStringBranch_Direct(t *testing.T) {
	// Test with ChanType to hit default case
	chanType := &ast.ChanType{
		Value: &ast.Ident{Name: "int"},
	}
	result := typeToString(chanType)
	if result != "unknown" {
		t.Errorf("expected 'unknown' for ChanType, got %s", result)
	}

	// Test SelectorExpr where X is not an Ident
	selectorExpr := &ast.SelectorExpr{
		X: &ast.BasicLit{Kind: token.STRING, Value: `"test"`}, // X is not an Ident
		Sel: &ast.Ident{Name: "field"},
	}
	result = typeToString(selectorExpr)
	if result != "field" {
		t.Errorf("expected 'field' for selector with non-Ident X, got %s", result)
	}
}

// TestHasConditionalReturnPath_Direct tests hasConditionalReturnPath with all statement types.
func TestHasConditionalReturnPath_Direct(t *testing.T) {
	// Test ForStmt with return in body
	forStmt := &ast.ForStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "i"}}},
			},
		},
	}
	result := hasConditionalReturnPath(forStmt)
	if !result {
		t.Error("expected true for ForStmt with return")
	}

	// Test RangeStmt with return in body
	rangeStmt := &ast.RangeStmt{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "v"}}},
			},
		},
	}
	result = hasConditionalReturnPath(rangeStmt)
	if !result {
		t.Error("expected true for RangeStmt with return")
	}

	// Test SwitchStmt with return in case
	switchStmt := &ast.SwitchStmt{
		Tag: &ast.Ident{Name: "x"},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.CaseClause{
					Body: []ast.Stmt{
						&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "1"}}},
					},
				},
			},
		},
	}
	result = hasConditionalReturnPath(switchStmt)
	if !result {
		t.Error("expected true for SwitchStmt with return in case")
	}
}

// TestCountInterfaceDepth_Direct tests countInterfaceDepth with sibling interfaces at same depth.
func TestCountInterfaceDepth_Direct(t *testing.T) {
	// Create interface with sibling embedded interfaces at same depth
	iface := &ast.InterfaceType{
		Methods: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{},
					},
				},
				{
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{},
					},
				},
			},
		},
	}
	depth := countInterfaceDepth(iface)
	if depth != 1 {
		t.Errorf("expected depth 1, got %d", depth)
	}
}
