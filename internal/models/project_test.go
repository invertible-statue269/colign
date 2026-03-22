package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProjectModel(t *testing.T) {
	p := &Project{
		ID:          1,
		Name:        "My App",
		Slug:        "my-app",
		Description: "Test project",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, "My App", p.Name)
	assert.Equal(t, "my-app", p.Slug)
}

func TestProjectMemberModel(t *testing.T) {
	pm := &ProjectMember{
		ID:        1,
		ProjectID: 1,
		UserID:    1,
		Role:      RoleOwner,
	}

	assert.Equal(t, RoleOwner, pm.Role)
}

func TestProjectMemberRoles(t *testing.T) {
	roles := []Role{RoleOwner, RoleEditor, RoleViewer}
	expected := []string{"owner", "editor", "viewer"}

	for i, role := range roles {
		assert.Equal(t, expected[i], string(role))
	}
}
