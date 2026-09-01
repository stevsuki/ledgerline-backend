package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type MenuRepository struct {
	mock.Mock
}

var _ domain.MenuRepository = (*MenuRepository)(nil)

func (m *MenuRepository) ListByRole(ctx context.Context, roleID uuid.UUID) ([]domain.Menu, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Menu), args.Error(1)
}
