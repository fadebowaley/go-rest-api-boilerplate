package document

import (
	"errors"
	"fmt"
	"io"
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

func (h *Handler) Upload(c *gin.Context) {
	grantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	docType := c.PostForm("document_type")
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("File is required"))
		return
	}
	file.Close()

	userID := contextutil.GetUserID(c)
	resp, err := h.svc.Upload(c.Request.Context(), uint(grantID), userID, docType, fileHeader)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidDocumentType):
			_ = c.Error(apiErrors.BadRequest(fmt.Sprintf("Valid types: %v", ValidDocumentTypes)))
		case errors.Is(err, ErrInvalidMimeType):
			_ = c.Error(apiErrors.BadRequest("File type not supported. Allowed: PDF, JPEG, PNG, TIFF, DOC, DOCX, XLS, XLSX"))
		case errors.Is(err, ErrFileTooLarge):
			_ = c.Error(apiErrors.BadRequest("Maximum file size is 20MB"))
		case errors.Is(err, ErrGrantNotFound):
			_ = c.Error(apiErrors.NotFound("Grant not found"))
		case errors.Is(err, ErrNotDraftGrant):
			_ = c.Error(apiErrors.BadRequest("Documents can only be uploaded for draft applications"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusCreated, apiErrors.Success(resp))

	h.auditSvc.LogAction(c.Request.Context(), userID, "documents", strconv.FormatUint(uint64(resp.ID), 10), audit.ActionUploaded, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) List(c *gin.Context) {
	grantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}

	userID := contextutil.GetUserID(c)
	isAdmin := contextutil.IsAdmin(c)

	resp, err := h.svc.ListByGrant(c.Request.Context(), uint(grantID), userID, isAdmin)
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Grant not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) GetByID(c *gin.Context) {
	grantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}
	docID, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid document ID"))
		return
	}

	userID := contextutil.GetUserID(c)
	isAdmin := contextutil.IsAdmin(c)

	resp, err := h.svc.GetByID(c.Request.Context(), uint(grantID), uint(docID), userID, isAdmin)
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Document not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))
}

func (h *Handler) Replace(c *gin.Context) {
	grantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}
	docID, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid document ID"))
		return
	}

	_, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("File is required"))
		return
	}

	userID := contextutil.GetUserID(c)
	resp, err := h.svc.Replace(c.Request.Context(), uint(grantID), uint(docID), userID, fileHeader)
	if err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			_ = c.Error(apiErrors.NotFound("Document not found"))
		case errors.Is(err, ErrNotOwner):
			_ = c.Error(apiErrors.Forbidden("You do not own this document"))
		case errors.Is(err, ErrCannotModify):
			_ = c.Error(apiErrors.BadRequest("Cannot modify a verified or rejected document"))
		case errors.Is(err, ErrFileTooLarge):
			_ = c.Error(apiErrors.BadRequest("Maximum file size is 20MB"))
		case errors.Is(err, ErrInvalidMimeType):
			_ = c.Error(apiErrors.BadRequest("File type not supported"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))

	h.auditSvc.LogAction(c.Request.Context(), userID, "documents", strconv.FormatUint(uint64(resp.ID), 10), audit.ActionUpdated, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) Delete(c *gin.Context) {
	grantID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid grant ID"))
		return
	}
	docID, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid document ID"))
		return
	}

	userID := contextutil.GetUserID(c)
	isAdmin := contextutil.IsAdmin(c)

	if err := h.svc.Delete(c.Request.Context(), uint(grantID), uint(docID), userID, isAdmin); err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			_ = c.Error(apiErrors.NotFound("Document not found"))
		case errors.Is(err, ErrNotOwner):
			_ = c.Error(apiErrors.Forbidden("You do not own this document"))
		case errors.Is(err, ErrCannotModify):
			_ = c.Error(apiErrors.BadRequest("Cannot delete a verified or rejected document"))
		default:
			_ = c.Error(apiErrors.InternalServerError(err))
		}
		return
	}

	c.Status(http.StatusNoContent)

	h.auditSvc.LogAction(c.Request.Context(), userID, "documents", strconv.FormatUint(uint64(docID), 10), audit.ActionDeleted, nil, nil, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) Download(c *gin.Context) {
	docID, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid document ID"))
		return
	}

	reader, fileName, mimeType, err := h.svc.Download(c.Request.Context(), uint(docID))
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Document not found"))
		return
	}
	defer reader.Close()

	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}

func (h *Handler) Verify(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	docID, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid document ID"))
		return
	}

	var req ReviewDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = ReviewDocumentRequest{}
	}

	resp, err := h.svc.Verify(c.Request.Context(), uint(docID), req.Notes)
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Document not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))

	h.auditSvc.LogAction(c.Request.Context(), userID, "documents", strconv.FormatUint(uint64(resp.ID), 10), audit.ActionVerified, nil, resp, c.ClientIP(), c.Request.UserAgent())
}

func (h *Handler) Reject(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	docID, err := strconv.ParseUint(c.Param("docId"), 10, 64)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid document ID"))
		return
	}

	var req ReviewDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = ReviewDocumentRequest{}
	}

	resp, err := h.svc.Reject(c.Request.Context(), uint(docID), req.Notes)
	if err != nil {
		_ = c.Error(apiErrors.NotFound("Document not found"))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(resp))

	h.auditSvc.LogAction(c.Request.Context(), userID, "documents", strconv.FormatUint(uint64(resp.ID), 10), audit.ActionRejected, nil, resp, c.ClientIP(), c.Request.UserAgent())
}
