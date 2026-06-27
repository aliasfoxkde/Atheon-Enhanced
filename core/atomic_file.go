package core

import (
	"os"

	"github.com/aliasfoxkde/Atheon/internal/atomicio"
)

// atomicWriteFile is a wrapper around atomicio.WriteFile for core package use.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	return atomicio.WriteFile(path, data, perm)
}
