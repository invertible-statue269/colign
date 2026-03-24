package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/aiconfig"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockProposalGen struct {
	ch  <-chan SectionChunk
	err error
}

func (m *mockProposalGen) GenerateProposal(_ context.Context, _ *aiconfig.AIConfig, _ GenerateProposalInput) (<-chan SectionChunk, error) {
	return m.ch, m.err
}

type mockACGen struct {
	result []GeneratedAC
	err    error
}

func (m *mockACGen) GenerateAC(_ context.Context, _ *aiconfig.AIConfig, _ GenerateACInput) ([]GeneratedAC, error) {
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// writeAIError tests
// ---------------------------------------------------------------------------

func TestWriteAIError_Unauthenticated(t *testing.T) {
	w := httptest.NewRecorder()
	writeAIError(w, errUnauthenticated)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWriteAIError_RateLimited(t *testing.T) {
	w := httptest.NewRecorder()
	writeAIError(w, errRateLimited)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestWriteAIError_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	writeAIError(w, errBadRequest)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWriteAIError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	writeAIError(w, errNotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWriteAIError_AINotConfigured(t *testing.T) {
	w := httptest.NewRecorder()
	writeAIError(w, errAINotConfigured)
	assert.Equal(t, http.StatusPreconditionFailed, w.Code)
}

// ---------------------------------------------------------------------------
// SSE formatting test (unit-level, no DB)
// ---------------------------------------------------------------------------

func TestHandleGenerateProposal_SSE(t *testing.T) {
	// Build a channel with two chunks and close it.
	ch := make(chan SectionChunk, 2)
	ch <- SectionChunk{Section: "overview", Text: "hello"}
	ch <- SectionChunk{Section: "overview", Text: " world"}
	close(ch)

	gen := &mockProposalGen{ch: ch}
	cfg := &aiconfig.AIConfig{Provider: "openai", Model: "gpt-4o"}

	// Call the SSE writer directly (bypasses DB/auth).
	w := httptest.NewRecorder()
	writeSSEProposal(w, r_noop(), gen, cfg, GenerateProposalInput{Description: "desc"})

	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	body := w.Body.String()
	assert.Contains(t, body, "data: ")
	assert.Contains(t, body, "[DONE]")

	// Verify each event is valid JSON with the expected fields.
	lines := strings.Split(body, "\n")
	var dataLines []string
	for _, l := range lines {
		if strings.HasPrefix(l, "data: ") && !strings.Contains(l, "[DONE]") {
			dataLines = append(dataLines, strings.TrimPrefix(l, "data: "))
		}
	}
	require.Len(t, dataLines, 2)
	for _, dl := range dataLines {
		var chunk SectionChunk
		require.NoError(t, json.Unmarshal([]byte(dl), &chunk))
		assert.Equal(t, "overview", chunk.Section)
	}
}

// ---------------------------------------------------------------------------
// AC JSON endpoint test (unit-level, no DB)
// ---------------------------------------------------------------------------

func TestHandleGenerateAC_JSON(t *testing.T) {
	expected := []GeneratedAC{
		{
			Scenario: "User logs in",
			Steps: []ACStep{
				{Keyword: "Given", Text: "user is on login page"},
				{Keyword: "When", Text: "user submits credentials"},
				{Keyword: "Then", Text: "user is redirected to dashboard"},
			},
		},
	}

	gen := &mockACGen{result: expected}
	cfg := &aiconfig.AIConfig{Provider: "openai", Model: "gpt-4o"}

	w := httptest.NewRecorder()
	writeACJSON(w, r_noop(), gen, cfg, GenerateACInput{Proposal: "proposal text"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var got []GeneratedAC
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	require.Len(t, got, 1)
	assert.Equal(t, "User logs in", got[0].Scenario)
	require.Len(t, got[0].Steps, 3)
	assert.Equal(t, "Given", got[0].Steps[0].Keyword)
}

// r_noop returns a minimal *http.Request for tests that don't need a real one.
func r_noop() *http.Request {
	return httptest.NewRequest(http.MethodPost, "/", nil)
}
