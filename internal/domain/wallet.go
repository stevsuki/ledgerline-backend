package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WalletType string

const (
	WalletTypeBank    WalletType = "bank"
	WalletTypeEwallet WalletType = "ewallet"
	WalletTypeCard    WalletType = "card"
	WalletTypeCash    WalletType = "cash"
)

func (t WalletType) Valid() bool {
	switch t {
	case WalletTypeBank, WalletTypeEwallet, WalletTypeCard, WalletTypeCash:
		return true
	}
	return false
}

// Day of month a card statement falls due; the table checks the same range.
const (
	MinDueDay = 1
	MaxDueDay = 31
)

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Type      WalletType
	Icon      string
	Currency  Currency
	Reference string
	Balance   int64
	// A balance only ever changes because someone typed it, so its age is tracked apart.
	BalanceUpdatedAt time.Time
	BalanceUpdatedBy *uuid.UUID
	IncludeInTotal   bool
	// Cards only; nil on every other type.
	CreditLimit *int64
	DueDay      *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *uuid.UUID
	UpdatedBy   *uuid.UUID
	// Only ever set on a soft-deleted row, which no read returns yet.
	DeletedBy *uuid.UUID
}

// CurrencyAmount: one currency's total, for the currencies the headline cannot state.
type CurrencyAmount struct {
	Currency Currency
	Amount   int64
}

// WalletOverview: the balance summary. Only base-currency wallets reach the
// headline; card debt and other currencies stay on their own lines, because
// there is no exchange rate to fold them in with.
type WalletOverview struct {
	BaseCurrency   Currency
	TotalHeld      int64
	CountedWallets int
	OwedOnCards    int64
	HeldByCurrency []CurrencyAmount
}

type WalletRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]Wallet, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Wallet, error)
	Create(ctx context.Context, wallet *Wallet) error
	Update(ctx context.Context, wallet *Wallet) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Overview(ctx context.Context, userID uuid.UUID) (WalletOverview, error)
}

type CreateWalletInput struct {
	Name           string
	Type           WalletType
	Icon           string
	Currency       Currency
	Reference      string
	Balance        int64
	IncludeInTotal bool
	CreditLimit    *int64
	DueDay         *int
}

type UpdateWalletInput struct {
	Name           *string
	Type           *WalletType
	Icon           *string
	Currency       *Currency
	Reference      *string
	Balance        *int64
	IncludeInTotal *bool
	CreditLimit    *int64
	DueDay         *int
}

type WalletService interface {
	List(ctx context.Context, userID uuid.UUID) ([]Wallet, error)
	Create(ctx context.Context, userID uuid.UUID, input CreateWalletInput) (*Wallet, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (*Wallet, error)
	Update(ctx context.Context, userID, id uuid.UUID, input UpdateWalletInput) (*Wallet, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Overview(ctx context.Context, userID uuid.UUID) (WalletOverview, error)
}
