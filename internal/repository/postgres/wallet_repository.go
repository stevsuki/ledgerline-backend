package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

// defaultWalletOrder: newest first, id as the tie breaker. Stable order matters:
// the UI colours the summary bar by a wallet's position in this list.
const defaultWalletOrder = "created_at ASC, id ASC"

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Wallet, error) {
	var rows []model.WalletModel
	err := dbFrom(ctx, r.db).
		Where("user_id = ?", userID).
		Order(defaultWalletOrder).
		Find(&rows).Error
	if err != nil {
		return nil, walletErrors.wrap("list wallets", err)
	}
	return model.WalletsToDomain(rows), nil
}

// GetByID always includes user_id so other users cannot reach this data.
func (r *walletRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Wallet, error) {
	var row model.WalletModel
	err := dbFrom(ctx, r.db).First(&row, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		return nil, walletErrors.wrap("get wallet", err)
	}
	return row.ToDomain(), nil
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	actor := domain.ActorFrom(ctx)
	wallet.CreatedBy, wallet.UpdatedBy = actor, actor
	// The opening balance is a figure someone typed, so it is stamped like any later edit.
	wallet.BalanceUpdatedBy = actor

	row := model.WalletFromDomain(wallet)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return walletErrors.wrap("create wallet", err)
	}

	wallet.CreatedAt = row.CreatedAt
	wallet.UpdatedAt = row.UpdatedAt
	wallet.BalanceUpdatedAt = row.BalanceUpdatedAt
	return nil
}

func (r *walletRepository) Update(ctx context.Context, wallet *domain.Wallet) error {
	wallet.UpdatedBy = domain.ActorFrom(ctx)

	// balance_updated_at only moves when the figure actually changes: the UI prints
	// its age ("Updated 8 days ago"), so an unrelated rename must not reset it.
	// Inside an UPDATE, a bare column on the right-hand side is still the old row.
	balanceChanged := "balance IS DISTINCT FROM ?"

	// user_id in the WHERE so a wallet read for one user can never be written by another.
	result := dbFrom(ctx, r.db).Model(&model.WalletModel{}).
		Where("id = ? AND user_id = ?", wallet.ID, wallet.UserID).
		Updates(map[string]any{
			"name":             wallet.Name,
			"type":             string(wallet.Type),
			"icon":             wallet.Icon,
			"currency":         string(wallet.Currency),
			"reference":        wallet.Reference,
			"balance":          wallet.Balance,
			"include_in_total": wallet.IncludeInTotal,
			"credit_limit":     wallet.CreditLimit,
			"due_day":          wallet.DueDay,
			"updated_by":       wallet.UpdatedBy,
			"balance_updated_at": gorm.Expr(
				"CASE WHEN "+balanceChanged+" THEN NOW() ELSE balance_updated_at END",
				wallet.Balance),
			"balance_updated_by": gorm.Expr(
				"CASE WHEN "+balanceChanged+" THEN ? ELSE balance_updated_by END",
				wallet.Balance, wallet.UpdatedBy),
		})
	if result.Error != nil {
		return walletErrors.wrap("update wallet", result.Error)
	}
	if result.RowsAffected == 0 {
		return walletErrors.wrap("update wallet", gorm.ErrRecordNotFound)
	}

	var updated model.WalletModel
	err := dbFrom(ctx, r.db).Select("updated_at", "balance_updated_at", "balance_updated_by").
		First(&updated, "id = ?", wallet.ID).Error
	if err == nil {
		wallet.UpdatedAt = updated.UpdatedAt
		wallet.BalanceUpdatedAt = updated.BalanceUpdatedAt
		wallet.BalanceUpdatedBy = updated.BalanceUpdatedBy
	}
	return nil
}

// Delete: soft delete stamped with who did it, in one statement.
func (r *walletRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	result := dbFrom(ctx, r.db).
		Model(&model.WalletModel{}).
		Where("id = ? AND user_id = ?", id, userID).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		})
	if result.Error != nil {
		return walletErrors.wrap("delete wallet", result.Error)
	}
	if result.RowsAffected == 0 {
		return walletErrors.wrap("delete wallet", gorm.ErrRecordNotFound)
	}
	return nil
}

// overviewRow: one currency's totals, split by sign so debt never hides inside the headline.
type overviewRow struct {
	Currency  string
	Held      int64
	Owed      int64
	HeldCount int
	Total     int64
}

// Overview sums in SQL rather than over a fetched list: the rule for what counts
// belongs here, and one grouped read cannot disagree with itself.
func (r *walletRepository) Overview(ctx context.Context, userID uuid.UUID) (domain.WalletOverview, error) {
	var rows []overviewRow
	err := dbFrom(ctx, r.db).
		Model(&model.WalletModel{}).
		Select(
			"currency",
			"COALESCE(SUM(balance) FILTER (WHERE balance >= 0), 0) AS held",
			"COALESCE(SUM(balance) FILTER (WHERE balance < 0), 0) AS owed",
			"COUNT(*) FILTER (WHERE balance >= 0) AS held_count",
			"COALESCE(SUM(balance), 0) AS total",
		).
		Where("user_id = ? AND include_in_total", userID).
		Group("currency").
		Order("currency ASC").
		Scan(&rows).Error
	if err != nil {
		return domain.WalletOverview{}, walletErrors.wrap("wallet overview", err)
	}

	overview := domain.WalletOverview{
		BaseCurrency:   domain.BaseCurrency,
		HeldByCurrency: []domain.CurrencyAmount{},
	}
	for _, row := range rows {
		currency := domain.Currency(row.Currency)
		if currency == domain.BaseCurrency {
			overview.TotalHeld = row.Held
			overview.CountedWallets = row.HeldCount
			overview.OwedOnCards = row.Owed
			continue
		}
		overview.HeldByCurrency = append(overview.HeldByCurrency,
			domain.CurrencyAmount{Currency: currency, Amount: row.Total})
	}
	return overview, nil
}
