package domain

import (
	"context"

	"github.com/google/uuid"
)

// Menu: one entry of the sidebar. Groups carry children, pages carry access.
type Menu struct {
	ID        uuid.UUID
	ParentID  *uuid.UUID
	Code      string // stable identifier permission checks rely on
	Name      string
	Path      string
	Icon      string
	SortOrder int
	Access    MenuAccess
	Children  []Menu
}

// MenuAccess: what one role may do on one menu; all false when no row exists.
type MenuAccess struct {
	CanCreate  bool
	CanRead    bool
	CanUpdate  bool
	CanDelete  bool
	CanApprove bool
}

// MenuRepository: port to storage for the menus table.
type MenuRepository interface {
	// ListByRole returns every active menu, flat, with one role's access flags.
	ListByRole(ctx context.Context, roleID uuid.UUID) ([]Menu, error)
}
