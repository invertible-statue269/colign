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
			Description: "Read a spec document for a change. For proposals, the content field is a JSON string with keys: problem, scope, outOfScope, approach.",
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
			Description: "Write or update a spec document for a change. For proposals, content must be a JSON string with keys: problem (required), scope (required), outOfScope (optional), approach (optional). Example: {\"problem\":\"...\",\"scope\":\"...\",\"outOfScope\":\"...\",\"approach\":\"...\"}. For other doc types, content is plain markdown.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
					"doc_type":  {Type: "string", Description: "Document type: proposal, design, spec, tasks"},
					"content":   {Type: "string", Description: "For proposal: JSON with problem, scope, outOfScope, approach. For others: markdown text."},
				},
				Required: []string{"change_id", "doc_type", "content"},
			},
		},
		{
			Name:        "create_task",
			Description: "Create a new implementation task for a change",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id":   {Type: "integer", Description: "Change ID"},
					"title":       {Type: "string", Description: "Task title"},
					"description": {Type: "string", Description: "Task description (optional)"},
					"status":      {Type: "string", Description: "Initial status: todo, in_progress, done (default: todo)"},
				},
				Required: []string{"change_id", "title"},
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
		{
			Name:        "list_acceptance_criteria",
			Description: "List acceptance criteria (Given/When/Then scenarios) for a change",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
				},
				Required: []string{"change_id"},
			},
		},
		{
			Name:        "create_acceptance_criteria",
			Description: "Create an acceptance criteria with BDD-style Given/When/Then steps for a change",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_id": {Type: "integer", Description: "Change ID"},
					"scenario":  {Type: "string", Description: "Scenario name describing the test case"},
					"steps":     {Type: "string", Description: "JSON array of steps, each with keyword (Given/When/Then/And/But) and text. Example: [{\"keyword\":\"Given\",\"text\":\"a user is logged in\"},{\"keyword\":\"When\",\"text\":\"they click logout\"},{\"keyword\":\"Then\",\"text\":\"they are redirected to login page\"}]"},
				},
				Required: []string{"change_id", "scenario", "steps"},
			},
		},
	}
}
