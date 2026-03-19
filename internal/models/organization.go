package models

import (
	"time"

	"github.com/uptrace/bun"
)

type OrgRole string

const (
	OrgRoleOwner  OrgRole = "owner"
	OrgRoleAdmin  OrgRole = "admin"
	OrgRoleMember OrgRole = "member"
)

type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:o"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Slug      string    `bun:"slug,notnull,unique"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	Members []OrganizationMember `bun:"rel:has-many,join:id=organization_id"`
}

type OrganizationMember struct {
	bun.BaseModel `bun:"table:organization_members,alias:om"`

	ID             int64     `bun:"id,pk,autoincrement"`
	OrganizationID int64     `bun:"organization_id,notnull"`
	UserID         int64     `bun:"user_id,notnull"`
	Role           OrgRole   `bun:"role,notnull,default:'member'"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Organization *Organization `bun:"rel:belongs-to,join:organization_id=id"`
	User         *User         `bun:"rel:belongs-to,join:user_id=id"`
}
