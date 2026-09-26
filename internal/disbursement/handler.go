package disbursement

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
// @Summary Create a disbursement
// @Description Create a new disbursement for an approved grant
// @Tags disbursements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateDisbursementRequest true "Disbursement data"
// @Success 201 {object} errors.Response{success=bool,data=DisbursementResponse}
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/disbursements [post]
func (h *Handler) Create(c *gin.Context) {
	userID := contextutil.GetUserID(c)

	var req CreateDisbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	result, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, ErrGrantNotApproved) {
			_ = c.Error(apiErrors.BadRequest("Grant is not approved"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(result))

	h.auditSvc.LogAction(c.Request.Context(), userID, "disbursements", strconv.FormatUint(uint64(result.ID), 10), audit.ActionCreated, nil, result, c.ClientIP(), c.Request.UserAgent())
}

// ListByGrant godoc
// @Summary List disbursements for a grant
// @Tags disbursements
// @Produce json
// @Security BearerAuth
// @Param grantId path int true "Grant ID"
// @Success 200 {object} errors.Response{success=bool,data=[]DisbursementResponse}
// @Router /api/v1/grants/{grantId}/disbursements [get]
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
// @Summary Get a disbursement by ID
// @Tags disbursements
// @Produce json
// @Security BearerAuth
// @Param id path int true "Disbursement ID"
// @Success 200 {object} errors.Response{success=bool,data=DisbursementResponse}
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/disbursements/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid disbursement ID"))
		return
	}

	result, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrDisbursementNotFound) {
			_ = c.Error(apiErrors.NotFound("Disbursement not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))
}

// Update godoc
// @Summary Update a disbursement
// @Tags disbursements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Disbursement ID"
// @Param request body UpdateDisbursementRequest true "Update data"
// @Success 200 {object} errors.Response{success=bool,data=DisbursementResponse}
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/disbursements/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid disbursement ID"))
		return
	}

	userID := contextutil.GetUserID(c)

	var req UpdateDisbursementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	oldResp, _ := h.svc.GetByID(c.Request.Context(), uint(id))

	result, err := h.svc.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrDisbursementNotFound) {
			_ = c.Error(apiErrors.NotFound("Disbursement not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(result))

	h.auditSvc.LogAction(c.Request.Context(), userID, "disbursements", strconv.FormatUint(uint64(result.ID), 10), audit.ActionUpdated, oldResp, result, c.ClientIP(), c.Request.UserAgent())
}

// Delete godoc
// @Summary Delete a disbursement
// @Tags disbursements
// @Produce json
// @Security BearerAuth
// @Param id path int true "Disbursement ID"
// @Success 204 {object} errors.Response
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/disbursements/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid disbursement ID"))
		return
	}

	userID := contextutil.GetUserID(c)

	oldResp, _ := h.svc.GetByID(c.Request.Context(), uint(id))

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrDisbursementNotFound) {
			_ = c.Error(apiErrors.NotFound("Disbursement not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)

	if oldResp != nil {
		h.auditSvc.LogAction(c.Request.Context(), userID, "disbursements", strconv.FormatUint(uint64(oldResp.ID), 10), audit.ActionDeleted, oldResp, nil, c.ClientIP(), c.Request.UserAgent())
	}
}
