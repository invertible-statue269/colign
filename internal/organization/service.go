package organization

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

var ErrOrgNotFound = errors.New("organization not found")

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListByUser(ctx context.Context, userID int64) ([]models.Organization, error) {
	var orgs []models.Organization
	err := s.db.NewSelect().Model(&orgs).
		Join("JOIN organization_members AS om ON om.organization_id = o.id").
		Where("om.user_id = ?", userID).
		OrderExpr("o.created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*models.Organization, error) {
	org := new(models.Organization)
	err := s.db.NewSelect().Model(org).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

func (s *Service) IsMember(ctx context.Context, orgID, userID int64) (bool, error) {
	return s.db.NewSelect().Model((*models.OrganizationMember)(nil)).
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Exists(ctx)
}

func (s *Service) ListMembers(ctx context.Context, orgID int64) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := s.db.NewSelect().Model(&members).
		Relation("User").
		Where("om.organization_id = ?", orgID).
		OrderExpr("om.created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (s *Service) InviteMember(ctx context.Context, orgID int64, email string, role models.OrgRole) (*models.OrganizationMember, error) {
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found: they must sign up first")
		}
		return nil, err
	}

	member := &models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         user.ID,
		Role:           role,
	}
	if _, err := s.db.NewInsert().Model(member).
		On("CONFLICT (organization_id, user_id) DO UPDATE").
		Set("role = EXCLUDED.role").
		Exec(ctx); err != nil {
		return nil, err
	}

	member.User = user
	return member, nil
}

func (s *Service) RemoveMember(ctx context.Context, orgID, userID int64) error {
	_, err := s.db.NewDelete().Model((*models.OrganizationMember)(nil)).
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

func (s *Service) UpdateMemberRole(ctx context.Context, orgID, userID int64, role models.OrgRole) (*models.OrganizationMember, error) {
	member := new(models.OrganizationMember)
	_, err := s.db.NewUpdate().Model(member).
		Set("role = ?", role).
		Where("organization_id = ?", orgID).
		Where("user_id = ?", userID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (s *Service) Update(ctx context.Context, id int64, name string) (*models.Organization, error) {
	org := new(models.Organization)
	err := s.db.NewSelect().Model(org).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	org.Name = name
	org.UpdatedAt = time.Now()

	if _, err := s.db.NewUpdate().Model(org).WherePK().Exec(ctx); err != nil {
		return nil, err
	}
	return org, nil
}
