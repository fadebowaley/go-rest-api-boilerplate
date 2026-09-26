package audit

import (
	"context"
	"time"

	"github.com/fadebowaley/applico/internal/contextutil"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filters FilterParams, page, perPage int) ([]AuditLog, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) List(ctx context.Context, filters FilterParams, page, perPage int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditLog{})

	if tenantID := contextutil.GetTenantID(ctx); tenantID != 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if filters.UserID != 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.EntityType != "" {
		query = query.Where("entity_type = ?", filters.EntityType)
	}
	if filters.EntityID != "" {
		query = query.Where("entity_id = ?", filters.EntityID)
	}
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if !filters.Since.IsZero() {
		query = query.Where("created_at >= ?", filters.Since)
	}
	if !filters.Until.IsZero() {
		query = query.Where("created_at <= ?", filters.Until)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

type FilterParams struct {
	UserID     uint
	EntityType string
	EntityID   string
	Action     string
	Since      time.Time
	Until      time.Time
}
