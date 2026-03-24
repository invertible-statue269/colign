package authz

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/casbin/casbin/v2"
	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/auth"
)

// ProjectScoped is implemented by protobuf Request messages that include a project_id field.
type ProjectScoped interface {
	GetProjectId() int64
}

// RBACInterceptor enforces role-based access control on Connect RPCs using Casbin.
type RBACInterceptor struct {
	db                *bun.DB
	enforcer          *casbin.Enforcer
	jwtManager        *auth.JWTManager
	apiTokenValidator auth.APITokenValidator
}

// NewRBACInterceptor creates a new RBAC interceptor.
func NewRBACInterceptor(db *bun.DB, enforcer *casbin.Enforcer, jwtManager *auth.JWTManager, apiTokenValidator auth.APITokenValidator) *RBACInterceptor {
	return &RBACInterceptor{db: db, enforcer: enforcer, jwtManager: jwtManager, apiTokenValidator: apiTokenValidator}
}

// WrapUnary implements connect.Interceptor.
func (i *RBACInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := req.Spec().Procedure

		// Skip RPCs that don't need RBAC
		if IsSkipped(procedure) {
			return next(ctx, req)
		}

		// Look up the auth rule
		rule, ok := GetRule(procedure)
		if !ok {
			slog.Warn("unmapped RPC denied", "procedure", procedure)
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("access denied"))
		}

		// Extract project_id from request
		scoped, ok := req.Any().(ProjectScoped)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("missing project context"))
		}

		projectID := scoped.GetProjectId()
		if projectID == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("project_id is required"))
		}

		// Authenticate: resolve user from Authorization header
		header := req.Header().Get("Authorization")
		claims, err := auth.ResolveFromHeader(i.jwtManager, i.apiTokenValidator, ctx, header)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
		}

		// Look up user's role in the project
		role, err := i.lookupRole(ctx, projectID, claims.UserID)
		if err != nil {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("not a project member"))
		}

		// Enforce with Casbin
		allowed, err := i.enforcer.Enforce(role, rule.Resource, rule.Action)
		if err != nil {
			slog.Error("casbin enforcement error", "error", err, "procedure", procedure)
			return nil, connect.NewError(connect.CodeInternal, errors.New("authorization error"))
		}

		if !allowed {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("insufficient permissions"))
		}

		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor (no-op for streaming).
func (i *RBACInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor (no-op for streaming).
func (i *RBACInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// lookupRole queries project_members to find the user's role in the project.
func (i *RBACInterceptor) lookupRole(ctx context.Context, projectID, userID int64) (string, error) {
	var role string
	err := i.db.NewSelect().
		TableExpr("project_members").
		ColumnExpr("role").
		Where("project_id = ?", projectID).
		Where("user_id = ?", userID).
		Scan(ctx, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("not a member")
		}
		return "", err
	}
	return role, nil
}
