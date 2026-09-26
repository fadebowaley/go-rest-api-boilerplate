package stats

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Summary(ctx context.Context, tenantID uint) (*SummaryResponse, error)
	ProgramStats(ctx context.Context, tenantID, programID uint) (*ProgramStat, error)
	WorkflowStats(ctx context.Context, tenantID uint) (*WorkflowStatsResponse, error)
	FinanceSummary(ctx context.Context, tenantID uint) (*FinanceSummaryResponse, error)
	AuditSummary(ctx context.Context, tenantID uint) (*AuditSummaryResponse, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Summary(ctx context.Context, tenantID uint) (*SummaryResponse, error) {
	type grantCounts struct {
		Total    int64
		Draft    int64
		Submitted int64
		Approved int64
		Rejected int64
		Disbursed int64
	}
	var gc grantCounts
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0) AS draft,
			COALESCE(SUM(CASE WHEN status = 'submitted' THEN 1 ELSE 0 END), 0) AS submitted,
			COALESCE(SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END), 0) AS approved,
			COALESCE(SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0) AS rejected,
			COALESCE(SUM(CASE WHEN status IN ('disbursed','closed') THEN 1 ELSE 0 END), 0) AS disbursed
		FROM grant_applications WHERE tenant_id = ? AND deleted_at IS NULL
	`, tenantID).Scan(&gc).Error; err != nil {
		return nil, fmt.Errorf("failed to count grants: %w", err)
	}

	var requested, disbursed float64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(requested_amount), 0) FROM grant_applications WHERE tenant_id = ? AND deleted_at IS NULL
	`, tenantID).Scan(&requested).Error; err != nil {
		return nil, fmt.Errorf("failed to sum requested: %w", err)
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM disbursements WHERE tenant_id = ? AND deleted_at IS NULL AND status = 'paid'
	`, tenantID).Scan(&disbursed).Error; err != nil {
		return nil, fmt.Errorf("failed to sum disbursed: %w", err)
	}

	var activePrograms int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM grant_programs WHERE tenant_id = ? AND status = 'active' AND deleted_at IS NULL
	`, tenantID).Scan(&activePrograms).Error; err != nil {
		return nil, fmt.Errorf("failed to count programs: %w", err)
	}

	var pendingWorkflows int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM application_workflows aw
		JOIN grant_applications ga ON ga.id = aw.application_id
		WHERE ga.tenant_id = ? AND aw.status = 'in_progress'
	`, tenantID).Scan(&pendingWorkflows).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending workflows: %w", err)
	}

	type recentRow struct {
		ID             uint
		ProjectTitle   string
		Status         string
		RequestedAmount float64
		CreatedAt      time.Time
		ProgramName    *string
	}
	var rows []recentRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT ga.id, ga.project_title, ga.status, ga.requested_amount, ga.created_at, gp.name AS program_name
		FROM grant_applications ga
		LEFT JOIN grant_programs gp ON gp.id = ga.program_id
		WHERE ga.tenant_id = ? AND ga.deleted_at IS NULL
		ORDER BY ga.created_at DESC LIMIT 10
	`, tenantID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch recent grants: %w", err)
	}
	recent := make([]RecentGrant, len(rows))
	for i, r := range rows {
		pn := ""
		if r.ProgramName != nil {
			pn = *r.ProgramName
		}
		recent[i] = RecentGrant{
			ID:              r.ID,
			ProjectTitle:    r.ProjectTitle,
			Status:          r.Status,
			RequestedAmount: r.RequestedAmount,
			CreatedAt:       r.CreatedAt,
			ProgramName:     pn,
		}
	}

	return &SummaryResponse{
		TotalGrants:     gc.Total,
		DraftGrants:     gc.Draft,
		SubmittedGrants: gc.Submitted,
		ApprovedGrants:  gc.Approved,
		RejectedGrants:  gc.Rejected,
		DisbursedGrants: gc.Disbursed,
		TotalRequested:  requested,
		TotalDisbursed:  disbursed,
		ActivePrograms:  activePrograms,
		PendingWorkflows: pendingWorkflows,
		RecentGrants:    recent,
	}, nil
}

