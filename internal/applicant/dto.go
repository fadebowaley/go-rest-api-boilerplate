package applicant

import (
	"encoding/json"
	"time"
)

type CreateApplicantRequest struct {
	Type        string          `json:"type" validate:"required"`
	Name        string          `json:"name" validate:"required,min=2,max=255"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Address     string          `json:"address"`
	ProfileData json.RawMessage `json:"profile_data"`
}

type UpdateApplicantRequest struct {
	Type        *string          `json:"type"`
	Name        *string          `json:"name" validate:"omitempty,min=2,max=255"`
	Email       *string          `json:"email"`
	Phone       *string          `json:"phone"`
	Address     *string          `json:"address"`
	Status      *string          `json:"status"`
	ProfileData *json.RawMessage `json:"profile_data"`
}

type ApplicantResponse struct {
	ID          uint            `json:"id"`
	TenantID    uint            `json:"tenant_id"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Email       string          `json:"email,omitempty"`
	Phone       string          `json:"phone,omitempty"`
	Address     string          `json:"address,omitempty"`
	ProfileData json.RawMessage `json:"profile_data"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func ToApplicantResponse(a *Applicant) ApplicantResponse {
	return ApplicantResponse{
		ID:          a.ID,
		TenantID:    a.TenantID,
		Type:        a.Type,
		Name:        a.Name,
		Email:       a.Email,
		Phone:       a.Phone,
		Address:     a.Address,
		ProfileData: a.ProfileData,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
