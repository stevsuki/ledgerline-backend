package domain

import "errors"

// Domain sentinel errors; mapped to HTTP status in the handler layer.
var (
	ErrNotFound            = errors.New("resource not found")
	ErrConflict            = errors.New("resource already exists")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrTokenExpired        = errors.New("token expired")
	ErrTokenInvalid        = errors.New("token invalid")
	ErrInternal            = errors.New("internal server error")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInvalidOTP          = errors.New("invalid OTP")
	ErrMaxAttemptsExceeded = errors.New("maximum OTP attempts exceeded")
	ErrAccountLocked       = errors.New("account temporarily locked")
)
