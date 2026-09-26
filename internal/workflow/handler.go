package workflow

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

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))

	userID := contextutil.GetUserID(c)
	h.auditSvc.LogAction(c.Request.Context(), userID, "workflow_templates", strconv.FormatUint(uint64(resp.ID), 10), audit.ActionCreated, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid template ID"))
		return
	}

	resp, err := h.svc.GetTemplate(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow template not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) ListTemplates(c *gin.Context) {
	resp, err := h.svc.ListTemplates(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid template ID"))
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.UpdateTemplate(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			_ = c.Error(apiErrors.NotFound("Workflow template not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))

	userID := contextutil.GetUserID(c)
	h.auditSvc.LogAction(c.Request.Context(), userID, "workflow_templates", strconv.FormatUint(id, 10), audit.ActionUpdated, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid template ID"))
		return
	}

	if err := h.svc.DeleteTemplate(c.Request.Context(), uint(id)); err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow template not found"))
		return
	}

	c.Status(http.StatusNoContent)

	userID := contextutil.GetUserID(c)
	h.auditSvc.LogAction(c.Request.Context(), userID, "workflow_templates", strconv.FormatUint(id, 10), audit.ActionDeleted, nil, nil, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) CreateStep(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid template ID"))
		return
	}

	var req CreateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.CreateStep(c.Request.Context(), uint(templateID), &req)
	if err != nil {
		if errors.Is(err, ErrTemplateNotFound) {
			_ = c.Error(apiErrors.NotFound("Workflow template not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))

	userID := contextutil.GetUserID(c)
	h.auditSvc.LogAction(c.Request.Context(), userID, "workflow_steps", strconv.FormatUint(uint64(resp.ID), 10), audit.ActionCreated, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) ListSteps(c *gin.Context) {
	templateID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid template ID"))
		return
	}

	resp, err := h.svc.ListSteps(c.Request.Context(), uint(templateID))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow template not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) UpdateStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid step ID"))
		return
	}

	var req UpdateStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.UpdateStep(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrStepNotFound) {
			_ = c.Error(apiErrors.NotFound("Workflow step not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))

	userID := contextutil.GetUserID(c)
	h.auditSvc.LogAction(c.Request.Context(), userID, "workflow_steps", strconv.FormatUint(id, 10), audit.ActionUpdated, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) DeleteStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid step ID"))
		return
	}

	if err := h.svc.DeleteStep(c.Request.Context(), uint(id)); err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow step not found"))
		return
	}

	c.Status(http.StatusNoContent)

	userID := contextutil.GetUserID(c)
	h.auditSvc.LogAction(c.Request.Context(), userID, "workflow_steps", strconv.FormatUint(id, 10), audit.ActionDeleted, nil, nil, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) GetStep(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("stepId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid step ID"))
		return
	}

	resp, err := h.svc.GetStep(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow step not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) GetWorkflow(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid application ID"))
		return
	}

	resp, err := h.svc.GetWorkflow(c.Request.Context(), uint(appID))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) Approve(c *gin.Context) {
	h.actOnStep(c, ActionApprove)
}

func (h *Handler) Reject(c *gin.Context) {
	h.actOnStep(c, ActionReject)
}

func (h *Handler) Return(c *gin.Context) {
	h.actOnStep(c, ActionReturn)
}

func (h *Handler) actOnStep(c *gin.Context, action string) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid application ID"))
		return
	}

	var req WorkflowActionRequest
	_ = c.ShouldBindJSON(&req)

	actorID := contextutil.GetUserID(c)
	roles := contextutil.GetRoles(c)

	resp, err := h.svc.ActOnStep(c.Request.Context(), uint(appID), actorID, action, req.Comment, roles)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoWorkflowStarted):
			_ = c.Error(apiErrors.NotFound("No workflow found for this application"))
		case errors.Is(err, ErrWorkflowCompleted):
			_ = c.Error(apiErrors.BadRequest("Workflow is already completed"))
		case errors.Is(err, ErrInvalidAction):
			_ = c.Error(apiErrors.BadRequest("Invalid action"))
		case errors.Is(err, ErrCommentRequired):
			_ = c.Error(apiErrors.BadRequest("Comment is required for rejection or return"))
		case errors.Is(err, ErrNotYourTurn):
			_ = c.Error(apiErrors.Forbidden("You are not authorized to act on this step"))
		case errors.Is(err, ErrAlreadyActed):
			_ = c.Error(apiErrors.BadRequest("You have already acted on this step"))
		case errors.Is(err, ErrStepAlreadyApproved):
			_ = c.Error(apiErrors.BadRequest("This step has already been approved"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))

	h.auditSvc.LogAction(c.Request.Context(), actorID, "workflows", strconv.FormatUint(uint64(appID), 10), action, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) GetHistory(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid application ID"))
		return
	}

	resp, err := h.svc.GetHistory(c.Request.Context(), uint(appID))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Workflow not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) ListPendingApprovals(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	role := c.DefaultQuery("role", "")

	resp, err := h.svc.ListPendingApprovals(c.Request.Context(), userID, role)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}
