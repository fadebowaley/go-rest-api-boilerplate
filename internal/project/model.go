package project

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	StatusPlanned    = "planned"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

var ValidStatuses = []string{StatusPlanned, StatusInProgress, StatusCompleted, StatusCancelled}

type ProjectUpdate struct {
	ID                 uint             `gorm:"primaryKey" json:"id"`
	TenantID           uint             `gorm:"not null;index" json:"tenant_id"`
	GrantApplicationID uint             `gorm:"not null;index" json:"grant_application_id"`
	Title              string           `gorm:"type:varchar(500);not null" json:"title"`
	Description        string           `gorm:"type:text" json:"description,omitempty"`
	Status             string           `gorm:"type:varchar(30);not null;default:'planned'" json:"status"`
	MilestoneDate      *time.Time       `json:"milestone_date,omitempty"`
	Photos             json.RawMessage  `gorm:"type:jsonb;default:'[]'" json:"photos,omitempty"`
	Invoices           json.RawMessage  `gorm:"type:jsonb;default:'[]'" json:"invoices,omitempty"`
	Notes              string           `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy          uint             `gorm:"not null" json:"created_by"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (ProjectUpdate) TableName() string {
	return "project_updates"
}
