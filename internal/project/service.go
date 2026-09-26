package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrUpdateNotFound = errors.New("project update not found")
	ErrInvalidStatus  = errors.New("invalid status")
)

type GrantChecker interface {
	IsApproved(ctx context.Context, grantID uint) (bool, error)
	GetGrantStatus(ctx context.Context, grantID uint) (string, error)
}

type Service interface {
	Create(ctx context.Context, userID uint, req *CreateUpdateRequest) (*UpdateResponse, error)
	GetByID(ctx context.Context, id uint) (*UpdateResponse, error)
	ListByGrant(ctx context.Context, grantID uint) (*[]UpdateResponse, error)
	Update(ctx context.Context, id uint, req *UpdateUpdateRequest) (*UpdateResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo         Repository
	grantChecker GrantChecker
}

func NewService(repo Repository, grantChecker GrantChecker) Service {
	return &service{repo: repo, grantChecker: grantChecker}
}

func (s *service) Create(ctx context.Context, userID uint, req *CreateUpdateRequest) (*UpdateResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	status, err := s.grantChecker.GetGrantStatus(ctx, req.GrantApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check grant status: %w", err)
	}
	if status != "approved" && status != "disbursed" {
		return nil, fmt.Errorf("grant is not approved or disbursed")
	}

	var milestoneDate *time.Time
	if req.MilestoneDate != "" {
		t, parseErr := time.Parse(time.RFC3339, req.MilestoneDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid milestone_date format: %w", parseErr)
		}
		milestoneDate = &t
	}

	photos := req.Photos
	if photos == nil {
		photos = []byte("[]")
	}
	invoices := req.Invoices
	if invoices == nil {
		invoices = []byte("[]")
	}

	u := &ProjectUpdate{
		TenantID:           tenantID,
		GrantApplicationID: req.GrantApplicationID,
		Title:              req.Title,
		Description:        req.Description,
		Status:             StatusPlanned,
		MilestoneDate:      milestoneDate,
		Photos:             photos,
		Invoices:           invoices,
		Notes:              req.Notes,
		CreatedBy:          userID,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create project update: %w", err)
	}

	resp := ToUpdateResponse(u)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*UpdateResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find project update: %w", err)
	}
	if u == nil {
		return nil, ErrUpdateNotFound
	}
	resp := ToUpdateResponse(u)
	return &resp, nil
}

func (s *service) ListByGrant(ctx context.Context, grantID uint) (*[]UpdateResponse, error) {
	list, err := s.repo.ListByGrant(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list project updates: %w", err)
	}
	responses := make([]UpdateResponse, len(list))
	for i, u := range list {
		responses[i] = ToUpdateResponse(&u)
	}
	return &responses, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateUpdateRequest) (*UpdateResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find project update: %w", err)
	}
	if u == nil {
		return nil, ErrUpdateNotFound
	}

	if req.Title != nil {
		u.Title = *req.Title
	}
	if req.Description != nil {
		u.Description = *req.Description
	}
	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		u.Status = *req.Status
	}
	if req.MilestoneDate != nil {
		if *req.MilestoneDate == "" {
			u.MilestoneDate = nil
		} else {
			t, parseErr := time.Parse(time.RFC3339, *req.MilestoneDate)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid milestone_date format: %w", parseErr)
			}
			u.MilestoneDate = &t
		}
	}
	if req.Photos != nil {
		u.Photos = *req.Photos
	}
	if req.Invoices != nil {
		u.Invoices = *req.Invoices
	}
	if req.Notes != nil {
		u.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update project update: %w", err)
	}

	resp := ToUpdateResponse(u)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find project update: %w", err)
	}
	if u == nil {
		return ErrUpdateNotFound
	}
	return s.repo.Delete(ctx, id)
}

func isValidStatus(status string) bool {
	for _, s := range ValidStatuses {
		if s == status {
			return true
		}
	}
	return false
}
