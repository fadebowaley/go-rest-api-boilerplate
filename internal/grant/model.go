package grant

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	StatusDraft               = "draft"
	StatusSubmitted           = "submitted"
	StatusUnderReview         = "under_review"
	StatusPendingClarification = "pending_clarification"
	StatusApproved            = "approved"
	StatusRejected            = "rejected"
	StatusDisbursed           = "disbursed"
	StatusClosed              = "closed"
)

var ValidStatuses = []string{
	StatusDraft, StatusSubmitted, StatusUnderReview,
	StatusPendingClarification, StatusApproved, StatusRejected,
	StatusDisbursed, StatusClosed,
}

var ValidTransitions = map[string][]string{
	StatusDraft:               {StatusSubmitted},
	StatusSubmitted:           {StatusUnderReview, StatusRejected},
	StatusUnderReview:         {StatusApproved, StatusRejected, StatusPendingClarification},
	StatusPendingClarification: {StatusSubmitted, StatusRejected},
	StatusApproved:            {StatusDisbursed},
	StatusDisbursed:           {StatusClosed},
	StatusRejected:            {},
	StatusClosed:              {},
}

type GrantApplication struct {
	ID                     uint             `gorm:"primaryKey" json:"id"`
	TenantID               uint             `gorm:"not null;index" json:"tenant_id"`
	ProgramID              *uint            `gorm:"index" json:"program_id,omitempty"`
	ApplicantID            *uint            `gorm:"index" json:"applicant_id,omitempty"`
	ParishID               *uint            `gorm:"index" json:"parish_id,omitempty"`
	UserID                 uint             `gorm:"not null;index" json:"user_id"`
	WorkflowTemplateID     *uint            `json:"workflow_template_id,omitempty"`
	CurrentStepID          *uint            `json:"current_step_id,omitempty"`
	ProjectTitle           string           `gorm:"not null" json:"project_title"`
	ProjectDescription     string           `gorm:"type:text" json:"project_description"`
	BuildingStage          string           `gorm:"type:varchar(50)" json:"building_stage,omitempty"`
	RequestedAmount        float64          `gorm:"type:decimal(15,2);not null;default:0" json:"requested_amount"`
	EstimatedProjectCost   float64          `gorm:"type:decimal(15,2);not null;default:0" json:"estimated_project_cost"`
	ProjectLocation        string           `gorm:"type:text" json:"project_location"`
	ProjectStartDate       *time.Time       `json:"project_start_date"`
	ExpectedCompletionDate *time.Time       `json:"expected_completion_date"`
	Justification          string           `gorm:"type:text" json:"justification"`
	FormResponses          json.RawMessage  `gorm:"type:jsonb;default:'{}'" json:"form_responses"`
	FormTemplateID         *uint            `json:"form_template_id,omitempty"`
	FormVersionID          *uint            `json:"form_version_id,omitempty"`
	Status                 string           `gorm:"type:varchar(30);not null;default:draft;index" json:"status"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	DeletedAt              gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (GrantApplication) TableName() string {
	return "grant_applications"
}

func IsValidStatus(s string) bool {
	for _, vs := range ValidStatuses {
		if vs == s {
			return true
		}
	}
	return false
}

func CanTransition(from, to string) bool {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == to {
			return true
		}
	}
	return false
}
