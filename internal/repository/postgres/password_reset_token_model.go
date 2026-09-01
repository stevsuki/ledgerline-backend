package postgres

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type passwordResetTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	OTPHash   string    `gorm:"size:255;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	Attempts  int        `gorm:"not null;default:0"`
	UsedAt    *time.Time `gorm:"default:null"`
}

func (passwordResetTokenModel) TableName() string { return "password_reset_tokens" }

func (m passwordResetTokenModel) toDomain() *domain.PasswordResetToken {
	return &domain.PasswordResetToken{
		ID:        m.ID,
		UserID:    m.UserID,
		OTPHash:   m.OTPHash,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
		Attempts:  m.Attempts,
		UsedAt:    m.UsedAt,
	}
}

func passwordResetTokenFromDomain(t *domain.PasswordResetToken) passwordResetTokenModel {
	return passwordResetTokenModel{
		ID:        t.ID,
		UserID:    t.UserID,
		OTPHash:   t.OTPHash,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
		Attempts:  t.Attempts,
		UsedAt:    t.UsedAt,
	}
}
