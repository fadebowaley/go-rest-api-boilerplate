package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/datatypes"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrTenantNotFound       = errors.New("tenant not found")
	ErrTenantSlugExists     = errors.New("tenant slug already exists")
	ErrUserNotInTenant      = errors.New("user is not a member of this tenant")
	ErrInvalidTenantType    = errors.New("invalid tenant type")
	ErrNotTenantAdmin       = errors.New("user is not a tenant admin")
	ErrUserAlreadyInTenant  = errors.New("user is already a member of this tenant")
	ErrUserNotFoundByEmail  = errors.New("user with this email not found")
)

type Service interface {
	Create(ctx context.Context, req *CreateTenantRequest) (*TenantResponse, error)
	GetByID(ctx context.Context, id uint) (*TenantResponse, error)
	List(ctx context.Context) (*[]TenantResponse, error)
	Update(ctx context.Context, id uint, req *UpdateTenantRequest) (*TenantResponse, error)
	Delete(ctx context.Context, id uint) error

	AddUser(ctx context.Context, tenantID uint, req *AddTenantUserRequest) (*TenantUserResponse, error)
	RemoveUser(ctx context.Context, tenantID, userID uint) error
	GetUsers(ctx context.Context, tenantID uint) (*[]TenantUserResponse, error)
	GetUserTenants(ctx context.Context, userID uint) (*[]TenantUserResponse, error)
	GetUserTenant(ctx context.Context, userID uint, tenantID uint) (*TenantUserResponse, error)
	InviteUser(ctx context.Context, inviterID uint, tenantID uint, req *InviteUserRequest) (*TenantUserResponse, error)

	CreateRole(ctx context.Context, userID uint, tenantID uint, req *CreateTenantRoleRequest) (*TenantRoleResponse, error)
	GetRole(ctx context.Context, userID uint, roleID uint) (*TenantRoleResponse, error)
	ListRoles(ctx context.Context, userID uint, tenantID uint) (*[]TenantRoleResponse, error)
	UpdateRole(ctx context.Context, userID uint, roleID uint, req *UpdateTenantRoleRequest) (*TenantRoleResponse, error)
	DeleteRole(ctx context.Context, userID uint, roleID uint) error
}

// UserFinder allows the tenant service to look up users by email
type UserFinder interface {
	FindByEmail(ctx context.Context, email string) (uint, string, error)
}

type service struct {
	repo  Repository
	users UserFinder
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func NewServiceWithUsers(repo Repository, users UserFinder) Service {
	return &service{repo: repo, users: users}
}

func (s *service) Create(ctx context.Context, req *CreateTenantRequest) (*TenantResponse, error) {
	if !isValidType(req.Type) {
		return nil, ErrInvalidTenantType
	}

	req.Slug = strings.ToLower(strings.ReplaceAll(req.Slug, " ", "-"))

	tenant := &Tenant{
		Name:     req.Name,
		Slug:     req.Slug,
		Type:     req.Type,
		Settings: datatypes.JSON(req.Settings),
		Status:   "active",
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		if isDuplicateError(err) {
			return nil, ErrTenantSlugExists
		}
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	resp := ToTenantResponse(tenant)
	tenantCtx := contextutil.GetTenantContext(ctx)
	if tenantCtx != nil {
		tenantCtx.SetTenantID(tenant.ID)
	}

	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*TenantResponse, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	resp := ToTenantResponse(tenant)
	return &resp, nil
}

func (s *service) List(ctx context.Context) (*[]TenantResponse, error) {
	tenants, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]TenantResponse, len(tenants))
	for i, t := range tenants {
		responses[i] = ToTenantResponse(&t)
	}
	return &responses, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateTenantRequest) (*TenantResponse, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.Slug != nil {
		tenant.Slug = strings.ToLower(strings.ReplaceAll(*req.Slug, " ", "-"))
	}
	if req.Type != nil {
		if !isValidType(*req.Type) {
			return nil, ErrInvalidTenantType
		}
		tenant.Type = *req.Type
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Settings != nil {
		tenant.Settings = datatypes.JSON(*req.Settings)
	}

	if err := s.repo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	resp := ToTenantResponse(tenant)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrTenantNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) AddUser(ctx context.Context, tenantID uint, req *AddTenantUserRequest) (*TenantUserResponse, error) {
	if _, err := s.repo.GetByID(ctx, tenantID); err != nil {
		return nil, ErrTenantNotFound
	}

	rolesJSON, err := marshalRoles(req.Roles)
	if err != nil {
		return nil, fmt.Errorf("invalid roles: %w", err)
	}

	tu := &TenantUser{
		TenantID: tenantID,
		UserID:   req.UserID,
		Roles:    rolesJSON,
	}

	if err := s.repo.AddUser(ctx, tu); err != nil {
		return nil, fmt.Errorf("failed to add user to tenant: %w", err)
	}

	resp, err := ToTenantUserResponse(tu)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *service) RemoveUser(ctx context.Context, tenantID, userID uint) error {
	return s.repo.RemoveUser(ctx, tenantID, userID)
}

func (s *service) GetUsers(ctx context.Context, tenantID uint) (*[]TenantUserResponse, error) {
	users, err := s.repo.GetUsers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	responses := make([]TenantUserResponse, len(users))
	for i, u := range users {
		resp, err := ToTenantUserResponse(&u)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}
	return &responses, nil
}

func (s *service) InviteUser(ctx context.Context, inviterID uint, tenantID uint, req *InviteUserRequest) (*TenantUserResponse, error) {
	// Verify inviter is a tenant_admin
	inviterTU, err := s.repo.GetUserTenant(ctx, inviterID, tenantID)
	if err != nil {
		return nil, ErrNotTenantAdmin
	}
	var inviterRoles []string
	if err := json.Unmarshal(inviterTU.Roles, &inviterRoles); err != nil {
		return nil, fmt.Errorf("failed to parse inviter roles: %w", err)
	}
	if !contains(inviterRoles, "tenant_admin") {
		return nil, ErrNotTenantAdmin
	}

	// Check tenant exists
	if _, err := s.repo.GetByID(ctx, tenantID); err != nil {
		return nil, ErrTenantNotFound
	}

	// Find invited user
	if s.users == nil {
		return nil, errors.New("user lookup not available")
	}
	invitedUserID, _, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrUserNotFoundByEmail
	}

	// Check not already a member
	existing, _ := s.repo.GetUserTenant(ctx, invitedUserID, tenantID)
	if existing != nil {
		return nil, ErrUserAlreadyInTenant
	}

	rolesJSON, err := marshalRoles(req.Roles)
	if err != nil {
		return nil, fmt.Errorf("invalid roles: %w", err)
	}

	tu := &TenantUser{
		TenantID: tenantID,
		UserID:   invitedUserID,
		Roles:    rolesJSON,
	}
	if err := s.repo.AddUser(ctx, tu); err != nil {
		return nil, fmt.Errorf("failed to add user to tenant: %w", err)
	}

	resp, err := ToTenantUserResponse(tu)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *service) requireTenantAdmin(ctx context.Context, userID, tenantID uint) error {
	tu, err := s.repo.GetUserTenant(ctx, userID, tenantID)
	if err != nil {
		return ErrNotTenantAdmin
	}
	var roles []string
	if err := json.Unmarshal(tu.Roles, &roles); err != nil {
		return fmt.Errorf("failed to parse user roles: %w", err)
	}
	if !contains(roles, "tenant_admin") {
		return ErrNotTenantAdmin
	}
	return nil
}

