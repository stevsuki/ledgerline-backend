package domain

import (
	"errors"
	"fmt"
	"time"
)

// Kind classifies a failure; the HTTP layer turns it into a status code.
type Kind uint8

const (
	KindInternal Kind = iota
	KindNotFound
	KindConflict
	KindInvalidInput
	KindUnauthorized
	KindForbidden
	KindRateLimited
	KindTimeout
)

// Coarse sentinels, kept so any layer can ask "is this a not-found?" without unwrapping.
var (
	ErrInternal        = errors.New("internal error")
	ErrNotFound        = errors.New("resource not found")
	ErrConflict        = errors.New("resource already exists")
	ErrInvalidInput    = errors.New("invalid input")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrTooManyRequests = errors.New("too many requests")
)

// sentinel: the coarse error a kind answers errors.Is for.
func (k Kind) sentinel() error {
	switch k {
	case KindInternal:
		return ErrInternal
	case KindNotFound:
		return ErrNotFound
	case KindConflict:
		return ErrConflict
	case KindInvalidInput:
		return ErrInvalidInput
	case KindUnauthorized:
		return ErrUnauthorized
	case KindForbidden:
		return ErrForbidden
	case KindRateLimited:
		return ErrTooManyRequests
	default:
		return nil
	}
}

// Error carries everything the HTTP layer needs: Kind picks the status, Code is the
// contract the frontend switches on, Message is safe to display, cause is log-only.
type Error struct {
	Kind       Kind
	Code       string
	Message    string
	Field      string        // optional: the request field to highlight
	RetryAfter time.Duration // optional: fills the Retry-After header
	cause      error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return e.Code + ": " + e.Message
}

// Unwrap exposes the cause to errors.Is/As, never to the client.
func (e *Error) Unwrap() error { return e.cause }

// Is matches a coarse sentinel by kind, and another domain error by code.
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return t.Code == e.Code
	}
	s := e.Kind.sentinel()
	return s != nil && target == s
}

func newError(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

func NotFound(code, message string) *Error     { return newError(KindNotFound, code, message) }
func Conflict(code, message string) *Error     { return newError(KindConflict, code, message) }
func InvalidInput(code, message string) *Error { return newError(KindInvalidInput, code, message) }
func Unauthorized(code, message string) *Error { return newError(KindUnauthorized, code, message) }
func Forbidden(code, message string) *Error    { return newError(KindForbidden, code, message) }
func RateLimited(code, message string) *Error  { return newError(KindRateLimited, code, message) }

// clone keeps the package-level errors below immutable.
func (e *Error) clone() *Error {
	c := *e
	return &c
}

// WithField names the request field that caused the failure.
func (e *Error) WithField(field string) *Error {
	c := e.clone()
	c.Field = field
	return c
}

// WithCause attaches the underlying error for the log.
func (e *Error) WithCause(cause error) *Error {
	c := e.clone()
	c.cause = cause
	return c
}

// WithRetryAfter tells the client how long to wait.
func (e *Error) WithRetryAfter(d time.Duration) *Error {
	c := e.clone()
	c.RetryAfter = d
	return c
}

// WithMessage overrides the display message, keeping code and kind.
func (e *Error) WithMessage(message string) *Error {
	c := e.clone()
	c.Message = message
	return c
}

// AsError pulls the domain error out of a wrapped chain.
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// Predefined failures shared across services; treat them as read-only templates.
var (
	ErrInvalidCredentials  = Unauthorized(CodeInvalidCredentials, "invalid email or password")
	ErrTokenExpired        = Unauthorized(CodeTokenExpired, "the token has expired")
	ErrTokenInvalid        = Unauthorized(CodeTokenInvalid, "the token is invalid")
	ErrTokenMissing        = Unauthorized(CodeTokenMissing, "an Authorization header formatted as 'Bearer <token>' is required")
	ErrAuthRequired        = Unauthorized(CodeUnauthorized, "authentication is required")
	ErrAccessDenied        = Forbidden(CodeForbidden, "you do not have access to this resource")
	ErrInvalidOTP          = InvalidInput(CodeInvalidOTP, "the OTP is invalid or has expired")
	ErrMaxAttemptsExceeded = RateLimited(CodeOTPMaxAttempts, "too many wrong OTP attempts, request a new code")
	ErrAccountLocked       = RateLimited(CodeAccountLocked, "too many failed sign-in attempts, please try again later")
	ErrTimeout             = newError(KindTimeout, CodeTimeout, "the request took too long to complete")
)

// A bare sentinel returned by hand still deserves its status, just with a generic code.
var sentinelFallbacks = []*Error{
	NotFound(CodeNotFound, "resource not found"),
	Conflict(CodeConflict, "resource already exists"),
	InvalidInput(CodeInvalidInput, "invalid input"),
	Unauthorized(CodeUnauthorized, "authentication failed"),
	Forbidden(CodeForbidden, "access denied"),
	RateLimited(CodeTooManyRequests, "too many requests, please try again later"),
}

// FromSentinel upgrades a coarse sentinel into a full domain error.
func FromSentinel(err error) (*Error, bool) {
	for _, e := range sentinelFallbacks {
		if errors.Is(err, e.Kind.sentinel()) {
			return e, true
		}
	}
	return nil, false
}
