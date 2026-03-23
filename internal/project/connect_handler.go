package project

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	projectv1 "github.com/gobenpark/colign/gen/proto/project/v1"
	"github.com/gobenpark/colign/gen/proto/project/v1/projectv1connect"
	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/models"
)

type ConnectHandler struct {
	service           *Service
	jwtManager        *auth.JWTManager
	apiTokenValidator auth.APITokenValidator
}

var _ projectv1connect.ProjectServiceHandler = (*ConnectHandler)(nil)

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

func (h *ConnectHandler) CreateProject(ctx context.Context, req *connect.Request[projectv1.CreateProjectRequest]) (*connect.Response[projectv1.CreateProjectResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	project, err := h.service.Create(ctx, CreateProjectInput{
		Name:           req.Msg.Name,
		Description:    req.Msg.Description,
		UserID:         claims.UserID,
		OrganizationID: claims.OrgID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.CreateProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ConnectHandler) GetProject(ctx context.Context, req *connect.Request[projectv1.GetProjectRequest]) (*connect.Response[projectv1.GetProjectResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	project, members, labels, err := h.service.GetBySlug(ctx, req.Msg.Slug, claims.OrgID)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoMembers := make([]*projectv1.ProjectMember, len(members))
	for i, m := range members {
		protoMembers[i] = memberToProto(&m)
	}

	pp := projectToProto(project)
	for i := range labels {
		pp.Labels = append(pp.Labels, labelToProto(&labels[i]))
	}

	return connect.NewResponse(&projectv1.GetProjectResponse{
		Project: pp,
		Members: protoMembers,
	}), nil
}

func (h *ConnectHandler) ListProjects(ctx context.Context, req *connect.Request[projectv1.ListProjectsRequest]) (*connect.Response[projectv1.ListProjectsResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	projects, err := h.service.ListByUser(ctx, claims.UserID, claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoProjects := make([]*projectv1.Project, len(projects))
	for i, p := range projects {
		protoProjects[i] = projectToProto(&p)
	}

	return connect.NewResponse(&projectv1.ListProjectsResponse{
		Projects: protoProjects,
	}), nil
}

func (h *ConnectHandler) UpdateProject(ctx context.Context, req *connect.Request[projectv1.UpdateProjectRequest]) (*connect.Response[projectv1.UpdateProjectResponse], error) {
	input := UpdateProjectInput{
		ID:          req.Msg.Id,
		Name:        req.Msg.Name,
		Description: req.Msg.Description,
	}

	if req.Msg.Status != nil {
		input.Status = req.Msg.Status
	}
	if req.Msg.Priority != nil {
		input.Priority = req.Msg.Priority
	}
	if req.Msg.Health != nil {
		input.Health = req.Msg.Health
	}
	if req.Msg.LeadId != nil {
		v := *req.Msg.LeadId
		if v == 0 {
			input.ClearLead = true
		} else {
			input.LeadID = &v
		}
	}
	if req.Msg.StartDate != nil {
		v := *req.Msg.StartDate
		if v == "" {
			input.ClearStart = true
		} else {
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}
			input.StartDate = &t
		}
	}
	if req.Msg.TargetDate != nil {
		v := *req.Msg.TargetDate
		if v == "" {
			input.ClearTarget = true
		} else {
			t, err := time.Parse("2006-01-02", v)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, err)
			}
			input.TargetDate = &t
		}
	}
	if req.Msg.Icon != nil {
		input.Icon = req.Msg.Icon
	}
	if req.Msg.Color != nil {
		input.Color = req.Msg.Color
	}

	project, err := h.service.Update(ctx, input)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.UpdateProjectResponse{
		Project: projectToProto(project),
	}), nil
}

func (h *ConnectHandler) DeleteProject(ctx context.Context, req *connect.Request[projectv1.DeleteProjectRequest]) (*connect.Response[projectv1.DeleteProjectResponse], error) {
	if err := h.service.Delete(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&projectv1.DeleteProjectResponse{}), nil
}

func (h *ConnectHandler) InviteMember(ctx context.Context, req *connect.Request[projectv1.InviteMemberRequest]) (*connect.Response[projectv1.InviteMemberResponse], error) {
	member, err := h.service.InviteMember(ctx, InviteMemberInput{
		ProjectID: req.Msg.ProjectId,
		Email:     req.Msg.Email,
		Role:      models.Role(req.Msg.Role),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.InviteMemberResponse{
		Member: memberToProto(member),
	}), nil
}

func (h *ConnectHandler) CreateLabel(ctx context.Context, req *connect.Request[projectv1.CreateLabelRequest]) (*connect.Response[projectv1.CreateLabelResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	label, err := h.service.CreateLabel(ctx, claims.OrgID, req.Msg.Name, req.Msg.Color)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.CreateLabelResponse{
		Label: labelToProto(label),
	}), nil
}

func (h *ConnectHandler) ListLabels(ctx context.Context, req *connect.Request[projectv1.ListLabelsRequest]) (*connect.Response[projectv1.ListLabelsResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	labels, err := h.service.ListLabels(ctx, claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoLabels := make([]*projectv1.ProjectLabel, len(labels))
	for i, l := range labels {
		protoLabels[i] = labelToProto(&l)
	}

	return connect.NewResponse(&projectv1.ListLabelsResponse{
		Labels: protoLabels,
	}), nil
}

func (h *ConnectHandler) AssignLabel(ctx context.Context, req *connect.Request[projectv1.AssignLabelRequest]) (*connect.Response[projectv1.AssignLabelResponse], error) {
	if err := h.service.AssignLabel(ctx, req.Msg.ProjectId, req.Msg.LabelId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&projectv1.AssignLabelResponse{}), nil
}

func (h *ConnectHandler) RemoveLabel(ctx context.Context, req *connect.Request[projectv1.RemoveLabelRequest]) (*connect.Response[projectv1.RemoveLabelResponse], error) {
	if err := h.service.RemoveLabel(ctx, req.Msg.ProjectId, req.Msg.LabelId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&projectv1.RemoveLabelResponse{}), nil
}

func (h *ConnectHandler) CreateChange(ctx context.Context, req *connect.Request[projectv1.CreateChangeRequest]) (*connect.Response[projectv1.CreateChangeResponse], error) {
	change, err := h.service.CreateChange(ctx, req.Msg.ProjectId, req.Msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.CreateChangeResponse{
		Change: changeToProto(change),
	}), nil
}

func (h *ConnectHandler) ListChanges(ctx context.Context, req *connect.Request[projectv1.ListChangesRequest]) (*connect.Response[projectv1.ListChangesResponse], error) {
	changes, err := h.service.ListChanges(ctx, req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoChanges := make([]*projectv1.Change, len(changes))
	for i, c := range changes {
		protoChanges[i] = changeToProto(&c)
	}

	return connect.NewResponse(&projectv1.ListChangesResponse{
		Changes: protoChanges,
	}), nil
}

func (h *ConnectHandler) GetChange(ctx context.Context, req *connect.Request[projectv1.GetChangeRequest]) (*connect.Response[projectv1.GetChangeResponse], error) {
	change, err := h.service.GetChange(ctx, req.Msg.Id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.GetChangeResponse{
		Change: changeToProto(change),
	}), nil
}

func (h *ConnectHandler) DeleteChange(ctx context.Context, req *connect.Request[projectv1.DeleteChangeRequest]) (*connect.Response[projectv1.DeleteChangeResponse], error) {
	if err := h.service.DeleteChange(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&projectv1.DeleteChangeResponse{}), nil
}

func projectToProto(p *models.Project) *projectv1.Project {
	pp := &projectv1.Project{
		Id:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Status:      string(p.Status),
		Priority:    string(p.Priority),
		Health:      string(p.Health),
		Icon:        p.Icon,
		Color:       p.Color,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	if p.LeadID != nil {
		pp.LeadId = p.LeadID
	}
	if p.Lead != nil {
		pp.LeadName = p.Lead.Name
		pp.LeadAvatarUrl = p.Lead.AvatarURL
	}
	if p.StartDate != nil {
		s := p.StartDate.Format("2006-01-02")
		pp.StartDate = &s
	}
	if p.TargetDate != nil {
		s := p.TargetDate.Format("2006-01-02")
		pp.TargetDate = &s
	}

	return pp
}

func memberToProto(m *models.ProjectMember) *projectv1.ProjectMember {
	pm := &projectv1.ProjectMember{
		Id:        m.ID,
		ProjectId: m.ProjectID,
		UserId:    m.UserID,
		Role:      string(m.Role),
	}
	if m.User != nil {
		pm.UserName = m.User.Name
		pm.UserEmail = m.User.Email
	}
	return pm
}

func changeToProto(c *models.Change) *projectv1.Change {
	return &projectv1.Change{
		Id:        c.ID,
		ProjectId: c.ProjectID,
		Name:      c.Name,
		Stage:     string(c.Stage),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func labelToProto(l *models.ProjectLabel) *projectv1.ProjectLabel {
	return &projectv1.ProjectLabel{
		Id:    l.ID,
		Name:  l.Name,
		Color: l.Color,
	}
}

func (h *ConnectHandler) Search(ctx context.Context, req *connect.Request[projectv1.SearchRequest]) (*connect.Response[projectv1.SearchResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}

	results, err := h.service.Search(ctx, req.Msg.Query, claims.UserID, claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoResults := make([]*projectv1.SearchResult, len(results))
	for i, r := range results {
		protoResults[i] = &projectv1.SearchResult{
			Type:      r.Type,
			Id:        r.ID,
			Title:     r.Title,
			Subtitle:  r.Subtitle,
			Slug:      r.Slug,
			ProjectId: r.ProjectID,
		}
	}

	return connect.NewResponse(&projectv1.SearchResponse{
		Results: protoResults,
	}), nil
}
