package disbursement

import "time"

type CreateDisbursementRequest struct {
	GrantApplicationID uint    `json:"grant_application_id" binding:"required"`
	Amount             float64 `json:"amount" binding:"required,min=0.01"`
	Currency           string  `json:"currency"`
	PaymentDate        string  `json:"payment_date"`
	Notes              string  `json:"notes"`
}

type UpdateDisbursementRequest struct {
	Amount      *float64 `json:"amount" binding:"omitempty,min=0.01"`
	Currency    *string  `json:"currency"`
	Status      *string  `json:"status" binding:"omitempty,oneof=scheduled paid cancelled"`
	PaymentDate *string  `json:"payment_date"`
	EvidenceURL *string  `json:"evidence_url"`
	Notes       *string  `json:"notes"`
}

type DisbursementResponse struct {
	ID                 uint       `json:"id"`
	TenantID           uint       `json:"tenant_id"`
	GrantApplicationID uint       `json:"grant_application_id"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	PaymentDate        *time.Time `json:"payment_date,omitempty"`
	EvidenceURL        string     `json:"evidence_url,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	PaidBy             *uint      `json:"paid_by,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func ToDisbursementResponse(d *Disbursement) DisbursementResponse {
	return DisbursementResponse{
		ID:                 d.ID,
		TenantID:           d.TenantID,
		GrantApplicationID: d.GrantApplicationID,
		Amount:             d.Amount,
		Currency:           d.Currency,
		Status:             d.Status,
		PaymentDate:        d.PaymentDate,
		EvidenceURL:        d.EvidenceURL,
		Notes:              d.Notes,
		PaidBy:             d.PaidBy,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}
}
