package domain

import (
	"context"

	"github.com/google/uuid"
)

type MasterCategory struct {
	ID   uuid.UUID
	Name string
}

type MasterCategoryRepository interface {
	List(ctx context.Context) ([]MasterCategory, error)
}
