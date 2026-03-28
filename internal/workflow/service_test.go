package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gobenpark/colign/internal/models"
)

func TestShouldAutoAdvance(t *testing.T) {
	gate := NewGateChecker()

	// Draft with proposal done → should advance
	assert.True(t, gate.AllMet(models.StageDraft, GateInput{HasProposal: true}),
		"should auto-advance from Draft when proposal exists")

	// Spec with spec and approvals done → should advance
	assert.True(t, gate.AllMet(models.StageSpec, GateInput{
		HasDesign:       true,
		ApprovalsNeeded: 2,
		ApprovalsDone:   2,
	}), "should auto-advance from Spec when spec complete and approvals met")

	// Spec without enough approvals → should not advance
	assert.False(t, gate.AllMet(models.StageSpec, GateInput{
		HasDesign:       true,
		ApprovalsNeeded: 2,
		ApprovalsDone:   1,
	}), "should not auto-advance from Spec with insufficient approvals")
}
