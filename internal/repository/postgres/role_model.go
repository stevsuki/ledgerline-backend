package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type roleModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"size:50;not null"`
	Description string    `gorm:"size:255"`
	IsSystem    bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Read-only: filled by the users join in List, excluded from every write.
	UserCount int            `gorm:"->;column:user_count"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (roleModel) TableName() string { return "roles" }

func (m roleModel) toDomain() *domain.Role {
	return &domain.Role{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsSystem:    m.IsSystem,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		UserCount:   m.UserCount,
	}
}

func roleFromDomain(r *domain.Role) roleModel {
	return roleModel{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func rolesToDomain(models []roleModel) []domain.Role {
	out := make([]domain.Role, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out
}

type roleMenuPermissionModel struct {
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

func (roleMenuPermissionModel) TableName() string { return "role_menu_permissions" }

func (m roleMenuPermissionModel) toDomain() *domain.RoleMenuPermission {
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

func roleMenuPermissionsToDomain(models []roleMenuPermissionModel) []domain.RoleMenuPermission {
	out := make([]domain.RoleMenuPermission, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out
}

func roleMenuPermissionFromDomain(p *domain.RoleMenuPermission) roleMenuPermissionModel {
	return roleMenuPermissionModel{
		RoleID:     p.RoleID,
		MenuID:     p.MenuID,
		CanCreate:  p.CanCreate,
		CanRead:    p.CanRead,
		CanUpdate:  p.CanUpdate,
		CanDelete:  p.CanDelete,
		CanApprove: p.CanApprove,
	}
}
