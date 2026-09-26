package grant

import (
	"encoding/json"
	"time"
)

type CreateGrantRequest struct {
	ProgramID              *uint           `json:"program_id"`
	ApplicantID            *uint           `json:"applicant_id"`
	ParishID               *uint           `json:"parish_id"`
	ProjectTitle           string          `json:"project_title" binding:"required,min=3,max=500"`
	ProjectDescription     string          `json:"project_description"`
	BuildingStage          string          `json:"building_stage"`
	RequestedAmount        float64         `json:"requested_amount" binding:"required,min=0"`
	EstimatedProjectCost   float64         `json:"estimated_project_cost" binding:"required,min=0"`
	ProjectLocation        string          `json:"project_location"`
	ProjectStartDate       string          `json:"project_start_date"`
	ExpectedCompletionDate string          `json:"expected_completion_date"`
	Justification          string          `json:"justification"`
	FormResponses          json.RawMessage `json:"form_responses,omitempty"`
}

type UpdateGrantRequest struct {
	ProjectTitle           string  `json:"project_title" binding:"omitempty,min=3,max=500"`
	ProjectDescription     string  `json:"project_description"`
	BuildingStage          string  `json:"building_stage"`
	RequestedAmount        *float64 `json:"requested_amount" binding:"omitempty,min=0"`
	EstimatedProjectCost   *float64 `json:"estimated_project_cost" binding:"omitempty,min=0"`
	ProjectLocation        string  `json:"project_location"`
	ProjectStartDate       string  `json:"project_start_date"`
	ExpectedCompletionDate string  `json:"expected_completion_date"`
	Justification          string  `json:"justification"`
}

type GrantFilterParams struct {
	Status   string
	ParishID uint
	Region   string
	Province string
	DateFrom string
	DateTo   string
	Search   string
	Sort     string
	Order    string
}

type GrantResponse struct {
	ID                     uint             `json:"id"`
	TenantID               uint             `json:"tenant_id"`
	ProgramID              *uint            `json:"program_id,omitempty"`
	ApplicantID            *uint            `json:"applicant_id,omitempty"`
	ParishID               *uint            `json:"parish_id,omitempty"`
	UserID                 uint             `json:"user_id"`
	WorkflowTemplateID     *uint            `json:"workflow_template_id,omitempty"`
	CurrentStepID          *uint            `json:"current_step_id,omitempty"`
	ProjectTitle           string           `json:"project_title"`
	ProjectDescription     string           `json:"project_description"`
	BuildingStage          string           `json:"building_stage"`
	RequestedAmount        float64          `json:"requested_amount"`
	EstimatedProjectCost   float64          `json:"estimated_project_cost"`
	ProjectLocation        string           `json:"project_location"`
	ProjectStartDate       string           `json:"project_start_date"`
	ExpectedCompletionDate string           `json:"expected_completion_date"`
	Justification          string           `json:"justification"`
	FormResponses          json.RawMessage  `json:"form_responses"`
	FormTemplateID         *uint            `json:"form_template_id,omitempty"`
	FormVersionID          *uint            `json:"form_version_id,omitempty"`
	Status                 string           `json:"status"`
	CreatedAt              string           `json:"created_at"`
	UpdatedAt              string           `json:"updated_at"`
}

type GrantListResponse struct {
	Grants     []GrantResponse `json:"grants"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

func ToGrantResponse(g *GrantApplication) GrantResponse {
	r := GrantResponse{
		ID:                 g.ID,
		TenantID:           g.TenantID,
		ProgramID:          g.ProgramID,
		ApplicantID:        g.ApplicantID,
		ParishID:           g.ParishID,
		UserID:             g.UserID,
		WorkflowTemplateID: g.WorkflowTemplateID,
		CurrentStepID:      g.CurrentStepID,
		ProjectTitle:       g.ProjectTitle,
		ProjectDescription: g.ProjectDescription,
		BuildingStage:      g.BuildingStage,
		RequestedAmount:    g.RequestedAmount,
		EstimatedProjectCost: g.EstimatedProjectCost,
		ProjectLocation:    g.ProjectLocation,
		Justification:      g.Justification,
		FormResponses:      g.FormResponses,
		FormTemplateID:     g.FormTemplateID,
		FormVersionID:      g.FormVersionID,
		Status:             g.Status,
		CreatedAt:          g.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          g.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if g.ProjectStartDate != nil {
		r.ProjectStartDate = g.ProjectStartDate.Format("2006-01-02")
	}
	if g.ExpectedCompletionDate != nil {
		r.ExpectedCompletionDate = g.ExpectedCompletionDate.Format("2006-01-02")
	}
	return r
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}
