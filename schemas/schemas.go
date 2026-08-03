package schemas

import (
	"context"
)

// Schema represents a validation schema.
type Schema struct {
	ID          string
	Name        string
	Version     string
	Description string
	Type        string
	Source      string
}

// Validator defines the interface for schema validation.
type Validator interface {
	// Validate checks if the data conforms to the schema.
	Validate(ctx context.Context, data []byte) error

	// ValidateFile validates a file against the schema.
	ValidateFile(ctx context.Context, path string) error
}

// Loader defines the interface for loading schemas.
type Loader interface {
	// Load retrieves a schema by ID.
	Load(ctx context.Context, id string) (*Schema, error)

	// LoadAll returns all available schemas.
	LoadAll(ctx context.Context) ([]*Schema, error)
}

// Compiler defines the interface for compiling schemas.
type Compiler interface {
	// Compile converts a schema into a Validator.
	Compile(ctx context.Context, schema *Schema) (Validator, error)
}
