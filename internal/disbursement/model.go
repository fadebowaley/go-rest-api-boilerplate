package disbursement

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusScheduled = "scheduled"
	StatusPaid      = "paid"
	StatusCancelled = "cancelled"
)

var ValidStatuses = []string{StatusScheduled, StatusPaid, StatusCancelled}

type Disbursement struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	TenantID           uint           `gorm:"not null;index" json:"tenant_id"`
	GrantApplicationID uint           `gorm:"not null;index" json:"grant_application_id"`
	Amount             float64        `gorm:"type:decimal(15,2);not null;default:0" json:"amount"`
	Currency           string         `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	Status             string         `gorm:"type:varchar(30);not null;default:'scheduled'" json:"status"`
	PaymentDate        *time.Time     `json:"payment_date,omitempty"`
	EvidenceURL        string         `gorm:"type:text" json:"evidence_url,omitempty"`
	Notes              string         `gorm:"type:text" json:"notes,omitempty"`
	PaidBy             *uint          `json:"paid_by,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Disbursement) TableName() string {
	return "disbursements"
}
