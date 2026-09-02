package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// maxOTPAttempts caps wrong guesses per token; 6 digits are brute-forceable otherwise.
const maxOTPAttempts = 5

// Per-account brute force limit; the per-IP rate limiter misses rotating addresses.
const (
	maxLoginAttempts  = 5
	loginLockDuration = 15 * time.Minute
)

// resetCooldown throttles password reset requests per account.
const resetCooldown = 60 * time.Second

type authService struct {
	userRepo               domain.UserRepository
	hasher                 domain.PasswordHasher
	token                  domain.TokenManager
	mailer                 domain.Mailer
	otp                    domain.OTPGenerator
	passwordResetTokenRepo domain.PasswordResetTokenRepository
	menuRepo               domain.MenuRepository
	tx                     domain.TxManager
	otpTTL                 time.Duration
	auditLogRepo           domain.AuditLogRepository
}

func NewAuthService(
	userRepo domain.UserRepository,
	hasher domain.PasswordHasher,
	token domain.TokenManager,
	mailer domain.Mailer,
	otp domain.OTPGenerator,
	passwordResetTokenRepo domain.PasswordResetTokenRepository,
	menuRepo domain.MenuRepository,
	tx domain.TxManager,
	otpTTL time.Duration,
	auditLogRepo domain.AuditLogRepository,
) domain.AuthService {
	return &authService{userRepo: userRepo, hasher: hasher, token: token, mailer: mailer, otp: otp, passwordResetTokenRepo: passwordResetTokenRepo, menuRepo: menuRepo, tx: tx, otpTTL: otpTTL, auditLogRepo: auditLogRepo}
}

func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.Conflict(domain.CodeUserEmailTaken, "email is already registered").WithField("email")
	}

	hashed, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	user := &domain.User{
		ID:           id,
		Email:        email,
		FullName:     strings.TrimSpace(input.FullName),
		PasswordHash: hashed,
		RoleID:       domain.RoleIDUser,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(ctx context.Context, input domain.LoginInput, meta domain.RequestMeta) (*domain.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		// Only an unknown address is a credential problem; anything else is a real failure.
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		s.auditLoginFailure(ctx, user, meta, domain.AuditSeverityCritical,
			domain.NewAccountLockedDetail(user.FailedLoginAttempts, user.Email))
		return nil, domain.ErrAccountLocked.WithRetryAfter(time.Until(*user.LockedUntil))
	}

	if err := s.hasher.Compare(user.PasswordHash, input.Password); err != nil {
		if user.LockedUntil != nil {
			if err := s.userRepo.ClearFailedLogins(ctx, user.ID); err != nil {
				return nil, err
			}
		}

		attempts, err := s.userRepo.IncrementFailedLogin(ctx, user.ID)
		if err != nil {
			return nil, err
		}

		locked := attempts >= maxLoginAttempts
		if locked {
			if err := s.userRepo.LockUntil(ctx, user.ID, time.Now().Add(loginLockDuration)); err != nil {
				return nil, err
			}
		}

		// Locking someone out is worth more than a warning in the audit feed.
		severity := domain.AuditSeverityWarning
		detail := domain.NewWrongPasswordDetail(attempts, maxLoginAttempts, user.Email)
		if locked {
			severity = domain.AuditSeverityCritical
			detail = domain.NewAccountLockedDetail(attempts, user.Email)
		}
		s.auditLoginFailure(ctx, user, meta, severity, detail)

		if locked {
			return nil, domain.ErrAccountLocked.WithRetryAfter(loginLockDuration)
		}
		return nil, domain.ErrInvalidCredentials
	}

	token, err := s.issueTokens(user)
	if err != nil {
		return nil, err
	}

	// Only write when there is something to clear.
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		if err := s.userRepo.ClearFailedLogins(ctx, user.ID); err != nil {
			return nil, err
		}
	}

	recordAudit(ctx, s.auditLogRepo, &domain.AuditLog{
		UserID:       user.ID,
		UserFullName: user.FullName,
		RoleName:     user.RoleName,
		Action:       "auth.login",
		Status:       domain.AuditStatusSuccess,
		Severity:     domain.AuditSeverityInfo,
		Module:       domain.AuditModuleAuth,
		IPAddress:    &meta.IPAddress,
		Details:      domain.NewSessionDetail(domain.SessionMethodPassword, meta.UserAgent),
	})

	return token, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	claims, err := s.token.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Re-read the user so role changes take effect immediately.
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrTokenInvalid
		}
		return nil, err
	}
	return s.issueTokens(user)
}

