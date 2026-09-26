package form

import (
	"errors"
	"fmt"
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

func (h *Handler) Create(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	var req CreateFormTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "form_templates", strconv.FormatUint(uint64(result.ID), 10), audit.ActionCreated, nil, result, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusCreated, apiErrors.Success(result))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid ID"))
		return
	}

	result, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrFormTemplateNotFound) {
			_ = c.Error(apiErrors.NotFound("Form template not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

func (h *Handler) List(c *gin.Context) {
	result, err := h.svc.List(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

func (h *Handler) Update(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid ID"))
		return
	}

	var req UpdateFormTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.svc.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrFormTemplateNotFound) {
			_ = c.Error(apiErrors.NotFound("Form template not found"))
			return
		}
		if errors.Is(err, ErrNotDraft) {
			_ = c.Error(apiErrors.BadRequest("Only draft templates can be edited"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "form_templates", strconv.FormatUint(uint64(id), 10), audit.ActionUpdated, nil, result, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

func (h *Handler) Delete(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrFormTemplateNotFound) {
			_ = c.Error(apiErrors.NotFound("Form template not found"))
			return
		}
		if errors.Is(err, ErrAlreadyPublished) {
			_ = c.Error(apiErrors.BadRequest("Cannot delete a published template"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "form_templates", strconv.FormatUint(id, 10), audit.ActionDeleted, nil, nil, c.ClientIP(), c.Request.UserAgent())

	c.Status(http.StatusNoContent)
}

func (h *Handler) Publish(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid ID"))
		return
	}

	notes := c.Query("notes")

	result, err := h.svc.Publish(c.Request.Context(), uint(id), notes)
	if err != nil {
		if errors.Is(err, ErrFormTemplateNotFound) {
			_ = c.Error(apiErrors.NotFound("Form template not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	h.auditSvc.LogAction(c.Request.Context(), userID, "form_templates", strconv.FormatUint(id, 10), audit.ActionUpdated, nil, fmt.Sprintf("published version %d", result.VersionNumber), c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

func (h *Handler) ListVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid ID"))
		return
	}

	result, err := h.svc.ListVersions(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}
