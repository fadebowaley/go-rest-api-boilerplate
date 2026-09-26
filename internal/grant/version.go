package grant

import (
	"encoding/json"
	"time"
)

type GrantApplicationVersion struct {
	ID                   uint             `gorm:"primaryKey" json:"id"`
	GrantApplicationID   uint             `gorm:"not null;index" json:"grant_application_id"`
	VersionNumber        int              `gorm:"not null" json:"version_number"`
	FormTemplateID       *uint            `json:"form_template_id,omitempty"`
	FormVersionID        *uint            `json:"form_version_id,omitempty"`
	FormResponseSnapshot json.RawMessage  `gorm:"type:jsonb;not null;default:'{}'" json:"form_response_snapshot"`
	DocumentSnapshot     json.RawMessage  `gorm:"type:jsonb;not null;default:'[]'" json:"document_snapshot"`
	StatusSnapshot       string           `gorm:"type:varchar(30);not null;default:draft" json:"status_snapshot"`
	Reason               string           `gorm:"type:varchar(30);not null;default:submitted" json:"reason"`
	CreatedBy            uint             `gorm:"not null" json:"created_by"`
	CreatedAt            time.Time        `json:"created_at"`
}

func (GrantApplicationVersion) TableName() string {
	return "grant_application_versions"
}
