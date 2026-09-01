// Package mocks: stub implementations of domain interfaces for unit tests.
package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type UserRepository struct {
	mock.Mock
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (m *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

type PasswordHasher struct {
	mock.Mock
}

var _ domain.PasswordHasher = (*PasswordHasher)(nil)

func (m *PasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *PasswordHasher) Compare(hash, password string) error {
	return m.Called(hash, password).Error(0)
}

type TokenManager struct {
	mock.Mock
}

var _ domain.TokenManager = (*TokenManager)(nil)

func (m *TokenManager) GenerateAccessToken(claims domain.TokenClaims) (string, int64, error) {
	args := m.Called(claims)
	return args.String(0), int64(args.Int(1)), args.Error(2)
}

func (m *TokenManager) GenerateRefreshToken(claims domain.TokenClaims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *TokenManager) Verify(token string) (*domain.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

func (m *TokenManager) VerifyRefreshToken(token string) (*domain.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

type UserService struct {
	mock.Mock
}

var _ domain.UserService = (*UserService)(nil)

func (m *UserService) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *UserService) Update(ctx context.Context, id uuid.UUID, input domain.UpdateUserInput) (*domain.User, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type Mailer struct {
	mock.Mock
}

var _ domain.Mailer = (*Mailer)(nil)

func (m *Mailer) SendPasswordResetOTP(ctx context.Context, data domain.PasswordResetOTPMail) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

type OTPGenerator struct {
	mock.Mock
}

var _ domain.OTPGenerator = (*OTPGenerator)(nil)

func (m *OTPGenerator) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type PasswordResetTokenRepository struct {
	mock.Mock
}

var _ domain.PasswordResetTokenRepository = (*PasswordResetTokenRepository)(nil)

func (m *PasswordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *PasswordResetTokenRepository) GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}

func (m *PasswordResetTokenRepository) Update(ctx context.Context, token *domain.PasswordResetToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *PasswordResetTokenRepository) DeleteActiveByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *PasswordResetTokenRepository) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *PasswordResetTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordResetToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordResetToken), args.Error(1)
}

func (m *TokenManager) GenerateResetToken(claims domain.TokenClaims) (string, int64, error) {
	args := m.Called(claims)
	return args.String(0), int64(args.Int(1)), args.Error(2)
}

func (m *TokenManager) VerifyResetToken(token string) (*domain.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

func (m *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return m.Called(ctx, userID, passwordHash).Error(0)
}

// expects reports whether a test set an expectation for this method. Checking
// the whole mock is not enough: a login test sets expectations on GetByEmail
// while wanting nothing to do with the lockout calls.
func expects(m *mock.Mock, method string) bool {
	for _, call := range m.ExpectedCalls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// The lockout calls stay silent unless a test asks for them: most login tests
// are not about lockout.
func (m *UserRepository) IncrementFailedLogin(ctx context.Context, userID uuid.UUID) (int, error) {
	if !expects(&m.Mock, "IncrementFailedLogin") {
		return 1, nil
	}
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *UserRepository) LockUntil(ctx context.Context, userID uuid.UUID, until time.Time) error {
	if !expects(&m.Mock, "LockUntil") {
		return nil
	}
	return m.Called(ctx, userID, until).Error(0)
}

func (m *UserRepository) ClearFailedLogins(ctx context.Context, userID uuid.UUID) error {
	if !expects(&m.Mock, "ClearFailedLogins") {
		return nil
	}
	return m.Called(ctx, userID).Error(0)
}

type MenuRepository struct {
	mock.Mock
}

var _ domain.MenuRepository = (*MenuRepository)(nil)

func (m *MenuRepository) ListByRole(ctx context.Context, roleID uuid.UUID) ([]domain.Menu, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Menu), args.Error(1)
}
