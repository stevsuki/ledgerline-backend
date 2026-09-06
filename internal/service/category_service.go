package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type CategoryService struct {
	categoryRepo domain.CategoryRepository
}

func NewCategoryService(categoryRepo domain.CategoryRepository) domain.CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
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
		ID:               id,
		UserID:           userID,
		MasterCategoryID: input.MasterCategoryID,
		Name:             name,
		Type:             input.Type,
		Icon:             strings.TrimSpace(input.Icon),
		Color:            strings.TrimSpace(input.Color),
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *CategoryService) Update(ctx context.Context, userID, id uuid.UUID, input domain.UpdateCategoryInput) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id, userID)
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

	if input.MasterCategoryID != nil {
		category.MasterCategoryID = *input.MasterCategoryID
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

func (s *CategoryService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.categoryRepo.Delete(ctx, id, userID)
}

func (s *CategoryService) OptionsCategoryType(
	ctx context.Context, userID uuid.UUID, slug string,
) ([]domain.OptionCategoryType, error) {
	types, ok := domain.CategoryTypesForSlug(slug)
	if !ok {
		return nil, domain.InvalidInput(
			domain.CodeCategoryInvalidSlug, "category option slug must be filter or budget",
		).WithField("slug")
	}
	return s.categoryRepo.OptionsCategoryType(ctx, userID, types)
}
