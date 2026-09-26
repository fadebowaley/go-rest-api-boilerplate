You are a senior React frontend engineer.

We have a borrowed React frontend boilerplate from another application. Your task is to decouple the existing authentication system and rewire it cleanly for Applico, our multi-tenant grant management platform.

Do not build grant modules, dashboards, workflows, forms, or extra pages yet. Focus only on authentication foundation.

Backend auth APIs available:
- POST /api/v1/auth/login
- POST /api/v1/auth/register
- POST /api/v1/auth/refresh
- POST /api/v1/auth/logout
- GET /api/v1/auth/me
- GET /api/v1/auth/me/tenants

Primary goal:
A user should be able to log in, store session safely, fetch profile, fetch tenant memberships, persist session across refresh, logout, and be redirected correctly.

Tasks:

1. Remove old app auth coupling
- Identify all old auth references from the borrowed app
- Remove old API URLs
- Remove old token names
- Remove old user/session assumptions
- Remove old role checks
- Remove old hardcoded dashboard redirects
- Remove unused auth helper files

2. Create clean auth API layer
Create:

src/lib/api/client.ts
src/lib/api/auth.ts

The API client must:
- use environment variable for API base URL
- attach access token to requests
- handle JSON responses
- handle errors consistently
- support refresh token flow
- logout cleanly on refresh failure

3. Create Applico auth types
Create:

src/types/auth.ts

Types should include:
- User
- LoginRequest
- LoginResponse
- RegisterRequest
- AuthSession
- TenantMembership
- Role
- Permission

Do not assume old app user fields. Match backend response shape. If unclear, create adaptable types and document expected API response.

4. Create auth storage utility
Create:

src/lib/auth/storage.ts

Handle:
- save access token
- save refresh token
- get access token
- get refresh token
- clear session
- save active tenant id
- get active tenant id
- clear active tenant id

Use consistent key names:
- applico_access_token
- applico_refresh_token
- applico_active_tenant_id

5. Create auth context/provider
Create:

src/features/auth/AuthProvider.tsx

It must expose:
- user
- tenants
- activeTenant
- isAuthenticated
- isLoading
- login()
- logout()
- refreshSession()
- loadMe()
- setActiveTenant()

On app start:
- check existing token
- call /auth/me
- call /auth/me/tenants
- restore active tenant if valid
- auto-select tenant if user has only one tenant
- allow platform admin to continue without tenant

6. Fix login page
Wire login page to:
POST /api/v1/auth/login

On successful login:
- store tokens
- fetch /auth/me
- fetch /auth/me/tenants
- redirect based on user type:
  - platform_admin → /admin
  - tenant user with one tenant → /dashboard
  - tenant user with multiple tenants → /select-tenant
  - user with no tenant and not platform admin → /no-access

7. Add tenant selection page
Create:

/select-tenant

It should:
- list tenant memberships
- allow user to choose tenant
- save active tenant id
- redirect to /dashboard

8. Add protected route handling
Create route guards for:
- authenticated routes
- guest-only routes
- platform admin routes
- tenant-required routes

Rules:
- unauthenticated users go to /login
- authenticated users should not stay on /login
- platform_admin can access /admin without active tenant
- normal users need active tenant for /dashboard
- no tenant users go to /no-access

9. Fix logout
Logout should:
- call backend logout if endpoint is available
- clear local tokens
- clear active tenant
- clear auth state
- redirect to /login

10. Add auth debug screen temporarily
Create a temporary route:

/auth-debug

Show:
- authenticated status
- user object
- tenants
- active tenant
- access token exists yes/no
- refresh token exists yes/no

This is for development only.

11. Environment setup
Add:

VITE_API_BASE_URL=http://localhost:8080/api/v1

or the correct React env variable depending on the boilerplate.

12. Acceptance criteria
Auth is complete only when:

- Old borrowed app auth logic is removed
- Login works with Applico backend
- Token is stored with Applico key names
- Refreshing browser keeps user logged in
- /auth/me loads correctly
- /auth/me/tenants loads correctly
- Single-tenant user goes to dashboard
- Multi-tenant user goes to tenant selector
- Platform admin goes to admin dashboard
- Logout clears everything
- Expired token refresh works
- Failed refresh logs user out
- Protected routes block unauthenticated access

Do not build other Applico features yet.
