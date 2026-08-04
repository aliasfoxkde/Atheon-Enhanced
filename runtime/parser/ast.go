package parser

import (
	"fmt"
	"go/ast"
)

// WalkAST traverses the AST tree and calls the visitor function for each node.
// If the visitor returns false, traversal stops.
func WalkAST(node ast.Node, visitor func(ast.Node) bool) {
	if node == nil {
		return
	}

	if !visitor(node) {
		return
	}

	switch n := node.(type) {
	case *ast.File:
		if n == nil {
			return
		}
		for _, decl := range n.Decls {
			WalkAST(decl, visitor)
		}
		for _, comment := range n.Comments {
			WalkAST(comment, visitor)
		}

	case *ast.GenDecl:
		if n == nil {
			return
		}
		for _, spec := range n.Specs {
			WalkAST(spec, visitor)
		}

	case *ast.ValueSpec:
		if n == nil {
			return
		}
		for _, name := range n.Names {
			WalkAST(name, visitor)
		}
		WalkAST(n.Type, visitor)
		for _, value := range n.Values {
			WalkAST(value, visitor)
		}

	case *ast.ImportSpec:
		if n == nil {
			return
		}
		WalkAST(n.Name, visitor)
		WalkAST(n.Path, visitor)

	case *ast.TypeSpec:
		if n == nil {
			return
		}
		WalkAST(n.Name, visitor)
		WalkAST(n.Type, visitor)

	case *ast.FuncDecl:
		if n.Name != nil {
			WalkAST(n.Name, visitor)
		}
		if n.Type != nil {
			WalkAST(n.Type, visitor)
		}
		if n.Body != nil {
			WalkAST(n.Body, visitor)
		}

	case *ast.BlockStmt:
		for _, stmt := range n.List {
			WalkAST(stmt, visitor)
		}

	case *ast.IfStmt:
		WalkAST(n.Init, visitor)
		WalkAST(n.Cond, visitor)
		WalkAST(n.Body, visitor)
		WalkAST(n.Else, visitor)

	case *ast.ForStmt:
		WalkAST(n.Init, visitor)
		WalkAST(n.Cond, visitor)
		WalkAST(n.Post, visitor)
		WalkAST(n.Body, visitor)

	case *ast.RangeStmt:
		WalkAST(n.X, visitor)
		for _, stmt := range n.Body.List {
			WalkAST(stmt, visitor)
		}

	case *ast.SwitchStmt:
		WalkAST(n.Init, visitor)
		WalkAST(n.Tag, visitor)
		WalkAST(n.Body, visitor)

	case *ast.SelectStmt:
		for _, clause := range n.Body.List {
			WalkAST(clause, visitor)
		}

	case *ast.TypeSwitchStmt:
		WalkAST(n.Init, visitor)
		WalkAST(n.Assign, visitor)
		WalkAST(n.Body, visitor)

	case *ast.CommClause:
		WalkAST(n.Comm, visitor)
		for _, stmt := range n.Body {
			WalkAST(stmt, visitor)
		}

	case *ast.CaseClause:
		for _, expr := range n.List {
			WalkAST(expr, visitor)
		}
		for _, stmt := range n.Body {
			WalkAST(stmt, visitor)
		}

	case *ast.AssignStmt:
		for _, lhs := range n.Lhs {
			WalkAST(lhs, visitor)
		}
		for _, rhs := range n.Rhs {
			WalkAST(rhs, visitor)
		}

	case *ast.CallExpr:
		WalkAST(n.Fun, visitor)
		for _, arg := range n.Args {
			WalkAST(arg, visitor)
		}

	case *ast.ReturnStmt:
		for _, result := range n.Results {
			WalkAST(result, visitor)
		}

	case *ast.ExprStmt:
		WalkAST(n.X, visitor)

	case *ast.SendStmt:
		WalkAST(n.Chan, visitor)
		WalkAST(n.Value, visitor)

	case *ast.IncDecStmt:
		WalkAST(n.X, visitor)

	case *ast.UnaryExpr:
		WalkAST(n.X, visitor)

	case *ast.BinaryExpr:
		WalkAST(n.X, visitor)
		WalkAST(n.Y, visitor)

	case *ast.ParenExpr:
		WalkAST(n.X, visitor)

	case *ast.SelectorExpr:
		WalkAST(n.X, visitor)
		WalkAST(n.Sel, visitor)

	case *ast.IndexExpr:
		WalkAST(n.X, visitor)
		WalkAST(n.Index, visitor)

	case *ast.IndexListExpr:
		WalkAST(n.X, visitor)
		for _, index := range n.Indices {
			WalkAST(index, visitor)
		}

	case *ast.SliceExpr:
		WalkAST(n.X, visitor)
		WalkAST(n.Low, visitor)
		WalkAST(n.High, visitor)
		WalkAST(n.Max, visitor)

	case *ast.TypeAssertExpr:
		WalkAST(n.X, visitor)
		WalkAST(n.Type, visitor)

	case *ast.StarExpr:
		WalkAST(n.X, visitor)

	case *ast.KeyValueExpr:
		WalkAST(n.Key, visitor)
		WalkAST(n.Value, visitor)

	case *ast.CompositeLit:
		WalkAST(n.Type, visitor)
		for _, elts := range n.Elts {
			WalkAST(elts, visitor)
		}

	case *ast.FuncLit:
		WalkAST(n.Type, visitor)
		WalkAST(n.Body, visitor)

	case *ast.FuncType:
		WalkAST(n.Params, visitor)
		WalkAST(n.Results, visitor)

	case *ast.StructType:
		WalkAST(n.Fields, visitor)

	case *ast.InterfaceType:
		WalkAST(n.Methods, visitor)

	case *ast.MapType:
		WalkAST(n.Key, visitor)
		WalkAST(n.Value, visitor)

	case *ast.ChanType:
		WalkAST(n.Value, visitor)

	case *ast.BasicLit:
		// No children

	case *ast.Ident:
		// No children

	case *ast.Ellipsis:
		WalkAST(n.Elt, visitor)

	case *ast.Package:
		for _, file := range n.Files {
			WalkAST(file, visitor)
		}

	case *ast.CommentGroup:
		for _, comment := range n.List {
			WalkAST(comment, visitor)
		}

	case *ast.Field:
		if n == nil {
			return
		}
		for _, name := range n.Names {
			WalkAST(name, visitor)
		}
		WalkAST(n.Type, visitor)
		WalkAST(n.Tag, visitor)

	case *ast.FieldList:
		if n == nil {
			return
		}
		if n.List != nil {
			for _, field := range n.List {
				WalkAST(field, visitor)
			}
		}

	case *ast.LabeledStmt:
		WalkAST(n.Label, visitor)
		WalkAST(n.Stmt, visitor)

	case *ast.GoStmt:
		WalkAST(n.Call, visitor)

	case *ast.DeferStmt:
		WalkAST(n.Call, visitor)

	case *ast.EmptyStmt:
		// No children

	case *ast.BranchStmt:
		// No children

	case *ast.DeclStmt:
		WalkAST(n.Decl, visitor)

	case *ast.BadStmt:
		// No children

	case *ast.BadExpr:
		// No children

	case *ast.BadDecl:
		// No children

	case *ast.Comment:
		// No children

	case *ast.ArrayType:
		WalkAST(n.Len, visitor)
		WalkAST(n.Elt, visitor)
	}
}

