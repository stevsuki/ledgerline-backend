package domain

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
)

type RequestMeta struct {
	IPAddress string
	UserAgent string
}

// NewRequestMeta normalizes the address to IPv4 where one exists.
func NewRequestMeta(ipAddress, userAgent string) RequestMeta {
	return RequestMeta{IPAddress: toIPv4(ipAddress), UserAgent: userAgent}
}

// toIPv4 prefers the IPv4 form, keeping a real IPv6 address as it is.
func toIPv4(address string) string {
	ip := net.ParseIP(address)
	if ip == nil {
		return address
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	if ip.IsLoopback() {
		return "127.0.0.1"
	}
	return ip.String()
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds
}

type TokenClaims struct {
	UserID   uuid.UUID
	FullName string
	Email    string
	RoleID   uuid.UUID
	RoleName string
	IssuedAt time.Time // set on parse; ignored when generating
}

// ResetToken: short-lived grant handed out once an OTP checks out.
type ResetToken struct {
	Token     string
	ExpiresIn int64 // seconds
}

// TokenManager: port for token issuing & verification.
type TokenManager interface {
	GenerateAccessToken(claims TokenClaims) (token string, expiresIn int64, err error)
	GenerateRefreshToken(claims TokenClaims) (string, error)
	Verify(token string) (*TokenClaims, error)
	VerifyRefreshToken(token string) (*TokenClaims, error)
	GenerateResetToken(claims TokenClaims) (token string, expiresIn int64, err error)
	VerifyResetToken(token string) (*TokenClaims, error)
}

// PasswordHasher: port for password hashing.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type RegisterInput struct {
	Email    string
	FullName string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type ForgotPasswordInput struct {
	Email string
}

type VerifyOTPResetPasswordInput struct {
	Email string
	OTP   string
}

type ResetPasswordInput struct {
	ResetToken         string
	NewPassword        string
	ConfirmNewPassword string
}

// Profile: what /auth/me answers -- who you are plus the sidebar you may see.
type Profile struct {
	User  User
	Menus []Menu
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*User, error)
	Login(ctx context.Context, input LoginInput, meta RequestMeta) (*TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	Me(ctx context.Context, userID uuid.UUID) (*Profile, error)
	ForgotPassword(ctx context.Context, input ForgotPasswordInput) error
	VerifyOTPResetPassword(ctx context.Context, input VerifyOTPResetPasswordInput) (*ResetToken, error)
	ResetPassword(ctx context.Context, input ResetPasswordInput) error
}
