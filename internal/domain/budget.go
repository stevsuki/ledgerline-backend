package domain

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
)

// Share of a limit an alert fires at; the table checks the same range.
const (
	MinAlertThresholdPercent = 1
	MaxAlertThresholdPercent = 100
	FullPercent              = 100
)

type Budget struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	Currency              Currency
	MonthlyLimit          int64
	Spent                 int64
	AlertThresholdPercent int
	IsFixed               bool
	Rollover              bool
	CreatedAt             time.Time
	CreatedBy             *uuid.UUID
	UpdatedAt             time.Time
	UpdatedBy             *uuid.UUID
	DeletedBy             *uuid.UUID
}

// PercentOf: the one rounding rule for every share this package reports, so a
// card and the headline above it can never round the same figure differently.
func PercentOf(part, whole int64) int {
	if whole <= 0 {
		return 0
	}
	return int(math.Round(float64(part) / float64(whole) * FullPercent))
}

// Remaining is negative once the limit is passed.
func (b Budget) Remaining() int64 { return b.MonthlyLimit - b.Spent }

func (b Budget) UsedPercent() int { return PercentOf(b.Spent, b.MonthlyLimit) }

func (b Budget) IsOver() bool { return b.Spent > b.MonthlyLimit }

type BudgetRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Budget, error)
	Create(ctx context.Context, budget *Budget) error
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Usage(ctx context.Context, userID uuid.UUID) ([]BudgetUsage, error)
	// ExistsByCategory: whether a live budget still limits this category.
	ExistsByCategory(ctx context.Context, categoryID, userID uuid.UUID) (bool, error)
}

type CreateBudgetInput struct {
	CategoryID            uuid.UUID
	Currency              Currency
	MonthlyLimit          int64
	AlertThresholdPercent int
	IsFixed               bool
	Rollover              bool
}

type UpdateBudgetInput struct {
	CategoryID            *uuid.UUID
	Currency              *Currency
	MonthlyLimit          *int64
	AlertThresholdPercent *int
	IsFixed               *bool
	Rollover              *bool
}

type BudgetService interface {
	List(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Budget, error)
	Create(ctx context.Context, userID uuid.UUID, input CreateBudgetInput) (*Budget, error)
	Update(ctx context.Context, userID, id uuid.UUID, input UpdateBudgetInput) (*Budget, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Overview(ctx context.Context, userID uuid.UUID) (BudgetOverview, error)
}

// BudgetUsage: one budget joined to the category it limits, with what that
// category has spent this cycle.
type BudgetUsage struct {
	BudgetID              uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	Currency              Currency
	MonthlyLimit          int64
	Spent                 int64
	AlertThresholdPercent int
	IsFixed               bool
	Rollover              bool
}

// BudgetShare: one slice of the allocation bar. Percent is of the total limit,
// never of the spend, and the slices always add up to 100.
type BudgetShare struct {
	CategoryID   uuid.UUID
	CategoryName string
	Color        string
	Percent      int
}

// BudgetAttention: a budget that is over its limit or has reached its own
// alert threshold. Figures only; the wording belongs to the client.
type BudgetAttention struct {
	BudgetID              uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	MonthlyLimit          int64
	Spent                 int64
	Remaining             int64
	UsedPercent           int
	AlertThresholdPercent int
	IsFixed               bool
	IsOver                bool
}

// BudgetOverview: the summary panel on the budgets screen. Only base-currency
// budgets are summed, the way WalletOverview keeps its headline in one currency.
type BudgetOverview struct {
	Currency       Currency
	TotalAllocated int64
	TotalSpent     int64
	// Negative once the allocation is passed.
	TotalLeft     int64
	CategoryCount int
	// Budgets in another currency, summed nowhere: there is no rate to fold them in with.
	UncountedBudgets int
	UsedPercent      int
	IsOver           bool
	// Days between today and the last day of the month.
	DaysLeft int
	// How much of the cycle has run, which is what UsedPercent is early or late against.
	CycleElapsedPercent int
	Shares              []BudgetShare
	NeedAttention       []BudgetAttention
}
