package document

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

const (
	TypeLandDocument         = "land_document"
	TypeSurveyPlan           = "survey_plan"
	TypeBuildingPlan         = "building_plan"
	TypeCostEstimate         = "cost_estimate"
	TypeRecommendationLetter = "recommendation_letter"
	TypeFinancialStatement   = "financial_statement"
	TypeSitePhoto            = "site_photo"
	TypeInvoice              = "invoice"
	TypeReceipt              = "receipt"
	TypeOther                = "other"

	StatusPending  = "pending"
	StatusVerified = "verified"
	StatusRejected = "rejected"
)

var ValidDocumentTypes = []string{
	TypeLandDocument, TypeSurveyPlan, TypeBuildingPlan, TypeCostEstimate,
	TypeRecommendationLetter, TypeFinancialStatement, TypeSitePhoto,
	TypeInvoice, TypeReceipt, TypeOther,
}

var AllowedMimeTypes = map[string]string{
	"application/pdf":              ".pdf",
	"image/jpeg":                   ".jpg",
	"image/png":                    ".png",
	"image/tiff":                   ".tiff",
	"application/msword":           ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel":     ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
}

type Document struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	TenantID           uint           `gorm:"not null;index" json:"tenant_id"`
	GrantApplicationID uint           `gorm:"not null;index" json:"grant_application_id"`
	UserID             uint           `gorm:"not null;index" json:"user_id"`
	DocumentType       string         `gorm:"type:varchar(50);not null" json:"document_type"`
	FileName           string         `gorm:"type:varchar(255);not null" json:"file_name"`
	OriginalName       string         `gorm:"type:varchar(255);not null" json:"original_name"`
	MimeType           string         `gorm:"type:varchar(100);not null" json:"mime_type"`
	FileSize           int64          `gorm:"not null;default:0" json:"file_size"`
	FilePath           string         `gorm:"type:varchar(500);not null" json:"-"`
	Status             string         `gorm:"type:varchar(30);not null;default:pending" json:"status"`
	Notes              string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Document) TableName() string {
	return "documents"
}

func IsValidDocumentType(docType string) bool {
	for _, t := range ValidDocumentTypes {
		if t == docType {
			return true
		}
	}
	return false
}

func IsAllowedMimeType(mimeType string) bool {
	_, ok := AllowedMimeTypes[mimeType]
	return ok
}

type StorageBackend interface {
	Store(ctx context.Context, path string, reader io.Reader) error
	Delete(ctx context.Context, path string) error
	Retrieve(ctx context.Context, path string) (io.ReadCloser, error)
}

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{BasePath: basePath}
}

func (s *LocalStorage) Store(ctx context.Context, path string, reader io.Reader) error {
	fullPath := filepath.Join(s.BasePath, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, reader)
	return err
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	return os.Remove(filepath.Join(s.BasePath, path))
}

func (s *LocalStorage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.BasePath, path))
}
