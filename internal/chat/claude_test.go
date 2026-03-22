package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMessages(t *testing.T) {
	history := []Message{
		{Role: "user", Content: "I need user authentication"},
		{Role: "assistant", Content: "What kind of auth do you need?"},
	}

	messages := BuildMessages("You are a spec assistant", history, "Add OAuth support")
	require.Len(t, messages, 5, "expected 5 messages (system context pair + 2 history + 1 new)")
	assert.Equal(t, "user", messages[0].Role, "first message should be system context as user role")
	assert.Equal(t, "Add OAuth support", messages[len(messages)-1].Content, "last message should be the new user message")
}

func TestMessage(t *testing.T) {
	m := Message{Role: "user", Content: "hello"}
	assert.Equal(t, "user", m.Role)
}
