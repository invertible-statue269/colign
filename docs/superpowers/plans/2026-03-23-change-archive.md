# Change Archive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add archiving for completed Changes so active project views stay clean.

**Architecture:** Add `archived_at` nullable timestamp to the Change model, with per-project ArchivePolicy for manual/auto mode. Archive/unarchive RPCs live in ProjectService alongside existing Change CRUD. Auto-archive evaluates on task completion events and via a daily in-process ticker.

**Tech Stack:** Go (bun ORM, connect-rpc), PostgreSQL, protobuf/buf, Next.js (React, connect-es), Tailwind CSS, shadcn/ui

**Spec:** `docs/superpowers/specs/2026-03-23-change-archive-design.md`

---

## File Structure

### Backend — New Files
| File | Responsibility |
|---|---|
| `migrations/000018_add_archive.up.sql` | Add `archived_at` column + `archive_policies` table |
| `migrations/000018_add_archive.down.sql` | Rollback migration |
| `internal/models/archive.go` | `ArchivePolicy` model |
| `internal/archive/service.go` | Archive/unarchive logic, auto-archive evaluation |
| `internal/archive/service_test.go` | Unit tests for archive service |
| `internal/archive/cron.go` | Daily auto-archive ticker |
| `internal/archive/cron_test.go` | Tests for cron logic |

### Backend — Modified Files
| File | Change |
|---|---|
| `internal/models/change.go` | Add `ArchivedAt *time.Time` field |
| `proto/project/v1/project.proto` | Add `archived_at` to Change, archive RPCs, filter to ListChanges |
| `internal/project/service.go` | Update `ListChanges` to filter by archive status |
| `internal/project/connect_handler.go` | Add archive/unarchive/policy handlers, update `changeToProto` |
| `internal/workflow/service.go` | Block advance/revert on archived changes |
| `internal/task/service.go` | Trigger auto-archive evaluation on task status → done |
| `internal/server/server.go` | Wire archive service, start cron |
| `web/src/app/projects/[slug]/settings/page.tsx` | Add "Archive" settings tab |

### Frontend — Modified Files
| File | Change |
|---|---|
| `web/src/app/projects/[slug]/page.tsx` | Active/Archived tabs in ChangesTab |
| `web/src/app/projects/[slug]/changes/[changeId]/page.tsx` | Archive/Unarchive buttons |
| `web/src/lib/project.ts` | Export projectClient (already exists, used for new RPCs) |

---

## Task 1: Database Migration

**Files:**
- Create: `migrations/000018_add_archive.up.sql`
- Create: `migrations/000018_add_archive.down.sql`

- [ ] **Step 1: Write the up migration**

```sql
-- migrations/000018_add_archive.up.sql
ALTER TABLE changes ADD COLUMN archived_at TIMESTAMPTZ;
CREATE INDEX idx_changes_active ON changes (project_id) WHERE archived_at IS NULL;

CREATE TABLE archive_policies (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    mode TEXT NOT NULL DEFAULT 'manual' CHECK (mode IN ('manual', 'auto')),
    trigger_type TEXT NOT NULL DEFAULT 'tasks_done' CHECK (trigger_type IN ('tasks_done', 'days_after_ready', 'tasks_done_and_days')),
    days_delay INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- [ ] **Step 2: Write the down migration**

```sql
-- migrations/000018_add_archive.down.sql
DROP INDEX IF EXISTS idx_changes_active;
ALTER TABLE changes DROP COLUMN IF EXISTS archived_at;
DROP TABLE IF EXISTS archive_policies;
```

- [ ] **Step 3: Commit**

```bash
git add migrations/000018_add_archive.up.sql migrations/000018_add_archive.down.sql
git commit -m "feat: add archive migration for changes and archive_policies"
```

---

## Task 2: Models

**Files:**
- Modify: `internal/models/change.go` (add ArchivedAt field)
- Create: `internal/models/archive.go` (ArchivePolicy model)

- [ ] **Step 1: Add ArchivedAt to Change model**

In `internal/models/change.go`, add to the `Change` struct:

```go
ArchivedAt *time.Time `bun:"archived_at"`
```

- [ ] **Step 2: Create ArchivePolicy model**

Create `internal/models/archive.go`:

```go
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
	TriggerTasksDone        ArchiveTrigger = "tasks_done"
	TriggerDaysAfterReady   ArchiveTrigger = "days_after_ready"
	TriggerTasksDoneAndDays ArchiveTrigger = "tasks_done_and_days"
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
```

- [ ] **Step 3: Commit**

```bash
git add internal/models/change.go internal/models/archive.go
git commit -m "feat: add ArchivePolicy model and ArchivedAt field to Change"
```

---

## Task 3: Proto Definitions

**Files:**
- Modify: `proto/project/v1/project.proto`

- [ ] **Step 1: Update Change message and add archive messages**

In `proto/project/v1/project.proto`:

1. Add to `Change` message:
```protobuf
optional google.protobuf.Timestamp archived_at = 7;
```

2. Add `filter` to `ListChangesRequest`:
```protobuf
message ListChangesRequest {
  int64 project_id = 1;
  optional string filter = 2; // "active" (default), "archived", "all"
}
```

3. Add archive messages:
```protobuf
// Archive
message ArchiveChangeRequest {
  int64 change_id = 1;
}
message ArchiveChangeResponse {
  Change change = 1;
}

