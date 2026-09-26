package tenant

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/contextutil"
	apiErrors "github.com/fadebowaley/applico/internal/errors"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTenantType):
			_ = c.Error(apiErrors.BadRequest("Invalid tenant type. Valid: faith_based, ngo, foundation, csr, government, development"))
		case errors.Is(err, ErrTenantSlugExists):
			_ = c.Error(apiErrors.Conflict("Tenant slug already exists"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Tenant not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) List(c *gin.Context) {
	resp, err := h.svc.List(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTenantNotFound):
			_ = c.Error(apiErrors.NotFound("Tenant not found"))
		case errors.Is(err, ErrInvalidTenantType):
			_ = c.Error(apiErrors.BadRequest("Invalid tenant type"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		_ = c.Error(apiErrors.NotFound("Tenant not found"))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) AddUser(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	var req AddTenantUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.AddUser(c.Request.Context(), uint(tenantID), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTenantNotFound):
			_ = c.Error(apiErrors.NotFound("Tenant not found"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))
}

func (h *Handler) RemoveUser(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	if err := h.svc.RemoveUser(c.Request.Context(), uint(tenantID), uint(userID)); err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetUsers(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	resp, err := h.svc.GetUsers(c.Request.Context(), uint(tenantID))
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) CreateRole(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	var req CreateTenantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	userID := contextutil.GetUserID(c)
	resp, err := h.svc.CreateRole(c.Request.Context(), userID, uint(tenantID), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotTenantAdmin):
			_ = c.Error(apiErrors.Forbidden("Only tenant admins can manage roles"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))
}

func (h *Handler) ListRoles(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	userID := contextutil.GetUserID(c)
	resp, err := h.svc.ListRoles(c.Request.Context(), userID, uint(tenantID))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotTenantAdmin):
			_ = c.Error(apiErrors.Forbidden("Only tenant admins can view roles"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) GetRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid role ID"))
		return
	}

	userID := contextutil.GetUserID(c)
	resp, err := h.svc.GetRole(c.Request.Context(), userID, uint(roleID))
	if err != nil {
		switch {
		case errors.Is(err, ErrTenantNotFound):
			_ = c.Error(apiErrors.NotFound("Role not found"))
		case errors.Is(err, ErrNotTenantAdmin):
			_ = c.Error(apiErrors.Forbidden("Only tenant admins can view roles"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) UpdateRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid role ID"))
		return
	}

	var req UpdateTenantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	userID := contextutil.GetUserID(c)
	resp, err := h.svc.UpdateRole(c.Request.Context(), userID, uint(roleID), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrTenantNotFound):
			_ = c.Error(apiErrors.NotFound("Role not found"))
		case errors.Is(err, ErrNotTenantAdmin):
			_ = c.Error(apiErrors.Forbidden("Only tenant admins can manage roles"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) DeleteRole(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid role ID"))
		return
	}

	userID := contextutil.GetUserID(c)
	err = h.svc.DeleteRole(c.Request.Context(), userID, uint(roleID))
	if err != nil {
		switch {
		case errors.Is(err, ErrTenantNotFound):
			_ = c.Error(apiErrors.NotFound("Role not found"))
		case errors.Is(err, ErrNotTenantAdmin):
			_ = c.Error(apiErrors.Forbidden("Only tenant admins can manage roles"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) InviteUser(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	inviterID := contextutil.GetUserID(c)

	resp, err := h.svc.InviteUser(c.Request.Context(), inviterID, uint(tenantID), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotTenantAdmin):
			_ = c.Error(apiErrors.Forbidden("You are not a tenant admin"))
		case errors.Is(err, ErrTenantNotFound):
			_ = c.Error(apiErrors.NotFound("Tenant not found"))
		case errors.Is(err, ErrUserNotFoundByEmail):
			_ = c.Error(apiErrors.NotFound("No user found with this email"))
		case errors.Is(err, ErrUserAlreadyInTenant):
			_ = c.Error(apiErrors.Conflict("User is already a member of this tenant"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))
}

func (h *Handler) GetMyTenant(c *gin.Context) {
	tenantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid tenant ID"))
		return
	}

	userID := contextutil.GetUserID(c)

	// Verify the user is a member of this tenant
	_, err = h.svc.GetUserTenant(c.Request.Context(), userID, uint(tenantID))
	if err != nil {
		_ = c.Error(apiErrors.Forbidden("You are not a member of this tenant"))
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), uint(tenantID))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Tenant not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) GetMyTenants(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	resp, err := h.svc.GetUserTenants(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}
