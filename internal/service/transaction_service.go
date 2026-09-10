package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type TransactionService struct {
	transactionRepo domain.TransactionRepository
	categoryRepo    domain.CategoryRepository
	walletRepo      domain.WalletRepository
}

func NewTransactionService(transactionRepo domain.TransactionRepository, categoryRepo domain.CategoryRepository, walletRepo domain.WalletRepository) domain.TransactionService {
	return &TransactionService{transactionRepo: transactionRepo, categoryRepo: categoryRepo, walletRepo: walletRepo}
}

func (t *TransactionService) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return t.transactionRepo.List(ctx, userID, filter)
}

func (t *TransactionService) Summary(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) (domain.TransactionSummary, error) {
	return t.transactionRepo.Summary(ctx, userID, filter)
}

func (t *TransactionService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	return t.transactionRepo.GetByID(ctx, id, userID)
}

func (t *TransactionService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction name is required").WithField("name")
	}

	note := strings.TrimSpace(input.Note)
	if note == "" {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction note is required").WithField("note")
	}

	if !input.Type.Valid() {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidType, "transaction type must be debit or credit").WithField("type")
	}

	if !input.Currency.Valid() {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalidCurrency,
			"transaction currency must be IDR, USD, or SGD").WithField("currency")
	}

	if err := checkAmountSign(input.Amount, input.Type); err != nil {
		return nil, err
	}

	wallet, err := t.walletRepo.GetByID(ctx, input.WalletID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidWallet,
				"wallet_id does not refer to one of your wallets").WithField("wallet_id")
		}
		return nil, err
	}
	if err := checkWalletCurrency(input.Currency, wallet); err != nil {
		return nil, err
	}

	category, err := t.categoryRepo.GetByID(ctx, input.CategoryID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidCategory,
				"category_id does not refer to one of your categories").WithField("category_id")
		}
		return nil, err
	}
	if err := checkCategoryType(input.Type, category); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	transaction := &domain.Transaction{
		ID:         id,
		UserID:     userID,
		WalletID:   input.WalletID,
		CategoryID: input.CategoryID,
		Name:       name,
		Note:       note,
		Type:       input.Type,
		Amount:     input.Amount,
		Currency:   input.Currency,
		OccurredAt: input.OccurredAt,
	}

	if err := t.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (t *TransactionService) Update(ctx context.Context, id, userID uuid.UUID, input domain.UpdateTransactionInput) (*domain.Transaction, error) {
	transaction, err := t.transactionRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Kept for the pairing checks at the end, which only read once every field the
	// patch touches has been applied.
	var wallet *domain.Wallet
	var category *domain.Category

	if input.WalletID != nil {
		wallet, err = t.walletRepo.GetByID(ctx, *input.WalletID, userID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, domain.InvalidInput(domain.CodeTransactionInvalidWallet,
					"wallet_id does not refer to one of your wallets").WithField("wallet_id")
			}
			return nil, err
		}
		transaction.WalletID = *input.WalletID
	}

	if input.CategoryID != nil {
		category, err = t.categoryRepo.GetByID(ctx, *input.CategoryID, userID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, domain.InvalidInput(domain.CodeTransactionInvalidCategory,
					"category_id does not refer to one of your categories").WithField("category_id")
			}
			return nil, err
		}
		transaction.CategoryID = *input.CategoryID
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction name must not be empty").WithField("name")
		}
		transaction.Name = name
	}

	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		if note == "" {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidData, "transaction note must not be empty").WithField("note")
		}
		transaction.Note = note
	}

	if input.Type != nil {
		if !input.Type.Valid() {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidType, "transaction type must be income or expense").WithField("type")
		}
		transaction.Type = *input.Type
	}

	if input.Currency != nil {
		if !input.Currency.Valid() {
			return nil, domain.InvalidInput(domain.CodeTransactionInvalidCurrency,
				"transaction currency must be IDR, USD, or SGD").WithField("currency")
		}
		transaction.Currency = *input.Currency
	}

	if input.Amount != nil {
		transaction.Amount = *input.Amount
	}

	if input.OccurredAt != nil {
		transaction.OccurredAt = *input.OccurredAt
	}

	if err := checkAmountSign(transaction.Amount, transaction.Type); err != nil {
		return nil, err
	}

	// A currency change has to answer to the wallet it sits in even when the wallet
	// itself was not touched, and the same the other way around — so whichever side
	// the patch left unread is read now.
	if input.Currency != nil && wallet == nil {
		if wallet, err = t.walletRepo.GetByID(ctx, transaction.WalletID, userID); err != nil {
			return nil, err
		}
	}
	if wallet != nil {
		if err := checkWalletCurrency(transaction.Currency, wallet); err != nil {
			return nil, err
		}
	}

	if input.Type != nil && category == nil {
		if category, err = t.categoryRepo.GetByID(ctx, transaction.CategoryID, userID); err != nil {
			return nil, err
		}
	}
	if category != nil {
		if err := checkCategoryType(transaction.Type, category); err != nil {
			return nil, err
		}
	}

	if err := t.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, err
	}
	return transaction, nil
}

func (t *TransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return t.transactionRepo.Delete(ctx, id, userID)
}

// checkWalletCurrency: a wallet holds one currency, and its balance sums only the rows
// kept in that currency — a row in another one would be money the wallet can never
// account for, so it is refused instead of filed where nothing will count it.
func checkWalletCurrency(currency domain.Currency, wallet *domain.Wallet) error {
	if currency == wallet.Currency {
		return nil
	}
	return domain.InvalidInput(domain.CodeTransactionCurrencyMismatch,
		fmt.Sprintf("%s holds %s, so this transaction cannot be kept in %s",
			wallet.Name, wallet.Currency, currency)).WithField("currency")
}

