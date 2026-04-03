package notification

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func (s *Service) List(ctx context.Context, userID int64, filter string) ([]models.Notification, int, error) {
	var notifications []models.Notification
	q := s.db.NewSelect().Model(&notifications).
		Where("n.user_id = ?", userID).
		Relation("Actor").
		Relation("Change").
		Relation("Project").
		OrderExpr("n.created_at DESC").
		Limit(50)

	switch filter {
	case "unread":
		q = q.Where("n.read = FALSE")
	case "review_request", "comment", "mention", "stage_change", "invite":
		q = q.Where("n.type = ?", filter)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, 0, err
	}

	// Unread count
	count, err := s.db.NewSelect().Model((*models.Notification)(nil)).
		Where("user_id = ?", userID).
		Where("read = FALSE").
		Count(ctx)
	if err != nil {
		return notifications, 0, err
	}

	return notifications, count, nil
}

func (s *Service) MarkRead(ctx context.Context, userID int64, notificationID int64, read bool) error {
	_, err := s.db.NewUpdate().Model((*models.Notification)(nil)).
		Set("read = ?", read).
		Where("id = ?", notificationID).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

func (s *Service) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := s.db.NewUpdate().Model((*models.Notification)(nil)).
		Set("read = TRUE").
		Where("user_id = ?", userID).
		Where("read = FALSE").
		Exec(ctx)
	return err
}

func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return s.db.NewSelect().Model((*models.Notification)(nil)).
		Where("user_id = ?", userID).
		Where("read = FALSE").
		Count(ctx)
}

type CreateInput struct {
	UserID           int64
	Type             models.NotificationType
	ActorID          int64
	ChangeID         int64
	ProjectID        int64
	Stage            string
	CommentPreview   string
	MentionedUserIDs []int64
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*models.Notification, error) {
	n := &models.Notification{
		UserID:           input.UserID,
		Type:             input.Type,
		ActorID:          input.ActorID,
		ChangeID:         input.ChangeID,
		ProjectID:        input.ProjectID,
		Stage:            input.Stage,
		CommentPreview:   input.CommentPreview,
		MentionedUserIDs: input.MentionedUserIDs,
		CreatedAt:        time.Now(),
	}
	if _, err := s.db.NewInsert().Model(n).Exec(ctx); err != nil {
		return nil, err
	}
	return n, nil
}

// MentionedUserInfo holds the user data needed for rendering mentions in previews.
type MentionedUserInfo struct {
	ID    int64  `bun:"id"`
	Name  string `bun:"name"`
	Email string `bun:"email"`
}

// LookupUsers returns user info for the given IDs.
func (s *Service) LookupUsers(ctx context.Context, ids []int64) (map[int64]MentionedUserInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []MentionedUserInfo
	if err := s.db.NewSelect().
		TableExpr("users").
		ColumnExpr("id, name, email").
		Where("id IN (?)", bun.List(ids)).
		Scan(ctx, &users); err != nil {
		return nil, err
	}
	result := make(map[int64]MentionedUserInfo, len(users))
	for _, u := range users {
		result[u.ID] = u
	}
	return result, nil
}