// FindNodes returns all nodes in the AST that match the predicate.
func FindNodes(node ast.Node, predicate func(ast.Node) bool) []ast.Node {
	var result []ast.Node
	WalkAST(node, func(n ast.Node) bool {
		if predicate(n) {
			result = append(result, n)
		}
		return true
	})
	return result
}

// GetNodeKind returns a string describing the kind of the node.
func GetNodeKind(node ast.Node) string {
	if node == nil {
		return "nil"
	}

	switch node.(type) {
	case *ast.File:
		return "File"
	case *ast.Package:
		return "Package"
	case *ast.TypeSpec:
		return "TypeSpec"
	case *ast.ValueSpec:
		return "ValueSpec"
	case *ast.ImportSpec:
		return "ImportSpec"
	case *ast.GenDecl:
		return "GenDecl"
	case *ast.FuncDecl:
		return "FuncDecl"
	case *ast.FuncLit:
		return "FuncLit"
	case *ast.FuncType:
		return "FuncType"
	case *ast.Ident:
		return "Ident"
	case *ast.BasicLit:
		return "BasicLit"
	case *ast.BinaryExpr:
		return "BinaryExpr"
	case *ast.UnaryExpr:
		return "UnaryExpr"
	case *ast.CallExpr:
		return "CallExpr"
	case *ast.AssignStmt:
		return "AssignStmt"
	case *ast.ReturnStmt:
		return "ReturnStmt"
	case *ast.IfStmt:
		return "IfStmt"
	case *ast.ForStmt:
		return "ForStmt"
	case *ast.RangeStmt:
		return "RangeStmt"
	case *ast.SwitchStmt:
		return "SwitchStmt"
	case *ast.SelectStmt:
		return "SelectStmt"
	case *ast.TypeSwitchStmt:
		return "TypeSwitchStmt"
	case *ast.CaseClause:
		return "CaseClause"
	case *ast.CommClause:
		return "CommClause"
	case *ast.BlockStmt:
		return "BlockStmt"
	case *ast.ExprStmt:
		return "ExprStmt"
	case *ast.SendStmt:
		return "SendStmt"
	case *ast.IncDecStmt:
		return "IncDecStmt"
	case *ast.StarExpr:
		return "StarExpr"
	case *ast.SelectorExpr:
		return "SelectorExpr"
	case *ast.IndexExpr:
		return "IndexExpr"
	case *ast.IndexListExpr:
		return "IndexListExpr"
	case *ast.SliceExpr:
		return "SliceExpr"
	case *ast.TypeAssertExpr:
		return "TypeAssertExpr"
	case *ast.ParenExpr:
		return "ParenExpr"
	case *ast.KeyValueExpr:
		return "KeyValueExpr"
	case *ast.CompositeLit:
		return "CompositeLit"
	case *ast.StructType:
		return "StructType"
	case *ast.InterfaceType:
		return "InterfaceType"
	case *ast.MapType:
		return "MapType"
	case *ast.ChanType:
		return "ChanType"
	case *ast.ArrayType:
		return "ArrayType"
	case *ast.Ellipsis:
		return "Ellipsis"
	case *ast.Field:
		return "Field"
	case *ast.FieldList:
		return "FieldList"
	case *ast.LabeledStmt:
		return "LabeledStmt"
	case *ast.GoStmt:
		return "GoStmt"
	case *ast.DeferStmt:
		return "DeferStmt"
	case *ast.EmptyStmt:
		return "EmptyStmt"
	case *ast.BranchStmt:
		return "BranchStmt"
	case *ast.DeclStmt:
		return "DeclStmt"
	case *ast.BadStmt:
		return "BadStmt"
	case *ast.BadExpr:
		return "BadExpr"
	case *ast.BadDecl:
		return "BadDecl"
	case *ast.Comment:
		return "Comment"
	case *ast.CommentGroup:
		return "CommentGroup"
	default:
		return fmt.Sprintf("Unknown(%T)", node)
	}
}
