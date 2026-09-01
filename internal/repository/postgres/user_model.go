package postgres

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type userModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"size:255;not null"`
	FullName     string    `gorm:"size:100;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	// Read-only: filled by the roles join, excluded from every insert & update.
	RoleName string `gorm:"->;column:role_name"`
	// default tags let GORM omit these on insert so the database fills them.
	Status              string    `gorm:"size:20;not null;default:enabled"`
	PasswordChangedAt   time.Time `gorm:"default:now()"`
	FailedLoginAttempts int       `gorm:"not null;default:0"`
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

func (userModel) TableName() string { return "users" }

func (m userModel) toDomain() *domain.User {
	return &domain.User{
		ID:                  m.ID,
		Email:               m.Email,
		FullName:            m.FullName,
		PasswordHash:        m.PasswordHash,
		RoleID:              m.RoleID,
		RoleName:            m.RoleName,
		Status:              domain.Status(m.Status),
		PasswordChangedAt:   m.PasswordChangedAt,
		FailedLoginAttempts: m.FailedLoginAttempts,
		LockedUntil:         m.LockedUntil,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func userFromDomain(u *domain.User) userModel {
	return userModel{
		ID:                u.ID,
		Email:             u.Email,
		FullName:          u.FullName,
		PasswordHash:      u.PasswordHash,
		RoleID:            u.RoleID,
		Status:            string(u.Status),
		PasswordChangedAt: u.PasswordChangedAt,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

func usersToDomain(models []userModel) []domain.User {
	out := make([]domain.User, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out
}
