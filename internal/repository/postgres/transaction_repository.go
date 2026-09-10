package postgres

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
	"gorm.io/gorm"
)

const defaultTransactionOrder = "users.occurred_at DESC, users.id ASC"

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

// filtered: the predicates List and Summary must agree on, so both describe the same set.
func (r *transactionRepository) filtered(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) *gorm.DB {
	query := dbFrom(ctx, r.db).Model(&model.TransactionModel{}).
		Where("transactions.user_id = ?", userID)

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(transactions.name) LIKE ?", keyword)
	}

	if filter.CategoryID != nil {
		query = query.Where("transactions.category_id = ?", filter.CategoryID)
	}

	if filter.WalletID != nil {
		query = query.Where("transactions.wallet_id = ?", filter.WalletID)
	}

	if filter.Type != nil {
		query = query.Where("transactions.type = ?", string(*filter.Type))
	}

	if filter.OccurredFrom != nil {
		query = query.Where("transactions.occurred_at >= ?", filter.OccurredFrom)
	}

	if filter.OccurredTo != nil {
		query = query.Where("transactions.occurred_at <= ?", filter.OccurredTo)
	}

	return query
}

// transactionSummaryRow: one currency's totals, split by type so income and expense stay separate.
type transactionSummaryRow struct {
	Currency string
	MoneyIn  int64
	MoneyOut int64
	Net      int64
	Entries  int
}

// Summary sums in SQL over the whole filtered set, ignoring limit and offset.
func (r *transactionRepository) Summary(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) (domain.TransactionSummary, error) {
	var rows []transactionSummaryRow
	err := r.filtered(ctx, userID, filter).
		Select(
			"transactions.currency",
			"COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'income'), 0) AS money_in",
			"COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'expense'), 0) AS money_out",
			"COALESCE(SUM(transactions.amount), 0) AS net",
			"COUNT(*) AS entries",
		).
		Group("transactions.currency").
		Order("transactions.currency ASC").
		Scan(&rows).Error
	if err != nil {
		return domain.TransactionSummary{}, transactionErrors.wrap("summarize transactions", err)
	}

	summary := domain.TransactionSummary{
		BaseCurrency: domain.BaseCurrency,
		ByCurrency:   []domain.TransactionCurrencyTotal{},
	}
	for _, row := range rows {
		currency := domain.Currency(row.Currency)
		if currency == domain.BaseCurrency {
			summary.MoneyIn = row.MoneyIn
			summary.MoneyOut = row.MoneyOut
			summary.Net = row.Net
			continue
		}

		summary.OtherEntries += row.Entries
		summary.ByCurrency = append(summary.ByCurrency, domain.TransactionCurrencyTotal{
			Currency: currency,
			MoneyIn:  row.MoneyIn,
			MoneyOut: row.MoneyOut,
			Net:      row.Net,
			Entries:  row.Entries,
		})
	}
	return summary, nil
}

func (r *transactionRepository) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int, error) {
	query := r.filtered(ctx, userID, filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, transactionErrors.wrap("count transactions", err)
	}
	if total == 0 {
		return []domain.Transaction{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultTransactionOrder
	}

	var rows []model.TransactionModel
	err := query.
		Select("transactions.*, wallets.name AS wallet_name, categories.name AS category_name").
		Joins("LEFT JOIN wallets ON wallets.id = transactions.wallet_id").
		Joins("LEFT JOIN categories ON categories.id = transactions.category_id").
		Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&rows).Error
	if err != nil {
		return nil, 0, transactionErrors.wrap("list transactions", err)
	}
	return model.TransactionsToDomain(rows), int(total), nil
}

// GetByID always includes user_id so other users cannot reach this data.
func (r *transactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	var row model.TransactionModel
	err := dbFrom(ctx, r.db).Model(&model.TransactionModel{}).
		Select("transactions.*, wallets.name AS wallet_name, categories.name AS category_name").
		Joins("LEFT JOIN wallets ON wallets.id = transactions.wallet_id").
		Joins("LEFT JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.id = ? AND transactions.user_id = ?", id, userID).
		First(&row).Error
	if err != nil {
		return nil, transactionErrors.wrap("get transaction", err)
	}

	return row.ToDomain(), nil
}

