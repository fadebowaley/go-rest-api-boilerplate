package form

import (
	"encoding/json"
	"time"
)

type CreateFormTemplateRequest struct {
	Name             string          `json:"name" validate:"required,min=3,max=255"`
	Description      string          `json:"description"`
	SchemaDefinition json.RawMessage `json:"schema_definition"`
	Status           string          `json:"status"`
}

type UpdateFormTemplateRequest struct {
	Name             *string          `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description      *string          `json:"description,omitempty"`
	SchemaDefinition *json.RawMessage `json:"schema_definition,omitempty"`
	Status           *string          `json:"status,omitempty"`
}

type FormTemplateResponse struct {
	ID               uint            `json:"id"`
	TenantID         uint            `json:"tenant_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	SchemaDefinition json.RawMessage `json:"schema_definition"`
	Status           string          `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type FormVersionResponse struct {
	ID               uint            `json:"id"`
	FormTemplateID   uint            `json:"form_template_id"`
	VersionNumber    int             `json:"version_number"`
	SchemaDefinition json.RawMessage `json:"schema_definition"`
	Notes            string          `json:"notes,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

func ToFormTemplateResponse(t *FormTemplate) FormTemplateResponse {
	return FormTemplateResponse{
		ID:               t.ID,
		TenantID:         t.TenantID,
		Name:             t.Name,
		Description:      t.Description,
		SchemaDefinition: t.SchemaDefinition,
		Status:           t.Status,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

func ToFormVersionResponse(v *FormVersion) FormVersionResponse {
	return FormVersionResponse{
		ID:               v.ID,
		FormTemplateID:   v.FormTemplateID,
		VersionNumber:    v.VersionNumber,
		SchemaDefinition: v.SchemaDefinition,
		Notes:            v.Notes,
		CreatedAt:        v.CreatedAt,
	}
}
