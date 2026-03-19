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
