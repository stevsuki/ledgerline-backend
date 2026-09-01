package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

// CreateRoleRequestDTO: create role payload (admin only), lengths from migration 000006.
type CreateRoleRequestDTO struct {
	Name        string                               `json:"name" binding:"required,min=2,max=50" example:"Finance Staff"`
	Description string                               `json:"description" binding:"omitempty,max=255" example:"Handles day to day transactions"`
	Permissions []CreateRoleMenuPermissionRequestDTO `json:"permissions"`
}

type CreateRoleMenuPermissionRequestDTO struct {
	MenuID     uuid.UUID `json:"menu_id" binding:"required" example:"b0000000-0000-0000-0000-000000000002"`
	CanCreate  bool      `json:"can_create" example:"true"`
	CanRead    bool      `json:"can_read" example:"true"`
	CanUpdate  bool      `json:"can_update" example:"true"`
	CanDelete  bool      `json:"can_delete" example:"true"`
	CanApprove bool      `json:"can_approve" example:"true"`
}

func (r CreateRoleRequestDTO) ToInput() domain.CreateRoleInput {
	permissions := make([]domain.CreateRoleMenuPermissionInput, 0, len(r.Permissions))
	for i := range r.Permissions {
		p := r.Permissions[i]
		permissions = append(permissions, domain.CreateRoleMenuPermissionInput{
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

// UpdateRoleRequestDTO: pointers so partial updates are detectable.
type UpdateRoleRequestDTO struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=50" example:"Finance Lead"`
	Description *string `json:"description" binding:"omitempty,max=255" example:"Approves transactions above the limit"`
	// Omitted leaves the permissions untouched; an empty array clears them.
	Permissions *[]CreateRoleMenuPermissionRequestDTO `json:"permissions"`
}

func (r UpdateRoleRequestDTO) ToInput() domain.UpdateRoleInput {
	input := domain.UpdateRoleInput{Name: r.Name, Description: r.Description}
	if r.Permissions == nil {
		return input
	}

	permissions := make([]domain.CreateRoleMenuPermissionInput, 0, len(*r.Permissions))
	for _, p := range *r.Permissions {
		permissions = append(permissions, domain.CreateRoleMenuPermissionInput{
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

// roleSort: only these columns may enter ORDER BY, qualified because the list joins users.
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

// ListRoleQueryDTO: query params for the role list.
type ListRoleQueryDTO struct {
	Search  string `form:"search" example:"finance"`
	Sort    string `form:"sort" binding:"omitempty,max=100" example:"name,-created_at"`
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100" example:"10"`
}

// OrderBy: raw sort param -> whitelisted SQL clause.
func (q ListRoleQueryDTO) OrderBy() (string, error) { return roleSort.OrderBy(q.Sort) }

// RoleResponseDTO: the role shape that is safe to send to clients.
type RoleResponseDTO struct {
	ID          uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000001"`
	Name        string    `json:"name" example:"Admin"`
	Description string    `json:"description" example:"Built-in role with access to every menu"`
	// Built-in roles must not be renamed or deleted; the UI hides those actions.
	IsSystem  bool      `json:"is_system" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-02T15:04:05Z"`
	// How many users are assigned to this role; filled by the list only.
	UserCount   int                             `json:"user_count" example:"12"`
	Permissions []RoleMenuPermissionResponseDTO `json:"permissions"`
}

type RoleMenuPermissionResponseDTO struct {
	MenuID     uuid.UUID `json:"menu_id" example:"b0000000-0000-0000-0000-000000000002"`
	CanCreate  bool      `json:"can_create" example:"true"`
	CanRead    bool      `json:"can_read" example:"true"`
	CanUpdate  bool      `json:"can_update" example:"true"`
	CanDelete  bool      `json:"can_delete" example:"true"`
	CanApprove bool      `json:"can_approve" example:"true"`
}

func NewRoleMenuPermissionResponseDTO(p domain.RoleMenuPermission) RoleMenuPermissionResponseDTO {
	return RoleMenuPermissionResponseDTO{
		MenuID:     p.MenuID,
		CanCreate:  p.CanCreate,
		CanRead:    p.CanRead,
		CanUpdate:  p.CanUpdate,
		CanDelete:  p.CanDelete,
		CanApprove: p.CanApprove,
	}
}

func NewRoleMenuPermissionResponseDTOs(permissions []domain.RoleMenuPermission) []RoleMenuPermissionResponseDTO {
	out := make([]RoleMenuPermissionResponseDTO, 0, len(permissions))
	for i := range permissions {
		out = append(out, NewRoleMenuPermissionResponseDTO(permissions[i]))
	}
	return out
}

func NewRoleResponseDTO(r *domain.Role) RoleResponseDTO {
	return RoleResponseDTO{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		UserCount:   r.UserCount,
		Permissions: NewRoleMenuPermissionResponseDTOs(r.Permissions),
	}
}

func NewRoleResponseDTOs(roles []domain.Role) []RoleResponseDTO {
	out := make([]RoleResponseDTO, 0, len(roles))
	for i := range roles {
		out = append(out, NewRoleResponseDTO(&roles[i]))
	}
	return out
}
