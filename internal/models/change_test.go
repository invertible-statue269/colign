package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeStages(t *testing.T) {
	stages := []ChangeStage{StageDraft, StageDesign, StageReview, StageReady}
	expected := []string{"draft", "design", "review", "ready"}

	for i, stage := range stages {
		assert.Equal(t, expected[i], string(stage))
	}
}

func TestChangeModel(t *testing.T) {
	c := &Change{
		ProjectID: 1,
		Name:      "add-user-auth",
		Stage:     StageDraft,
	}

	assert.Equal(t, StageDraft, c.Stage)
	assert.Equal(t, "add-user-auth", c.Name)
}

func TestChangeStageOrder(t *testing.T) {
	order := StageOrder()
	require.Len(t, order, 4)
	assert.Equal(t, StageDraft, order[0])
	assert.Equal(t, StageReady, order[3])
}
