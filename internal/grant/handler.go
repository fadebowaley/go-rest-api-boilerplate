package grant

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/audit"
	"github.com/fadebowaley/applico/internal/contextutil"
	apiErrors "github.com/fadebowaley/applico/internal/errors"
	"github.com/fadebowaley/applico/internal/middleware"
)

type Handler struct {
	service  Service
	auditSvc audit.Service
}

func NewHandler(service Service, auditSvc audit.Service) *Handler {
	return &Handler{service: service, auditSvc: auditSvc}
}

// CreateGrant godoc
// @Summary Create grant application
// @Description Create a new grant application as draft
// @Tags grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateGrantRequest true "Grant data"
// @Success 201 {object} errors.Response{success=bool,data=GrantResponse}
// @Router /api/v1/grants [post]
func (h *Handler) CreateGrant(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	var req CreateGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.service.CreateGrant(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, ErrParishRequired) {
			_ = c.Error(apiErrors.BadRequest("Parish is required"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(result))

	h.auditSvc.LogAction(c.Request.Context(), userID, "grants", strconv.FormatUint(uint64(result.ID), 10), audit.ActionCreated, nil, result, c.ClientIP(), c.Request.UserAgent())
}

// ListMyGrants godoc
// @Summary List my grant applications
// @Description Get paginated list of current user's grant applications
// @Tags grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param search query string false "Search by title"
// @Success 200 {object} errors.Response{success=bool,data=GrantListResponse}
// @Router /api/v1/grants [get]
func (h *Handler) ListMyGrants(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	pagination := middleware.ParsePaginationParams(c)
	filters := ParseGrantFilters(c)

	result, err := h.service.ListMyGrants(c.Request.Context(), userID, filters, pagination.Page, pagination.PerPage)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// GetGrant godoc
// @Summary Get grant application
// @Description Get a grant application by ID (own grants only)
// @Tags grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Grant ID"
// @Success 200 {object} errors.Response{success=bool,data=GrantResponse}
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/grants/{id} [get]
func (h *Handler) GetGrant(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	result, err := h.service.GetGrant(c.Request.Context(), userID, uint(id))
	if err != nil {
		if errors.Is(err, ErrGrantNotFound) {
			_ = c.Error(apiErrors.NotFound("Grant not found"))
			return
		}
		if errors.Is(err, ErrForbidden) {
			_ = c.Error(apiErrors.Forbidden("You do not have access to this grant"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// UpdateGrant godoc
// @Summary Update grant application
// @Description Update a draft grant application
// @Tags grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Grant ID"
// @Param request body UpdateGrantRequest true "Update data"
// @Success 200 {object} errors.Response{success=bool,data=GrantResponse}
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/grants/{id} [put]
func (h *Handler) UpdateGrant(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	var req UpdateGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.service.UpdateGrant(c.Request.Context(), userID, uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrGrantNotFound) {
			_ = c.Error(apiErrors.NotFound("Grant not found"))
			return
		}
		if errors.Is(err, ErrForbidden) {
			_ = c.Error(apiErrors.Forbidden("You do not have access to this grant"))
			return
		}
		if errors.Is(err, ErrNotDraft) {
			_ = c.Error(apiErrors.BadRequest("Only draft applications can be updated"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))

	h.auditSvc.LogAction(c.Request.Context(), userID, "grants", strconv.FormatUint(uint64(result.ID), 10), audit.ActionUpdated, nil, result, c.ClientIP(), c.Request.UserAgent())
}

// DeleteGrant godoc
// @Summary Delete grant application
// @Description Delete a draft grant application
// @Tags grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Grant ID"
// @Success 204
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/grants/{id} [delete]
func (h *Handler) DeleteGrant(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	if err := h.service.DeleteGrant(c.Request.Context(), userID, uint(id)); err != nil {
		if errors.Is(err, ErrGrantNotFound) {
			_ = c.Error(apiErrors.NotFound("Grant not found"))
			return
		}
		if errors.Is(err, ErrForbidden) {
			_ = c.Error(apiErrors.Forbidden("You do not have access to this grant"))
			return
		}
		if errors.Is(err, ErrNotDraft) {
			_ = c.Error(apiErrors.BadRequest("Only draft applications can be deleted"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)

	h.auditSvc.LogAction(c.Request.Context(), userID, "grants", strconv.FormatUint(uint64(id), 10), audit.ActionDeleted, nil, nil, c.ClientIP(), c.Request.UserAgent())
}

// SubmitGrant godoc
// @Summary Submit grant application
// @Description Submit a draft grant application for review
// @Tags grants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Grant ID"
// @Success 200 {object} errors.Response{success=bool,data=GrantResponse}
// @Router /api/v1/grants/{id}/submit [post]
func (h *Handler) SubmitGrant(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	result, err := h.service.SubmitGrant(c.Request.Context(), userID, uint(id))
	if err != nil {
		if errors.Is(err, ErrGrantNotFound) {
			_ = c.Error(apiErrors.NotFound("Grant not found"))
			return
		}
		if errors.Is(err, ErrForbidden) {
			_ = c.Error(apiErrors.Forbidden("You do not have access to this grant"))
			return
		}
		if errors.Is(err, ErrInvalidTransition) {
			_ = c.Error(apiErrors.BadRequest("Grant cannot be submitted from its current status"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))

	h.auditSvc.LogAction(c.Request.Context(), userID, "grants", strconv.FormatUint(uint64(result.ID), 10), audit.ActionSubmitted, nil, result, c.ClientIP(), c.Request.UserAgent())
}

// Admin endpoints

// ListAllGrants godoc
// @Summary List all grant applications (Admin)
// @Description Get paginated list of all grant applications with filters
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param parish_id query int false "Filter by parish ID"
// @Param search query string false "Search by title"
// @Success 200 {object} errors.Response{success=bool,data=GrantListResponse}
// @Router /api/v1/admin/grants [get]
func (h *Handler) ListAllGrants(c *gin.Context) {
	pagination := middleware.ParsePaginationParams(c)
	filters := ParseGrantFilters(c)

	result, err := h.service.ListAllGrants(c.Request.Context(), filters, pagination.Page, pagination.PerPage)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// GetGrantAsAdmin godoc
// @Summary Get any grant application (Admin)
// @Description Get any grant application by ID (admin access)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Grant ID"
// @Success 200 {object} errors.Response{success=bool,data=GrantResponse}
// @Router /api/v1/admin/grants/{id} [get]
func (h *Handler) GetGrantAsAdmin(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	result, err := h.service.GetGrantAsAdmin(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrGrantNotFound) {
			_ = c.Error(apiErrors.NotFound("Grant not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// ParseGrantFilters parses query parameters into GrantFilterParams
func ParseGrantFilters(c *gin.Context) GrantFilterParams {
	filters := GrantFilterParams{
		Status:   c.Query("status"),
		Region:   c.Query("region"),
		Province: c.Query("province"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		Search:   c.Query("search"),
		Sort:     c.DefaultQuery("sort", "created_at"),
		Order:    c.DefaultQuery("order", "desc"),
	}

	if parishIDStr := c.Query("parish_id"); parishIDStr != "" {
		if id, err := strconv.ParseUint(parishIDStr, 10, 32); err == nil {
			filters.ParishID = uint(id)
		}
	}

	if filters.Order != "asc" && filters.Order != "desc" {
		filters.Order = "desc"
	}

	return filters
}
