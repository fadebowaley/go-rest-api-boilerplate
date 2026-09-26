package contextutil

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/auth"
)

// GetTenantID extracts the tenant ID from a context set by RequireTenant middleware.
// Returns 0 if no tenant context is present.
func GetTenantID(ctx context.Context) uint {
	if tc := GetTenantContext(ctx); tc != nil {
		return tc.TenantID()
	}
	return 0
}

func GetUser(c *gin.Context) *auth.Claims {
	value, exists := c.Get(auth.KeyUser)
	if !exists {
		return nil
	}

	claims, ok := value.(*auth.Claims)
	if !ok {
		return nil
	}

	return claims
}

func MustGetUser(c *gin.Context) (*auth.Claims, error) {
	claims := GetUser(c)
	if claims == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return claims, nil
}

func GetUserID(c *gin.Context) uint {
	claims := GetUser(c)
	if claims == nil {
		return 0
	}
	return claims.UserID
}

func MustGetUserID(c *gin.Context) (uint, error) {
	userID := GetUserID(c)
	if userID == 0 {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

func GetEmail(c *gin.Context) string {
	claims := GetUser(c)
	if claims == nil {
		return ""
	}
	return claims.Email
}

func IsAuthenticated(c *gin.Context) bool {
	return GetUser(c) != nil
}

func CanAccessUser(c *gin.Context, targetUserID uint) bool {
	if IsAdmin(c) {
		return true
	}
	authenticatedUserID := GetUserID(c)
	return authenticatedUserID == targetUserID
}

func GetUserName(c *gin.Context) string {
	claims := GetUser(c)
	if claims == nil {
		return ""
	}
	return claims.Name
}

func HasRole(c *gin.Context, role string) bool {
	claims := GetUser(c)
	if claims == nil {
		return false
	}
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func GetRoles(c *gin.Context) []string {
	claims := GetUser(c)
	if claims == nil {
		return []string{}
	}
	return claims.Roles
}

func IsAdmin(c *gin.Context) bool {
	return HasRole(c, "admin") || HasRole(c, "super_admin") || HasRole(c, "homeland_admin")
}

// HasPermission checks if the authenticated user has a specific permission
func HasPermission(c *gin.Context, permission string) bool {
	claims := GetUser(c)
	if claims == nil {
		return false
	}
	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetPermissions returns the list of permissions for the authenticated user
func GetPermissions(c *gin.Context) []string {
	claims := GetUser(c)
	if claims == nil {
		return []string{}
	}
	return claims.Permissions
}

// GetTenantIDs returns the list of tenant IDs the authenticated user belongs to
func GetTenantIDs(c *gin.Context) []uint {
	claims := GetUser(c)
	if claims == nil {
		return []uint{}
	}
	return claims.TenantIDs
}

// HasTenantPermission checks if the user has a specific permission scoped to a tenant.
// This checks tenant_roles permissions (not global role_permissions).
func HasTenantPermission(c *gin.Context, tenantID uint, permission string) bool {
	claims := GetUser(c)
	if claims == nil || claims.TenantPermissions == nil {
		return false
	}
	tid := fmt.Sprintf("%d", tenantID)
	perms, ok := claims.TenantPermissions[tid]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// GetTenantPermissions returns the tenant-scoped permissions for a specific tenant.
func GetTenantPermissions(c *gin.Context, tenantID uint) []string {
	claims := GetUser(c)
	if claims == nil || claims.TenantPermissions == nil {
		return []string{}
	}
	tid := fmt.Sprintf("%d", tenantID)
	perms, ok := claims.TenantPermissions[tid]
	if !ok {
		return []string{}
	}
	return perms
}

// HasPermissionInTenant checks if the user has a permission considering BOTH
// global permissions (from role_permissions) AND tenant-scoped permissions (from tenant_roles).
// This is the main authorization check for tenant-scoped routes.
func HasPermissionInTenant(c *gin.Context, tenantID uint, permission string) bool {
	if HasPermission(c, permission) {
		return true
	}
	return HasTenantPermission(c, tenantID, permission)
}
