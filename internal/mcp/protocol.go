package mcp

import "encoding/json"

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func ListTools() []Tool {
	return []Tool{
		{
			Name:        "list_projects",
			Description: "List all projects the user has access to",
			InputSchema: InputSchema{Type: "object"},
		},
		{
			Name:        "get_change",
			Description: "Get details of a specific change including its stage and artifacts",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
				},
				Required: []string{"change_id"},
			},
		},
		{
			Name:        "read_spec",
			Description: "Read a spec document for a change",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
					"doc_type":  {Type: "string", Description: "Document type: proposal, design, spec, tasks"},
				},
				Required: []string{"change_id", "doc_type"},
			},
		},
		{
			Name:        "write_spec",
			Description: "Write or update a spec document for a change",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
					"doc_type":  {Type: "string", Description: "Document type: proposal, design, spec, tasks"},
					"content":   {Type: "string", Description: "Document content in markdown"},
				},
				Required: []string{"change_id", "doc_type", "content"},
			},
		},
		{
			Name:        "list_tasks",
			Description: "List implementation tasks for a change",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
				},
				Required: []string{"change_id"},
			},
		},
		{
			Name:        "update_task",
			Description: "Update a task's status",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"task_id": {Type: "integer", Description: "Task ID"},
					"status":  {Type: "string", Description: "New status: todo, in_progress, done"},
				},
				Required: []string{"task_id", "status"},
			},
		},
		{
			Name:        "suggest_spec",
			Description: "Get AI suggestions for improving a spec document",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
					"doc_type":  {Type: "string", Description: "Document type to improve"},
				},
				Required: []string{"change_id", "doc_type"},
			},
		},
	}
}