func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	actor := domain.ActorFrom(ctx)
	transaction.CreatedBy, transaction.UpdatedBy = actor, actor

	row := model.TransactionFromDomain(transaction)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return transactionErrors.wrap("create transaction", err)
	}

	transaction.CreatedAt = row.CreatedAt
	transaction.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	transaction.UpdatedBy = domain.ActorFrom(ctx)

	result := dbFrom(ctx, r.db).Model(&model.TransactionModel{}).
		Where("id = ? AND user_id = ?", transaction.ID, transaction.UserID).
		Updates(map[string]any{
			"category_id": transaction.CategoryID,
			"wallet_id":   transaction.WalletID,
			"name":        transaction.Name,
			"note":        transaction.Note,
			"type":        string(transaction.Type),
			"amount":      transaction.Amount,
			"currency":    string(transaction.Currency),
			"occurred_at": transaction.OccurredAt,
			"updated_by":  transaction.UpdatedBy,
		})
	if result.Error != nil {
		return transactionErrors.wrap("update transaction", result.Error)
	}
	if result.RowsAffected == 0 {
		return transactionErrors.wrap("update transaction", gorm.ErrRecordNotFound)
	}

	var updated model.TransactionModel
	err := dbFrom(ctx, r.db).Select("updated_at").First(&updated, "id = ?", transaction.ID).Error
	if err == nil {
		transaction.UpdatedAt = updated.UpdatedAt
	}
	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result := dbFrom(ctx, r.db).
		Model(&model.TransactionModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		})
	if result.Error != nil {
		return transactionErrors.wrap("delete transaction", result.Error)
	}
	if result.RowsAffected == 0 {
		return transactionErrors.wrap("delete transaction", gorm.ErrRecordNotFound)
	}
	return nil
}

/* ── the dashboard's two reads ─────────────────────────────────────────── */

// overviewTotalsRow: one currency with both periods beside each other, so the base currency
// and the ones the totals cannot absorb come back from the same grouped read.
type overviewTotalsRow struct {
	Currency               string
	MoneyIn                int64
	MoneyOut               int64
	IncomeEntries          int
	ExpenseEntries         int
	Entries                int
	PreviousMoneyIn        int64
	PreviousMoneyOut       int64
	PreviousIncomeEntries  int
	PreviousExpenseEntries int
}

// overviewCategoryRow: one category, both periods. The donut needs this month; the "driven
// by" note needs the difference, which is a comparison rather than a total.
type overviewCategoryRow struct {
	CategoryID    uuid.UUID
	Name          string
	Icon          string
	Color         string
	Spent         int64
	PreviousSpent int64
}

// The windows are passed in rather than derived here, so every figure the dashboard prints
// is measured against one set of month boundaries instead of each query keeping its own.
const overviewTotalsSelect = `transactions.currency,
	COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'income'  AND transactions.occurred_at >= ? AND transactions.occurred_at < ?), 0) AS money_in,
	COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'expense' AND transactions.occurred_at >= ? AND transactions.occurred_at < ?), 0) AS money_out,
	COUNT(*) FILTER (WHERE transactions.type = 'income'  AND transactions.occurred_at >= ? AND transactions.occurred_at < ?) AS income_entries,
	COUNT(*) FILTER (WHERE transactions.type = 'expense' AND transactions.occurred_at >= ? AND transactions.occurred_at < ?) AS expense_entries,
	COUNT(*) FILTER (WHERE transactions.occurred_at >= ? AND transactions.occurred_at < ?) AS entries,
	COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'income'  AND transactions.occurred_at >= ? AND transactions.occurred_at < ?), 0) AS previous_money_in,
	COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'expense' AND transactions.occurred_at >= ? AND transactions.occurred_at < ?), 0) AS previous_money_out,
	COUNT(*) FILTER (WHERE transactions.type = 'income'  AND transactions.occurred_at >= ? AND transactions.occurred_at < ?) AS previous_income_entries,
	COUNT(*) FILTER (WHERE transactions.type = 'expense' AND transactions.occurred_at >= ? AND transactions.occurred_at < ?) AS previous_expense_entries`

const overviewCategorySelect = `transactions.category_id,
	categories.name,
	categories.icon,
	categories.color,
	COALESCE(-SUM(transactions.amount) FILTER (WHERE transactions.occurred_at >= ? AND transactions.occurred_at < ?), 0)::bigint AS spent,
	COALESCE(-SUM(transactions.amount) FILTER (WHERE transactions.occurred_at >= ? AND transactions.occurred_at < ?), 0)::bigint AS previous_spent`

