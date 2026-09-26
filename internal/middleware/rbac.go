package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/contextutil"
	"github.com/fadebowaley/applico/internal/errors"
)

// RequireRole returns a middleware that checks if the user has the specified role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !contextutil.HasRole(c, role) {
			c.JSON(http.StatusForbidden, errors.Forbidden("insufficient permissions"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin returns a middleware that checks if the user is an admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequirePermission returns a middleware that checks if the user has the specified permission
// checking BOTH global permissions and tenant-scoped permissions.
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if contextutil.HasPermission(c, permission) {
			c.Next()
			return
		}
		// Also check tenant-scoped permissions if a tenant context is active
		if tenantID := contextutil.GetTenantID(c.Request.Context()); tenantID > 0 {
			if contextutil.HasTenantPermission(c, tenantID, permission) {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, errors.Forbidden("insufficient permissions"))
		c.Abort()
	}
}

// RequireAnyRole returns a middleware that checks if the user has any of the specified roles
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, role := range roles {
			if contextutil.HasRole(c, role) {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, errors.Forbidden("insufficient permissions"))
		c.Abort()
	}
}

// RequireAnyPermission returns a middleware that checks if the user has any of the specified permissions,
// checking BOTH global permissions and tenant-scoped permissions.
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := contextutil.GetTenantID(c.Request.Context())
		for _, perm := range permissions {
			if contextutil.HasPermission(c, perm) {
				c.Next()
				return
			}
			if tenantID > 0 && contextutil.HasTenantPermission(c, tenantID, perm) {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, errors.Forbidden("insufficient permissions"))
		c.Abort()
	}
}

// RequireEnterpriseAdmin returns a middleware for Super Admin and Homeland Admin roles
func RequireEnterpriseAdmin() gin.HandlerFunc {
	return RequireAnyRole("super_admin", "homeland_admin")
}
