package domain

import (
	"context"

	"github.com/google/uuid"
)

// MasterCategoryIDOthers: the bucket a category falls back to, seeded in 000023.
var MasterCategoryIDOthers = uuid.MustParse("00000000-0000-0000-0000-000000000007")

type MasterCategory struct {
	ID   uuid.UUID
	Name string
}

type MasterCategoryRepository interface {
	List(ctx context.Context) ([]MasterCategory, error)
}
