package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

// Parser wraps the go/ast and go/parser packages.
type Parser struct {
	fset *token.FileSet
}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

// ParseFile parses a Go source file and returns its AST.
func (p *Parser) ParseFile(filename string) (*ast.File, error) {
	return parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
}

// ParseDir parses all Go source files in a directory and returns a map
// of filename to AST.
func (p *Parser) ParseDir(dir string) (map[string]*ast.File, error) {
	files := make(map[string]*ast.File)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		f, err := parser.ParseFile(p.fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		files[path] = f
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// ParseContent parses Go source from a string and returns its AST.
func (p *Parser) ParseContent(filename string, content string) (*ast.File, error) {
	return parser.ParseFile(p.fset, filename, content, parser.ParseComments)
}
