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

type TransactionRepository interface {
	List(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]Transaction, int, error)
	Summary(ctx context.Context, userID uuid.UUID, filter TransactionFilter) (TransactionSummary, error)
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
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Transaction, error)
	Create(ctx context.Context, userID uuid.UUID, input CreateTransactionInput) (*Transaction, error)
	Update(ctx context.Context, id, userID uuid.UUID, input UpdateTransactionInput) (*Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
