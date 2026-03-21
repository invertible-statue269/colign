package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gobenpark/colign/internal/models"
)

func TestShouldTriggerTaskGeneration(t *testing.T) {
	tests := []struct {
		from    models.ChangeStage
		to      models.ChangeStage
		trigger bool
	}{
		{models.StageReview, models.StageReady, true},
		{models.StageDraft, models.StageDesign, false},
		{models.StageDesign, models.StageReview, false},
		{models.StageReady, models.StageReview, false}, // revert
	}

	for _, tt := range tests {
		assert.Equal(t, tt.trigger, ShouldTriggerTaskGeneration(tt.from, tt.to),
			"ShouldTriggerTaskGeneration(%s, %s)", tt.from, tt.to)
	}
}
