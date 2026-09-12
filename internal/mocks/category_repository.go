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

func (m *CategoryRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	args := m.Called(ctx, userID)
	categories, _ := args.Get(0).([]domain.Category)
	return categories, args.Error(1)
}

func (m *CategoryRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Category, error) {
	args := m.Called(ctx, id, userID)
	category, _ := args.Get(0).(*domain.Category)
	return category, args.Error(1)
}

func (m *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *CategoryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *CategoryRepository) ResolveForUser(
	ctx context.Context, userID, id uuid.UUID,
) (*domain.Category, error) {
	args := m.Called(ctx, userID, id)
	category, _ := args.Get(0).(*domain.Category)
	return category, args.Error(1)
}

func (m *CategoryRepository) OptionsCategoryType(
	ctx context.Context, userID uuid.UUID, types []string,
) ([]domain.OptionCategoryType, error) {
	args := m.Called(ctx, userID, types)
	options, _ := args.Get(0).([]domain.OptionCategoryType)
	return options, args.Error(1)
}
