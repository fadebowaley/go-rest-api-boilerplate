package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrTemplateNotFound   = errors.New("workflow template not found")
	ErrStepNotFound       = errors.New("workflow step not found")
	ErrWorkflowNotFound   = errors.New("workflow not found for this application")
	ErrNoWorkflowStarted  = errors.New("no active workflow for this application")
	ErrInvalidAction      = errors.New("invalid action")
	ErrNotYourTurn         = errors.New("you are not authorized to act on this step")
	ErrWorkflowCompleted   = errors.New("workflow is already completed")
	ErrCommentRequired     = errors.New("comment is required for rejection or return")
	ErrAlreadyActed        = errors.New("you have already acted on this step")
	ErrStepAlreadyApproved = errors.New("this step has already been approved")
)

type Service interface {
	CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*TemplateResponse, error)
	GetTemplate(ctx context.Context, id uint) (*TemplateResponse, error)
	ListTemplates(ctx context.Context) (*[]TemplateResponse, error)
	UpdateTemplate(ctx context.Context, id uint, req *UpdateTemplateRequest) (*TemplateResponse, error)
	DeleteTemplate(ctx context.Context, id uint) error

	CreateStep(ctx context.Context, templateID uint, req *CreateStepRequest) (*StepResponse, error)
	GetStep(ctx context.Context, id uint) (*StepResponse, error)
	ListSteps(ctx context.Context, templateID uint) (*[]StepResponse, error)
	UpdateStep(ctx context.Context, id uint, req *UpdateStepRequest) (*StepResponse, error)
	DeleteStep(ctx context.Context, id uint) error

	StartWorkflow(ctx context.Context, applicationID, workflowTemplateID uint) (*ApplicationWorkflowResponse, error)
	GetWorkflow(ctx context.Context, applicationID uint) (*ApplicationWorkflowResponse, error)
	ActOnStep(ctx context.Context, applicationID, actorID uint, action, comment string, roles []string) (*ApplicationWorkflowResponse, error)
	GetHistory(ctx context.Context, applicationID uint) (*[]ActionResponse, error)
	ListPendingApprovals(ctx context.Context, userID uint, role string) (*[]PendingApprovalResponse, error)
	SetGrantUpdater(updater GrantStatusUpdater)
}

type GrantStatusUpdater interface {
	UpdateGrantStatus(ctx context.Context, grantID uint, status string) error
}

type service struct {
	repo          Repository
	grantUpdater  GrantStatusUpdater
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func NewServiceWithUpdater(repo Repository, updater GrantStatusUpdater) Service {
	return &service{repo: repo, grantUpdater: updater}
}

func (s *service) SetGrantUpdater(updater GrantStatusUpdater) {
	s.grantUpdater = updater
}

func (s *service) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*TemplateResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	t := &WorkflowTemplate{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.CreateTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to create workflow template: %w", err)
	}

	resp := ToTemplateResponse(t)
	return &resp, nil
}

func (s *service) GetTemplate(ctx context.Context, id uint) (*TemplateResponse, error) {
	t, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, ErrTemplateNotFound
	}
	resp := ToTemplateResponse(t)
	return &resp, nil
}

func (s *service) ListTemplates(ctx context.Context) (*[]TemplateResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	templates, err := s.repo.ListTemplatesByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	responses := make([]TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = ToTemplateResponse(&t)
	}
	return &responses, nil
}

func (s *service) UpdateTemplate(ctx context.Context, id uint, req *UpdateTemplateRequest) (*TemplateResponse, error) {
	t, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = *req.Description
	}

	if err := s.repo.UpdateTemplate(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	resp := ToTemplateResponse(t)
	return &resp, nil
}

func (s *service) DeleteTemplate(ctx context.Context, id uint) error {
	if _, err := s.repo.GetTemplateByID(ctx, id); err != nil {
		return ErrTemplateNotFound
	}
	return s.repo.DeleteTemplate(ctx, id)
}

func (s *service) CreateStep(ctx context.Context, templateID uint, req *CreateStepRequest) (*StepResponse, error) {
	if _, err := s.repo.GetTemplateByID(ctx, templateID); err != nil {
		return nil, ErrTemplateNotFound
	}

	step := &WorkflowStep{
		WorkflowTemplateID: templateID,
		Name:               req.Name,
		StepOrder:          req.StepOrder,
		AssigneeRoles:      req.AssigneeRoles,
	}

	if err := s.repo.CreateStep(ctx, step); err != nil {
		return nil, fmt.Errorf("failed to create step: %w", err)
	}

	resp := ToStepResponse(step)
	return &resp, nil
}

