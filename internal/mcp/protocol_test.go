package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolDefinitions(t *testing.T) {
	tools := ListTools()
	require.NotEmpty(t, tools, "expected at least one tool defined")

	expectedTools := []string{
		"list_projects", "get_change", "read_spec",
		"write_spec", "list_tasks", "update_task", "suggest_spec",
	}

	toolMap := make(map[string]bool)
	for _, tool := range tools {
		toolMap[tool.Name] = true
	}

	for _, name := range expectedTools {
		assert.True(t, toolMap[name], "missing tool: %s", name)
	}
}

func TestJSONRPCRequest(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	var req JSONRPCRequest
	require.NoError(t, json.Unmarshal([]byte(raw), &req))
	assert.Equal(t, "tools/list", req.Method)
}
