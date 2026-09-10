package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type BudgetModel struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID                uuid.UUID `gorm:"type:uuid;not null;index"`
	CategoryID            uuid.UUID `gorm:"type:uuid;not null;index"`
	Currency              string    `gorm:"size:3;not null"`
	MonthlyLimit          int64     `gorm:"not null"`
	AlertThresholdPercent int       `gorm:"not null;default:80"`
	IsFixed               bool      `gorm:"not null;default:false"`
	Rollover              bool      `gorm:"not null;default:false"`
	CreatedAt             time.Time
	CreatedBy             *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt             time.Time
	UpdatedBy             *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt             gorm.DeletedAt `gorm:"index"`
	DeletedBy             *uuid.UUID     `gorm:"type:uuid"`
}

func (BudgetModel) TableName() string { return "budgets" }

// BudgetRow: a budget read back with its category columns.
type BudgetRow struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	CategoryID            uuid.UUID
	CategoryName          string
	Icon                  string
	Color                 string
	Currency              string
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

func (b *BudgetRow) ToDomain() *domain.Budget {
	return &domain.Budget{
		ID:                    b.ID,
		UserID:                b.UserID,
		CategoryID:            b.CategoryID,
		CategoryName:          b.CategoryName,
		Icon:                  b.Icon,
		Color:                 b.Color,
		Currency:              domain.Currency(b.Currency),
		MonthlyLimit:          b.MonthlyLimit,
		Spent:                 b.Spent,
		CarriedOver:           b.CarriedOver,
		AlertThresholdPercent: b.AlertThresholdPercent,
		IsFixed:               b.IsFixed,
		Rollover:              b.Rollover,
		CreatedAt:             b.CreatedAt,
		CreatedBy:             b.CreatedBy,
		UpdatedAt:             b.UpdatedAt,
		UpdatedBy:             b.UpdatedBy,
		DeletedBy:             b.DeletedBy,
	}
}

func BudgetFromDomain(budget *domain.Budget) BudgetModel {
	return BudgetModel{
		ID:                    budget.ID,
		UserID:                budget.UserID,
		CategoryID:            budget.CategoryID,
		Currency:              string(budget.Currency),
		MonthlyLimit:          budget.MonthlyLimit,
		AlertThresholdPercent: budget.AlertThresholdPercent,
		IsFixed:               budget.IsFixed,
		Rollover:              budget.Rollover,
		CreatedAt:             budget.CreatedAt,
		CreatedBy:             budget.CreatedBy,
		UpdatedAt:             budget.UpdatedAt,
		UpdatedBy:             budget.UpdatedBy,
	}
}

func BudgetRowsToDomain(rows []BudgetRow) []domain.Budget {
	out := make([]domain.Budget, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.ToDomain())
	}
	return out
}
