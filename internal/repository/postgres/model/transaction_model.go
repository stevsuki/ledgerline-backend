package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"gorm.io/gorm"
)

type TransactionModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index"`
	WalletID   uuid.UUID `gorm:"type:uuid;not null;index"`

	CategoryName string `gorm:"->"`
	WalletName   string `gorm:"->"`

	Name       string    `gorm:"size:150;not null"`
	Note       string    `gorm:"size:255"`
	Type       string    `gorm:"size:20;not null"`
	Amount     int64     `gorm:"not null"`
	Currency   string    `gorm:"size:3;not null"`
	OccurredAt time.Time `gorm:"not null"`
	CreatedAt  time.Time
	CreatedBy  *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt  time.Time
	UpdatedBy  *uuid.UUID     `gorm:"type:uuid"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	DeletedBy  *uuid.UUID     `gorm:"type:uuid"`
}

func (TransactionModel) TableName() string {
	return "transactions"
}

func (t *TransactionModel) ToDomain() *domain.Transaction {
	return &domain.Transaction{
		ID:           t.ID,
		UserID:       t.UserID,
		CategoryID:   t.CategoryID,
		WalletID:     t.WalletID,
		CategoryName: t.CategoryName,
		WalletName:   t.WalletName,
		Name:         t.Name,
		Note:         t.Note,
		Type:         domain.TransactionType(t.Type),
		Amount:       t.Amount,
		Currency:     domain.Currency(t.Currency),
		OccurredAt:   t.OccurredAt,
		CreatedAt:    t.CreatedAt,
		CreatedBy:    t.CreatedBy,
		UpdatedAt:    t.UpdatedAt,
		UpdatedBy:    t.UpdatedBy,
		DeletedBy:    t.DeletedBy,
	}
}

func TransactionFromDomain(t *domain.Transaction) TransactionModel {
	return TransactionModel{
		ID:         t.ID,
		UserID:     t.UserID,
		CategoryID: t.CategoryID,
		WalletID:   t.WalletID,
		Name:       t.Name,
		Note:       t.Note,
		Type:       string(t.Type),
		Amount:     t.Amount,
		Currency:   string(t.Currency),
		OccurredAt: t.OccurredAt,
		CreatedAt:  t.CreatedAt,
		CreatedBy:  t.CreatedBy,
		UpdatedAt:  t.UpdatedAt,
		UpdatedBy:  t.UpdatedBy,
		DeletedBy:  t.DeletedBy,
	}
}

func TransactionsToDomain(ts []TransactionModel) []domain.Transaction {
	domainTs := make([]domain.Transaction, len(ts))
	for i, t := range ts {
		domainTs[i] = *t.ToDomain()
	}
	return domainTs
}
