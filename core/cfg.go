package core

import (
	"fmt"
	"go/ast"
	"go/token"
)

// CFG represents a control flow graph for a function.
type CFG struct {
	FuncName    string
	Blocks      []*CFGBlock
	EntryBlock  *CFGBlock
	ExitBlock   *CFGBlock
	Acquired    []*ResourceAcquisition // Resources acquired in this function
	Transactions []*Transaction       // Transactions started in this function
}

// CFGBlock represents a basic block in the CFG.
type CFGBlock struct {
	ID      int
	Stmts   []ast.Stmt
	Succs   []*CFGBlock // Successor blocks
	Preds   []*CFGBlock // Predecessor blocks
	Terminates bool     // Block ends with return/panic
}

// ResourceAcquisition represents a resource that needs release.
type ResourceAcquisition struct {
	AcquiredAt token.Pos
	AcquiredBy string  // Function name that acquired (e.g., "Lock", "Open")
	Variable   string  // Variable name holding the resource
	MustRelease string // Function name that releases (e.g., "Unlock", "Close")
	Line       int
}

// Transaction represents a transaction that needs commit/rollback.
type Transaction struct {
	BeginAt    token.Pos
	BeginBy    string  // Function name that began (e.g., "Begin")
	Variable   string  // Variable name holding the transaction
	MustEnd    string  // How it must end ("Commit", "Rollback")
	Line       int
}

// ProofObligation represents an obligation that must be satisfied.
type ProofObligation struct {
	ObligationType string // "MustRelease", "MustClose", "MustCommit"
	CreatedAt      token.Pos
	AcquiredBy     string
	Variable       string
	Satisfied      bool
	SatisfiedAt    token.Pos
	Line           int
}

// BuildCFG builds a control flow graph for a function.
func BuildCFG(funcDecl *ast.FuncDecl) *CFG {
	cfg := &CFG{
		FuncName: funcDecl.Name.Name,
		Blocks:   []*CFGBlock{},
	}

	if funcDecl.Body == nil {
		return cfg
	}

	// Create entry block
	entry := &CFGBlock{ID: 0, Stmts: []ast.Stmt{}}
	cfg.Blocks = append(cfg.Blocks, entry)
	cfg.EntryBlock = entry

	// Build CFG by traversing the AST
	blocks := buildBlocksFromStmt(funcDecl.Body, cfg)

	// Connect blocks sequentially
	for i := 0; i < len(blocks)-1; i++ {
		blocks[i].Succs = append(blocks[i].Succs, blocks[i+1])
		blocks[i+1].Preds = append(blocks[i+1].Preds, blocks[i])
	}

	// Mark blocks that terminate
	if len(blocks) > 0 {
		last := blocks[len(blocks)-1]
		if terminates(last) {
			last.Terminates = true
		}
		cfg.ExitBlock = last
	}

	// Track resource acquisitions and transactions
	trackResourceAcquisitions(funcDecl.Body, cfg)

	return cfg
}

// buildBlocksFromStmt converts a statement into a list of basic blocks.
func buildBlocksFromStmt(stmt ast.Stmt, cfg *CFG) []*CFGBlock {
	var blocks []*CFGBlock

	ast.Inspect(stmt, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.IfStmt:
			// Split into condition block and body blocks
			condBlock := &CFGBlock{ID: len(cfg.Blocks), Stmts: []ast.Stmt{s}}
			cfg.Blocks = append(cfg.Blocks, condBlock)
			blocks = append(blocks, condBlock)

			// Process body and else
			if s.Body != nil {
				bodyBlocks := buildBlocksFromStmt(s.Body, cfg)
				blocks = append(blocks, bodyBlocks...)
			}
			if s.Else != nil {
				elseBlocks := buildBlocksFromStmt(s.Else, cfg)
				blocks = append(blocks, elseBlocks...)
			}

			return false

		case *ast.ForStmt:
			loopBlock := &CFGBlock{ID: len(cfg.Blocks), Stmts: []ast.Stmt{s}}
			cfg.Blocks = append(cfg.Blocks, loopBlock)
			blocks = append(blocks, loopBlock)

			if s.Body != nil {
				bodyBlocks := buildBlocksFromStmt(s.Body, cfg)
				blocks = append(blocks, bodyBlocks...)
			}

			return false

		case *ast.RangeStmt:
			rangeBlock := &CFGBlock{ID: len(cfg.Blocks), Stmts: []ast.Stmt{s}}
			cfg.Blocks = append(cfg.Blocks, rangeBlock)
			blocks = append(blocks, rangeBlock)

			if s.Body != nil {
				bodyBlocks := buildBlocksFromStmt(s.Body, cfg)
				blocks = append(blocks, bodyBlocks...)
			}

			return false

		case *ast.SwitchStmt:
			switchBlock := &CFGBlock{ID: len(cfg.Blocks), Stmts: []ast.Stmt{s}}
			cfg.Blocks = append(cfg.Blocks, switchBlock)
			blocks = append(blocks, switchBlock)
			return false

		case *ast.ReturnStmt:
			retBlock := &CFGBlock{ID: len(cfg.Blocks), Stmts: []ast.Stmt{s}, Terminates: true}
			cfg.Blocks = append(cfg.Blocks, retBlock)
			blocks = append(blocks, retBlock)
			return false
		}
		return true
	})

	if len(blocks) == 0 {
		// No control flow statements found, create a single block
		block := &CFGBlock{ID: len(cfg.Blocks), Stmts: []ast.Stmt{stmt}}
		cfg.Blocks = append(cfg.Blocks, block)
		blocks = append(blocks, block)
	}

	return blocks
}

