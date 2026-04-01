package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewStreamableHandler creates an http.Handler for the MCP Streamable HTTP transport.
// It uses the official MCP Go SDK and delegates tool calls to the existing handler logic.
func NewStreamableHandler(apiURL, apiToken string) http.Handler {
	clients := newAPIClients(apiURL, apiToken)

	mcpServer := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "colign-mcp",
		Version: "0.1.0",
	}, nil)

	registerTools(mcpServer, clients)

	return sdkmcp.NewStreamableHTTPHandler(func(r *http.Request) *sdkmcp.Server {
		return mcpServer
	}, nil)
}

// NewStreamableHandlerWithAuth creates an http.Handler that extracts the API token
// from the Authorization header per-request, allowing multi-tenant usage.
func NewStreamableHandlerWithAuth(apiURL string, opts ...clientOption) http.Handler {
	return sdkmcp.NewStreamableHTTPHandler(func(r *http.Request) *sdkmcp.Server {
		token := r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if token == "" {
			return nil // results in 400
		}

		clients := newAPIClients(apiURL, token, opts...)

		server := sdkmcp.NewServer(&sdkmcp.Implementation{
			Name:    "colign-mcp",
			Version: "0.1.0",
		}, nil)
		registerTools(server, clients)
		return server
	}, &sdkmcp.StreamableHTTPOptions{
		Stateless: true,
	})
}

func registerTools(s *sdkmcp.Server, clients *apiClients) {
	for _, tool := range ListTools() {
		schema, _ := json.Marshal(tool.InputSchema)
		s.AddTool(&sdkmcp.Tool{
			Meta:        sdkmcp.Meta(tool.Meta),
			Annotations: toSDKToolAnnotations(tool.Annotations),
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: json.RawMessage(schema),
		}, makeToolHandler(tool, clients))
	}
}

func makeToolHandler(tool Tool, clients *apiClients) sdkmcp.ToolHandler {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
		// Create a temporary server-like struct to reuse existing handler logic
		s := &Server{clients: clients}

		result, err := tool.Handler(s, ctx, req.Params.Arguments)
		if err != nil {
			return &sdkmcp.CallToolResult{
				Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			}, nil
		}

		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(resultJSON)}},
		}, nil
	}
}

func toSDKToolAnnotations(a *ToolAnnotations) *sdkmcp.ToolAnnotations {
	if a == nil {
		return nil
	}
	return &sdkmcp.ToolAnnotations{
		DestructiveHint: a.DestructiveHint,
		IdempotentHint:  a.IdempotentHint,
		OpenWorldHint:   a.OpenWorldHint,
		ReadOnlyHint:    a.ReadOnlyHint,
		Title:           a.Title,
	}
}