message UnarchiveChangeRequest {
  int64 change_id = 1;
}
message UnarchiveChangeResponse {
  Change change = 1;
}

// Archive Policy
message ArchivePolicy {
  int64 project_id = 1;
  string mode = 2;
  string trigger = 3;
  int32 days_delay = 4;
}

message GetArchivePolicyRequest {
  int64 project_id = 1;
}
message GetArchivePolicyResponse {
  ArchivePolicy policy = 1;
}

message UpdateArchivePolicyRequest {
  int64 project_id = 1;
  string mode = 2;
  string trigger = 3;
  int32 days_delay = 4;
}
message UpdateArchivePolicyResponse {
  ArchivePolicy policy = 1;
}
```

4. Add RPCs to `ProjectService`:
```protobuf
rpc ArchiveChange(ArchiveChangeRequest) returns (ArchiveChangeResponse) {}
rpc UnarchiveChange(UnarchiveChangeRequest) returns (UnarchiveChangeResponse) {}
rpc GetArchivePolicy(GetArchivePolicyRequest) returns (GetArchivePolicyResponse) {}
rpc UpdateArchivePolicy(UpdateArchivePolicyRequest) returns (UpdateArchivePolicyResponse) {}
```

- [ ] **Step 2: Generate protobuf code**

```bash
cd proto && buf generate
```

- [ ] **Step 3: Commit**

```bash
git add proto/ gen/
git commit -m "feat: add archive proto definitions and generate code"
```

---

## Task 4: Archive Service (Backend Core)

**Files:**
- Create: `internal/archive/service.go`
- Create: `internal/archive/service_test.go`

- [ ] **Step 1: Write tests for Archive and Unarchive**

Create `internal/archive/service_test.go`:

```go
package archive

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/gobenpark/colign/internal/models"
)

func setupTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := bun.NewDB(mockDB, pgdialect.New())
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestArchive_NotReady(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)

	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "archived_at", "created_at", "updated_at"}).
		AddRow(1, 1, "test", "design", nil, time.Now(), time.Now())
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	_, err := svc.Archive(context.Background(), 1, 1)
	require.ErrorIs(t, err, ErrNotReady)
}

func TestArchive_AlreadyArchived(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "archived_at", "created_at", "updated_at"}).
		AddRow(1, 1, "test", "ready", now, time.Now(), time.Now())
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	_, err := svc.Archive(context.Background(), 1, 1)
	require.ErrorIs(t, err, ErrAlreadyArchived)
}

func TestUnarchive_NotArchived(t *testing.T) {
	db, mock := setupTestDB(t)
	svc := NewService(db)

	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "stage", "archived_at", "created_at", "updated_at"}).
		AddRow(1, 1, "test", "ready", nil, time.Now(), time.Now())
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	_, err := svc.Unarchive(context.Background(), 1, 1)
	require.ErrorIs(t, err, ErrNotArchived)
}

func TestShouldAutoArchive_TasksDone(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	result := svc.shouldAutoArchive(
		models.TriggerTasksDone,
		0,        // days_delay
		true,     // allTasksDone
		&now,     // readyAt
	)
	assert.True(t, result)
}

func TestShouldAutoArchive_TasksDone_NotAllDone(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	result := svc.shouldAutoArchive(
		models.TriggerTasksDone,
		0,
		false,    // not all done
		&now,
	)
	assert.False(t, result)
}

func TestShouldAutoArchive_DaysAfterReady(t *testing.T) {
	svc := &Service{}
	readyAt := time.Now().Add(-8 * 24 * time.Hour) // 8 days ago

	result := svc.shouldAutoArchive(
		models.TriggerDaysAfterReady,
		7,        // 7 days delay
		false,    // tasks don't matter
		&readyAt,
	)
	assert.True(t, result)
}

func TestShouldAutoArchive_DaysAfterReady_TooSoon(t *testing.T) {
	svc := &Service{}
	readyAt := time.Now().Add(-3 * 24 * time.Hour) // 3 days ago

	result := svc.shouldAutoArchive(
		models.TriggerDaysAfterReady,
		7,
		false,
		&readyAt,
	)
	assert.False(t, result)
}

func TestShouldAutoArchive_TasksDoneAndDays(t *testing.T) {
	svc := &Service{}
	readyAt := time.Now().Add(-8 * 24 * time.Hour)

	result := svc.shouldAutoArchive(
		models.TriggerTasksDoneAndDays,
		7,
		true,     // all done
		&readyAt,
	)
	assert.True(t, result)
}

