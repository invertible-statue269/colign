package archive

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/models"
)

func TestArchive_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	result, err := svc.Archive(ctx, 1, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrChangeNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUnarchive_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	result, err := svc.Unarchive(ctx, 1, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrChangeNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPolicy_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	result, err := svc.GetPolicy(ctx, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrChangeNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdatePolicy_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	result, err := svc.UpdatePolicy(ctx, &models.ArchivePolicy{ProjectID: 1}, wrongOrgID)
	require.ErrorIs(t, err, ErrChangeNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}
