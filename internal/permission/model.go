package permission

import "time"

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Group       string    `gorm:"not null;index" json:"group"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey"`
	PermissionID uint      `gorm:"primaryKey"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