func TestShouldAutoArchive_TasksDoneAndDays_OnlyTasksDone(t *testing.T) {
	svc := &Service{}
	readyAt := time.Now().Add(-3 * 24 * time.Hour) // only 3 days

	result := svc.shouldAutoArchive(
		models.TriggerTasksDoneAndDays,
		7,
		true,     // tasks done but days not elapsed
		&readyAt,
	)
	assert.False(t, result)
}

func TestShouldAutoArchive_NoReadyAt(t *testing.T) {
	svc := &Service{}

	result := svc.shouldAutoArchive(
		models.TriggerDaysAfterReady,
		7,
		false,
		nil, // no ready timestamp
	)
	assert.False(t, result)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/ben/Projects/colign && go test ./internal/archive/... -v
```

Expected: compilation errors (Service, shouldAutoArchive not defined)

- [ ] **Step 3: Implement archive service**

Create `internal/archive/service.go`:

```go
package archive

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

var (
	ErrChangeNotFound   = errors.New("change not found")
	ErrNotReady         = errors.New("only ready changes can be archived")
	ErrAlreadyArchived  = errors.New("change is already archived")
	ErrNotArchived      = errors.New("change is not archived")
)

type Service struct {
	db *bun.DB
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Archive(ctx context.Context, changeID int64, userID int64) (*models.Change, error) {
	change := new(models.Change)
	if err := s.db.NewSelect().Model(change).Where("id = ?", changeID).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	if change.Stage != models.StageReady {
		return nil, ErrNotReady
	}
	if change.ArchivedAt != nil {
		return nil, ErrAlreadyArchived
	}

	now := time.Now()
	change.ArchivedAt = &now
	if _, err := s.db.NewUpdate().Model(change).
		Column("archived_at").
		WherePK().
		Exec(ctx); err != nil {
		return nil, err
	}

	// Record workflow event
	event := &models.WorkflowEvent{
		ChangeID:  changeID,
		FromStage: string(change.Stage),
		ToStage:   string(change.Stage),
		Action:    "archive",
		UserID:    userID,
	}
	if _, err := s.db.NewInsert().Model(event).Exec(ctx); err != nil {
		slog.Error("failed to record archive event", "error", err)
	}

	return change, nil
}

func (s *Service) Unarchive(ctx context.Context, changeID int64, userID int64) (*models.Change, error) {
	change := new(models.Change)
	if err := s.db.NewSelect().Model(change).Where("id = ?", changeID).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	if change.ArchivedAt == nil {
		return nil, ErrNotArchived
	}

	change.ArchivedAt = nil
	if _, err := s.db.NewUpdate().Model(change).
		Column("archived_at").
		WherePK().
		Exec(ctx); err != nil {
		return nil, err
	}

	event := &models.WorkflowEvent{
		ChangeID:  changeID,
		FromStage: string(change.Stage),
		ToStage:   string(change.Stage),
		Action:    "unarchive",
		UserID:    userID,
	}
	if _, err := s.db.NewInsert().Model(event).Exec(ctx); err != nil {
		slog.Error("failed to record unarchive event", "error", err)
	}

	return change, nil
}

func (s *Service) GetPolicy(ctx context.Context, projectID int64) (*models.ArchivePolicy, error) {
	policy := new(models.ArchivePolicy)
	err := s.db.NewSelect().Model(policy).Where("project_id = ?", projectID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default policy
			return &models.ArchivePolicy{
				ProjectID: projectID,
				Mode:      models.ArchiveModeManual,
				Trigger:   models.TriggerTasksDone,
				DaysDelay: 0,
			}, nil
		}
		return nil, err
	}
	return policy, nil
}

func (s *Service) UpdatePolicy(ctx context.Context, policy *models.ArchivePolicy) (*models.ArchivePolicy, error) {
	policy.UpdatedAt = time.Now()
	if _, err := s.db.NewInsert().Model(policy).
		On("CONFLICT (project_id) DO UPDATE").
		Set("mode = EXCLUDED.mode").
		Set("trigger_type = EXCLUDED.trigger_type").
		Set("days_delay = EXCLUDED.days_delay").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx); err != nil {
		return nil, err
	}
	return policy, nil
}

// EvaluateAutoArchive checks if a change should be auto-archived based on project policy.
// Called when a task is completed or by the daily cron.
func (s *Service) EvaluateAutoArchive(ctx context.Context, changeID int64) (bool, error) {
	change := new(models.Change)
	if err := s.db.NewSelect().Model(change).Where("id = ?", changeID).Scan(ctx); err != nil {
		return false, err
	}

	if change.Stage != models.StageReady || change.ArchivedAt != nil {
		return false, nil
	}

	policy, err := s.GetPolicy(ctx, change.ProjectID)
	if err != nil {
		return false, err
	}
	if policy.Mode != models.ArchiveModeAuto {
		return false, nil
	}

	// Check all tasks done
	allDone, err := s.allTasksDone(ctx, changeID)
	if err != nil {
		return false, err
	}

	// Get ready timestamp from workflow events
	readyAt, err := s.getReadyTimestamp(ctx, changeID)
	if err != nil {
		return false, err
	}

	if !s.shouldAutoArchive(policy.TriggerType, policy.DaysDelay, allDone, readyAt) {
		return false, nil
	}

	// Auto-archive
	now := time.Now()
	change.ArchivedAt = &now
	if _, err := s.db.NewUpdate().Model(change).
		Column("archived_at").
		WherePK().
		Exec(ctx); err != nil {
		return false, err
	}

	event := &models.WorkflowEvent{
		ChangeID:  changeID,
		FromStage: string(change.Stage),
		ToStage:   string(change.Stage),
		Action:    "auto_archive",
		Reason:    "auto-archive policy triggered",
	}
	if _, err := s.db.NewInsert().Model(event).Exec(ctx); err != nil {
		slog.Error("failed to record auto_archive event", "error", err)
	}

	return true, nil
}

