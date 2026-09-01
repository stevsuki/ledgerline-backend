package domain

import (
	"context"
	"time"
)

// PasswordResetOTPMail: data for the password reset OTP email.
type PasswordResetOTPMail struct {
	Email     string
	FullName  string
	OTP       string
	ExpiresIn time.Duration
}

// Mailer: port for outbound email delivery.
type Mailer interface {
	SendPasswordResetOTP(ctx context.Context, data PasswordResetOTPMail) error
}
