package postgres

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

// defaultRoleOrder: fallback when the filter arrives without OrderBy.
// Built-in roles first, matching the default the DTO hands in.
const defaultRoleOrder = "roles.is_system DESC, roles.name ASC"

func (r *roleRepository) List(ctx context.Context, filter domain.RoleFilter) ([]domain.Role, int, error) {
	query := dbFrom(ctx, r.db).Model(&roleModel{})

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(roles.name) LIKE ? OR LOWER(roles.description) LIKE ?", keyword, keyword)
	}

	// Counted before the join: one row per role, not one per user.
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, wrapErr("count roles", err)
	}
	if total == 0 {
		return []domain.Role{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultRoleOrder
	}

	// LEFT JOIN so a role nobody uses still comes back, with user_count 0.
	// Soft-deleted users are excluded by hand; GORM's scope does not reach
	// into a join written as a string.
	var models []roleModel
	err := query.
		Select("roles.*, COUNT(users.id) AS user_count").
		Joins("LEFT JOIN users ON users.role_id = roles.id AND users.deleted_at IS NULL").
		Group("roles.id").
		Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&models).Error
	if err != nil {
		return nil, 0, wrapErr("list roles", err)
	}
	return rolesToDomain(models), int(total), nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var model roleModel
	if err := dbFrom(ctx, r.db).First(&model, "id = ?", id).Error; err != nil {
		return nil, wrapErr("get role", err)
	}
	return model.toDomain(), nil
}

// Create writes the whole aggregate: the role plus role.Permissions, in one
// transaction, so a rejected permission never leaves a half-built role behind.
func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	return dbFrom(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		model := roleFromDomain(role)
		if err := tx.Create(&model).Error; err != nil {
			return wrapErr("create role", err)
		}

		role.CreatedAt = model.CreatedAt
		role.UpdatedAt = model.UpdatedAt

		if len(role.Permissions) == 0 {
			return nil
		}

		permissions := make([]roleMenuPermissionModel, 0, len(role.Permissions))
		for i := range role.Permissions {
			role.Permissions[i].RoleID = role.ID
			permissions = append(permissions, roleMenuPermissionFromDomain(&role.Permissions[i]))
		}

		if err := tx.Create(&permissions).Error; err != nil {
			return wrapErr("create role permissions", err)
		}
		for i := range permissions {
			role.Permissions[i].CreatedAt = permissions[i].CreatedAt
			role.Permissions[i].UpdatedAt = permissions[i].UpdatedAt
		}
		return nil
	})
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role, permissions []domain.RoleMenuPermission) error {
	return dbFrom(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		model := roleFromDomain(role)
		if err := tx.Save(&model).Error; err != nil {
			return wrapErr("update role", err)
		}

		// Columns with database defaults come back via RETURNING.
		role.UpdatedAt = model.UpdatedAt

		if permissions == nil {
			return nil
		}

		// The request carries the full matrix, so stale rows have to go first.
		if err := tx.Where("role_id = ?", role.ID).Delete(&roleMenuPermissionModel{}).Error; err != nil {
			return wrapErr("clear role permissions", err)
		}
		if len(permissions) == 0 {
			role.Permissions = []domain.RoleMenuPermission{}
			return nil
		}

		models := make([]roleMenuPermissionModel, 0, len(permissions))
		for i := range permissions {
			permissions[i].RoleID = role.ID
			models = append(models, roleMenuPermissionFromDomain(&permissions[i]))
		}
		if err := tx.Create(&models).Error; err != nil {
			return wrapErr("create role permissions", err)
		}
		for i := range models {
			permissions[i].CreatedAt = models[i].CreatedAt
			permissions[i].UpdatedAt = models[i].UpdatedAt
		}
		role.Permissions = permissions
		return nil
	})
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := dbFrom(ctx, r.db).Delete(&roleModel{}, "id = ?", id).Error; err != nil {
		return wrapErr("delete role", err)
	}
	return nil
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.RoleMenuPermission, error) {
	var models []roleMenuPermissionModel
	err := dbFrom(ctx, r.db).
		Where("role_id = ?", roleID).
		Order("menu_id").
		Find(&models).Error
	if err != nil {
		return nil, wrapErr("get role permissions", err)
	}
	return roleMenuPermissionsToDomain(models), nil
}
