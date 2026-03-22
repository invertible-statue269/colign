package project

import (
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
