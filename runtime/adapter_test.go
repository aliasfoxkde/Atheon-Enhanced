package runtime

import (
	"testing"

	"github.com/aliasfoxkde/Atheon/runtime/config"
	"github.com/aliasfoxkde/Atheon/runtime/diagnostics"
)

// mockAdapter implements Adapter for testing
type mockAdapter struct {
	name        string
	version     string
	initialized bool
	validated   bool
}

func (m *mockAdapter) Name() string {
	return m.name
}

func (m *mockAdapter) Version() string {
	return m.version
}

func (m *mockAdapter) Initialize(cfg *config.Config) error {
	m.initialized = true
	return nil
}

func (m *mockAdapter) Validate() error {
	m.validated = true
	return nil
}

func (m *mockAdapter) Analyze(input interface{}) (*diagnostics.Diagnostics, error) {
	return &diagnostics.Diagnostics{}, nil
}

func (m *mockAdapter) FormatOutput(d *diagnostics.Diagnostics, format string) (interface{}, error) {
	return d, nil
}

func (m *mockAdapter) Cleanup() error {
	return nil
}

func TestNewAdapterRegistry(t *testing.T) {
	r := NewAdapterRegistry()
	if r == nil {
		t.Fatal("NewAdapterRegistry returned nil")
	}
	if r.adapters == nil {
		t.Error("adapters map is nil")
	}
}

func TestAdapterRegistryRegister(t *testing.T) {
	r := NewAdapterRegistry()
	adapter := &mockAdapter{name: "test", version: "1.0"}

	err := r.Register(adapter)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, ok := r.Get("test")
	if !ok {
		t.Error("adapter not found after registration")
	}
	if got.Name() != "test" {
		t.Errorf("expected name 'test', got %q", got.Name())
	}
}

func TestAdapterRegistryRegisterNil(t *testing.T) {
	r := NewAdapterRegistry()
	err := r.Register(nil)
	if err != ErrNilAdapter {
		t.Errorf("expected ErrNilAdapter, got %v", err)
	}
}

func TestAdapterRegistryRegisterEmptyName(t *testing.T) {
	r := NewAdapterRegistry()
	adapter := &mockAdapter{name: "", version: "1.0"}
	err := r.Register(adapter)
	if err != ErrInvalidAdapterName {
		t.Errorf("expected ErrInvalidAdapterName, got %v", err)
	}
}

func TestAdapterRegistryRegisterAlreadyExists(t *testing.T) {
	r := NewAdapterRegistry()
	adapter := &mockAdapter{name: "test", version: "1.0"}

	err := r.Register(adapter)
	if err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	err = r.Register(adapter)
	if err != ErrAdapterAlreadyRegistered {
		t.Errorf("expected ErrAdapterAlreadyRegistered, got %v", err)
	}
}

func TestAdapterRegistryGet(t *testing.T) {
	r := NewAdapterRegistry()
	adapter := &mockAdapter{name: "test", version: "1.0"}
	r.Register(adapter)

	got, ok := r.Get("test")
	if !ok {
		t.Error("Get returned ok=false")
	}
	if got.Name() != "test" {
		t.Errorf("expected 'test', got %q", got.Name())
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Get should return false for nonexistent adapter")
	}
}

func TestAdapterRegistryList(t *testing.T) {
	r := NewAdapterRegistry()
	if len(r.List()) != 0 {
		t.Error("List should be empty for new registry")
	}

	r.Register(&mockAdapter{name: "test1", version: "1.0"})
	r.Register(&mockAdapter{name: "test2", version: "1.0"})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 adapters, got %d", len(list))
	}
}

func TestAdapterRegistryUnregister(t *testing.T) {
	r := NewAdapterRegistry()
	adapter := &mockAdapter{name: "test", version: "1.0"}
	r.Register(adapter)

	ok := r.Unregister("test")
	if !ok {
		t.Error("Unregister should return true")
	}

	_, ok = r.Get("test")
	if ok {
		t.Error("adapter should not exist after Unregister")
	}

	ok = r.Unregister("nonexistent")
	if ok {
		t.Error("Unregister should return false for nonexistent adapter")
	}
}

func TestAdapterError(t *testing.T) {
	err := &AdapterError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", err.Error())
	}
}
