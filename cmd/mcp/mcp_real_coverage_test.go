package main

import (
	"encoding/json"
	"testing"

	"atheon/core"
)

func TestToolList(t *testing.T) {
	tools := toolList()

	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	// Verify tool structure
	for _, tool := range tools {
		if tool["name"] == nil {
			t.Error("tool must have a name")
		}
		if tool["description"] == nil {
			t.Error("tool must have a description")
		}
		if tool["inputSchema"] == nil {
			t.Error("tool must have an input schema")
		}
	}
}

func TestHandleCallScanString(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_string","arguments":{"content":"AKIAIOSFODNN7EXAMPLE","source":"test"}}`)

	result, err := handleCall(params)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Error("expected result from scan_string")
	}

	// Verify result structure
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Error("expected map result")
		return
	}

	if resultMap["content"] == nil {
		t.Error("expected content in result")
	}
}

func TestHandleCallScanFile(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_file","arguments":{"path":"/tmp/test.txt"}}`)

	result, err := handleCall(params)

	// File might not exist, but shouldn't panic
	if err != nil {
		// Expected for non-existent files
		t.Logf("Expected error for non-existent file: %v", err)
	}

	// Check if we got a valid result
	if result != nil {
		// Result should be a map with content
		switch v := result.(type) {
		case map[string]any:
			if v["content"] == nil {
				t.Error("expected content in result")
			}
		default:
			t.Errorf("unexpected result type: %T", v)
		}
	}
}

func TestHandleCallScanDir(t *testing.T) {
	dir := t.TempDir()
	pathJSON, _ := json.Marshal(dir)
	params := json.RawMessage(`{"name":"scan_dir","arguments":{"path":` + string(pathJSON) + `,"categories":["secrets"]}}`)

	result, err := handleCall(params)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result == nil {
		t.Error("expected result from scan_dir")
	} else {
		// Result should be a map, not an interface
		switch v := result.(type) {
		case map[string]any:
			if v["content"] == nil {
				t.Error("expected content in result")
			}
		default:
			t.Logf("unexpected result type: %T", v)
		}
	}
}

func TestHandleCallInvalidTool(t *testing.T) {
	params := json.RawMessage(`{"name":"invalid_tool","arguments":{}}`)

	result, err := handleCall(params)

	if err == nil {
		t.Error("expected error for invalid tool")
	}

	if result != nil {
		t.Error("expected nil result for invalid tool")
	}
}

func TestHandleCallMissingArguments(t *testing.T) {
	params := json.RawMessage(`{"name":"scan_string"}`)

	result, err := handleCall(params)

	// Should handle missing arguments gracefully
	if err != nil {
		t.Logf("Expected error for missing arguments: %v", err)
	}

	// Should return some result (even if error)
	if result == nil && err == nil {
		t.Error("expected either result or error")
	}
}

func TestHandleCallInvalidParams(t *testing.T) {
	params := json.RawMessage(`invalid json`)

	result, err := handleCall(params)

	if err == nil {
		t.Error("expected error for invalid params")
	}

	if result != nil {
		t.Error("expected nil result for invalid params")
	}
}

func TestTextResult(t *testing.T) {
	tests := []struct {
		name     string
		findings []core.Finding
	}{
		{
			name:     "no findings",
			findings: []core.Finding{},
		},
		{
			name: "single finding",
			findings: []core.Finding{
				{Pattern: "test", File: "test.txt", Line: 1, Content: "content"},
			},
		},
		{
			name: "multiple findings",
			findings: []core.Finding{
				{Pattern: "test1", File: "test1.txt", Line: 1, Content: "content1"},
				{Pattern: "test2", File: "test2.txt", Line: 2, Content: "content2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := textResult(tt.findings)

			if result == nil {
				t.Error("expected result from textResult")
				return
			}

			// result is already map[string]any, not an interface
			if result["content"] == nil {
				t.Error("expected content in result")
			}
		})
	}
}

func TestMCPToolNames(t *testing.T) {
	tools := toolList()

	expectedNames := []string{"scan_string", "scan_file", "scan_dir"}
	foundNames := make(map[string]bool)

	for _, tool := range tools {
		name, ok := tool["name"].(string)
		if !ok {
			continue
		}
		foundNames[name] = true
	}

	for _, expectedName := range expectedNames {
		if !foundNames[expectedName] {
			t.Errorf("expected to find tool %s", expectedName)
		}
	}
}

func TestMCPToolDescriptions(t *testing.T) {
	tools := toolList()

	for _, tool := range tools {
		description, ok := tool["description"].(string)
		if !ok {
			t.Errorf("tool description should be string")
			continue
		}

		if description == "" {
			t.Error("tool should have a description")
		}
	}
}

func TestMCPToolInputSchemas(t *testing.T) {
	tools := toolList()

	for _, tool := range tools {
		schema, ok := tool["inputSchema"].(map[string]any)
		if !ok {
			t.Errorf("tool inputSchema should be map")
			continue
		}

		// Check required fields
		if schema["type"] != "object" {
			t.Errorf("inputSchema type should be object")
		}

		if schema["properties"] == nil {
			t.Errorf("inputSchema should have properties")
		}

		if schema["required"] == nil {
			t.Errorf("inputSchema should have required fields")
		}
	}
}
