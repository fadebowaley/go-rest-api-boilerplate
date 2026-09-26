package disbursement

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	Create(ctx context.Context, d *Disbursement) error
	GetByID(ctx context.Context, id uint) (*Disbursement, error)
	ListByGrant(ctx context.Context, grantID uint) ([]Disbursement, error)
	Update(ctx context.Context, d *Disbursement) error
	Delete(ctx context.Context, id uint) error
	GetTotalByGrant(ctx context.Context, grantID uint) (float64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, d *Disbursement) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Disbursement, error) {
	var d Disbursement
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, nil
	}
	result := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&d, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &d, nil
}

func (r *repository) ListByGrant(ctx context.Context, grantID uint) ([]Disbursement, error) {
	var list []Disbursement
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

func (r *repository) Update(ctx context.Context, d *Disbursement) error {
	var existing Disbursement
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", d.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&Disbursement{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) GetTotalByGrant(ctx context.Context, grantID uint) (float64, error) {
	var total float64
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return 0, nil
	}
	err := r.db.WithContext(ctx).
		Model(&Disbursement{}).
		Where("grant_application_id = ? AND tenant_id = ?", grantID, tenantID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
