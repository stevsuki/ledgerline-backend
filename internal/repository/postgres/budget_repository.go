package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

const defaultBudgetOrder = "budgets.created_at DESC, budgets.id DESC"

type budgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) domain.BudgetRepository {
	return &budgetRepository{db: db}
}

// readQuery: every read joins the category a budget limits.
func (r *budgetRepository) readQuery(ctx context.Context) *gorm.DB {
	return dbFrom(ctx, r.db).
		Model(&model.BudgetModel{}).
		Select(
			"budgets.id",
			"budgets.user_id",
			"budgets.category_id",
			"categories.name AS category_name",
			"categories.icon",
			"categories.color",
			"budgets.currency",
			"budgets.monthly_limit",
			spentThisCycle,
			carriedOverFromLastCycle,
			"budgets.alert_threshold_percent",
			"budgets.is_fixed",
			"budgets.rollover",
			"budgets.created_at",
			"budgets.created_by",
			"budgets.updated_at",
			"budgets.updated_by",
			"budgets.deleted_by",
		).
		Joins("JOIN categories ON categories.id = budgets.category_id AND categories.deleted_at IS NULL")
}

func (r *budgetRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	var rows []model.BudgetRow
	err := r.readQuery(ctx).
		Where("budgets.user_id = ?", userID).
		Order(defaultBudgetOrder).
		Scan(&rows).Error
	if err != nil {
		return nil, budgetErrors.wrap("list budgets", err)
	}
	return model.BudgetRowsToDomain(rows), nil
}

// GetByID: Scan never raises not-found, so the row count does.
func (r *budgetRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Budget, error) {
	var row model.BudgetRow
	result := r.readQuery(ctx).
		Where("budgets.id = ? AND budgets.user_id = ?", id, userID).
		Scan(&row)
	if result.Error != nil {
		return nil, budgetErrors.wrap("get budget", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, budgetErrors.wrap("get budget", gorm.ErrRecordNotFound)
	}
	return row.ToDomain(), nil
}

func (r *budgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	actor := domain.ActorFrom(ctx)
	budget.CreatedBy, budget.UpdatedBy = actor, actor

	row := model.BudgetFromDomain(budget)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return budgetErrors.wrap("create budget", err)
	}

	budget.CreatedAt = row.CreatedAt
	budget.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *budgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	budget.UpdatedBy = domain.ActorFrom(ctx)

	result := dbFrom(ctx, r.db).Model(&model.BudgetModel{}).
		Where("id = ? AND user_id = ?", budget.ID, budget.UserID).
		Updates(map[string]any{
			"category_id":             budget.CategoryID,
			"currency":                budget.Currency,
			"monthly_limit":           budget.MonthlyLimit,
			"alert_threshold_percent": budget.AlertThresholdPercent,
			"is_fixed":                budget.IsFixed,
			"rollover":                budget.Rollover,
			"updated_by":              budget.UpdatedBy,
		})
	if result.Error != nil {
		return budgetErrors.wrap("update budget", result.Error)
	}
	if result.RowsAffected == 0 {
		return budgetErrors.wrap("update budget", gorm.ErrRecordNotFound)
	}

	var updated model.BudgetModel
	err := dbFrom(ctx, r.db).Select("updated_at").First(&updated, "id = ?", budget.ID).Error
	if err == nil {
		budget.UpdatedAt = updated.UpdatedAt
	}
	return nil
}

func (r *budgetRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result := dbFrom(ctx, r.db).
		Model(&model.BudgetModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		})
	if result.Error != nil {
		return budgetErrors.wrap("delete budget", result.Error)
	}
	if result.RowsAffected == 0 {
		return budgetErrors.wrap("delete budget", gorm.ErrRecordNotFound)
	}
	return nil
}

type budgetUsageRow struct {
	BudgetID              uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	Currency              string
	MonthlyLimit          int64
	Spent                 int64
	CarriedOver           int64
	AlertThresholdPercent int
	IsFixed               bool
	Rollover              bool
}

// The cycle a budget is measured over: the calendar month, the same one the overview counts days of.
const (
	cycleStart     = "date_trunc('month', NOW())"
	nextCycleStart = cycleStart + " + INTERVAL '1 month'"
	lastCycleStart = cycleStart + " - INTERVAL '1 month'"
)

// expenseSum: what the budget's own category and currency spent inside a window, as a positive
// number because the table stores expense amounts negative.
func expenseSum(from, to string) string {
	return `COALESCE((
		SELECT -SUM(t.amount)
		FROM transactions t
		WHERE t.user_id = budgets.user_id
		  AND t.category_id = budgets.category_id
		  AND t.currency = budgets.currency
		  AND t.type = 'expense'
		  AND t.deleted_at IS NULL
		  AND t.occurred_at >= ` + from + `
		  AND t.occurred_at < ` + to + `
	), 0)`
}

var (
	spentThisCycle = expenseSum(cycleStart, nextCycleStart) + "::bigint AS spent"

	// carriedOverFromLastCycle: a rollover budget keeps what it did not spend last cycle, and only if
	// it already existed then. Last cycle's limit is read as the current one, the table keeps no history.
	carriedOverFromLastCycle = `CASE
		WHEN budgets.rollover AND budgets.created_at < ` + cycleStart + `
		THEN GREATEST(budgets.monthly_limit - ` + expenseSum(lastCycleStart, cycleStart) + `, 0)
		ELSE 0
	END::bigint AS carried_over`
)

func (r *budgetRepository) Usage(ctx context.Context, userID uuid.UUID) ([]domain.BudgetUsage, error) {
	var rows []budgetUsageRow
	err := dbFrom(ctx, r.db).
		Model(&model.BudgetModel{}).
		Select(
			"budgets.id AS budget_id",
			"budgets.category_id",
			"categories.name AS category_name",
			"categories.icon",
			"categories.color",
			"budgets.currency",
			"budgets.monthly_limit",
			"budgets.alert_threshold_percent",
			"budgets.is_fixed",
			"budgets.rollover",
			spentThisCycle,
			carriedOverFromLastCycle,
		).
		Joins("JOIN categories ON categories.id = budgets.category_id AND categories.deleted_at IS NULL").
		Where("budgets.user_id = ?", userID).
		Order(defaultBudgetOrder).
		Scan(&rows).Error
	if err != nil {
		return nil, budgetErrors.wrap("budget usage", err)
	}

	usages := make([]domain.BudgetUsage, 0, len(rows))
	for _, row := range rows {
		usages = append(usages, domain.BudgetUsage{
			BudgetID:              row.BudgetID,
			CategoryID:            row.CategoryID,
			CategoryName:          row.CategoryName,
			Icon:                  row.Icon,
			Color:                 row.Color,
			Currency:              domain.Currency(row.Currency),
			MonthlyLimit:          row.MonthlyLimit,
			Spent:                 row.Spent,
			CarriedOver:           row.CarriedOver,
			AlertThresholdPercent: row.AlertThresholdPercent,
			IsFixed:               row.IsFixed,
			Rollover:              row.Rollover,
		})
	}
	return usages, nil
}

func (r *budgetRepository) ExistsByCategory(ctx context.Context, categoryID, userID uuid.UUID) (bool, error) {
	var count int64
	err := dbFrom(ctx, r.db).
		Model(&model.BudgetModel{}).
		Where("category_id = ? AND user_id = ?", categoryID, userID).
		Count(&count).Error
	if err != nil {
		return false, budgetErrors.wrap("budget exists by category", err)
	}
	return count > 0, nil
}