func (s *Service) shouldAutoArchive(trigger models.ArchiveTrigger, daysDelay int, allTasksDone bool, readyAt *time.Time) bool {
	switch trigger {
	case models.TriggerTasksDone:
		return allTasksDone
	case models.TriggerDaysAfterReady:
		if readyAt == nil {
			return false
		}
		return time.Since(*readyAt) >= time.Duration(daysDelay)*24*time.Hour
	case models.TriggerTasksDoneAndDays:
		if !allTasksDone || readyAt == nil {
			return false
		}
		return time.Since(*readyAt) >= time.Duration(daysDelay)*24*time.Hour
	default:
		return false
	}
}

func (s *Service) allTasksDone(ctx context.Context, changeID int64) (bool, error) {
	var total int
	if err := s.db.NewSelect().TableExpr("tasks").
		ColumnExpr("COUNT(*)").
		Where("change_id = ?", changeID).
		Scan(ctx, &total); err != nil {
		return false, err
	}
	if total == 0 {
		return false, nil // no tasks = not done
	}

	var doneCount int
	if err := s.db.NewSelect().TableExpr("tasks").
		ColumnExpr("COUNT(*)").
		Where("change_id = ?", changeID).
		Where("status = ?", "done").
		Scan(ctx, &doneCount); err != nil {
		return false, err
	}

	return doneCount == total, nil
}

func (s *Service) getReadyTimestamp(ctx context.Context, changeID int64) (*time.Time, error) {
	var readyAt time.Time
	err := s.db.NewSelect().TableExpr("workflow_events").
		ColumnExpr("created_at").
		Where("change_id = ?", changeID).
		Where("to_stage = ?", "ready").
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx, &readyAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &readyAt, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/ben/Projects/colign && go test ./internal/archive/... -v
```

Expected: all `shouldAutoArchive` tests pass

- [ ] **Step 5: Commit**

```bash
git add internal/archive/ internal/models/archive.go
git commit -m "feat: add archive service with auto-archive evaluation logic"
```

---

## Task 5: Block Workflow on Archived Changes

**Files:**
- Modify: `internal/workflow/service.go`
- Modify: `internal/workflow/service_test.go`

- [ ] **Step 1: Write test for archived change blocking**

Add to `internal/workflow/service_test.go`:

```go
func TestAdvance_ArchivedChange(t *testing.T) {
	// Advancing an archived change should return ErrChangeArchived
}
```

- [ ] **Step 2: Add ErrChangeArchived and guard in workflow service**

In `internal/workflow/service.go`, add the error:

```go
var ErrChangeArchived = errors.New("cannot modify archived change")
```

In `Advance()`, after fetching the change, add:

```go
if change.ArchivedAt != nil {
    return "", ErrChangeArchived
}
```

Add the same guard in `Revert()` and `EvaluateAndAdvance()`.

- [ ] **Step 3: Run tests**

```bash
cd /Users/ben/Projects/colign && go test ./internal/workflow/... -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/workflow/
git commit -m "feat: block workflow transitions on archived changes"
```

---

## Task 6: Update ListChanges with Archive Filter

**Files:**
- Modify: `internal/project/service.go` (ListChanges method)
- Modify: `internal/project/connect_handler.go` (ListChanges handler + changeToProto)

- [ ] **Step 1: Update ListChanges service method**

In `internal/project/service.go`, change `ListChanges` signature and add filter:

```go
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
```

- [ ] **Step 2: Update connect handler**

In `internal/project/connect_handler.go`:

Update `ListChanges` handler to pass filter:

```go
func (h *ConnectHandler) ListChanges(ctx context.Context, req *connect.Request[projectv1.ListChangesRequest]) (*connect.Response[projectv1.ListChangesResponse], error) {
	filter := "active"
	if req.Msg.Filter != nil {
		filter = *req.Msg.Filter
	}
	changes, err := h.service.ListChanges(ctx, req.Msg.ProjectId, filter)
	// ... rest unchanged
}
```

Update `changeToProto` to include `ArchivedAt`:

```go
func changeToProto(c *models.Change) *projectv1.Change {
	pc := &projectv1.Change{
		Id:        c.ID,
		ProjectId: c.ProjectID,
		Name:      c.Name,
		Stage:     string(c.Stage),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
	if c.ArchivedAt != nil {
		pc.ArchivedAt = timestamppb.New(*c.ArchivedAt)
	}
	return pc
}
```

- [ ] **Step 3: Run build to verify compilation**

```bash
cd /Users/ben/Projects/colign && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/project/
git commit -m "feat: add archive filter to ListChanges and update changeToProto"
```

---

## Task 7: Archive Connect Handlers

**Files:**
- Modify: `internal/project/connect_handler.go`

- [ ] **Step 1: Add Archive/Unarchive/Policy handlers**

Add to `internal/project/connect_handler.go`. The `ConnectHandler` needs an `archiveService` field:

```go
type ConnectHandler struct {
	service           *Service
	archiveService    *archive.Service
	jwtManager        *auth.JWTManager
	apiTokenValidator auth.APITokenValidator
}

func NewConnectHandler(service *Service, archiveService *archive.Service, jwtManager *auth.JWTManager, apiTokenValidator auth.APITokenValidator) *ConnectHandler {
	return &ConnectHandler{service: service, archiveService: archiveService, jwtManager: jwtManager, apiTokenValidator: apiTokenValidator}
}
```

**IMPORTANT:** Also update the `NewConnectHandler` call in `internal/server/server.go` to pass the archive service (create with `archive.NewService(s.db)`) so that the build doesn't break. This must happen in this same task.

Add a helper to check project role from a change ID (editor+ for archive/unarchive, owner for policy):

```go
func (h *ConnectHandler) checkChangeRole(ctx context.Context, changeID int64, userID int64, requiredRole string) error {
	var role string
	err := h.service.db.NewSelect().
		TableExpr("project_members pm").
		ColumnExpr("pm.role").
		Join("JOIN changes ch ON ch.project_id = pm.project_id").
		Where("ch.id = ?", changeID).
		Where("pm.user_id = ?", userID).
		Scan(ctx, &role)
	if err != nil {
		return connect.NewError(connect.CodePermissionDenied, errors.New("not a project member"))
	}
	if requiredRole == "owner" && role != "owner" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("owner access required"))
	}
	return nil
}
```

Then add handlers with permission checks:

```go
func (h *ConnectHandler) ArchiveChange(ctx context.Context, req *connect.Request[projectv1.ArchiveChangeRequest]) (*connect.Response[projectv1.ArchiveChangeResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}
	if err := h.checkChangeRole(ctx, req.Msg.ChangeId, claims.UserID, "editor"); err != nil {
		return nil, err
	}

	change, err := h.archiveService.Archive(ctx, req.Msg.ChangeId, claims.UserID)
	if err != nil {
		if errors.Is(err, archive.ErrChangeNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, archive.ErrNotReady) || errors.Is(err, archive.ErrAlreadyArchived) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.ArchiveChangeResponse{
		Change: changeToProto(change),
	}), nil
}

