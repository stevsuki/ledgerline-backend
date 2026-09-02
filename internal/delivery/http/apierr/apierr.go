// Package apierr: the single place an error becomes an HTTP response.
// Handlers and middleware both go through Write, so every failure leaves the
// same envelope: {success, message, code, errors, request_id}.
package apierr

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/logger"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
	"github.com/stevensuki/ledgerline-backend/pkg/validator"
)

// StatusClientClosedRequest: nginx's code for "the caller hung up"; log-only, never sent.
const StatusClientClosedRequest = 499

// Status maps a domain kind to its HTTP status; the only table that does so.
func Status(kind domain.Kind) int {
	switch kind {
	case domain.KindNotFound:
		return http.StatusNotFound
	case domain.KindConflict:
		return http.StatusConflict
	case domain.KindInvalidInput:
		return http.StatusBadRequest
	case domain.KindUnauthorized:
		return http.StatusUnauthorized
	case domain.KindForbidden:
		return http.StatusForbidden
	case domain.KindRateLimited:
		return http.StatusTooManyRequests
	case domain.KindTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

// Write turns any error into the response for it.
func Write(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if domainErr, ok := domain.AsError(err); ok {
		writeDomain(c, domainErr)
		return
	}

	// Nobody is left to answer; abort quietly instead of logging a fake 500.
	if errors.Is(err, context.Canceled) {
		logger.FromContext(c.Request.Context()).Debug("request cancelled by client")
		c.AbortWithStatus(StatusClientClosedRequest)
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		logger.FromContext(c.Request.Context()).Warn("request timed out", slog.Any("error", err))
		writeDomain(c, domain.ErrTimeout)
		return
	}

	if errors.Is(err, pagination.ErrInvalidSort) {
		response.Fail(c, http.StatusUnprocessableEntity, domain.CodeValidation,
			"the query parameters are invalid",
			[]response.FieldError{{Field: "sort", Message: sortDetail(err)}})
		return
	}

	if fallback, ok := domain.FromSentinel(err); ok {
		writeDomain(c, fallback.WithCause(err))
		return
	}

	// Unexpected: the client learns nothing, the log learns everything.
	logger.FromContext(c.Request.Context()).Error("unhandled error", slog.Any("error", err))
	_ = c.Error(err)
	response.Fail(c, http.StatusInternalServerError, domain.CodeInternal,
		"an internal server error occurred", nil)
}

// WriteBind reports a binding or validation failure with per-field detail.
func WriteBind(c *gin.Context, err error) {
	if fieldErrors := validator.Translate(err); len(fieldErrors) > 0 {
		response.Fail(c, http.StatusUnprocessableEntity, domain.CodeValidation,
			"the submitted data is invalid", fieldErrors)
		return
	}
	response.Fail(c, http.StatusBadRequest, domain.CodeBadRequest,
		"the request format cannot be processed", nil)
}

func writeDomain(c *gin.Context, e *domain.Error) {
	status := Status(e.Kind)

	// A domain error that still maps to 500 is a bug, not an expected outcome.
	if status == http.StatusInternalServerError {
		logger.FromContext(c.Request.Context()).Error("internal domain error", slog.Any("error", e))
		_ = c.Error(e)
	}
	if e.RetryAfter > 0 {
		c.Header("Retry-After", strconv.Itoa(int(math.Ceil(e.RetryAfter.Seconds()))))
	}

	var fields []response.FieldError
	if e.Field != "" {
		fields = []response.FieldError{{Field: e.Field, Message: e.Message}}
	}
	response.Fail(c, status, e.Code, e.Message, fields)
}

// sortDetail drops the sentinel prefix so the field message reads on its own.
func sortDetail(err error) string {
	return strings.TrimPrefix(err.Error(), pagination.ErrInvalidSort.Error()+": ")
}
