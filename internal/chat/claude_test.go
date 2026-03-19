package chat

import "testing"

func TestBuildMessages(t *testing.T) {
	history := []Message{
		{Role: "user", Content: "I need user authentication"},
		{Role: "assistant", Content: "What kind of auth do you need?"},
	}

	messages := BuildMessages("You are a spec assistant", history, "Add OAuth support")
	if len(messages) != 5 {
		t.Fatalf("expected 5 messages (system context pair + 2 history + 1 new), got %d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Error("first message should be system context as user role")
	}
	if messages[len(messages)-1].Content != "Add OAuth support" {
		t.Error("last message should be the new user message")
	}
}

func TestMessage(t *testing.T) {
	m := Message{Role: "user", Content: "hello"}
	if m.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", m.Role)
	}
}