func (r *repository) ProgramStats(ctx context.Context, tenantID, programID uint) (*ProgramStat, error) {
	var ps ProgramStat
	ps.ProgramID = programID

	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(gp.name, '') AS program_name,
			COUNT(*) AS total_grants,
			COALESCE(SUM(CASE WHEN ga.status = 'approved' THEN 1 ELSE 0 END), 0) AS approved_grants,
			COALESCE(SUM(CASE WHEN ga.status = 'rejected' THEN 1 ELSE 0 END), 0) AS rejected_grants,
			COALESCE(SUM(CASE WHEN ga.status IN ('draft','submitted','under_review','pending_clarification') THEN 1 ELSE 0 END), 0) AS pending_grants,
			COALESCE(SUM(ga.requested_amount), 0) AS total_requested
		FROM grant_applications ga
		LEFT JOIN grant_programs gp ON gp.id = ga.program_id
		WHERE ga.tenant_id = ? AND ga.program_id = ? AND ga.deleted_at IS NULL
		GROUP BY gp.name
	`, tenantID, programID).Scan(&ps).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate program stats: %w", err)
	}
	ps.ProgramID = programID

	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(d.amount), 0) FROM disbursements d
		JOIN grant_applications ga ON ga.id = d.grant_application_id
		WHERE ga.tenant_id = ? AND ga.program_id = ? AND d.status = 'paid' AND d.deleted_at IS NULL
	`, tenantID, programID).Scan(&ps.TotalDisbursed).Error; err != nil {
		return nil, fmt.Errorf("failed to sum program disbursements: %w", err)
	}

	return &ps, nil
}

func (r *repository) WorkflowStats(ctx context.Context, tenantID uint) (*WorkflowStatsResponse, error) {
	var resp WorkflowStatsResponse

	type wfCounts struct {
		TotalActive    int64
		TotalCompleted int64
		TotalRejected  int64
	}
	var wc wfCounts
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN aw.status = 'in_progress' THEN 1 ELSE 0 END), 0) AS total_active,
			COALESCE(SUM(CASE WHEN aw.status = 'completed' THEN 1 ELSE 0 END), 0) AS total_completed,
			COALESCE(SUM(CASE WHEN aw.status = 'rejected' THEN 1 ELSE 0 END), 0) AS total_rejected
		FROM application_workflows aw
		JOIN grant_applications ga ON ga.id = aw.application_id
		WHERE ga.tenant_id = ?
	`, tenantID).Scan(&wc).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate workflow stats: %w", err)
	}
	resp.TotalActive = wc.TotalActive
	resp.TotalCompleted = wc.TotalCompleted
	resp.TotalRejected = wc.TotalRejected

	var avgHours float64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (aw.completed_at - aw.started_at)) / 3600), 0)
		FROM application_workflows aw
		JOIN grant_applications ga ON ga.id = aw.application_id
		WHERE ga.tenant_id = ? AND aw.status = 'completed' AND aw.completed_at IS NOT NULL
	`, tenantID).Scan(&avgHours).Error; err != nil {
		return nil, fmt.Errorf("failed to compute avg workflow duration: %w", err)
	}
	resp.AvgCompletionHours = avgHours

	type stepRow struct {
		StepName     string
		TemplateName string
		PendingCount int64
		AvgHours     float64
	}
	var steps []stepRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			ws.name AS step_name,
			wt.name AS template_name,
			COUNT(*) AS pending_count
		FROM application_workflows aw
		JOIN grant_applications ga ON ga.id = aw.application_id
		JOIN workflow_steps ws ON ws.id = aw.current_step_id
		JOIN workflow_templates wt ON wt.id = aw.workflow_template_id
		WHERE ga.tenant_id = ? AND aw.status = 'in_progress'
		GROUP BY ws.name, wt.name
		ORDER BY pending_count DESC
	`, tenantID).Scan(&steps).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate step breakdown: %w", err)
	}
	resp.StepBreakdown = make([]WorkflowStat, len(steps))
	for i, s := range steps {
		resp.StepBreakdown[i] = WorkflowStat{
			StepName:     s.StepName,
			TemplateName: s.TemplateName,
			PendingCount: s.PendingCount,
		}
	}

	return &resp, nil
}

func (r *repository) FinanceSummary(ctx context.Context, tenantID uint) (*FinanceSummaryResponse, error) {
	type programRow struct {
		ProgramID      uint    `gorm:"column:program_id"`
		ProgramName    string  `gorm:"column:program_name"`
		TotalRequested float64 `gorm:"column:total_requested"`
		TotalApproved  float64 `gorm:"column:total_approved"`
		TotalDisbursed float64 `gorm:"column:total_disbursed"`
		GrantCount     int64   `gorm:"column:grant_count"`
	}
	var programs []programRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(gp.id, 0) AS program_id,
			COALESCE(gp.name, 'Unassigned') AS program_name,
			COALESCE(SUM(ga.requested_amount), 0) AS total_requested,
			COALESCE(SUM(CASE WHEN ga.status = 'approved' THEN ga.requested_amount ELSE 0 END), 0) AS total_approved,
			COALESCE(SUM(CASE WHEN d.status = 'paid' THEN d.amount ELSE 0 END), 0) AS total_disbursed,
			COUNT(ga.id) AS grant_count
		FROM grant_applications ga
		LEFT JOIN grant_programs gp ON gp.id = ga.program_id
		LEFT JOIN disbursements d ON d.grant_application_id = ga.id AND d.deleted_at IS NULL
		WHERE ga.tenant_id = ? AND ga.deleted_at IS NULL
		GROUP BY gp.id, gp.name
		ORDER BY total_requested DESC
	`, tenantID).Scan(&programs).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate finance summary: %w", err)
	}

	programBreakdown := make([]ProgramFinanceSummary, len(programs))
	for i, p := range programs {
		programBreakdown[i] = ProgramFinanceSummary{
			ProgramID:      p.ProgramID,
			ProgramName:    p.ProgramName,
			TotalRequested: p.TotalRequested,
			TotalApproved:  p.TotalApproved,
			TotalDisbursed: p.TotalDisbursed,
			GrantCount:     p.GrantCount,
		}
	}

	type grantTotals struct {
		TotalRequested float64 `gorm:"column:total_requested"`
		TotalApproved  float64 `gorm:"column:total_approved"`
		TotalPending   float64 `gorm:"column:total_pending"`
		TotalCancelled float64 `gorm:"column:total_cancelled"`
	}
	var gt grantTotals
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(requested_amount), 0) AS total_requested,
			COALESCE(SUM(CASE WHEN status = 'approved' THEN requested_amount ELSE 0 END), 0) AS total_approved,
			COALESCE(SUM(CASE WHEN status IN ('draft','submitted','under_review','pending_clarification') THEN requested_amount ELSE 0 END), 0) AS total_pending,
			COALESCE(SUM(CASE WHEN status IN ('rejected','cancelled') THEN requested_amount ELSE 0 END), 0) AS total_cancelled
		FROM grant_applications
		WHERE tenant_id = ? AND deleted_at IS NULL
	`, tenantID).Scan(&gt).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate grant financial totals: %w", err)
	}
	totalRequested, totalApproved, totalPending, totalCancelled := gt.TotalRequested, gt.TotalApproved, gt.TotalPending, gt.TotalCancelled

	var totalDisbursed float64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM disbursements WHERE tenant_id = ? AND status = 'paid' AND deleted_at IS NULL
	`, tenantID).Scan(&totalDisbursed).Error; err != nil {
		return nil, fmt.Errorf("failed to sum disbursed: %w", err)
	}

	disbursementRate := 0.0
	if totalApproved > 0 {
		disbursementRate = totalDisbursed / totalApproved * 100
	}

	return &FinanceSummaryResponse{
		TotalApproved:    totalApproved,
		TotalRequested:   totalRequested,
		TotalDisbursed:   totalDisbursed,
		TotalPending:     totalPending,
		TotalCancelled:   totalCancelled,
		DisbursementRate: disbursementRate,
		ProgramBreakdown: programBreakdown,
	}, nil
}

