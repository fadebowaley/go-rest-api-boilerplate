package form

import (
	"context"
	"errors"
	"fmt"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrFormTemplateNotFound = errors.New("form template not found")
	ErrTenantRequired       = errors.New("tenant context required")
	ErrNotDraft             = errors.New("only draft templates can be edited")
	ErrAlreadyPublished     = errors.New("template is already published")
)

type Service interface {
	Create(ctx context.Context, req *CreateFormTemplateRequest) (*FormTemplateResponse, error)
	GetByID(ctx context.Context, id uint) (*FormTemplateResponse, error)
	List(ctx context.Context) ([]FormTemplateResponse, error)
	Update(ctx context.Context, id uint, req *UpdateFormTemplateRequest) (*FormTemplateResponse, error)
	Delete(ctx context.Context, id uint) error
	Publish(ctx context.Context, id uint, notes string) (*FormVersionResponse, error)
	ListVersions(ctx context.Context, templateID uint) ([]FormVersionResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateFormTemplateRequest) (*FormTemplateResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	schema := req.SchemaDefinition
	if schema == nil {
		schema = []byte("{}")
	}

	t := &FormTemplate{
		TenantID:         tenantID,
		Name:             req.Name,
		Description:      req.Description,
		SchemaDefinition: schema,
		Status:           status,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to create form template: %w", err)
	}

	resp := ToFormTemplateResponse(t)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*FormTemplateResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFormTemplateNotFound
	}

	resp := ToFormTemplateResponse(t)
	return &resp, nil
}

func (s *service) List(ctx context.Context) ([]FormTemplateResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	templates, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list form templates: %w", err)
	}

	resp := make([]FormTemplateResponse, len(templates))
	for i, t := range templates {
		resp[i] = ToFormTemplateResponse(&t)
	}
	return resp, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateFormTemplateRequest) (*FormTemplateResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFormTemplateNotFound
	}

	if t.Status != "draft" {
		return nil, ErrNotDraft
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.SchemaDefinition != nil {
		t.SchemaDefinition = *req.SchemaDefinition
	}
	if req.Status != nil {
		t.Status = *req.Status
	}

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to update form template: %w", err)
	}

	resp := ToFormTemplateResponse(t)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrFormTemplateNotFound
	}
	if t.Status == "published" {
		return ErrAlreadyPublished
	}
	return s.repo.Delete(ctx, id)
}

func (s *service) Publish(ctx context.Context, id uint, notes string) (*FormVersionResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFormTemplateNotFound
	}

	latestVersion, err := s.repo.GetLatestVersion(ctx, id)
	nextVersion := 1
	if err == nil && latestVersion != nil {
		nextVersion = latestVersion.VersionNumber + 1
	}

	v := &FormVersion{
		FormTemplateID:   t.ID,
		VersionNumber:    nextVersion,
		SchemaDefinition: t.SchemaDefinition,
		Notes:            notes,
	}

	if err := s.repo.CreateVersion(ctx, v); err != nil {
		return nil, fmt.Errorf("failed to create form version: %w", err)
	}

	t.Status = "published"
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to update template status: %w", err)
	}

	resp := ToFormVersionResponse(v)
	return &resp, nil
}

func (s *service) ListVersions(ctx context.Context, templateID uint) ([]FormVersionResponse, error) {
	versions, err := s.repo.ListVersions(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to list form versions: %w", err)
	}

	resp := make([]FormVersionResponse, len(versions))
	for i, v := range versions {
		resp[i] = ToFormVersionResponse(&v)
	}
	return resp, nil
}
