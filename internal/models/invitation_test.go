package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPendingInvitationModel(t *testing.T) {
	inv := &PendingInvitation{
		ProjectID: 1,
		Email:     "new@example.com",
		Role:      RoleEditor,
		Token:     "abc123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	assert.Equal(t, "new@example.com", inv.Email)
	assert.Equal(t, RoleEditor, inv.Role)
}
