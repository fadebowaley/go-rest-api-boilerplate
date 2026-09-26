# MissionGrant Suite — Multi-Tenant Grant Management Platform

## Vision

A generic, multi-tenant grant management platform that serves faith-based organizations, NGOs, foundations, CSR teams, government agencies, and development programs. Any organization can create grant programs, define applicant types, configure application forms, set up approval workflows, manage disbursements, and track outcomes.

No hardcoded RCCG-specific concepts. All organization-specific configuration is data-driven.

## Architecture

### Multi-Tenant Data Model

Every business record belongs to a tenant (organization). Tenant isolation is enforced at the database query level via `tenant_id`.

```
tenants ──┬── tenant_users (user membership per tenant)
          ├── grant_programs (grant offerings per tenant)
          ├── applicants (generic applying entities)
          ├── workflow_templates (approval workflow designs)
          ├── grant_applications (scoped to program + applicant)
          ├── documents (scoped to application)
          ├── disbursements (scoped to application)
          ├── project_updates (scoped to application)
          └── audit_logs
```

### Users

Users are global identities (email/password for auth). Membership in a tenant is managed via `tenant_users`, which stores the user's roles per tenant as a JSONB array.

### Tenants

Each tenant represents an organization. A tenant configures:
- Organization name, type, and settings
- Grant programs with custom form schemas
- Applicant types accepted
- Required document types
- Approval workflow templates with configurable steps
- Roles and permissions

### Grant Programs

Each tenant creates grant programs. A program defines:
- Application form schema (JSONB) — dynamic fields
- Required document types
- Accepted applicant types
- Budget, timeline, and status

### Applicants

Generic entities that apply for grants. Each has:
- A configurable type (e.g. "church", "school", "community", "individual", "cooperative")
- `profile_data` (JSONB) for any dynamic fields per tenant/applicant type

### Workflows

Configurable approval workflows replace fixed approval levels:
- **Workflow templates** — designed per tenant and optionally assigned to programs
- **Workflow steps** — ordered steps with configurable name, assignee roles, and required actions
- **Application workflows** — runtime instances tracking the current step and history

### JSONB for Dynamic Data

- `tenants.settings` — per-tenant configuration
- `applicants.profile_data` — dynamic profile fields per applicant type
- `grant_applications.form_responses` — dynamic answers to program form schema
- `workflow_steps.assignee_roles` — which roles can act at each step
- `tenant_users.roles` — user's role assignments per tenant
- `grant_programs.form_schema` — defines the dynamic application form fields

## Core Modules

### 1. Authentication & Users ✅
- Global JWT authentication
- User registration, login, password reset
- Tenant membership via tenant_users
- Roles per tenant (not global)

### 2. Tenant Management ✅
- CRUD for tenants
- Tenant settings (JSONB)
- Tenant user management

### 3. Grant Program Management ✅
- CRUD for grant programs per tenant
- Configurable form schema
- Required document types per program

### 4. Applicant Management ✅
- CRUD for applicants per tenant
- Configurable applicant types
- JSONB profile data

### 5. Grant Application Module ✅
- Create draft application
- Submit to program
- Dynamic form responses (JSONB)
- Link to applicant + program
- Status driven by workflow

### 6. Document Management ✅
- Upload per application
- Document types per tenant config
- Storage abstraction (local → S3/MinIO later)

### 7. Configurable Workflow ✅
- Design workflow templates with steps
- Assign templates to programs
- Step-by-step approval tracking (approve, reject, return)
- Auto-start workflow on grant submission
- Action history
- Pending approvals listing

### 8. Finance & Disbursement ✅
- Only create disbursements against approved grants
- Track scheduled vs paid vs cancelled
- Currency support per disbursement
- Payment evidence URL
- Update status (scheduled → paid/cancelled)
- Grant status auto-updated to approved/rejected on workflow completion

### 9. Project Monitoring ✅
- Milestone and progress tracking per grant
- Create project updates (only for approved grants)
- Status tracking: planned → in_progress → completed
- Photo and invoice URL arrays per update
- Milestone date tracking

### 10. Dashboard & Reports (next)
- Tenant-scoped analytics
- Program-level stats
- Approval bottleneck analysis

## Database Schema

### tenants
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| name | VARCHAR(255) | Organization name |
| slug | VARCHAR(100) UNIQUE | URL-friendly identifier |
| type | VARCHAR(50) | faith_based, ngo, foundation, csr, government, development |
| settings | JSONB | Configurable tenant settings |
| status | VARCHAR(30) | active, suspended |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ | |

### tenant_users
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| user_id | FK → users | |
| roles | JSONB | ["admin", "reviewer", ...] |
| created_at | TIMESTAMPTZ | |

### grant_programs
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| name | VARCHAR(255) | |
| description | TEXT | |
| form_schema | JSONB | Dynamic application field definitions |
| required_document_types | JSONB | ["budget", "identification", ...] |
| applicant_types | JSONB | ["church", "school", ...] |
| workflow_template_id | FK → workflow_templates | |
| status | VARCHAR(30) | active, closed |
| budget_amount | DECIMAL(15,2) | |
| created_at, updated_at, deleted_at | | |

### applicants
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| type | VARCHAR(100) | Configurable per program |
| name | VARCHAR(255) | |
| email | VARCHAR(255) | |
| phone | VARCHAR(50) | |
| address | TEXT | |
| profile_data | JSONB | Dynamic profile fields |
| status | VARCHAR(30) | active, inactive |
| created_at, updated_at, deleted_at | | |

