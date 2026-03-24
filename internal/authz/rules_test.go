package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRule(t *testing.T) {
	rule, ok := GetRule("/task.v1.TaskService/CreateTask")
	assert.True(t, ok)
	assert.Equal(t, "task", rule.Resource)
	assert.Equal(t, "create", rule.Action)
}

func TestGetRule_NotFound(t *testing.T) {
	_, ok := GetRule("/unknown.v1.Service/Method")
	assert.False(t, ok)
}

func TestIsSkipped(t *testing.T) {
	assert.True(t, IsSkipped("/auth.v1.AuthService/Login"))
	assert.True(t, IsSkipped("/project.v1.ProjectService/CreateProject"))
	assert.False(t, IsSkipped("/task.v1.TaskService/CreateTask"))
}

func TestAllRPCsCovered(t *testing.T) {
	// Verify no RPC is in both maps
	for rpc := range rpcRules {
		assert.False(t, skipRPCs[rpc], "RPC %s is in both rpcRules and skipRPCs", rpc)
	}
}
