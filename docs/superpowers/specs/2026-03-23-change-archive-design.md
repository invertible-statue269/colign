# Change Archive Feature Design

## Problem

프로젝트가 오래 지속되면 Change 목록이 수십~수백 개로 쌓여서 "지금 진행 중인 게 뭔지" 파악하기 어려워진다. 완료된 Change와 활성 Change를 분리해야 한다.

## Solution

`archived_at` 타임스탬프 필드를 Change 모델에 추가하고, 프로젝트별 아카이브 정책(수동/자동)을 설정할 수 있게 한다.

## Data Model

### Change 모델 변경

`changes` 테이블에 nullable 컬럼 추가:

```sql
ALTER TABLE changes ADD COLUMN archived_at TIMESTAMPTZ;
CREATE INDEX idx_changes_active ON changes (project_id) WHERE archived_at IS NULL;
```

```go
type Change struct {
    ...existing fields...
    ArchivedAt *time.Time `bun:"archived_at"`
}
```

- `archived_at IS NULL` → Active
- `archived_at IS NOT NULL` → Archived
- 복원 시 `archived_at = NULL`로 되돌림
- Archived 상태의 Change는 stage 전환(advance/revert) 불가 — 먼저 unarchive해야 함

### ArchivePolicy 모델 (신규)

