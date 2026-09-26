package workflow

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type WorkflowTemplate struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"not null;index" json:"tenant_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkflowTemplate) TableName() string {
	return "workflow_templates"
}

type WorkflowStep struct {
	ID                           uint            `gorm:"primaryKey" json:"id"`
	WorkflowTemplateID           uint            `gorm:"not null;index" json:"workflow_template_id"`
	Name                         string          `gorm:"type:varchar(255);not null" json:"name"`
	StepOrder                    int             `gorm:"not null;default:0" json:"step_order"`
	AssigneeRoles                json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"assignee_roles"`
	AssigneeUserID               *uint           `gorm:"index" json:"assignee_user_id,omitempty"`
	SLAHours                     int             `gorm:"not null;default:0" json:"sla_hours"`
	RequireDocumentVerification  bool            `gorm:"not null;default:false" json:"require_document_verification"`
	IsFinalApproval              bool            `gorm:"not null;default:false" json:"is_final_approval"`
	CreatedAt                    time.Time       `json:"created_at"`
	UpdatedAt                    time.Time       `json:"updated_at"`
}

func (WorkflowStep) TableName() string {
	return "workflow_steps"
}

const (
	WorkflowStatusInProgress = "in_progress"
	WorkflowStatusCompleted  = "completed"
	WorkflowStatusRejected   = "rejected"

	ActionApprove = "approved"
	ActionReject  = "rejected"
	ActionReturn  = "returned"
)

type ApplicationWorkflow struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	TenantID           uint       `gorm:"not null;index" json:"tenant_id"`
	ApplicationID      uint       `gorm:"not null;index" json:"application_id"`
	WorkflowTemplateID uint       `gorm:"not null" json:"workflow_template_id"`
	CurrentStepID      *uint      `gorm:"index" json:"current_step_id"`
	Status             string     `gorm:"type:varchar(30);not null;default:in_progress" json:"status"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (ApplicationWorkflow) TableName() string {
	return "application_workflows"
}

type WorkflowAction struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	TenantID              uint      `gorm:"not null;index" json:"tenant_id"`
	ApplicationWorkflowID uint      `gorm:"not null;index" json:"application_workflow_id"`
	StepID                uint      `gorm:"not null;index" json:"step_id"`
	ActorID               uint      `gorm:"not null;index" json:"actor_id"`
	Action                string    `gorm:"type:varchar(30);not null" json:"action"`
	Comment               string    `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

func (WorkflowAction) TableName() string {
	return "workflow_actions"
}
