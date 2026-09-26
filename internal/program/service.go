package program

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrProgramNotFound = errors.New("grant program not found")
)

type Service interface {
	Create(ctx context.Context, req *CreateProgramRequest) (*ProgramResponse, error)
	GetByID(ctx context.Context, id uint) (*ProgramResponse, error)
	List(ctx context.Context) (*[]ProgramResponse, error)
	Update(ctx context.Context, id uint, req *UpdateProgramRequest) (*ProgramResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req *CreateProgramRequest) (*ProgramResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	var startDate, endDate *time.Time
	if req.ApplicationStartDate != nil {
		t, err := time.Parse("2006-01-02", *req.ApplicationStartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format, use YYYY-MM-DD")
		}
		startDate = &t
	}
	if req.ApplicationEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.ApplicationEndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format, use YYYY-MM-DD")
		}
		endDate = &t
	}

	program := &GrantProgram{
		TenantID:             tenantID,
		Name:                 req.Name,
		Description:          req.Description,
		FormSchema:           req.FormSchema,
		RequiredDocumentTypes: req.RequiredDocumentTypes,
		ApplicantTypes:       req.ApplicantTypes,
		WorkflowTemplateID:   req.WorkflowTemplateID,
		Status:               "active",
		Currency:             "USD",
		EligibilityRules:     json.RawMessage(`{}`),
		ApplicationStartDate: startDate,
		ApplicationEndDate:   endDate,
	}
	if req.BudgetAmount != nil {
		program.BudgetAmount = *req.BudgetAmount
	}
	if req.Currency != nil {
		program.Currency = *req.Currency
	}
	if req.EligibilityRules != nil {
		program.EligibilityRules = req.EligibilityRules
	}

	if err := s.repo.Create(ctx, program); err != nil {
		return nil, fmt.Errorf("failed to create program: %w", err)
	}

	resp := ToProgramResponse(program)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*ProgramResponse, error) {
	program, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProgramNotFound
	}
	resp := ToProgramResponse(program)
	return &resp, nil
}

func (s *service) List(ctx context.Context) (*[]ProgramResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	programs, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	responses := make([]ProgramResponse, len(programs))
	for i, p := range programs {
		responses[i] = ToProgramResponse(&p)
	}
	return &responses, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateProgramRequest) (*ProgramResponse, error) {
	program, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProgramNotFound
	}

	if req.Name != nil {
		program.Name = *req.Name
	}
	if req.Description != nil {
		program.Description = *req.Description
	}
	if req.FormSchema != nil {
		program.FormSchema = *req.FormSchema
	}
	if req.RequiredDocumentTypes != nil {
		program.RequiredDocumentTypes = *req.RequiredDocumentTypes
	}
	if req.ApplicantTypes != nil {
		program.ApplicantTypes = *req.ApplicantTypes
	}
	if req.WorkflowTemplateID != nil {
		program.WorkflowTemplateID = req.WorkflowTemplateID
	}
	if req.Status != nil {
		program.Status = *req.Status
	}
	if req.BudgetAmount != nil {
		program.BudgetAmount = *req.BudgetAmount
	}
	if req.Currency != nil {
		program.Currency = *req.Currency
	}
	if req.EligibilityRules != nil {
		program.EligibilityRules = *req.EligibilityRules
	}
	if req.ApplicationStartDate != nil {
		t, err := time.Parse("2006-01-02", *req.ApplicationStartDate)
		if err == nil {
			program.ApplicationStartDate = &t
		}
	}
	if req.ApplicationEndDate != nil {
		t, err := time.Parse("2006-01-02", *req.ApplicationEndDate)
		if err == nil {
			program.ApplicationEndDate = &t
		}
	}

	if err := s.repo.Update(ctx, program); err != nil {
		return nil, fmt.Errorf("failed to update program: %w", err)
	}

	resp := ToProgramResponse(program)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrProgramNotFound
	}
	return s.repo.Delete(ctx, id)
}
