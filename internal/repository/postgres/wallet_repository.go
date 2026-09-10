package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

// defaultWalletOrder: oldest first, id as the tie breaker.
const defaultWalletOrder = "created_at ASC, id ASC"

// currentBalance: the stated balance moved by every transaction filed since it was
// stated. Strictly after balance_updated_at, because the figure its owner gave for
// that moment already contains everything up to it; counting those rows again would
// apply them twice. Income is stored positive and expense negative, so a plain SUM
// is the whole arithmetic. A row in another currency is left out rather than added
// at a rate nothing here knows.
const currentBalance = `wallets.balance + COALESCE((
	SELECT SUM(t.amount)
	FROM transactions t
	WHERE t.wallet_id = wallets.id
	  AND t.user_id = wallets.user_id
	  AND t.currency = wallets.currency
	  AND t.deleted_at IS NULL
	  AND t.occurred_at > wallets.balance_updated_at
), 0)`

const currentBalanceColumn = currentBalance + "::bigint AS current_balance"

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Wallet, error) {
	var rows []model.WalletModel
	err := dbFrom(ctx, r.db).
		Model(&model.WalletModel{}).
		Select("wallets.*", currentBalanceColumn).
		Where("user_id = ?", userID).
		Order(defaultWalletOrder).
		Find(&rows).Error
	if err != nil {
		return nil, walletErrors.wrap("list wallets", err)
	}
	return model.WalletsToDomain(rows), nil
}

// Options: id and name only, for the wallet filter dropdown.
func (r *walletRepository) Options(ctx context.Context, userID uuid.UUID) ([]domain.WalletOption, error) {
	var rows []domain.WalletOption
	err := dbFrom(ctx, r.db).Model(&model.WalletModel{}).
		Select("id", "name").
		Where("user_id = ?", userID).
		Order(defaultWalletOrder).
		Find(&rows).Error
	if err != nil {
		return nil, walletErrors.wrap("options wallet", err)
	}
	return rows, nil
}

// GetByID always includes user_id so other users cannot reach this data.
func (r *walletRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Wallet, error) {
	var row model.WalletModel
	err := dbFrom(ctx, r.db).
		Model(&model.WalletModel{}).
		Select("wallets.*", currentBalanceColumn).
		First(&row, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		return nil, walletErrors.wrap("get wallet", err)
	}
	return row.ToDomain(), nil
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	actor := domain.ActorFrom(ctx)
	wallet.CreatedBy, wallet.UpdatedBy = actor, actor
	wallet.BalanceUpdatedBy = actor

	// The opening balance is stated now, and saying so is load-bearing: the running
	// balance counts transactions after this moment, so a zero value here would pull
	// in every row the wallet ever carried, backdated ones included.
	if wallet.BalanceUpdatedAt.IsZero() {
		wallet.BalanceUpdatedAt = time.Now()
	}

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

	balanceChanged := "balance IS DISTINCT FROM ?"

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
	Currency       string
	Held           int64
	Owed           int64
	HeldCount      int
	Overdrawn      int64
	OverdrawnCount int
	Total          int64
}

// Overview sums in SQL rather than over a fetched list.
//
// It groups over the same per-wallet expression the cards are read with, reached
// through a derived table so it is written once: a headline summing the stated
// balance while the cards below it state the current one is a panel arguing with
// itself, which is the whole reason `getWalletsScreen()` reads the two together.
func (r *walletRepository) Overview(ctx context.Context, userID uuid.UUID) (domain.WalletOverview, error) {
	counted := dbFrom(ctx, r.db).
		Model(&model.WalletModel{}).
		Select("currency", "type", currentBalanceColumn).
		Where("user_id = ? AND include_in_total", userID)

	var rows []overviewRow
	err := dbFrom(ctx, r.db).
		Table("(?) AS counted", counted).
		Select(
			"currency",
			"COALESCE(SUM(current_balance) FILTER (WHERE current_balance >= 0), 0) AS held",
			// A card in the red owes; anything else in the red is overdrawn, and the two
			// are not the same statement. Before the balance followed its transactions
			// only a card could go negative, so one sum could answer for both.
			"COALESCE(SUM(current_balance) FILTER (WHERE current_balance < 0 AND type = 'card'), 0) AS owed",
			"COUNT(*) FILTER (WHERE current_balance >= 0) AS held_count",
			"COALESCE(SUM(current_balance) FILTER (WHERE current_balance < 0 AND type <> 'card'), 0) AS overdrawn",
			"COUNT(*) FILTER (WHERE current_balance < 0 AND type <> 'card') AS overdrawn_count",
			"COALESCE(SUM(current_balance), 0) AS total",
		).
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
			overview.Overdrawn = row.Overdrawn
			overview.OverdrawnWallets = row.OverdrawnCount
			continue
		}
		overview.HeldByCurrency = append(overview.HeldByCurrency,
			domain.CurrencyAmount{Currency: currency, Amount: row.Total})
	}
	return overview, nil
}
