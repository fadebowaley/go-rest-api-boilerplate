package stats

import (
	"context"
	"fmt"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Service interface {
	Summary(ctx context.Context) (*SummaryResponse, error)
	ProgramStats(ctx context.Context, programID uint) (*ProgramStat, error)
	WorkflowStats(ctx context.Context) (*WorkflowStatsResponse, error)
	FinanceSummary(ctx context.Context) (*FinanceSummaryResponse, error)
	AuditSummary(ctx context.Context) (*AuditSummaryResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Summary(ctx context.Context) (*SummaryResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}
	return s.repo.Summary(ctx, tenantID)
}

func (s *service) ProgramStats(ctx context.Context, programID uint) (*ProgramStat, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}
	stat, err := s.repo.ProgramStats(ctx, tenantID, programID)
	if err != nil {
		return nil, err
	}
	if stat == nil {
		return nil, fmt.Errorf("program not found")
	}
	return stat, nil
}

func (s *service) WorkflowStats(ctx context.Context) (*WorkflowStatsResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}
	return s.repo.WorkflowStats(ctx, tenantID)
}

func (s *service) FinanceSummary(ctx context.Context) (*FinanceSummaryResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}
	return s.repo.FinanceSummary(ctx, tenantID)
}

func (s *service) AuditSummary(ctx context.Context) (*AuditSummaryResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}
	return s.repo.AuditSummary(ctx, tenantID)
}
