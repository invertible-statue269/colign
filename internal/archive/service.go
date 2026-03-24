package archive

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

var (
	ErrChangeNotFound  = errors.New("change not found")
	ErrNotReady        = errors.New("change is not in ready stage")
	ErrAlreadyArchived = errors.New("change is already archived")
	ErrNotArchived     = errors.New("change is not archived")
)

// Service handles archiving and unarchiving of changes.
type Service struct {
	db *bun.DB
}

// NewService creates a new archive service.
func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

// Archive sets archived_at on a change. Only allowed for changes in the Ready
// stage that are not already archived. Records a WorkflowEvent.
func (s *Service) Archive(ctx context.Context, changeID, userID int64, orgID int64) (*models.Change, error) {
	change := new(models.Change)
	err := s.db.NewSelect().Model(change).
		Join("JOIN projects AS p ON p.id = ch.project_id").
		Where("ch.id = ?", changeID).
		Where("p.organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	if change.Stage != models.StageReady {
		return nil, ErrNotReady
	}

	if change.ArchivedAt != nil {
		return nil, ErrAlreadyArchived
	}

	now := time.Now()
	_, err = s.db.NewUpdate().Model((*models.Change)(nil)).
		Set("archived_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", changeID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	change.ArchivedAt = &now
	change.UpdatedAt = now

	event := &models.WorkflowEvent{
		ChangeID:  changeID,
		FromStage: string(change.Stage),
		ToStage:   string(change.Stage),
		Action:    "archive",
		Reason:    "manually archived",
		UserID:    userID,
	}
	if _, err := s.db.NewInsert().Model(event).Exec(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to record archive workflow event", "error", err, "change_id", changeID)
	}

	return change, nil
}

// Unarchive clears archived_at on a change. Only allowed for archived changes.
// Records a WorkflowEvent.
func (s *Service) Unarchive(ctx context.Context, changeID, userID int64, orgID int64) (*models.Change, error) {
	change := new(models.Change)
	err := s.db.NewSelect().Model(change).
		Join("JOIN projects AS p ON p.id = ch.project_id").
		Where("ch.id = ?", changeID).
		Where("p.organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	if change.ArchivedAt == nil {
		return nil, ErrNotArchived
	}

	now := time.Now()
	_, err = s.db.NewUpdate().Model((*models.Change)(nil)).
		Set("archived_at = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", changeID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	change.ArchivedAt = nil
	change.UpdatedAt = now

	event := &models.WorkflowEvent{
		ChangeID:  changeID,
		FromStage: string(change.Stage),
		ToStage:   string(change.Stage),
		Action:    "unarchive",
		Reason:    "manually unarchived",
		UserID:    userID,
	}
	if _, err := s.db.NewInsert().Model(event).Exec(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to record unarchive workflow event", "error", err, "change_id", changeID)
	}

	return change, nil
}

// GetPolicy returns the archive policy for a project. If none exists, returns
// a default policy with manual mode.
func (s *Service) GetPolicy(ctx context.Context, projectID int64, orgID int64) (*models.ArchivePolicy, error) {
	// Verify project belongs to org
	exists, err := s.db.NewSelect().Model((*models.Project)(nil)).
		Where("id = ?", projectID).
		Where("organization_id = ?", orgID).
		Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrChangeNotFound
	}

	policy := new(models.ArchivePolicy)
	err = s.db.NewSelect().Model(policy).Where("project_id = ?", projectID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.ArchivePolicy{
				ProjectID:   projectID,
				Mode:        models.ArchiveModeManual,
				TriggerType: models.TriggerTasksDone,
				DaysDelay:   0,
			}, nil
		}
		return nil, err
	}
	return policy, nil
}

// UpdatePolicy upserts an archive policy using ON CONFLICT (project_id) DO UPDATE.
func (s *Service) UpdatePolicy(ctx context.Context, policy *models.ArchivePolicy, orgID int64) (*models.ArchivePolicy, error) {
	// Verify project belongs to org
	exists, err := s.db.NewSelect().Model((*models.Project)(nil)).
		Where("id = ?", policy.ProjectID).
		Where("organization_id = ?", orgID).
		Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrChangeNotFound
	}

	now := time.Now()
	policy.UpdatedAt = now
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}

	_, err = s.db.NewInsert().Model(policy).
		On("CONFLICT (project_id) DO UPDATE").
		Set("mode = EXCLUDED.mode").
		Set("trigger_type = EXCLUDED.trigger_type").
		Set("days_delay = EXCLUDED.days_delay").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