### grant_applications (refactored + extended)
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| program_id | FK → grant_programs | |
| applicant_id | FK → applicants | (was parish_id) |
| user_id | FK → users | Submitter |
| workflow_template_id | FK → workflow_templates | |
| current_step_id | FK → workflow_steps | |
| project_title | VARCHAR(500) | |
| project_description | TEXT | |
| requested_amount | DECIMAL(15,2) | |
| estimated_project_cost | DECIMAL(15,2) | |
| form_responses | JSONB | Dynamic answers to program form |
| status | VARCHAR(30) | draft, submitted, in_review, approved, rejected, closed |
| created_at, updated_at, deleted_at | | |

### workflow_templates
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| name | VARCHAR(255) | "Standard Approval", "Fast Track" |
| description | TEXT | |
| created_at, updated_at, deleted_at | | |

### workflow_steps
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| workflow_template_id | FK → workflow_templates | |
| name | VARCHAR(255) | "Initial Review", "Committee" |
| step_order | INTEGER | |
| assignee_roles | JSONB | ["reviewer", "committee"] |
| created_at, updated_at | | |

### application_workflows (runtime)
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| application_id | FK → grant_applications | |
| workflow_template_id | FK → workflow_templates | |
| current_step_id | FK → workflow_steps | |
| status | VARCHAR(30) | in_progress, completed, rejected |
| created_at, updated_at | | |

### workflow_actions (history)
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| application_workflow_id | FK → application_workflows | |
| step_id | FK → workflow_steps | |
| actor_id | FK → users | |
| action | VARCHAR(30) | approved, rejected, returned |
| comment | TEXT | |
| created_at | TIMESTAMPTZ | |

### documents (refactored)
Add `tenant_id` FK → tenants

### audit_logs (refactored)
Add `tenant_id` FK → tenants

### disbursements
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| grant_application_id | FK → grant_applications | |
| amount | DECIMAL(15,2) | |
| currency | VARCHAR(3) | USD, EUR, etc. |
| status | VARCHAR(30) | scheduled, paid, cancelled |
| payment_date | TIMESTAMPTZ | When payment made or scheduled |
| evidence_url | TEXT | Link to receipt/proof |
| notes | TEXT | Internal notes |
| paid_by | FK → users | Who processed payment |
| created_at, updated_at, deleted_at | | |

### project_updates
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | |
| tenant_id | FK → tenants | |
| grant_application_id | FK → grant_applications | |
| title | VARCHAR(500) | Milestone or update title |
| description | TEXT | Detailed progress description |
| status | VARCHAR(30) | planned, in_progress, completed, cancelled |
| milestone_date | TIMESTAMPTZ | Target or actual completion date |
| photos | JSONB | Array of photo URLs |
| invoices | JSONB | Array of invoice URLs |
| notes | TEXT | |
| created_by | FK → users | |
| created_at, updated_at, deleted_at | | |

## API Routes

### Public
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`

### Authenticated (scoped to user's tenants)
- `GET /api/v1/auth/me`
- `PUT /api/v1/auth/me`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

### Tenant-scoped (require active tenant context)
- `GET/POST /api/v1/tenants` — tenant management (super admin)
- `GET/POST /api/v1/tenants/:id/users` — tenant user management
- `GET/POST /api/v1/programs` — grant programs for current tenant
- `GET/PUT/DELETE /api/v1/programs/:id`
- `GET/POST /api/v1/applicants` — applicants for current tenant
- `GET/PUT/DELETE /api/v1/applicants/:id`
- `GET/POST /api/v1/workflows` — workflow templates
- `GET/PUT/DELETE /api/v1/workflows/:id`
- `GET/POST /api/v1/grants` — grant applications
- `GET/PUT/DELETE /api/v1/grants/:id`
- `POST /api/v1/grants/:id/submit`
- `GET/POST /api/v1/grants/:id/documents`
- `GET/PUT/DELETE /api/v1/grants/:id/documents/:docId`
- `GET /api/v1/grants/:id/documents/:docId/download`
- `GET /api/v1/grants/:id/workflow` — get runtime workflow state
- `POST /api/v1/grants/:id/workflow/approve` — approve current step
- `POST /api/v1/grants/:id/workflow/reject` — reject application
- `POST /api/v1/grants/:id/workflow/return` — return to applicant
- `GET /api/v1/grants/:id/workflow/history` — action history
- `GET /api/v1/workflow/pending` — pending approvals for current user
- `POST /api/v1/workflows/:id/steps` — manage workflow template steps
- `GET/POST /api/v1/grants/:id/disbursements` — per-grant disbursements
- `GET/PUT/DELETE /api/v1/disbursements/:id`
- `GET/POST /api/v1/grants/:id/updates` — per-grant project updates
- `GET/PUT/DELETE /api/v1/project-updates/:id`

### Admin
- `GET/POST /api/v1/admin/grants`
- `POST /api/v1/admin/documents/:id/verify`
- `POST /api/v1/admin/documents/:id/reject`
- `GET/POST/PUT/DELETE /api/v1/admin/tenants/:id`
- `GET /api/v1/admin/audit-logs`

## Engineering Requirements
- Clean Architecture: Handler → Service → Repository
- Database migrations with golang-migrate
- Request validation
- Structured error responses (apiErrors throughout)
- Pagination on all list endpoints
- Tenant-scoped query middleware
- RBAC per tenant
- Swagger documentation
- Unit and integration tests
- Docker Compose for API + PostgreSQL
