package memory

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/gobenpark/colign/internal/models"
)

func setupTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	db := bun.NewDB(mockDB, pgdialect.New())
	db.RegisterModel((*models.ProjectLabelAssignment)(nil))
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestGet_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mem, err := svc.Get(ctx, 1, wrongOrgID)
	require.NoError(t, err)
	require.Nil(t, mem)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	result, err := svc.Save(ctx, 1, "content", 1, wrongOrgID)
	require.ErrorIs(t, err, ErrProjectNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}
