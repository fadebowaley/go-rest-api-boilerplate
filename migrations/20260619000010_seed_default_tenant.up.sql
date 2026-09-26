BEGIN;

INSERT INTO tenants (id, name, slug, type, settings)
VALUES (1, 'RCCG Homeland Mission', 'rccg-homeland', 'faith_based', '{"description": "Default tenant for RCCG Homeland Mission grant management"}')
ON CONFLICT (id) DO NOTHING;

UPDATE grant_applications SET tenant_id = 1 WHERE tenant_id IS NULL;

UPDATE documents SET tenant_id = 1 FROM grant_applications
WHERE documents.grant_application_id = grant_applications.id AND documents.tenant_id IS NULL;

UPDATE audit_logs SET tenant_id = 1 WHERE tenant_id IS NULL;

COMMIT;
