package apierr_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/delivery/http/apierr"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/response"
)

func writeErr(t *testing.T, err error) (*httptest.ResponseRecorder, response.Error) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Writer.Header().Set(response.HeaderRequestID, "req-123")

	apierr.Write(c, err)

	var body response.Error
	if rec.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	}
	return rec, body
}

func TestWriteDomainError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"not found", domain.NotFound(domain.CodeUserNotFound, "user not found"), http.StatusNotFound, domain.CodeUserNotFound},
		{"conflict", domain.Conflict(domain.CodeUserEmailTaken, "taken"), http.StatusConflict, domain.CodeUserEmailTaken},
		{"forbidden", domain.ErrAccessDenied, http.StatusForbidden, domain.CodeForbidden},
		{"rate limited", domain.ErrAccountLocked, http.StatusTooManyRequests, domain.CodeAccountLocked},
		{"timeout", context.DeadlineExceeded, http.StatusGatewayTimeout, domain.CodeTimeout},
		{"bare sentinel", domain.ErrNotFound, http.StatusNotFound, domain.CodeNotFound},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, domain.CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec, body := writeErr(t, tt.err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantCode, body.Code)
			assert.False(t, body.Success)
			assert.Equal(t, "req-123", body.RequestID)
		})
	}
}

// The op prefix a repository adds for the log must never reach the client.
func TestWriteKeepsInternalDetailOutOfTheBody(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("create user: %w",
		domain.Conflict(domain.CodeUserEmailTaken, "email is already registered").
			WithCause(errors.New("duplicate key value violates unique constraint")))

	rec, body := writeErr(t, err)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "email is already registered", body.Message)
	assert.NotContains(t, body.Message, "create user")
	assert.NotContains(t, body.Message, "unique constraint")
}

func TestWriteFieldAndRetryAfter(t *testing.T) {
	t.Parallel()

	err := domain.InvalidInput(domain.CodeWalletInvalidType, "wallet type must be bank, ewallet, card or cash").
		WithField("type")
	_, body := writeErr(t, err)

	require.Len(t, body.Errors, 1)
	assert.Equal(t, "type", body.Errors[0].Field)

	rec, _ := writeErr(t, domain.ErrAccountLocked.WithRetryAfter(90*time.Second))
	assert.Equal(t, "90", rec.Header().Get("Retry-After"))
}

func TestWriteClientDisconnectIsNotAServerError(t *testing.T) {
	t.Parallel()

	rec, _ := writeErr(t, context.Canceled)
	assert.Equal(t, apierr.StatusClientClosedRequest, rec.Code)
	assert.Empty(t, rec.Body.String())
}
