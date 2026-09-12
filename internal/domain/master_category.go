package domain

import (
	"context"

	"github.com/google/uuid"
)

// MasterCategory: a category every account can pick without owning it yet.
//
// The rows are offered alongside the account's own in the options list. Picking
// one adopts it — a category of the account's own is created from this row's
// name, type, icon and colour — and from then on the account's copy is what is
// read. Nothing links the two afterwards.
type MasterCategory struct {
	ID    uuid.UUID
	Name  string
	Type  string
	Icon  string
	Color string
}

type MasterCategoryRepository interface {
	List(ctx context.Context) ([]MasterCategory, error)
}
