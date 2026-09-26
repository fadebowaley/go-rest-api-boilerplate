package grant

import (
	"context"
	"errors"
	"fmt"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type WorkflowStarter interface {
	StartWorkflow(ctx context.Context, applicationID, workflowTemplateID uint) error
}

var (
	ErrGrantNotFound       = errors.New("grant application not found")
	ErrNotDraft            = errors.New("only draft applications can be modified")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrForbidden           = errors.New("you do not have permission to modify this grant")
	ErrParishRequired      = errors.New("parish is required")
	ErrTenantRequired      = errors.New("tenant context required")
)

type Service interface {
	CreateGrant(ctx context.Context, userID uint, req *CreateGrantRequest) (*GrantResponse, error)
	GetGrant(ctx context.Context, userID uint, grantID uint) (*GrantResponse, error)
	GetGrantAsAdmin(ctx context.Context, grantID uint) (*GrantResponse, error)
	UpdateGrant(ctx context.Context, userID uint, grantID uint, req *UpdateGrantRequest) (*GrantResponse, error)
	DeleteGrant(ctx context.Context, userID uint, grantID uint) error
	SubmitGrant(ctx context.Context, userID uint, grantID uint) (*GrantResponse, error)
	ListMyGrants(ctx context.Context, userID uint, filters GrantFilterParams, page, perPage int) (*GrantListResponse, error)
	ListAllGrants(ctx context.Context, filters GrantFilterParams, page, perPage int) (*GrantListResponse, error)
	IsApproved(ctx context.Context, grantID uint) (bool, error)
	GetRequestedAmount(ctx context.Context, grantID uint) (float64, error)
	GetGrantStatus(ctx context.Context, grantID uint) (string, error)
	UpdateGrantStatus(ctx context.Context, grantID uint, status string) error
}

type service struct {
	repo       Repository
	workflowSvc WorkflowStarter
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func NewServiceWithWorkflow(repo Repository, workflowSvc WorkflowStarter) Service {
	return &service{repo: repo, workflowSvc: workflowSvc}
}

func (s *service) CreateGrant(ctx context.Context, userID uint, req *CreateGrantRequest) (*GrantResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, ErrTenantRequired
	}

	var workflowTemplateID *uint
	if req.ProgramID != nil {
		program, progErr := s.repo.FindProgramByID(ctx, *req.ProgramID)
		if progErr == nil && program != nil {
			workflowTemplateID = program.WorkflowTemplateID
		}
	}

	grant := &GrantApplication{
		TenantID:               tenantID,
		ProgramID:              req.ProgramID,
		ApplicantID:            req.ApplicantID,
		ParishID:               req.ParishID,
		UserID:                 userID,
		ProjectTitle:           req.ProjectTitle,
		ProjectDescription:     req.ProjectDescription,
		BuildingStage:          req.BuildingStage,
		RequestedAmount:        req.RequestedAmount,
		EstimatedProjectCost:   req.EstimatedProjectCost,
		ProjectLocation:        req.ProjectLocation,
		ProjectStartDate:       parseDate(req.ProjectStartDate),
		ExpectedCompletionDate: parseDate(req.ExpectedCompletionDate),
		Justification:          req.Justification,
		FormResponses:          req.FormResponses,
		WorkflowTemplateID:     workflowTemplateID,
		Status:                 StatusDraft,
	}

	if err := s.repo.Create(ctx, grant); err != nil {
		return nil, fmt.Errorf("failed to create grant: %w", err)
	}

	resp := ToGrantResponse(grant)
	return &resp, nil
}

func (s *service) GetGrant(ctx context.Context, userID uint, grantID uint) (*GrantResponse, error) {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return nil, ErrGrantNotFound
	}
	if grant.UserID != userID {
		return nil, ErrForbidden
	}

	resp := ToGrantResponse(grant)
	return &resp, nil
}

func (s *service) IsApproved(ctx context.Context, grantID uint) (bool, error) {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return false, fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return false, nil
	}
	return grant.Status == StatusApproved, nil
}

func (s *service) GetGrantStatus(ctx context.Context, grantID uint) (string, error) {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return "", fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return "", ErrGrantNotFound
	}
	return grant.Status, nil
}

func (s *service) GetRequestedAmount(ctx context.Context, grantID uint) (float64, error) {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return 0, fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return 0, ErrGrantNotFound
	}
	return grant.RequestedAmount, nil
}

func (s *service) UpdateGrantStatus(ctx context.Context, grantID uint, status string) error {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return ErrGrantNotFound
	}
	grant.Status = status
	return s.repo.Update(ctx, grant)
}

func (s *service) GetGrantAsAdmin(ctx context.Context, grantID uint) (*GrantResponse, error) {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return nil, ErrGrantNotFound
	}

	resp := ToGrantResponse(grant)
	return &resp, nil
}

