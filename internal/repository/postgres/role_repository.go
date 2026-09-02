package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

// defaultRoleOrder: fallback without OrderBy, built-in roles first as the DTO defaults.
const defaultRoleOrder = "roles.is_system DESC, roles.name ASC"

func (r *roleRepository) List(ctx context.Context, filter domain.RoleFilter) ([]domain.Role, int, error) {
	query := dbFrom(ctx, r.db).Model(&model.RoleModel{})

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(roles.name) LIKE ? OR LOWER(roles.description) LIKE ?", keyword, keyword)
	}

	// Counted before the join: one row per role, not one per user.
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, roleErrors.wrap("count roles", err)
	}
	if total == 0 {
		return []domain.Role{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultRoleOrder
	}

	// LEFT JOIN keeps unused roles at user_count 0; deleted_at is filtered by hand here.
	var rows []model.RoleModel
	err := query.
		Select("roles.*, COUNT(users.id) AS user_count").
		Joins("LEFT JOIN users ON users.role_id = roles.id AND users.deleted_at IS NULL").
		Group("roles.id").
		Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&rows).Error
	if err != nil {
		return nil, 0, roleErrors.wrap("list roles", err)
	}
	return model.RolesToDomain(rows), int(total), nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var row model.RoleModel
	if err := dbFrom(ctx, r.db).First(&row, "id = ?", id).Error; err != nil {
		return nil, roleErrors.wrap("get role", err)
	}
	return row.ToDomain(), nil
}

// Create writes the role and its permissions in one transaction.
func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	actor := domain.ActorFrom(ctx)
	role.CreatedBy, role.UpdatedBy = actor, actor

	return dbFrom(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		row := model.RoleFromDomain(role)
		if err := tx.Create(&row).Error; err != nil {
			return roleErrors.wrap("create role", err)
		}

		role.CreatedAt = row.CreatedAt
		role.UpdatedAt = row.UpdatedAt

		if len(role.Permissions) == 0 {
			return nil
		}

		permissions := make([]model.RoleMenuPermissionModel, 0, len(role.Permissions))
		for i := range role.Permissions {
			role.Permissions[i].RoleID = role.ID
			permissions = append(permissions, model.RoleMenuPermissionFromDomain(&role.Permissions[i]))
		}

		if err := tx.Create(&permissions).Error; err != nil {
			return roleErrors.wrap("create role permissions", err)
		}
		for i := range permissions {
			role.Permissions[i].CreatedAt = permissions[i].CreatedAt
			role.Permissions[i].UpdatedAt = permissions[i].UpdatedAt
		}
		return nil
	})
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role, permissions []domain.RoleMenuPermission) error {
	// Save writes the whole row, so created_by survives only because GetByID read it back.
	role.UpdatedBy = domain.ActorFrom(ctx)

	return dbFrom(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		row := model.RoleFromDomain(role)
		if err := tx.Save(&row).Error; err != nil {
			return roleErrors.wrap("update role", err)
		}

		// Columns with database defaults come back via RETURNING.
		role.UpdatedAt = row.UpdatedAt

		if permissions == nil {
			return nil
		}

		// The request carries the full matrix, so stale rows have to go first.
		if err := tx.Where("role_id = ?", role.ID).Delete(&model.RoleMenuPermissionModel{}).Error; err != nil {
			return roleErrors.wrap("clear role permissions", err)
		}
		if len(permissions) == 0 {
			role.Permissions = []domain.RoleMenuPermission{}
			return nil
		}

		rows := make([]model.RoleMenuPermissionModel, 0, len(permissions))
		for i := range permissions {
			permissions[i].RoleID = role.ID
			rows = append(rows, model.RoleMenuPermissionFromDomain(&permissions[i]))
		}
		if err := tx.Create(&rows).Error; err != nil {
			return roleErrors.wrap("create role permissions", err)
		}
		for i := range rows {
			permissions[i].CreatedAt = rows[i].CreatedAt
			permissions[i].UpdatedAt = rows[i].UpdatedAt
		}
		role.Permissions = permissions
		return nil
	})
}

// Delete: soft delete stamped with who did it, in one statement. See userRepository.Delete.
func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := dbFrom(ctx, r.db).
		Model(&model.RoleModel{}).
		Where("id = ?", id).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		}).Error
	if err != nil {
		return roleErrors.wrap("delete role", err)
	}
	return nil
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.RoleMenuPermission, error) {
	var rows []model.RoleMenuPermissionModel
	err := dbFrom(ctx, r.db).
		Where("role_id = ?", roleID).
		Order("menu_id").
		Find(&rows).Error
	if err != nil {
		return nil, roleErrors.wrap("get role permissions", err)
	}
	return model.RoleMenuPermissionsToDomain(rows), nil
}
