ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_projects_organization;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
