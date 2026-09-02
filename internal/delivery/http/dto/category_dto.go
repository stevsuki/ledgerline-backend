package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

type CreateCategoryRequestDTO struct {
	Name string `json:"name" binding:"required,min=2,max=100" example:"Gaji"`
	Type string `json:"type" binding:"required,oneof=income expense" example:"income"`
}

func (r CreateCategoryRequestDTO) ToInput() domain.CreateCategoryInput {
	return domain.CreateCategoryInput{Name: r.Name, Type: domain.CategoryType(r.Type)}
}

type UpdateCategoryRequestDTO struct {
	Name *string `json:"name" binding:"omitempty,min=2,max=100" example:"Gaji Bulanan"`
	Type *string `json:"type" binding:"omitempty,oneof=income expense" example:"expense"`
}

func (r UpdateCategoryRequestDTO) ToInput() domain.UpdateCategoryInput {
	input := domain.UpdateCategoryInput{Name: r.Name}
	if r.Type != nil {
		t := domain.CategoryType(*r.Type)
		input.Type = &t
	}
	return input
}

// categorySort: only columns in this map may enter ORDER BY.
var categorySort = pagination.Sortable{
	Allowed: pagination.Whitelist{
		"name":       "name",
		"type":       "type",
		"created_at": "created_at",
		"updated_at": "updated_at",
	},
	Default:    "-created_at",
	TieBreaker: "id",
}

// ListCategoryQueryDTO: query params for the category list.
type ListCategoryQueryDTO struct {
	Search  string `form:"search" example:"gaji"`
	Type    string `form:"type" binding:"omitempty,oneof=income expense" example:"income"`
	Sort    string `form:"sort" binding:"omitempty,max=100" example:"-created_at,name"`
	Page    int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100" example:"10"`
}

// OrderBy: raw sort param -> whitelisted SQL clause.
func (q ListCategoryQueryDTO) OrderBy() (string, error) { return categorySort.OrderBy(q.Sort) }

type CategoryResponseDTO struct {
	ID        uuid.UUID `json:"id" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
	Name      string    `json:"name" example:"Gaji"`
	Type      string    `json:"type" example:"income"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-02T15:04:05Z"`
	// deleted_by is left out: a deleted category is never in a response.
	CreatedBy *uuid.UUID `json:"created_by" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
	UpdatedBy *uuid.UUID `json:"updated_by" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
}

func NewCategoryResponseDTO(c *domain.Category) CategoryResponseDTO {
	return CategoryResponseDTO{
		ID:        c.ID,
		Name:      c.Name,
		Type:      string(c.Type),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
}

func NewCategoryResponseDTOs(categories []domain.Category) []CategoryResponseDTO {
	out := make([]CategoryResponseDTO, 0, len(categories))
	for i := range categories {
		out = append(out, NewCategoryResponseDTO(&categories[i]))
	}
	return out
}
