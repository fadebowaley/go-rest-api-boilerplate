BEGIN;

ALTER TABLE grant_applications
    ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id),
    ADD COLUMN IF NOT EXISTS program_id INTEGER REFERENCES grant_programs(id),
    ADD COLUMN IF NOT EXISTS applicant_id INTEGER REFERENCES applicants(id),
    ADD COLUMN IF NOT EXISTS workflow_template_id INTEGER REFERENCES workflow_templates(id),
    ADD COLUMN IF NOT EXISTS current_step_id INTEGER REFERENCES workflow_steps(id),
    ADD COLUMN IF NOT EXISTS form_responses JSONB DEFAULT '{}'::jsonb;

ALTER TABLE grant_applications
    ALTER COLUMN parish_id DROP NOT NULL,
    ALTER COLUMN building_stage DROP NOT NULL;

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id);

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id);

CREATE INDEX IF NOT EXISTS idx_grant_applications_tenant ON grant_applications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_grant_applications_program ON grant_applications(program_id);
CREATE INDEX IF NOT EXISTS idx_grant_applications_applicant ON grant_applications(applicant_id);
CREATE INDEX IF NOT EXISTS idx_documents_tenant ON documents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant ON audit_logs(tenant_id);

COMMIT;
