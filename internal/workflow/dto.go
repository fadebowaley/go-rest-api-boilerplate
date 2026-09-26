package workflow

import (
	"encoding/json"
	"time"
)

type CreateTemplateRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=255"`
	Description string `json:"description"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=255"`
	Description *string `json:"description"`
}

type CreateStepRequest struct {
	Name          string          `json:"name" validate:"required,min=2,max=255"`
	StepOrder     int             `json:"step_order"`
	AssigneeRoles json.RawMessage `json:"assignee_roles"`
}

type UpdateStepRequest struct {
	Name          *string          `json:"name" validate:"omitempty,min=2,max=255"`
	StepOrder     *int             `json:"step_order"`
	AssigneeRoles *json.RawMessage `json:"assignee_roles"`
}

type TemplateResponse struct {
	ID          uint      `json:"id"`
	TenantID    uint      `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StepResponse struct {
	ID                 uint            `json:"id"`
	WorkflowTemplateID uint            `json:"workflow_template_id"`
	Name               string          `json:"name"`
	StepOrder          int             `json:"step_order"`
	AssigneeRoles      json.RawMessage `json:"assignee_roles"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

func ToTemplateResponse(t *WorkflowTemplate) TemplateResponse {
	return TemplateResponse{
		ID:          t.ID,
		TenantID:    t.TenantID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToStepResponse(s *WorkflowStep) StepResponse {
	return StepResponse{
		ID:                 s.ID,
		WorkflowTemplateID: s.WorkflowTemplateID,
		Name:               s.Name,
		StepOrder:          s.StepOrder,
		AssigneeRoles:      s.AssigneeRoles,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

type WorkflowActionRequest struct {
	Comment string `json:"comment"`
}

type ApplicationWorkflowResponse struct {
	ID                 uint                `json:"id"`
	ApplicationID      uint                `json:"application_id"`
	WorkflowTemplateID uint                `json:"workflow_template_id"`
	CurrentStepID      *uint               `json:"current_step_id"`
	Status             string              `json:"status"`
	StartedAt          time.Time           `json:"started_at"`
	CompletedAt        *time.Time          `json:"completed_at,omitempty"`
	Actions            []ActionResponse    `json:"actions,omitempty"`
}

type ActionResponse struct {
	ID        uint      `json:"id"`
	StepID    uint      `json:"step_id"`
	StepName  string    `json:"step_name"`
	ActorID   uint      `json:"actor_id"`
	Action    string    `json:"action"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type PendingApprovalResponse struct {
	ApplicationID      uint   `json:"application_id"`
	ProjectTitle       string `json:"project_title"`
	WorkflowID         uint   `json:"workflow_id"`
	CurrentStepID      uint   `json:"current_step_id"`
	CurrentStepName    string `json:"current_step_name"`
}

func ToAppWorkflowResponse(w *ApplicationWorkflow) ApplicationWorkflowResponse {
	return ApplicationWorkflowResponse{
		ID:                 w.ID,
		ApplicationID:      w.ApplicationID,
		WorkflowTemplateID: w.WorkflowTemplateID,
		CurrentStepID:      w.CurrentStepID,
		Status:             w.Status,
		StartedAt:          w.StartedAt,
		CompletedAt:        w.CompletedAt,
	}
}