// getPolicy returns the archive policy for a project without org verification.
// Used internally by system-level operations (cron, auto-archive).
func (s *Service) getPolicy(ctx context.Context, projectID int64) (*models.ArchivePolicy, error) {
	policy := new(models.ArchivePolicy)
	err := s.db.NewSelect().Model(policy).Where("project_id = ?", projectID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.ArchivePolicy{
				ProjectID:   projectID,
				Mode:        models.ArchiveModeManual,
				TriggerType: models.TriggerTasksDone,
				DaysDelay:   0,
			}, nil
		}
		return nil, err
	}
	return policy, nil
}

// EvaluateAutoArchive checks whether a change should be auto-archived based on
// the project's archive policy. Returns true if the change was archived.
func (s *Service) EvaluateAutoArchive(ctx context.Context, changeID int64) (bool, error) {
	change := new(models.Change)
	err := s.db.NewSelect().Model(change).Where("id = ?", changeID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrChangeNotFound
		}
		return false, err
	}

	// Only auto-archive changes in Ready stage that are not yet archived
	if change.Stage != models.StageReady || change.ArchivedAt != nil {
		return false, nil
	}

	policy, err := s.getPolicy(ctx, change.ProjectID)
	if err != nil {
		return false, err
	}

	if policy.Mode != models.ArchiveModeAuto {
		return false, nil
	}

	tasksDone, err := s.allTasksDone(ctx, changeID)
	if err != nil {
		return false, err
	}

	readyAt, err := s.getReadyTimestamp(ctx, changeID)
	if err != nil {
		return false, err
	}

	if !shouldAutoArchive(policy.TriggerType, policy.DaysDelay, tasksDone, readyAt) {
		return false, nil
	}

	// Archive the change
	now := time.Now()
	_, err = s.db.NewUpdate().Model((*models.Change)(nil)).
		Set("archived_at = ?", now).
		Set("updated_at = ?", now).
		Where("id = ?", changeID).
		Exec(ctx)
	if err != nil {
		return false, err
	}

	event := &models.WorkflowEvent{
		ChangeID:  changeID,
		FromStage: string(change.Stage),
		ToStage:   string(change.Stage),
		Action:    "archive",
		Reason:    "auto-archived by policy",
	}
	if _, err := s.db.NewInsert().Model(event).Exec(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to record auto-archive workflow event", "error", err, "change_id", changeID)
	}

	return true, nil
}

// shouldAutoArchive is a pure function that evaluates whether a change should
// be auto-archived based on the trigger type, delay, task status, and ready timestamp.
func shouldAutoArchive(trigger models.ArchiveTrigger, daysDelay int, allTasksDone bool, readyAt *time.Time) bool {
	switch trigger {
	case models.TriggerTasksDone:
		return allTasksDone
	case models.TriggerDaysAfterReady:
		if readyAt == nil {
			return false
		}
		return readyAt.Add(time.Duration(daysDelay) * 24 * time.Hour).Before(time.Now())
	case models.TriggerTasksDoneAndDays:
		if readyAt == nil {
			return false
		}
		return allTasksDone && readyAt.Add(time.Duration(daysDelay)*24*time.Hour).Before(time.Now())
	default:
		return false
	}
}

// allTasksDone checks whether all tasks for a change have status "done".
// Returns false if no tasks exist.
func (s *Service) allTasksDone(ctx context.Context, changeID int64) (bool, error) {
	totalCount, err := s.db.NewSelect().
		TableExpr("tasks").
		Where("change_id = ?", changeID).
		Count(ctx)
	if err != nil {
		return false, err
	}

	if totalCount == 0 {
		return false, nil
	}

	doneCount, err := s.db.NewSelect().
		TableExpr("tasks").
		Where("change_id = ?", changeID).
		Where("status = ?", models.TaskDone).
		Count(ctx)
	if err != nil {
		return false, err
	}

	return totalCount == doneCount, nil
}

// getReadyTimestamp returns the timestamp of the most recent workflow event
// where the change transitioned to the "ready" stage.
func (s *Service) getReadyTimestamp(ctx context.Context, changeID int64) (*time.Time, error) {
	event := new(models.WorkflowEvent)
	err := s.db.NewSelect().Model(event).
		Where("change_id = ?", changeID).
		Where("to_stage = ?", "ready").
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &event.CreatedAt, nil
}
