package project

import (
	"encoding/json"
	"time"
)

type CreateUpdateRequest struct {
	GrantApplicationID uint             `json:"grant_application_id" binding:"required"`
	Title              string           `json:"title" binding:"required,min=2,max=500"`
	Description        string           `json:"description"`
	MilestoneDate      string           `json:"milestone_date"`
	Photos             json.RawMessage  `json:"photos"`
	Invoices           json.RawMessage  `json:"invoices"`
	Notes              string           `json:"notes"`
}

type UpdateUpdateRequest struct {
	Title         *string          `json:"title" binding:"omitempty,min=2,max=500"`
	Description   *string          `json:"description"`
	Status        *string          `json:"status" binding:"omitempty,oneof=planned in_progress completed cancelled"`
	MilestoneDate *string          `json:"milestone_date"`
	Photos        *json.RawMessage `json:"photos"`
	Invoices      *json.RawMessage `json:"invoices"`
	Notes         *string          `json:"notes"`
}

type UpdateResponse struct {
	ID                 uint              `json:"id"`
	TenantID           uint              `json:"tenant_id"`
	GrantApplicationID uint              `json:"grant_application_id"`
	Title              string            `json:"title"`
	Description        string            `json:"description,omitempty"`
	Status             string            `json:"status"`
	MilestoneDate      *time.Time        `json:"milestone_date,omitempty"`
	Photos             json.RawMessage   `json:"photos,omitempty"`
	Invoices           json.RawMessage   `json:"invoices,omitempty"`
	Notes              string            `json:"notes,omitempty"`
	CreatedBy          uint              `json:"created_by"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

func ToUpdateResponse(u *ProjectUpdate) UpdateResponse {
	return UpdateResponse{
		ID:                 u.ID,
		TenantID:           u.TenantID,
		GrantApplicationID: u.GrantApplicationID,
		Title:              u.Title,
		Description:        u.Description,
		Status:             u.Status,
		MilestoneDate:      u.MilestoneDate,
		Photos:             u.Photos,
		Invoices:           u.Invoices,
		Notes:              u.Notes,
		CreatedBy:          u.CreatedBy,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}
}
