package workflow

import (
	"context"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	CreateTemplate(ctx context.Context, t *WorkflowTemplate) error
	GetTemplateByID(ctx context.Context, id uint) (*WorkflowTemplate, error)
	ListTemplatesByTenant(ctx context.Context, tenantID uint) ([]WorkflowTemplate, error)
	UpdateTemplate(ctx context.Context, t *WorkflowTemplate) error
	DeleteTemplate(ctx context.Context, id uint) error

	CreateStep(ctx context.Context, s *WorkflowStep) error
	GetStepByID(ctx context.Context, id uint) (*WorkflowStep, error)
	ListStepsByTemplate(ctx context.Context, templateID uint) ([]WorkflowStep, error)
	UpdateStep(ctx context.Context, s *WorkflowStep) error
	DeleteStep(ctx context.Context, id uint) error

	CreateAppWorkflow(ctx context.Context, w *ApplicationWorkflow) error
	GetAppWorkflowByApplication(ctx context.Context, appID uint) (*ApplicationWorkflow, error)
	GetAppWorkflowByID(ctx context.Context, id uint) (*ApplicationWorkflow, error)
	UpdateAppWorkflow(ctx context.Context, w *ApplicationWorkflow) error
	CreateAction(ctx context.Context, a *WorkflowAction) error
	ListActionsByWorkflow(ctx context.Context, workflowID uint) ([]WorkflowAction, error)
	ListPendingApprovals(ctx context.Context, userID uint, role string) ([]PendingApprovalResponse, error)
	ActionExists(ctx context.Context, workflowID, stepID, actorID uint) (bool, error)
	HasApprovedAction(ctx context.Context, workflowID, stepID uint) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateTemplate(ctx context.Context, t *WorkflowTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repository) GetTemplateByID(ctx context.Context, id uint) (*WorkflowTemplate, error) {
	var t WorkflowTemplate
	err := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) ListTemplatesByTenant(ctx context.Context, tenantID uint) ([]WorkflowTemplate, error) {
	var templates []WorkflowTemplate
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&templates).Error
	return templates, err
}

func (r *repository) UpdateTemplate(ctx context.Context, t *WorkflowTemplate) error {
	var existing WorkflowTemplate
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", t.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *repository) DeleteTemplate(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&WorkflowTemplate{}, id).Error
}

func (r *repository) CreateStep(ctx context.Context, s *WorkflowStep) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repository) GetStepByID(ctx context.Context, id uint) (*WorkflowStep, error) {
	var s WorkflowStep
	err := r.db.WithContext(ctx).
		Joins("JOIN workflow_templates wt ON wt.id = workflow_steps.workflow_template_id").
		Where("workflow_steps.id = ? AND wt.tenant_id = ?", id, contextutil.GetTenantID(ctx)).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repository) ListStepsByTemplate(ctx context.Context, templateID uint) ([]WorkflowStep, error) {
	var steps []WorkflowStep
	err := r.db.WithContext(ctx).
		Where("workflow_template_id = ?", templateID).
		Order("step_order ASC").
		Find(&steps).Error
	return steps, err
}

func (r *repository) UpdateStep(ctx context.Context, s *WorkflowStep) error {
	var existing WorkflowStep
	err := r.db.WithContext(ctx).
		Joins("JOIN workflow_templates wt ON wt.id = workflow_steps.workflow_template_id").
		Where("workflow_steps.id = ? AND wt.tenant_id = ?", s.ID, contextutil.GetTenantID(ctx)).
		First(&existing).Error
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *repository) DeleteStep(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND workflow_template_id IN (SELECT id FROM workflow_templates WHERE tenant_id = ?)", id, contextutil.GetTenantID(ctx)).
		Delete(&WorkflowStep{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) CreateAppWorkflow(ctx context.Context, w *ApplicationWorkflow) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *repository) GetAppWorkflowByApplication(ctx context.Context, appID uint) (*ApplicationWorkflow, error) {
	var w ApplicationWorkflow
	err := r.db.WithContext(ctx).
		Joins("JOIN grant_applications ga ON ga.id = application_workflows.application_id").
		Where("application_workflows.application_id = ? AND ga.tenant_id = ?", appID, contextutil.GetTenantID(ctx)).
		First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repository) GetAppWorkflowByID(ctx context.Context, id uint) (*ApplicationWorkflow, error) {
	var w ApplicationWorkflow
	err := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *repository) UpdateAppWorkflow(ctx context.Context, w *ApplicationWorkflow) error {
	var existing ApplicationWorkflow
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", w.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *repository) CreateAction(ctx context.Context, a *WorkflowAction) error {
	var w ApplicationWorkflow
	if err := r.db.WithContext(ctx).Where("id = ?", a.ApplicationWorkflowID).First(&w).Error; err != nil {
		return err
	}
	a.TenantID = w.TenantID
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repository) ListActionsByWorkflow(ctx context.Context, workflowID uint) ([]WorkflowAction, error) {
	var actions []WorkflowAction
	err := r.db.WithContext(ctx).
		Joins("JOIN application_workflows aw ON aw.id = workflow_actions.application_workflow_id").
		Where("workflow_actions.application_workflow_id = ? AND aw.tenant_id = ?", workflowID, contextutil.GetTenantID(ctx)).
		Order("workflow_actions.created_at ASC").
		Find(&actions).Error
	return actions, err
}

func (r *repository) ListPendingApprovals(ctx context.Context, userID uint, role string) ([]PendingApprovalResponse, error) {
	var results []PendingApprovalResponse
	err := r.db.WithContext(ctx).Raw(`
		SELECT aw.id AS workflow_id, aw.application_id, ga.project_title,
		       aw.current_step_id, ws.name AS current_step_name
		FROM application_workflows aw
		JOIN grant_applications ga ON ga.id = aw.application_id
		JOIN workflow_steps ws ON ws.id = aw.current_step_id
		WHERE aw.status = 'in_progress'
		AND ga.tenant_id = ?
		AND (ws.assignee_roles @> ?::jsonb OR ? = ANY(SELECT jsonb_array_elements_text(ws.assignee_roles)))
		AND NOT EXISTS (
		    SELECT 1 FROM workflow_actions wa
		    WHERE wa.application_workflow_id = aw.id AND wa.step_id = aw.current_step_id
		)
	`, contextutil.GetTenantID(ctx), `["`+role+`"]`, role).Scan(&results).Error
	return results, err
}

func (r *repository) ActionExists(ctx context.Context, workflowID, stepID, actorID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&WorkflowAction{}).
		Where("application_workflow_id = ? AND step_id = ? AND actor_id = ?", workflowID, stepID, actorID).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) HasApprovedAction(ctx context.Context, workflowID, stepID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&WorkflowAction{}).
		Where("application_workflow_id = ? AND step_id = ? AND action = ?", workflowID, stepID, ActionApprove).
		Count(&count).Error
	return count > 0, err
}
