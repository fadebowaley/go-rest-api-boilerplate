BEGIN;

CREATE TABLE IF NOT EXISTS form_templates (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema_definition JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_form_templates_tenant ON form_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_form_templates_status ON form_templates(status);

CREATE TABLE IF NOT EXISTS form_versions (
    id SERIAL PRIMARY KEY,
    form_template_id INTEGER NOT NULL REFERENCES form_templates(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    schema_definition JSONB NOT NULL DEFAULT '{}'::jsonb,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(form_template_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_form_versions_template ON form_versions(form_template_id);

CREATE TABLE IF NOT EXISTS grant_application_versions (
    id SERIAL PRIMARY KEY,
    grant_application_id INTEGER NOT NULL REFERENCES grant_applications(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    form_template_id INTEGER REFERENCES form_templates(id),
    form_version_id INTEGER REFERENCES form_versions(id),
    form_response_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    document_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,
    status_snapshot VARCHAR(30) NOT NULL DEFAULT 'draft',
    reason VARCHAR(30) NOT NULL DEFAULT 'submitted',
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(grant_application_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_gav_application ON grant_application_versions(grant_application_id);
CREATE INDEX IF NOT EXISTS idx_gav_reason ON grant_application_versions(reason);

ALTER TABLE grant_applications
    ADD COLUMN IF NOT EXISTS form_template_id INTEGER REFERENCES form_templates(id),
    ADD COLUMN IF NOT EXISTS form_version_id INTEGER REFERENCES form_versions(id);

ALTER TABLE grant_programs
    ADD COLUMN IF NOT EXISTS form_template_id INTEGER REFERENCES form_templates(id);

COMMIT;
