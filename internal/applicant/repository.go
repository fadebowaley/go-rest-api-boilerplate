package applicant

import (
	"context"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	Create(ctx context.Context, a *Applicant) error
	GetByID(ctx context.Context, id uint) (*Applicant, error)
	ListByTenant(ctx context.Context, tenantID uint) ([]Applicant, error)
	Update(ctx context.Context, a *Applicant) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, a *Applicant) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Applicant, error) {
	var a Applicant
	err := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) ListByTenant(ctx context.Context, tenantID uint) ([]Applicant, error) {
	var applicants []Applicant
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&applicants).Error
	return applicants, err
}

func (r *repository) Update(ctx context.Context, a *Applicant) error {
	var existing Applicant
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", a.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&Applicant{}, id).Error
}
