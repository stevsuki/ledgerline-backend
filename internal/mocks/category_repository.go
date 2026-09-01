package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type CategoryRepository struct {
	mock.Mock
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

func (m *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *CategoryRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Category, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}

func (m *CategoryRepository) List(ctx context.Context, filter domain.CategoryFilter) ([]domain.Category, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]domain.Category), args.Int(1), args.Error(2)
}

func (m *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *CategoryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *CategoryRepository) ExistsByName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	args := m.Called(ctx, userID, name)
	return args.Bool(0), args.Error(1)
}