func (h *ConnectHandler) UnarchiveChange(ctx context.Context, req *connect.Request[projectv1.UnarchiveChangeRequest]) (*connect.Response[projectv1.UnarchiveChangeResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}
	if err := h.checkChangeRole(ctx, req.Msg.ChangeId, claims.UserID, "editor"); err != nil {
		return nil, err
	}

	change, err := h.archiveService.Unarchive(ctx, req.Msg.ChangeId, claims.UserID)
	if err != nil {
		if errors.Is(err, archive.ErrChangeNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		if errors.Is(err, archive.ErrNotArchived) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.UnarchiveChangeResponse{
		Change: changeToProto(change),
	}), nil
}

func (h *ConnectHandler) GetArchivePolicy(ctx context.Context, req *connect.Request[projectv1.GetArchivePolicyRequest]) (*connect.Response[projectv1.GetArchivePolicyResponse], error) {
	policy, err := h.archiveService.GetPolicy(ctx, req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.GetArchivePolicyResponse{
		Policy: &projectv1.ArchivePolicy{
			ProjectId: policy.ProjectID,
			Mode:      string(policy.Mode),
			Trigger:   string(policy.TriggerType),
			DaysDelay: int32(policy.DaysDelay),
		},
	}), nil
}

func (h *ConnectHandler) UpdateArchivePolicy(ctx context.Context, req *connect.Request[projectv1.UpdateArchivePolicyRequest]) (*connect.Response[projectv1.UpdateArchivePolicyResponse], error) {
	claims, err := h.extractClaims(ctx, req.Header().Get("Authorization"))
	if err != nil {
		return nil, err
	}
	// Owner-only
	var role string
	if dbErr := h.service.db.NewSelect().
		TableExpr("project_members").
		ColumnExpr("role").
		Where("project_id = ?", req.Msg.ProjectId).
		Where("user_id = ?", claims.UserID).
		Scan(ctx, &role); dbErr != nil || role != "owner" {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("owner access required"))
	}

	policy := &models.ArchivePolicy{
		ProjectID:   req.Msg.ProjectId,
		Mode:        models.ArchiveMode(req.Msg.Mode),
		TriggerType: models.ArchiveTrigger(req.Msg.Trigger),
		DaysDelay:   int(req.Msg.DaysDelay),
	}

	updated, err := h.archiveService.UpdatePolicy(ctx, policy)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&projectv1.UpdateArchivePolicyResponse{
		Policy: &projectv1.ArchivePolicy{
			ProjectId: updated.ProjectID,
			Mode:      string(updated.Mode),
			Trigger:   string(updated.TriggerType),
			DaysDelay: int32(updated.DaysDelay),
		},
	}), nil
}
```

- [ ] **Step 2: Run build**

```bash
cd /Users/ben/Projects/colign && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add internal/project/
git commit -m "feat: add archive/unarchive/policy connect handlers"
```

---

## Task 8: Auto-Archive Trigger on Task Completion

**Files:**
- Modify: `internal/task/service.go`

- [ ] **Step 1: Add archive service dependency to task service**

In `internal/task/service.go`, add an optional `ArchiveEvaluator` interface:

```go
type ArchiveEvaluator interface {
	EvaluateAutoArchive(ctx context.Context, changeID int64) (bool, error)
}

