package postgres

import (
	"context"
	"errors"
	"fmt"
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

// categoryListRow: the two sources scanned as one. A master row has no id of
// the account's, so the columns it cannot fill come back NULL.
type categoryListRow struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Name      string
	Type      string
	Icon      string
	Color     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	IsOwn     bool
	IsBuiltIn bool
}

/*
The categories screen reads the same two sources the pickers do.

A picker offering a bucket the list never mentions is a screen disagreeing with
itself: somebody files a transaction under "Housing", then goes looking for
Housing in their categories and finds nothing. So the shared master rows appear
here too, marked as not yet the account's.

They are shown, not owned. A master row carries no timestamps and cannot be
deleted — there is nothing to delete. Editing one adopts it first, which is the
same swap the writers make, so a person who changes its colour simply ends up
holding it.
*/
const listQuery = `
	SELECT c.id, c.user_id, c.name, c.type, c.icon, c.color,
	       c.created_at, c.updated_at, TRUE AS is_own,
	       EXISTS (
		   SELECT 1 FROM master_categories m2
		    WHERE m2.type = c.type AND LOWER(m2.name) = LOWER(c.name)
	       ) AS is_built_in
	FROM categories c
	WHERE c.user_id = ? AND c.deleted_at IS NULL

	UNION ALL

	SELECT m.id, NULL, m.name, m.type, m.icon, m.color,
	       NULL, NULL, FALSE AS is_own, TRUE AS is_built_in
	FROM master_categories m
	WHERE NOT EXISTS (
		SELECT 1 FROM categories c2
		WHERE c2.user_id = ? AND c2.deleted_at IS NULL
		  AND c2.type = m.type AND LOWER(c2.name) = LOWER(m.name)
	)

	ORDER BY is_own DESC, created_at ASC NULLS LAST, name ASC`

func (r *categoryRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	var rows []categoryListRow
	if err := dbFrom(ctx, r.db).Raw(listQuery, userID, userID).Scan(&rows).Error; err != nil {
		return nil, categoryErrors.wrap("list categories", err)
	}

	categories := make([]domain.Category, 0, len(rows))
	for _, row := range rows {
		category := domain.Category{
			ID:        row.ID,
			Name:      row.Name,
			Type:      row.Type,
			Icon:      row.Icon,
			Color:     row.Color,
			IsOwn:     row.IsOwn,
			IsBuiltIn: row.IsBuiltIn,
		}
		if row.UserID != nil {
			category.UserID = *row.UserID
		}
		if row.CreatedAt != nil {
			category.CreatedAt = *row.CreatedAt
		}
		if row.UpdatedAt != nil {
			category.UpdatedAt = *row.UpdatedAt
		}
		categories = append(categories, category)
	}
	return categories, nil
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
			"name":       category.Name,
			"type":       category.Type,
			"icon":       category.Icon,
			"color":      category.Color,
			"updated_by": category.UpdatedBy,
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

// adoptMaster: the account's own copy of a master row, created if it has none.
//
// ON CONFLICT covers the account that already holds a category of that name and
// direction — made by hand, or adopted a moment ago from another tab. Nothing is
// overwritten in that case; the row it already has is the answer.
const adoptMaster = `
	INSERT INTO categories (id, user_id, name, type, icon, color, created_by, updated_by)
	SELECT gen_random_uuid(), ?, m.name, m.type, m.icon, m.color, ?, ?
	FROM master_categories m
	WHERE m.id = ?
	ON CONFLICT DO NOTHING`

/*
ResolveForUser turns the id a picker handed back into a category the account owns.

The options list is the account's own categories together with the master rows it
has not taken up yet, so an id arriving here may name either. Own first: a master
row and a category never share an id, and the lookup is already scoped by user, so
the order costs nothing but decides the common case in one read.

Adoption is idempotent on purpose. The list somebody picked from is a snapshot,
and the same master row may be adopted twice — from another tab, or by a retry
after a save failed further down. Running this twice has to hand back the same
category rather than a duplicate or an error.
*/
func (r *categoryRepository) ResolveForUser(
	ctx context.Context, userID, id uuid.UUID,
) (*domain.Category, error) {
	own, err := r.GetByID(ctx, id, userID)
	if err == nil {
		return own, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	if err := dbFrom(ctx, r.db).Exec(adoptMaster, userID, userID, userID, id).Error; err != nil {
		return nil, categoryErrors.wrap("adopt master category", err)
	}

	// Read back by what the row was made from rather than by a returned id: the
	// insert states nothing when it conflicts, and that is exactly the case where
	// the account already had one.
	var row model.CategoryModel
	err = dbFrom(ctx, r.db).Model(&model.CategoryModel{}).
		Joins("JOIN master_categories m ON m.type = categories.type AND LOWER(m.name) = LOWER(categories.name)").
		Where("m.id = ? AND categories.user_id = ?", id, userID).
		First(&row).Error
	if err != nil {
		// Neither the account's own nor a master row: the id names no category.
		return nil, categoryErrors.wrap("resolve category", err)
	}
	return row.ToDomain(), nil
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

/*
OptionsCategoryType: everything a picker may offer. No types means all of them.

Two sources, not one. The account's own categories, and the master rows it has not
taken up yet — an account that has never recorded anything owns nothing, and a
picker offering it nothing is a dead end. A master row drops off the list the
moment the account holds a category of the same name and direction, so the same
bucket is never offered twice.

They are matched by name because there is no link to match on: a category keeps no
trace of the master row it was adopted from. Rename an adopted one and its master
row comes back on offer, which is the honest reading — a different name is a
different bucket.

The account's own come first, oldest first, as they did when this read one table;
the master rows follow, alphabetically, having no creation date of the account's.
*/
const optionsQuery = `
	SELECT c.id, c.name, c.type, c.created_at
	FROM categories c
	WHERE c.user_id = ? AND c.deleted_at IS NULL%s

	UNION ALL

	SELECT m.id, m.name, m.type, NULL AS created_at
	FROM master_categories m
	WHERE NOT EXISTS (
		SELECT 1 FROM categories c2
		WHERE c2.user_id = ? AND c2.deleted_at IS NULL
		  AND c2.type = m.type AND LOWER(c2.name) = LOWER(m.name)
	)%s

	ORDER BY created_at ASC NULLS LAST, name ASC`

func (r *categoryRepository) OptionsCategoryType(
	ctx context.Context, userID uuid.UUID, types []string,
) ([]domain.OptionCategoryType, error) {
	ownFilter, masterFilter := "", ""
	args := []any{userID}
	if len(types) > 0 {
		ownFilter = " AND c.type IN (?)"
		args = append(args, types)
	}
	args = append(args, userID)
	if len(types) > 0 {
		masterFilter = " AND m.type IN (?)"
		args = append(args, types)
	}

	var rows []domain.OptionCategoryType
	query := fmt.Sprintf(optionsQuery, ownFilter, masterFilter)
	if err := dbFrom(ctx, r.db).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, categoryErrors.wrap("options category type", err)
	}
	return rows, nil
}
