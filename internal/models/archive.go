package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ArchiveMode string

const (
	ArchiveModeManual ArchiveMode = "manual"
	ArchiveModeAuto   ArchiveMode = "auto"
)

type ArchiveTrigger string

const (
	TriggerTasksDone         ArchiveTrigger = "tasks_done"
	TriggerDaysAfterApproved ArchiveTrigger = "days_after_approved"
	TriggerTasksDoneAndDays  ArchiveTrigger = "tasks_done_and_days"
)

type ArchivePolicy struct {
	bun.BaseModel `bun:"table:archive_policies,alias:ap"`

	ID          int64          `bun:"id,pk,autoincrement"`
	ProjectID   int64          `bun:"project_id,notnull,unique"`
	Mode        ArchiveMode    `bun:"mode,notnull,default:'manual'"`
	TriggerType ArchiveTrigger `bun:"trigger_type,notnull,default:'tasks_done'"`
	DaysDelay   int            `bun:"days_delay,notnull,default:0"`
	CreatedAt   time.Time      `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time      `bun:"updated_at,notnull,default:current_timestamp"`
}
