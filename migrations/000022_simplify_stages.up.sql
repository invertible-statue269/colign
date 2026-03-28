-- Drop existing CHECK constraints first
ALTER TABLE changes DROP CONSTRAINT IF EXISTS changes_stage_check;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_type_check;
ALTER TABLE archive_policies DROP CONSTRAINT IF EXISTS archive_policies_trigger_type_check;

-- Migrate stage values: design→spec, review→approved, ready→approved
UPDATE changes SET stage = 'spec' WHERE stage = 'design';
UPDATE changes SET stage = 'approved' WHERE stage IN ('review', 'ready');

-- Add new CHECK constraint on changes.stage
ALTER TABLE changes ADD CONSTRAINT changes_stage_check CHECK (stage IN ('draft', 'spec', 'approved'));

-- Migrate document type: design→spec
UPDATE documents SET type = 'spec' WHERE type = 'design';

-- Add new CHECK constraint on documents.type
ALTER TABLE documents ADD CONSTRAINT documents_type_check CHECK (type IN ('proposal', 'spec', 'tasks'));

-- Migrate workflow event stage references
UPDATE workflow_events SET from_stage = 'spec' WHERE from_stage = 'design';
UPDATE workflow_events SET from_stage = 'approved' WHERE from_stage IN ('review', 'ready');
UPDATE workflow_events SET to_stage = 'spec' WHERE to_stage = 'design';
UPDATE workflow_events SET to_stage = 'approved' WHERE to_stage IN ('review', 'ready');

-- Update archive trigger type: days_after_ready→days_after_approved
UPDATE archive_policies SET trigger_type = 'days_after_approved' WHERE trigger_type = 'days_after_ready';
ALTER TABLE archive_policies ADD CONSTRAINT archive_policies_trigger_type_check
    CHECK (trigger_type IN ('tasks_done', 'days_after_approved', 'tasks_done_and_days'));
