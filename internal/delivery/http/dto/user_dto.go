// Package dto: request/response models specific to the HTTP transport.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

// CreateUserRequestDTO: create user payload (admin only).
type CreateUserRequestDTO struct {
	Email    string `json:"email" binding:"required,email" example:"budi@example.com"`
	FullName string `json:"full_name" binding:"required,min=3,max=100" example:"Budi Santoso"`
	Password string `json:"password" binding:"required,min=8,max=72" example:"Rahasia123!"`
	// Omitted means the built-in User role; an unknown id is rejected as 400.
	RoleID uuid.UUID `json:"role_id" binding:"omitempty" example:"00000000-0000-0000-0000-000000000002"`
}

func (r CreateUserRequestDTO) ToInput() domain.CreateUserInput {
	return domain.CreateUserInput{
		Email:    r.Email,
		FullName: r.FullName,
		Password: r.Password,
		RoleID:   r.RoleID,
	}
}

// UpdateUserRequestDTO: pointers so partial updates are detectable.
type UpdateUserRequestDTO struct {
	FullName *string    `json:"full_name" binding:"omitempty,min=3,max=100" example:"Budi Santoso"`
	RoleID   *uuid.UUID `json:"role_id" binding:"omitempty" example:"00000000-0000-0000-0000-000000000001"`
}

func (r UpdateUserRequestDTO) ToInput() domain.UpdateUserInput {
	return domain.UpdateUserInput{FullName: r.FullName, RoleID: r.RoleID}
}

// userSort: only columns in this map may enter ORDER BY.
var userSort = pagination.Sortable{
	// Qualified: the list query joins roles, which owns columns of the same name.
	Allowed: pagination.Whitelist{
		"email":      "users.email",
		"full_name":  "users.full_name",
		"role":       "roles.name",
		"created_at": "users.created_at",
		"updated_at": "users.updated_at",
	},
	Default:    "-created_at",
	TieBreaker: "users.id",
}

// ListUserQueryDTO: query params for the user list.
type ListUserQueryDTO struct {
	Search  string `form:"search" example:"budi"`
	Sort    string `form:"sort" binding:"omitempty,max=100" example:"-created_at,full_name"`
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100" example:"10"`
}

// OrderBy: raw sort param -> whitelisted SQL clause.
func (q ListUserQueryDTO) OrderBy() (string, error) { return userSort.OrderBy(q.Sort) }

// UserResponseDTO: the user shape that is safe to send to clients.
type UserResponseDTO struct {
	ID        uuid.UUID `json:"id" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
	Email     string    `json:"email" example:"budi@example.com"`
	FullName  string    `json:"full_name" example:"Budi Santoso"`
	RoleID    uuid.UUID `json:"role_id" example:"00000000-0000-0000-0000-000000000002"`
	Role      string    `json:"role" example:"User"`
	Status    string    `json:"status" example:"enabled"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-02T15:04:05Z"`
}

func NewUserResponseDTO(u *domain.User) UserResponseDTO {
	return UserResponseDTO{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		RoleID:    u.RoleID,
		Role:      u.RoleName,
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func NewUserResponseDTOs(users []domain.User) []UserResponseDTO {
	out := make([]UserResponseDTO, 0, len(users))
	for i := range users {
		out = append(out, NewUserResponseDTO(&users[i]))
	}
	return out
}
