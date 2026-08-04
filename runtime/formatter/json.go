package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// JSONFormatter implements JSON output for diagnostics.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSONFormatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format outputs diagnostics in JSON format.
func (f *JSONFormatter) Format(d *diagnostics.Diagnostics) ([]byte, error) {
	if d == nil {
		d = diagnostics.NewDiagnostics()
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return append(data, '\n'), nil
}

// ContentType returns the MIME type for JSON output.
func (f *JSONFormatter) ContentType() string {
	return "application/json"
}

// FileExtension returns the file extension for JSON output.
func (f *JSONFormatter) FileExtension() string {
	return ".json"
}
