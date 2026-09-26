BEGIN;
DROP INDEX IF EXISTS idx_grant_applications_tenant;
DROP INDEX IF EXISTS idx_grant_applications_program;
DROP INDEX IF EXISTS idx_grant_applications_applicant;
DROP INDEX IF EXISTS idx_documents_tenant;
DROP INDEX IF EXISTS idx_audit_logs_tenant;

ALTER TABLE grant_applications DROP COLUMN IF EXISTS form_responses;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS current_step_id;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS workflow_template_id;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS applicant_id;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS program_id;
ALTER TABLE grant_applications DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE documents DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS tenant_id;
COMMIT;
