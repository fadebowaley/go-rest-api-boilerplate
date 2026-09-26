package program

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type GrantProgram struct {
	ID                   uint            `gorm:"primaryKey" json:"id"`
	TenantID             uint            `gorm:"not null;index" json:"tenant_id"`
	Name                 string          `gorm:"type:varchar(255);not null" json:"name"`
	Description          string          `gorm:"type:text" json:"description"`
	FormSchema           json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"form_schema"`
	RequiredDocumentTypes json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"required_document_types"`
	ApplicantTypes       json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"applicant_types"`
	WorkflowTemplateID   *uint           `gorm:"" json:"workflow_template_id,omitempty"`
	Status               string          `gorm:"type:varchar(30);not null;default:active" json:"status"`
	BudgetAmount         float64         `gorm:"type:decimal(15,2);default:0" json:"budget_amount"`
	Currency             string          `gorm:"type:varchar(3);not null;default:USD" json:"currency"`
	EligibilityRules     json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"eligibility_rules"`
	ApplicationStartDate *time.Time      `gorm:"type:date" json:"application_start_date,omitempty"`
	ApplicationEndDate   *time.Time      `gorm:"type:date" json:"application_end_date,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (GrantProgram) TableName() string {
	return "grant_programs"
}
