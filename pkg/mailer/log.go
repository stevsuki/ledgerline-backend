package mailer

import (
	"context"
	"log/slog"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// LogMailer: a dev mailer that prints emails to the log instead of sending them.
type LogMailer struct {
	log *slog.Logger
}

var _ domain.Mailer = (*LogMailer)(nil)

func NewLog(log *slog.Logger) *LogMailer {
	return &LogMailer{log: log}
}

func (m *LogMailer) SendPasswordResetOTP(_ context.Context, data domain.PasswordResetOTPMail) error {
	m.log.Info("password reset OTP email (log mode)",
		slog.String("to", data.Email),
		slog.String("otp", data.OTP),
		slog.Duration("expires_in", data.ExpiresIn),
	)
	return nil
}
