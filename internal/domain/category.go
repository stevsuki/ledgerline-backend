package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// The two values categories_type_check allows.
const (
	CategoryTypeIncome  = "income"
	CategoryTypeExpense = "expense"
)

func ValidCategoryType(t string) bool {
	return t == CategoryTypeIncome || t == CategoryTypeExpense
}

// The screens that ask for category options; each one may use its own types.
const (
	CategoryOptionSlugFilter = "filter"
	CategoryOptionSlugBudget = "budget"
)

// CategoryTypesForSlug: types a slug may show; nil means every type.
func CategoryTypesForSlug(slug string) (types []string, ok bool) {
	switch slug {
	case CategoryOptionSlugFilter:
		return nil, true
	case CategoryOptionSlugBudget:
		return []string{CategoryTypeExpense}, true
	default:
		return nil, false
	}
}

// Category: a spending or income category owned by one user.
type Category struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	MasterCategoryID uuid.UUID
	Name             string
	Type             string
	Icon             string
	Color            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CreatedBy        *uuid.UUID
	UpdatedBy        *uuid.UUID
	DeletedBy        *uuid.UUID
}

type CategoryRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	SeedDefaults(ctx context.Context, userID uuid.UUID) error
	OptionsCategoryType(ctx context.Context, userID uuid.UUID, types []string) ([]OptionCategoryType, error)
}

type CreateCategoryInput struct {
	Name             string
	MasterCategoryID uuid.UUID
	Type             string
	Icon             string
	Color            string
}

// UpdateCategoryInput: pointers so a partial update is detectable.
type UpdateCategoryInput struct {
	Name             *string
	MasterCategoryID *uuid.UUID
	Type             *string
	Icon             *string
	Color            *string
}

type CategoryService interface {
	List(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	Create(ctx context.Context, userID uuid.UUID, category CreateCategoryInput) (*Category, error)
	Update(ctx context.Context, userID, id uuid.UUID, category UpdateCategoryInput) (*Category, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	OptionsCategoryType(ctx context.Context, userID uuid.UUID, slug string) ([]OptionCategoryType, error)
}

type OptionCategoryType struct {
	ID   uuid.UUID
	Name string
}
