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

// Category: a spending or income category owned by one user.
type Category struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Type      string
	Icon      string
	Color     string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	// IsOwn: set by reads that mix in the shared master rows. False marks a row
	// the account can use but does not hold yet — it has no id of its own here,
	// no timestamps, and nothing to delete. Writes never set it.
	IsOwn bool
	// IsBuiltIn: this name and direction are one of the shared master rows.
	//
	// Certain while the row still is one. Once taken up it is a copy carrying no
	// trace of where it came from, so this becomes a reading of the same rule the
	// list already applies — a name and a direction are what make a bucket, which
	// is how a master row knows to stop offering itself. Rename an adopted one
	// and it stops counting as built in, which is the same answer the picker
	// gives when its master row reappears there.
	IsBuiltIn bool
}

type CategoryRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	// ResolveForUser: the id as one of the account's own categories, adopting the
	// master row it names when the account has no category from it yet.
	ResolveForUser(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	OptionsCategoryType(ctx context.Context, userID uuid.UUID, types []string) ([]OptionCategoryType, error)
}

type CreateCategoryInput struct {
	Name  string
	Type  string
	Icon  string
	Color string
}

// UpdateCategoryInput: pointers so a partial update is detectable.
type UpdateCategoryInput struct {
	Name  *string
	Type  *string
	Icon  *string
	Color *string
}

type CategoryService interface {
	List(ctx context.Context, userID uuid.UUID) ([]Category, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Category, error)
	Create(ctx context.Context, userID uuid.UUID, category CreateCategoryInput) (*Category, error)
	Update(ctx context.Context, userID, id uuid.UUID, category UpdateCategoryInput) (*Category, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	OptionsCategoryType(ctx context.Context, userID uuid.UUID, categoryType string) ([]OptionCategoryType, error)
}

// OptionCategoryType: one pickable category. The list is the account's own
// categories together with the master rows it has not taken up yet, so an id
// here may name either — which is what ResolveForUser settles on the way in.
type OptionCategoryType struct {
	ID   uuid.UUID
	Name string
	Type string
}
