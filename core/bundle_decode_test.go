package core

import (
	"testing"
)

// decodeJSONStrict is tested via the public LoadBundle API.
// These are supplementary unit tests for edge cases.

func TestDecodeJSONStrict_ObjectUnknownField(t *testing.T) {
	type validStruct struct {
		Name string `json:"name"`
	}
	var out validStruct

	// JSON with unknown field using Go field name (validates against Go names)
	data := []byte(`{"Name":"test","UnknownField":99}`)

	err := decodeJSONStrict(data, &out)
	if err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestDecodeJSONStrict_ObjectValid(t *testing.T) {
	type validStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var out validStruct

	// Note: decodeJSONStrict validates using Go field names, not JSON names
	// So we use Go field names in JSON for this test
	data := []byte(`{"Name":"test","Age":99}`)

	err := decodeJSONStrict(data, &out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out.Name != "test" || out.Age != 99 {
		t.Errorf("got %+v, want Name=test Age=99", out)
	}
}

func TestDecodeJSONStrict_ArrayValid(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}
	var out []item

	data := []byte(`[{"id":1},{"id":2}]`)

	err := decodeJSONStrict(data, &out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("got %d items, want 2", len(out))
	}
}

func TestDecodeJSONStrict_EmptyData(t *testing.T) {
	var out struct{}
	err := decodeJSONStrict([]byte{}, &out)
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}

func TestDecodeJSONStrict_InvalidJSON(t *testing.T) {
	var out struct{}
	err := decodeJSONStrict([]byte(`{invalid`), &out)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDecodeJSONStrict_UnexpectedRootType(t *testing.T) {
	var out struct{}
	err := decodeJSONStrict([]byte(`"string"`), &out)
	if err == nil {
		t.Error("expected error for unexpected root type, got nil")
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
	}{
		{[]byte("  hello  "), []byte("hello")},
		{[]byte("\t\nhello\r\n"), []byte("hello")},
		{[]byte("no_whitespace"), []byte("no_whitespace")},
		{[]byte("   "), []byte("")},
		{[]byte(""), []byte("")},
	}
	for _, tt := range tests {
		got := trimSpace(tt.input)
		if string(got) != string(tt.expected) {
			t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDecodeBundleDefs_V2Schema(t *testing.T) {
	data := []byte(`{"schema_version":2,"data":[{"name":"test","category":"test","match":"test","enabled":true}]}`)

	defs, err := decodeBundleDefs(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("got %d defs, want 1", len(defs))
	}
}

func TestDecodeBundleDefs_V2UnknownField(t *testing.T) {
	data := []byte(`{"schema_version":2,"data":[],"unknown_field":true}`)

	_, err := decodeBundleDefs(data)
	if err == nil {
		t.Error("expected error for unknown field in v2 envelope, got nil")
	}
}

func TestDecodeBundleDefs_V1Array(t *testing.T) {
	data := []byte(`[{"name":"test","category":"test","match":"test","enabled":true}]`)

	defs, err := decodeBundleDefs(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("got %d defs, want 1", len(defs))
	}
}

func TestDecodeBundleDefs_EmptyData(t *testing.T) {
	_, err := decodeBundleDefs([]byte{})
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}
}

func TestDecodeBundleDefs_WhitespaceOnly(t *testing.T) {
	_, err := decodeBundleDefs([]byte("   \n\t  "))
	if err == nil {
		t.Error("expected error for whitespace-only data, got nil")
	}
}

func TestDecodeBundleDefs_InvalidJSON(t *testing.T) {
	_, err := decodeBundleDefs([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDecodeBundleDefs_UnsupportedVersion(t *testing.T) {
	data := []byte(`{"schema_version":99,"data":[]}`)
	_, err := decodeBundleDefs(data)
	if err == nil {
		t.Error("expected error for unsupported schema version, got nil")
	}
}

func TestDecodeBundleDefs_V2InvalidJSON(t *testing.T) {
	data := []byte(`{"schema_version":2,"data":[{"invalid"}]`)
	_, err := decodeBundleDefs(data)
	if err == nil {
		t.Error("expected error for invalid JSON in v2 data, got nil")
	}
}

func TestNormalizeConfidence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"high", "high"},
		{"medium", "medium"},
		{"low", "low"},
		{"HIGH", "high"}, // case insensitive
		{"Medium", "medium"},
		{"invalid", "medium"}, // defaults to medium
		{"", "medium"},        // defaults to medium
	}
	for _, tt := range tests {
		got := normalizeConfidence(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeConfidence(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
