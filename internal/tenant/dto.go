package tenant

import (
	"encoding/json"
	"time"
)

type CreateTenantRequest struct {
	Name     string          `json:"name" validate:"required,min=2,max=255"`
	Slug     string          `json:"slug" validate:"required,min=2,max=100"`
	Type     string          `json:"type" validate:"required"`
	Settings json.RawMessage `json:"settings"`
}

type UpdateTenantRequest struct {
	Name     *string          `json:"name" validate:"omitempty,min=2,max=255"`
	Slug     *string          `json:"slug" validate:"omitempty,min=2,max=100"`
	Type     *string          `json:"type"`
	Status   *string          `json:"status"`
	Settings *json.RawMessage `json:"settings"`
}

type AddTenantUserRequest struct {
	UserID uint     `json:"user_id" validate:"required"`
	Roles  []string `json:"roles" validate:"required"`
}

type InviteUserRequest struct {
	Email string   `json:"email" validate:"required,email"`
	Roles []string `json:"roles" validate:"required"`
}

type CreateTenantRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UpdateTenantRoleRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

type TenantRoleResponse struct {
	ID          uint     `json:"id"`
	TenantID    uint     `json:"tenant_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func ToTenantRoleResponse(r *TenantRole) TenantRoleResponse {
	var perms []string
	if r.Permissions != nil {
		if err := json.Unmarshal(r.Permissions, &perms); err != nil {
			perms = []string{}
		}
	}
	return TenantRoleResponse{
		ID:          r.ID,
		TenantID:    r.TenantID,
		Name:        r.Name,
		Description: r.Description,
		Permissions: perms,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type TenantResponse struct {
	ID        uint            `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Type      string          `json:"type"`
	Settings  json.RawMessage `json:"settings"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type TenantUserResponse struct {
	ID        uint      `json:"id"`
	TenantID  uint      `json:"tenant_id"`
	UserID    uint      `json:"user_id"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}

func ToTenantResponse(t *Tenant) TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		Type:      t.Type,
		Settings:  json.RawMessage(t.Settings),
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func ToTenantUserResponse(tu *TenantUser) (TenantUserResponse, error) {
	var roles []string
	if err := json.Unmarshal(tu.Roles, &roles); err != nil {
		roles = []string{}
	}
	return TenantUserResponse{
		ID:        tu.ID,
		TenantID:  tu.TenantID,
		UserID:    tu.UserID,
		Roles:     roles,
		CreatedAt: tu.CreatedAt,
	}, nil
}