func (s *service) GetStep(ctx context.Context, id uint) (*StepResponse, error) {
	step, err := s.repo.GetStepByID(ctx, id)
	if err != nil {
		return nil, ErrStepNotFound
	}
	resp := ToStepResponse(step)
	return &resp, nil
}

func (s *service) ListSteps(ctx context.Context, templateID uint) (*[]StepResponse, error) {
	if _, err := s.repo.GetTemplateByID(ctx, templateID); err != nil {
		return nil, ErrTemplateNotFound
	}

	steps, err := s.repo.ListStepsByTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	responses := make([]StepResponse, len(steps))
	for i, step := range steps {
		responses[i] = ToStepResponse(&step)
	}
	return &responses, nil
}

func (s *service) UpdateStep(ctx context.Context, id uint, req *UpdateStepRequest) (*StepResponse, error) {
	step, err := s.repo.GetStepByID(ctx, id)
	if err != nil {
		return nil, ErrStepNotFound
	}

	if req.Name != nil {
		step.Name = *req.Name
	}
	if req.StepOrder != nil {
		step.StepOrder = *req.StepOrder
	}
	if req.AssigneeRoles != nil {
		step.AssigneeRoles = *req.AssigneeRoles
	}

	if err := s.repo.UpdateStep(ctx, step); err != nil {
		return nil, fmt.Errorf("failed to update step: %w", err)
	}

	resp := ToStepResponse(step)
	return &resp, nil
}

func (s *service) DeleteStep(ctx context.Context, id uint) error {
	if _, err := s.repo.GetStepByID(ctx, id); err != nil {
		return ErrStepNotFound
	}
	return s.repo.DeleteStep(ctx, id)
}

