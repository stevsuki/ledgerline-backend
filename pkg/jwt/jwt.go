// Package jwt: domain.TokenManager implementation using JWT HS256.
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
	tokenTypeReset   = "reset"
)

type Claims struct {
	jwt.RegisteredClaims
	Email     string `json:"email"`
	RoleID    string `json:"role_id"`
	TokenType string `json:"token_type"`
}

type Manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	resetTTL   time.Duration
}

func NewManager(secret, issuer string, accessTTL, refreshTTL, resetTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		resetTTL:   resetTTL,
	}
}

// Compile-time check against the domain contract.
var _ domain.TokenManager = (*Manager)(nil)

func (m *Manager) GenerateAccessToken(c domain.TokenClaims) (string, int64, error) {
	token, err := m.generate(c, tokenTypeAccess, m.accessTTL)
	if err != nil {
		return "", 0, err
	}
	return token, int64(m.accessTTL.Seconds()), nil
}

func (m *Manager) GenerateRefreshToken(c domain.TokenClaims) (string, error) {
	return m.generate(c, tokenTypeRefresh, m.refreshTTL)
}

// GenerateResetToken: short-lived token that authorises one password reset.
func (m *Manager) GenerateResetToken(c domain.TokenClaims) (string, int64, error) {
	token, err := m.generate(c, tokenTypeReset, m.resetTTL)
	if err != nil {
		return "", 0, err
	}
	return token, int64(m.resetTTL.Seconds()), nil
}

func (m *Manager) generate(c domain.TokenClaims, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.UserID.String(),
			Issuer:    m.issuer,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		Email:     c.Email,
		RoleID:    c.RoleID.String(),
		TokenType: tokenType,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Verify: validate an access token.
func (m *Manager) Verify(token string) (*domain.TokenClaims, error) {
	return m.parse(token, tokenTypeAccess)
}

// VerifyRefreshToken: ensure the token is of the refresh type.
func (m *Manager) VerifyRefreshToken(token string) (*domain.TokenClaims, error) {
	return m.parse(token, tokenTypeRefresh)
}

// VerifyResetToken: ensure the token is of the reset type.
func (m *Manager) VerifyResetToken(token string) (*domain.TokenClaims, error) {
	return m.parse(token, tokenTypeReset)
}

func (m *Manager) parse(tokenString, wantType string) (*domain.TokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || claims.TokenType != wantType {
		return nil, domain.ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	issuedAt := time.Time{}
	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}

	roleID, err := uuid.Parse(claims.RoleID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	return &domain.TokenClaims{
		UserID:   userID,
		Email:    claims.Email,
		RoleID:   roleID,
		IssuedAt: issuedAt,
	}, nil
}
