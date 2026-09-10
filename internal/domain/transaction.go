package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CategoryID   uuid.UUID
	WalletID     uuid.UUID
	CategoryName string
	WalletName   string
	Name         string
	Note         string
	Type         TransactionType
	Amount       int64
	Currency     Currency
	OccurredAt   time.Time
	CreatedAt    time.Time
	CreatedBy    *uuid.UUID
	UpdatedAt    time.Time
	UpdatedBy    *uuid.UUID
	DeletedBy    *uuid.UUID
}

type TransactionFilter struct {
	Search       string
	Limit        int
	Offset       int
	OrderBy      string // ORDER BY clause, may only be filled via pagination.Sortable
	CategoryID   *uuid.UUID
	WalletID     *uuid.UUID
	Type         *TransactionType
	OccurredFrom *time.Time
	OccurredTo   *time.Time
}

type TransactionType string

const (
	TransactionTypeExpense TransactionType = "expense"
	TransactionTypeIncome  TransactionType = "income"
)

func (t TransactionType) Valid() bool {
	switch t {
	case TransactionTypeExpense, TransactionTypeIncome:
		return true
	}
	return false
}

// MatchesCategoryType: the two vocabularies are the same words, and a row has to
// agree with the category it is filed under — an expense under an income category
// would count as spending on one screen and as earning on the next.
func (t TransactionType) MatchesCategoryType(categoryType string) bool {
	return string(t) == categoryType
}

// TransactionCurrencyTotal: one currency's totals, for the currencies the headline cannot state.
type TransactionCurrencyTotal struct {
	Currency Currency
	MoneyIn  int64
	MoneyOut int64
	Net      int64
	Entries  int
}

// TransactionSummary: totals over everything the filter matches, not just the page.
// The headline is base currency only; anything else stays unconverted in ByCurrency.
type TransactionSummary struct {
	BaseCurrency Currency
	MoneyIn      int64
	MoneyOut     int64
	Net          int64
	OtherEntries int
	ByCurrency   []TransactionCurrencyTotal
}

// TimeRange: half-open, [From, To). Half-open is the only shape where a month boundary
// falls in exactly one period, so two periods compared against each other cannot both
// claim the same midnight.
type TimeRange struct {
	From time.Time
	To   time.Time
}

// CategorySpend: one category's expense inside a period, stated positive even though the
// rows behind it are negative, because a spend is read as a magnitude.
type CategorySpend struct {
	CategoryID uuid.UUID
	Name       string
	Icon       string
	Color      string
	Spent      int64
}

// TransactionPeriodTotals: one period, base currency only. MoneyOut stays negative, as the
// rows are stored, which is what lets Net be a plain sum of the two.
type TransactionPeriodTotals struct {
	MoneyIn        int64
	MoneyOut       int64
	Net            int64
	IncomeEntries  int
	ExpenseEntries int
	ByCategory     []CategorySpend
}

// TransactionOverview: the month the dashboard reports, beside the month its deltas are
// measured against. Both are counted the same way, because every delta divides one by the
// other and two different definitions would make the percentage meaningless.
type TransactionOverview struct {
	BaseCurrency Currency
	PeriodStart  time.Time
	Current      TransactionPeriodTotals
	Previous     TransactionPeriodTotals
	// OtherEntries: rows outside the base currency in the reported month, counted but never
	// added in, the same way the wallets and budgets totals leave them out.
	OtherEntries int
}

// TransactionTrendBucket: how the trend chart divides its window.
type TransactionTrendBucket string

const (
	TrendBucketWeek  TransactionTrendBucket = "week"
	TrendBucketMonth TransactionTrendBucket = "month"
)

// TransactionTrendRange: the two windows the chart offers, and what each one buckets by.
type TransactionTrendRange string

const (
	TrendRangeWeekly  TransactionTrendRange = "weekly"
	TrendRangeMonthly TransactionTrendRange = "monthly"
)

func (r TransactionTrendRange) Valid() bool {
	switch r {
	case TrendRangeWeekly, TrendRangeMonthly:
		return true
	}
	return false
}

// TrendMonths: how many months the monthly range covers, the current one included.
const TrendMonths = 6

// TransactionTrendPoint: one bar. Empty periods are points too — a chart that simply omits
// a quiet week draws a shape the months did not have.
type TransactionTrendPoint struct {
	Start    time.Time
	MoneyIn  int64
	MoneyOut int64
}

type TransactionRepository interface {
	List(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]Transaction, int, error)
	Summary(ctx context.Context, userID uuid.UUID, filter TransactionFilter) (TransactionSummary, error)
	Overview(ctx context.Context, userID uuid.UUID, current, previous TimeRange) (TransactionOverview, error)
	Trend(ctx context.Context, userID uuid.UUID, window TimeRange, bucket TransactionTrendBucket) ([]TransactionTrendPoint, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Transaction, error)
	Create(ctx context.Context, transaction *Transaction) error
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type CreateTransactionInput struct {
	CategoryID uuid.UUID
	WalletID   uuid.UUID
	Name       string
	Note       string
	Type       TransactionType
	Amount     int64
	Currency   Currency
	OccurredAt time.Time
}

type UpdateTransactionInput struct {
	CategoryID *uuid.UUID
	WalletID   *uuid.UUID
	Name       *string
	Note       *string
	Type       *TransactionType
	Amount     *int64
	Currency   *Currency
	OccurredAt *time.Time
}

type TransactionService interface {
	List(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]Transaction, int, error)
	Summary(ctx context.Context, userID uuid.UUID, filter TransactionFilter) (TransactionSummary, error)
	// Overview: the given month and the one before it. month is any instant inside the month
	// wanted; the service takes the calendar from there.
	Overview(ctx context.Context, userID uuid.UUID, month time.Time) (TransactionOverview, error)
	Trend(ctx context.Context, userID uuid.UUID, trendRange TransactionTrendRange, now time.Time) ([]TransactionTrendPoint, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Transaction, error)
	Create(ctx context.Context, userID uuid.UUID, input CreateTransactionInput) (*Transaction, error)
	Update(ctx context.Context, id, userID uuid.UUID, input UpdateTransactionInput) (*Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
