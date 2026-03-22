package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskStatuses(t *testing.T) {
	statuses := []TaskStatus{TaskTodo, TaskInProgress, TaskDone}
	expected := []string{"todo", "in_progress", "done"}

	for i, s := range statuses {
		assert.Equal(t, expected[i], string(s))
	}
}
