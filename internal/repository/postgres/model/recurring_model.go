package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type RecurringModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	CategoryID  uuid.UUID `gorm:"type:uuid;not null;index"`
	WalletID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"size:100;not null"`
	Amount      int64     `gorm:"not null"`
	Currency    string    `gorm:"size:3;not null"`
	Frequency   string    `gorm:"size:10;not null"`
	NextDueDate time.Time `gorm:"type:date;not null"`
	AnchorDay   int       `gorm:"not null"`
	IsPaused    bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time
	CreatedBy   *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	DeletedBy   *uuid.UUID     `gorm:"type:uuid"`
}

func (RecurringModel) TableName() string { return "recurrings" }

// RecurringRow: a recurring item read back with its category and wallet columns.
type RecurringRow struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CategoryID   uuid.UUID
	WalletID     uuid.UUID
	CategoryName string
	WalletName   string
	Icon         string
	Name         string
	Amount       int64
	Currency     string
	Frequency    string
	NextDueDate  time.Time
	AnchorDay    int
	IsPaused     bool
	CreatedAt    time.Time
	CreatedBy    *uuid.UUID
	UpdatedAt    time.Time
	UpdatedBy    *uuid.UUID
	DeletedBy    *uuid.UUID
}

func (r *RecurringRow) ToDomain() *domain.Recurring {
	return &domain.Recurring{
		ID:           r.ID,
		UserID:       r.UserID,
		CategoryID:   r.CategoryID,
		WalletID:     r.WalletID,
		CategoryName: r.CategoryName,
		WalletName:   r.WalletName,
		Icon:         r.Icon,
		Name:         r.Name,
		Amount:       r.Amount,
		Currency:     domain.Currency(r.Currency),
		Frequency:    domain.RecurringFrequency(r.Frequency),
		NextDueDate:  r.NextDueDate,
		AnchorDay:    r.AnchorDay,
		IsPaused:     r.IsPaused,
		CreatedAt:    r.CreatedAt,
		CreatedBy:    r.CreatedBy,
		UpdatedAt:    r.UpdatedAt,
		UpdatedBy:    r.UpdatedBy,
		DeletedBy:    r.DeletedBy,
	}
}

func RecurringFromDomain(recurring *domain.Recurring) RecurringModel {
	return RecurringModel{
		ID:          recurring.ID,
		UserID:      recurring.UserID,
		CategoryID:  recurring.CategoryID,
		WalletID:    recurring.WalletID,
		Name:        recurring.Name,
		Amount:      recurring.Amount,
		Currency:    string(recurring.Currency),
		Frequency:   string(recurring.Frequency),
		NextDueDate: recurring.NextDueDate,
		AnchorDay:   recurring.AnchorDay,
		IsPaused:    recurring.IsPaused,
		CreatedAt:   recurring.CreatedAt,
		CreatedBy:   recurring.CreatedBy,
		UpdatedAt:   recurring.UpdatedAt,
		UpdatedBy:   recurring.UpdatedBy,
	}
}

func RecurringRowsToDomain(rows []RecurringRow) []domain.Recurring {
	out := make([]domain.Recurring, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.ToDomain())
	}
	return out
}
