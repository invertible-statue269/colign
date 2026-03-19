package mcp

import (
	"encoding/json"
	"testing"
)

func TestToolDefinitions(t *testing.T) {
	tools := ListTools()
	if len(tools) == 0 {
		t.Fatal("expected at least one tool defined")
	}

	expectedTools := []string{
		"list_projects", "get_change", "read_spec",
		"write_spec", "list_tasks", "update_task", "suggest_spec",
	}

	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}

	for _, name := range expectedTools {
		if !toolMap[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestJSONRPCRequest(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	var req JSONRPCRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if req.Method != "tools/list" {
		t.Errorf("expected method 'tools/list', got '%s'", req.Method)
	}
}
