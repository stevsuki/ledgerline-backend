package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// CategoryModel: sizes and nullability follow migration 000022.
type CategoryModel struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	// Nullable: ON DELETE SET NULL cuts the link when its master row goes.
	MasterCategoryID *uuid.UUID `gorm:"type:uuid"`
	Name             string     `gorm:"size:100;not null"`
	Type             string     `gorm:"size:20;not null"`
	// Both are NOT NULL DEFAULT '' (migration 000026), so a row written before
	// they existed scans back without needing a pointer.
	Icon      string `gorm:"size:50;not null;default:''"`
	Color     string `gorm:"size:10;not null;default:''"`
	CreatedAt time.Time
	CreatedBy *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	DeletedBy *uuid.UUID     `gorm:"type:uuid"`
}

func (CategoryModel) TableName() string { return "categories" }

func (m CategoryModel) ToDomain() *domain.Category {
	category := &domain.Category{
		ID:        m.ID,
		UserID:    m.UserID,
		Name:      m.Name,
		Type:      m.Type,
		Icon:      m.Icon,
		Color:     m.Color,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
		DeletedBy: m.DeletedBy,
	}
	// A NULL link reads as the zero uuid, which is the domain's "no master".
	if m.MasterCategoryID != nil {
		category.MasterCategoryID = *m.MasterCategoryID
	}
	return category
}

func CategoryFromDomain(c *domain.Category) CategoryModel {
	return CategoryModel{
		ID:               c.ID,
		UserID:           c.UserID,
		MasterCategoryID: MasterCategoryRef(c.MasterCategoryID),
		Name:             c.Name,
		Type:             c.Type,
		Icon:             c.Icon,
		Color:            c.Color,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		CreatedBy:        c.CreatedBy,
		UpdatedBy:        c.UpdatedBy,
	}
}

// MasterCategoryRef: the zero uuid is stored as NULL, so a category with no
// master reads back the same way it was written. The repository writes the
// column directly on update, which is why this is not kept private.
func MasterCategoryRef(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func CategoriesToDomain(models []CategoryModel) []domain.Category {
	out := make([]domain.Category, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}
