package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type AuditLogRepository struct {
	mock.Mock
}

var _ domain.AuditLogRepository = (*AuditLogRepository)(nil)

// Create is best effort in the services, so an unconfigured mock stays silent.
func (m *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if len(m.ExpectedCalls) == 0 {
		return nil
	}
	return m.Called(ctx, log).Error(0)
}

func (m *AuditLogRepository) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.AuditLog), args.Int(1), args.Error(2)
}

func (m *AuditLogRepository) Overview(ctx context.Context, window time.Duration) (domain.AuditLogOverview, error) {
	args := m.Called(ctx, window)
	return args.Get(0).(domain.AuditLogOverview), args.Error(1)
}

func (m *AuditLogRepository) DistinctActors(ctx context.Context) ([]domain.AuditActorOption, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.AuditActorOption), args.Error(1)
}

func (m *AuditLogRepository) ListRows(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.AuditLog), args.Error(1)
}
