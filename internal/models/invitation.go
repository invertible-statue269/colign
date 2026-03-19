package models

import (
	"time"

	"github.com/uptrace/bun"
)

type PendingInvitation struct {
	bun.BaseModel `bun:"table:pending_invitations,alias:pi"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ProjectID int64     `bun:"project_id,notnull"`
	Email     string    `bun:"email,notnull"`
	Role      Role      `bun:"role,notnull"`
	Token     string    `bun:"token,notnull,unique"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Project *Project `bun:"rel:belongs-to,join:project_id=id"`
}
