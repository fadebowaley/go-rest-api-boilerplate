package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/contextutil"
	"github.com/fadebowaley/applico/internal/tenant"
)

// RequireTenant is middleware that extracts tenant_id from the X-Tenant-ID header,
// verifies the user is a member (or is super_admin), and makes it available in the request context.
// All tenant-scoped routes should use this middleware.
func RequireTenant(tenantSvc tenant.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr == "" {
			tenantIDStr = c.Query("tenant_id")
		}
		if tenantIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TENANT_REQUIRED",
					"message": "X-Tenant-ID header or tenant_id query parameter is required",
				},
			})
			return
		}

		tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
		if err != nil || tenantID == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TENANT_ID",
					"message": "Invalid tenant ID",
				},
			})
			return
		}

		// super_admin bypass — can access any tenant
		if contextutil.HasRole(c, "super_admin") {
			tc := contextutil.NewTenantContext(uint(tenantID))
			ctx := contextutil.WithTenantContext(c.Request.Context(), tc)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		// Verify user is a member of this tenant
		userID := contextutil.GetUserID(c)
		_, err = tenantSvc.GetUserTenant(c.Request.Context(), userID, uint(tenantID))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "NOT_TENANT_MEMBER",
					"message": "You are not a member of this tenant",
				},
			})
			return
		}

		tc := contextutil.NewTenantContext(uint(tenantID))
		ctx := contextutil.WithTenantContext(c.Request.Context(), tc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
