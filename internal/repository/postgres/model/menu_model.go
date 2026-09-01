package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type MenuModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ParentID  *uuid.UUID `gorm:"type:uuid"`
	Code      string     `gorm:"size:100;not null"`
	Name      string     `gorm:"size:100;not null"`
	Path      *string    `gorm:"size:255"`
	Icon      *string    `gorm:"size:50"`
	SortOrder int        `gorm:"not null;default:0"`
	IsActive  bool       `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Read-only: filled by the permissions join, excluded from every write.
	CanCreate  bool `gorm:"->;column:can_create"`
	CanRead    bool `gorm:"->;column:can_read"`
	CanUpdate  bool `gorm:"->;column:can_update"`
	CanDelete  bool `gorm:"->;column:can_delete"`
	CanApprove bool `gorm:"->;column:can_approve"`
}

func (MenuModel) TableName() string { return "menus" }

func (m MenuModel) ToDomain() *domain.Menu {
	menu := &domain.Menu{
		ID:        m.ID,
		ParentID:  m.ParentID,
		Code:      m.Code,
		Name:      m.Name,
		SortOrder: m.SortOrder,
		Access: domain.MenuAccess{
			CanCreate:  m.CanCreate,
			CanRead:    m.CanRead,
			CanUpdate:  m.CanUpdate,
			CanDelete:  m.CanDelete,
			CanApprove: m.CanApprove,
		},
	}
	if m.Path != nil {
		menu.Path = *m.Path
	}
	if m.Icon != nil {
		menu.Icon = *m.Icon
	}
	return menu
}

func MenusToDomain(models []MenuModel) []domain.Menu {
	out := make([]domain.Menu, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}