// checkCategoryType: income belongs under an income category and expense under an
// expense one, or the same row would read as earning on one screen and spending on
// the next — and a budget would measure against a category that cannot be spent on.
func checkCategoryType(transactionType domain.TransactionType, category *domain.Category) error {
	if transactionType.MatchesCategoryType(category.Type) {
		return nil
	}
	return domain.InvalidInput(domain.CodeTransactionCategoryMismatch,
		fmt.Sprintf("%s is a %s category, so it cannot carry a %s transaction",
			category.Name, category.Type, transactionType)).WithField("category_id")
}

func checkAmountSign(amount int64, transactionType domain.TransactionType) error {
	switch {
	case amount == 0:
		return domain.InvalidInput(domain.CodeTransactionInvalidAmount,
			"amount must not be zero").WithField("amount")
	case transactionType == domain.TransactionTypeIncome && amount < 0:
		return domain.InvalidInput(domain.CodeTransactionInvalidAmount,
			"income transaction must have positive amount").WithField("amount")
	case transactionType == domain.TransactionTypeExpense && amount > 0:
		return domain.InvalidInput(domain.CodeTransactionInvalidAmount,
			"expense transaction must have negative amount").WithField("amount")
	}
	return nil
}

/* ── the dashboard ─────────────────────────────────────────────────────── */

// monthStart: the first instant of the month an instant falls in, in its own zone.
func monthStart(at time.Time) time.Time {
	return time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, at.Location())
}

// Overview: the month asked for and the month before it, each as a half-open range so the
// midnight between them belongs to exactly one of the two.
func (t *TransactionService) Overview(
	ctx context.Context, userID uuid.UUID, month time.Time,
) (domain.TransactionOverview, error) {
	start := monthStart(month)
	current := domain.TimeRange{From: start, To: start.AddDate(0, 1, 0)}
	previous := domain.TimeRange{From: start.AddDate(0, -1, 0), To: start}

	return t.transactionRepo.Overview(ctx, userID, current, previous)
}

// Trend: the window the range names, with a point for every period in it.
func (t *TransactionService) Trend(
	ctx context.Context, userID uuid.UUID, trendRange domain.TransactionTrendRange, now time.Time,
) ([]domain.TransactionTrendPoint, error) {
	if !trendRange.Valid() {
		return nil, domain.InvalidInput(domain.CodeTransactionInvalid,
			"trend range must be weekly or monthly").WithField("range")
	}

	buckets, bucket := trendBuckets(trendRange, now)
	if len(buckets) == 0 {
		return []domain.TransactionTrendPoint{}, nil
	}

	window := domain.TimeRange{
		From: buckets[0],
		To:   bucketEnd(buckets[len(buckets)-1], bucket),
	}

	points, err := t.transactionRepo.Trend(ctx, userID, window, bucket)
	if err != nil {
		return nil, err
	}
	return fillTrend(buckets, bucket, points), nil
}

// weekStart: the Monday of the week an instant falls in, matching what Postgres
// `date_trunc('week', …)` answers, so a row cannot land outside the bucket it was counted in.
func weekStart(at time.Time) time.Time {
	daysSinceMonday := (int(at.Weekday()) + 6) % 7
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	return day.AddDate(0, 0, -daysSinceMonday)
}

func bucketEnd(start time.Time, bucket domain.TransactionTrendBucket) time.Time {
	if bucket == domain.TrendBucketWeek {
		return start.AddDate(0, 0, 7)
	}
	return start.AddDate(0, 1, 0)
}

// trendBuckets: every period the chart is meant to draw, earliest first.
//
// Weekly is the weeks of the month being reported, because that is the range the label
// states; monthly is the last TrendMonths months, this one included. An empty period is
// still a bar — a chart that omits a quiet week draws a shape the month did not have.
func trendBuckets(
	trendRange domain.TransactionTrendRange, now time.Time,
) ([]time.Time, domain.TransactionTrendBucket) {
	start := monthStart(now)

	if trendRange == domain.TrendRangeWeekly {
		nextMonth := start.AddDate(0, 1, 0)
		weeks := []time.Time{}
		for week := weekStart(start); week.Before(nextMonth); week = week.AddDate(0, 0, 7) {
			weeks = append(weeks, week)
		}
		return weeks, domain.TrendBucketWeek
	}

	months := make([]time.Time, 0, domain.TrendMonths)
	for i := domain.TrendMonths - 1; i >= 0; i-- {
		months = append(months, start.AddDate(0, -i, 0))
	}
	return months, domain.TrendBucketMonth
}

// fillTrend: the database's rows laid onto the periods that were asked for.
//
// A row is matched by the period it falls inside rather than by an equal timestamp: the
// bucket start comes back from Postgres in the session's zone, and comparing instants for
// equality across zones is how a bar silently loses its figures.
func fillTrend(
	buckets []time.Time, bucket domain.TransactionTrendBucket, points []domain.TransactionTrendPoint,
) []domain.TransactionTrendPoint {
	filled := make([]domain.TransactionTrendPoint, 0, len(buckets))

	for _, start := range buckets {
		end := bucketEnd(start, bucket)
		point := domain.TransactionTrendPoint{Start: start}

		for _, row := range points {
			if !row.Start.Before(start) && row.Start.Before(end) {
				point.MoneyIn += row.MoneyIn
				point.MoneyOut += row.MoneyOut
			}
		}
		filled = append(filled, point)
	}
	return filled
}
