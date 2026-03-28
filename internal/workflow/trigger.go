package workflow

import "github.com/gobenpark/colign/internal/models"

// ShouldTriggerTaskGeneration returns true when a change transitions to Approved stage.
func ShouldTriggerTaskGeneration(from, to models.ChangeStage) bool {
	return from == models.StageSpec && to == models.StageApproved
}