func (s *service) UpdateGrant(ctx context.Context, userID uint, grantID uint, req *UpdateGrantRequest) (*GrantResponse, error) {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return nil, ErrGrantNotFound
	}
	if grant.UserID != userID {
		return nil, ErrForbidden
	}
	if grant.Status != StatusDraft {
		return nil, ErrNotDraft
	}

	if req.ProjectTitle != "" {
		grant.ProjectTitle = req.ProjectTitle
	}
	if req.ProjectDescription != "" {
		grant.ProjectDescription = req.ProjectDescription
	}
	if req.BuildingStage != "" {
		grant.BuildingStage = req.BuildingStage
	}
	if req.RequestedAmount != nil {
		grant.RequestedAmount = *req.RequestedAmount
	}
	if req.EstimatedProjectCost != nil {
		grant.EstimatedProjectCost = *req.EstimatedProjectCost
	}
	if req.ProjectLocation != "" {
		grant.ProjectLocation = req.ProjectLocation
	}
	if req.ProjectStartDate != "" {
		grant.ProjectStartDate = parseDate(req.ProjectStartDate)
	}
	if req.ExpectedCompletionDate != "" {
		grant.ExpectedCompletionDate = parseDate(req.ExpectedCompletionDate)
	}
	if req.Justification != "" {
		grant.Justification = req.Justification
	}

	if err := s.repo.Update(ctx, grant); err != nil {
		return nil, fmt.Errorf("failed to update grant: %w", err)
	}

	resp := ToGrantResponse(grant)
	return &resp, nil
}

func (s *service) DeleteGrant(ctx context.Context, userID uint, grantID uint) error {
	grant, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return fmt.Errorf("failed to find grant: %w", err)
	}
	if grant == nil {
		return ErrGrantNotFound
	}
	if grant.UserID != userID {
		return ErrForbidden
	}
	if grant.Status != StatusDraft {
		return ErrNotDraft
	}

	if err := s.repo.Delete(ctx, grantID); err != nil {
		return fmt.Errorf("failed to delete grant: %w", err)
	}

	return nil
}

func (s *service) snapshotVersion(ctx context.Context, grant *GrantApplication, userID uint, reason string) error {
	latestVersion, _ := s.repo.GetLatestVersion(ctx, grant.ID)
	nextVersion := 1
	if latestVersion != nil {
		nextVersion = latestVersion.VersionNumber + 1
	}

	v := &GrantApplicationVersion{
		GrantApplicationID:    grant.ID,
		VersionNumber:         nextVersion,
		FormTemplateID:        grant.FormTemplateID,
		FormVersionID:         grant.FormVersionID,
		FormResponseSnapshot:  grant.FormResponses,
		StatusSnapshot:        grant.Status,
		Reason:                reason,
		CreatedBy:             userID,
	}

	return s.repo.CreateVersion(ctx, v)
}

func (s *service) SubmitGrant(ctx context.Context, userID uint, grantID uint) (*GrantResponse, error) {
	grantApp, err := s.repo.FindByID(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find grant: %w", err)
	}
	if grantApp == nil {
		return nil, ErrGrantNotFound
	}
	if grantApp.UserID != userID {
		return nil, ErrForbidden
	}

	if !CanTransition(grantApp.Status, StatusSubmitted) {
		return nil, ErrInvalidTransition
	}

	grantApp.Status = StatusSubmitted
	if err := s.repo.Update(ctx, grantApp); err != nil {
		return nil, fmt.Errorf("failed to submit grant: %w", err)
	}

	if err := s.snapshotVersion(ctx, grantApp, userID, "submitted"); err != nil {
		return nil, fmt.Errorf("grant submitted but failed to create version snapshot: %w", err)
	}

	if s.workflowSvc != nil && grantApp.WorkflowTemplateID != nil {
		if wfErr := s.workflowSvc.StartWorkflow(ctx, grantApp.ID, *grantApp.WorkflowTemplateID); wfErr != nil {
			return nil, fmt.Errorf("grant submitted but failed to start workflow: %w", wfErr)
		}
	}

	resp := ToGrantResponse(grantApp)
	return &resp, nil
}

func (s *service) ListMyGrants(ctx context.Context, userID uint, filters GrantFilterParams, page, perPage int) (*GrantListResponse, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 || perPage > 100 {
		return nil, fmt.Errorf("perPage must be between 1 and 100")
	}

	grants, total, err := s.repo.List(ctx, userID, filters, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list grants: %w", err)
	}

	responses := make([]GrantResponse, len(grants))
	for i, g := range grants {
		responses[i] = ToGrantResponse(&g)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &GrantListResponse{
		Grants:     responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *service) ListAllGrants(ctx context.Context, filters GrantFilterParams, page, perPage int) (*GrantListResponse, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 || perPage > 100 {
		return nil, fmt.Errorf("perPage must be between 1 and 100")
	}

	grants, total, err := s.repo.ListAll(ctx, filters, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list grants: %w", err)
	}

	responses := make([]GrantResponse, len(grants))
	for i, g := range grants {
		responses[i] = ToGrantResponse(&g)
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &GrantListResponse{
		Grants:     responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}
