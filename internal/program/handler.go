package program

import (
	"errors"
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

func (h *Handler) Create(c *gin.Context) {
	var req CreateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid program ID"))
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Program not found"))
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
		_ = c.Error(apiErrors.BadRequest("Invalid program ID"))
		return
	}

	var req UpdateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrProgramNotFound) {
			_ = c.Error(apiErrors.NotFound("Program not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid program ID"))
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		_ = c.Error(apiErrors.NotFound("Program not found"))
		return
	}

	c.Status(http.StatusNoContent)
}
