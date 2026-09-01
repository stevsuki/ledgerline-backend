package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Role: a named set of menu permissions a user can be assigned to.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsSystem    bool // built-in role, must not be deleted or renamed
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Filled by the list query only; a single-role read leaves it zero.
	UserCount int
	// Written with the role on create; on read only a single-role query fills it.
	Permissions []RoleMenuPermission
}

// RoleMenuPermission: what one role may do on one menu, keyed by the role/menu pair.
type RoleMenuPermission struct {
	RoleID     uuid.UUID
	MenuID     uuid.UUID
	CanCreate  bool
	CanRead    bool
	CanUpdate  bool
	CanDelete  bool
	CanApprove bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type RoleFilter struct {
	Search  string
	Limit   int
	Offset  int
	OrderBy string // ORDER BY clause, may only be filled via pagination.Sortable
}

// RoleRepository: port to storage for the roles table.
type RoleRepository interface {
	List(ctx context.Context, filter RoleFilter) ([]Role, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	Create(ctx context.Context, role *Role) error
	// Update replaces the permission set when permissions is non-nil.
	Update(ctx context.Context, role *Role, permissions []RoleMenuPermission) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]RoleMenuPermission, error)
}

type CreateRoleInput struct {
	Name        string
	Description string
	Permissions []CreateRoleMenuPermissionInput
}

type CreateRoleMenuPermissionInput struct {
	MenuID     uuid.UUID
	CanCreate  bool
	CanRead    bool
	CanUpdate  bool
	CanDelete  bool
	CanApprove bool
}

type UpdateRoleInput struct {
	Name        *string
	Description *string
	// nil leaves the permissions untouched; an empty slice clears them.
	Permissions *[]CreateRoleMenuPermissionInput
}

type RoleService interface {
	List(ctx context.Context, filter RoleFilter) ([]Role, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	Create(ctx context.Context, input CreateRoleInput) (*Role, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
