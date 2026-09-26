BEGIN;

-- super_admin gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT DO NOTHING;

-- admin gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- homeland_admin gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'homeland_admin'
ON CONFLICT DO NOTHING;

-- tenant_admin gets all permissions EXCEPT platform admin (full control within their tenant)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'tenant_admin'
  AND p."group" NOT LIKE 'admin%'
ON CONFLICT DO NOTHING;

-- grant_officer gets grant processing permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'grant_officer'
  AND p.name IN (
    'grant:read', 'grant:list',
    'grant:approve', 'grant:reject', 'grant:return',
    'document:list', 'document:read', 'document:download',
    'document:verify', 'document:reject',
    'workflow:read', 'workflow:history',
    'workflow:approve', 'workflow:reject', 'workflow:return',
    'workflow:pending:list',
    'dashboard:summary', 'dashboard:workflow:stats'
  )
ON CONFLICT DO NOTHING;

-- finance_officer gets disbursement permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'finance_officer'
  AND p.name IN (
    'grant:read', 'grant:list',
    'disbursement:create', 'disbursement:read', 'disbursement:list',
    'disbursement:update', 'disbursement:delete',
    'dashboard:summary', 'dashboard:finance:summary'
  )
ON CONFLICT DO NOTHING;

-- auditor gets read-only + audit permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'auditor'
  AND p.name IN (
    'grant:read', 'grant:list',
    'document:list', 'document:read', 'document:download',
    'workflow:read', 'workflow:history',
    'dashboard:summary', 'dashboard:audit:summary',
    'disbursement:read', 'disbursement:list'
  )
ON CONFLICT DO NOTHING;

-- applicant gets grant submission permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'applicant'
  AND p.name IN (
    'grant:create', 'grant:list', 'grant:read', 'grant:submit',
    'document:upload', 'document:list', 'document:read', 'document:delete',
    'update:create', 'update:list'
  )
ON CONFLICT DO NOTHING;

-- program_manager gets program management permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'program_manager'
  AND p."group" IN ('program', 'applicant', 'dashboard')
ON CONFLICT DO NOTHING;

-- reviewer gets review permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'reviewer'
  AND p.name IN (
    'grant:read', 'grant:list',
    'grant:approve', 'grant:reject', 'grant:return',
    'document:list', 'document:read', 'document:download',
    'workflow:read', 'workflow:history',
    'workflow:approve', 'workflow:reject', 'workflow:return',
    'workflow:pending:list',
    'dashboard:summary', 'dashboard:workflow:stats'
  )
ON CONFLICT DO NOTHING;

-- committee_member gets review + approval permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'committee_member'
  AND p.name IN (
    'grant:read', 'grant:list',
    'grant:approve', 'grant:reject',
    'document:list', 'document:read',
    'workflow:read', 'workflow:history',
    'workflow:approve', 'workflow:reject',
    'workflow:pending:list',
    'dashboard:summary'
  )
ON CONFLICT DO NOTHING;

-- regional/provincial/zonal/area reviewers get read + review permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name IN ('regional_reviewer', 'provincial_reviewer', 'zonal_reviewer', 'area_reviewer')
  AND p.name IN (
    'grant:read', 'grant:list',
    'grant:approve', 'grant:reject', 'grant:return',
    'document:list', 'document:read',
    'workflow:read', 'workflow:history',
    'workflow:approve', 'workflow:reject', 'workflow:return',
    'workflow:pending:list',
    'dashboard:summary'
  )
ON CONFLICT DO NOTHING;

-- parish_pastor gets grant submission + read permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'parish_pastor'
  AND p.name IN (
    'grant:create', 'grant:list', 'grant:read', 'grant:submit',
    'document:upload', 'document:list', 'document:read',
    'update:create', 'update:list',
    'dashboard:summary'
  )
ON CONFLICT DO NOTHING;

COMMIT;
