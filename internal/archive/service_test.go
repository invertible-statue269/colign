package archive

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/gobenpark/colign/internal/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	db := bun.NewDB(mockDB, pgdialect.New())
	db.RegisterModel((*models.ProjectLabelAssignment)(nil))
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

// ---------------------------------------------------------------------------
// shouldAutoArchive — pure function tests (no DB)
// ---------------------------------------------------------------------------

func TestShouldAutoArchive(t *testing.T) {
	now := time.Now()
	readyAt := now.Add(-48 * time.Hour) // 2 days ago

	t.Run("tasks_done: true when all tasks done", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerTasksDone, 0, true, &readyAt)
		assert.True(t, result)
	})

	t.Run("tasks_done: false when tasks not done", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerTasksDone, 0, false, &readyAt)
		assert.False(t, result)
	})

	t.Run("days_after_ready: true when enough days passed", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerDaysAfterReady, 1, false, &readyAt)
		assert.True(t, result)
	})

	t.Run("days_after_ready: false when not enough days passed", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerDaysAfterReady, 5, false, &readyAt)
		assert.False(t, result)
	})

	t.Run("days_after_ready: false when readyAt is nil", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerDaysAfterReady, 1, false, nil)
		assert.False(t, result)
	})

	t.Run("tasks_done_and_days: true when both conditions met", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerTasksDoneAndDays, 1, true, &readyAt)
		assert.True(t, result)
	})

	t.Run("tasks_done_and_days: false when only tasks done", func(t *testing.T) {
		recentReady := now.Add(-1 * time.Hour) // 1 hour ago
		result := shouldAutoArchive(models.TriggerTasksDoneAndDays, 5, true, &recentReady)
		assert.False(t, result)
	})

	t.Run("tasks_done_and_days: false when only days passed", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerTasksDoneAndDays, 1, false, &readyAt)
		assert.False(t, result)
	})

	t.Run("tasks_done_and_days: false when readyAt is nil", func(t *testing.T) {
		result := shouldAutoArchive(models.TriggerTasksDoneAndDays, 1, true, nil)
		assert.False(t, result)
	})

	t.Run("unknown trigger: returns false", func(t *testing.T) {
		result := shouldAutoArchive(models.ArchiveTrigger("unknown"), 0, true, &readyAt)
		assert.False(t, result)
	})
}

// ---------------------------------------------------------------------------
// Archive — sqlmock tests
// ---------------------------------------------------------------------------

func TestArchive(t *testing.T) {
	t.Run("error when change not found", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		mock.ExpectQuery("SELECT").
			WillReturnError(sql.ErrNoRows)

		result, err := svc.Archive(ctx, 999, 1, 1)
		assert.Nil(t, result)
		require.ErrorIs(t, err, ErrChangeNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when change not in ready stage", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "change_type", "created_at", "updated_at", "archived_at"}).
			AddRow(int64(1), int64(1), "Test Change", "design", "feature", time.Now(), time.Now(), nil)
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, err := svc.Archive(ctx, 1, 1, 1)
		assert.Nil(t, result)
		require.ErrorIs(t, err, ErrNotReady)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when already archived", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		archivedAt := time.Now()
		rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "change_type", "created_at", "updated_at", "archived_at"}).
			AddRow(int64(1), int64(1), "Test Change", "ready", "feature", time.Now(), time.Now(), archivedAt)
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, err := svc.Archive(ctx, 1, 1, 1)
		assert.Nil(t, result)
		require.ErrorIs(t, err, ErrAlreadyArchived)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success archives ready change", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "change_type", "created_at", "updated_at", "archived_at"}).
			AddRow(int64(1), int64(1), "Test Change", "ready", "feature", time.Now(), time.Now(), nil)
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		// UPDATE archived_at
		mock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// INSERT workflow event
		mock.ExpectQuery("INSERT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

		result, err := svc.Archive(ctx, 1, 1, 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotNil(t, result.ArchivedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------------------------------------------------------------------
// Unarchive — sqlmock tests
// ---------------------------------------------------------------------------

func TestUnarchive(t *testing.T) {
	t.Run("error when change not found", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		mock.ExpectQuery("SELECT").
			WillReturnError(sql.ErrNoRows)

		result, err := svc.Unarchive(ctx, 999, 1, 1)
		assert.Nil(t, result)
		require.ErrorIs(t, err, ErrChangeNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error when change not archived", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "change_type", "created_at", "updated_at", "archived_at"}).
			AddRow(int64(1), int64(1), "Test Change", "ready", "feature", time.Now(), time.Now(), nil)
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		result, err := svc.Unarchive(ctx, 1, 1, 1)
		assert.Nil(t, result)
		require.ErrorIs(t, err, ErrNotArchived)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success unarchives archived change", func(t *testing.T) {
		db, mock := setupTestDB(t)
		svc := NewService(db)
		ctx := context.Background()

		archivedAt := time.Now()
		rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "change_type", "created_at", "updated_at", "archived_at"}).
			AddRow(int64(1), int64(1), "Test Change", "ready", "feature", time.Now(), time.Now(), archivedAt)
		mock.ExpectQuery("SELECT").
			WillReturnRows(rows)

		// UPDATE archived_at = NULL
		mock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// INSERT workflow event
		mock.ExpectQuery("INSERT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

		result, err := svc.Unarchive(ctx, 1, 1, 1)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.ArchivedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
