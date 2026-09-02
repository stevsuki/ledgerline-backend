package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

func (t CategoryType) Valid() bool {
	return t == CategoryTypeIncome || t == CategoryTypeExpense
}

// Category: a transaction category owned by one user.
type Category struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Type      CategoryType
	CreatedAt time.Time
	UpdatedAt time.Time
	// Actor ids; the owner writes their own categories, so these track UserID.
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	// Only ever set on a soft-deleted row, which no read returns yet.
	DeletedBy *uuid.UUID
}

// CategoryFilter: UserID is required so data from different users never mixes.
type CategoryFilter struct {
	UserID  uuid.UUID
	Search  string
	Type    CategoryType
	Limit   int
	Offset  int
	OrderBy string // ORDER BY clause, may only be filled via pagination.Sortable
}

type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Category, error)
	List(ctx context.Context, filter CategoryFilter) ([]Category, int, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ExistsByName(ctx context.Context, userID uuid.UUID, name string) (bool, error)
}

type CreateCategoryInput struct {
	Name string
	Type CategoryType
}

type UpdateCategoryInput struct {
	Name *string
	Type *CategoryType
}

type CategoryService interface {
	Create(ctx context.Context, userID uuid.UUID, input CreateCategoryInput) (*Category, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	List(ctx context.Context, filter CategoryFilter) ([]Category, int, error)
	Update(ctx context.Context, userID, id uuid.UUID, input UpdateCategoryInput) (*Category, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}