type Service struct {
	db               *bun.DB
	archiveEvaluator ArchiveEvaluator
}

func NewService(db *bun.DB, opts ...Option) *Service {
	s := &Service{db: db}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Option func(*Service)

func WithArchiveEvaluator(ae ArchiveEvaluator) Option {
	return func(s *Service) { s.archiveEvaluator = ae }
}
```

- [ ] **Step 2: Trigger evaluation on task status change to "done"**

In the `Update` method, after the DB update succeeds, add:

```go
// After successful update, check auto-archive if task completed
if status != nil && models.TaskStatus(*status) == models.TaskStatusDone && s.archiveEvaluator != nil {
	if _, err := s.archiveEvaluator.EvaluateAutoArchive(ctx, task.ChangeID); err != nil {
		slog.Error("auto-archive evaluation failed", "error", err, "change_id", task.ChangeID)
	}
}
```

- [ ] **Step 3: Run build**

```bash
cd /Users/ben/Projects/colign && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add internal/task/
git commit -m "feat: trigger auto-archive evaluation on task completion"
```

---

## Task 9: Daily Auto-Archive Cron

**Files:**
- Create: `internal/archive/cron.go`
- Create: `internal/archive/cron_test.go`

- [ ] **Step 1: Write cron test**

Create `internal/archive/cron_test.go`:

```go
package archive

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCronInterval(t *testing.T) {
	assert.Equal(t, 24*time.Hour, cronInterval)
}
```

- [ ] **Step 2: Implement cron**

Create `internal/archive/cron.go`:

```go
package archive

import (
	"context"
	"log/slog"
	"time"

	"github.com/gobenpark/colign/internal/models"
)

const cronInterval = 24 * time.Hour

// StartCron launches a background goroutine that evaluates auto-archive
// for all Ready changes daily. Cancel the context to stop.
func (s *Service) StartCron(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(cronInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("archive cron stopped")
				return
			case <-ticker.C:
				s.runAutoArchiveScan(ctx)
			}
		}
	}()
	slog.Info("archive cron started", "interval", cronInterval)
}

