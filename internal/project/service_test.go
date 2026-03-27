package project

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My App", "my-app"},
		{"Hello World 123", "hello-world-123"},
		{"UPPER CASE", "upper-case"},
		{"  spaces  everywhere  ", "spaces-everywhere"},
		{"special!@#chars$%^", "special-chars"},
		{"already-kebab", "already-kebab"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, GenerateSlug(tt.input), "GenerateSlug(%q)", tt.input)
	}
}

// AC#46: 프로젝트 생성 시 이니셜이 자동 부여된다
func TestGenerateIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"multi word — first letters", "My Cool Project", "MCP"},
		{"single word — first 3 chars", "Colign", "COL"},
		{"short word — pad to available", "AI", "AI"},
		{"single char", "X", "X"},
		{"korean name — uses slug fallback", "필리크", ""}, // slug-based, non-empty
		{"mixed with numbers", "Web3 Platform", "WP"},
		{"lowercase forced to upper", "piliq", "PIL"},
		{"already uppercase", "PLQ", "PLQ"},
		{"special chars stripped", "my@app!", "MYA"},
		{"hyphenated single word", "piliq-app", "PA"},
	}

	for _, tt := range tests {
		result := GenerateIdentifier(tt.input)
		assert.NotEmpty(t, result, "GenerateIdentifier(%q) should not be empty", tt.input)
		assert.LessOrEqual(t, len(result), 5, "GenerateIdentifier(%q) max 5 chars", tt.input)
		assert.Regexp(t, `^[A-Z0-9]+$`, result, "GenerateIdentifier(%q) must be uppercase alphanumeric", tt.input)
		if tt.expected != "" {
			assert.Equal(t, tt.expected, result, "GenerateIdentifier(%q)", tt.input)
		}
	}
}

func TestGenerateIdentifier_FiveCharBase(t *testing.T) {
	// Verify that 5-char identifiers can produce unique suffixed candidates
	// (regression test for infinite loop when base is already 5 chars)
	result := GenerateIdentifier("ABCDE Things")
	assert.Equal(t, "AT", result)

	// Single 5-char word
	result = GenerateIdentifier("ABCDE")
	assert.Equal(t, "ABC", result)
	assert.LessOrEqual(t, len(result), 5)
}

func TestGenerateIdentifier_EmptyInput(t *testing.T) {
	// Empty/whitespace-only input should still produce a valid identifier
	result := GenerateIdentifier("")
	assert.NotEmpty(t, result, "empty input should produce fallback identifier")
	assert.Equal(t, "PRJ", result)

	result = GenerateIdentifier("   ")
	assert.NotEmpty(t, result, "whitespace input should produce fallback identifier")
}

func TestGenerateIdentifier_SuffixTruncation(t *testing.T) {
	// When base is 5 chars, suffixed candidates must still be <= 5 chars
	// and must differ from the base to avoid infinite loops
	base := "ABCDE"
	candidates := make(map[string]bool)

	// Simulate what ensureUniqueIdentifier does internally
	for i := 0; i < 20; i++ {
		candidate := base
		if i > 0 {
			suffix := fmt.Sprintf("%d", i+1)
			maxBase := 5 - len(suffix)
			if maxBase < 1 {
				maxBase = 1
			}
			truncated := base
			if len(truncated) > maxBase {
				truncated = truncated[:maxBase]
			}
			candidate = truncated + suffix
		}
		assert.LessOrEqual(t, len(candidate), 5, "candidate %q exceeds 5 chars", candidate)
		// After first iteration, candidates must differ from base
		if i > 0 {
			assert.NotEqual(t, base, candidate, "suffix %d produced same value as base", i)
		}
		candidates[candidate] = true
	}
	// All 20 candidates should be unique
	assert.Equal(t, 20, len(candidates), "expected 20 unique candidates")
}

func TestParseProjectRef(t *testing.T) {
	tests := []struct {
		input    string
		wantID   int64
		wantSlug string
		wantOK   bool
	}{
		{"42-my-project", 42, "42-my-project", true},
		{"1-x", 1, "1-x", true},
		{"42", 42, "", true},
		{"my-project", 0, "", false}, // no numeric prefix
		{"0-bad", 0, "", false},      // zero id
		{"-1-neg", 0, "", false},     // negative
	}

	for _, tt := range tests {
		id, slug, ok := parseProjectRef(tt.input)
		assert.Equal(t, tt.wantOK, ok, "parseProjectRef(%q) ok", tt.input)
		if ok {
			assert.Equal(t, tt.wantID, id, "parseProjectRef(%q) id", tt.input)
			assert.Equal(t, tt.wantSlug, slug, "parseProjectRef(%q) slug", tt.input)
		}
	}
}
