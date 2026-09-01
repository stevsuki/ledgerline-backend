package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

// CreateRoleRequest: create role payload (admin only).
// Lengths mirror the column sizes in migration 000006.
type CreateRoleRequest struct {
	Name        string                            `json:"name" binding:"required,min=2,max=50" example:"Finance Staff"`
	Description string                            `json:"description" binding:"omitempty,max=255" example:"Handles day to day transactions"`
	Permissions []CreateRoleMenuPermissionRequest `json:"permissions"`
}

type CreateRoleMenuPermissionRequest struct {
	MenuID     uuid.UUID `json:"menu_id" binding:"required" example:"b0000000-0000-0000-0000-000000000002"`
	CanCreate  bool      `json:"can_create" example:"true"`
	CanRead    bool      `json:"can_read" example:"true"`
	CanUpdate  bool      `json:"can_update" example:"true"`
	CanDelete  bool      `json:"can_delete" example:"true"`
	CanApprove bool      `json:"can_approve" example:"true"`
}

func (r CreateRoleRequest) ToInput() domain.CreateRoleInput {
	permissions := make([]domain.CreateRoleMenuPermissionRequest, 0, len(r.Permissions))
	for i := range r.Permissions {
		p := r.Permissions[i]
		permissions = append(permissions, domain.CreateRoleMenuPermissionRequest{
			MenuID:     p.MenuID,
			CanCreate:  p.CanCreate,
			CanRead:    p.CanRead,
			CanUpdate:  p.CanUpdate,
			CanDelete:  p.CanDelete,
			CanApprove: p.CanApprove,
		})
	}

	return domain.CreateRoleInput{
		Name:        r.Name,
		Description: r.Description,
		Permissions: permissions,
	}
}

// UpdateRoleRequest: pointers so partial updates are detectable.
type UpdateRoleRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=50" example:"Finance Lead"`
	Description *string `json:"description" binding:"omitempty,max=255" example:"Approves transactions above the limit"`
	// Omitted leaves the permissions untouched; an empty array clears them.
	Permissions *[]CreateRoleMenuPermissionRequest `json:"permissions"`
}

func (r UpdateRoleRequest) ToInput() domain.UpdateRoleInput {
	input := domain.UpdateRoleInput{Name: r.Name, Description: r.Description}
	if r.Permissions == nil {
		return input
	}

	permissions := make([]domain.CreateRoleMenuPermissionRequest, 0, len(*r.Permissions))
	for _, p := range *r.Permissions {
		permissions = append(permissions, domain.CreateRoleMenuPermissionRequest{
			MenuID:     p.MenuID,
			CanCreate:  p.CanCreate,
			CanRead:    p.CanRead,
			CanUpdate:  p.CanUpdate,
			CanDelete:  p.CanDelete,
			CanApprove: p.CanApprove,
		})
	}
	input.Permissions = &permissions
	return input
}

// roleSort: only columns in this map may enter ORDER BY.
// Qualified: the list query joins users, which owns columns of the same name.
var roleSort = pagination.Sortable{
	Allowed: pagination.Whitelist{
		"name":       "roles.name",
		"is_system":  "roles.is_system",
		"created_at": "roles.created_at",
		"updated_at": "roles.updated_at",
		"user_count": "user_count",
	},
	// Built-in roles first keeps Admin and User at the top of the first page.
	Default:    "-is_system,name",
	TieBreaker: "roles.id",
}

// ListRoleQuery: query params for the role list.
type ListRoleQuery struct {
	Search  string `form:"search" example:"finance"`
	Sort    string `form:"sort" binding:"omitempty,max=100" example:"name,-created_at"`
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100" example:"10"`
}

// OrderBy: raw sort param -> whitelisted SQL clause.
func (q ListRoleQuery) OrderBy() (string, error) { return roleSort.OrderBy(q.Sort) }

// RoleResponse: the role shape that is safe to send to clients.
type RoleResponse struct {
	ID          uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000001"`
	Name        string    `json:"name" example:"Admin"`
	Description string    `json:"description" example:"Built-in role with access to every menu"`
	// Built-in roles must not be renamed or deleted; the UI hides those actions.
	IsSystem  bool      `json:"is_system" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-02T15:04:05Z"`
	// How many users are assigned to this role; filled by the list only.
	UserCount   int                          `json:"user_count" example:"12"`
	Permissions []RoleMenuPermissionResponse `json:"permissions"`
}

type RoleMenuPermissionResponse struct {
	// RoleID     uuid.UUID `json:"role_id" example:"00000000-0000-0000-0000-000000000001"`
	MenuID     uuid.UUID `json:"menu_id" example:"b0000000-0000-0000-0000-000000000002"`
	CanCreate  bool      `json:"can_create" example:"true"`
	CanRead    bool      `json:"can_read" example:"true"`
	CanUpdate  bool      `json:"can_update" example:"true"`
	CanDelete  bool      `json:"can_delete" example:"true"`
	CanApprove bool      `json:"can_approve" example:"true"`
	// CreatedAt  time.Time `json:"created_at" example:"2026-01-02T15:04:05Z"`
	// UpdatedAt  time.Time `json:"updated_at" example:"2026-01-02T15:04:05Z"`
}

func NewRoleMenuPermissionResponse(p domain.RoleMenuPermission) RoleMenuPermissionResponse {
	return RoleMenuPermissionResponse{
		// RoleID:     p.RoleID,
		MenuID:     p.MenuID,
		CanCreate:  p.CanCreate,
		CanRead:    p.CanRead,
		CanUpdate:  p.CanUpdate,
		CanDelete:  p.CanDelete,
		CanApprove: p.CanApprove,
		// CreatedAt:  p.CreatedAt,
		// UpdatedAt:  p.UpdatedAt,
	}
}

func NewRoleMenuPermissionResponses(permissions []domain.RoleMenuPermission) []RoleMenuPermissionResponse {
	out := make([]RoleMenuPermissionResponse, 0, len(permissions))
	for i := range permissions {
		out = append(out, NewRoleMenuPermissionResponse(permissions[i]))
	}
	return out
}

func NewRoleResponse(r *domain.Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		UserCount:   r.UserCount,
		Permissions: NewRoleMenuPermissionResponses(r.Permissions),
	}
}

func NewRoleResponses(roles []domain.Role) []RoleResponse {
	out := make([]RoleResponse, 0, len(roles))
	for i := range roles {
		out = append(out, NewRoleResponse(&roles[i]))
	}
	return out
}
