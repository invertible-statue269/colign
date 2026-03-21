package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/models"
)

func TestGateConditions_Draft(t *testing.T) {
	gate := NewGateChecker()

	// Draft → Design: proposal must exist
	conditions := gate.Check(models.StageDraft, GateInput{
		HasProposal: false,
		HasDesign:   false,
	})

	require.NotEmpty(t, conditions, "expected conditions for Draft gate")
	assert.False(t, conditions[0].Met, "proposal condition should not be met")

	// With proposal
	conditions = gate.Check(models.StageDraft, GateInput{
		HasProposal: true,
	})
	assert.True(t, conditions[0].Met, "proposal condition should be met")
}

func TestGateConditions_Design(t *testing.T) {
	gate := NewGateChecker()

	// Design → Review: design must exist
	conditions := gate.Check(models.StageDesign, GateInput{
		HasProposal: true,
		HasDesign:   false,
	})

	require.NotEmpty(t, conditions, "expected conditions for Design gate")
	assert.Equal(t, "design", conditions[0].Name)
	assert.False(t, conditions[0].Met, "design condition should not be met")

	// With design
	conditions = gate.Check(models.StageDesign, GateInput{
		HasProposal: true,
		HasDesign:   true,
	})

	for _, c := range conditions {
		assert.True(t, c.Met, "condition %s should be met", c.Name)
	}
}

func TestGateConditions_AllMet(t *testing.T) {
	gate := NewGateChecker()

	assert.True(t, gate.AllMet(models.StageDraft, GateInput{HasProposal: true}),
		"Draft gate should pass with proposal")
	assert.False(t, gate.AllMet(models.StageDraft, GateInput{HasProposal: false}),
		"Draft gate should fail without proposal")
}
