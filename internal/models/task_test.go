package models

import "testing"

func TestTaskStatuses(t *testing.T) {
	statuses := []TaskStatus{TaskTodo, TaskInProgress, TaskDone}
	expected := []string{"todo", "in_progress", "done"}

	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], s)
		}
	}
}
