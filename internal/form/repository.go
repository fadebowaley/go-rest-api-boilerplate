package form

import (
	"context"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	Create(ctx context.Context, t *FormTemplate) error
	GetByID(ctx context.Context, id uint) (*FormTemplate, error)
	ListByTenant(ctx context.Context, tenantID uint) ([]FormTemplate, error)
	Update(ctx context.Context, t *FormTemplate) error
	Delete(ctx context.Context, id uint) error

	CreateVersion(ctx context.Context, v *FormVersion) error
	ListVersions(ctx context.Context, templateID uint) ([]FormVersion, error)
	GetVersionByID(ctx context.Context, id uint) (*FormVersion, error)
	GetLatestVersion(ctx context.Context, templateID uint) (*FormVersion, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, t *FormTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*FormTemplate, error) {
	var t FormTemplate
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, contextutil.GetTenantID(ctx)).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) ListByTenant(ctx context.Context, tenantID uint) ([]FormTemplate, error) {
	var templates []FormTemplate
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Find(&templates).Error
	return templates, err
}

func (r *repository) Update(ctx context.Context, t *FormTemplate) error {
	var existing FormTemplate
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", t.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&FormTemplate{}, id).Error
}

func (r *repository) CreateVersion(ctx context.Context, v *FormVersion) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *repository) ListVersions(ctx context.Context, templateID uint) ([]FormVersion, error) {
	var versions []FormVersion
	err := r.db.WithContext(ctx).
		Where("form_template_id = ?", templateID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

func (r *repository) GetVersionByID(ctx context.Context, id uint) (*FormVersion, error) {
	var v FormVersion
	err := r.db.WithContext(ctx).First(&v, id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *repository) GetLatestVersion(ctx context.Context, templateID uint) (*FormVersion, error) {
	var v FormVersion
	err := r.db.WithContext(ctx).
		Where("form_template_id = ?", templateID).
		Order("version_number DESC").
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}
