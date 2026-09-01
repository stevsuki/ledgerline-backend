package mocks

import (
	"context"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// TxManager runs the unit of work inline. Transaction behaviour belongs to the
// database, so service tests only need the callback to be executed.
type TxManager struct{}

var _ domain.TxManager = (*TxManager)(nil)

func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