func (s *Service) runAutoArchiveScan(ctx context.Context) {
	var changes []models.Change
	err := s.db.NewSelect().Model(&changes).
		Where("stage = ?", models.StageReady).
		Where("archived_at IS NULL").
		Scan(ctx)
	if err != nil {
		slog.Error("archive cron: failed to list ready changes", "error", err)
		return
	}

	for _, change := range changes {
		archived, err := s.EvaluateAutoArchive(ctx, change.ID)
		if err != nil {
			slog.Error("archive cron: evaluation failed",
				"error", err,
				"change_id", change.ID,
			)
			continue
		}
		if archived {
			slog.Info("archive cron: auto-archived change",
				"change_id", change.ID,
			)
		}
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd /Users/ben/Projects/colign && go test ./internal/archive/... -v
```

- [ ] **Step 4: Commit**

```bash
git add internal/archive/cron.go internal/archive/cron_test.go
git commit -m "feat: add daily auto-archive cron scanner"
```

---

## Task 10: Wire Archive Service in Server

**Files:**
- Modify: `internal/server/server.go`

- [ ] **Step 1: Wire dependencies**

In `internal/server/server.go`:

1. Add import:
```go
import "github.com/gobenpark/colign/internal/archive"
```

2. Add a `cronCancel` field to the `Server` struct:
```go
type Server struct {
	// ...existing fields...
	cronCancel context.CancelFunc
}
```

3. In the server setup, after existing service initialization:
```go
// Archive service
archiveService := archive.NewService(s.db)
cronCtx, cronCancel := context.WithCancel(context.Background())
s.cronCancel = cronCancel
archiveService.StartCron(cronCtx)
```

4. In the server's `Shutdown()` or cleanup method:
```go
if s.cronCancel != nil {
	s.cronCancel()
}
```

5. Update project `NewConnectHandler` call to pass archiveService:
```go
projectConnectHandler := project.NewConnectHandler(projectService, archiveService, s.jwtManager, apiTokenService)
```

6. Update `task.NewService` call to pass the archive evaluator:
```go
taskService := task.NewService(s.db, task.WithArchiveEvaluator(archiveService))
```

- [ ] **Step 2: Run build**

```bash
cd /Users/ben/Projects/colign && go build ./cmd/api/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/server/
git commit -m "feat: wire archive service and cron into server"
```

---

## Task 11: Frontend — Active/Archived Tabs

**Files:**
- Modify: `web/src/app/projects/[slug]/page.tsx` (ChangesTab component)

- [ ] **Step 1: Add archive filter state and tabs**

In the `ChangesTab` component, add state and fetch logic. Note: `ChangesTab` receives `projectId` as a prop from the parent page which resolves it via `getProject(slug)`.

```tsx
const [archiveFilter, setArchiveFilter] = useState<"active" | "archived">("active");
const [activeCount, setActiveCount] = useState(0);
const [archivedCount, setArchivedCount] = useState(0);

// On initial load, fetch both to get counts
useEffect(() => {
  Promise.all([
    projectClient.listChanges({ projectId: BigInt(projectId), filter: "active" }),
    projectClient.listChanges({ projectId: BigInt(projectId), filter: "archived" }),
  ]).then(([activeRes, archivedRes]) => {
    setActiveCount(activeRes.changes.length);
    setArchivedCount(archivedRes.changes.length);
  });
}, [projectId]);
```

Add tab toggle above the change list. Use i18n keys for all UI strings (per project rule — never hardcode UI text):

```tsx
<div className="flex gap-1 rounded-lg bg-muted/50 p-1">
  <button
    onClick={() => setArchiveFilter("active")}
    className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
      archiveFilter === "active"
        ? "bg-background text-foreground shadow-sm"
        : "text-muted-foreground hover:text-foreground"
    }`}
  >
    {t("changes.filter.active")} ({activeCount})
  </button>
  <button
    onClick={() => setArchiveFilter("archived")}
    className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
      archiveFilter === "archived"
        ? "bg-background text-foreground shadow-sm"
        : "text-muted-foreground hover:text-foreground"
    }`}
  >
    {t("changes.filter.archived")} ({archivedCount})
  </button>
</div>
```

Add the required i18n keys to the translation files:
```json
{
  "changes.filter.active": "Active",
  "changes.filter.archived": "Archived",
  "changes.archive": "Archive",
  "changes.restore": "Restore",
  "settings.archive.title": "Archive Policy",
  "settings.archive.description": "Configure how completed changes are archived",
  "settings.archive.mode": "Mode",
  "settings.archive.mode.manual": "Manual",
  "settings.archive.mode.auto": "Automatic",
  "settings.archive.trigger": "Trigger",
  "settings.archive.trigger.tasks_done": "All tasks completed",
  "settings.archive.trigger.days_after_ready": "Days after ready",
  "settings.archive.trigger.tasks_done_and_days": "All tasks completed + days",
  "settings.archive.days_delay": "Days delay",
  "settings.archive.save": "Save Policy"
}
```

- [ ] **Step 2: Verify in browser**

Start dev server and navigate to a project's Changes tab. Verify:
- Active/Archived tabs render
- Switching tabs changes the displayed list
- Counts update correctly

- [ ] **Step 3: Commit**

```bash
cd /Users/ben/Projects/colign/web && git add . && git commit -m "feat: add active/archived tabs to changes list"
```

---

## Task 12: Frontend — Archive/Unarchive Buttons

**Files:**
- Modify: `web/src/app/projects/[slug]/changes/[changeId]/page.tsx`

- [ ] **Step 1: Add archive/unarchive buttons to change detail page**

In the change detail page, add buttons based on state:

```tsx
{/* Show Archive button when stage is Ready and not archived */}
{change.stage === "ready" && !change.archivedAt && (
  <Button
    variant="outline"
    onClick={async () => {
      await projectClient.archiveChange({ changeId: BigInt(change.id) });
      // refetch change
    }}
  >
    <Archive className="mr-2 h-4 w-4" />
    {t("changes.archive")}
  </Button>
)}

