package project

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
// Cross-tenant isolation tests
// ---------------------------------------------------------------------------

const wrongOrgID int64 = 999

func TestUpdate_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// SELECT project WHERE id=? AND organization_id=? returns no rows
	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.Update(ctx, UpdateProjectInput{ID: 1, Name: "hacked"}, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// DELETE WHERE id=? AND organization_id=? affects 0 rows
	mock.ExpectExec("DELETE").
		WillReturnResult(driver.RowsAffected(0))

	err := svc.Delete(ctx, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateChange_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// SELECT EXISTS (...) returns false — project not in org
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := svc.CreateChange(ctx, 1, "test", wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetChange_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// SELECT with JOIN returns no rows
	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetChange(ctx, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteChange_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// DELETE WHERE id=? AND project_id IN (SELECT ...) affects 0 rows
	mock.ExpectExec("DELETE").
		WillReturnResult(driver.RowsAffected(0))

	err := svc.DeleteChange(ctx, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListChanges_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// SELECT EXISTS (...) returns false — project not in org
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := svc.ListChanges(ctx, 1, "active", wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignLabel_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// First SELECT EXISTS for project — returns false
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := svc.AssignLabel(ctx, 1, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRemoveLabel_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	// SELECT EXISTS for project — returns false
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := svc.RemoveLabel(ctx, 1, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
