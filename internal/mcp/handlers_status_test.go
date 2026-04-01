package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/models"
)

func TestParseCompositeChangeStatus(t *testing.T) {
	tests := []struct {
		raw       string
		stage     models.ChangeStage
		subStatus models.SubStatus
	}{
		{raw: "draft(in_progress)", stage: models.StageDraft, subStatus: models.SubStatusInProgress},
		{raw: "draft(ready)", stage: models.StageDraft, subStatus: models.SubStatusReady},
		{raw: "spec(in_progress)", stage: models.StageSpec, subStatus: models.SubStatusInProgress},
		{raw: "spec(ready)", stage: models.StageSpec, subStatus: models.SubStatusReady},
	}

	for _, tt := range tests {
		stage, subStatus, err := parseCompositeChangeStatus(tt.raw)
		require.NoError(t, err, tt.raw)
		assert.Equal(t, tt.stage, stage)
		assert.Equal(t, tt.subStatus, subStatus)
	}
}

func TestParseCompositeChangeStatusRejectsInvalidValues(t *testing.T) {
	_, _, err := parseCompositeChangeStatus("draft")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestParseCompositeChangeStatusRejectsApproved(t *testing.T) {
	_, _, err := parseCompositeChangeStatus("approved")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use approve_change")
}

func TestFormatCompositeChangeStatus(t *testing.T) {
	assert.Equal(t, "draft(ready)", formatCompositeChangeStatus(models.StageDraft, models.SubStatusReady))
	assert.Equal(t, "approved", formatCompositeChangeStatus(models.StageApproved, models.SubStatusInProgress))
}
