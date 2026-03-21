package organization

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	organizationv1 "github.com/gobenpark/colign/gen/proto/organization/v1"
	"github.com/gobenpark/colign/gen/proto/organization/v1/organizationv1connect"
	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/models"
)

type ConnectHandler struct {
	service           *Service
	jwtManager        *auth.JWTManager
	apiTokenValidator auth.APITokenValidator
}

var _ organizationv1connect.OrganizationServiceHandler = (*ConnectHandler)(nil)

func NewConnectHandler(service *Service, jwtManager *auth.JWTManager, apiTokenValidator auth.APITokenValidator) *ConnectHandler {
	return &ConnectHandler{service: service, jwtManager: jwtManager, apiTokenValidator: apiTokenValidator}
}

func (h *ConnectHandler) extractClaims(ctx context.Context, header string) (*auth.Claims, error) {
	claims, err := auth.ResolveFromHeader(h.jwtManager, h.apiTokenValidator, ctx, header)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	return claims, nil
}

func (h *ConnectHandler) ListOrganizations(ctx context.Context, req *connect.Request[organizationv1.ListOrganizationsRequest]) (*connect.Response[organizationv1.ListOrganizationsResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	orgs, err := h.service.ListByUser(ctx, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoOrgs := make([]*organizationv1.Organization, len(orgs))
	for i, o := range orgs {
		protoOrgs[i] = orgToProto(&o)
	}

	return connect.NewResponse(&organizationv1.ListOrganizationsResponse{
		Organizations: protoOrgs,
		CurrentOrgId:  claims.OrgID,
	}), nil
}

func (h *ConnectHandler) SwitchOrganization(ctx context.Context, req *connect.Request[organizationv1.SwitchOrganizationRequest]) (*connect.Response[organizationv1.SwitchOrganizationResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	// Verify user is member of target org
	isMember, err := h.service.IsMember(ctx, req.Msg.OrganizationId, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !isMember {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("not a member of this organization"))
	}

	org, err := h.service.GetByID(ctx, req.Msg.OrganizationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// Generate new token pair with new org_id
	tokenPair, err := h.jwtManager.GenerateTokenPair(claims.UserID, claims.Email, claims.Name, req.Msg.OrganizationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&organizationv1.SwitchOrganizationResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		Organization: orgToProto(org),
	}), nil
}

func (h *ConnectHandler) UpdateOrganization(ctx context.Context, req *connect.Request[organizationv1.UpdateOrganizationRequest]) (*connect.Response[organizationv1.UpdateOrganizationResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	// Only allow updating the current org
	if req.Msg.Id != claims.OrgID {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("can only update current organization"))
	}

	org, err := h.service.Update(ctx, req.Msg.Id, req.Msg.Name)
	if err != nil {
		if errors.Is(err, ErrOrgNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&organizationv1.UpdateOrganizationResponse{
		Organization: orgToProto(org),
	}), nil
}

func (h *ConnectHandler) ListMembers(ctx context.Context, req *connect.Request[organizationv1.ListMembersRequest]) (*connect.Response[organizationv1.ListMembersResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	members, err := h.service.ListMembers(ctx, claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoMembers := make([]*organizationv1.OrganizationMember, len(members))
	for i, m := range members {
		protoMembers[i] = memberToProto(&m)
	}

	return connect.NewResponse(&organizationv1.ListMembersResponse{
		Members: protoMembers,
	}), nil
}

func (h *ConnectHandler) InviteOrgMember(ctx context.Context, req *connect.Request[organizationv1.InviteOrgMemberRequest]) (*connect.Response[organizationv1.InviteOrgMemberResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	role := models.OrgRole(req.Msg.Role)
	if role == "" {
		role = "member"
	}

	member, err := h.service.InviteMember(ctx, claims.OrgID, req.Msg.Email, role)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&organizationv1.InviteOrgMemberResponse{
		Member: memberToProto(member),
	}), nil
}

func (h *ConnectHandler) RemoveOrgMember(ctx context.Context, req *connect.Request[organizationv1.RemoveOrgMemberRequest]) (*connect.Response[organizationv1.RemoveOrgMemberResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	if req.Msg.UserId == claims.UserID {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("cannot remove yourself"))
	}

	if err := h.service.RemoveMember(ctx, claims.OrgID, req.Msg.UserId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&organizationv1.RemoveOrgMemberResponse{}), nil
}

func (h *ConnectHandler) UpdateOrgMemberRole(ctx context.Context, req *connect.Request[organizationv1.UpdateOrgMemberRoleRequest]) (*connect.Response[organizationv1.UpdateOrgMemberRoleResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	member, err := h.service.UpdateMemberRole(ctx, claims.OrgID, req.Msg.UserId, models.OrgRole(req.Msg.Role))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&organizationv1.UpdateOrgMemberRoleResponse{
		Member: memberToProto(member),
	}), nil
}

func memberToProto(m *models.OrganizationMember) *organizationv1.OrganizationMember {
	proto := &organizationv1.OrganizationMember{
		Id:             m.ID,
		OrganizationId: m.OrganizationID,
		UserId:         m.UserID,
		Role:           string(m.Role),
	}
	if m.User != nil {
		proto.UserName = m.User.Name
		proto.UserEmail = m.User.Email
	}
	return proto
}

func orgToProto(o *models.Organization) *organizationv1.Organization {
	return &organizationv1.Organization{
		Id:        o.ID,
		Name:      o.Name,
		Slug:      o.Slug,
		CreatedAt: timestamppb.New(o.CreatedAt),
	}
}
