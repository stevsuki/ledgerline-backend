package postgres

import (
	"context"
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
