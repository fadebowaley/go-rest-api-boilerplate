package form

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type FormTemplate struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	TenantID         uint            `gorm:"not null;index" json:"tenant_id"`
	Name             string          `gorm:"type:varchar(255);not null" json:"name"`
	Description      string          `gorm:"type:text" json:"description"`
	SchemaDefinition json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"schema_definition"`
	Status           string          `gorm:"type:varchar(20);not null;default:draft;index" json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (FormTemplate) TableName() string {
	return "form_templates"
}

type FormVersion struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	FormTemplateID   uint            `gorm:"not null;index" json:"form_template_id"`
	VersionNumber    int             `gorm:"not null" json:"version_number"`
	SchemaDefinition json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"schema_definition"`
	Notes            string          `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

func (FormVersion) TableName() string {
	return "form_versions"
}
