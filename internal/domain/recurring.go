package domain

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
)

type RecurringFrequency string

const (
	RecurringWeekly    RecurringFrequency = "weekly"
	RecurringMonthly   RecurringFrequency = "monthly"
	RecurringQuarterly RecurringFrequency = "quarterly"
	RecurringAnnual    RecurringFrequency = "annual"
)

const (
	daysPerWeek      = 7
	monthsPerQuarter = 3
	monthsPerYear    = 12
	weeksPerYear     = 52

	// RecurringDueSoonDays: how far ahead "due soon" looks.
	RecurringDueSoonDays = 7
)

func (f RecurringFrequency) Valid() bool {
	switch f {
	case RecurringWeekly, RecurringMonthly, RecurringQuarterly, RecurringAnnual:
		return true
	}
	return false
}

func (f RecurringFrequency) months() int {
	switch f {
	case RecurringQuarterly:
		return monthsPerQuarter
	case RecurringAnnual:
		return monthsPerYear
	default:
		return 1
	}
}

// Next: the due date after from; anchorDay keeps a 31st from drifting to the 28th.
func (f RecurringFrequency) Next(from time.Time, anchorDay int) time.Time {
	if f == RecurringWeekly {
		return from.AddDate(0, 0, daysPerWeek)
	}

	first := time.Date(from.Year(), from.Month()+time.Month(f.months()), 1, 0, 0, 0, 0, time.UTC)
	lastDay := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(anchorDay, lastDay)-1)
}

// MonthlyAmount: what one occurrence costs spread over a month.
func (f RecurringFrequency) MonthlyAmount(amount int64) int64 {
	if f == RecurringWeekly {
		return int64(math.Round(float64(amount) * weeksPerYear / monthsPerYear))
	}
	return int64(math.Round(float64(amount) / float64(f.months())))
}

type Recurring struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CategoryID   uuid.UUID
	WalletID     uuid.UUID
	CategoryName string
	WalletName   string
	Icon         string
	Name         string
	Amount       int64
	Currency     Currency
	Frequency    RecurringFrequency
	NextDueDate  time.Time
	AnchorDay    int
	IsPaused     bool
	CreatedAt    time.Time
	CreatedBy    *uuid.UUID
	UpdatedAt    time.Time
	UpdatedBy    *uuid.UUID
	DeletedBy    *uuid.UUID
}

// RecurringOverview: the cards above the list, base currency only.
type RecurringOverview struct {
	Currency          Currency
	MonthlyCommitment int64
	DueSoonTotal      int64
	DueSoonCount      int
	OverdueCount      int
	ActiveCount       int
	PausedCount       int
	UncountedItems    int
}

type RecurringRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Recurring, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Recurring, error)
	Create(ctx context.Context, recurring *Recurring) error
	Update(ctx context.Context, recurring *Recurring) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type CreateRecurringInput struct {
	CategoryID  uuid.UUID
	WalletID    uuid.UUID
	Name        string
	Amount      int64
	Frequency   RecurringFrequency
	NextDueDate time.Time
	IsPaused    bool
}

type UpdateRecurringInput struct {
	CategoryID  *uuid.UUID
	WalletID    *uuid.UUID
	Name        *string
	Amount      *int64
	Frequency   *RecurringFrequency
	NextDueDate *time.Time
	IsPaused    *bool
}

// ConfirmRecurringInput: what the person may correct before the payment is recorded.
type ConfirmRecurringInput struct {
	Amount     *int64
	OccurredAt *time.Time
	Note       string
}

type RecurringService interface {
	List(ctx context.Context, userID uuid.UUID) ([]Recurring, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Recurring, error)
	Create(ctx context.Context, userID uuid.UUID, input CreateRecurringInput) (*Recurring, error)
	Update(ctx context.Context, userID, id uuid.UUID, input UpdateRecurringInput) (*Recurring, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Overview(ctx context.Context, userID uuid.UUID, today time.Time) (RecurringOverview, error)
	Confirm(ctx context.Context, userID, id uuid.UUID, input ConfirmRecurringInput) (*Recurring, error)
}
