package project

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrNotAuthorized   = errors.New("not authorized")
	slugRegexp         = regexp.MustCompile(`[^a-z0-9-]+`)
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func GenerateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = slugRegexp.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	// Collapse multiple dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

func (s *Service) ensureUniqueSlug(ctx context.Context, slug string, orgID int64) (string, error) {
	baseSlug := slug
	for i := 0; ; i++ {
		candidate := baseSlug
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", baseSlug, i+1)
		}
		exists, err := s.db.NewSelect().Model((*models.Project)(nil)).
			Where("slug = ?", candidate).
			Where("organization_id = ?", orgID).
			Exists(ctx)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

type CreateProjectInput struct {
	Name           string
	Description    string
	UserID         int64
	OrganizationID int64
}

func (s *Service) Create(ctx context.Context, input CreateProjectInput) (*models.Project, error) {
	slug := GenerateSlug(input.Name)
	uniqueSlug, err := s.ensureUniqueSlug(ctx, slug, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	project := &models.Project{
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Slug:           uniqueSlug,
		Description:    input.Description,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.NewInsert().Model(project).Exec(ctx); err != nil {
		return nil, err
	}

	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    input.UserID,
		Role:      models.RoleOwner,
	}
	if _, err := tx.NewInsert().Model(member).Exec(ctx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string, orgID int64) (*models.Project, []models.ProjectMember, []models.ProjectLabel, error) {
	project := new(models.Project)
	err := s.db.NewSelect().Model(project).
		Where("slug = ?", slug).
		Where("organization_id = ?", orgID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil, ErrProjectNotFound
		}
		return nil, nil, nil, err
	}

	// Load Lead user
	if project.LeadID != nil {
		lead := new(models.User)
		if err := s.db.NewSelect().Model(lead).Where("id = ?", *project.LeadID).Scan(ctx); err == nil {
			project.Lead = lead
		}
	}

	// Load Labels via junction table
	var labels []models.ProjectLabel
	var assignments []models.ProjectLabelAssignment
	err = s.db.NewSelect().Model(&assignments).
		Relation("Label").
		Where("project_id = ?", project.ID).
		Scan(ctx)
	if err == nil {
		for _, a := range assignments {
			if a.Label != nil {
				labels = append(labels, *a.Label)
			}
		}
	}

	var members []models.ProjectMember
	err = s.db.NewSelect().Model(&members).
		Relation("User").
		Where("pm.project_id = ?", project.ID).
		Scan(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	return project, members, labels, nil
}

func (s *Service) ListByUser(ctx context.Context, userID int64, orgID int64) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.NewSelect().Model(&projects).
		Join("JOIN project_members AS pm ON pm.project_id = p.id").
		Where("pm.user_id = ?", userID).
		Where("p.organization_id = ?", orgID).
		OrderExpr("p.updated_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	for i := range projects {
		if projects[i].LeadID == nil {
			continue
		}
		lead := new(models.User)
		if err := s.db.NewSelect().Model(lead).Where("id = ?", *projects[i].LeadID).Scan(ctx); err == nil {
			projects[i].Lead = lead
		}
	}

	return projects, nil
}

type UpdateProjectInput struct {
	ID          int64
	Name        string
	Description string
	Status      *string
	Priority    *string
	Health      *string
	LeadID      *int64
	ClearLead   bool
	StartDate   *time.Time
	ClearStart  bool
	TargetDate  *time.Time
	ClearTarget bool
	Icon        *string
	Color       *string
}

func (s *Service) Update(ctx context.Context, input UpdateProjectInput) (*models.Project, error) {
	project := new(models.Project)
	err := s.db.NewSelect().Model(project).Where("id = ?", input.ID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	project.Description = input.Description
	if input.Status != nil {
		project.Status = models.ProjectStatus(*input.Status)
	}
	if input.Priority != nil {
		project.Priority = models.ProjectPriority(*input.Priority)
	}
	if input.Health != nil {
		project.Health = models.ProjectHealth(*input.Health)
	}
	if input.ClearLead {
		project.LeadID = nil
	} else if input.LeadID != nil {
		project.LeadID = input.LeadID
	}
	if input.ClearStart {
		project.StartDate = nil
	} else if input.StartDate != nil {
		project.StartDate = input.StartDate
	}
	if input.ClearTarget {
		project.TargetDate = nil
	} else if input.TargetDate != nil {
		project.TargetDate = input.TargetDate
	}
	if input.Icon != nil {
		project.Icon = *input.Icon
	}
	if input.Color != nil {
		project.Color = *input.Color
	}
	project.UpdatedAt = time.Now()

	if _, err := s.db.NewUpdate().Model(project).WherePK().Exec(ctx); err != nil {
		return nil, err
	}

	// Reload with Lead relation
	if project.LeadID != nil {
		lead := new(models.User)
		if err := s.db.NewSelect().Model(lead).Where("id = ?", *project.LeadID).Scan(ctx); err == nil {
			project.Lead = lead
		}
	}

	return project, nil
}

// Label management

func (s *Service) CreateLabel(ctx context.Context, orgID int64, name, color string) (*models.ProjectLabel, error) {
	label := &models.ProjectLabel{
		OrganizationID: orgID,
		Name:           name,
		Color:          color,
	}
	if _, err := s.db.NewInsert().Model(label).Exec(ctx); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *Service) ListLabels(ctx context.Context, orgID int64) ([]models.ProjectLabel, error) {
	var labels []models.ProjectLabel
	err := s.db.NewSelect().Model(&labels).
		Where("organization_id = ?", orgID).
		OrderExpr("name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (s *Service) AssignLabel(ctx context.Context, projectID, labelID int64) error {
	assignment := &models.ProjectLabelAssignment{
		ProjectID: projectID,
		LabelID:   labelID,
	}
	_, err := s.db.NewInsert().Model(assignment).
		On("CONFLICT (project_id, label_id) DO NOTHING").
		Exec(ctx)
	return err
}

func (s *Service) RemoveLabel(ctx context.Context, projectID, labelID int64) error {
	_, err := s.db.NewDelete().Model((*models.ProjectLabelAssignment)(nil)).
		Where("project_id = ?", projectID).
		Where("label_id = ?", labelID).
		Exec(ctx)
	return err
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*models.Project)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (s *Service) CreateChange(ctx context.Context, projectID int64, name string) (*models.Change, error) {
	change := &models.Change{
		ProjectID: projectID,
		Name:      name,
		Stage:     models.StageDraft,
	}

	if _, err := s.db.NewInsert().Model(change).Exec(ctx); err != nil {
		return nil, err
	}
	return change, nil
}

func (s *Service) ListChanges(ctx context.Context, projectID int64, filter string) ([]models.Change, error) {
	var changes []models.Change
	q := s.db.NewSelect().Model(&changes).
		Where("project_id = ?", projectID).
		OrderExpr("created_at DESC")

	switch filter {
	case "archived":
		q = q.Where("archived_at IS NOT NULL")
	case "all":
		// no filter
	default: // "active" or empty
		q = q.Where("archived_at IS NULL")
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return changes, nil
}

func (s *Service) GetChange(ctx context.Context, id int64) (*models.Change, error) {
	change := new(models.Change)
	err := s.db.NewSelect().Model(change).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return change, nil
}

func (s *Service) DeleteChange(ctx context.Context, id int64) error {
	_, err := s.db.NewDelete().Model((*models.Change)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

type InviteMemberInput struct {
	ProjectID int64
	Email     string
	Role      models.Role
}

func (s *Service) InviteMember(ctx context.Context, input InviteMemberInput) (*models.ProjectMember, error) {
	// Find user by email
	user := new(models.User)
	err := s.db.NewSelect().Model(user).Where("email = ?", input.Email).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User not registered — create pending invitation
			tokenBytes := make([]byte, 32)
			if _, err := rand.Read(tokenBytes); err != nil {
				return nil, err
			}
			invitation := &models.PendingInvitation{
				ProjectID: input.ProjectID,
				Email:     input.Email,
				Role:      input.Role,
				Token:     hex.EncodeToString(tokenBytes),
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}
			if _, err := s.db.NewInsert().Model(invitation).
				On("CONFLICT (project_id, email) DO UPDATE").
				Set("role = EXCLUDED.role").
				Set("token = EXCLUDED.token").
				Set("expires_at = EXCLUDED.expires_at").
				Exec(ctx); err != nil {
				return nil, err
			}
			// TODO: send invitation email with token
			return nil, nil
		}
		return nil, err
	}

	member := &models.ProjectMember{
		ProjectID: input.ProjectID,
		UserID:    user.ID,
		Role:      input.Role,
	}

	if _, err := s.db.NewInsert().Model(member).
		On("CONFLICT (project_id, user_id) DO UPDATE").
		Set("role = EXCLUDED.role").
		Exec(ctx); err != nil {
		return nil, err
	}

	// Also add to organization if not already a member
	project := new(models.Project)
	if err := s.db.NewSelect().Model(project).Where("id = ?", input.ProjectID).Scan(ctx); err == nil && project.OrganizationID > 0 {
		orgMember := &models.OrganizationMember{
			OrganizationID: project.OrganizationID,
			UserID:         user.ID,
			Role:           "member",
		}
		_, _ = s.db.NewInsert().Model(orgMember).
			On("CONFLICT (organization_id, user_id) DO NOTHING").
			Exec(ctx)
	}

	return member, nil
}

// ProcessPendingInvitations converts pending invitations to project memberships
// after a new user registers. Called during user registration flow.
func (s *Service) ProcessPendingInvitations(ctx context.Context, userID int64, email string) error {
	var invitations []models.PendingInvitation
	err := s.db.NewSelect().Model(&invitations).
		Where("email = ?", email).
		Where("expires_at > ?", time.Now()).
		Scan(ctx)
	if err != nil {
		return err
	}

	for _, inv := range invitations {
		member := &models.ProjectMember{
			ProjectID: inv.ProjectID,
			UserID:    userID,
			Role:      inv.Role,
		}
		if _, err := s.db.NewInsert().Model(member).
			On("CONFLICT (project_id, user_id) DO NOTHING").
			Exec(ctx); err != nil {
			return err
		}
	}

	// Clean up processed invitations
	if len(invitations) > 0 {
		_, err = s.db.NewDelete().Model((*models.PendingInvitation)(nil)).
			Where("email = ?", email).
			Exec(ctx)
	}

	return err
}

type SearchResult struct {
	Type      string
	ID        int64
	Title     string
	Subtitle  string
	Slug      string
	ProjectID int64
}

func (s *Service) Search(ctx context.Context, query string, userID, orgID int64) ([]SearchResult, error) {
	if query == "" {
		return nil, nil
	}
	like := "%" + query + "%"
	var results []SearchResult

	// Search projects
	var projects []models.Project
	if err := s.db.NewSelect().Model(&projects).
		Join("JOIN project_members AS pm ON pm.project_id = p.id").
		Where("pm.user_id = ?", userID).
		Where("p.organization_id = ?", orgID).
		Where("p.name ILIKE ?", like).
		Limit(5).
		Scan(ctx); err == nil {
		for _, p := range projects {
			results = append(results, SearchResult{
				Type:      "project",
				ID:        p.ID,
				Title:     p.Name,
				Subtitle:  string(p.Status),
				Slug:      p.Slug,
				ProjectID: p.ID,
			})
		}
	}

	// Search changes
	var changes []models.Change
	if err := s.db.NewSelect().Model(&changes).
		Relation("Project").
		Join("JOIN project_members AS pm ON pm.project_id = c.project_id").
		Where("pm.user_id = ?", userID).
		Where("c.name ILIKE ?", like).
		Limit(10).
		Scan(ctx); err == nil {
		for _, c := range changes {
			subtitle := string(c.Stage)
			slug := ""
			if c.Project != nil {
				slug = c.Project.Slug
			}
			results = append(results, SearchResult{
				Type:      "change",
				ID:        c.ID,
				Title:     c.Name,
				Subtitle:  subtitle,
				Slug:      slug,
				ProjectID: c.ProjectID,
			})
		}
	}

	// Search tasks
	var tasks []models.Task
	if err := s.db.NewSelect().Model(&tasks).
		Relation("Change").
		Relation("Change.Project").
		Join("JOIN changes AS ch ON ch.id = \"task\".change_id").
		Join("JOIN project_members AS pm ON pm.project_id = ch.project_id").
		Where("pm.user_id = ?", userID).
		Where("\"task\".title ILIKE ?", like).
		Limit(5).
		Scan(ctx); err == nil {
		for _, t := range tasks {
			slug := ""
			var projectID int64
			if t.Change != nil {
				projectID = t.Change.ProjectID
				if t.Change.Project != nil {
					slug = t.Change.Project.Slug
				}
			}
			results = append(results, SearchResult{
				Type:      "task",
				ID:        t.ID,
				Title:     t.Title,
				Subtitle:  string(t.Status),
				Slug:      slug,
				ProjectID: projectID,
			})
		}
	}

	return results, nil
}
