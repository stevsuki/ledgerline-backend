package domain_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

func TestErrorMatchesCoarseSentinel(t *testing.T) {
	t.Parallel()

	err := domain.Conflict(domain.CodeUserEmailTaken, "email is already registered")

	assert.ErrorIs(t, err, domain.ErrConflict)
	assert.NotErrorIs(t, err, domain.ErrNotFound)
}

func TestErrorMatchesThroughWrapping(t *testing.T) {
	t.Parallel()

	wrapped := fmt.Errorf("create user: %w", domain.NotFound(domain.CodeUserNotFound, "user not found"))

	assert.ErrorIs(t, wrapped, domain.ErrNotFound)

	got, ok := domain.AsError(wrapped)
	require.True(t, ok)
	assert.Equal(t, domain.CodeUserNotFound, got.Code)
}

func TestErrorMatchesByCodeAcrossCopies(t *testing.T) {
	t.Parallel()

	copied := domain.ErrInvalidOTP.WithCause(errors.New("bcrypt mismatch"))

	assert.ErrorIs(t, copied, domain.ErrInvalidOTP)
	assert.NotErrorIs(t, copied, domain.ErrTokenExpired)
}

func TestWithersDoNotMutateTheTemplate(t *testing.T) {
	t.Parallel()

	locked := domain.ErrAccountLocked.WithRetryAfter(5 * time.Minute).WithField("email")

	assert.Equal(t, 5*time.Minute, locked.RetryAfter)
	assert.Equal(t, "email", locked.Field)
	assert.Zero(t, domain.ErrAccountLocked.RetryAfter, "the shared template must stay untouched")
	assert.Empty(t, domain.ErrAccountLocked.Field)
}

func TestCauseStaysOutOfTheClientMessage(t *testing.T) {
	t.Parallel()

	err := domain.NotFound(domain.CodeUserNotFound, "user not found").
		WithCause(errors.New("get user: sql: no rows in result set"))

	assert.Equal(t, "user not found", err.Message, "Message is what the client sees")
	assert.Contains(t, err.Error(), "sql: no rows", "Error() is what the log sees")
}

func TestFromSentinel(t *testing.T) {
	t.Parallel()

	got, ok := domain.FromSentinel(fmt.Errorf("wrapped: %w", domain.ErrForbidden))
	require.True(t, ok)
	assert.Equal(t, domain.CodeForbidden, got.Code)
	assert.Equal(t, domain.KindForbidden, got.Kind)

	_, ok = domain.FromSentinel(errors.New("something else"))
	assert.False(t, ok)
}
