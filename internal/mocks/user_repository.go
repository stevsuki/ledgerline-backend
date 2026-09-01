// Package mocks: stub domain interfaces for unit tests, one file per interface.
package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type UserRepository struct {
	mock.Mock
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (m *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *UserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return m.Called(ctx, userID, passwordHash).Error(0)
}

func (m *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// expects reports whether a test set an expectation for this method.
func expects(m *mock.Mock, method string) bool {
	for _, call := range m.ExpectedCalls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// The three lockout calls below stay silent unless a test asks for them.
func (m *UserRepository) IncrementFailedLogin(ctx context.Context, userID uuid.UUID) (int, error) {
	if !expects(&m.Mock, "IncrementFailedLogin") {
		return 1, nil
	}
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *UserRepository) LockUntil(ctx context.Context, userID uuid.UUID, until time.Time) error {
	if !expects(&m.Mock, "LockUntil") {
		return nil
	}
	return m.Called(ctx, userID, until).Error(0)
}

func (m *UserRepository) ClearFailedLogins(ctx context.Context, userID uuid.UUID) error {
	if !expects(&m.Mock, "ClearFailedLogins") {
		return nil
	}
	return m.Called(ctx, userID).Error(0)
}
