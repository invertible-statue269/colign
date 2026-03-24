package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// NewChatModel creates an eino BaseChatModel for the given provider.
func NewChatModel(ctx context.Context, provider, modelName, apiKey string) (model.BaseChatModel, error) {
	switch provider {
	case "openai":
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			Model:  modelName,
			APIKey: apiKey,
		})
	case "anthropic":
		return claude.NewChatModel(ctx, &claude.Config{
			Model:     modelName,
			APIKey:    apiKey,
			MaxTokens: 4096,
		})
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// TestConnection verifies that the given provider credentials are valid
// by making a minimal API call.
func TestConnection(ctx context.Context, provider, modelName, apiKey string) error {
	chatModel, err := NewChatModel(ctx, provider, modelName, apiKey)
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}

	// Make a minimal completion request to verify credentials.
	_, err = chatModel.Generate(ctx, []*schema.Message{
		{Role: schema.User, Content: "hi"},
	})
	if err != nil {
		return fmt.Errorf("test connection: %w", err)
	}
	return nil
}
