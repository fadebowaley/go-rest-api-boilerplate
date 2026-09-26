package disbursement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fadebowaley/applico/internal/contextutil"
)

var (
	ErrDisbursementNotFound  = errors.New("disbursement not found")
	ErrGrantNotApproved      = errors.New("grant is not approved")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrExceedsGrantAmount    = errors.New("total disbursement cannot exceed grant requested amount")
)

type GrantChecker interface {
	IsApproved(ctx context.Context, grantID uint) (bool, error)
	GetRequestedAmount(ctx context.Context, grantID uint) (float64, error)
}

type Service interface {
	Create(ctx context.Context, userID uint, req *CreateDisbursementRequest) (*DisbursementResponse, error)
	GetByID(ctx context.Context, id uint) (*DisbursementResponse, error)
	ListByGrant(ctx context.Context, grantID uint) (*[]DisbursementResponse, error)
	Update(ctx context.Context, id uint, req *UpdateDisbursementRequest) (*DisbursementResponse, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	repo         Repository
	grantChecker GrantChecker
}

func NewService(repo Repository, grantChecker GrantChecker) Service {
	return &service{repo: repo, grantChecker: grantChecker}
}

func (s *service) Create(ctx context.Context, userID uint, req *CreateDisbursementRequest) (*DisbursementResponse, error) {
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, fmt.Errorf("tenant context required")
	}

	approved, err := s.grantChecker.IsApproved(ctx, req.GrantApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check grant status: %w", err)
	}
	if !approved {
		return nil, ErrGrantNotApproved
	}

	requestedAmount, err := s.grantChecker.GetRequestedAmount(ctx, req.GrantApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grant requested amount: %w", err)
	}
	totalDisbursed, err := s.repo.GetTotalByGrant(ctx, req.GrantApplicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total disbursed: %w", err)
	}
	if totalDisbursed+req.Amount > requestedAmount {
		return nil, ErrExceedsGrantAmount
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	var paymentDate *time.Time
	if req.PaymentDate != "" {
		t, parseErr := time.Parse(time.RFC3339, req.PaymentDate)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid payment_date format: %w", parseErr)
		}
		paymentDate = &t
	}

	d := &Disbursement{
		TenantID:           tenantID,
		GrantApplicationID: req.GrantApplicationID,
		Amount:             req.Amount,
		Currency:           currency,
		Status:             StatusScheduled,
		PaymentDate:        paymentDate,
		Notes:              req.Notes,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to create disbursement: %w", err)
	}

	resp := ToDisbursementResponse(d)
	return &resp, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*DisbursementResponse, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find disbursement: %w", err)
	}
	if d == nil {
		return nil, ErrDisbursementNotFound
	}
	resp := ToDisbursementResponse(d)
	return &resp, nil
}

func (s *service) ListByGrant(ctx context.Context, grantID uint) (*[]DisbursementResponse, error) {
	list, err := s.repo.ListByGrant(ctx, grantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list disbursements: %w", err)
	}
	responses := make([]DisbursementResponse, len(list))
	for i, d := range list {
		responses[i] = ToDisbursementResponse(&d)
	}
	return &responses, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateDisbursementRequest) (*DisbursementResponse, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find disbursement: %w", err)
	}
	if d == nil {
		return nil, ErrDisbursementNotFound
	}

	if req.Amount != nil {
		d.Amount = *req.Amount
	}
	if req.Currency != nil {
		d.Currency = *req.Currency
	}
	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		d.Status = *req.Status
	}
	if req.PaymentDate != nil {
		if *req.PaymentDate == "" {
			d.PaymentDate = nil
		} else {
			t, parseErr := time.Parse(time.RFC3339, *req.PaymentDate)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid payment_date format: %w", parseErr)
			}
			d.PaymentDate = &t
		}
	}
	if req.EvidenceURL != nil {
		d.EvidenceURL = *req.EvidenceURL
	}
	if req.Notes != nil {
		d.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, d); err != nil {
		return nil, fmt.Errorf("failed to update disbursement: %w", err)
	}

	resp := ToDisbursementResponse(d)
	return &resp, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find disbursement: %w", err)
	}
	if d == nil {
		return ErrDisbursementNotFound
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
