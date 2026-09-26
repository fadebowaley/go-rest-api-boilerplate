package stats

import "time"

type ProgramStat struct {
	ProgramID          uint    `json:"program_id"`
	ProgramName        string  `json:"program_name"`
	TotalGrants        int64   `json:"total_grants"`
	ApprovedGrants     int64   `json:"approved_grants"`
	RejectedGrants     int64   `json:"rejected_grants"`
	PendingGrants      int64   `json:"pending_grants"`
	TotalRequested     float64 `json:"total_requested"`
	TotalDisbursed     float64 `json:"total_disbursed"`
}

type SummaryResponse struct {
	TotalGrants    int64         `json:"total_grants"`
	DraftGrants    int64         `json:"draft_grants"`
	SubmittedGrants int64        `json:"submitted_grants"`
	ApprovedGrants int64         `json:"approved_grants"`
	RejectedGrants int64         `json:"rejected_grants"`
	DisbursedGrants int64        `json:"disbursed_grants"`
	TotalRequested float64       `json:"total_requested"`
	TotalDisbursed float64       `json:"total_disbursed"`
	ActivePrograms int64         `json:"active_programs"`
	PendingWorkflows int64       `json:"pending_workflows"`
	RecentGrants   []RecentGrant `json:"recent_grants"`
}

type RecentGrant struct {
	ID           uint      `json:"id"`
	ProjectTitle string    `json:"project_title"`
	Status       string    `json:"status"`
	RequestedAmount float64 `json:"requested_amount"`
	CreatedAt    time.Time `json:"created_at"`
	ProgramName  string    `json:"program_name,omitempty"`
}

type WorkflowStat struct {
	StepName       string `json:"step_name"`
	TemplateName   string `json:"template_name"`
	PendingCount   int64  `json:"pending_count"`
	AvgDurationHours float64 `json:"avg_duration_hours,omitempty"`
}

type WorkflowStatsResponse struct {
	TotalActive     int64          `json:"total_active"`
	TotalCompleted  int64          `json:"total_completed"`
	TotalRejected   int64          `json:"total_rejected"`
	AvgCompletionHours float64     `json:"avg_completion_hours"`
	StepBreakdown   []WorkflowStat `json:"step_breakdown"`
}

type FinanceSummaryResponse struct {
	TotalApproved    float64                  `json:"total_approved"`
	TotalRequested   float64                  `json:"total_requested"`
	TotalDisbursed   float64                  `json:"total_disbursed"`
	TotalPending     float64                  `json:"total_pending"`
	TotalCancelled   float64                  `json:"total_cancelled"`
	DisbursementRate float64                  `json:"disbursement_rate"`
	ProgramBreakdown []ProgramFinanceSummary  `json:"program_breakdown"`
}

type ProgramFinanceSummary struct {
	ProgramID      uint    `json:"program_id"`
	ProgramName    string  `json:"program_name"`
	TotalRequested float64 `json:"total_requested"`
	TotalApproved  float64 `json:"total_approved"`
	TotalDisbursed float64 `json:"total_disbursed"`
	GrantCount     int64   `json:"grant_count"`
}

type AuditSummaryResponse struct {
	TotalActions  int64         `json:"total_actions"`
	ActionsByType []ActionCount `json:"actions_by_type"`
	RecentActions []RecentAction `json:"recent_actions"`
}

type ActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

type RecentAction struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	CreatedAt  time.Time `json:"created_at"`
}
