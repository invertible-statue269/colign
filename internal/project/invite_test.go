package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/models"
)

func TestValidateInviteRole(t *testing.T) {
	tests := []struct {
		role  string
		valid bool
	}{
		{"owner", true},
		{"editor", true},
		{"viewer", true},
		{"admin", false},
		{"", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.valid, ValidateRole(tt.role), "ValidateRole(%q)", tt.role)
	}
}

func TestParseRole(t *testing.T) {
	role, err := ParseRole("editor")
	require.NoError(t, err)
	assert.Equal(t, models.RoleEditor, role)

	_, err = ParseRole("invalid")
	assert.Error(t, err, "expected error for invalid role")
}
