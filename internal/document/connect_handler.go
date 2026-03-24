package document

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	documentv1 "github.com/gobenpark/colign/gen/proto/document/v1"
	"github.com/gobenpark/colign/gen/proto/document/v1/documentv1connect"
	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/models"
)

type ConnectHandler struct {
	service           *Service
	jwtManager        *auth.JWTManager
	apiTokenValidator auth.APITokenValidator
}

var _ documentv1connect.DocumentServiceHandler = (*ConnectHandler)(nil)

func NewConnectHandler(service *Service, jwtManager *auth.JWTManager, apiTokenValidator auth.APITokenValidator) *ConnectHandler {
	return &ConnectHandler{service: service, jwtManager: jwtManager, apiTokenValidator: apiTokenValidator}
}

func (h *ConnectHandler) GetDocument(ctx context.Context, req *connect.Request[documentv1.GetDocumentRequest]) (*connect.Response[documentv1.GetDocumentResponse], error) {
	claims, err := auth.ResolveFromHeader(h.jwtManager, h.apiTokenValidator, ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	doc, err := h.service.Get(ctx, req.Msg.ChangeId, models.DocumentType(req.Msg.Type), claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &documentv1.GetDocumentResponse{}
	if doc != nil {
		resp.Document = docToProto(doc)
	}
	return connect.NewResponse(resp), nil
}

func (h *ConnectHandler) SaveDocument(ctx context.Context, req *connect.Request[documentv1.SaveDocumentRequest]) (*connect.Response[documentv1.SaveDocumentResponse], error) {
	claims, err := auth.ResolveFromHeader(h.jwtManager, h.apiTokenValidator, ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	doc, err := h.service.Save(ctx, SaveInput{
		ChangeID: req.Msg.ChangeId,
		Type:     models.DocumentType(req.Msg.Type),
		Title:    req.Msg.Title,
		Content:  req.Msg.Content,
		UserID:   claims.UserID,
	}, claims.OrgID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&documentv1.SaveDocumentResponse{
		Document: docToProto(doc),
	}), nil
}

func docToProto(d *models.Document) *documentv1.Document {
	return &documentv1.Document{
		Id:        d.ID,
		ChangeId:  d.ChangeID,
		Type:      string(d.Type),
		Title:     d.Title,
		Content:   d.Content,
		Version:   int32(d.Version),
		CreatedAt: timestamppb.New(d.CreatedAt),
		UpdatedAt: timestamppb.New(d.UpdatedAt),
	}
}
