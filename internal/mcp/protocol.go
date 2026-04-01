package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
)

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

// ToolHandler is the function signature for tool execution.
type ToolHandler func(s *Server, ctx context.Context, args json.RawMessage) (any, error)

type ToolAnnotations struct {
	DestructiveHint *bool  `json:"destructiveHint,omitempty"`
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`
	OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`
	Title           string `json:"title,omitempty"`
}

type Tool struct {
	Meta        map[string]any   `json:"_meta,omitempty"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	InputSchema InputSchema      `json:"inputSchema"`
	ReadOnly    bool             `json:"-"`
	Destructive *bool            `json:"-"`
	Idempotent  bool             `json:"-"`
	OpenWorld   *bool            `json:"-"`
	Handler     ToolHandler      `json:"-"`
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

var (
	toolRegistryMu sync.RWMutex
	toolRegistry   []Tool
	toolIndex      = make(map[string]int)
)

func RegisterTool(t Tool) {
	normalized := normalizeTool(t)

	toolRegistryMu.Lock()
	defer toolRegistryMu.Unlock()

	if _, exists := toolIndex[normalized.Name]; exists {
		panic(fmt.Sprintf("duplicate MCP tool registration: %s", normalized.Name))
	}

	toolIndex[normalized.Name] = len(toolRegistry)
	toolRegistry = append(toolRegistry, normalized)
}

func ListTools() []Tool {
	toolRegistryMu.RLock()
	defer toolRegistryMu.RUnlock()
	return slices.Clone(toolRegistry)
}

func FindTool(name string) (Tool, bool) {
	toolRegistryMu.RLock()
	defer toolRegistryMu.RUnlock()

	i, ok := toolIndex[name]
	if !ok {
		return Tool{}, false
	}
	return toolRegistry[i], true
}

// callTool dispatches a tool call to the registered handler.
func (s *Server) callTool(ctx context.Context, name string, args json.RawMessage) (any, error) {
	t, ok := FindTool(name)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Handler(s, ctx, args)
}

func normalizeTool(t Tool) Tool {
	if t.Annotations == nil {
		t.Annotations = &ToolAnnotations{}
	}
	t.Annotations.ReadOnlyHint = t.ReadOnly
	if t.Destructive != nil {
		t.Annotations.DestructiveHint = t.Destructive
	} else if t.ReadOnly {
		t.Annotations.DestructiveHint = boolPtr(false)
	}
	t.Annotations.IdempotentHint = t.Idempotent
	if t.OpenWorld != nil {
		t.Annotations.OpenWorldHint = t.OpenWorld
	}

	return t
}

func boolPtr(v bool) *bool {
	return &v
}
