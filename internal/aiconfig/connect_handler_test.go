package aiconfig

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	aiconfigv1 "github.com/gobenpark/colign/gen/proto/aiconfig/v1"
	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newHandlerTestDB creates a bun.DB with m2m model registration for handler tests.
func newHandlerTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := bun.NewDB(sqlDB, pgdialect.New())
	db.RegisterModel((*models.ProjectLabelAssignment)(nil))
	db.RegisterModel((*models.ChangeLabelAssignment)(nil))
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db, mock
}

func makeAuthHeader(t *testing.T, jwtManager *auth.JWTManager, orgID int64) string {
	t.Helper()
	tok, err := jwtManager.GenerateAccessToken(1, "user@example.com", "Test User", orgID)
	require.NoError(t, err)
	return "Bearer " + tok
}

func reqWithAuth[T any](msg *T, authHeader string) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", authHeader)
	return req
}

const testOrgID = int64(10)

// ---------------------------------------------------------------------------
// GetAIConfig
// ---------------------------------------------------------------------------

func TestGetAIConfig_Unauthenticated(t *testing.T) {
	db, _ := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	req := connect.NewRequest(&aiconfigv1.GetAIConfigRequest{ProjectId: 1})
	// No Authorization header

	_, err := h.GetAIConfig(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
}

func TestGetAIConfig_ProjectNotInOrg(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.GetAIConfigRequest{ProjectId: 1}, authHeader)

	// Project ownership check returns no rows (wrong org)
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := h.GetAIConfig(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAIConfig_NoConfig(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.GetAIConfigRequest{ProjectId: 1}, authHeader)

	// Project ownership check — project exists in org
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// GetByProjectID — no config found
	mock.ExpectQuery(`SELECT .+ FROM "ai_configs"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "provider", "model", "api_key_encrypted", "key_version", "include_project_context", "created_at", "updated_at"}))

	resp, err := h.GetAIConfig(context.Background(), req)
	require.NoError(t, err)
	assert.Nil(t, resp.Msg.Config)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAIConfig_Found(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.GetAIConfigRequest{ProjectId: 1}, authHeader)

	encKey, err := Encrypt("sk-test-key-1234567890", svcKey, 1)
	require.NoError(t, err)

	now := time.Now()

	// Project ownership check
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// GetByProjectID
	mock.ExpectQuery(`SELECT .+ FROM "ai_configs"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "project_id", "provider", "model", "api_key_encrypted", "key_version", "include_project_context", "created_at", "updated_at"}).
				AddRow(5, 1, "anthropic", "claude-3-5-sonnet", encKey, 1, true, now, now),
		)

	resp, err := h.GetAIConfig(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Config)
	assert.Equal(t, int64(5), resp.Msg.Config.Id)
	assert.Equal(t, int64(1), resp.Msg.Config.ProjectId)
	assert.Equal(t, "anthropic", resp.Msg.Config.Provider)
	assert.Equal(t, "claude-3-5-sonnet", resp.Msg.Config.Model)
	// Verify the API key is masked, not plaintext
	assert.Contains(t, resp.Msg.Config.ApiKeyMasked, "...")
	assert.True(t, resp.Msg.Config.IncludeProjectContext)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// SaveAIConfig
// ---------------------------------------------------------------------------

