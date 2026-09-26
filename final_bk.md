You are a senior Go backend engineer and enterprise API architect.

We have built the backend API for Applico, a multi-tenant grant management platform. The API is based on go-rest-api-boilerplate and supports tenants, users, roles, grant programs, applicants, grant applications, documents, workflows, approvals, disbursements, project updates, dashboards, audit logs, and reporting.

Your task is NOT to add random new features. Your task is to stabilize, harden, verify, document, and make the backend frontend-ready.

1. Run a full backend lifecycle test

Create and verify the full journey:

- Platform admin creates tenant
- Platform admin creates/assigns tenant users
- Tenant admin creates grant program
- Tenant admin creates form template
- Tenant admin creates document requirements
- Tenant admin creates workflow template and workflow steps
- Applicant profile is created
- Applicant creates draft grant application
- Applicant uploads documents
- Applicant submits application
- Workflow instance starts automatically
- Reviewer views pending approval
- Reviewer verifies/rejects documents
- Reviewer approves, rejects, or returns grant
- Returned grant can be edited and re-submitted
- Final approval changes grant status to approved
- Finance creates disbursement
- Finance marks disbursement as paid
- Applicant submits project update
- Admin views dashboard summary
- Admin views audit trail

Fix all broken lifecycle points.

2. Enforce tenant isolation

Review every query, repository, service, and handler.

Every tenant-scoped business record must be protected by tenant_id.

Tables to verify:

- users / tenant_users
- roles
- permissions
- grant_programs
- form_templates
- applicants
- grant_applications
- documents
- workflow_templates
- workflow_steps
- workflow_instances
- workflow_actions
- disbursements
- project_updates
- audit_logs
- notifications

No user from Tenant A must access, modify, approve, or view Tenant B’s records, even by guessing IDs.

Add tests for cross-tenant access denial.

3. Verify RBAC and permission checks

Create a clear permission matrix and enforce it.

Minimum roles:

- platform_admin
- tenant_admin
- program_manager
- applicant
- reviewer
- finance_officer
- auditor

Verify:

- applicant cannot approve grants
- reviewer cannot manage tenants
- finance officer cannot alter workflow templates
- auditor can read but not modify
- tenant admin cannot access another tenant
- platform admin can manage tenants but should not bypass tenant audit trail silently

Add tests for permission denial.

4. Harden workflow engine

Workflow must support:

- configurable workflow templates
- ordered workflow steps
- role-based assignment
- specific-user assignment where available
- current step tracking
- approval history
- approve
- reject
- return for correction
- re-submit after return
- final approval status update
- comments required for reject/return
- immutable workflow history

Verify that workflow cannot:

- skip steps
- approve twice
- approve after rejection
- approve a draft
- approve without required permission
- continue after final approval

5. Harden dynamic form engine

Verify that grant programs support configurable forms.

Each grant application should preserve:

- form_template_id
- form_version_id
- form_response JSONB
- promoted reporting fields such as requested_amount, project_title, applicant_id, program_id, status, submitted_at

Ensure form templates support:

- versioning
- publishing
- draft/edit mode
- required fields
- field validation
- safe JSON schema validation
- preserving old submitted form versions

Do not allow changes to a published form version to corrupt old applications.

6. Harden document management

Verify document flow:

- upload document
- list documents
- replace document
- delete document only when allowed
- verify document
- reject document with notes
- preserve document metadata
- preserve document history
- support required document checking before submission or before approval

Every document must store:

- tenant_id
- grant_application_id
- document_type
- file_name
- file_path or storage_key
- mime_type
- file_size
- uploaded_by
- verification_status
- verified_by
- verified_at
- rejection_reason

Prepare storage abstraction for local now and S3/MinIO later.

7. Harden disbursement module

Verify:

- disbursement can only be created after grant approval
- total disbursement cannot exceed approved amount
- finance officer permission is required
- payment status lifecycle is valid
- payment reference is unique where required
- paid date is recorded
- cancellation rules are safe
- audit logs are created

Statuses should include:

- scheduled
- processing
- paid
- cancelled
- failed

8. Harden audit logging

Audit logs must be automatic and consistent.

Track:

- actor_user_id
- tenant_id
- entity_type
- entity_id
- action
- old_value
- new_value
- ip_address
- user_agent
- created_at

Audit must cover:

- login
- user invite/role change
- tenant creation/update
- grant program creation/update
- form template publish/update
- applicant creation/update
- grant draft creation/update
- grant submission
- document upload/replace/delete/verify/reject
- workflow approve/reject/return
- disbursement creation/update/payment
- project update creation/update

Audit logs should be append-only. Do not allow normal users to update or delete audit logs.

9. Standardize API responses

Make all frontend-facing responses consistent.

Success format:

{
  "success": true,
  "message": "Operation successful",
  "data": {},
  "meta": {}
}

Error format:

{
  "success": false,
  "message": "Human readable error",
  "error": {
    "code": "ERROR_CODE",
    "details": {}
  }
}

Pagination format:

{
  "success": true,
  "data": [],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}

Use consistent HTTP codes:

- 200 success
- 201 created
- 400 validation error
- 401 unauthenticated
- 403 unauthorized
- 404 not found
- 409 conflict
- 422 business rule failure
- 500 server error

10. Finalize Swagger/OpenAPI documentation

Swagger must include:

- all endpoints
- auth requirements
- tenant header/context requirement
- request examples
- response examples
- validation errors
- pagination examples
- file upload examples
- workflow action examples
- role permission notes

Group docs by:

- Auth
- Platform Admin
- Tenant Management
- Users & Roles
- Grant Programs
- Forms
- Applicants
- Grant Applications
- Documents
- Workflow
- Disbursements
- Project Updates
- Dashboard
- Audit Logs

11. Create seed and demo scripts

Create seed data for:

- platform admin
- tenant: RCCG Homeland Mission
- tenant admin
- program manager
- reviewer
- finance officer
- auditor
- applicant
- grant program: Church Building Support Grant
- form template
- required documents
- workflow template
- workflow steps

Also create at least three reusable tenant/program templates:

- Faith-based building grant
- NGO education grant
- SME support grant
- Community development grant

12. Create API test collection

Create Postman/Insomnia/HTTP collection covering:

- auth
- tenant setup
- user role assignment
- grant program setup
- form setup
- workflow setup
- applicant setup
- application draft
- document upload
- submit
- approve
- return
- resubmit
- reject
- disbursement
- project update
- dashboard
- audit logs

The collection must run in order as a full demo.

13. Add integration tests

At minimum, add tests for:

- tenant isolation
- RBAC denial
- grant lifecycle
- workflow progression
- return and resubmission
- document verification
- disbursement limit enforcement
- audit log creation
- form version preservation

14. Frontend readiness checklist

Before frontend begins, confirm:

- API base URL is configurable
- CORS is configured
- JWT refresh flow works
- tenant context is clear
- all enums are documented
- all list endpoints support pagination
- all list endpoints support useful filters
- file upload works from browser clients
- Swagger is accurate
- demo seed can be loaded with one command
- full lifecycle works without manual database changes

Final deliverable:

Produce a backend readiness report with:

1. What was tested
2. What passed
3. What failed and was fixed
4. Remaining risks
5. API endpoints ready for frontend
6. Known limitations
7. Recommended frontend build order

Do this as backend stabilization and product readiness work, not feature expansion.
