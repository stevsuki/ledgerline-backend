package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type WalletService struct {
	walletRepo   domain.WalletRepository
	auditLogRepo domain.AuditLogRepository
}

func NewWalletService(walletRepo domain.WalletRepository, auditLogRepo domain.AuditLogRepository) domain.WalletService {
	return &WalletService{walletRepo: walletRepo, auditLogRepo: auditLogRepo}
}

func (s *WalletService) List(ctx context.Context, userID uuid.UUID) ([]domain.Wallet, error) {
	return s.walletRepo.List(ctx, userID)
}

func (s *WalletService) Overview(ctx context.Context, userID uuid.UUID) (domain.WalletOverview, error) {
	return s.walletRepo.Overview(ctx, userID)
}

// checkCardFields validates the card-only fields against the type they will be
// stored with. Sending them for anything but a card is a mistake worth naming.
func checkCardFields(walletType domain.WalletType, limit *int64, dueDay *int) error {
	if walletType != domain.WalletTypeCard {
		if limit != nil || dueDay != nil {
			return domain.InvalidInput(domain.CodeWalletInvalidCard,
				"credit_limit and due_day apply to a card wallet only").WithField("type")
		}
		return nil
	}

	if limit != nil && *limit < 0 {
		return domain.InvalidInput(domain.CodeWalletInvalidCard,
			"credit_limit must not be negative").WithField("credit_limit")
	}
	if dueDay != nil && (*dueDay < domain.MinDueDay || *dueDay > domain.MaxDueDay) {
		return domain.InvalidInput(domain.CodeWalletInvalidCard,
			fmt.Sprintf("due_day must be between %d and %d", domain.MinDueDay, domain.MaxDueDay)).
			WithField("due_day")
	}
	return nil
}

func (s *WalletService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateWalletInput) (*domain.Wallet, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.InvalidInput(domain.CodeWalletInvalidData, "wallet name is required").WithField("name")
	}

	if !input.Type.Valid() {
		return nil, domain.InvalidInput(domain.CodeWalletInvalidType, "wallet type must be bank, ewallet, card, or cash").WithField("type")
	}

	if !input.Currency.Valid() {
		return nil, domain.InvalidInput(domain.CodeWalletInvalidCurrency, "wallet currency must be IDR, USD, or SGD").WithField("currency")
	}

	if err := checkCardFields(input.Type, input.CreditLimit, input.DueDay); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	wallet := &domain.Wallet{
		ID:             id,
		UserID:         userID,
		Name:           name,
		Type:           input.Type,
		Icon:           strings.TrimSpace(input.Icon),
		Currency:       input.Currency,
		Reference:      strings.TrimSpace(input.Reference),
		Balance:        input.Balance,
		IncludeInTotal: input.IncludeInTotal,
		CreditLimit:    input.CreditLimit,
		DueDay:         input.DueDay,
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Wallet, error) {
	return s.walletRepo.GetByID(ctx, id, userID)
}

func (s *WalletService) Update(ctx context.Context, userID, id uuid.UUID, input domain.UpdateWalletInput) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.InvalidInput(domain.CodeWalletInvalidData, "wallet name must not be empty").WithField("name")
		}
		wallet.Name = name
	}

	if input.Type != nil {
		if !input.Type.Valid() {
			return nil, domain.InvalidInput(domain.CodeWalletInvalidType, "wallet type must be bank, ewallet, card, or cash").WithField("type")
		}
		wallet.Type = *input.Type
	}

	if input.Currency != nil {
		if !input.Currency.Valid() {
			return nil, domain.InvalidInput(domain.CodeWalletInvalidCurrency, "wallet currency must be IDR, USD, or SGD").WithField("currency")
		}
		wallet.Currency = *input.Currency
	}

	if input.Icon != nil {
		wallet.Icon = strings.TrimSpace(*input.Icon)
	}

	if input.Reference != nil {
		wallet.Reference = strings.TrimSpace(*input.Reference)
	}

	if input.Balance != nil {
		wallet.Balance = *input.Balance
	}

	if input.IncludeInTotal != nil {
		wallet.IncludeInTotal = *input.IncludeInTotal
	}

	if input.CreditLimit != nil {
		wallet.CreditLimit = input.CreditLimit
	}
	if input.DueDay != nil {
		wallet.DueDay = input.DueDay
	}

	// Judged against the type the wallet ends up with. Switching away from card
	// clears the two fields rather than failing: nothing displays them any more.
	if wallet.Type != domain.WalletTypeCard {
		if input.CreditLimit != nil || input.DueDay != nil {
			return nil, domain.InvalidInput(domain.CodeWalletInvalidCard,
				"credit_limit and due_day apply to a card wallet only").WithField("type")
		}
		wallet.CreditLimit, wallet.DueDay = nil, nil
	} else if err := checkCardFields(wallet.Type, wallet.CreditLimit, wallet.DueDay); err != nil {
		return nil, err
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.walletRepo.Delete(ctx, id, userID)
}
