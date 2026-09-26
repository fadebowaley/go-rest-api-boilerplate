BEGIN;

DROP INDEX IF EXISTS idx_application_workflows_tenant;
DROP INDEX IF EXISTS idx_workflow_actions_tenant;
DROP INDEX IF EXISTS idx_workflow_steps_assignee;

ALTER TABLE application_workflows DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE workflow_actions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE grant_programs DROP COLUMN IF EXISTS currency;
ALTER TABLE grant_programs DROP COLUMN IF EXISTS eligibility_rules;
ALTER TABLE workflow_steps DROP COLUMN IF EXISTS assignee_user_id;
ALTER TABLE workflow_steps DROP COLUMN IF EXISTS sla_hours;
ALTER TABLE workflow_steps DROP COLUMN IF EXISTS require_document_verification;
ALTER TABLE workflow_steps DROP COLUMN IF EXISTS is_final_approval;

COMMIT;
