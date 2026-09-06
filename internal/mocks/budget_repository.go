package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type BudgetRepository struct {
	mock.Mock
}

var _ domain.BudgetRepository = (*BudgetRepository)(nil)

func (m *BudgetRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	args := m.Called(ctx, userID)
	budgets, _ := args.Get(0).([]domain.Budget)
	return budgets, args.Error(1)
}

func (m *BudgetRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Budget, error) {
	args := m.Called(ctx, id, userID)
	budget, _ := args.Get(0).(*domain.Budget)
	return budget, args.Error(1)
}

func (m *BudgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	return m.Called(ctx, budget).Error(0)
}

func (m *BudgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	return m.Called(ctx, budget).Error(0)
}

func (m *BudgetRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *BudgetRepository) Usage(ctx context.Context, userID uuid.UUID) ([]domain.BudgetUsage, error) {
	args := m.Called(ctx, userID)
	usages, _ := args.Get(0).([]domain.BudgetUsage)
	return usages, args.Error(1)
}

func (m *BudgetRepository) ExistsByCategory(ctx context.Context, categoryID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, categoryID, userID)
	return args.Bool(0), args.Error(1)
}
