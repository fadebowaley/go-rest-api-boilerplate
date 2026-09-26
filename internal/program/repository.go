package program

import (
	"context"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	Create(ctx context.Context, p *GrantProgram) error
	GetByID(ctx context.Context, id uint) (*GrantProgram, error)
	ListByTenant(ctx context.Context, tenantID uint) ([]GrantProgram, error)
	Update(ctx context.Context, p *GrantProgram) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, p *GrantProgram) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*GrantProgram, error) {
	var p GrantProgram
	err := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *repository) ListByTenant(ctx context.Context, tenantID uint) ([]GrantProgram, error) {
	var programs []GrantProgram
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&programs).Error
	return programs, err
}

func (r *repository) Update(ctx context.Context, p *GrantProgram) error {
	var existing GrantProgram
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", p.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&GrantProgram{}, id).Error
}
