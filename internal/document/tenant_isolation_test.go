package document

import (
	"context"
	"database/sql"
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
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestGet_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	doc, err := svc.Get(ctx, 1, models.DocumentType("proposal"), wrongOrgID)
	require.NoError(t, err)
	require.Nil(t, doc)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	result, err := svc.Save(ctx, SaveInput{ChangeID: 1, Type: "proposal"}, wrongOrgID)
	require.ErrorIs(t, err, ErrDocumentNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRestore_CrossTenantBlocked(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)
	ctx := context.Background()

	wrongOrgID := int64(999)

	// First query: fetch DocumentVersion by documentID + version
	dvRows := sqlmock.NewRows([]string{"id", "document_id", "content", "version", "user_id"}).
		AddRow(int64(1), int64(1), "old content", 1, int64(1))
	mock.ExpectQuery("SELECT").WillReturnRows(dvRows)

	// Second query: fetch Document by id
	docRows := sqlmock.NewRows([]string{"id", "change_id", "type", "title", "content", "version"}).
		AddRow(int64(1), int64(1), "proposal", "Title", "current content", 2)
	mock.ExpectQuery("SELECT").WillReturnRows(docRows)

	// Third query: EXISTS check for org ownership — returns false
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	result, err := svc.Restore(ctx, 1, 1, 1, wrongOrgID)
	require.ErrorIs(t, err, ErrDocumentNotFound)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}
