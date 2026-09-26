BEGIN;

CREATE TABLE IF NOT EXISTS tenant_roles (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_tenant_roles_tenant_id ON tenant_roles(tenant_id);
CREATE INDEX idx_tenant_roles_deleted_at ON tenant_roles(deleted_at);

COMMIT;
