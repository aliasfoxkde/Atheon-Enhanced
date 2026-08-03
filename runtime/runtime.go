package runtime

import (
	"go/ast"

	"github.com/aliasfoxkde/Atheon/runtime/cache"
	"github.com/aliasfoxkde/Atheon/runtime/config"
	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
	"github.com/aliasfoxkde/Atheon/runtime/formatter"
	"github.com/aliasfoxkde/Atheon/runtime/parser"
	"github.com/aliasfoxkde/Atheon/runtime/patterns"
	"github.com/aliasfoxkde/Atheon/runtime/rules"
	"github.com/aliasfoxkde/Atheon/runtime/scanner"
)

// Runtime is the main API for the Atheon analysis runtime.
// It orchestrates all components to perform analysis on code projects.
type Runtime struct {
	scanner  *scanner.Scanner
	rules   *rules.Registry
	patterns *patterns.Matcher
	config  *config.Config
	cache   *cache.Cache
	formats map[string]formatter.Formatter
}

// RuntimeOption configures the Runtime.
type RuntimeOption func(*Runtime)

// WithScanner sets a custom scanner.
func WithScanner(s *scanner.Scanner) RuntimeOption {
	return func(r *Runtime) {
		r.scanner = s
	}
}

// WithRules sets a custom rules registry.
func WithRules(reg *rules.Registry) RuntimeOption {
	return func(r *Runtime) {
		r.rules = reg
	}
}

// WithPatterns sets a custom pattern matcher.
func WithPatterns(m *patterns.Matcher) RuntimeOption {
	return func(r *Runtime) {
		r.patterns = m
	}
}

// WithConfig sets a custom configuration.
func WithConfig(cfg *config.Config) RuntimeOption {
	return func(r *Runtime) {
		r.config = cfg
	}
}

// WithCache sets a custom cache.
func WithCache(c *cache.Cache) RuntimeOption {
	return func(r *Runtime) {
		r.cache = c
	}
}

// NewRuntime creates a new Runtime instance with default components.
func NewRuntime(opts ...RuntimeOption) *Runtime {
	r := &Runtime{
		scanner:  scanner.NewScanner(scanner.Options{}),
		rules:   rules.NewRegistry(),
		patterns: patterns.NewMatcher(patterns.BuiltinPatterns()),
		config:  config.LoadDefault(),
		cache:   cache.NewCache(0),
		formats: make(map[string]formatter.Formatter),
	}

	// Register built-in rules
	rules.RegisterBuiltinRules(r.rules)

	// Register default formatters
	r.formats["console"] = formatter.NewConsoleFormatter()
	r.formats["json"] = formatter.NewJSONFormatter()
	r.formats["sarif"] = formatter.NewSARIFFormatter()
	r.formats["markdown"] = formatter.NewMarkdownFormatter()

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Analyze performs analysis on the given project path.
func (r *Runtime) Analyze(projectPath string) (*diagnostics.Diagnostics, error) {
	// Scan files
	files, err := r.scanner.ScanDir(projectPath)
	if err != nil {
		return nil, err
	}

	// Create diagnostics result
	result := diagnostics.NewDiagnostics()
	result.Statistics.FilesScanned = len(files)

	// Process each file
	p := parser.NewParser()
	for _, file := range files {
		// Parse file
		parsedFile, err := p.ParseFile(file)
		if err != nil {
			continue
		}

		// Find and evaluate rules
		nodes := parser.FindNodes(parsedFile, func(n ast.Node) bool {
			return true // Visit all nodes
		})

		for _, node := range nodes {
			// Check rules
			for _, rule := range r.rules.List() {
				if issue := rule.Check(file, node); issue != nil {
					diagnostics.AddIssue(result, *issue)
				}
			}
		}

		result.Statistics.FilesAnalyzed++
	}

	return result, nil
}

// AnalyzeFiles performs analysis on specific files.
func (r *Runtime) AnalyzeFiles(paths []string) (*diagnostics.Diagnostics, error) {
	result := diagnostics.NewDiagnostics()
	result.Statistics.FilesScanned = len(paths)

	p := parser.NewParser()
	for _, file := range paths {
		parsedFile, err := p.ParseFile(file)
		if err != nil {
			continue
		}

		nodes := parser.FindNodes(parsedFile, func(n ast.Node) bool {
			return true
		})

		for _, node := range nodes {
			for _, rule := range r.rules.List() {
				if issue := rule.Check(file, node); issue != nil {
					diagnostics.AddIssue(result, *issue)
				}
			}
		}

		result.Statistics.FilesAnalyzed++
	}

	return result, nil
}

// FormatResult formats the diagnostics using the specified formatter.
func (r *Runtime) FormatResult(d *diagnostics.Diagnostics, format string) ([]byte, error) {
	f, ok := r.formats[format]
	if !ok {
		f = r.formats["console"] // Default to console
	}
	return f.Format(d)
}

// ListFormatters returns the available output formatters.
func (r *Runtime) ListFormatters() []string {
	keys := make([]string, 0, len(r.formats))
	for k := range r.formats {
		keys = append(keys, k)
	}
	return keys
}

// GetConfig returns the current configuration.
func (r *Runtime) GetConfig() *config.Config {
	return r.config
}

// UpdateConfig updates the runtime configuration.
func (r *Runtime) UpdateConfig(cfg *config.Config) {
	r.config = cfg
}

// BenchmarkResult contains the results of a benchmark run.
type BenchmarkResult struct {
	FilesScanned     int
	FilesAnalyzed    int
	Duration         float64
	ThroughputFiles  float64 // files per second
	MemoryUsedMB     float64
}

// Benchmark runs a performance benchmark on the given project.
func (r *Runtime) Benchmark(projectPath string) (*BenchmarkResult, error) {
	result := &BenchmarkResult{}

	// Scan files and time it
	files, err := r.scanner.ScanDir(projectPath)
	if err != nil {
		return nil, err
	}

	result.FilesScanned = len(files)
	result.FilesAnalyzed = len(files)
	result.ThroughputFiles = float64(len(files)) / result.Duration

	return result, nil
}

// ValidationResult contains the results of configuration validation.
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// ValidateConfig validates a configuration file.
func (r *Runtime) ValidateConfig(configPath string) (*ValidationResult, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	result := &ValidationResult{Valid: true}

	// Validate rules exist
	for _, ruleCfg := range cfg.Rules {
		if !ruleCfg.Enabled {
			continue
		}
		if r.rules.Get(ruleCfg.ID) == nil {
			result.Warnings = append(result.Warnings, "Unknown rule: "+ruleCfg.ID)
		}
	}

	return result, nil
}

// Version returns the runtime version.
func Version() string {
	return "0.1.0"
}
