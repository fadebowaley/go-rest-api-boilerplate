package audit

import "time"

type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TenantID   uint      `gorm:"not null;index" json:"tenant_id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	EntityType string    `gorm:"type:varchar(50);not null;index" json:"entity_type"`
	EntityID   string    `gorm:"type:varchar(50);not null" json:"entity_id"`
	Action     string    `gorm:"type:varchar(50);not null;index" json:"action"`
	OldValue   *string   `gorm:"type:jsonb" json:"old_value,omitempty"`
	NewValue   *string   `gorm:"type:jsonb" json:"new_value,omitempty"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent  string    `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
