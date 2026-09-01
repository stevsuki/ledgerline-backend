package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// txKey is private so the transaction can only be put into a context here.
type txKey struct{}

type txManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) domain.TxManager { return &txManager{db: db} }

// Do runs fn in a transaction; a nested Do joins the outer one as a savepoint.
func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return dbFrom(ctx, m.db).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFrom returns the transaction ctx carries, else the pool; every repository goes through it.
func dbFrom(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}
