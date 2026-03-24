package acceptance

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func setupTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	db := bun.NewDB(mockDB, pgdialect.New())
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestACUpdate_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	// SELECT with JOINs returns no rows — wrong org
	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	result, err := svc.Update(ctx, 1, UpdateInput{Scenario: "test"}, wrongOrgID)
	assert.Nil(t, result)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestACToggle_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	// SELECT with JOINs returns no rows — wrong org
	mock.ExpectQuery("SELECT").
		WillReturnError(sql.ErrNoRows)

	result, err := svc.Toggle(ctx, 1, true, wrongOrgID)
	assert.Nil(t, result)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestACDelete_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	// DELETE returns 0 rows affected — wrong org
	mock.ExpectExec("DELETE").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.Delete(ctx, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