func (s *service) CreateRole(ctx context.Context, userID uint, tenantID uint, req *CreateTenantRoleRequest) (*TenantRoleResponse, error) {
	if err := s.requireTenantAdmin(ctx, userID, tenantID); err != nil {
		return nil, err
	}

	permsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("invalid permissions: %w", err)
	}

	role := &TenantRole{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: permsJSON,
	}

	if err := s.repo.CreateTenantRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create tenant role: %w", err)
	}

	resp := ToTenantRoleResponse(role)
	return &resp, nil
}

func (s *service) GetRole(ctx context.Context, userID uint, roleID uint) (*TenantRoleResponse, error) {
	role, err := s.repo.GetTenantRoleByID(ctx, roleID)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	if err := s.requireTenantAdmin(ctx, userID, role.TenantID); err != nil {
		return nil, err
	}

	resp := ToTenantRoleResponse(role)
	return &resp, nil
}

func (s *service) ListRoles(ctx context.Context, userID uint, tenantID uint) (*[]TenantRoleResponse, error) {
	if err := s.requireTenantAdmin(ctx, userID, tenantID); err != nil {
		return nil, err
	}

	roles, err := s.repo.ListTenantRoles(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant roles: %w", err)
	}

	responses := make([]TenantRoleResponse, len(roles))
	for i, r := range roles {
		responses[i] = ToTenantRoleResponse(&r)
	}
	return &responses, nil
}

func (s *service) UpdateRole(ctx context.Context, userID uint, roleID uint, req *UpdateTenantRoleRequest) (*TenantRoleResponse, error) {
	role, err := s.repo.GetTenantRoleByID(ctx, roleID)
	if err != nil {
		return nil, ErrTenantNotFound
	}

	if err := s.requireTenantAdmin(ctx, userID, role.TenantID); err != nil {
		return nil, err
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Permissions != nil {
		permsJSON, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("invalid permissions: %w", err)
		}
		role.Permissions = permsJSON
	}

	if err := s.repo.UpdateTenantRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update tenant role: %w", err)
	}

	resp := ToTenantRoleResponse(role)
	return &resp, nil
}

func (s *service) DeleteRole(ctx context.Context, userID uint, roleID uint) error {
	role, err := s.repo.GetTenantRoleByID(ctx, roleID)
	if err != nil {
		return ErrTenantNotFound
	}

	if err := s.requireTenantAdmin(ctx, userID, role.TenantID); err != nil {
		return err
	}

	return s.repo.DeleteTenantRole(ctx, roleID)
}

func (s *service) GetUserTenants(ctx context.Context, userID uint) (*[]TenantUserResponse, error) {
	users, err := s.repo.GetUserTenants(ctx, userID)
	if err != nil {
		return nil, err
	}
	responses := make([]TenantUserResponse, len(users))
	for i, u := range users {
		resp, err := ToTenantUserResponse(&u)
		if err != nil {
			return nil, err
		}
		responses[i] = resp
	}
	return &responses, nil
}

func (s *service) GetUserTenant(ctx context.Context, userID uint, tenantID uint) (*TenantUserResponse, error) {
	tu, err := s.repo.GetUserTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, ErrUserNotInTenant
	}
	resp, err := ToTenantUserResponse(tu)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func isValidType(t string) bool {
	for _, vt := range ValidTenantTypes {
		if t == vt {
			return true
		}
	}
	return false
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key")
}

func marshalRoles(roles []string) ([]byte, error) {
	return json.Marshal(roles)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
