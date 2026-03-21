package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gobenpark/colign/internal/models"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from    models.ChangeStage
		to      models.ChangeStage
		allowed bool
	}{
		{models.StageDraft, models.StageDesign, true},
		{models.StageDesign, models.StageReview, true},
		{models.StageReview, models.StageReady, true},
		// Can revert
		{models.StageDesign, models.StageDraft, true},
		{models.StageReview, models.StageDesign, true},
		{models.StageReady, models.StageReview, true},
		// Cannot skip
		{models.StageDraft, models.StageReview, false},
		{models.StageDraft, models.StageReady, false},
		{models.StageDesign, models.StageReady, false},
		// Cannot self-transition
		{models.StageDraft, models.StageDraft, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.allowed, CanTransition(tt.from, tt.to),
			"CanTransition(%s, %s)", tt.from, tt.to)
	}
}

func TestNextStage(t *testing.T) {
	tests := []struct {
		current  models.ChangeStage
		expected models.ChangeStage
		hasNext  bool
	}{
		{models.StageDraft, models.StageDesign, true},
		{models.StageDesign, models.StageReview, true},
		{models.StageReview, models.StageReady, true},
		{models.StageReady, "", false},
	}

	for _, tt := range tests {
		next, ok := NextStage(tt.current)
		assert.Equal(t, tt.hasNext, ok, "NextStage(%s) hasNext", tt.current)
		if ok {
			assert.Equal(t, tt.expected, next, "NextStage(%s)", tt.current)
		}
	}
}

func TestPrevStage(t *testing.T) {
	tests := []struct {
		current  models.ChangeStage
		expected models.ChangeStage
		hasPrev  bool
	}{
		{models.StageDraft, "", false},
		{models.StageDesign, models.StageDraft, true},
		{models.StageReview, models.StageDesign, true},
		{models.StageReady, models.StageReview, true},
	}

	for _, tt := range tests {
		prev, ok := PrevStage(tt.current)
		assert.Equal(t, tt.hasPrev, ok, "PrevStage(%s) hasPrev", tt.current)
		if ok {
			assert.Equal(t, tt.expected, prev, "PrevStage(%s)", tt.current)
		}
	}
}
