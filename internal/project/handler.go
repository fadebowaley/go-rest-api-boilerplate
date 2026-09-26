package project

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/audit"
	"github.com/fadebowaley/applico/internal/contextutil"
	apiErrors "github.com/fadebowaley/applico/internal/errors"
)

type Handler struct {
	svc      Service
	auditSvc audit.Service
}

func NewHandler(svc Service, auditSvc audit.Service) *Handler {
	return &Handler{svc: svc, auditSvc: auditSvc}
}

// Create godoc
// @Summary Create a project update
// @Description Add a milestone or progress update for a grant
// @Tags project_updates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUpdateRequest true "Update data"
// @Success 201 {object} errors.Response{success=bool,data=UpdateResponse}
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/grants/{grantId}/updates [post]
func (h *Handler) Create(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	var req CreateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "project_updates", strconv.FormatUint(uint64(result.ID), 10), audit.ActionCreated, nil, result, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusCreated, apiErrors.Success(result))
}

// ListByGrant godoc
// @Summary List project updates for a grant
// @Tags project_updates
// @Produce json
// @Security BearerAuth
// @Param grantId path int true "Grant ID"
// @Success 200 {object} errors.Response{success=bool,data=[]UpdateResponse}
// @Router /api/v1/grants/{grantId}/updates [get]
func (h *Handler) ListByGrant(c *gin.Context) {
	grantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	result, err := h.svc.ListByGrant(c.Request.Context(), uint(grantID))
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// GetByID godoc
// @Summary Get a project update by ID
// @Tags project_updates
// @Produce json
// @Security BearerAuth
// @Param id path int true "Update ID"
// @Success 200 {object} errors.Response{success=bool,data=UpdateResponse}
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/project-updates/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid update ID"))
		return
	}

	result, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrUpdateNotFound) {
			_ = c.Error(apiErrors.NotFound("Project update not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// Update godoc
// @Summary Update a project update
// @Tags project_updates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Update ID"
// @Param request body UpdateUpdateRequest true "Update data"
// @Success 200 {object} errors.Response{success=bool,data=UpdateResponse}
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/project-updates/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid update ID"))
		return
	}

	var req UpdateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.svc.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrUpdateNotFound) {
			_ = c.Error(apiErrors.NotFound("Project update not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "project_updates", strconv.FormatUint(uint64(result.ID), 10), audit.ActionUpdated, nil, result, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// Delete godoc
// @Summary Delete a project update
// @Tags project_updates
// @Produce json
// @Security BearerAuth
// @Param id path int true "Update ID"
// @Success 204 {object} errors.Response
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/project-updates/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid update ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrUpdateNotFound) {
			_ = c.Error(apiErrors.NotFound("Project update not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "project_updates", strconv.FormatUint(id, 10), audit.ActionDeleted, nil, nil, c.ClientIP(), c.Request.UserAgent())

	c.Status(http.StatusNoContent)
}