```sql
CREATE TABLE archive_policies (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL UNIQUE REFERENCES projects(id),
    mode TEXT NOT NULL DEFAULT 'manual',
    trigger TEXT NOT NULL DEFAULT 'tasks_done',
    days_delay INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

```go
type ArchivePolicy struct {
    ID        int64     `bun:"id,pk,autoincrement"`
    ProjectID int64     `bun:"project_id,notnull,unique"`
    Mode      string    `bun:"mode,notnull,default:'manual'"`
    Trigger   string    `bun:"trigger,notnull,default:'tasks_done'"`
    DaysDelay int       `bun:"days_delay,notnull,default:0"`
    CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
    UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
```

필드 설명:
- `mode`: `manual` (기본값) 또는 `auto`
- `trigger`: 자동 모드일 때 트리거 조건
  - `tasks_done` — 모든 Task 완료 시 즉시
  - `days_after_ready` — Ready 전환 후 N일 경과
  - `tasks_done_and_days` — 모든 Task 완료 + N일 경과
- `days_delay`: 일수 기반 트리거에서 사용하는 경과 일수

**ArchivePolicy 부재 시 처리:** `archive_policies` 행이 없는 프로젝트는 `mode: manual`로 간주. 설정 저장 시 upsert로 생성/갱신.

### Migration

단일 마이그레이션 파일로 처리:
1. `changes` 테이블에 `archived_at` 컬럼 추가
2. `archive_policies` 테이블 생성
3. Partial index 생성

기존 프로젝트에 대한 `archive_policies` 행 backfill 불필요 — 코드에서 부재 시 manual 기본값 처리.

## API

### Proto 변경

`Change` 메시지에 `archived_at` 필드 추가:

```protobuf
message Change {
  int64 id = 1;
  int64 project_id = 2;
  string name = 3;
  string stage = 4;
  google.protobuf.Timestamp created_at = 5;
  google.protobuf.Timestamp updated_at = 6;
  optional google.protobuf.Timestamp archived_at = 7;
}
```

`ListChangesRequest`에 필터 추가:

```protobuf
message ListChangesRequest {
  int64 project_id = 1;
  optional string filter = 2;  // "active" (기본값), "archived", "all"
}
```

`optional string`으로 "active", "archived", "all" 세 가지 상태를 표현. 미설정 시 "active" 기본값.

### 새로 추가할 RPC

`ProjectService`에 추가 (Change CRUD가 이미 있는 서비스):

| RPC | Request | Response | Description |
|---|---|---|---|
| `ArchiveChange` | `change_id` | `Change` | archived_at = now() |
| `UnarchiveChange` | `change_id` | `Change` | archived_at = null |
| `GetArchivePolicy` | `project_id` | `ArchivePolicy` | 프로젝트 아카이브 설정 조회 |
| `UpdateArchivePolicy` | `project_id, mode, trigger, days_delay` | `ArchivePolicy` | 프로젝트 아카이브 설정 저장 (upsert) |

### 권한

- `ArchiveChange` / `UnarchiveChange`: 프로젝트 editor 이상
- `GetArchivePolicy` / `UpdateArchivePolicy`: 프로젝트 owner만

## Auto-Archive Logic

### Ready 시점 기록

`days_after_ready` 트리거를 위해 Ready 전환 시점이 필요. `workflow_events` 테이블에서 `to_stage = 'ready'`인 가장 최근 이벤트의 `created_at`을 사용.

- Ready에서 revert 후 다시 Ready로 진입하면, 가장 최근 Ready 전환 시점이 기준이 됨
- 별도 컬럼 추가 없이 기존 데이터 활용

### 트리거 시점

1. **Task 완료 이벤트** — Task 상태가 done으로 변경될 때 `tasks_done`, `tasks_done_and_days` 트리거 평가
2. **크론잡 (1일 1회)** — `days_after_ready`, `tasks_done_and_days` 트리거의 시간 경과 조건 평가

### 크론잡 구현

API 프로세스 내에서 goroutine + ticker로 구현 (별도 바이너리 불필요):

```go
// 1일 1회 실행, 멱등성 보장 (이미 archived면 skip)
func (s *ArchiveService) RunDailyCheck(ctx context.Context)
```

- 단일 레플리카 환경이므로 분산 잠금 불필요
- 멱등: 이미 `archived_at`이 있는 Change는 건너뜀

### 평가 흐름

```
이벤트 발생 (Task 완료 or 크론잡)
  │
  ├─ Change의 프로젝트 ArchivePolicy 조회 (없으면 manual → 종료)
  ├─ mode == "manual"? → 종료
  ├─ Change가 Ready 단계? → 아닌 경우 종료
  ├─ 이미 archived? → 종료
  ├─ trigger 조건 평가:
  │    ├─ tasks_done: 모든 Task done인가?
  │    ├─ days_after_ready: 최근 Ready 전환 시점 + days_delay ≤ now?
  │    └─ tasks_done_and_days: 위 두 조건 모두 충족?
  │
  └─ 조건 충족 → archived_at = now(), WorkflowEvent 기록
```

### 이력 기록

아카이브/복원 액션은 WorkflowEvent로 기록. `from_stage`와 `to_stage`는 모두 현재 stage 값(보통 "ready")으로 설정:

| Action | from_stage | to_stage | UserID |
|---|---|---|---|
| `archive` | "ready" | "ready" | 실행한 사용자 |
| `auto_archive` | "ready" | "ready" | 0 (시스템) |
| `unarchive` | "ready" | "ready" | 실행한 사용자 |

stage 전환이 아니므로 `from_stage == to_stage`이지만, 기존 모델을 재사용하여 별도 테이블 없이 이력을 남긴다.

## Frontend

### Changes 탭 - Active/Archived 토글

프로젝트 Changes 목록 상단에 탭 추가:

```
[Active (5)]  [Archived (12)]
```

- Active가 기본 선택
- 각 탭에 해당 개수 표시
- 초기 로드 시 `ListChanges(filter="active")`와 `ListChanges(filter="archived")` 병렬 호출하여 양쪽 개수 확보
- 탭 전환 시 캐시된 데이터 사용 (이미 로드됨)

### Change 상세 페이지

- Ready 단계일 때 "아카이브" 버튼 노출
- Archived 상태에서 "복원" 버튼 노출
- Archived 상태에서 workflow advance/revert 버튼 비활성화

### 프로젝트 설정 페이지

아카이브 섹션 (owner만 접근 가능):

```
아카이브
├── 모드: [수동 ▼] / [자동 ▼]
└── (자동 선택 시 펼쳐짐)
      ├── 트리거: [모든 Task 완료 시 ▼]
      └── (days 관련 트리거 선택 시)
            └── 경과 일수: [7] 일
```

## Scope Out

- Epic/Initiative 계층 — 아카이브로 목록 관리 문제를 먼저 해결, 그룹핑은 나중에
- 아카이브 시 읽기 전용 모드 — 복원하면 바로 편집 가능
- 일괄 아카이브 — 초기 버전에서는 개별 아카이브만 지원
- MCP 도구 추가 — 추후 archive/unarchive MCP 도구 추가 가능