// terminates checks if a block ends with a terminating statement.
func terminates(block *CFGBlock) bool {
	if len(block.Stmts) == 0 {
		return false
	}
	last := block.Stmts[len(block.Stmts)-1]
	switch last.(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	}
	return false
}

// trackResourceAcquisitions finds resource acquisitions and transactions in a function.
func trackResourceAcquisitions(stmt ast.Stmt, cfg *CFG) {
	ast.Inspect(stmt, func(n ast.Node) bool {
		switch call := n.(type) {
		case *ast.CallExpr:
			name := getCallName(call)
			methodName := getMethodName(call)

			// Check for lock acquisitions (Lock, RLock, WLock take no args)
			for _, acquireFunc := range []string{"Lock", "RLock", "WLock"} {
				if name == acquireFunc || name == "sync."+acquireFunc || methodName == acquireFunc {
					cfg.Acquired = append(cfg.Acquired, &ResourceAcquisition{
						AcquiredAt:  call.Pos(),
						AcquiredBy:  acquireFunc,
						MustRelease: getReleaseFunc(acquireFunc),
						Line:        0,
					})
				}
			}

			// Check for Open (file operations)
			if name == "Open" || name == "os.Open" || name == "os.Create" {
				cfg.Acquired = append(cfg.Acquired, &ResourceAcquisition{
					AcquiredAt:  call.Pos(),
					AcquiredBy:  "Open",
					MustRelease: "Close",
					Line:        0,
				})
			}

			// Check for transaction Begin
			for _, beginFunc := range []string{"Begin", "BeginTx"} {
				if name == beginFunc || name == "db."+beginFunc || name == "sql."+beginFunc {
					cfg.Transactions = append(cfg.Transactions, &Transaction{
						BeginAt: call.Pos(),
						BeginBy: beginFunc,
						MustEnd: "Commit/Rollback",
						Line:    0,
					})
				}
			}

		case *ast.AssignStmt:
			// Check for assignments like: f, err := os.Open(...)
			// 'call' is already *ast.AssignStmt in this case
			if len(call.Rhs) > 0 {
				if callExpr, ok := call.Rhs[0].(*ast.CallExpr); ok {
					name := getCallName(callExpr)
					if name == "Open" || name == "os.Open" || name == "Create" || name == "os.Create" {
						// Only add acquisition for first lhs (the resource handle)
						// e.g., f, err := os.Open() - we only track f
						if len(call.Lhs) > 0 {
							if ident, ok := call.Lhs[0].(*ast.Ident); ok {
								cfg.Acquired = append(cfg.Acquired, &ResourceAcquisition{
									AcquiredAt:  call.Pos(),
									AcquiredBy:  "Open",
									Variable:    ident.Name,
									MustRelease: "Close",
									Line:        0,
								})
							}
						}
					}
				}
			}
		}
		return true
	})
}

// getMethodName returns just the method name from a call expression (e.g., "Lock" from "mu.Lock").
func getMethodName(call *ast.CallExpr) string {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name
	}
	return ""
}

// getReleaseFunc returns the corresponding release function for an acquire function.
func getReleaseFunc(acquire string) string {
	switch acquire {
	case "Lock":
		return "Unlock"
	case "RLock":
		return "RUnlock"
	case "WLock":
		return "WUnlock"
	default:
		return "Release"
	}
}

