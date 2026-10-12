package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

const dateLayout = "2006-01-02"

func parseDate(value string) time.Time {
	date, _ := time.Parse(dateLayout, value)
	return date
}

type CreateRecurringRequestDTO struct {
	CategoryID  uuid.UUID `json:"category_id" binding:"required" example:"b0000000-0000-0000-0000-000000000002"`
	WalletID    uuid.UUID `json:"wallet_id" binding:"required" example:"b0000000-0000-0000-0000-000000000002"`
	Name        string    `json:"name" binding:"required,max=100" example:"Biznet Home"`
	Amount      int64     `json:"amount" binding:"required" example:"425000"`
	Frequency   string    `json:"frequency" binding:"required" example:"monthly" enum:"weekly,monthly,quarterly,annual"`
	NextDueDate string    `json:"next_due_date" binding:"required,datetime=2006-01-02" example:"2026-09-01"`
	IsPaused    bool      `json:"is_paused" example:"false"`
}

func (c CreateRecurringRequestDTO) ToInput() domain.CreateRecurringInput {
	return domain.CreateRecurringInput{
		CategoryID:  c.CategoryID,
		WalletID:    c.WalletID,
		Name:        c.Name,
		Amount:      c.Amount,
		Frequency:   domain.RecurringFrequency(c.Frequency),
		NextDueDate: parseDate(c.NextDueDate),
		IsPaused:    c.IsPaused,
	}
}

type UpdateRecurringRequestDTO struct {
	CategoryID  *uuid.UUID `json:"category_id" binding:"omitempty" example:"b0000000-0000-0000-0000-000000000002"`
	WalletID    *uuid.UUID `json:"wallet_id" binding:"omitempty" example:"b0000000-0000-0000-0000-000000000002"`
	Name        *string    `json:"name" binding:"omitempty,max=100" example:"Biznet Home"`
	Amount      *int64     `json:"amount" binding:"omitempty" example:"425000"`
	Frequency   *string    `json:"frequency" binding:"omitempty" example:"monthly" enum:"weekly,monthly,quarterly,annual"`
	NextDueDate *string    `json:"next_due_date" binding:"omitempty,datetime=2006-01-02" example:"2026-09-01"`
	IsPaused    *bool      `json:"is_paused" example:"false"`
}

func (c UpdateRecurringRequestDTO) ToInput() domain.UpdateRecurringInput {
	input := domain.UpdateRecurringInput{
		CategoryID: c.CategoryID,
		WalletID:   c.WalletID,
		Name:       c.Name,
		Amount:     c.Amount,
		IsPaused:   c.IsPaused,
	}
	if c.Frequency != nil {
		frequency := domain.RecurringFrequency(*c.Frequency)
		input.Frequency = &frequency
	}
	if c.NextDueDate != nil {
		date := parseDate(*c.NextDueDate)
		input.NextDueDate = &date
	}
	return input
}

type ConfirmRecurringRequestDTO struct {
	Amount     *int64     `json:"amount" binding:"omitempty" example:"425000"`
	OccurredAt *time.Time `json:"occurred_at" binding:"omitempty" example:"2026-09-01T09:00:00+07:00"`
	Note       string     `json:"note" binding:"omitempty,max=255" example:"September bill"`
}

func (c ConfirmRecurringRequestDTO) ToInput() domain.ConfirmRecurringInput {
	return domain.ConfirmRecurringInput{Amount: c.Amount, OccurredAt: c.OccurredAt, Note: c.Note}
}

type OverviewRecurringQuery struct {
	Date string `form:"date" binding:"omitempty,datetime=2006-01-02" example:"2026-08-27"`
}

// Today: the caller's own day when it states one, else the server's.
func (q OverviewRecurringQuery) Today(now time.Time) time.Time {
	if q.Date == "" {
		return now
	}
	return parseDate(q.Date)
}

type RecurringResponseDTO struct {
	ID           uuid.UUID `json:"id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryID   uuid.UUID `json:"category_id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryName string    `json:"category_name" example:"Utilities"`
	WalletID     uuid.UUID `json:"wallet_id" example:"b0000000-0000-0000-0000-000000000002"`
	WalletName   string    `json:"wallet_name" example:"BCA Payroll"`
	Icon         string    `json:"icon" example:"wifi"`
	Name         string    `json:"name" example:"Biznet Home"`
	Amount       int64     `json:"amount" example:"425000"`
	Currency     string    `json:"currency" example:"IDR" enum:"IDR,USD,SGD"`
	Frequency    string    `json:"frequency" example:"monthly" enum:"weekly,monthly,quarterly,annual"`
	NextDueDate  string    `json:"next_due_date" example:"2026-09-01"`
	IsPaused     bool      `json:"is_paused" example:"false"`
}

func NewRecurringResponseDTO(r *domain.Recurring) RecurringResponseDTO {
	return RecurringResponseDTO{
		ID:           r.ID,
		CategoryID:   r.CategoryID,
		CategoryName: r.CategoryName,
		WalletID:     r.WalletID,
		WalletName:   r.WalletName,
		Icon:         r.Icon,
		Name:         r.Name,
		Amount:       r.Amount,
		Currency:     string(r.Currency),
		Frequency:    string(r.Frequency),
		NextDueDate:  r.NextDueDate.Format(dateLayout),
		IsPaused:     r.IsPaused,
	}
}

func NewRecurringResponseDTOs(items []domain.Recurring) []RecurringResponseDTO {
	response := make([]RecurringResponseDTO, 0, len(items))
	for i := range items {
		response = append(response, NewRecurringResponseDTO(&items[i]))
	}
	return response
}

type RecurringOverviewResponseDTO struct {
	Currency          string `json:"currency" example:"IDR"`
	MonthlyCommitment int64  `json:"monthly_commitment" example:"5705000"`
	DueSoonTotal      int64  `json:"due_soon_total" example:"5105000"`
	DueSoonCount      int    `json:"due_soon_count" example:"3"`
	DueSoonDays       int    `json:"due_soon_days" example:"7"`
	OverdueCount      int    `json:"overdue_count" example:"0"`
	ActiveCount       int    `json:"active_count" example:"6"`
	PausedCount       int    `json:"paused_count" example:"1"`
	UncountedItems    int    `json:"uncounted_items" example:"0"`
}

func NewRecurringOverviewResponseDTO(o domain.RecurringOverview) RecurringOverviewResponseDTO {
	return RecurringOverviewResponseDTO{
		Currency:          string(o.Currency),
		MonthlyCommitment: o.MonthlyCommitment,
		DueSoonTotal:      o.DueSoonTotal,
		DueSoonCount:      o.DueSoonCount,
		DueSoonDays:       domain.RecurringDueSoonDays,
		OverdueCount:      o.OverdueCount,
		ActiveCount:       o.ActiveCount,
		PausedCount:       o.PausedCount,
		UncountedItems:    o.UncountedItems,
	}
}
