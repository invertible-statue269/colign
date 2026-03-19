package models

import (
	"testing"
	"time"
)

func TestPendingInvitationModel(t *testing.T) {
	inv := &PendingInvitation{
		ProjectID: 1,
		Email:     "new@example.com",
		Role:      RoleEditor,
		Token:     "abc123",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if inv.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", inv.Email)
	}
	if inv.Role != RoleEditor {
		t.Errorf("expected role editor, got %s", inv.Role)
	}
}
