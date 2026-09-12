package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type CategoryService struct {
	categoryRepo    domain.CategoryRepository
	budgetRepo      domain.BudgetRepository
	transactionRepo domain.TransactionRepository
}

func NewCategoryService(
	categoryRepo domain.CategoryRepository,
	budgetRepo domain.BudgetRepository,
	transactionRepo domain.TransactionRepository,
) domain.CategoryService {
	return &CategoryService{
		categoryRepo:    categoryRepo,
		budgetRepo:      budgetRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	return s.categoryRepo.List(ctx, userID)
}

func (s *CategoryService) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	return s.categoryRepo.GetByID(ctx, id, userID)
}

func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateCategoryInput) (*domain.Category, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.InvalidInput(domain.CodeCategoryInvalidData, "category name is required").WithField("name")
	}

	if !domain.ValidCategoryType(input.Type) {
		return nil, domain.InvalidInput(domain.CodeCategoryInvalidType, "category type must be income or expense").WithField("type")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	category := &domain.Category{
		ID:     id,
		UserID: userID,
		Name:   name,
		Type:   input.Type,
		Icon:   strings.TrimSpace(input.Icon),
		Color:  strings.TrimSpace(input.Color),
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, userID, id uuid.UUID, input domain.UpdateCategoryInput) (*domain.Category, error) {
	// The list offers shared rows beside the account's own, so a pencil may land
	// on one it does not hold yet. Adopting it here is the same swap the writers
	// make: what gets edited is always a row of the account's.
	category, err := s.categoryRepo.ResolveForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.InvalidInput(domain.CodeCategoryInvalidData, "category name must not be empty").WithField("name")
		}
		category.Name = name
	}

	if input.Type != nil {
		if !domain.ValidCategoryType(*input.Type) {
			return nil, domain.InvalidInput(domain.CodeCategoryInvalidType, "category type must be income or expense").WithField("type")
		}
		category.Type = *input.Type
	}

	if input.Icon != nil {
		category.Icon = strings.TrimSpace(*input.Icon)
	}
	if input.Color != nil {
		category.Color = strings.TrimSpace(*input.Color)
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

/*
Delete refuses while anything is filed under the category.

Removal here is a soft delete, so the ON DELETE RESTRICT on transactions and
budgets never fires — the row stays and the reference stays valid. What breaks
instead is quieter: the dashboard's category ring and the budgets list both join
categories with a not-deleted filter, while the month's totals join nothing at
all. Delete a category still carrying spending and the stat card keeps counting
it while the ring stops drawing it, with nothing on screen admitting that the
two no longer agree.

Budgets were already guarded. Transactions were not, and that was the half that
could put two different figures for the same month on one screen.
*/
func (s *CategoryService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	budgeted, err := s.budgetRepo.ExistsByCategory(ctx, id, userID)
	if err != nil {
		return err
	}
	if budgeted {
		return domain.Conflict(domain.CodeCategoryInUse,
			"a budget still limits this category; remove the budget first")
	}

	recorded, err := s.transactionRepo.ExistsByCategory(ctx, id, userID)
	if err != nil {
		return err
	}
	if recorded {
		return domain.Conflict(domain.CodeCategoryInUse,
			"transactions are still filed under this category; move or remove them first")
	}

	return s.categoryRepo.Delete(ctx, id, userID)
}

// OptionsCategoryType: what a picker may offer, narrowed to one direction when
// the caller names one. An empty type means both.
//
// Which direction a screen wants is the screen's business, not this list's — a
// budget asks for expense because that is all it can limit, and the rule that
// makes it so is enforced where budgets are written, not here.
func (s *CategoryService) OptionsCategoryType(
	ctx context.Context, userID uuid.UUID, categoryType string,
) ([]domain.OptionCategoryType, error) {
	var types []string
	if categoryType != "" {
		if !domain.ValidCategoryType(categoryType) {
			return nil, domain.InvalidInput(
				domain.CodeCategoryInvalidType, "category type must be income or expense",
			).WithField("type")
		}
		types = []string{categoryType}
	}
	return s.categoryRepo.OptionsCategoryType(ctx, userID, types)
}
