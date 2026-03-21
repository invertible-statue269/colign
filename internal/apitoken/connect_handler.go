package apitoken

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	apitokenv1 "github.com/gobenpark/colign/gen/proto/apitoken/v1"
	"github.com/gobenpark/colign/gen/proto/apitoken/v1/apitokenv1connect"
	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/models"
)

type ConnectHandler struct {
	service    *Service
	jwtManager *auth.JWTManager
}

var _ apitokenv1connect.ApiTokenServiceHandler = (*ConnectHandler)(nil)

func NewConnectHandler(service *Service, jwtManager *auth.JWTManager) *ConnectHandler {
	return &ConnectHandler{service: service, jwtManager: jwtManager}
}

func (h *ConnectHandler) extractClaims(header string) (*auth.Claims, error) {
	claims, err := auth.ExtractClaims(h.jwtManager, header)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	return claims, nil
}

func (h *ConnectHandler) CreateApiToken(ctx context.Context, req *connect.Request[apitokenv1.CreateApiTokenRequest]) (*connect.Response[apitokenv1.CreateApiTokenResponse], error) {
	claims, err := h.extractClaims(req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	token, rawToken, err := h.service.Create(ctx, claims.UserID, claims.OrgID, req.Msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&apitokenv1.CreateApiTokenResponse{
		Token:    tokenToProto(token),
		RawToken: rawToken,
	}), nil
}

func (h *ConnectHandler) ListApiTokens(ctx context.Context, req *connect.Request[apitokenv1.ListApiTokensRequest]) (*connect.Response[apitokenv1.ListApiTokensResponse], error) {
	claims, err := h.extractClaims(req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	tokens, err := h.service.List(ctx, claims.UserID, claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoTokens := make([]*apitokenv1.ApiToken, len(tokens))
	for i, t := range tokens {
		protoTokens[i] = tokenToProto(&t)
	}

	return connect.NewResponse(&apitokenv1.ListApiTokensResponse{
		Tokens: protoTokens,
	}), nil
}

func (h *ConnectHandler) DeleteApiToken(ctx context.Context, req *connect.Request[apitokenv1.DeleteApiTokenRequest]) (*connect.Response[apitokenv1.DeleteApiTokenResponse], error) {
	claims, err := h.extractClaims(req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	if err := h.service.Delete(ctx, claims.UserID, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&apitokenv1.DeleteApiTokenResponse{}), nil
}

func tokenToProto(t *models.APIToken) *apitokenv1.ApiToken {
	proto := &apitokenv1.ApiToken{
		Id:        t.ID,
		Name:      t.Name,
		Prefix:    t.Prefix,
		CreatedAt: timestamppb.New(t.CreatedAt),
	}
	if t.LastUsedAt != nil {
		proto.LastUsedAt = timestamppb.New(*t.LastUsedAt)
	}
	return proto
}
