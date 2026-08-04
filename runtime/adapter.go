package runtime

import (
	"github.com/aliasfoxkde/Atheon/runtime/config"
	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// Adapter defines the interface for all Atheon adapters.
// Adapters are the integration points for different AI coding agents
// and development environments.
type Adapter interface {
	// Name returns the adapter's unique identifier.
	Name() string

	// Version returns the adapter version.
	Version() string

	// Initialize sets up the adapter with configuration.
	Initialize(cfg *config.Config) error

	// Validate checks if the adapter is properly configured.
	Validate() error

	// Analyze runs analysis on the provided input.
	// Each adapter interprets "input" differently:
	// - CLI: project path
	// - PI: workspace and changed files
	// - GitHub Actions: repository
	Analyze(input interface{}) (*diagnostics.Diagnostics, error)

	// FormatOutput converts diagnostics to the adapter's native format.
	FormatOutput(d *diagnostics.Diagnostics, format string) (interface{}, error)

	// Cleanup releases any resources held by the adapter.
	Cleanup() error
}

// AdapterRegistry manages available adapters.
type AdapterRegistry struct {
	adapters map[string]Adapter
}

// NewAdapterRegistry creates a new registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{
		adapters: make(map[string]Adapter),
	}
}

// Register adds an adapter to the registry.
func (r *AdapterRegistry) Register(adapter Adapter) error {
	if adapter == nil {
		return ErrNilAdapter
	}
	name := adapter.Name()
	if name == "" {
		return ErrInvalidAdapterName
	}
	if _, exists := r.adapters[name]; exists {
		return ErrAdapterAlreadyRegistered
	}
	r.adapters[name] = adapter
	return nil
}

// Get retrieves an adapter by name.
func (r *AdapterRegistry) Get(name string) (Adapter, bool) {
	adapter, ok := r.adapters[name]
	return adapter, ok
}

// List returns all registered adapter names.
func (r *AdapterRegistry) List() []string {
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// Unregister removes an adapter from the registry.
func (r *AdapterRegistry) Unregister(name string) bool {
	if _, exists := r.adapters[name]; exists {
		delete(r.adapters, name)
		return true
	}
	return false
}

// Errors for adapter operations.
var (
	ErrNilAdapter               = &AdapterError{"adapter cannot be nil"}
	ErrInvalidAdapterName       = &AdapterError{"adapter name cannot be empty"}
	ErrAdapterAlreadyRegistered = &AdapterError{"adapter already registered"}
	ErrAdapterNotFound          = &AdapterError{"adapter not found"}
)

// AdapterError represents an adapter-related error.
type AdapterError struct {
	Message string
}

func (e *AdapterError) Error() string {
	return e.Message
}
