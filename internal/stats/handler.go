package stats

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apiErrors "github.com/fadebowaley/applico/internal/errors"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Summary godoc
// @Summary Tenant-level dashboard summary
// @Description Aggregated counts, amounts, and recent grants for the current tenant
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=SummaryResponse}
// @Router /api/v1/dashboard/summary [get]
func (h *Handler) Summary(c *gin.Context) {
	result, err := h.svc.Summary(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// ProgramStats godoc
// @Summary Program-level statistics
// @Description Grant counts and financial totals for a specific program
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param id path int true "Program ID"
// @Success 200 {object} errors.Response{success=bool,data=ProgramStat}
// @Router /api/v1/dashboard/program-stats/{id} [get]
func (h *Handler) ProgramStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid program ID"))
		return
	}

	result, err := h.svc.ProgramStats(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Program not found"))
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// WorkflowStats godoc
// @Summary Workflow bottleneck analysis
// @Description Active, completed, rejected workflow counts and per-step bottlenecks
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=WorkflowStatsResponse}
// @Router /api/v1/dashboard/workflow-stats [get]
func (h *Handler) WorkflowStats(c *gin.Context) {
	result, err := h.svc.WorkflowStats(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// FinanceSummary godoc
// @Summary Finance summary dashboard
// @Description Aggregated financial data across programs with approval, disbursement rates
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=FinanceSummaryResponse}
// @Router /api/v1/dashboard/finance-summary [get]
func (h *Handler) FinanceSummary(c *gin.Context) {
	result, err := h.svc.FinanceSummary(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// AuditSummary godoc
// @Summary Audit summary dashboard
// @Description Aggregated audit log counts and recent activity
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=AuditSummaryResponse}
// @Router /api/v1/dashboard/audit-summary [get]
func (h *Handler) AuditSummary(c *gin.Context) {
	result, err := h.svc.AuditSummary(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}
	c.JSON(http.StatusOK, apiErrors.Success(result))
}
