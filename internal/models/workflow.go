package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ApprovalPolicy struct {
	bun.BaseModel `bun:"table:approval_policies,alias:ap"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ProjectID int64     `bun:"project_id,notnull,unique"`
	Policy    string    `bun:"policy,notnull,default:'owner_one'"` // owner_one, editor_two, all, auto_pass
	MinCount  int       `bun:"min_count,notnull,default:1"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Project *Project `bun:"rel:belongs-to,join:project_id=id"`
}

type Approval struct {
	bun.BaseModel `bun:"table:approvals,alias:apv"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ChangeID  int64     `bun:"change_id,notnull"`
	UserID    int64     `bun:"user_id,notnull"`
	Status    string    `bun:"status,notnull"` // approved, changes_requested
	Comment   string    `bun:"comment"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Change *Change `bun:"rel:belongs-to,join:change_id=id"`
	User   *User   `bun:"rel:belongs-to,join:user_id=id"`
}

type WorkflowEvent struct {
	bun.BaseModel `bun:"table:workflow_events,alias:we"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ChangeID  int64     `bun:"change_id,notnull"`
	FromStage string    `bun:"from_stage,notnull"`
	ToStage   string    `bun:"to_stage,notnull"`
	Action    string    `bun:"action,notnull"` // auto_advance, revert, approve, request_changes
	Reason    string    `bun:"reason"`
	UserID    int64     `bun:"user_id"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Change *Change `bun:"rel:belongs-to,join:change_id=id"`
	User   *User   `bun:"rel:belongs-to,join:user_id=id"`
}
