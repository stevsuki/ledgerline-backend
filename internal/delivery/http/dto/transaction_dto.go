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

/* ── the dashboard ─────────────────────────────────────────────────────── */

// CategorySpendResponseDTO: one slice of a month's spending, stated positive.
type CategorySpendResponseDTO struct {
	CategoryID uuid.UUID `json:"category_id" example:"b0000000-0000-0000-0000-000000000002"`
	Name       string    `json:"name" example:"Food & Drink"`
	Icon       string    `json:"icon" example:"food"`
	Color      string    `json:"color" example:"c2"`
	Spent      int64     `json:"spent" example:"2700000"`
}

// TransactionPeriodResponseDTO: one period. `money_out` stays negative, as the rows are
// stored, so `net` is the two added together rather than a subtraction the client invents.
type TransactionPeriodResponseDTO struct {
	MoneyIn        int64                      `json:"money_in" example:"18400000"`
	MoneyOut       int64                      `json:"money_out" example:"-6958000"`
	Net            int64                      `json:"net" example:"11442000"`
	IncomeEntries  int                        `json:"income_entries" example:"14"`
	ExpenseEntries int                        `json:"expense_entries" example:"40"`
	ByCategory     []CategorySpendResponseDTO `json:"by_category"`
}

// TransactionOverviewResponseDTO: the dashboard's month beside the month it is compared
// with. Figures only — every percentage the cards print is a ratio of two of these, and
// the client decides how to say "no comparison" when the earlier month is empty.
type TransactionOverviewResponseDTO struct {
	BaseCurrency string                       `json:"base_currency" example:"IDR"`
	Month        string                       `json:"month" example:"2026-09"`
	Current      TransactionPeriodResponseDTO `json:"current"`
	Previous     TransactionPeriodResponseDTO `json:"previous"`
	OtherEntries int                          `json:"other_entries" example:"2"`
}

const monthLayout = "2006-01"

func newTransactionPeriodResponseDTO(period domain.TransactionPeriodTotals) TransactionPeriodResponseDTO {
	categories := make([]CategorySpendResponseDTO, 0, len(period.ByCategory))
	for _, entry := range period.ByCategory {
		categories = append(categories, CategorySpendResponseDTO{
			CategoryID: entry.CategoryID,
			Name:       entry.Name,
			Icon:       entry.Icon,
			Color:      entry.Color,
			Spent:      entry.Spent,
		})
	}

	return TransactionPeriodResponseDTO{
		MoneyIn:        period.MoneyIn,
		MoneyOut:       period.MoneyOut,
		Net:            period.Net,
		IncomeEntries:  period.IncomeEntries,
		ExpenseEntries: period.ExpenseEntries,
		ByCategory:     categories,
	}
}

func NewTransactionOverviewResponseDTO(overview domain.TransactionOverview) TransactionOverviewResponseDTO {
	return TransactionOverviewResponseDTO{
		BaseCurrency: string(overview.BaseCurrency),
		Month:        overview.PeriodStart.Format(monthLayout),
		Current:      newTransactionPeriodResponseDTO(overview.Current),
		Previous:     newTransactionPeriodResponseDTO(overview.Previous),
		OtherEntries: overview.OtherEntries,
	}
}

// TransactionTrendPointResponseDTO: one bar. It carries the period it starts on rather than
// a label, because naming a week or a month is wording, and wording belongs to the client.
type TransactionTrendPointResponseDTO struct {
	Start    time.Time `json:"start" example:"2026-09-01T00:00:00+07:00"`
	MoneyIn  int64     `json:"money_in" example:"4200000"`
	MoneyOut int64     `json:"money_out" example:"-1740000"`
}

type TransactionTrendResponseDTO struct {
	BaseCurrency string                             `json:"base_currency" example:"IDR"`
	Range        string                             `json:"range" example:"monthly" enum:"weekly,monthly"`
	Points       []TransactionTrendPointResponseDTO `json:"points"`
}

func NewTransactionTrendResponseDTO(
	trendRange domain.TransactionTrendRange, points []domain.TransactionTrendPoint,
) TransactionTrendResponseDTO {
	out := make([]TransactionTrendPointResponseDTO, 0, len(points))
	for _, point := range points {
		out = append(out, TransactionTrendPointResponseDTO{
			Start:    point.Start,
			MoneyIn:  point.MoneyIn,
			MoneyOut: point.MoneyOut,
		})
	}

	return TransactionTrendResponseDTO{
		BaseCurrency: string(domain.BaseCurrency),
		Range:        string(trendRange),
		Points:       out,
	}
}

// OverviewTransactionsQuery: which month to report. Absent means the month the server is in,
// which is what the dashboard asks for.
type OverviewTransactionsQuery struct {
	Month string `form:"month" binding:"omitempty,len=7" example:"2026-09"`
}

// Period answers the month asked for, or the current one when the field is empty.
func (q OverviewTransactionsQuery) Period(now time.Time) (time.Time, error) {
	if q.Month == "" {
		return now, nil
	}

	month, err := time.ParseInLocation(monthLayout, q.Month, now.Location())
	if err != nil {
		return time.Time{}, domain.InvalidInput(domain.CodeTransactionInvalid,
			"month must be written as YYYY-MM").WithField("month")
	}
	return month, nil
}

// TrendTransactionsQuery: which window the chart draws.
type TrendTransactionsQuery struct {
	Range string `form:"range" binding:"omitempty,oneof=weekly monthly" example:"monthly"`
}

func (q TrendTransactionsQuery) TrendRange() domain.TransactionTrendRange {
	if q.Range == "" {
		return domain.TrendRangeWeekly
	}
	return domain.TransactionTrendRange(q.Range)
}
