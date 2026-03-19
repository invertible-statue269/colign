package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const claudeAPIURL = "https://api.anthropic.com/v1/messages"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeClient struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		model:      "claude-sonnet-4-20250514",
	}
}

func BuildMessages(systemContext string, history []Message, newMessage string) []Message {
	messages := make([]Message, 0, len(history)+2)

	// Add system context as first user message (Claude API uses system parameter separately)
	if systemContext != "" {
		messages = append(messages, Message{Role: "user", Content: systemContext})
		messages = append(messages, Message{Role: "assistant", Content: "Understood. I'll help with spec writing based on this context."})
	}

	messages = append(messages, history...)
	messages = append(messages, Message{Role: "user", Content: newMessage})
	return messages
}

type claudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
}

type StreamCallback func(text string) error

func (c *ClaudeClient) StreamMessage(ctx context.Context, system string, messages []Message, callback StreamCallback) error {
	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  messages,
		Stream:    true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("claude API error %d: %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Type == "content_block_delta" && event.Delta.Text != "" {
			if err := callback(event.Delta.Text); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
