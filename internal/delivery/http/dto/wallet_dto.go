package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// CreateWalletRequestDTO: create wallet payload, lengths and enums from migration 000018.
type CreateWalletRequestDTO struct {
	Name     string `json:"name" binding:"required,min=2,max=100" example:"BCA Payroll"`
	Type     string `json:"type" binding:"required,oneof=bank ewallet card cash" example:"bank"`
	Currency string `json:"currency" binding:"required,oneof=IDR USD SGD" example:"IDR"`
	// reference and icon are nullable in the table, so they stay optional here.
	Reference string `json:"reference" binding:"omitempty,max=50" example:"1234567890"`
	Icon      string `json:"icon" binding:"omitempty,max=50" example:"bank"`
	// No "required" on these two: it would reject a 0 balance and include_in_total=false.
	Balance        int64 `json:"balance" example:"41200000"`
	IncludeInTotal bool  `json:"include_in_total" example:"true"`
	// Card wallets only; sending either on another type is rejected.
	CreditLimit *int64 `json:"credit_limit" binding:"omitempty,min=0" example:"25000000"`
	DueDay      *int   `json:"due_day" binding:"omitempty,min=1,max=31" example:"18"`
}

func (r CreateWalletRequestDTO) ToInput() domain.CreateWalletInput {
	return domain.CreateWalletInput{
		Name:           r.Name,
		Type:           domain.WalletType(r.Type),
		Icon:           r.Icon,
		Currency:       domain.Currency(r.Currency),
		Reference:      r.Reference,
		Balance:        r.Balance,
		IncludeInTotal: r.IncludeInTotal,
		CreditLimit:    r.CreditLimit,
		DueDay:         r.DueDay,
	}
}

// UpdateWalletRequestDTO: pointers so partial updates are detectable.
type UpdateWalletRequestDTO struct {
	Name           *string `json:"name" binding:"omitempty,min=2,max=100" example:"BCA Utama"`
	Type           *string `json:"type" binding:"omitempty,oneof=bank ewallet card cash" example:"bank"`
	Currency       *string `json:"currency" binding:"omitempty,oneof=IDR USD SGD" example:"IDR"`
	Reference      *string `json:"reference" binding:"omitempty,max=50" example:"1234567890"`
	Icon           *string `json:"icon" binding:"omitempty,max=50" example:"bank"`
	Balance        *int64  `json:"balance" example:"41200000"`
	IncludeInTotal *bool   `json:"include_in_total" example:"true"`
	CreditLimit    *int64  `json:"credit_limit" binding:"omitempty,min=0" example:"25000000"`
	DueDay         *int    `json:"due_day" binding:"omitempty,min=1,max=31" example:"18"`
}

func (r UpdateWalletRequestDTO) ToInput() domain.UpdateWalletInput {
	input := domain.UpdateWalletInput{
		Name:           r.Name,
		Icon:           r.Icon,
		Reference:      r.Reference,
		Balance:        r.Balance,
		IncludeInTotal: r.IncludeInTotal,
		CreditLimit:    r.CreditLimit,
		DueDay:         r.DueDay,
	}
	if r.Type != nil {
		t := domain.WalletType(*r.Type)
		input.Type = &t
	}
	if r.Currency != nil {
		c := domain.Currency(*r.Currency)
		input.Currency = &c
	}
	return input
}

// WalletResponseDTO: raw figures only. Widths, colours, and formatted money are
// the client's to decide; sending them here would freeze the design in the API.
type WalletResponseDTO struct {
	ID        uuid.UUID `json:"id" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
	Name      string    `json:"name" example:"BCA Payroll"`
	Type      string    `json:"type" example:"bank"`
	Icon      string    `json:"icon" example:"bank"`
	Currency  string    `json:"currency" example:"IDR"`
	Reference string    `json:"reference" example:"1234567890"`
	Balance   int64     `json:"balance" example:"41200000"`
	// When the balance itself last changed, which is not when the row last changed.
	BalanceUpdatedAt time.Time `json:"balance_updated_at" example:"2026-08-26T15:04:05Z"`
	IncludeInTotal   bool      `json:"include_in_total" example:"true"`
	// null on every type but card.
	CreditLimit *int64     `json:"credit_limit" example:"25000000"`
	DueDay      *int       `json:"due_day" example:"18"`
	CreatedAt   time.Time  `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2026-01-02T15:04:05Z"`
	CreatedBy   *uuid.UUID `json:"created_by" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
	UpdatedBy   *uuid.UUID `json:"updated_by" example:"6f1e2b7e-2c8a-4c1f-9f3e-6a0f1c2d3e4b"`
}

func NewWalletResponseDTO(w *domain.Wallet) WalletResponseDTO {
	return WalletResponseDTO{
		ID:               w.ID,
		Name:             w.Name,
		Type:             string(w.Type),
		Icon:             w.Icon,
		Currency:         string(w.Currency),
		Reference:        w.Reference,
		Balance:          w.Balance,
		BalanceUpdatedAt: w.BalanceUpdatedAt,
		IncludeInTotal:   w.IncludeInTotal,
		CreditLimit:      w.CreditLimit,
		DueDay:           w.DueDay,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		CreatedBy:        w.CreatedBy,
		UpdatedBy:        w.UpdatedBy,
	}
}

func NewWalletResponseDTOs(wallets []domain.Wallet) []WalletResponseDTO {
	out := make([]WalletResponseDTO, 0, len(wallets))
	for i := range wallets {
		out = append(out, NewWalletResponseDTO(&wallets[i]))
	}
	return out
}

type CurrencyAmountResponseDTO struct {
	Currency string `json:"currency" example:"USD"`
	Amount   int64  `json:"amount" example:"1480"`
}

type WalletOverviewResponseDTO struct {
	BaseCurrency string `json:"base_currency" example:"IDR"`
	TotalHeld    int64  `json:"total_held" example:"43680000"`
	// How many wallets total_held is made of.
	CountedWallets int `json:"counted_wallets" example:"3"`
	// Negative, or 0 when nothing is owed.
	OwedOnCards    int64                       `json:"owed_on_cards" example:"-3240000"`
	HeldByCurrency []CurrencyAmountResponseDTO `json:"held_by_currency"`
}

func NewWalletOverviewResponseDTO(o domain.WalletOverview) WalletOverviewResponseDTO {
	held := make([]CurrencyAmountResponseDTO, 0, len(o.HeldByCurrency))
	for _, item := range o.HeldByCurrency {
		held = append(held, CurrencyAmountResponseDTO{
			Currency: string(item.Currency),
			Amount:   item.Amount,
		})
	}

	return WalletOverviewResponseDTO{
		BaseCurrency:   string(o.BaseCurrency),
		TotalHeld:      o.TotalHeld,
		CountedWallets: o.CountedWallets,
		OwedOnCards:    o.OwedOnCards,
		HeldByCurrency: held,
	}
}
