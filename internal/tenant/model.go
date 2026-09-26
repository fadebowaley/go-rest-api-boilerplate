package tenant

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Tenant struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Type      string         `gorm:"type:varchar(50);not null;default:faith_based" json:"type"`
	Settings  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	Status    string         `gorm:"type:varchar(30);not null;default:active" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Tenant) TableName() string {
	return "tenants"
}

type TenantUser struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"not null;uniqueIndex:idx_tenant_user" json:"tenant_id"`
	UserID    uint           `gorm:"not null;uniqueIndex:idx_tenant_user" json:"user_id"`
	Roles     datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"roles"`
	CreatedAt time.Time      `json:"created_at"`
}

func (TenantUser) TableName() string {
	return "tenant_users"
}

type TenantRole struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	Permissions datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TenantRole) TableName() string {
	return "tenant_roles"
}

var ValidTenantTypes = []string{
	"faith_based", "ngo", "foundation", "csr", "government", "development",
}