func (s *service) StartWorkflow(ctx context.Context, applicationID, workflowTemplateID uint) (*ApplicationWorkflowResponse, error) {
	steps, err := s.repo.ListStepsByTemplate(ctx, workflowTemplateID)
	if err != nil || len(steps) == 0 {
		return nil, ErrStepNotFound
	}

	tenantID := contextutil.GetTenantID(ctx)
	wf := &ApplicationWorkflow{
		TenantID:           tenantID,
		ApplicationID:      applicationID,
		WorkflowTemplateID: workflowTemplateID,
		CurrentStepID:      &steps[0].ID,
		Status:             WorkflowStatusInProgress,
		StartedAt:          time.Now(),
	}

	if err := s.repo.CreateAppWorkflow(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	resp := ToAppWorkflowResponse(wf)
	return &resp, nil
}

func (s *service) GetWorkflow(ctx context.Context, applicationID uint) (*ApplicationWorkflowResponse, error) {
	wf, err := s.repo.GetAppWorkflowByApplication(ctx, applicationID)
	if err != nil {
		return nil, ErrNoWorkflowStarted
	}

	actions, _ := s.repo.ListActionsByWorkflow(ctx, wf.ID)
	resp := ToAppWorkflowResponse(wf)
	resp.Actions = make([]ActionResponse, len(actions))
	for i, a := range actions {
		step, _ := s.repo.GetStepByID(ctx, a.StepID)
		stepName := ""
		if step != nil {
			stepName = step.Name
		}
		resp.Actions[i] = ActionResponse{
			ID:        a.ID,
			StepID:    a.StepID,
			StepName:  stepName,
			ActorID:   a.ActorID,
			Action:    a.Action,
			Comment:   a.Comment,
			CreatedAt: a.CreatedAt,
		}
	}
	return &resp, nil
}

func (s *service) ActOnStep(ctx context.Context, applicationID, actorID uint, action, comment string, roles []string) (*ApplicationWorkflowResponse, error) {
	if action != ActionApprove && action != ActionReject && action != ActionReturn {
		return nil, ErrInvalidAction
	}

	if (action == ActionReject || action == ActionReturn) && comment == "" {
		return nil, ErrCommentRequired
	}

	wf, err := s.repo.GetAppWorkflowByApplication(ctx, applicationID)
	if err != nil {
		return nil, ErrNoWorkflowStarted
	}

	if wf.Status != WorkflowStatusInProgress {
		return nil, ErrWorkflowCompleted
	}

	if wf.CurrentStepID == nil {
		return nil, ErrNoWorkflowStarted
	}

	step, err := s.repo.GetStepByID(ctx, *wf.CurrentStepID)
	if err != nil {
		return nil, ErrStepNotFound
	}

	var assigneeRoles []string
	if err := json.Unmarshal(step.AssigneeRoles, &assigneeRoles); err != nil {
		return nil, fmt.Errorf("failed to parse assignee roles: %w", err)
	}
	if len(assigneeRoles) > 0 {
		hasRole := false
		for _, stepRole := range assigneeRoles {
			for _, userRole := range roles {
				if stepRole == userRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		if !hasRole {
			return nil, ErrNotYourTurn
		}
	}

	approved, err := s.repo.HasApprovedAction(ctx, wf.ID, step.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check step approval status: %w", err)
	}
	if approved {
		return nil, ErrStepAlreadyApproved
	}

	exists, err := s.repo.ActionExists(ctx, wf.ID, step.ID, actorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing action: %w", err)
	}
	if exists {
		return nil, ErrAlreadyActed
	}

	wa := &WorkflowAction{
		TenantID:              contextutil.GetTenantID(ctx),
		ApplicationWorkflowID: wf.ID,
		StepID:                step.ID,
		ActorID:               actorID,
		Action:                action,
		Comment:               comment,
	}

	if err := s.repo.CreateAction(ctx, wa); err != nil {
		return nil, fmt.Errorf("failed to record action: %w", err)
	}

	switch action {
	case ActionApprove:
		steps, _ := s.repo.ListStepsByTemplate(ctx, wf.WorkflowTemplateID)
		var nextStep *WorkflowStep
		for i, st := range steps {
			if st.ID == step.ID && i+1 < len(steps) {
				nextStep = &steps[i+1]
				break
			}
		}
		if nextStep != nil {
			wf.CurrentStepID = &nextStep.ID
		} else {
			wf.Status = WorkflowStatusCompleted
			now := time.Now()
			wf.CompletedAt = &now
			wf.CurrentStepID = nil
			if s.grantUpdater != nil {
				_ = s.grantUpdater.UpdateGrantStatus(ctx, applicationID, "approved")
			}
		}

	case ActionReject:
		wf.Status = WorkflowStatusRejected
		now := time.Now()
		wf.CompletedAt = &now
		if s.grantUpdater != nil {
			_ = s.grantUpdater.UpdateGrantStatus(ctx, applicationID, "rejected")
		}

	case ActionReturn:
		steps, _ := s.repo.ListStepsByTemplate(ctx, wf.WorkflowTemplateID)
		var prevStep *WorkflowStep
		for i, st := range steps {
			if st.ID == step.ID && i > 0 {
				prevStep = &steps[i-1]
				break
			}
		}
		if prevStep != nil {
			wf.CurrentStepID = &prevStep.ID
		}
	}

	if err := s.repo.UpdateAppWorkflow(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return s.GetWorkflow(ctx, applicationID)
}

func (s *service) GetHistory(ctx context.Context, applicationID uint) (*[]ActionResponse, error) {
	wf, err := s.repo.GetAppWorkflowByApplication(ctx, applicationID)
	if err != nil {
		return nil, ErrNoWorkflowStarted
	}

	actions, err := s.repo.ListActionsByWorkflow(ctx, wf.ID)
	if err != nil {
		return nil, err
	}

	responses := make([]ActionResponse, len(actions))
	for i, a := range actions {
		step, _ := s.repo.GetStepByID(ctx, a.StepID)
		stepName := ""
		if step != nil {
			stepName = step.Name
		}
		responses[i] = ActionResponse{
			ID:        a.ID,
			StepID:    a.StepID,
			StepName:  stepName,
			ActorID:   a.ActorID,
			Action:    a.Action,
			Comment:   a.Comment,
			CreatedAt: a.CreatedAt,
		}
	}
	return &responses, nil
}

func (s *service) ListPendingApprovals(ctx context.Context, userID uint, role string) (*[]PendingApprovalResponse, error) {
	results, err := s.repo.ListPendingApprovals(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	return &results, nil
}
