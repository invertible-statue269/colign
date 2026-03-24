package acceptance

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

var ErrNotFound = errors.New("acceptance criteria not found")

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, ac *models.AcceptanceCriteria) error {
	_, err := s.db.NewInsert().Model(ac).Exec(ctx)
	return err
}

func (s *Service) List(ctx context.Context, changeID int64) ([]models.AcceptanceCriteria, error) {
	var items []models.AcceptanceCriteria
	err := s.db.NewSelect().Model(&items).
		Where("change_id = ?", changeID).
		OrderExpr("sort_order ASC, id ASC").
		Scan(ctx)
	return items, err
}

type UpdateInput struct {
	Scenario  string
	Steps     []models.ACStep
	SortOrder int
	TestRef   *string
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput, orgID int64) (*models.AcceptanceCriteria, error) {
	ac := new(models.AcceptanceCriteria)
	err := s.db.NewSelect().Model(ac).
		Join("JOIN changes AS c ON c.id = ac.change_id").
		Join("JOIN projects AS p ON p.id = c.project_id").
		Where("ac.id = ?", id).
		Where("p.organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	ac.Scenario = input.Scenario
	ac.Steps = input.Steps
	ac.SortOrder = input.SortOrder
	if input.TestRef != nil {
		ac.TestRef = *input.TestRef
	}
	ac.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().Model(ac).WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

func (s *Service) Toggle(ctx context.Context, id int64, met bool, orgID int64) (*models.AcceptanceCriteria, error) {
	ac := new(models.AcceptanceCriteria)
	err := s.db.NewSelect().Model(ac).
		Join("JOIN changes AS c ON c.id = ac.change_id").
		Join("JOIN projects AS p ON p.id = c.project_id").
		Where("ac.id = ?", id).
		Where("p.organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	ac.Met = met
	ac.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().Model(ac).Column("met", "updated_at").WherePK().Exec(ctx)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

func (s *Service) Delete(ctx context.Context, id int64, orgID int64) error {
	res, err := s.db.NewDelete().Model((*models.AcceptanceCriteria)(nil)).
		Where("id = ?", id).
		Where("change_id IN (SELECT c.id FROM changes c JOIN projects p ON p.id = c.project_id WHERE p.organization_id = ?)", orgID).
		Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
