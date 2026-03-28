-- Revert archive trigger type
ALTER TABLE archive_policies DROP CONSTRAINT IF EXISTS archive_policies_trigger_type_check;
UPDATE archive_policies SET trigger_type = 'days_after_ready' WHERE trigger_type = 'days_after_approved';
ALTER TABLE archive_policies ADD CONSTRAINT archive_policies_trigger_type_check
    CHECK (trigger_type IN ('tasks_done', 'days_after_ready', 'tasks_done_and_days'));

-- Revert workflow event stage references
UPDATE workflow_events SET to_stage = 'ready' WHERE to_stage = 'approved';
UPDATE workflow_events SET to_stage = 'design' WHERE to_stage = 'spec';
UPDATE workflow_events SET from_stage = 'ready' WHERE from_stage = 'approved';
UPDATE workflow_events SET from_stage = 'design' WHERE from_stage = 'spec';

-- Revert document type
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_type_check;
UPDATE documents SET type = 'design' WHERE type = 'spec';
ALTER TABLE documents ADD CONSTRAINT documents_type_check CHECK (type IN ('proposal', 'design', 'tasks'));

-- Revert stage values (approved→ready, spec→design)
ALTER TABLE changes DROP CONSTRAINT IF EXISTS changes_stage_check;
UPDATE changes SET stage = 'design' WHERE stage = 'spec';
UPDATE changes SET stage = 'ready' WHERE stage = 'approved';
ALTER TABLE changes ADD CONSTRAINT changes_stage_check CHECK (stage IN ('draft', 'design', 'review', 'ready'));
