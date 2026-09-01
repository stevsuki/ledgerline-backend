package domain

import "context"

// TxManager runs a unit of work inside a single database transaction.
//
// The ctx passed to fn carries the transaction, and every repository picks it
// up on its own, so callers never touch the driver. Returning an error from fn
// rolls the whole unit back.
//
// Keep external side effects (email, HTTP calls) outside fn: a transaction can
// still roll back after they have already happened.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