func (r *repository) AuditSummary(ctx context.Context, tenantID uint) (*AuditSummaryResponse, error) {
	var totalActions int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM audit_logs WHERE tenant_id = ?
	`, tenantID).Scan(&totalActions).Error; err != nil {
		return nil, fmt.Errorf("failed to count audit actions: %w", err)
	}

	type actionCountRow struct {
		Action string `gorm:"column:action"`
		Count  int64  `gorm:"column:count"`
	}
	var rows []actionCountRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT action, COUNT(*) AS count FROM audit_logs WHERE tenant_id = ? GROUP BY action ORDER BY count DESC
	`, tenantID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate actions by type: %w", err)
	}

	actionsByType := make([]ActionCount, len(rows))
	for i, row := range rows {
		actionsByType[i] = ActionCount{Action: row.Action, Count: row.Count}
	}

	type recentRow struct {
		ID         uint
		UserID     uint
		EntityType string
		EntityID   string
		Action     string
		CreatedAt  time.Time
	}
	var recentRows []recentRow
	if err := r.db.WithContext(ctx).Raw(`
		SELECT id, user_id, entity_type, entity_id, action, created_at
		FROM audit_logs WHERE tenant_id = ?
		ORDER BY created_at DESC LIMIT 20
	`, tenantID).Scan(&recentRows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch recent audit actions: %w", err)
	}

	recentActions := make([]RecentAction, len(recentRows))
	for i, r := range recentRows {
		recentActions[i] = RecentAction{
			ID:         r.ID,
			UserID:     r.UserID,
			EntityType: r.EntityType,
			EntityID:   r.EntityID,
			Action:     r.Action,
			CreatedAt:  r.CreatedAt,
		}
	}

	return &AuditSummaryResponse{
		TotalActions:  totalActions,
		ActionsByType: actionsByType,
		RecentActions: recentActions,
	}, nil
}
