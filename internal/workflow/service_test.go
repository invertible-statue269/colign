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

	// Design with design done → should advance
	assert.True(t, gate.AllMet(models.StageDesign, GateInput{
		HasDesign: true,
	}), "should auto-advance from Design when design complete")

	// Review with approvals done → should advance
	assert.True(t, gate.AllMet(models.StageReview, GateInput{
		ApprovalsNeeded: 2,
		ApprovalsDone:   2,
	}), "should auto-advance from Review when approvals met")

	// Review without enough approvals → should not advance
	assert.False(t, gate.AllMet(models.StageReview, GateInput{
		ApprovalsNeeded: 2,
		ApprovalsDone:   1,
	}), "should not auto-advance from Review with insufficient approvals")
}
