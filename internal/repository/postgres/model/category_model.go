package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type CategoryModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"size:100;not null"`
	Type      string    `gorm:"size:20;not null"`
	CreatedAt time.Time
	CreatedBy *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	DeletedBy *uuid.UUID     `gorm:"type:uuid"`
}

func (CategoryModel) TableName() string { return "categories" }

func (m CategoryModel) ToDomain() *domain.Category {
	return &domain.Category{
		ID:        m.ID,
		UserID:    m.UserID,
		Name:      m.Name,
		Type:      domain.CategoryType(m.Type),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		CreatedBy: m.CreatedBy,
		UpdatedBy: m.UpdatedBy,
		DeletedBy: m.DeletedBy,
	}
}

func CategoryFromDomain(c *domain.Category) CategoryModel {
	return CategoryModel{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Type:      string(c.Type),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		CreatedBy: c.CreatedBy,
		UpdatedBy: c.UpdatedBy,
	}
}

func CategoriesToDomain(models []CategoryModel) []domain.Category {
	out := make([]domain.Category, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}
