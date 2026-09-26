package user

import "time"

const (
	RoleUser              = "user"
	RoleAdmin             = "admin"
	RoleSuperAdmin        = "super_admin"
	RoleHomelandAdmin     = "homeland_admin"
	RoleGrantOfficer      = "grant_officer"
	RoleFinanceOfficer    = "finance_officer"
	RoleAuditor           = "auditor"
	RoleCommitteeMember   = "committee_member"
	RoleRegionalReviewer  = "regional_reviewer"
	RoleProvincialReviewer = "provincial_reviewer"
	RoleZonalReviewer     = "zonal_reviewer"
	RoleAreaReviewer      = "area_reviewer"
	RoleParishPastor      = "parish_pastor"
)

// Role represents a user role in the system
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Role model
func (Role) TableName() string {
	return "roles"
}

// AllEnterpriseRoles returns all enterprise role names
func AllEnterpriseRoles() []string {
	return []string{
		RoleSuperAdmin,
		RoleHomelandAdmin,
		RoleGrantOfficer,
		RoleFinanceOfficer,
		RoleAuditor,
		RoleCommitteeMember,
		RoleRegionalReviewer,
		RoleProvincialReviewer,
		RoleZonalReviewer,
		RoleAreaReviewer,
		RoleParishPastor,
	}
}
