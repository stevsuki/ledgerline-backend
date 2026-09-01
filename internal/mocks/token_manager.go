package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

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

func (m *TokenManager) GenerateResetToken(claims domain.TokenClaims) (string, int64, error) {
	args := m.Called(claims)
	return args.String(0), int64(args.Int(1)), args.Error(2)
}

// Each Verify calls m.Called itself: testify names the expectation after the caller.
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

func (m *TokenManager) VerifyResetToken(token string) (*domain.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}
