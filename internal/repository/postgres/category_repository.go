package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

// defaultCategoryOrder: oldest first, id as the tie breaker.
const defaultCategoryOrder = "created_at ASC, id ASC"

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	var rows []model.CategoryModel
	err := dbFrom(ctx, r.db).
		Where("user_id = ?", userID).
		Order(defaultCategoryOrder).
		Find(&rows).Error
	if err != nil {
		return nil, categoryErrors.wrap("list categories", err)
	}
	return model.CategoriesToDomain(rows), nil
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

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	category.UpdatedBy = domain.ActorFrom(ctx)

	result := dbFrom(ctx, r.db).Model(&model.CategoryModel{}).
		Where("id = ? AND user_id = ?", category.ID, category.UserID).
		Updates(map[string]any{
			"master_category_id": model.MasterCategoryRef(category.MasterCategoryID),
			"name":               category.Name,
			"type":               category.Type,
			"icon":               category.Icon,
			"color":              category.Color,
			"updated_by":         category.UpdatedBy,
		})
	if result.Error != nil {
		return categoryErrors.wrap("update category", result.Error)
	}
	if result.RowsAffected == 0 {
		return categoryErrors.wrap("update category", gorm.ErrRecordNotFound)
	}

	var updated model.CategoryModel
	err := dbFrom(ctx, r.db).Select("updated_at").First(&updated, "id = ?", category.ID).Error
	if err == nil {
		category.UpdatedAt = updated.UpdatedAt
	}
	return nil
}

// SeedDefaults gives a brand new user one category per master row.
func (r *categoryRepository) SeedDefaults(ctx context.Context, userID uuid.UUID) error {
	err := dbFrom(ctx, r.db).Exec(`
		INSERT INTO categories (id, user_id, master_category_id, name, type, created_by, updated_by)
		SELECT gen_random_uuid(), ?, m.id, m.name, ?, ?, ?
		FROM master_categories m
	`, userID, domain.CategoryTypeExpense, userID, userID).Error
	if err != nil {
		return categoryErrors.wrap("seed default categories", err)
	}
	return nil
}

// Delete: soft delete stamped with who did it, in one statement.
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

// OptionsCategoryType: id and name only, for the pickers. No types means all of them.
func (r *categoryRepository) OptionsCategoryType(
	ctx context.Context, userID uuid.UUID, types []string,
) ([]domain.OptionCategoryType, error) {
	query := dbFrom(ctx, r.db).Model(&model.CategoryModel{}).
		Select("id", "name").
		Where("user_id = ?", userID)
	if len(types) > 0 {
		query = query.Where("type IN ?", types)
	}

	var rows []domain.OptionCategoryType
	if err := query.Order(defaultCategoryOrder).Find(&rows).Error; err != nil {
		return nil, categoryErrors.wrap("options category type", err)
	}
	return rows, nil
}
