package domain

import "context"

// TxManager runs fn in one transaction, carried by ctx; keep email and HTTP calls out of fn.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
