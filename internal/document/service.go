package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fadebowaley/applico/internal/contextutil"
	"github.com/fadebowaley/applico/internal/grant"
)

var (
	ErrInvalidDocumentType = errors.New("invalid document type")
	ErrInvalidMimeType     = errors.New("file type not allowed")
	ErrDocumentNotFound    = errors.New("document not found")
	ErrFileTooLarge        = errors.New("file exceeds maximum size of 20MB")
	ErrGrantNotFound       = errors.New("grant not found")
	ErrNotDraftGrant       = errors.New("documents can only be uploaded for draft applications")
	ErrNotOwner            = errors.New("you do not own this document")
	ErrCannotModify        = errors.New("verified or rejected documents cannot be modified")
)

const MaxFileSize int64 = 20 * 1024 * 1024

type Service interface {
	Upload(ctx context.Context, grantID, userID uint, docType string, fileHeader *multipart.FileHeader) (*DocumentResponse, error)
	ListByGrant(ctx context.Context, grantID, userID uint, isAdmin bool) (*DocumentListResponse, error)
	GetByID(ctx context.Context, grantID, docID, userID uint, isAdmin bool) (*DocumentResponse, error)
	Replace(ctx context.Context, grantID, docID, userID uint, fileHeader *multipart.FileHeader) (*DocumentResponse, error)
	Delete(ctx context.Context, grantID, docID, userID uint, isAdmin bool) error
	Download(ctx context.Context, docID uint) (io.ReadCloser, string, string, error)
	Verify(ctx context.Context, docID uint, notes string) (*DocumentResponse, error)
	Reject(ctx context.Context, docID uint, notes string) (*DocumentResponse, error)
}

type service struct {
	docRepo  Repository
	grantSvc grant.Service
	storage  StorageBackend
}

func NewService(docRepo Repository, grantSvc grant.Service, storage StorageBackend) Service {
	return &service{
		docRepo:  docRepo,
		grantSvc: grantSvc,
		storage:  storage,
	}
}

func (s *service) Upload(ctx context.Context, grantID, userID uint, docType string, fileHeader *multipart.FileHeader) (*DocumentResponse, error) {
	if !IsValidDocumentType(docType) {
		return nil, ErrInvalidDocumentType
	}
	if fileHeader.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if !IsAllowedMimeType(mimeType) {
		return nil, ErrInvalidMimeType
	}

	grantApp, err := s.grantSvc.GetGrant(ctx, userID, grantID)
	if err != nil {
		return nil, ErrGrantNotFound
	}
	if grantApp.Status != "draft" {
		return nil, ErrNotDraftGrant
	}

	ext := AllowedMimeTypes[mimeType]
	fileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixMilli(), ext)
	filePath := fmt.Sprintf("grants/%d/%s", grantID, fileName)

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	if err := s.storage.Store(ctx, filePath, file); err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	doc := &Document{
		TenantID:           contextutil.GetTenantID(ctx),
		GrantApplicationID: grantID,
		UserID:             userID,
		DocumentType:       docType,
		FileName:           fileName,
		OriginalName:       fileHeader.Filename,
		MimeType:           mimeType,
		FileSize:           fileHeader.Size,
		FilePath:           filePath,
		Status:             StatusPending,
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		if cleanupErr := s.storage.Delete(ctx, filePath); cleanupErr != nil {
			log.Printf("failed to clean up file after DB error: %v", cleanupErr)
		}
		return nil, fmt.Errorf("failed to save document record: %w", err)
	}

	resp := ToDocumentResponse(doc)
	return &resp, nil
}

func (s *service) ListByGrant(ctx context.Context, grantID, userID uint, isAdmin bool) (*DocumentListResponse, error) {
	if !isAdmin {
		if _, err := s.grantSvc.GetGrant(ctx, userID, grantID); err != nil {
			return nil, ErrGrantNotFound
		}
	}
	docs, err := s.docRepo.ListByGrant(ctx, grantID)
	if err != nil {
		return nil, err
	}
	resp := ToDocumentListResponse(docs)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, grantID, docID, userID uint, isAdmin bool) (*DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, ErrDocumentNotFound
	}
	if doc.GrantApplicationID != grantID {
		return nil, ErrDocumentNotFound
	}
	if !isAdmin && doc.UserID != userID {
		return nil, ErrDocumentNotFound
	}
	resp := ToDocumentResponse(doc)
	return &resp, nil
}

func (s *service) Replace(ctx context.Context, grantID, docID, userID uint, fileHeader *multipart.FileHeader) (*DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, ErrDocumentNotFound
	}
	if doc.GrantApplicationID != grantID {
		return nil, ErrDocumentNotFound
	}
	if doc.UserID != userID {
		return nil, ErrNotOwner
	}
	if doc.Status != StatusPending {
		return nil, ErrCannotModify
	}
	if fileHeader.Size > MaxFileSize {
		return nil, ErrFileTooLarge
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if !IsAllowedMimeType(mimeType) {
		return nil, ErrInvalidMimeType
	}

	if err := s.storage.Delete(ctx, doc.FilePath); err != nil {
		log.Printf("failed to delete old file: %v", err)
	}

	ext := AllowedMimeTypes[mimeType]
	doc.FileName = fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().UnixMilli(), ext)
	doc.FilePath = fmt.Sprintf("grants/%d/%s", grantID, doc.FileName)
	doc.OriginalName = fileHeader.Filename
	doc.MimeType = mimeType
	doc.FileSize = fileHeader.Size

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if err := s.storage.Store(ctx, doc.FilePath, file); err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to update document record: %w", err)
	}

	resp := ToDocumentResponse(doc)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, grantID, docID, userID uint, isAdmin bool) error {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return ErrDocumentNotFound
	}
	if doc.GrantApplicationID != grantID {
		return ErrDocumentNotFound
	}
	if !isAdmin && doc.UserID != userID {
		return ErrNotOwner
	}
	if doc.Status != StatusPending {
		return ErrCannotModify
	}

	if err := s.storage.Delete(ctx, doc.FilePath); err != nil {
		log.Printf("failed to delete file from storage: %v", err)
	}

	return s.docRepo.Delete(ctx, docID)
}

func (s *service) Download(ctx context.Context, docID uint) (io.ReadCloser, string, string, error) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, "", "", ErrDocumentNotFound
	}
	reader, err := s.storage.Retrieve(ctx, doc.FilePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("file not found on storage: %w", err)
	}
	return reader, doc.OriginalName, doc.MimeType, nil
}

func (s *service) Verify(ctx context.Context, docID uint, notes string) (*DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, ErrDocumentNotFound
	}
	doc.Status = StatusVerified
	doc.Notes = strings.TrimSpace(notes)
	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	resp := ToDocumentResponse(doc)
	return &resp, nil
}

func (s *service) Reject(ctx context.Context, docID uint, notes string) (*DocumentResponse, error) {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, ErrDocumentNotFound
	}
	doc.Status = StatusRejected
	doc.Notes = strings.TrimSpace(notes)
	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	resp := ToDocumentResponse(doc)
	return &resp, nil
}
