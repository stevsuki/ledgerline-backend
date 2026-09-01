package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type PasswordResetTokenRepository struct {
	mock.Mock
}

var _ domain.PasswordResetTokenRepository = (*PasswordResetTokenRepository)(nil)

func (m *PasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *PasswordResetTokenRepository) Update(ctx context.Context, token *domain.PasswordResetToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *PasswordResetTokenRepository) DeleteActiveByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *PasswordResetTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordResetToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}
