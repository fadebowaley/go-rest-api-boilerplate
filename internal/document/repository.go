package document

import (
	"context"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type Repository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id uint) (*Document, error)
	ListByGrant(ctx context.Context, grantID uint) ([]Document, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, doc *Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Document, error) {
	var doc Document
	err := r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Where("id = ?", id).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *repository) ListByGrant(ctx context.Context, grantID uint) ([]Document, error) {
	var docs []Document
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", contextutil.GetTenantID(ctx)).
		Where("grant_application_id = ?", grantID).
		Order("created_at DESC").
		Find(&docs).Error
	return docs, err
}

func (r *repository) Update(ctx context.Context, doc *Document) error {
	var existing Document
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", doc.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&Document{}, id).Error
}
