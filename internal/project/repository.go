package project

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	Create(ctx context.Context, u *ProjectUpdate) error
	GetByID(ctx context.Context, id uint) (*ProjectUpdate, error)
	ListByGrant(ctx context.Context, grantID uint) ([]ProjectUpdate, error)
	Update(ctx context.Context, u *ProjectUpdate) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, u *ProjectUpdate) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*ProjectUpdate, error) {
	var u ProjectUpdate
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, nil
	}
	result := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&u, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &u, nil
}

func (r *repository) ListByGrant(ctx context.Context, grantID uint) ([]ProjectUpdate, error) {
	var list []ProjectUpdate
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return list, nil
	}
	result := r.db.WithContext(ctx).
		Where("grant_application_id = ?", grantID).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&list)
	return list, result.Error
}

func (r *repository) Update(ctx context.Context, u *ProjectUpdate) error {
	var existing ProjectUpdate
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", u.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&ProjectUpdate{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
