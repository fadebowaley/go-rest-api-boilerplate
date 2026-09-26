package permission

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Group       string `json:"group" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

type PermissionResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
}

type PermissionListResponse struct {
	Permissions []PermissionResponse `json:"permissions"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
	TotalPages  int                  `json:"total_pages"`
}

type AssignPermissionsRequest struct {
	RoleID         uint   `json:"role_id"`
	PermissionIDs  []uint `json:"permission_ids" binding:"required,min=1"`
}

type RolePermissionsResponse struct {
	RoleID      uint                `json:"role_id"`
	RoleName    string              `json:"role_name"`
	Permissions []PermissionResponse `json:"permissions"`
}

func ToPermissionResponse(p *Permission) PermissionResponse {
	return PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Group:       p.Group,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
