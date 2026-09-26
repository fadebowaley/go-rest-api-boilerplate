package permission

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type txKey struct{}

type Repository interface {
	Create(ctx context.Context, p *Permission) error
	FindByID(ctx context.Context, id uint) (*Permission, error)
	FindByName(ctx context.Context, name string) (*Permission, error)
	List(ctx context.Context, group string, page, perPage int) ([]Permission, int64, error)
	Delete(ctx context.Context, id uint) error

	AssignPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetPermissionsForRole(ctx context.Context, roleID uint) ([]Permission, error)
	GetPermissionsForRoles(ctx context.Context, roleNames []string) ([]Permission, error)
	GetPermissionsByUserID(ctx context.Context, userID uint) ([]string, error)
	GetRoleNameByID(ctx context.Context, roleID uint) (string, error)
	Transaction(ctx context.Context, fn func(context.Context) error) error
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

func (r *repository) Create(ctx context.Context, p *Permission) error {
	result := r.getDB(ctx).WithContext(ctx).Create(p)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id uint) (*Permission, error) {
	var p Permission
	result := r.getDB(ctx).WithContext(ctx).First(&p, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &p, nil
}

func (r *repository) FindByName(ctx context.Context, name string) (*Permission, error) {
	var p Permission
	result := r.getDB(ctx).WithContext(ctx).Where("name = ?", name).First(&p)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &p, nil
}

func (r *repository) List(ctx context.Context, group string, page, perPage int) ([]Permission, int64, error) {
	var permissions []Permission
	var total int64

	query := r.getDB(ctx).WithContext(ctx).Model(&Permission{})
	if group != "" {
		query = query.Where(`"group" = ?`, group)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Order("name ASC").Limit(perPage).Offset(offset).Find(&permissions).Error; err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.getDB(ctx).WithContext(ctx).Delete(&Permission{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) AssignPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
	tx := r.getDB(ctx).WithContext(ctx)

	if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
		return err
	}

	for _, permID := range permissionIDs {
		rp := &RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
			CreatedAt:    time.Now(),
		}
		if err := tx.Create(rp).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) GetPermissionsForRole(ctx context.Context, roleID uint) ([]Permission, error) {
	var permissions []Permission
	err := r.getDB(ctx).WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Order("permissions.name ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) GetPermissionsForRoles(ctx context.Context, roleNames []string) ([]Permission, error) {
	if len(roleNames) == 0 {
		return nil, nil
	}

	var permissions []Permission
	err := r.getDB(ctx).WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Where("roles.name IN ?", roleNames).
		Distinct("permissions.*").
		Order("permissions.name ASC").
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) GetPermissionsByUserID(ctx context.Context, userID uint) ([]string, error) {
	var perms []string
	err := r.getDB(ctx).WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.name").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Order("permissions.name ASC").
		Pluck("name", &perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *repository) GetRoleNameByID(ctx context.Context, roleID uint) (string, error) {
	var name string
	result := r.getDB(ctx).WithContext(ctx).Table("roles").Select("name").Where("id = ?", roleID).Take(&name)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", result.Error
	}
	return name, nil
}

func (r *repository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
