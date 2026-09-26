package applicant

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Applicant struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	TenantID    uint            `gorm:"not null;index" json:"tenant_id"`
	Type        string          `gorm:"type:varchar(100);not null;default:individual" json:"type"`
	Name        string          `gorm:"type:varchar(255);not null" json:"name"`
	Email       string          `gorm:"type:varchar(255)" json:"email"`
	Phone       string          `gorm:"type:varchar(50)" json:"phone"`
	Address     string          `gorm:"type:text" json:"address"`
	ProfileData json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"profile_data"`
	Status      string          `gorm:"type:varchar(30);not null;default:active" json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (Applicant) TableName() string {
	return "applicants"
}
