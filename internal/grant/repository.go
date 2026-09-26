package grant

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/fadebowaley/applico/internal/contextutil"
)

type txKey struct{}

type ProgramRef struct {
	ID                 uint
	WorkflowTemplateID *uint
}

type Repository interface {
	Create(ctx context.Context, grant *GrantApplication) error
	FindByID(ctx context.Context, id uint) (*GrantApplication, error)
	Update(ctx context.Context, grant *GrantApplication) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, userID uint, filters GrantFilterParams, page, perPage int) ([]GrantApplication, int64, error)
	ListAll(ctx context.Context, filters GrantFilterParams, page, perPage int) ([]GrantApplication, int64, error)
	Transaction(ctx context.Context, fn func(context.Context) error) error
	FindProgramByID(ctx context.Context, id uint) (*ProgramRef, error)
	CreateVersion(ctx context.Context, v *GrantApplicationVersion) error
	ListVersionsByGrant(ctx context.Context, grantID uint) ([]GrantApplicationVersion, error)
	GetLatestVersion(ctx context.Context, grantID uint) (*GrantApplicationVersion, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *repository) Create(ctx context.Context, grant *GrantApplication) error {
	return r.getDB(ctx).WithContext(ctx).Create(grant).Error
}

func (r *repository) FindByID(ctx context.Context, id uint) (*GrantApplication, error) {
	var grant GrantApplication
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, nil
	}
	result := r.getDB(ctx).WithContext(ctx).Where("tenant_id = ?", tenantID).First(&grant, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &grant, nil
}

func (r *repository) Update(ctx context.Context, grant *GrantApplication) error {
	var existing GrantApplication
	if err := r.getDB(ctx).WithContext(ctx).Where("id = ? AND tenant_id = ?", grant.ID, contextutil.GetTenantID(ctx)).First(&existing).Error; err != nil {
		return err
	}
	return r.getDB(ctx).WithContext(ctx).Save(grant).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.getDB(ctx).WithContext(ctx).Where("tenant_id = ?", contextutil.GetTenantID(ctx)).Delete(&GrantApplication{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) List(ctx context.Context, userID uint, filters GrantFilterParams, page, perPage int) ([]GrantApplication, int64, error) {
	var grants []GrantApplication
	var total int64

	query := r.getDB(ctx).WithContext(ctx).Model(&GrantApplication{}).Where("user_id = ?", userID).Where("tenant_id = ?", contextutil.GetTenantID(ctx))
	query = applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	sortField := "created_at"
	if filters.Sort != "" {
		sortField = filters.Sort
	}
	order := "DESC"
	if filters.Order == "asc" {
		order = "ASC"
	}

	if err := query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&grants).Error; err != nil {
		return nil, 0, err
	}

	return grants, total, nil
}

func (r *repository) ListAll(ctx context.Context, filters GrantFilterParams, page, perPage int) ([]GrantApplication, int64, error) {
	var grants []GrantApplication
	var total int64

	query := r.getDB(ctx).WithContext(ctx).Model(&GrantApplication{}).Where("tenant_id = ?", contextutil.GetTenantID(ctx))
	query = applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	sortField := "created_at"
	if filters.Sort != "" {
		sortField = filters.Sort
	}
	order := "DESC"
	if filters.Order == "asc" {
		order = "ASC"
	}

	if err := query.Order(sortField + " " + order).Limit(perPage).Offset(offset).Find(&grants).Error; err != nil {
		return nil, 0, err
	}

	return grants, total, nil
}

func applyFilters(query *gorm.DB, filters GrantFilterParams) *gorm.DB {
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.ParishID != 0 {
		query = query.Where("parish_id = ?", filters.ParishID)
	}
	if filters.Region != "" {
		query = query.Where("region = ?", filters.Region)
	}
	if filters.Province != "" {
		query = query.Where("province = ?", filters.Province)
	}
	if filters.DateFrom != "" {
		query = query.Where("created_at >= ?", filters.DateFrom)
	}
	if filters.DateTo != "" {
		query = query.Where("created_at <= ?", filters.DateTo)
	}
	if filters.Search != "" {
		pattern := "%" + filters.Search + "%"
		query = query.Where("project_title ILIKE ? OR project_description ILIKE ?", pattern, pattern)
	}
	return query
}

func (r *repository) FindProgramByID(ctx context.Context, id uint) (*ProgramRef, error) {
	var ref ProgramRef
	tenantID := contextutil.GetTenantID(ctx)
	if tenantID == 0 {
		return nil, nil
	}
	result := r.db.WithContext(ctx).Model(&ProgramRef{}).Select("id, workflow_template_id").Table("grant_programs").
		Where("tenant_id = ?", tenantID).
		First(&ref, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &ref, nil
}

func (r *repository) CreateVersion(ctx context.Context, v *GrantApplicationVersion) error {
	return r.getDB(ctx).WithContext(ctx).Create(v).Error
}

func (r *repository) ListVersionsByGrant(ctx context.Context, grantID uint) ([]GrantApplicationVersion, error) {
	var versions []GrantApplicationVersion
	err := r.getDB(ctx).WithContext(ctx).
		Where("grant_application_id = ?", grantID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

func (r *repository) GetLatestVersion(ctx context.Context, grantID uint) (*GrantApplicationVersion, error) {
	var v GrantApplicationVersion
	err := r.getDB(ctx).WithContext(ctx).
		Where("grant_application_id = ?", grantID).
		Order("version_number DESC").
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *repository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
