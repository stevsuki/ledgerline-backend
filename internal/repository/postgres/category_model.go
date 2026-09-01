package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type categoryModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"size:100;not null"`
	Type      string    `gorm:"size:20;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (categoryModel) TableName() string { return "categories" }

func (m categoryModel) toDomain() *domain.Category {
	return &domain.Category{
		ID:        m.ID,
		UserID:    m.UserID,
		Name:      m.Name,
		Type:      domain.CategoryType(m.Type),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func categoryFromDomain(c *domain.Category) categoryModel {
	return categoryModel{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name,
		Type:      string(c.Type),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func categoriesToDomain(models []categoryModel) []domain.Category {
	out := make([]domain.Category, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out
}
