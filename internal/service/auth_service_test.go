package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/mocks"
	"github.com/stevensuki/ledgerline-backend/internal/service"
)

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "budi@example.com",
		PasswordHash: "hashed",
		RoleID:       domain.RoleIDUser,
	}

	t.Run("logs in successfully", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)
		token := new(mocks.TokenManager)

		repo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)
		hasher.On("Compare", "hashed", "Rahasia123!").Return(nil)
		token.On("GenerateAccessToken", mock.AnythingOfType("domain.TokenClaims")).Return("access", 900, nil)
		token.On("GenerateRefreshToken", mock.AnythingOfType("domain.TokenClaims")).Return("refresh", nil)

		pair, err := service.NewAuthService(repo, hasher, token, new(mocks.Mailer), new(mocks.OTPGenerator), new(mocks.PasswordResetTokenRepository), new(mocks.MenuRepository), new(mocks.TxManager), 10*time.Minute, new(mocks.AuditLogRepository)).
			Login(context.Background(), domain.LoginInput{Email: "Budi@example.com", Password: "Rahasia123!"}, domain.RequestMeta{})

		require.NoError(t, err)
		assert.Equal(t, "access", pair.AccessToken)
		assert.Equal(t, "refresh", pair.RefreshToken)
		assert.EqualValues(t, 900, pair.ExpiresIn)
	})

	t.Run("a wrong password still returns ErrInvalidCredentials", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		repo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)
		hasher.On("Compare", "hashed", "wrong-password").Return(bcrypt.ErrMismatchedHashAndPassword)

		_, err := service.NewAuthService(repo, hasher, new(mocks.TokenManager), new(mocks.Mailer), new(mocks.OTPGenerator), new(mocks.PasswordResetTokenRepository), new(mocks.MenuRepository), new(mocks.TxManager), 10*time.Minute, new(mocks.AuditLogRepository)).
			Login(context.Background(), domain.LoginInput{Email: user.Email, Password: "wrong-password"}, domain.RequestMeta{})

		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("an unregistered email is not disclosed", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		repo.On("GetByEmail", mock.Anything, "hantu@example.com").Return(nil, domain.ErrNotFound)

		_, err := service.NewAuthService(repo, new(mocks.PasswordHasher), new(mocks.TokenManager), new(mocks.Mailer), new(mocks.OTPGenerator), new(mocks.PasswordResetTokenRepository), new(mocks.MenuRepository), new(mocks.TxManager), 10*time.Minute, new(mocks.AuditLogRepository)).
			Login(context.Background(), domain.LoginInput{Email: "hantu@example.com", Password: "Rahasia123!"}, domain.RequestMeta{})

		require.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})
}

func TestAuthService_Refresh(t *testing.T) {
	t.Parallel()

	user := &domain.User{ID: uuid.New(), Email: "budi@example.com", RoleID: domain.RoleIDUser}

	t.Run("a valid refresh token issues a new pair", func(t *testing.T) {
		t.Parallel()

		repo := new(mocks.UserRepository)
		token := new(mocks.TokenManager)

		token.On("VerifyRefreshToken", "refresh-token").
			Return(&domain.TokenClaims{UserID: user.ID, Email: user.Email, RoleID: user.RoleID}, nil)
		repo.On("GetByID", mock.Anything, user.ID).Return(user, nil)
		token.On("GenerateAccessToken", mock.AnythingOfType("domain.TokenClaims")).Return("new-access", 900, nil)
		token.On("GenerateRefreshToken", mock.AnythingOfType("domain.TokenClaims")).Return("new-refresh", nil)

		pair, err := service.NewAuthService(repo, new(mocks.PasswordHasher), token, new(mocks.Mailer), new(mocks.OTPGenerator), new(mocks.PasswordResetTokenRepository), new(mocks.MenuRepository), new(mocks.TxManager), 10*time.Minute, new(mocks.AuditLogRepository)).
			Refresh(context.Background(), "refresh-token")

		require.NoError(t, err)
		assert.Equal(t, "new-access", pair.AccessToken)
	})

	t.Run("an expired token is rejected", func(t *testing.T) {
		t.Parallel()

		token := new(mocks.TokenManager)
		token.On("VerifyRefreshToken", "expired").Return(nil, domain.ErrTokenExpired)

		_, err := service.NewAuthService(new(mocks.UserRepository), new(mocks.PasswordHasher), token, new(mocks.Mailer), new(mocks.OTPGenerator), new(mocks.PasswordResetTokenRepository), new(mocks.MenuRepository), new(mocks.TxManager), 10*time.Minute, new(mocks.AuditLogRepository)).
			Refresh(context.Background(), "expired")

		require.ErrorIs(t, err, domain.ErrTokenExpired)
	})
}
