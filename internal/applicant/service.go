package applicant

import (
	"context"
	"errors"
	"fmt"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrApplicantNotFound = errors.New("applicant not found")
)

type Service interface {
	Create(ctx context.Context, req *CreateApplicantRequest) (*ApplicantResponse, error)
	GetByID(ctx context.Context, id uint) (*ApplicantResponse, error)
	List(ctx context.Context) (*[]ApplicantResponse, error)
	Update(ctx context.Context, id uint, req *UpdateApplicantRequest) (*ApplicantResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateApplicantRequest) (*ApplicantResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	applicant := &Applicant{
		TenantID:    tenantID,
		Type:        req.Type,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		ProfileData: req.ProfileData,
		Status:      "active",
	}

	if err := s.repo.Create(ctx, applicant); err != nil {
		return nil, fmt.Errorf("failed to create applicant: %w", err)
	}

	resp := ToApplicantResponse(applicant)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*ApplicantResponse, error) {
	applicant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrApplicantNotFound
	}
	resp := ToApplicantResponse(applicant)
	return &resp, nil
}

func (s *service) List(ctx context.Context) (*[]ApplicantResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	applicants, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	responses := make([]ApplicantResponse, len(applicants))
	for i, a := range applicants {
		responses[i] = ToApplicantResponse(&a)
	}
	return &responses, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateApplicantRequest) (*ApplicantResponse, error) {
	applicant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrApplicantNotFound
	}

	if req.Name != nil {
		applicant.Name = *req.Name
	}
	if req.Type != nil {
		applicant.Type = *req.Type
	}
	if req.Email != nil {
		applicant.Email = *req.Email
	}
	if req.Phone != nil {
		applicant.Phone = *req.Phone
	}
	if req.Address != nil {
		applicant.Address = *req.Address
	}
	if req.Status != nil {
		applicant.Status = *req.Status
	}
	if req.ProfileData != nil {
		applicant.ProfileData = *req.ProfileData
	}

	if err := s.repo.Update(ctx, applicant); err != nil {
		return nil, fmt.Errorf("failed to update applicant: %w", err)
	}

	resp := ToApplicantResponse(applicant)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrApplicantNotFound
	}
	return s.repo.Delete(ctx, id)
}
