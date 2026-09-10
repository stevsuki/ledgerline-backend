package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/pkg/pagination"
)

var transactionSort = pagination.Sortable{
	Allowed: pagination.Whitelist{
		"name":        "transactions.name",
		"amount":      "transactions.amount",
		"note":        "transactions.note",
		"type":        "transactions.type",
		"category":    "transactions.category",
		"wallet":      "transactions.wallet",
		"occurred_at": "transactions.occurred_at",
		"created_at":  "transactions.created_at",
		"updated_at":  "transactions.updated_at",
	},
	Default:    "-occurred_at",
	TieBreaker: "transactions.id",
}

type ListTransactionsQuery struct {
	Search       string     `form:"search" example:"kopi tuku"`
	Sort         string     `form:"sort" binding:"omitempty,max=100" example:"-created_at,full_name"`
	Page         int        `form:"page" binding:"omitempty,min=1" example:"1"`
	PerPage      int        `form:"per_page" binding:"omitempty,min=1,max=100" example:"10"`
	CategoryID   *uuid.UUID `form:"category,parser=encoding.TextUnmarshaler" binding:"omitempty"`
	WalletID     *uuid.UUID `form:"wallet,parser=encoding.TextUnmarshaler" binding:"omitempty"`
	Type         string     `form:"type" binding:"omitempty,oneof=income expense" example:"expense"`
	OccurredFrom *time.Time `form:"occurred_from" binding:"omitempty" example:"2026-09-01T00:00:00Z"`
	OccurredTo   *time.Time `form:"occurred_to" binding:"omitempty" example:"2026-09-30T23:59:59Z"`
}

// OrderBy: raw sort param -> whitelisted SQL clause.
func (q ListTransactionsQuery) OrderBy() (string, error) { return transactionSort.OrderBy(q.Sort) }

