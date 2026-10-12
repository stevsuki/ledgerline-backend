package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

// defaultRecurringOrder: what is due first, paused items last.
const defaultRecurringOrder = "recurrings.is_paused ASC, recurrings.next_due_date ASC, recurrings.id ASC"

type recurringRepository struct {
	db *gorm.DB
}

func NewRecurringRepository(db *gorm.DB) domain.RecurringRepository {
	return &recurringRepository{db: db}
}

// readQuery: LEFT JOINs, so an item outlives a removed category or wallet.
func (r *recurringRepository) readQuery(ctx context.Context) *gorm.DB {
	return dbFrom(ctx, r.db).
		Model(&model.RecurringModel{}).
		Select(
			"recurrings.id",
			"recurrings.user_id",
			"recurrings.category_id",
			"recurrings.wallet_id",
			"categories.name AS category_name",
			"wallets.name AS wallet_name",
			"categories.icon",
			"recurrings.name",
			"recurrings.amount",
			"recurrings.currency",
			"recurrings.frequency",
			"recurrings.next_due_date",
			"recurrings.anchor_day",
			"recurrings.is_paused",
			"recurrings.created_at",
			"recurrings.created_by",
			"recurrings.updated_at",
			"recurrings.updated_by",
			"recurrings.deleted_by",
		).
		Joins("LEFT JOIN categories ON categories.id = recurrings.category_id").
		Joins("LEFT JOIN wallets ON wallets.id = recurrings.wallet_id")
}

func (r *recurringRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Recurring, error) {
	var rows []model.RecurringRow
	err := r.readQuery(ctx).
		Where("recurrings.user_id = ?", userID).
		Order(defaultRecurringOrder).
		Scan(&rows).Error
	if err != nil {
		return nil, recurringErrors.wrap("list recurring", err)
	}
	return model.RecurringRowsToDomain(rows), nil
}

func (r *recurringRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Recurring, error) {
	var row model.RecurringRow
	result := r.readQuery(ctx).
		Where("recurrings.id = ? AND recurrings.user_id = ?", id, userID).
		Scan(&row)
	if result.Error != nil {
		return nil, recurringErrors.wrap("get recurring", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, recurringErrors.wrap("get recurring", gorm.ErrRecordNotFound)
	}
	return row.ToDomain(), nil
}

func (r *recurringRepository) Create(ctx context.Context, recurring *domain.Recurring) error {
	actor := domain.ActorFrom(ctx)
	recurring.CreatedBy, recurring.UpdatedBy = actor, actor

	row := model.RecurringFromDomain(recurring)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return recurringErrors.wrap("create recurring", err)
	}

	recurring.CreatedAt = row.CreatedAt
	recurring.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *recurringRepository) Update(ctx context.Context, recurring *domain.Recurring) error {
	recurring.UpdatedBy = domain.ActorFrom(ctx)

	result := dbFrom(ctx, r.db).Model(&model.RecurringModel{}).
		Where("id = ? AND user_id = ?", recurring.ID, recurring.UserID).
		Updates(map[string]any{
			"category_id":   recurring.CategoryID,
			"wallet_id":     recurring.WalletID,
			"name":          recurring.Name,
			"amount":        recurring.Amount,
			"currency":      string(recurring.Currency),
			"frequency":     string(recurring.Frequency),
			"next_due_date": recurring.NextDueDate,
			"anchor_day":    recurring.AnchorDay,
			"is_paused":     recurring.IsPaused,
			"updated_by":    recurring.UpdatedBy,
		})
	if result.Error != nil {
		return recurringErrors.wrap("update recurring", result.Error)
	}
	if result.RowsAffected == 0 {
		return recurringErrors.wrap("update recurring", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *recurringRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result := dbFrom(ctx, r.db).
		Model(&model.RecurringModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		})
	if result.Error != nil {
		return recurringErrors.wrap("delete recurring", result.Error)
	}
	if result.RowsAffected == 0 {
		return recurringErrors.wrap("delete recurring", gorm.ErrRecordNotFound)
	}
	return nil
}
