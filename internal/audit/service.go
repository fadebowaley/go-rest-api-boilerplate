package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	ActionLogin               = "login"
	ActionLogout              = "logout"
	ActionCreated             = "created"
	ActionUpdated             = "updated"
	ActionDeleted             = "deleted"
	ActionSubmitted           = "submitted"
	ActionApproved            = "approved"
	ActionRejected            = "rejected"
	ActionReturned            = "returned"
	ActionEscalated           = "escalated"
	ActionDisbursed           = "disbursed"
	ActionUploaded            = "uploaded"
	ActionVerified            = "verified"
	ActionRoleChanged         = "role_changed"
	ActionPasswordReset       = "password_reset"
)

type Service interface {
	Log(ctx context.Context, entry *CreateAuditLogRequest) error
	LogAction(ctx context.Context, userID uint, entityType, entityID, action string, oldValue, newValue interface{}, ipAddress, userAgent string) error
	List(ctx context.Context, filters FilterParams, page, perPage int) ([]AuditLog, int64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Log(ctx context.Context, entry *CreateAuditLogRequest) error {
	oldStr, err := toJSONString(entry.OldValue)
	if err != nil {
		return fmt.Errorf("failed to marshal old value: %w", err)
	}
	newStr, err := toJSONString(entry.NewValue)
	if err != nil {
		return fmt.Errorf("failed to marshal new value: %w", err)
	}

	log := &AuditLog{
		UserID:     entry.UserID,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		Action:     entry.Action,
		OldValue:   oldStr,
		NewValue:   newStr,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  time.Now().UTC(),
	}

	return s.repo.Create(ctx, log)
}

func (s *service) LogAction(ctx context.Context, userID uint, entityType, entityID, action string, oldValue, newValue interface{}, ipAddress, userAgent string) error {
	entry := &CreateAuditLogRequest{
		UserID:     userID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		OldValue:   oldValue,
		NewValue:   newValue,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	return s.Log(ctx, entry)
}

func (s *service) List(ctx context.Context, filters FilterParams, page, perPage int) ([]AuditLog, int64, error) {
	if page < 1 {
		return nil, 0, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 {
		return nil, 0, fmt.Errorf("perPage must be >= 1")
	}
	if perPage > 100 {
		return nil, 0, fmt.Errorf("perPage must be <= 100")
	}

	return s.repo.List(ctx, filters, page, perPage)
}

type CreateAuditLogRequest struct {
	UserID     uint        `json:"user_id"`
	EntityType string      `json:"entity_type" binding:"required"`
	EntityID   string      `json:"entity_id" binding:"required"`
	Action     string      `json:"action" binding:"required"`
	OldValue   interface{} `json:"old_value"`
	NewValue   interface{} `json:"new_value"`
	IPAddress  string      `json:"ip_address"`
	UserAgent  string      `json:"user_agent"`
}

func toJSONString(v interface{}) (*string, error) {
	if v == nil {
		return nil, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	str := string(data)
	return &str, nil
}
