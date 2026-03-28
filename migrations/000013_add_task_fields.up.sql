ALTER TABLE tasks ADD COLUMN assignee_id BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN creator_id BIGINT REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_creator_id ON tasks(creator_id) WHERE creator_id IS NOT NULL;
