package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ChatSession struct {
	bun.BaseModel `bun:"table:chat_sessions,alias:cs"`

	ID        int64     `bun:"id,pk,autoincrement"`
	ChangeID  int64     `bun:"change_id,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Change   *Change       `bun:"rel:belongs-to,join:change_id=id"`
	Messages []ChatMessage `bun:"rel:has-many,join:id=session_id"`
}

type ChatMessage struct {
	bun.BaseModel `bun:"table:chat_messages,alias:cm"`

	ID        int64     `bun:"id,pk,autoincrement"`
	SessionID int64     `bun:"session_id,notnull"`
	Role      string    `bun:"role,notnull"` // user, assistant
	Content   string    `bun:"content,notnull,type:text"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Session *ChatSession `bun:"rel:belongs-to,join:session_id=id"`
}
