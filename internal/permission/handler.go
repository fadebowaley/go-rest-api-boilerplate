package permission

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
	"github.com/fadebowaley/applico/internal/middleware"
	apiErrors "github.com/fadebowaley/applico/internal/errors"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreatePermission godoc
// @Summary Create a new permission
// @Description Create a new permission definition (admin only)
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePermissionRequest true "Permission data"
// @Success 201 {object} errors.Response{success=bool,data=PermissionResponse}
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Failure 409 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/admin/permissions [post]
func (h *Handler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.service.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrPermissionExists) {
			_ = c.Error(apiErrors.Conflict("Permission already exists"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(result))
}

// ListPermissions godoc
// @Summary List all permissions
// @Description Get paginated list of permissions with optional group filter (admin only)
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param group query string false "Filter by group name"
// @Success 200 {object} errors.Response{success=bool,data=PermissionListResponse}
// @Router /api/v1/admin/permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	pagination := middleware.ParsePaginationParams(c)
	group := c.Query("group")

	result, err := h.service.ListPermissions(c.Request.Context(), group, pagination.Page, pagination.PerPage)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// DeletePermission godoc
// @Summary Delete a permission
// @Description Delete a permission definition (admin only)
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Permission ID"
// @Success 204
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/admin/permissions/{id} [delete]
func (h *Handler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid permission ID"))
		return
	}

	if err := h.service.DeletePermission(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			_ = c.Error(apiErrors.NotFound("Permission not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignPermissionsToRole godoc
// @Summary Assign permissions to a role
// @Description Replace all permissions for a role (admin only)
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AssignPermissionsRequest true "Role-permission assignment"
// @Success 200 {object} errors.Response{success=bool,data=object}
// @Router /api/v1/admin/roles/{roleId}/permissions [put]
func (h *Handler) AssignPermissionsToRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid role ID"))
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}
	req.RoleID = uint(roleID)

	if err := h.service.AssignPermissionsToRole(c.Request.Context(), &req); err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Permissions assigned successfully"}))
}

// GetPermissionsForRole godoc
// @Summary Get permissions for a role
// @Description Get all permissions assigned to a specific role
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roleId path int true "Role ID"
// @Success 200 {object} errors.Response{success=bool,data=RolePermissionsResponse}
// @Router /api/v1/admin/roles/{roleId}/permissions [get]
func (h *Handler) GetPermissionsForRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid role ID"))
		return
	}

	result, err := h.service.GetPermissionsForRole(c.Request.Context(), uint(roleID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = c.Error(apiErrors.NotFound("Role not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// GetMyPermissions godoc
// @Summary Get current user permissions
// @Description Get all permissions for the currently authenticated user
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=[]string}
// @Router /api/v1/auth/me/permissions [get]
func (h *Handler) GetMyPermissions(c *gin.Context) {
	userID, err := mustGetUserID(c)
	if err != nil {
		_ = c.Error(apiErrors.Unauthorized(err.Error()))
		return
	}

	perms, err := h.service.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	if perms == nil {
		perms = []string{}
	}

	c.JSON(http.StatusOK, apiErrors.Success(perms))
}

func mustGetUserID(c *gin.Context) (uint, error) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		return 0, errors.New("user not authenticated")
	}
	return userID, nil
}
