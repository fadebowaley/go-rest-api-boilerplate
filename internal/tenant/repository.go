package tenant

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uint) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	List(ctx context.Context) ([]Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uint) error

	AddUser(ctx context.Context, tu *TenantUser) error
	RemoveUser(ctx context.Context, tenantID, userID uint) error
	GetUsers(ctx context.Context, tenantID uint) ([]TenantUser, error)
	GetUserTenants(ctx context.Context, userID uint) ([]TenantUser, error)
	GetUserTenant(ctx context.Context, userID uint, tenantID uint) (*TenantUser, error)

	CreateTenantRole(ctx context.Context, role *TenantRole) error
	GetTenantRoleByID(ctx context.Context, id uint) (*TenantRole, error)
	ListTenantRoles(ctx context.Context, tenantID uint) ([]TenantRole, error)
	UpdateTenantRole(ctx context.Context, role *TenantRole) error
	DeleteTenantRole(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tenant *Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Tenant, error) {
	var t Tenant
	err := r.db.WithContext(ctx).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*Tenant, error) {
	var t Tenant
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) List(ctx context.Context) ([]Tenant, error) {
	var tenants []Tenant
	err := r.db.WithContext(ctx).Order("name ASC").Find(&tenants).Error
	return tenants, err
}

func (r *repository) Update(ctx context.Context, tenant *Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Tenant{}, id).Error
}

func (r *repository) AddUser(ctx context.Context, tu *TenantUser) error {
	return r.db.WithContext(ctx).Create(tu).Error
}

func (r *repository) RemoveUser(ctx context.Context, tenantID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Delete(&TenantUser{}).Error
}

func (r *repository) GetUsers(ctx context.Context, tenantID uint) ([]TenantUser, error) {
	var users []TenantUser
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

func (r *repository) GetUserTenants(ctx context.Context, userID uint) ([]TenantUser, error) {
	var users []TenantUser
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&users).Error
	return users, err
}

func (r *repository) CreateTenantRole(ctx context.Context, role *TenantRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *repository) GetTenantRoleByID(ctx context.Context, id uint) (*TenantRole, error) {
	var role TenantRole
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) ListTenantRoles(ctx context.Context, tenantID uint) ([]TenantRole, error) {
	var roles []TenantRole
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name ASC").Find(&roles).Error
	return roles, err
}

func (r *repository) UpdateTenantRole(ctx context.Context, role *TenantRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *repository) DeleteTenantRole(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&TenantRole{}, id).Error
}

func (r *repository) GetUserTenant(ctx context.Context, userID uint, tenantID uint) (*TenantUser, error) {
	var tu TenantUser
	err := r.db.WithContext(ctx).Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&tu).Error
	if err != nil {
		return nil, err
	}
	return &tu, nil
}
