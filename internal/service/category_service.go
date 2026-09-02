package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type categoryService struct {
	repo domain.CategoryRepository
}

func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateCategoryInput) (*domain.Category, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.InvalidInput(domain.CodeCategoryInvalidData, "category name is required").WithField("name")
	}
	if !input.Type.Valid() {
		return nil, domain.InvalidInput(domain.CodeCategoryInvalidType, "category type must be income or expense").WithField("type")
	}

	exists, err := s.repo.ExistsByName(ctx, userID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.Conflict(domain.CodeCategoryNameTaken, "a category with that name already exists").WithField("name")
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
	}
	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *categoryService) List(ctx context.Context, filter domain.CategoryFilter) ([]domain.Category, int, error) {
	if filter.Type != "" && !filter.Type.Valid() {
		return nil, 0, domain.InvalidInput(domain.CodeCategoryInvalidType, "category type must be income or expense").WithField("type")
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return s.repo.List(ctx, filter)
}

func (s *categoryService) Update(ctx context.Context, userID, id uuid.UUID, input domain.UpdateCategoryInput) (*domain.Category, error) {
	category, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.InvalidInput(domain.CodeCategoryInvalidData, "category name must not be empty").WithField("name")
		}
		// Check for duplicates only when the name actually changed.
		if !strings.EqualFold(name, category.Name) {
			exists, err := s.repo.ExistsByName(ctx, userID, name)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, domain.Conflict(domain.CodeCategoryNameTaken, "a category with that name already exists").WithField("name")
			}
		}
		category.Name = name
	}

	if input.Type != nil {
		if !input.Type.Valid() {
			return nil, domain.InvalidInput(domain.CodeCategoryInvalidType, "category type must be income or expense").WithField("type")
		}
		category.Type = *input.Type
	}

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}
