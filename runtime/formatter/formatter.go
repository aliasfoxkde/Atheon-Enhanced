package formatter

import "github.com/aliasfoxkde/Atheon/runtime/diagnostics"

type Formatter interface {
	Format(d *diagnostics.Diagnostics) ([]byte, error)
	ContentType() string
	FileExtension() string
}
