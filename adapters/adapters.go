// Package adapters provides the adapter framework for integrating external
// pattern and rule sources into the Atheon analysis pipeline.
package adapters

import (
	"context"
	"time"
)

// Source represents an external source of patterns or rules.
type Source struct {
	ID       string
	Name     string
	Type     string
	URL      string
	Priority int
	Enabled  bool
}

// Adapter defines the interface for interacting with external sources.
type Adapter interface {
	// Connect establishes a connection to the source.
	Connect(ctx context.Context) error

	// Fetch retrieves data from the source.
	Fetch(ctx context.Context, path string) ([]byte, error)

	// Close closes the connection to the source.
	Close(ctx context.Context) error
}

// Registry provides access to registered adapters.
type Registry interface {
	// Get retrieves an adapter by name.
	Get(name string) (Adapter, error)

	// Register adds an adapter to the registry.
	Register(name string, adapter Adapter) error

	// List returns all registered adapters.
	List() []string
}

// HealthChecker defines the interface for checking source health.
type HealthChecker interface {
	// Check verifies the source is reachable and healthy.
	Check(ctx context.Context, source *Source) error
}

// Updater defines the interface for updating from external sources.
type Updater interface {
	// Update checks for and applies updates from the source.
	Update(ctx context.Context, source *Source) error

	// LastUpdate returns the time of the last successful update.
	LastUpdate(sourceID string) (time.Time, error)
}