{/* Show Unarchive button when archived */}
{change.archivedAt && (
  <Button
    variant="outline"
    onClick={async () => {
      await projectClient.unarchiveChange({ changeId: BigInt(change.id) });
      // refetch change
    }}
  >
    <ArchiveRestore className="mr-2 h-4 w-4" />
    {t("changes.restore")}
  </Button>
)}
```

Import `Archive` and `ArchiveRestore` from `lucide-react`.

When archived, disable the workflow advance/revert buttons.

- [ ] **Step 2: Verify in browser**

- Ready stage change shows "Archive" button
- After archiving, shows "Restore" button
- Workflow buttons disabled when archived

- [ ] **Step 3: Commit**

```bash
cd /Users/ben/Projects/colign/web && git add . && git commit -m "feat: add archive/restore buttons to change detail page"
```

---

## Task 13: Frontend — Archive Policy Settings

**Files:**
- Modify: `web/src/app/projects/[slug]/settings/page.tsx`

- [ ] **Step 1: Add "Archive" tab to settings**

Add to the tabs array:

```tsx
{ id: "archive", label: "Archive Policy" },
```

Update the `SettingsTab` type to include `"archive"`.

- [ ] **Step 2: Add archive policy form**

All UI strings must use i18n `t()` keys (per project rule). The settings page has `slug` from `useParams()` but needs `projectId` (numeric) for API calls. Resolve it by fetching the project first (the existing page likely already does this or should be updated to do so).

```tsx
{activeTab === "archive" && (
  <Card className="border-border/50">
    <CardHeader>
      <CardTitle>{t("settings.archive.title")}</CardTitle>
      <CardDescription>
        {t("settings.archive.description")}
      </CardDescription>
    </CardHeader>
    <CardContent className="space-y-5">
      <div className="space-y-2">
        <Label>{t("settings.archive.mode")}</Label>
        <Select value={archiveMode} onValueChange={setArchiveMode}>
          <SelectTrigger className="cursor-pointer">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="manual">{t("settings.archive.mode.manual")}</SelectItem>
            <SelectItem value="auto">{t("settings.archive.mode.auto")}</SelectItem>
          </SelectContent>
        </Select>
      </div>
      {archiveMode === "auto" && (
        <>
          <div className="space-y-2">
            <Label>{t("settings.archive.trigger")}</Label>
            <Select value={archiveTrigger} onValueChange={setArchiveTrigger}>
              <SelectTrigger className="cursor-pointer">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="tasks_done">{t("settings.archive.trigger.tasks_done")}</SelectItem>
                <SelectItem value="days_after_ready">{t("settings.archive.trigger.days_after_ready")}</SelectItem>
                <SelectItem value="tasks_done_and_days">{t("settings.archive.trigger.tasks_done_and_days")}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {(archiveTrigger === "days_after_ready" || archiveTrigger === "tasks_done_and_days") && (
            <div className="space-y-2">
              <Label>{t("settings.archive.days_delay")}</Label>
              <Input
                type="number"
                min={1}
                max={90}
                value={archiveDaysDelay}
                onChange={(e) => setArchiveDaysDelay(Number(e.target.value))}
              />
            </div>
          )}
        </>
      )}
      <div className="flex items-center gap-3 pt-2">
        <Button onClick={() => handleSave("archive")} disabled={saving}>
          {saving ? t("common.saving") : t("settings.archive.save")}
        </Button>
        {saved === "archive" && <span className="text-sm text-emerald-400">{t("common.saved")}</span>}
      </div>
    </CardContent>
  </Card>
)}
```

- [ ] **Step 3: Add state and save handler**

```tsx
const [archiveMode, setArchiveMode] = useState("manual");
const [archiveTrigger, setArchiveTrigger] = useState("tasks_done");
const [archiveDaysDelay, setArchiveDaysDelay] = useState(7);
```

First, resolve `projectId` from `slug`. Add a `useEffect` to fetch the project at the top of the component:

```tsx
const [projectId, setProjectId] = useState<bigint | null>(null);

useEffect(() => {
  projectClient.getProject({ slug }).then(({ project }) => {
    if (project) setProjectId(project.id);
  });
}, [slug]);
```

In `handleSave`, add the archive case (guard with `projectId` check):

```tsx
if (section === "archive" && projectId) {
  await projectClient.updateArchivePolicy({
    projectId,
    mode: archiveMode,
    trigger: archiveTrigger,
    daysDelay: archiveDaysDelay,
  });
}
```

Load existing policy on mount:

```tsx
useEffect(() => {
  if (!projectId) return;
  projectClient.getArchivePolicy({ projectId })
    .then(({ policy }) => {
      if (policy) {
        setArchiveMode(policy.mode);
        setArchiveTrigger(policy.trigger);
        setArchiveDaysDelay(policy.daysDelay);
      }
    });
}, [projectId]);
```

- [ ] **Step 4: Verify in browser**

- Settings page shows "Archive Policy" tab
- Switching mode to "auto" reveals trigger options
- Selecting "days_after_ready" reveals days input
- Save persists and reloads correctly

- [ ] **Step 5: Commit**

```bash
cd /Users/ben/Projects/colign/web && git add . && git commit -m "feat: add archive policy settings tab"
```

---

## Task 14: Integration Verification

- [ ] **Step 1: Run all backend tests**

```bash
cd /Users/ben/Projects/colign && go test ./... -v
```

- [ ] **Step 2: Run linter**

```bash
cd /Users/ben/Projects/colign && golangci-lint run
```

- [ ] **Step 3: Build backend**

```bash
cd /Users/ben/Projects/colign && CGO_ENABLED=0 go build -o /tmp/colign-api ./cmd/api
```

- [ ] **Step 4: Build frontend**

```bash
cd /Users/ben/Projects/colign/web && npm run build
```

- [ ] **Step 5: Final commit if any fixes needed**

```bash
git add -A && git commit -m "fix: address lint and build issues for archive feature"
```
