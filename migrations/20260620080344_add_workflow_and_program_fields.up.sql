BEGIN;

-- Add tenant_id as nullable first, backfill, then set NOT NULL
ALTER TABLE application_workflows
    ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE;

UPDATE application_workflows aw
SET tenant_id = ga.tenant_id
FROM grant_applications ga
WHERE aw.application_id = ga.id AND aw.tenant_id IS NULL;

ALTER TABLE application_workflows
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE workflow_actions
    ADD COLUMN IF NOT EXISTS tenant_id INTEGER REFERENCES tenants(id) ON DELETE CASCADE;

UPDATE workflow_actions wa
SET tenant_id = aw.tenant_id
FROM application_workflows aw
WHERE wa.application_workflow_id = aw.id AND wa.tenant_id IS NULL;

ALTER TABLE workflow_actions
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE grant_programs
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    ADD COLUMN IF NOT EXISTS eligibility_rules JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE workflow_steps
    ADD COLUMN IF NOT EXISTS assignee_user_id INTEGER REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS sla_hours INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS require_document_verification BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_final_approval BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_application_workflows_tenant ON application_workflows(tenant_id);
CREATE INDEX IF NOT EXISTS idx_workflow_actions_tenant ON workflow_actions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_workflow_steps_assignee ON workflow_steps(assignee_user_id);

COMMIT;
