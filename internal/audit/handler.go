package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fadebowaley/applico/internal/contextutil"
	apiErrors "github.com/fadebowaley/applico/internal/errors"
	"github.com/fadebowaley/applico/internal/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type AuditLogResponse struct {
	ID         uint        `json:"id"`
	UserID     uint        `json:"user_id"`
	EntityType string      `json:"entity_type"`
	EntityID   string      `json:"entity_id"`
	Action     string      `json:"action"`
	OldValue   interface{} `json:"old_value,omitempty"`
	NewValue   interface{} `json:"new_value,omitempty"`
	IPAddress  string      `json:"ip_address,omitempty"`
	UserAgent  string      `json:"user_agent,omitempty"`
	CreatedAt  string      `json:"created_at"`
}

type AuditLogListResponse struct {
	Logs       []AuditLogResponse `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

func toAuditLogResponse(log *AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:         log.ID,
		UserID:     log.UserID,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		Action:     log.Action,
		OldValue:   log.OldValue,
		NewValue:   log.NewValue,
		IPAddress:  log.IPAddress,
		UserAgent:  log.UserAgent,
		CreatedAt:  log.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListAuditLogs godoc
// @Summary List audit logs
// @Description Get paginated list of audit logs with optional filters (admin only)
// @Tags audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param user_id query int false "Filter by user ID"
// @Param entity_type query string false "Filter by entity type"
// @Param entity_id query string false "Filter by entity ID"
// @Param action query string false "Filter by action"
// @Param since query string false "Filter since (RFC3339)"
// @Param until query string false "Filter until (RFC3339)"
// @Success 200 {object} errors.Response{success=bool,data=AuditLogListResponse}
// @Router /api/v1/admin/audit-logs [get]
func (h *Handler) ListAuditLogs(c *gin.Context) {
	pagination := middleware.ParsePaginationParams(c)

	filters := FilterParams{}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			filters.UserID = uint(id)
		}
	}
	filters.EntityType = c.Query("entity_type")
	filters.EntityID = c.Query("entity_id")
	filters.Action = c.Query("action")

	if sinceStr := c.Query("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filters.Since = t
		}
	}
	if untilStr := c.Query("until"); untilStr != "" {
		if t, err := time.Parse(time.RFC3339, untilStr); err == nil {
			filters.Until = t
		}
	}

	logs, total, err := h.service.List(c.Request.Context(), filters, pagination.Page, pagination.PerPage)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	responses := make([]AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = toAuditLogResponse(&log)
	}

	totalPages := int(total) / pagination.PerPage
	if int(total)%pagination.PerPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, apiErrors.Success(AuditLogListResponse{
		Logs:       responses,
		Total:      total,
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		TotalPages: totalPages,
	}))
}

// LogActionFromRequest godoc
// @Summary Log an audit action
// @Description Create a manual audit log entry (admin/integration)
// @Tags audit
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} errors.Response{success=bool,data=object}
// @Router /api/v1/admin/audit-logs [post]
func (h *Handler) LogActionFromRequest(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	var req CreateAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}
	req.UserID = userID
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	if req.EntityType == "" || req.EntityID == "" || req.Action == "" {
		_ = c.Error(apiErrors.BadRequest("entity_type, entity_id, and action are required"))
		return
	}

	if err := h.service.Log(c.Request.Context(), &req); err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(gin.H{"message": "Audit log created"}))
}
