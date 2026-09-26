BEGIN;

CREATE TABLE IF NOT EXISTS disbursements (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    grant_application_id INTEGER NOT NULL REFERENCES grant_applications(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(30) NOT NULL DEFAULT 'scheduled',
    payment_date TIMESTAMPTZ,
    evidence_url TEXT,
    notes TEXT,
    paid_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_disbursements_tenant_id ON disbursements(tenant_id);
CREATE INDEX idx_disbursements_grant_application_id ON disbursements(grant_application_id);
CREATE INDEX idx_disbursements_status ON disbursements(status);
CREATE INDEX idx_disbursements_deleted_at ON disbursements(deleted_at);

COMMIT;
