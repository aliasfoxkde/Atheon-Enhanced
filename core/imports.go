package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ImportGraph represents the import relationships between files.
type ImportGraph struct {
	Nodes map[string]*ImportNode
	Edges []ImportEdge
}

// ImportNode represents a single file in the import graph.
type ImportNode struct {
	FilePath string
	Package  string
	Imports  []string // List of files this file imports
}

// ImportEdge represents an import relationship.
type ImportEdge struct {
	From string
	To   string
}

// BuildImportGraph builds an import graph for all Go files in a directory.
func BuildImportGraph(dir string) (*ImportGraph, error) {
	graph := &ImportGraph{
		Nodes: make(map[string]*ImportNode),
		Edges: []ImportEdge{},
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		node, err := parseFileImports(path)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		graph.Nodes[path] = node

		// Create edges for each import
		for _, imported := range node.Imports {
			graph.Edges = append(graph.Edges, ImportEdge{
				From: path,
				To:   imported,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return graph, nil
}

// parseFileImports parses a Go file and extracts its imports.
func parseFileImports(filePath string) (*ImportNode, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	node := &ImportNode{
		FilePath: filePath,
		Package:  file.Name.Name,
		Imports:  []string{},
	}

	// Extract imports
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		for _, spec := range genDecl.Specs {
			if importSpec, ok := spec.(*ast.ImportSpec); ok {
				if importSpec.Path != nil {
					importPath := strings.Trim(importSpec.Path.Value, "\"")
					node.Imports = append(node.Imports, importPath)
				}
			}
		}
	}

	return node, nil
}

// DetectCircularImports finds circular import chains in the graph.
func DetectCircularImports(graph *ImportGraph) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	path := []string{}
	inStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		if inStack[node] {
			// Found a cycle - extract it from the path
			cycleStart := -1
			for i, n := range path {
				if n == node {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append([]string{}, path[cycleStart:]...)
				cycle = append(cycle, node)
				cycles = append(cycles, cycle)
			}
			return true
		}

		if visited[node] {
			return false
		}

		visited[node] = true
		inStack[node] = true
		path = append(path, node)

		// Find all files this node imports
		for _, edge := range graph.Edges {
			if edge.From == node {
				dfs(edge.To)
			}
		}

		path = path[:len(path)-1]
		inStack[node] = false
		return false
	}

	// Start DFS from each node
	for node := range graph.Nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// ImportFinding represents a circular import finding.
type ImportFinding struct {
	File     string
	Line     int
	Cycle    []string
	Rule     string
	Message  string
	Severity string
}

// ScanDirForCircularImports scans a directory for circular imports.
func ScanDirForCircularImports(dir string) ([]ImportFinding, error) {
	graph, err := BuildImportGraph(dir)
	if err != nil {
		return nil, err
	}

	var findings []ImportFinding

	// Simple cycle detection - check for direct A->B->A cycles
	for _, edge := range graph.Edges {
		// Check if there's a path back from edge.To to edge.From
		if hasPath(edge.To, edge.From, graph) {
			findings = append(findings, ImportFinding{
				File:     edge.From,
				Line:     0,
				Cycle:    []string{edge.From, edge.To, edge.From},
				Rule:     "circular-import",
				Message:  fmt.Sprintf("Circular import detected: %s imports %s which imports %s", filepath.Base(edge.From), filepath.Base(edge.To), filepath.Base(edge.From)),
				Severity: "high",
			})
		}
	}

	return findings, nil
}

// hasPath checks if there's a path from source to target in the graph.
func hasPath(source, target string, graph *ImportGraph) bool {
	visited := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		if node == target {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true

		for _, edge := range graph.Edges {
			if edge.From == node {
				if dfs(edge.To) {
					return true
				}
			}
		}
		return false
	}

	return dfs(source)
}
