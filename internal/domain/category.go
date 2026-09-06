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

// CategoryTypesForSlug: which types a slug may show. A nil slice means every
// type, and ok is false for a slug nothing serves.
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

// Category: a spending or income category owned by one user. MasterCategoryID
// names the master row it derives from, and is uuid.Nil when it derives from none.
type Category struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	MasterCategoryID uuid.UUID
	Name             string
	Type             string
	// Icon is a key from the client's own sprite, Color a step of its chart
	// ramp. "" means the category has none of its own and the reader resolves
	// one. Neither is checked beyond its length, exactly as wallets.icon is
	// not: the vocabulary belongs to whatever draws them.
	Icon      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
}

type CategoryRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	// SeedDefaults gives a brand new user one category per master row.
	SeedDefaults(ctx context.Context, userID uuid.UUID) error
	// OptionsCategoryType takes the types to show; an empty slice means all.
	OptionsCategoryType(ctx context.Context, userID uuid.UUID, types []string) ([]OptionCategoryType, error)
}

type CreateCategoryInput struct {
	Name             string
	MasterCategoryID uuid.UUID
	Type             string
	Icon             string
	Color            string
}

// UpdateCategoryInput: pointers so partial updates are detectable; nil leaves
// the field as it was.
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