func (s *authService) Me(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	menus, err := s.menuRepo.ListByRole(ctx, user.RoleID)
	if err != nil {
		return nil, err
	}

	return &domain.Profile{User: *user, Menus: visibleMenuTree(menus)}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Answer the same way for an unknown address, otherwise this endpoint lists who has an account.
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}

	existingForgotAttempt, err := s.passwordResetTokenRepo.GetByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("check existing password reset token: %w", err)
	}

	if existingForgotAttempt != nil {
		if wait := resetCooldown - time.Since(existingForgotAttempt.CreatedAt); wait > 0 {
			return domain.RateLimited(domain.CodeResetTooSoon,
				"a reset code was just sent, please wait before requesting another").WithRetryAfter(wait)
		}
	}

	otp, err := s.otp.Generate()
	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}

	hashedOTP, err := s.hasher.Hash(otp)
	if err != nil {
		return fmt.Errorf("hash OTP: %w", err)
	}

	tokenID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate token id: %w", err)
	}

	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.passwordResetTokenRepo.DeleteActiveByUserID(ctx, user.ID); err != nil {
			return fmt.Errorf("delete active password reset token: %w", err)
		}
		return s.passwordResetTokenRepo.Create(ctx, &domain.PasswordResetToken{
			ID:        tokenID,
			UserID:    user.ID,
			OTPHash:   hashedOTP,
			ExpiresAt: time.Now().Add(s.otpTTL),
		})
	})
	if err != nil {
		return err
	}

	err = s.mailer.SendPasswordResetOTP(ctx, domain.PasswordResetOTPMail{Email: email, FullName: user.FullName, OTP: otp})
	if err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}

func (s *authService) VerifyOTPResetPassword(ctx context.Context, input domain.VerifyOTPResetPasswordInput) (*domain.ResetToken, error) {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		// A missing account must look exactly like a wrong code.
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidOTP
		}
		return nil, err
	}

	token, err := s.passwordResetTokenRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidOTP
		}
		return nil, err
	}

	// Spent code: same answer as a wrong one, and not the session-token code the client acts on.
	if token.UsedAt != nil {
		return nil, domain.ErrInvalidOTP
	}

	// Check if attempts exceed the limit
	if token.Attempts >= maxOTPAttempts {
		return nil, domain.ErrMaxAttemptsExceeded
	}

	// Check if the token expired or not
	if time.Now().After(token.ExpiresAt) {
		return nil, domain.ErrInvalidOTP
	}

	// Check if the token valid or not
	if err = s.hasher.Compare(token.OTPHash, input.OTP); err != nil {
		attempts := token.Attempts + 1

		// Burn the token on the last allowed try so the user must request a new code.
		var burnedAt *time.Time
		if attempts >= maxOTPAttempts {
			now := time.Now()
			burnedAt = &now
		}

		if err = s.passwordResetTokenRepo.Update(ctx, &domain.PasswordResetToken{
			ID:       token.ID,
			Attempts: attempts,
			UsedAt:   burnedAt,
		}); err != nil {
			return nil, fmt.Errorf("increment OTP attempts: %w", err)
		}
		return nil, domain.ErrInvalidOTP
	}

	now := time.Now()
	if err = s.passwordResetTokenRepo.Update(ctx, &domain.PasswordResetToken{
		ID:       token.ID,
		Attempts: token.Attempts,
		UsedAt:   &now,
	}); err != nil {
		return nil, fmt.Errorf("mark OTP as used: %w", err)
	}

	// The OTP is spent; this token is what authorises the actual password change.
	resetToken, expiresIn, err := s.token.GenerateResetToken(domain.TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		RoleID: user.RoleID,
	})
	if err != nil {
		return nil, fmt.Errorf("generate reset token: %w", err)
	}

	return &domain.ResetToken{Token: resetToken, ExpiresIn: expiresIn}, nil
}

func (s *authService) ResetPassword(ctx context.Context, input domain.ResetPasswordInput) error {
	claims, err := s.token.VerifyResetToken(input.ResetToken)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrTokenInvalid
		}
		return err
	}

	if claims.IssuedAt.Before(user.PasswordChangedAt.Truncate(time.Second)) {
		return domain.ErrTokenInvalid
	}

	hashed, err := s.hasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	return s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.userRepo.UpdatePassword(ctx, user.ID, hashed); err != nil {
			return err
		}
		if err := s.passwordResetTokenRepo.DeleteActiveByUserID(ctx, user.ID); err != nil {
			return fmt.Errorf("clear password reset tokens: %w", err)
		}
		return nil
	})
}

func (s *authService) issueTokens(user *domain.User) (*domain.TokenPair, error) {
	claims := domain.TokenClaims{UserID: user.ID, FullName: user.FullName, Email: user.Email, RoleID: user.RoleID, RoleName: user.RoleName}

	accessToken, expiresIn, err := s.token.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.token.GenerateRefreshToken(claims)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// auditLoginFailure records one rejected sign-in, whatever rejected it.
func (s *authService) auditLoginFailure(
	ctx context.Context, user *domain.User, meta domain.RequestMeta,
	severity domain.AuditSeverity, detail domain.SessionFailedDetail,
) {
	recordAudit(ctx, s.auditLogRepo, &domain.AuditLog{
		UserID:       user.ID,
		UserFullName: user.FullName,
		RoleName:     user.RoleName,
		Action:       "auth.login",
		Status:       domain.AuditStatusFailed,
		Severity:     severity,
		Module:       domain.AuditModuleAuth,
		IPAddress:    &meta.IPAddress,
		Details:      detail,
	})
}