// CheckObligations checks if all proof obligations are satisfied.
func CheckObligations(cfg *CFG) []*ProofObligation {
	var unsatisfied []*ProofObligation

	for _, acq := range cfg.Acquired {
		// Check if there's a matching release call in all paths
		if !hasReleaseCall(cfg, acq) {
			unsatisfied = append(unsatisfied, &ProofObligation{
				ObligationType: "MustRelease",
				CreatedAt:      acq.AcquiredAt,
				AcquiredBy:     acq.AcquiredBy,
				Variable:       acq.Variable,
				Satisfied:      false,
				Line:           acq.Line,
			})
		}
	}

	for _, tx := range cfg.Transactions {
		// Check if there's a matching commit/rollback call in all paths
		if !hasEndCall(cfg, tx) {
			unsatisfied = append(unsatisfied, &ProofObligation{
				ObligationType: "MustCommit",
				CreatedAt:      tx.BeginAt,
				AcquiredBy:     tx.BeginBy,
				Variable:       tx.Variable,
				Satisfied:      false,
				Line:           tx.Line,
			})
		}
	}

	return unsatisfied
}

// hasReleaseCall checks if there's a release call in the function.
func hasReleaseCall(cfg *CFG, acq *ResourceAcquisition) bool {
	for _, block := range cfg.Blocks {
		for _, stmt := range block.Stmts {
			if hasReleaseInStmt(stmt, acq.MustRelease) {
				return true
			}
		}
	}
	return false
}

// hasReleaseInStmt checks if a statement contains a release call.
func hasReleaseInStmt(stmt ast.Stmt, releaseFunc string) bool {
	found := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if found {
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok {
			name := getCallName(call)
			// Check for: Close, mu.Close, f.Close, os.File.Close, Close()
			if name == releaseFunc ||
				name == "mu."+releaseFunc ||
				name == releaseFunc+"()" ||
				name == "."+releaseFunc {
				found = true
				return false
			}
			// Also check if the selector expression ends with .Close
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == releaseFunc {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

// hasEndCall checks if there's a commit/rollback call in the function.
func hasEndCall(cfg *CFG, tx *Transaction) bool {
	endFuncs := []string{"Commit", "Rollback", "Close"}
	for _, block := range cfg.Blocks {
		for _, stmt := range block.Stmts {
			for _, endFunc := range endFuncs {
				if hasReleaseInStmt(stmt, endFunc) {
					return true
				}
			}
		}
	}
	return false
}

// CFGFinding represents a finding from CFG analysis.
type CFGFinding struct {
	File     string
	Line     int
	Rule     string
	Message  string
	Severity string
}

// DetectLockNotReleased detects when a lock is acquired but not released on all paths.
func DetectLockNotReleased(fset *token.FileSet, file *ast.File) []CFGFinding {
	var findings []CFGFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		cfg := BuildCFG(funcDecl)
		obligations := CheckObligations(cfg)

		for _, ob := range obligations {
			if ob.ObligationType == "MustRelease" {
				findings = append(findings, CFGFinding{
					File:     fset.Position(funcDecl.Pos()).Filename,
					Line:     fset.Position(ob.CreatedAt).Line,
					Rule:     "lock-not-released",
					Message:  fmt.Sprintf("Lock acquired via %s() may not be released on all code paths", ob.AcquiredBy),
					Severity: "high",
				})
			}
		}
	}

	return findings
}

// DetectResourceLeak detects when a resource (file, connection) is opened but not closed.
func DetectResourceLeak(fset *token.FileSet, file *ast.File) []CFGFinding {
	var findings []CFGFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		cfg := BuildCFG(funcDecl)

		for _, acq := range cfg.Acquired {
			if acq.AcquiredBy == "Open" && !hasReleaseCall(cfg, acq) {
				findings = append(findings, CFGFinding{
					File:     fset.Position(funcDecl.Pos()).Filename,
					Line:     fset.Position(acq.AcquiredAt).Line,
					Rule:     "resource-leak",
					Message:  fmt.Sprintf("Resource opened via %s() may not be closed on all code paths", acq.AcquiredBy),
					Severity: "high",
				})
			}
		}
	}

	return findings
}

// DetectTransactionBug detects when a transaction is begun but not committed/rolled back.
func DetectTransactionBug(fset *token.FileSet, file *ast.File) []CFGFinding {
	var findings []CFGFinding

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}

		cfg := BuildCFG(funcDecl)

		for _, tx := range cfg.Transactions {
			if !hasEndCall(cfg, tx) {
				findings = append(findings, CFGFinding{
					File:     fset.Position(funcDecl.Pos()).Filename,
					Line:     fset.Position(tx.BeginAt).Line,
					Rule:     "transaction-not-ended",
					Message:  fmt.Sprintf("Transaction begun via %s() may not be committed or rolled back on all code paths", tx.BeginBy),
					Severity: "high",
				})
			}
		}
	}

	return findings
}
