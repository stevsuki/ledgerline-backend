package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type TransactionRepository struct {
	mock.Mock
}

var _ domain.TransactionRepository = (*TransactionRepository)(nil)

func (m *TransactionRepository) List(
	ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter,
) ([]domain.Transaction, int, error) {
	args := m.Called(ctx, userID, filter)
	transactions, _ := args.Get(0).([]domain.Transaction)
	return transactions, args.Int(1), args.Error(2)
}

func (m *TransactionRepository) Summary(
	ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter,
) (domain.TransactionSummary, error) {
	args := m.Called(ctx, userID, filter)
	summary, _ := args.Get(0).(domain.TransactionSummary)
	return summary, args.Error(1)
}

func (m *TransactionRepository) Overview(
	ctx context.Context, userID uuid.UUID, current, previous domain.TimeRange,
) (domain.TransactionOverview, error) {
	args := m.Called(ctx, userID, current, previous)
	overview, _ := args.Get(0).(domain.TransactionOverview)
	return overview, args.Error(1)
}

func (m *TransactionRepository) Trend(
	ctx context.Context,
	userID uuid.UUID,
	window domain.TimeRange,
	bucket domain.TransactionTrendBucket,
) ([]domain.TransactionTrendPoint, error) {
	args := m.Called(ctx, userID, window, bucket)
	points, _ := args.Get(0).([]domain.TransactionTrendPoint)
	return points, args.Error(1)
}

func (m *TransactionRepository) GetByID(
	ctx context.Context, id, userID uuid.UUID,
) (*domain.Transaction, error) {
	args := m.Called(ctx, id, userID)
	transaction, _ := args.Get(0).(*domain.Transaction)
	return transaction, args.Error(1)
}

func (m *TransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	return m.Called(ctx, transaction).Error(0)
}

func (m *TransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	return m.Called(ctx, transaction).Error(0)
}

func (m *TransactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *TransactionRepository) ExistsByCategory(
	ctx context.Context, categoryID, userID uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, categoryID, userID)
	return args.Bool(0), args.Error(1)
}
