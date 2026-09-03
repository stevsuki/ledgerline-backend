package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type RoleModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"size:50;not null"`
	Description string    `gorm:"size:255"`
	Icon        string    `gorm:"size:50"`
	IsSystem    bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time
	CreatedBy   *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID `gorm:"type:uuid"`
	// Read-only: filled by the users join in List, excluded from every write.
	UserCount int            `gorm:"->;column:user_count"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	DeletedBy *uuid.UUID     `gorm:"type:uuid"`
}

func (RoleModel) TableName() string { return "roles" }

func (m RoleModel) ToDomain() *domain.Role {
	return &domain.Role{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Icon:        m.Icon,
		IsSystem:    m.IsSystem,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
		DeletedBy:   m.DeletedBy,
		UserCount:   m.UserCount,
	}
}

func RoleFromDomain(r *domain.Role) RoleModel {
	return RoleModel{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Icon:        r.Icon,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CreatedBy:   r.CreatedBy,
		UpdatedBy:   r.UpdatedBy,
	}
}

func RolesToDomain(models []RoleModel) []domain.Role {
	out := make([]domain.Role, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}

type RoleMenuPermissionModel struct {
	RoleID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	MenuID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	CanCreate  bool      `gorm:"not null;default:false"`
	CanRead    bool      `gorm:"not null;default:false"`
	CanUpdate  bool      `gorm:"not null;default:false"`
	CanDelete  bool      `gorm:"not null;default:false"`
	CanApprove bool      `gorm:"not null;default:false"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (RoleMenuPermissionModel) TableName() string { return "role_menu_permissions" }

func (m RoleMenuPermissionModel) ToDomain() *domain.RoleMenuPermission {
	return &domain.RoleMenuPermission{
		RoleID:     m.RoleID,
		MenuID:     m.MenuID,
		CanCreate:  m.CanCreate,
		CanRead:    m.CanRead,
		CanUpdate:  m.CanUpdate,
		CanDelete:  m.CanDelete,
		CanApprove: m.CanApprove,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func RoleMenuPermissionsToDomain(models []RoleMenuPermissionModel) []domain.RoleMenuPermission {
	out := make([]domain.RoleMenuPermission, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}

func RoleMenuPermissionFromDomain(p *domain.RoleMenuPermission) RoleMenuPermissionModel {
	return RoleMenuPermissionModel{
		RoleID:     p.RoleID,
		MenuID:     p.MenuID,
		CanCreate:  p.CanCreate,
		CanRead:    p.CanRead,
		CanUpdate:  p.CanUpdate,
		CanDelete:  p.CanDelete,
		CanApprove: p.CanApprove,
	}
}
