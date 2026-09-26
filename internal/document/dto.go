package document

import (
	"mime/multipart"
	"time"
)

type UploadDocumentRequest struct {
	GrantID      uint
	UserID       uint
	DocumentType string
	File         *multipart.FileHeader
}

type ReviewDocumentRequest struct {
	Notes string `json:"notes"`
}

type DocumentResponse struct {
	ID                 uint      `json:"id"`
	GrantApplicationID uint      `json:"grant_application_id"`
	UserID             uint      `json:"user_id"`
	DocumentType       string    `json:"document_type"`
	FileName           string    `json:"file_name"`
	OriginalName       string    `json:"original_name"`
	MimeType           string    `json:"mime_type"`
	FileSize           int64     `json:"file_size"`
	Status             string    `json:"status"`
	Notes              string    `json:"notes,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DocumentListResponse struct {
	Documents []DocumentResponse `json:"documents"`
	Total     int64              `json:"total"`
}

func ToDocumentResponse(doc *Document) DocumentResponse {
	return DocumentResponse{
		ID:                 doc.ID,
		GrantApplicationID: doc.GrantApplicationID,
		UserID:             doc.UserID,
		DocumentType:       doc.DocumentType,
		FileName:           doc.FileName,
		OriginalName:       doc.OriginalName,
		MimeType:           doc.MimeType,
		FileSize:           doc.FileSize,
		Status:             doc.Status,
		Notes:              doc.Notes,
		CreatedAt:          doc.CreatedAt,
		UpdatedAt:          doc.UpdatedAt,
	}
}

func ToDocumentListResponse(docs []Document) DocumentListResponse {
	responses := make([]DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = ToDocumentResponse(&doc)
	}
	return DocumentListResponse{Documents: responses, Total: int64(len(docs))}
}
