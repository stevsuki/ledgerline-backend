package domain

import (
	"context"

	"github.com/google/uuid"
)

// MasterCategory: a category every account can pick without owning it yet.
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
