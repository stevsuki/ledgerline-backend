package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type CreateCategoryRequestDTO struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon" binding:"omitempty,max=50" example:"cup"`
	Color string `json:"color" binding:"omitempty,max=10" example:"c2"`
}

func (c CreateCategoryRequestDTO) ToInput() domain.CreateCategoryInput {
	return domain.CreateCategoryInput{
		Name:  c.Name,
		Type:  c.Type,
		Icon:  c.Icon,
		Color: c.Color,
	}
}

// UpdateCategoryRequestDTO: pointers so partial updates are detectable.
type UpdateCategoryRequestDTO struct {
	Name  *string `json:"name"`
	Type  *string `json:"type"`
	Icon  *string `json:"icon" binding:"omitempty,max=50" example:"cup"`
	Color *string `json:"color" binding:"omitempty,max=10" example:"c2"`
}

func (c UpdateCategoryRequestDTO) ToInput() domain.UpdateCategoryInput {
	return domain.UpdateCategoryInput{
		Name:  c.Name,
		Type:  c.Type,
		Icon:  c.Icon,
		Color: c.Color,
	}
}

type CategoryResponseDTO struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	// IsOwn false marks a shared row the account may use but does not hold yet.
	IsOwn            bool       `json:"is_own"`
	// IsBuiltIn: this name and direction are one of the shared master rows.
	IsBuiltIn        bool       `json:"is_built_in"`
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Icon             string     `json:"icon" example:"cup"`
	Color            string     `json:"color" example:"c2"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CreatedBy        *uuid.UUID `json:"created_by"`
	UpdatedBy        *uuid.UUID `json:"updated_by"`
	DeletedBy        *uuid.UUID `json:"deleted_by"`
}

func NewCategoryResponseDTO(c *domain.Category) CategoryResponseDTO {
	return CategoryResponseDTO{
		ID:               c.ID,
		UserID:           c.UserID,
		IsOwn:            c.IsOwn,
		IsBuiltIn:        c.IsBuiltIn,
		Name:             c.Name,
		Type:             c.Type,
		Icon:             c.Icon,
		Color:            c.Color,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		CreatedBy:        c.CreatedBy,
		UpdatedBy:        c.UpdatedBy,
		DeletedBy:        c.DeletedBy,
	}
}

func NewCategoryResponseDTOs(cs []domain.Category) []CategoryResponseDTO {
	var categories []CategoryResponseDTO
	for _, c := range cs {
		categories = append(categories, NewCategoryResponseDTO(&c))
	}
	return categories
}

// OptionCategoryTypeQueryDTO: the direction the caller wants, or none for both.
type OptionCategoryTypeQueryDTO struct {
	Type string `form:"type" binding:"omitempty,oneof=income expense" example:"expense"`
}

// OptionCategoryResponseDTO: an id here may name one of the caller's own
// categories or a master row it has not taken up yet. The caller does not have
// to tell them apart — it sends the id back as `category_id` either way.
type OptionCategoryResponseDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type" example:"expense"`
}

func NewOptionCategoryResponseDTOs(cs []domain.OptionCategoryType) []OptionCategoryResponseDTO {
	var categories []OptionCategoryResponseDTO
	for _, c := range cs {
		categories = append(categories, OptionCategoryResponseDTO{
			ID:   c.ID,
			Name: c.Name,
			Type: c.Type,
		})
	}
	return categories
}
