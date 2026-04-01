package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ACStep struct {
	Keyword string `json:"keyword"` // Given, When, Then, And, But
	Text    string `json:"text"`
}

type AcceptanceCriteria struct {
	bun.BaseModel `bun:"table:acceptance_criteria,alias:ac"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ChangeID  int64     `bun:"change_id,notnull"`
	Scenario  string    `bun:"scenario,notnull"`
	Steps     []ACStep  `bun:"steps,type:jsonb,notnull,default:'[]'"`
	Met       bool      `bun:"met,notnull,default:false"`
	SortOrder int       `bun:"sort_order,notnull,default:0"`
	TestRef   string    `bun:"test_ref,notnull,default:''"`
	CreatedBy *int64    `bun:"created_by"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Change *Change `bun:"rel:belongs-to,join:change_id=id"`
	User   *User   `bun:"rel:belongs-to,join:created_by=id"`
}
