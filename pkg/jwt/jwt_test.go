package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/jwt"
)

func newClaims() domain.TokenClaims {
	return domain.TokenClaims{UserID: uuid.New(), Email: "budi@example.com", RoleID: domain.RoleIDAdmin}
}

func TestManager_AccessToken(t *testing.T) {
	t.Parallel()

	manager := jwt.NewManager("very-secret-key-for-unit-tests-32", "ledgerline", 15*time.Minute, time.Hour, 10*time.Minute)
	claims := newClaims()

	token, expiresIn, err := manager.GenerateAccessToken(claims)
	require.NoError(t, err)
	assert.EqualValues(t, 900, expiresIn)

	got, err := manager.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, got.UserID)
	assert.Equal(t, claims.RoleID, got.RoleID)
}

func TestManager_RefreshTokenIsNotAcceptedAsAccessToken(t *testing.T) {
	t.Parallel()

	manager := jwt.NewManager("very-secret-key-for-unit-tests-32", "ledgerline", 15*time.Minute, time.Hour, 10*time.Minute)

	refresh, err := manager.GenerateRefreshToken(newClaims())
	require.NoError(t, err)

	_, err = manager.Verify(refresh)
	require.ErrorIs(t, err, domain.ErrTokenInvalid)

	_, err = manager.VerifyRefreshToken(refresh)
	require.NoError(t, err)
}

func TestManager_ExpiredToken(t *testing.T) {
	t.Parallel()

	manager := jwt.NewManager("very-secret-key-for-unit-tests-32", "ledgerline", -time.Minute, time.Hour, 10*time.Minute)

	token, _, err := manager.GenerateAccessToken(newClaims())
	require.NoError(t, err)

	_, err = manager.Verify(token)
	require.ErrorIs(t, err, domain.ErrTokenExpired)
}

func TestManager_DifferentSecretIsRejected(t *testing.T) {
	t.Parallel()

	issuer := jwt.NewManager("secret-a-for-issuer-mismatch-test", "ledgerline", time.Minute, time.Hour, 10*time.Minute)
	attacker := jwt.NewManager("secret-b-for-issuer-mismatch-test", "ledgerline", time.Minute, time.Hour, 10*time.Minute)

	token, _, err := issuer.GenerateAccessToken(newClaims())
	require.NoError(t, err)

	_, err = attacker.Verify(token)
	require.ErrorIs(t, err, domain.ErrTokenInvalid)
}
