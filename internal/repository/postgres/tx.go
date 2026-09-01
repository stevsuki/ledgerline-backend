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

// Do runs fn in a transaction. Nesting is safe: an inner Do joins the outer
// transaction as a savepoint instead of opening a second, independent one.
func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return dbFrom(ctx, m.db).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFrom returns the transaction carried by ctx, falling back to the pooled
// connection. Every repository goes through this: without it a call made inside
// TxManager.Do would quietly write on another connection, outside the
// transaction it is supposed to be part of.
func dbFrom(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}
