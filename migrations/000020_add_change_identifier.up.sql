-- Add project identifier (e.g. COL, PLQ) for change numbering
ALTER TABLE projects ADD COLUMN identifier TEXT;

-- Backfill: generate identifier from first 3 chars of uppercased slug
-- Handle duplicates by appending a digit
DO $$
DECLARE
    r RECORD;
    base_id TEXT;
    candidate TEXT;
    suffix INT;
BEGIN
    FOR r IN SELECT id, organization_id, slug FROM projects ORDER BY id LOOP
        base_id := UPPER(LEFT(REPLACE(r.slug, '-', ''), 3));
        IF base_id = '' THEN base_id := 'PRJ'; END IF;
        candidate := base_id;
        suffix := 2;
        WHILE EXISTS (
            SELECT 1 FROM projects
            WHERE organization_id = r.organization_id
              AND identifier = candidate
              AND id != r.id
        ) LOOP
            candidate := base_id || suffix::TEXT;
            suffix := suffix + 1;
        END LOOP;
        UPDATE projects SET identifier = candidate WHERE id = r.id;
    END LOOP;
END $$;

ALTER TABLE projects ALTER COLUMN identifier SET NOT NULL;
ALTER TABLE projects ADD CONSTRAINT chk_identifier_length CHECK (LENGTH(identifier) <= 5);
CREATE UNIQUE INDEX idx_projects_org_identifier ON projects(organization_id, identifier);

-- Add change number (project-scoped sequential counter)
ALTER TABLE changes ADD COLUMN number INT;

-- Backfill: assign numbers per project ordered by created_at
DO $$
DECLARE
    r RECORD;
    seq INT;
    prev_pid BIGINT := -1;
BEGIN
    FOR r IN SELECT id, project_id FROM changes ORDER BY project_id, created_at, id LOOP
        IF r.project_id != prev_pid THEN
            seq := 1;
            prev_pid := r.project_id;
        ELSE
            seq := seq + 1;
        END IF;
        UPDATE changes SET number = seq WHERE id = r.id;
    END LOOP;
END $$;

ALTER TABLE changes ALTER COLUMN number SET NOT NULL;
CREATE UNIQUE INDEX idx_changes_project_number ON changes(project_id, number);
