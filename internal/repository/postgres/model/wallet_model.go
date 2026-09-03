package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// WalletModel: sizes and nullability follow migration 000018.
type WalletModel struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Name   string    `gorm:"size:100;not null"`
	Type   string    `gorm:"size:20;not null"`
	// currency is VARCHAR(3), not 20.
	Currency string `gorm:"size:3;not null"`
	// reference and icon are nullable in the table.
	Reference        string `gorm:"size:50"`
	Icon             string `gorm:"size:50"`
	Balance          int64  `gorm:"not null;default:0"`
	BalanceUpdatedAt time.Time
	BalanceUpdatedBy *uuid.UUID `gorm:"type:uuid"`
	IncludeInTotal   bool       `gorm:"not null;default:true"`
	// Cards only.
	CreditLimit *int64
	DueDay      *int `gorm:"type:smallint"`
	CreatedAt   time.Time
	CreatedBy   *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	DeletedBy   *uuid.UUID     `gorm:"type:uuid"`
}

func (WalletModel) TableName() string { return "wallets" }

func (m WalletModel) ToDomain() *domain.Wallet {
	return &domain.Wallet{
		ID:               m.ID,
		UserID:           m.UserID,
		Name:             m.Name,
		Type:             domain.WalletType(m.Type),
		Icon:             m.Icon,
		Currency:         domain.Currency(m.Currency),
		Reference:        m.Reference,
		Balance:          m.Balance,
		BalanceUpdatedAt: m.BalanceUpdatedAt,
		BalanceUpdatedBy: m.BalanceUpdatedBy,
		IncludeInTotal:   m.IncludeInTotal,
		CreditLimit:      m.CreditLimit,
		DueDay:           m.DueDay,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		CreatedBy:        m.CreatedBy,
		UpdatedBy:        m.UpdatedBy,
		DeletedBy:        m.DeletedBy,
	}
}

func WalletFromDomain(w *domain.Wallet) WalletModel {
	return WalletModel{
		ID:               w.ID,
		UserID:           w.UserID,
		Name:             w.Name,
		Type:             string(w.Type),
		Icon:             w.Icon,
		Currency:         string(w.Currency),
		Reference:        w.Reference,
		Balance:          w.Balance,
		BalanceUpdatedAt: w.BalanceUpdatedAt,
		BalanceUpdatedBy: w.BalanceUpdatedBy,
		IncludeInTotal:   w.IncludeInTotal,
		CreditLimit:      w.CreditLimit,
		DueDay:           w.DueDay,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		CreatedBy:        w.CreatedBy,
		UpdatedBy:        w.UpdatedBy,
	}
}

func WalletsToDomain(models []WalletModel) []domain.Wallet {
	out := make([]domain.Wallet, 0, len(models))
	for _, m := range models {
		out = append(out, *m.ToDomain())
	}
	return out
}
