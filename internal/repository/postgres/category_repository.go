package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

// defaultCategoryOrder: fallback when the filter arrives without OrderBy.
const defaultCategoryOrder = "created_at DESC, id ASC"

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	actor := domain.ActorFrom(ctx)
	category.CreatedBy, category.UpdatedBy = actor, actor

	row := model.CategoryFromDomain(category)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return categoryErrors.wrap("create category", err)
	}

	category.CreatedAt = row.CreatedAt
	category.UpdatedAt = row.UpdatedAt
	return nil
}

// GetByID always includes user_id so other users cannot reach this data.
func (r *categoryRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Category, error) {
	var row model.CategoryModel
	err := dbFrom(ctx, r.db).First(&row, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		return nil, categoryErrors.wrap("get category", err)
	}
	return row.ToDomain(), nil
}

func (r *categoryRepository) List(ctx context.Context, filter domain.CategoryFilter) ([]domain.Category, int, error) {
	query := dbFrom(ctx, r.db).Model(&model.CategoryModel{}).Where("user_id = ?", filter.UserID)

	if filter.Search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(filter.Search)+"%")
	}
	if filter.Type != "" {
		query = query.Where("type = ?", string(filter.Type))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, categoryErrors.wrap("count categories", err)
	}
	if total == 0 {
		return []domain.Category{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultCategoryOrder
	}

	var rows []model.CategoryModel
	err := query.Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&rows).Error
	if err != nil {
		return nil, 0, categoryErrors.wrap("list categories", err)
	}
	return model.CategoriesToDomain(rows), int(total), nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	category.UpdatedBy = domain.ActorFrom(ctx)

	result := dbFrom(ctx, r.db).
		Model(&model.CategoryModel{}).
		Where("id = ? AND user_id = ?", category.ID, category.UserID).
		Updates(map[string]any{
			"name":       category.Name,
			"type":       string(category.Type),
			"updated_by": category.UpdatedBy,
		})
	if result.Error != nil {
		return categoryErrors.wrap("update category", result.Error)
	}
	if result.RowsAffected == 0 {
		return categoryErrors.wrap("update category", gorm.ErrRecordNotFound)
	}

	var updated model.CategoryModel
	err := dbFrom(ctx, r.db).Select("updated_at").
		First(&updated, "id = ?", category.ID).Error
	if err == nil {
		category.UpdatedAt = updated.UpdatedAt
	}
	return nil
}

// Delete: soft delete stamped with who did it, in one statement. See userRepository.Delete.
func (r *categoryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result := dbFrom(ctx, r.db).
		Model(&model.CategoryModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		})
	if result.Error != nil {
		return categoryErrors.wrap("delete category", result.Error)
	}
	if result.RowsAffected == 0 {
		return categoryErrors.wrap("delete category", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *categoryRepository) ExistsByName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	var count int64
	err := dbFrom(ctx, r.db).Model(&model.CategoryModel{}).
		Where("user_id = ? AND LOWER(name) = LOWER(?)", userID, name).
		Limit(1).Count(&count).Error
	if err != nil {
		return false, categoryErrors.wrap("check category name", err)
	}
	return count > 0, nil
}
