BEGIN;

INSERT INTO permissions (name, "group", description) VALUES

-- Auth
('auth:profile:read', 'auth', 'Read own profile'),
('auth:profile:update', 'auth', 'Update own profile'),
('auth:password:change', 'auth', 'Change own password'),

-- Users (own)
('user:read', 'users', 'Read user details'),
('user:update', 'users', 'Update user'),
('user:delete', 'users', 'Delete user'),

-- Tenant
('tenant:read', 'tenant', 'Read tenant details'),
('tenant:invite', 'tenant', 'Invite users to tenant'),
('tenant:users:list', 'tenant', 'List tenant users'),
('tenant:role:create', 'tenant', 'Create tenant role'),
('tenant:role:list', 'tenant', 'List tenant roles'),
('tenant:role:read', 'tenant', 'Read tenant role'),
('tenant:role:update', 'tenant', 'Update tenant role'),
('tenant:role:delete', 'tenant', 'Delete tenant role'),

-- Programs
('program:list', 'program', 'List programs'),
('program:create', 'program', 'Create program'),
('program:read', 'program', 'Read program details'),
('program:update', 'program', 'Update program'),
('program:delete', 'program', 'Delete program'),

-- Applicants
('applicant:list', 'applicant', 'List applicants'),
('applicant:create', 'applicant', 'Create applicant'),
('applicant:read', 'applicant', 'Read applicant details'),
('applicant:update', 'applicant', 'Update applicant'),
('applicant:delete', 'applicant', 'Delete applicant'),

-- Grants
('grant:create', 'grant', 'Create grant application'),
('grant:list', 'grant', 'List grant applications'),
('grant:read', 'grant', 'Read grant application details'),
('grant:update', 'grant', 'Update grant application'),
('grant:delete', 'grant', 'Delete grant application'),
('grant:submit', 'grant', 'Submit grant application'),
('grant:approve', 'grant', 'Approve grant application'),
('grant:reject', 'grant', 'Reject grant application'),
('grant:return', 'grant', 'Return grant application for revision'),

-- Documents
('document:upload', 'document', 'Upload document'),
('document:list', 'document', 'List documents'),
('document:read', 'document', 'Read document details'),
('document:update', 'document', 'Update document'),
('document:delete', 'document', 'Delete document'),
('document:download', 'document', 'Download document'),
('document:verify', 'document', 'Verify document'),
('document:reject', 'document', 'Reject document'),

-- Workflow
('workflow:read', 'workflow', 'Read workflow'),
('workflow:approve', 'workflow', 'Approve at workflow step'),
('workflow:reject', 'workflow', 'Reject at workflow step'),
('workflow:return', 'workflow', 'Return for revision'),
('workflow:history', 'workflow', 'View workflow history'),
('workflow:pending:list', 'workflow', 'List pending approvals'),
('workflow:template:list', 'workflow', 'List workflow templates'),
('workflow:template:create', 'workflow', 'Create workflow template'),
('workflow:template:read', 'workflow', 'Read workflow template'),
('workflow:template:update', 'workflow', 'Update workflow template'),
('workflow:template:delete', 'workflow', 'Delete workflow template'),
('workflow:step:list', 'workflow', 'List workflow steps'),
('workflow:step:create', 'workflow', 'Create workflow step'),
('workflow:step:read', 'workflow', 'Read workflow step'),
('workflow:step:update', 'workflow', 'Update workflow step'),
('workflow:step:delete', 'workflow', 'Delete workflow step'),

-- Disbursements
('disbursement:list', 'disbursement', 'List disbursements'),
('disbursement:create', 'disbursement', 'Create disbursement'),
('disbursement:read', 'disbursement', 'Read disbursement details'),
('disbursement:update', 'disbursement', 'Update disbursement'),
('disbursement:delete', 'disbursement', 'Delete disbursement'),

-- Project Updates
('update:list', 'project', 'List project updates'),
('update:create', 'project', 'Create project update'),
('update:read', 'project', 'Read project update'),
('update:update', 'project', 'Update project update'),
('update:delete', 'project', 'Delete project update'),

-- Forms
('form:template:list', 'form', 'List form templates'),
('form:template:create', 'form', 'Create form template'),
('form:template:read', 'form', 'Read form template'),
('form:template:update', 'form', 'Update form template'),
('form:template:delete', 'form', 'Delete form template'),
('form:template:publish', 'form', 'Publish form template'),
('form:template:version:list', 'form', 'List form template versions'),

-- Dashboard
('dashboard:summary', 'dashboard', 'View dashboard summary'),
('dashboard:finance:summary', 'dashboard', 'View finance summary'),
('dashboard:audit:summary', 'dashboard', 'View audit summary'),
('dashboard:program:stats', 'dashboard', 'View program statistics'),
('dashboard:workflow:stats', 'dashboard', 'View workflow statistics'),

-- Admin: Tenants
('admin:tenant:create', 'admin', 'Create tenant'),
('admin:tenant:list', 'admin', 'List all tenants'),
('admin:tenant:read', 'admin', 'Read any tenant'),
('admin:tenant:update', 'admin', 'Update any tenant'),
('admin:tenant:delete', 'admin', 'Delete tenant'),
('admin:tenant:user:add', 'admin', 'Add user to any tenant'),
('admin:tenant:user:list', 'admin', 'List users in any tenant'),
('admin:tenant:user:remove', 'admin', 'Remove user from any tenant'),

-- Admin: Users
('admin:user:list', 'admin', 'List all users'),
('admin:user:read', 'admin', 'Read any user'),
('admin:user:update', 'admin', 'Update any user'),
('admin:user:delete', 'admin', 'Delete user'),
('admin:user:role:assign', 'admin', 'Assign role to user'),
('admin:user:role:list', 'admin', 'List user roles'),
('admin:user:role:remove', 'admin', 'Remove role from user'),

-- Admin: Permissions
('admin:permission:create', 'admin', 'Create permission definition'),
('admin:permission:list', 'admin', 'List permission definitions'),
('admin:permission:delete', 'admin', 'Delete permission definition'),
('admin:role:permission:list', 'admin', 'List permissions for role'),
('admin:role:permission:assign', 'admin', 'Assign permissions to role'),

-- Admin: Grants
('admin:grant:list', 'admin', 'List all grants across tenants'),
('admin:grant:read', 'admin', 'Read any grant'),

-- Admin: Audit
('admin:audit:list', 'admin', 'List audit logs'),
('admin:audit:log', 'admin', 'Log action'),

-- Admin: Documents
('admin:document:verify', 'admin', 'Verify any document'),
('admin:document:reject', 'admin', 'Reject any document')

ON CONFLICT (name) DO NOTHING;

COMMIT;
