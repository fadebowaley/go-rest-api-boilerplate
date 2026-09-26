package permission

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
	ErrPermissionExists   = errors.New("permission already exists")
)

type Service interface {
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error)
	ListPermissions(ctx context.Context, group string, page, perPage int) (*PermissionListResponse, error)
	DeletePermission(ctx context.Context, id uint) error
	AssignPermissionsToRole(ctx context.Context, req *AssignPermissionsRequest) error
	GetPermissionsForRole(ctx context.Context, roleID uint) (*RolePermissionsResponse, error)
	GetUserPermissions(ctx context.Context, userID uint) ([]string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error) {
	existing, err := s.repo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}
	if existing != nil {
		return nil, ErrPermissionExists
	}

	p := &Permission{
		Name:        req.Name,
		Group:       req.Group,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	resp := ToPermissionResponse(p)
	return &resp, nil
}

func (s *service) ListPermissions(ctx context.Context, group string, page, perPage int) (*PermissionListResponse, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 {
		return nil, fmt.Errorf("perPage must be >= 1")
	}
	if perPage > 100 {
		return nil, fmt.Errorf("perPage must be <= 100")
	}

	permissions, total, err := s.repo.List(ctx, group, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	responses := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = ToPermissionResponse(&p)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &PermissionListResponse{
		Permissions: responses,
		Total:       total,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
	}, nil
}

func (s *service) DeletePermission(ctx context.Context, id uint) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find permission: %w", err)
	}
	if existing == nil {
		return ErrPermissionNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

func (s *service) AssignPermissionsToRole(ctx context.Context, req *AssignPermissionsRequest) error {
	if err := s.repo.AssignPermissionsToRole(ctx, req.RoleID, req.PermissionIDs); err != nil {
		return fmt.Errorf("failed to assign permissions: %w", err)
	}

	return nil
}

func (s *service) GetPermissionsForRole(ctx context.Context, roleID uint) (*RolePermissionsResponse, error) {
	permissions, err := s.repo.GetPermissionsForRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for role: %w", err)
	}

	responses := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		responses[i] = ToPermissionResponse(&p)
	}

	roleName, err := s.repo.GetRoleNameByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role name: %w", err)
	}

	return &RolePermissionsResponse{
		RoleID:      roleID,
		RoleName:    roleName,
		Permissions: responses,
	}, nil
}

func (s *service) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	perms, err := s.repo.GetPermissionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return perms, nil
}
