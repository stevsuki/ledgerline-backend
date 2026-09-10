package domain

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
)

// Alert threshold range; the table checks the same.
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
	CarriedOver           int64
	AlertThresholdPercent int
	IsFixed               bool
	Rollover              bool
	CreatedAt             time.Time
	CreatedBy             *uuid.UUID
	UpdatedAt             time.Time
	UpdatedBy             *uuid.UUID
	DeletedBy             *uuid.UUID
}

// PercentOf: the one rounding rule every share here uses.
func PercentOf(part, whole int64) int {
	if whole <= 0 {
		return 0
	}
	return int(math.Round(float64(part) / float64(whole) * FullPercent))
}

// EffectiveLimit: what this cycle may spend. A rollover budget carries last cycle's leftover,
// so the carry widens what is left to spend without touching the limit that was planned.
func (b Budget) EffectiveLimit() int64 { return b.MonthlyLimit + b.CarriedOver }

// Remaining: negative once the limit is passed.
func (b Budget) Remaining() int64 { return b.EffectiveLimit() - b.Spent }

func (b Budget) UsedPercent() int { return PercentOf(b.Spent, b.EffectiveLimit()) }

func (b Budget) IsOver() bool { return b.Spent > b.EffectiveLimit() }

type BudgetRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Budget, error)
	Create(ctx context.Context, budget *Budget) error
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Usage(ctx context.Context, userID uuid.UUID) ([]BudgetUsage, error)
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

// BudgetUsage: a budget with its category and its spend this cycle.
type BudgetUsage struct {
	BudgetID              uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	Currency              Currency
	MonthlyLimit          int64
	Spent                 int64
	CarriedOver           int64
	AlertThresholdPercent int
	IsFixed               bool
	Rollover              bool
}

func (u BudgetUsage) EffectiveLimit() int64 { return u.MonthlyLimit + u.CarriedOver }

// BudgetShare: one slice of the allocation bar.
type BudgetShare struct {
	CategoryID   uuid.UUID
	CategoryName string
	Color        string
	Percent      int
}

// BudgetAttention: a budget over its limit or past its threshold.
type BudgetAttention struct {
	BudgetID              uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	MonthlyLimit          int64
	Spent                 int64
	CarriedOver           int64
	Remaining             int64
	UsedPercent           int
	AlertThresholdPercent int
	IsFixed               bool
	IsOver                bool
}

// BudgetOverview: the summary panel, base currency only.
type BudgetOverview struct {
	Currency            Currency
	TotalAllocated      int64
	TotalCarriedOver    int64
	TotalSpent          int64
	TotalLeft           int64
	CategoryCount       int
	UncountedBudgets    int
	UsedPercent         int
	IsOver              bool
	DaysLeft            int
	CycleElapsedPercent int
	Shares              []BudgetShare
	NeedAttention       []BudgetAttention
}
