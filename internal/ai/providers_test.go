package ai_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/ai"
)

func TestNewChatModel_UnsupportedProvider(t *testing.T) {
	_, err := ai.NewChatModel(context.Background(), "unknown", "model", "key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestNewChatModel_OpenAI(t *testing.T) {
	m, err := ai.NewChatModel(context.Background(), "openai", "gpt-4o", "sk-fake-key")
	require.NoError(t, err)
	assert.NotNil(t, m)
}

func TestNewChatModel_Anthropic(t *testing.T) {
	m, err := ai.NewChatModel(context.Background(), "anthropic", "claude-sonnet-4-20250514", "sk-ant-fake-key")
	require.NoError(t, err)
	assert.NotNil(t, m)
}
