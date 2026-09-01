package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	OTPHash   string
	ExpiresAt time.Time
	CreatedAt time.Time
	Attempts  int
	UsedAt    *time.Time
}

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	Update(ctx context.Context, payload *PasswordResetToken) error
	ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error)
	DeleteActiveByUserID(ctx context.Context, userID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*PasswordResetToken, error)
}
