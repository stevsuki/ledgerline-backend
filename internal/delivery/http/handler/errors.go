// Package handler: HTTP adapter (parse request, call service, map response).
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/logger"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
	"github.com/stevensuki/ledgerline-backend/pkg/validator"
)

// handleError: central mapping of domain errors -> HTTP status.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
	case errors.Is(err, domain.ErrConflict):
		response.Fail(c, http.StatusConflict, "CONFLICT", err.Error(), nil)
	case errors.Is(err, pagination.ErrInvalidSort):
		response.Fail(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "invalid sort parameter",
			[]response.FieldError{{Field: "sort", Message: err.Error()}})
	case errors.Is(err, domain.ErrInvalidInput):
		response.Fail(c, http.StatusBadRequest, "INVALID_INPUT", err.Error(), nil)
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Fail(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password", nil)
	case errors.Is(err, domain.ErrTokenExpired):
		response.Fail(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "the token has expired", nil)
	case errors.Is(err, domain.ErrTokenInvalid), errors.Is(err, domain.ErrUnauthorized):
		response.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication failed", nil)
	case errors.Is(err, domain.ErrForbidden):
		response.Fail(c, http.StatusForbidden, "FORBIDDEN", "access denied", nil)
	case errors.Is(err, domain.ErrTooManyRequests):
		response.Fail(c, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "too many requests, please try again later", nil)
	case errors.Is(err, domain.ErrInvalidOTP):
		response.Fail(c, http.StatusBadRequest, "INVALID_OTP", "the provided OTP is invalid", nil)
	case errors.Is(err, domain.ErrMaxAttemptsExceeded):
		response.Fail(c, http.StatusTooManyRequests, "MAX_ATTEMPTS_EXCEEDED", "maximum OTP attempts exceeded", nil)
	case errors.Is(err, domain.ErrAccountLocked):
		response.Fail(c, http.StatusTooManyRequests, "ACCOUNT_LOCKED",
			"too many failed sign-in attempts, please try again later", nil)
	default:
		// Unexpected error: details go to the log only.
		logger.FromContext(c.Request.Context()).Error("unhandled error", slog.Any("error", err))
		_ = c.Error(err)
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR",
			"an internal server error occurred", nil)
	}
}

// handleBindError: binding/validation failure -> 422 + per-field detail.
func handleBindError(c *gin.Context, err error) {
	if fieldErrors := validator.Translate(err); len(fieldErrors) > 0 {
		response.Fail(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR",
			"the submitted data is invalid", fieldErrors)
		return
	}
	response.Fail(c, http.StatusBadRequest, "BAD_REQUEST",
		"the request format cannot be processed", nil)
}
