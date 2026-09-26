package program

import (
	"encoding/json"
	"time"
)

type CreateProgramRequest struct {
	Name                 string          `json:"name" validate:"required,min=2,max=255"`
	Description          string          `json:"description"`
	FormSchema           json.RawMessage `json:"form_schema"`
	RequiredDocumentTypes json.RawMessage `json:"required_document_types"`
	ApplicantTypes       json.RawMessage `json:"applicant_types"`
	WorkflowTemplateID   *uint           `json:"workflow_template_id"`
	BudgetAmount         *float64        `json:"budget_amount"`
	Currency             *string         `json:"currency"`
	EligibilityRules     json.RawMessage `json:"eligibility_rules"`
	ApplicationStartDate *string         `json:"application_start_date"`
	ApplicationEndDate   *string         `json:"application_end_date"`
}

type UpdateProgramRequest struct {
	Name                 *string          `json:"name" validate:"omitempty,min=2,max=255"`
	Description          *string          `json:"description"`
	FormSchema           *json.RawMessage `json:"form_schema"`
	RequiredDocumentTypes *json.RawMessage `json:"required_document_types"`
	ApplicantTypes       *json.RawMessage `json:"applicant_types"`
	WorkflowTemplateID   *uint            `json:"workflow_template_id"`
	Status               *string          `json:"status"`
	BudgetAmount         *float64         `json:"budget_amount"`
	Currency             *string          `json:"currency"`
	EligibilityRules     *json.RawMessage `json:"eligibility_rules"`
	ApplicationStartDate *string          `json:"application_start_date"`
	ApplicationEndDate   *string          `json:"application_end_date"`
}

type ProgramResponse struct {
	ID                   uint            `json:"id"`
	TenantID             uint            `json:"tenant_id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	FormSchema           json.RawMessage `json:"form_schema"`
	RequiredDocumentTypes json.RawMessage `json:"required_document_types"`
	ApplicantTypes       json.RawMessage `json:"applicant_types"`
	WorkflowTemplateID   *uint           `json:"workflow_template_id,omitempty"`
	Status               string          `json:"status"`
	BudgetAmount         float64         `json:"budget_amount"`
	Currency             string          `json:"currency"`
	EligibilityRules     json.RawMessage `json:"eligibility_rules"`
	ApplicationStartDate *time.Time      `json:"application_start_date,omitempty"`
	ApplicationEndDate   *time.Time      `json:"application_end_date,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func ToProgramResponse(p *GrantProgram) ProgramResponse {
	return ProgramResponse{
		ID:                   p.ID,
		TenantID:             p.TenantID,
		Name:                 p.Name,
		Description:          p.Description,
		FormSchema:           p.FormSchema,
		RequiredDocumentTypes: p.RequiredDocumentTypes,
		ApplicantTypes:       p.ApplicantTypes,
		WorkflowTemplateID:   p.WorkflowTemplateID,
		Status:               p.Status,
		BudgetAmount:         p.BudgetAmount,
		Currency:             p.Currency,
		EligibilityRules:     p.EligibilityRules,
		ApplicationStartDate: p.ApplicationStartDate,
		ApplicationEndDate:   p.ApplicationEndDate,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}