func TestSaveAIConfig_Unauthenticated(t *testing.T) {
	db, _ := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	req := connect.NewRequest(&aiconfigv1.SaveAIConfigRequest{ProjectId: 1})

	_, err := h.SaveAIConfig(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
}

func TestSaveAIConfig_ProjectNotInOrg(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.SaveAIConfigRequest{ProjectId: 1, Provider: "openai", Model: "gpt-4"}, authHeader)

	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := h.SaveAIConfig(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveAIConfig_Success(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.SaveAIConfigRequest{
		ProjectId:             1,
		Provider:              "anthropic",
		Model:                 "claude-3-5-sonnet",
		ApiKey:                "sk-ant-test-key-1234567890abcdef",
		IncludeProjectContext: true,
	}, authHeader)

	now := time.Now()

	// Project ownership check
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Upsert INSERT
	mock.ExpectQuery(`INSERT INTO "ai_configs"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(7, now, now),
		)

	resp, err := h.SaveAIConfig(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Config)
	assert.Equal(t, int64(7), resp.Msg.Config.Id)
	assert.Equal(t, int64(1), resp.Msg.Config.ProjectId)
	assert.Equal(t, "anthropic", resp.Msg.Config.Provider)
	assert.Contains(t, resp.Msg.Config.ApiKeyMasked, "...")
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// TestConnection
// ---------------------------------------------------------------------------

func TestTestConnection_NilFunc(t *testing.T) {
	db, _ := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	req := connect.NewRequest(&aiconfigv1.TestConnectionRequest{
		Provider: "anthropic",
		Model:    "claude-3-5-sonnet",
		ApiKey:   "sk-ant-test",
	})
	req.Header().Set("Authorization", "Bearer anything")

	_, err := h.TestConnection(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeUnimplemented, connectErr.Code())
}

func TestTestConnection_Success(t *testing.T) {
	db, _ := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, func(_ context.Context, _, _, _ string) error {
		return nil
	})

	req := connect.NewRequest(&aiconfigv1.TestConnectionRequest{Provider: "openai", Model: "gpt-4o", ApiKey: "sk-test"})
	resp, err := h.TestConnection(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Msg.Success)
}

func TestTestConnection_Failure(t *testing.T) {
	db, _ := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, func(_ context.Context, _, _, _ string) error {
		return errors.New("invalid API key")
	})

	req := connect.NewRequest(&aiconfigv1.TestConnectionRequest{Provider: "openai", Model: "gpt-4o", ApiKey: "sk-bad"})
	resp, err := h.TestConnection(context.Background(), req)
	require.NoError(t, err)
	assert.False(t, resp.Msg.Success)
	assert.Contains(t, resp.Msg.Error, "invalid API key")
}

// ---------------------------------------------------------------------------
// DeleteAIConfig
// ---------------------------------------------------------------------------

func TestDeleteAIConfig_Unauthenticated(t *testing.T) {
	db, _ := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	req := connect.NewRequest(&aiconfigv1.DeleteAIConfigRequest{ProjectId: 1})

	_, err := h.DeleteAIConfig(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
}

func TestDeleteAIConfig_ProjectNotInOrg(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.DeleteAIConfigRequest{ProjectId: 1}, authHeader)

	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	_, err := h.DeleteAIConfig(context.Background(), req)
	require.Error(t, err)
	var connectErr *connect.Error
	require.True(t, errors.As(err, &connectErr))
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAIConfig_Success(t *testing.T) {
	db, mock := newHandlerTestDB(t)
	svc := NewService(db, svcKey)
	jwtManager := auth.NewJWTManager("test-secret")
	h := NewConnectHandler(svc, jwtManager, nil, nil)

	authHeader := makeAuthHeader(t, jwtManager, testOrgID)
	req := reqWithAuth(&aiconfigv1.DeleteAIConfigRequest{ProjectId: 1}, authHeader)

	// Project ownership check
	mock.ExpectQuery(`SELECT`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Delete
	mock.ExpectExec(`DELETE FROM "ai_configs"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.DeleteAIConfig(context.Background(), req)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// configToProto helper
// ---------------------------------------------------------------------------

func TestConfigToProto_MasksAPIKey(t *testing.T) {
	key, err := Encrypt("sk-test-key-1234567890abcdef", svcKey, 1)
	require.NoError(t, err)

	decrypted := "sk-test-key-1234567890abcdef"
	proto := configToProto(&AIConfig{
		ID:                    1,
		ProjectID:             2,
		Provider:              "anthropic",
		Model:                 "claude-3-5-sonnet",
		APIKeyEncrypted:       key,
		IncludeProjectContext: true,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}, decrypted)

	assert.Equal(t, int64(1), proto.Id)
	assert.Equal(t, int64(2), proto.ProjectId)
	assert.Equal(t, "anthropic", proto.Provider)
	assert.Equal(t, "claude-3-5-sonnet", proto.Model)
	// API key must be masked
	assert.NotEqual(t, decrypted, proto.ApiKeyMasked)
	assert.Contains(t, proto.ApiKeyMasked, "...")
	assert.True(t, proto.IncludeProjectContext)
	assert.NotNil(t, proto.CreatedAt)
	assert.NotNil(t, proto.UpdatedAt)
}

// verifyInterfaceSatisfied_ is a compile-time check.
var _ = func() {
	var _ interface {
		GetAIConfig(context.Context, *connect.Request[aiconfigv1.GetAIConfigRequest]) (*connect.Response[aiconfigv1.GetAIConfigResponse], error)
		SaveAIConfig(context.Context, *connect.Request[aiconfigv1.SaveAIConfigRequest]) (*connect.Response[aiconfigv1.SaveAIConfigResponse], error)
		TestConnection(context.Context, *connect.Request[aiconfigv1.TestConnectionRequest]) (*connect.Response[aiconfigv1.TestConnectionResponse], error)
		DeleteAIConfig(context.Context, *connect.Request[aiconfigv1.DeleteAIConfigRequest]) (*connect.Response[aiconfigv1.DeleteAIConfigResponse], error)
	} = (*ConnectHandler)(nil)
	_ = http.MethodGet // avoid unused import
}
