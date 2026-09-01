package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type Mailer struct {
	mock.Mock
}

var _ domain.Mailer = (*Mailer)(nil)

func (m *Mailer) SendPasswordResetOTP(ctx context.Context, data domain.PasswordResetOTPMail) error {
	return m.Called(ctx, data).Error(0)
}
