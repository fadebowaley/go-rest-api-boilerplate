BEGIN;

CREATE TABLE IF NOT EXISTS project_updates (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    grant_application_id INTEGER NOT NULL REFERENCES grant_applications(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'planned',
    milestone_date TIMESTAMPTZ,
    photos JSONB DEFAULT '[]'::jsonb,
    invoices JSONB DEFAULT '[]'::jsonb,
    notes TEXT,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_project_updates_tenant_id ON project_updates(tenant_id);
CREATE INDEX idx_project_updates_grant_application_id ON project_updates(grant_application_id);
CREATE INDEX idx_project_updates_status ON project_updates(status);
CREATE INDEX idx_project_updates_deleted_at ON project_updates(deleted_at);

COMMIT;