func (r *transactionRepository) Overview(
	ctx context.Context, userID uuid.UUID, current, previous domain.TimeRange,
) (domain.TransactionOverview, error) {
	overview := domain.TransactionOverview{
		BaseCurrency: domain.BaseCurrency,
		PeriodStart:  current.From,
	}

	var totals []overviewTotalsRow
	err := dbFrom(ctx, r.db).
		Model(&model.TransactionModel{}).
		Select(overviewTotalsSelect,
			current.From, current.To, current.From, current.To,
			current.From, current.To, current.From, current.To, current.From, current.To,
			previous.From, previous.To, previous.From, previous.To,
			previous.From, previous.To, previous.From, previous.To).
		Where("transactions.user_id = ? AND transactions.occurred_at >= ? AND transactions.occurred_at < ?",
			userID, previous.From, current.To).
		Group("transactions.currency").
		Scan(&totals).Error
	if err != nil {
		return domain.TransactionOverview{}, transactionErrors.wrap("overview transactions", err)
	}

	for _, row := range totals {
		if domain.Currency(row.Currency) != domain.BaseCurrency {
			overview.OtherEntries += row.Entries
			continue
		}
		overview.Current = domain.TransactionPeriodTotals{
			MoneyIn:        row.MoneyIn,
			MoneyOut:       row.MoneyOut,
			Net:            row.MoneyIn + row.MoneyOut,
			IncomeEntries:  row.IncomeEntries,
			ExpenseEntries: row.ExpenseEntries,
		}
		overview.Previous = domain.TransactionPeriodTotals{
			MoneyIn:        row.PreviousMoneyIn,
			MoneyOut:       row.PreviousMoneyOut,
			Net:            row.PreviousMoneyIn + row.PreviousMoneyOut,
			IncomeEntries:  row.PreviousIncomeEntries,
			ExpenseEntries: row.PreviousExpenseEntries,
		}
	}

	var categories []overviewCategoryRow
	err = dbFrom(ctx, r.db).
		Model(&model.TransactionModel{}).
		Select(overviewCategorySelect, current.From, current.To, previous.From, previous.To).
		Joins("JOIN categories ON categories.id = transactions.category_id AND categories.deleted_at IS NULL").
		Where("transactions.user_id = ? AND transactions.type = 'expense' AND transactions.currency = ?"+
			" AND transactions.occurred_at >= ? AND transactions.occurred_at < ?",
			userID, domain.BaseCurrency, previous.From, current.To).
		Group("transactions.category_id, categories.name, categories.icon, categories.color").
		Order("spent DESC, categories.name ASC").
		Scan(&categories).Error
	if err != nil {
		return domain.TransactionOverview{}, transactionErrors.wrap("overview categories", err)
	}

	overview.Current.ByCategory = categorySpend(categories,
		func(row overviewCategoryRow) int64 { return row.Spent })
	overview.Previous.ByCategory = categorySpend(categories,
		func(row overviewCategoryRow) int64 { return row.PreviousSpent })
	return overview, nil
}

// categorySpend: one period's rows, spenders only. A category that spent nothing is not a
// slice of the month, and a zero slice would still take a legend entry and a colour.
func categorySpend(rows []overviewCategoryRow, spentOf func(overviewCategoryRow) int64) []domain.CategorySpend {
	spend := make([]domain.CategorySpend, 0, len(rows))
	for _, row := range rows {
		spent := spentOf(row)
		if spent <= 0 {
			continue
		}
		spend = append(spend, domain.CategorySpend{
			CategoryID: row.CategoryID,
			Name:       row.Name,
			Icon:       row.Icon,
			Color:      row.Color,
			Spent:      spent,
		})
	}
	sort.SliceStable(spend, func(i, j int) bool { return spend[i].Spent > spend[j].Spent })
	return spend
}

type trendRow struct {
	BucketStart time.Time
	MoneyIn     int64
	MoneyOut    int64
}

// Trend: one row per period that has rows in it. Filling the gaps is the service's job,
// because only it knows which periods the window was supposed to cover.
func (r *transactionRepository) Trend(
	ctx context.Context, userID uuid.UUID, window domain.TimeRange, bucket domain.TransactionTrendBucket,
) ([]domain.TransactionTrendPoint, error) {
	var rows []trendRow
	err := dbFrom(ctx, r.db).
		Model(&model.TransactionModel{}).
		Select("date_trunc(?, transactions.occurred_at) AS bucket_start,"+
			" COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'income'), 0) AS money_in,"+
			" COALESCE(SUM(transactions.amount) FILTER (WHERE transactions.type = 'expense'), 0) AS money_out",
			string(bucket)).
		Where("transactions.user_id = ? AND transactions.currency = ?"+
			" AND transactions.occurred_at >= ? AND transactions.occurred_at < ?",
			userID, domain.BaseCurrency, window.From, window.To).
		Group("bucket_start").
		Order("bucket_start ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, transactionErrors.wrap("trend transactions", err)
	}

	points := make([]domain.TransactionTrendPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, domain.TransactionTrendPoint{
			Start:    row.BucketStart,
			MoneyIn:  row.MoneyIn,
			MoneyOut: row.MoneyOut,
		})
	}
	return points, nil
}
