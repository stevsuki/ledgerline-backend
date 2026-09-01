package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) domain.MenuRepository { return &menuRepository{db: db} }

// ListByRole: every active menu with the flags of one role. LEFT JOIN so menus
// the role has no row for still come back, with every flag false.
func (r *menuRepository) ListByRole(ctx context.Context, roleID uuid.UUID) ([]domain.Menu, error) {
	var models []menuModel
	err := dbFrom(ctx, r.db).
		Model(&menuModel{}).
		Select(`menus.*,
			COALESCE(p.can_create, FALSE)  AS can_create,
			COALESCE(p.can_read, FALSE)    AS can_read,
			COALESCE(p.can_update, FALSE)  AS can_update,
			COALESCE(p.can_delete, FALSE)  AS can_delete,
			COALESCE(p.can_approve, FALSE) AS can_approve`).
		Joins("LEFT JOIN role_menu_permissions p ON p.menu_id = menus.id AND p.role_id = ?", roleID).
		Where("menus.is_active").
		Order("menus.sort_order").
		Find(&models).Error
	if err != nil {
		return nil, wrapErr("list menus by role", err)
	}
	return menusToDomain(models), nil
}
