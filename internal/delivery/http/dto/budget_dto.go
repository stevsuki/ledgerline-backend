package dto

import (
	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type CreateBudgetRequestDTO struct {
	CategoryID            uuid.UUID `json:"category_id" binding:"required" example:"b0000000-0000-0000-0000-000000000002"`
	Currency              string    `json:"currency" binding:"required" example:"IDR" enum:"IDR,USD,SGD"`
	MonthlyLimit          int64     `json:"monthly_limit" binding:"required" example:"10000000"`
	AlertThresholdPercent int       `json:"alert_threshold_percent" binding:"required" example:"80"`
	IsFixed               bool      `json:"is_fixed" example:"false"`
	Rollover              bool      `json:"rollover" example:"false"`
}

func (c CreateBudgetRequestDTO) ToInput() domain.CreateBudgetInput {
	return domain.CreateBudgetInput{
		CategoryID:            c.CategoryID,
		Currency:              domain.Currency(c.Currency),
		MonthlyLimit:          c.MonthlyLimit,
		AlertThresholdPercent: c.AlertThresholdPercent,
		IsFixed:               c.IsFixed,
		Rollover:              c.Rollover,
	}
}

type UpdateBudgetRequestDTO struct {
	CategoryID            *uuid.UUID `json:"category_id" binding:"omitempty" example:"b0000000-0000-0000-0000-000000000002"`
	Currency              *string    `json:"currency" binding:"omitempty" example:"IDR" enum:"IDR,USD,SGD"`
	MonthlyLimit          *int64     `json:"monthly_limit" binding:"omitempty" example:"10000000"`
	AlertThresholdPercent *int       `json:"alert_threshold_percent" binding:"omitempty" example:"80"`
	IsFixed               *bool      `json:"is_fixed" example:"false"`
	Rollover              *bool      `json:"rollover" example:"false"`
}

func (c UpdateBudgetRequestDTO) ToInput() domain.UpdateBudgetInput {
	input := domain.UpdateBudgetInput{
		CategoryID:            c.CategoryID,
		MonthlyLimit:          c.MonthlyLimit,
		AlertThresholdPercent: c.AlertThresholdPercent,
		IsFixed:               c.IsFixed,
		Rollover:              c.Rollover,
	}
	if c.Currency != nil {
		c := domain.Currency(*c.Currency)
		input.Currency = &c
	}
	return input
}

type BudgetResponseDTO struct {
	ID                    uuid.UUID `json:"id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryID            uuid.UUID `json:"category_id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryName          string    `json:"category_name" example:"Food & Drink"`
	Icon                  string    `json:"icon" example:"food"`
	Color                 string    `json:"color" example:"c2"`
	Currency              string    `json:"currency" example:"IDR" enum:"IDR,USD,SGD"`
	MonthlyLimit          int64     `json:"monthly_limit" example:"10000000"`
	Spent                 int64     `json:"spent" example:"8400000"`
	CarriedOver           int64     `json:"carried_over" example:"0"`
	Remaining             int64     `json:"remaining" example:"1600000"`
	UsedPercent           int       `json:"used_percent" example:"84"`
	IsOver                bool      `json:"is_over" example:"false"`
	AlertThresholdPercent int       `json:"alert_threshold_percent" example:"80"`
	IsFixed               bool      `json:"is_fixed" example:"false"`
	Rollover              bool      `json:"rollover" example:"false"`
}

func NewBudgetResponseDTO(budget *domain.Budget) BudgetResponseDTO {
	return BudgetResponseDTO{
		ID:                    budget.ID,
		CategoryID:            budget.CategoryID,
		CategoryName:          budget.CategoryName,
		Icon:                  budget.Icon,
		Color:                 budget.Color,
		Currency:              string(budget.Currency),
		MonthlyLimit:          budget.MonthlyLimit,
		Spent:                 budget.Spent,
		CarriedOver:           budget.CarriedOver,
		Remaining:             budget.Remaining(),
		UsedPercent:           budget.UsedPercent(),
		IsOver:                budget.IsOver(),
		AlertThresholdPercent: budget.AlertThresholdPercent,
		IsFixed:               budget.IsFixed,
		Rollover:              budget.Rollover,
	}
}

func NewBudgetResponseDTOs(budgets []domain.Budget) []BudgetResponseDTO {
	var response []BudgetResponseDTO
	for _, budget := range budgets {
		response = append(response, NewBudgetResponseDTO(&budget))
	}
	return response
}

type BudgetOverviewResponseDTO struct {
	Currency             string             `json:"currency" example:"IDR"`
	TotalBudgetAllocated int64              `json:"total_budget_allocated" example:"10000000"`
	TotalBudgetCarried   int64              `json:"total_budget_carried_over" example:"0"`
	TotalBudgetSpent     int64              `json:"total_budget_spent" example:"9000000"`
	TotalBudgetLeft      int64              `json:"total_budget_left" example:"1000000"`
	AcrossCategory       int                `json:"across_category" example:"6"`
	UncountedBudgets     int                `json:"uncounted_budgets" example:"0"`
	UsedPercent          int                `json:"used_percent" example:"90"`
	IsOver               bool               `json:"is_over" example:"false"`
	DaysLeft             int                `json:"days_left" example:"4"`
	CycleElapsedPercent  int                `json:"cycle_elapsed_percent" example:"87"`
	Shares               []BudgetShareDTO   `json:"shares"`
	Alert                []NeedAttentionDTO `json:"alert"`
}

// BudgetShareDTO: one slice of the allocation bar.
type BudgetShareDTO struct {
	CategoryID   uuid.UUID `json:"category_id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryName string    `json:"category_name" example:"Food & Drink"`
	Color        string    `json:"color" example:"c2"`
	Percent      int       `json:"percent" example:"24"`
}

// NeedAttentionDTO: figures only; the client writes the wording.
type NeedAttentionDTO struct {
	BudgetID              uuid.UUID `json:"budget_id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryID            uuid.UUID `json:"category_id" example:"b0000000-0000-0000-0000-000000000002"`
	CategoryName          string    `json:"category_name" example:"Food & Drink"`
	Icon                  string    `json:"icon" example:"food"`
	Color                 string    `json:"color" example:"c2"`
	MonthlyLimit          int64     `json:"monthly_limit" example:"4000000"`
	Spent                 int64     `json:"spent" example:"4320000"`
	CarriedOver           int64     `json:"carried_over" example:"0"`
	Remaining             int64     `json:"remaining" example:"-320000"`
	UsedPercent           int       `json:"used_percent" example:"108"`
	AlertThresholdPercent int       `json:"alert_threshold_percent" example:"80"`
	IsFixed               bool      `json:"is_fixed" example:"false"`
	IsOver                bool      `json:"is_over" example:"true"`
}

func NewBudgetOverviewResponseDTO(o domain.BudgetOverview) BudgetOverviewResponseDTO {
	shares := make([]BudgetShareDTO, 0, len(o.Shares))
	for _, share := range o.Shares {
		shares = append(shares, BudgetShareDTO{
			CategoryID:   share.CategoryID,
			CategoryName: share.CategoryName,
			Color:        share.Color,
			Percent:      share.Percent,
		})
	}

	alert := make([]NeedAttentionDTO, 0, len(o.NeedAttention))
	for _, item := range o.NeedAttention {
		alert = append(alert, NeedAttentionDTO{
			BudgetID:              item.BudgetID,
			CategoryID:            item.CategoryID,
			CategoryName:          item.CategoryName,
			Icon:                  item.Icon,
			Color:                 item.Color,
			MonthlyLimit:          item.MonthlyLimit,
			Spent:                 item.Spent,
			CarriedOver:           item.CarriedOver,
			Remaining:             item.Remaining,
			UsedPercent:           item.UsedPercent,
			AlertThresholdPercent: item.AlertThresholdPercent,
			IsFixed:               item.IsFixed,
			IsOver:                item.IsOver,
		})
	}

	return BudgetOverviewResponseDTO{
		Currency:             string(o.Currency),
		TotalBudgetAllocated: o.TotalAllocated,
		TotalBudgetCarried:   o.TotalCarriedOver,
		TotalBudgetSpent:     o.TotalSpent,
		TotalBudgetLeft:      o.TotalLeft,
		AcrossCategory:       o.CategoryCount,
		UncountedBudgets:     o.UncountedBudgets,
		UsedPercent:          o.UsedPercent,
		IsOver:               o.IsOver,
		DaysLeft:             o.DaysLeft,
		CycleElapsedPercent:  o.CycleElapsedPercent,
		Shares:               shares,
		Alert:                alert,
	}
}
