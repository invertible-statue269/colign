DROP INDEX IF EXISTS idx_changes_project_number;
ALTER TABLE changes DROP COLUMN IF EXISTS number;

DROP INDEX IF EXISTS idx_projects_org_identifier;
ALTER TABLE projects DROP COLUMN IF EXISTS identifier;