type CreateTransactionRequestDTO struct {
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	WalletID   uuid.UUID `json:"wallet_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	Amount     int64     `json:"amount" binding:"required"`
	Note       string    `json:"note"`
	Type       string    `json:"type" binding:"required"`
	Currency   string    `json:"currency" binding:"required"`
	OccurredAt time.Time `json:"occurred_at" binding:"required"`
}

func (r *CreateTransactionRequestDTO) ToInput() domain.CreateTransactionInput {
	return domain.CreateTransactionInput{
		CategoryID: r.CategoryID,
		WalletID:   r.WalletID,
		Name:       r.Name,
		Amount:     r.Amount,
		Note:       r.Note,
		Type:       domain.TransactionType(r.Type),
		Currency:   domain.Currency(r.Currency),
		OccurredAt: r.OccurredAt,
	}
}

type UpdateTransactionRequestDTO struct {
	CategoryID *uuid.UUID `json:"category_id" binding:"omitempty"`
	WalletID   *uuid.UUID `json:"wallet_id" binding:"omitempty"`
	Name       *string    `json:"name" binding:"omitempty"`
	Amount     *int64     `json:"amount" binding:"omitempty"`
	Note       *string    `json:"note"`
	Type       *string    `json:"type" binding:"omitempty"`
	Currency   *string    `json:"currency" binding:"omitempty"`
	OccurredAt *time.Time `json:"occurred_at" binding:"omitempty"`
}

func (r *UpdateTransactionRequestDTO) ToInput() domain.UpdateTransactionInput {
	input := domain.UpdateTransactionInput{
		CategoryID: r.CategoryID,
		WalletID:   r.WalletID,
		Name:       r.Name,
		Amount:     r.Amount,
		Note:       r.Note,
		OccurredAt: r.OccurredAt,
	}
	if r.Type != nil {
		r := domain.TransactionType(*r.Type)
		input.Type = &r
	}
	if r.Currency != nil {
		c := domain.Currency(*r.Currency)
		input.Currency = &c
	}
	return input
}

type TransactionResponseDTO struct {
	ID           uuid.UUID `json:"id"`
	CategoryID   uuid.UUID `json:"category_id"`
	WalletID     uuid.UUID `json:"wallet_id"`
	CategoryName string    `json:"category_name"`
	WalletName   string    `json:"wallet_name"`
	Name         string    `json:"name"`
	Note         string    `json:"note"`
	Type         string    `json:"type"`
	Amount       int64     `json:"amount"`
	Currency     string    `json:"currency"`
	OccurredAt   time.Time `json:"occurred_at" example:"2026-01-02T15:04:05Z"`
}

func NewTransactionResponseDTO(t *domain.Transaction) TransactionResponseDTO {
	return TransactionResponseDTO{
		ID:           t.ID,
		CategoryID:   t.CategoryID,
		WalletID:     t.WalletID,
		CategoryName: t.CategoryName,
		WalletName:   t.WalletName,
		Name:         t.Name,
		Note:         t.Note,
		Type:         string(t.Type),
		Amount:       t.Amount,
		Currency:     string(t.Currency),
		OccurredAt:   t.OccurredAt,
	}
}

// TransactionGroupResponseDTO: one day of the ledger, for the list screen only.
type TransactionGroupResponseDTO struct {
	OccurredAt time.Time                `json:"occurred_at" example:"2026-01-02T00:00:00Z"`
	Total      int64                    `json:"total" example:"-280000"`
	Items      []TransactionResponseDTO `json:"items"`
}

// TransactionCurrencyTotalResponseDTO: a currency the headline cannot state, left unconverted.
type TransactionCurrencyTotalResponseDTO struct {
	Currency string `json:"currency" example:"USD"`
	MoneyIn  int64  `json:"money_in" example:"118000"`
	MoneyOut int64  `json:"money_out" example:"-2000"`
	Net      int64  `json:"net" example:"116000"`
	Entries  int    `json:"entries" example:"2"`
}

// TransactionSummaryResponseDTO: totals over the whole filtered set, not the page.
type TransactionSummaryResponseDTO struct {
	BaseCurrency string                                `json:"base_currency" example:"IDR"`
	MoneyIn      int64                                 `json:"money_in" example:"21450000"`
	MoneyOut     int64                                 `json:"money_out" example:"-12780000"`
	Net          int64                                 `json:"net" example:"8670000"`
	OtherEntries int                                   `json:"other_entries" example:"2"`
	ByCurrency   []TransactionCurrencyTotalResponseDTO `json:"by_currency"`
}

func NewTransactionSummaryResponseDTO(s domain.TransactionSummary) TransactionSummaryResponseDTO {
	byCurrency := make([]TransactionCurrencyTotalResponseDTO, 0, len(s.ByCurrency))
	for _, item := range s.ByCurrency {
		byCurrency = append(byCurrency, TransactionCurrencyTotalResponseDTO{
			Currency: string(item.Currency),
			MoneyIn:  item.MoneyIn,
			MoneyOut: item.MoneyOut,
			Net:      item.Net,
			Entries:  item.Entries,
		})
	}

	return TransactionSummaryResponseDTO{
		BaseCurrency: string(s.BaseCurrency),
		MoneyIn:      s.MoneyIn,
		MoneyOut:     s.MoneyOut,
		Net:          s.Net,
		OtherEntries: s.OtherEntries,
		ByCurrency:   byCurrency,
	}
}

type TransactionListResponseDTO struct {
	Summary TransactionSummaryResponseDTO `json:"summary"`
	Groups  []TransactionGroupResponseDTO `json:"groups"`
}

// NewTransactionGroupResponseDTOs folds rows into one group per day, keeping the order the query returned.
func NewTransactionGroupResponseDTOs(ts []domain.Transaction) []TransactionGroupResponseDTO {
	groups := make([]TransactionGroupResponseDTO, 0, len(ts))
	for i := range ts {
		y, m, d := ts[i].OccurredAt.Date()
		day := time.Date(y, m, d, 0, 0, 0, 0, ts[i].OccurredAt.Location())

		item := NewTransactionResponseDTO(&ts[i])
		if n := len(groups); n > 0 && groups[n-1].OccurredAt.Equal(day) {
			groups[n-1].Items = append(groups[n-1].Items, item)
			groups[n-1].Total += item.Amount
			continue
		}
		groups = append(groups, TransactionGroupResponseDTO{
			OccurredAt: day,
			Total:      item.Amount,
			Items:      []TransactionResponseDTO{item},
		})
	}
	return groups
}
